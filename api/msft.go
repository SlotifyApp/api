package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/coreos/go-oidc/v3/oidc"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

const (
	MicrosoftLogin = "https://login.microsoftonline.com/%s"
)

// ErrMSALCache is raised when MSAL could not find something in its cache.
var ErrMSALCache = errors.New("MSAL could not find resource could not be found in MSAL cache")

// getMSFTScopes will return the requested scopes for a MSFT access token.
func getMSFTScopes() []string {
	return []string{
		oidc.ScopeOpenID, "profile", "email", "User.ReadWrite",
		"Calendars.ReadBasic", "Calendars.Read", "Calendars.ReadWrite",
		"Calendars.ReadWrite.Shared",
	}
}

// getMSFTEntraValues returns the MSFT Entra tenantID, clientID and clientSecret.
func getMSFTEntraValues() (MSFTEntraValues, error) {
	var tenantID string
	var clientID string
	var clientSecret string
	var present bool
	// Check if the microsoft entra environment vars are set
	if tenantID, present = os.LookupEnv(TenantIDEnvName); !present {
		return MSFTEntraValues{}, fmt.Errorf("failed to get %s env variable", TenantIDEnvName)
	}

	if clientID, present = os.LookupEnv(ClientIDEnvName); !present {
		return MSFTEntraValues{}, fmt.Errorf("failed to get %s env variable", ClientIDEnvName)
	}

	if clientSecret, present = os.LookupEnv(ClientSecretEnvName); !present {
		return MSFTEntraValues{}, fmt.Errorf("failed to get %s env variable", ClientSecretEnvName)
	}

	return MSFTEntraValues{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TenantID:     tenantID,
	}, nil
}

type MSFTTokenResult struct {
	Email         string
	FirstName     string
	LastName      string
	HomeAccountID string
}

// createMSALClient creates a new confiential MSAL Client.
func createMSALClient() (confidential.Client, error) {
	msftEntraVals, err := getMSFTEntraValues()
	if err != nil {
		return confidential.Client{}, fmt.Errorf("failed to get msft entra values: %w", err)
	}
	cred, err := confidential.NewCredFromSecret(msftEntraVals.ClientSecret)
	if err != nil {
		return confidential.Client{}, fmt.Errorf("failed to create new cred from secret: %w", err)
	}

	// Create new MSAL Client
	c, err := confidential.New(fmt.Sprintf(MicrosoftLogin, msftEntraVals.TenantID), msftEntraVals.ClientID, cred)
	if err != nil {
		return confidential.Client{}, fmt.Errorf("failed to create new confidential client: %w", err)
	}
	return c, nil
}

// GetMSFTAccessToken gets a MSFT access token for a user.
func GetMSFTAccessToken(ctx context.Context, c *confidential.Client,
	db *database.Database, userID uint32,
) (azcore.AccessToken, error) {
	ctx, cancel := context.WithTimeout(ctx, database.DatabaseTimeout)
	defer cancel()

	user, err := db.GetUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			log.Print("getMSFTaccesstoken: context cancelled")
			return azcore.AccessToken{},
				fmt.Errorf("failed to get user by id: context cancelled: %w", err)
		case errors.Is(err, context.DeadlineExceeded):
			log.Print("getMSFTaccesstoken: query timed out")
			return azcore.AccessToken{},
				fmt.Errorf("failed to get user by id: query timed out: %w", err)
		default:
			log.Print("getMSFTaccesstoken: query timed out")
			return azcore.AccessToken{},
				fmt.Errorf("failed to get user by id: %w", err)
		}
	}

	// HomeAccountID is NULL, should not happen. This is set when a user logs in.
	if !user.MsftHomeAccountID.Valid {
		return azcore.AccessToken{}, errors.New("msft_home_account_id was null")
	}

	// msal attempts to get account cache, if this fails user has to reauthenticate
	account, err := c.Account(ctx, user.MsftHomeAccountID.String)
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("msal failed to get account by home account id: %w: %w", ErrMSALCache, err)
	}

	if account.IsZero() {
		return azcore.AccessToken{}, errors.New("msal account returned was zero value")
	}

	// msal attempts to get account cache, if this fails user has to reauthenticate
	res, err := c.AcquireTokenSilent(ctx, getMSFTScopes(), confidential.WithSilentAccount(account))
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("msal failed to get token silently: %w: %w", ErrMSALCache, err)
	}

	tk := azcore.AccessToken{
		ExpiresOn: res.ExpiresOn,
		Token:     res.AccessToken,
	}

	return tk, nil
}

// MSFTAuthoriseByCode will exchange a authorisation code with an access token.
// The home account id is stored, using this an access token can be gained.
func MSFTAuthoriseByCode(ctx context.Context, c *confidential.Client, authCode string) (MSFTTokenResult, error) {
	// MSAL fn to exchange auth code for token
	res, err := c.AcquireTokenByAuthCode(ctx, authCode, "http://localhost:8080/api/auth/callback", getMSFTScopes())
	if err != nil {
		return MSFTTokenResult{}, fmt.Errorf("failed to get token by auth code: %w", err)
	}

	if res.IDToken.IsZero() {
		return MSFTTokenResult{}, errors.New("msft id token was zero value")
	}

	if res.Account.IsZero() {
		return MSFTTokenResult{}, errors.New("msft id token was zero value")
	}

	firstName := res.IDToken.GivenName
	lastName := res.IDToken.FamilyName
	// If the id token doesn't contain given and family names, then attempt to
	// manually split by getting Name
	if firstName == "" && lastName == "" {
		firstName, lastName = splitName(res.IDToken.Name)
	}

	return MSFTTokenResult{
		Email:         res.IDToken.Email,
		FirstName:     firstName,
		LastName:      lastName,
		HomeAccountID: res.Account.HomeAccountID,
	}, nil
}

type SlotifyAccessTokenProvider struct {
	accessToken azcore.AccessToken
}

func (satp SlotifyAccessTokenProvider) GetToken(_ context.Context,
	_ policy.TokenRequestOptions,
) (azcore.AccessToken, error) {
	return satp.accessToken, nil
}

// CreateMSFTGraphClient creates a MSGraph SDK Client.
func CreateMSFTGraphClient(accessToken azcore.AccessToken) (*msgraphsdk.GraphServiceClient, error) {
	satp := SlotifyAccessTokenProvider{
		accessToken: accessToken,
	}
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(satp, getMSFTScopes())
	if err != nil {
		return nil, fmt.Errorf("failed to create new msgraph service: %w", err)
	}
	return client, nil
}

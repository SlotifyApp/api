package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

// MSFTEntraValues stores details specific to our MSFT tenant.
type MSFTEntraValues struct {
	ClientID     string
	ClientSecret string
	TenantID     string
}

// MSFTTokenResult contains filtered fields from the MSAL client exchanging an auth code for an access token,
// got through the OpenID flow.
type MSFTTokenResult struct {
	Email         string
	FirstName     string
	LastName      string
	HomeAccountID string
}

// getMSFTScopes will return the requested scopes for a MSFT access token.
func getMSFTScopes() []string {
	return []string{
		oidc.ScopeOpenID, "profile", "email", "User.ReadWrite.All",
		"Calendars.ReadBasic", "Calendars.Read", "Calendars.ReadWrite",
		"Calendars.ReadWrite.Shared", "Group.Read.All", "Group.ReadWrite.All",
		"Place.Read.All", "People.Read.All",
	}
}

// getMSFTEntraValues reads and returns the MSFT Entra tenantID, clientID and clientSecret.
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

// getMSFTAccessToken gets a MSFT access token for a user.
func getMSFTAccessToken(ctx context.Context, c *confidential.Client,
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

// splitName splits a name into first name and last name.
// if there are fewer than 2 individual names, return the name.
// if there are more than 2 names, join the last names to form our own last name.
func splitName(name string) (string, string) {
	names := strings.Fields(name)
	var firstName string
	var lastName string

	if len(names) > 0 {
		firstName = names[0]
	}

	if len(names) > 1 {
		// Join the rest of the fields together to form the
		// last name
		lastName = strings.Join(names[1:], " ")
	}
	return firstName, lastName
}

// msftAuthoriseByCode will exchange a authorisation code for a MSFT access token, following OAuth2.
// Due to the OpenID protocol, we get user information as well eg. first name.
// The home account id is stored so access token for a user can be gained when the user is logged out.
func msftAuthoriseByCode(ctx context.Context,
	msalClient *confidential.Client,
	authCode string,
) (MSFTTokenResult, error) {
	backendURL, present := os.LookupEnv("BACKEND_URL")
	if !present {
		return MSFTTokenResult{}, errors.New("failed to get BACKEND_URL env value")
	}
	// exchange authorisation code for access token
	res, err := msalClient.AcquireTokenByAuthCode(ctx,
		authCode,
		fmt.Sprintf("%s/api/auth/callback", backendURL),
		getMSFTScopes(),
	)
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
	// manually split Name
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

// AccessTokenProvider allows for passing in an access token directly into the MSGraph SDK.
type AccessTokenProvider struct {
	accessToken azcore.AccessToken
}

func (atp AccessTokenProvider) GetToken(_ context.Context,
	_ policy.TokenRequestOptions,
) (azcore.AccessToken, error) {
	return atp.accessToken, nil
}

// createMSFTGraphClientWithAccessToken creates a MSGraph SDK Client given an access token.
func createMSFTGraphClientWithAccessToken(accessToken azcore.AccessToken) (*msgraphsdk.GraphServiceClient, error) {
	atp := AccessTokenProvider{
		accessToken: accessToken,
	}
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(atp, getMSFTScopes())
	if err != nil {
		return nil, fmt.Errorf("failed to create new msgraph service: %w", err)
	}
	return client, nil
}

// CreateMSFTGraphClient gets a MSFT access token for a user and creates a graph client with it.
func CreateMSFTGraphClient(ctx context.Context, msalClient *confidential.Client,
	db *database.Database, userID uint32,
) (*msgraphsdk.GraphServiceClient, error) {
	at, err := getMSFTAccessToken(ctx, msalClient, db, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get msft access token: %w", err)
	}

	graph, err := createMSFTGraphClientWithAccessToken(at)
	if err != nil || graph == nil {
		return nil, fmt.Errorf("failed to create msft graph client: %w", err)
	}
	return graph, nil
}

// getOrInsertUserByClaimEmail will get a user by the claim email,
// or if first time log in, it will create a new user.
func getOrInsertUserByClaimEmail(ctx context.Context,
	qtx *database.Queries, msftTokenRes MSFTTokenResult,
) (database.User, error) {
	email := msftTokenRes.Email
	// Double the timeout due to more db operations
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	count, err := qtx.CountUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, fmt.Errorf("failed to get user count by claim email: %w", err)
	}

	// User doesn't exist, first time signing so sign up
	if count == 0 {
		dbParams := database.CreateUserParams{
			Email:     email,
			FirstName: msftTokenRes.FirstName,
			LastName:  msftTokenRes.LastName,
		}
		_, err = qtx.CreateUser(ctx, dbParams)
		if err != nil {
			return database.User{}, fmt.Errorf("failed to create user for claim email: %w", err)
		}
	}

	var u database.User

	u, err = qtx.GetUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, fmt.Errorf("failed to get user with claim email: %w", err)
	}

	// Update user's home account id so it can be used when asking for a MSFT access token
	dbParams := database.UpdateUserHomeAccountIDParams{
		ID:                u.ID,
		MsftHomeAccountID: sql.NullString{String: msftTokenRes.HomeAccountID, Valid: true},
	}

	var rowsAffected int64
	rowsAffected, err = qtx.UpdateUserHomeAccountID(ctx, dbParams)
	if err != nil {
		return database.User{}, fmt.Errorf("failed to update user home account id: %w", err)
	}

	// UpdateUserHomeAccountID should only either update 1 or 0 rows.
	if rowsAffected > 1 {
		err = database.WrongNumberSQLRowsError{
			ActualRows:   rowsAffected,
			ExpectedRows: []int64{0, 1},
		}
		return database.User{}, fmt.Errorf("failed to update home account id: %w", err)
	}

	return u, nil
}

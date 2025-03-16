package api

import (
	"errors"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/oapi-codegen/runtime/types"
)

func GroupableToMSFTGroup(g models.Groupable) (MSFTGroup, error) {
	if g.GetId() == nil || g.GetDisplayName() == nil {
		return MSFTGroup{}, errors.New("no id or name for microsoft group")
	}

	return MSFTGroup{
		Id:   *g.GetId(),
		Name: *g.GetDisplayName(),
	}, nil
}

func UserableToMSFTUser(u models.Userable) (MSFTUser, error) {
	if u.GetId() == nil {
		return MSFTUser{}, errors.New("missing user ID")
	}

	email := "no-email@placeholder.com"
	if u.GetMail() != nil {
		email = *u.GetMail()
	}

	firstName := ""
	if u.GetGivenName() != nil {
		firstName = *u.GetGivenName()
	}

	lastName := ""
	if u.GetSurname() != nil {
		lastName = *u.GetSurname()
	}

	return MSFTUser{
		Email:     types.Email(email),
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

func PersonableToMSFTUser(u models.Personable) (MSFTUser, error) {
	if u.GetId() == nil || u.GetScoredEmailAddresses() == nil {
		return MSFTUser{}, errors.New("missing required fields")
	}
	firstName := "No First Name Found"
	if u.GetGivenName() != nil {
		firstName = *u.GetGivenName()
	}
	lastName := "No Last Name Found"
	if u.GetSurname() != nil {
		lastName = *u.GetSurname()
	}
	var email string
	if len(u.GetScoredEmailAddresses()) == 0 {
		email = "No Email Found"
	} else {
		if addr := u.GetScoredEmailAddresses()[0].GetAddress(); addr != nil {
			email = *addr
		} else {
			email = "No Email Found"
		}
	}
	return MSFTUser{
		Email:     types.Email(email),
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

func GetsToMSFTGroups(d models.DirectoryObjectCollectionResponseable) ([]MSFTGroup, error) {
	var groups []MSFTGroup
	if d.GetValue() != nil {
		for _, dirs := range d.GetValue() {
			if grp, ok := dirs.(models.Groupable); ok {
				group, err := GroupableToMSFTGroup(grp)
				if err != nil {
					return []MSFTGroup{}, errors.New("failed to convert groupable to microsoft group")
				}
				groups = append(groups, group)
			}
		}
		return groups, nil
	}
	return []MSFTGroup{}, errors.New("empty gets")
}

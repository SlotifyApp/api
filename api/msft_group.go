package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/oapi-codegen/runtime/types"
)

func GroupableToMSFTGroup(g models.Groupable) (MSFTGroup, error) {
	if g.GetId() == nil || g.GetDisplayName() == nil {
		return MSFTGroup{}, errors.New("no id or name for microsoft group")
	}

	id, err := strconv.ParseUint(*g.GetId(), 10, 32)
	if err != nil {
		return MSFTGroup{}, fmt.Errorf("error converting microsoft group id '%s' to uint32: %w", *g.GetId(), err)
	}
	return MSFTGroup{
		Id:   uint32(id),
		Name: *g.GetDisplayName(),
	}, nil
}

func UserableToMSFTGroupUser(u models.Userable) (MSFTGroupUser, error) {
	if u.GetId() == nil {
		return MSFTGroupUser{}, errors.New("missing user ID")
	}

	email := *u.GetMail()
	if email == "" {
		email = "no-email@placeholder.com"
	}

	firstName := ""
	if u.GetGivenName() != nil {
		firstName = *u.GetGivenName()
	}

	lastName := ""
	if u.GetSurname() != nil {
		lastName = *u.GetSurname()
	}

	return MSFTGroupUser{
		Email:     types.Email(email),
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

func PersonableToMSFTGroupUser(u models.Personable) (MSFTGroupUser, error) {
	if u.GetId() == nil || u.GetScoredEmailAddresses() == nil || u.GetGivenName() == nil || u.GetSurname() == nil {
		return MSFTGroupUser{}, errors.New("missing required fields")
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
	return MSFTGroupUser{
		Email:     types.Email(email),
		FirstName: *u.GetGivenName(),
		LastName:  *u.GetSurname(),
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

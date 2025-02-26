package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/oapi-codegen/runtime/types"
)

func GroupableToGroup(g models.Groupable) (Group, error) {
	if g.GetId() == nil || g.GetDisplayName() == nil {
		return Group{}, errors.New("no id or name for group")
	}

	id, err := strconv.ParseUint(*g.GetId(), 10, 32)
	if err != nil {
		return Group{}, fmt.Errorf("error converting group id '%s' to uint32: %w", *g.GetId(), err)
	}
	return Group{
		Id:   uint32(id),
		Name: *g.GetDisplayName(),
	}, nil
}

func UserableToUser(u models.Userable) (GroupUser, error) {
	if u.GetId() == nil || u.GetMail() == nil || u.GetGivenName() == nil || u.GetSurname() == nil {
		return GroupUser{}, errors.New("missing required fields")
	}
	return GroupUser{
		Email:     types.Email(*u.GetMail()),
		FirstName: *u.GetGivenName(),
		LastName:  *u.GetSurname(),
	}, nil
}

func GetsToGroups(d models.DirectoryObjectCollectionResponseable) ([]Group, error) {
	var groups []Group
	if d.GetValue() != nil {
		for _, dirs := range d.GetValue() {
			if grp, ok := dirs.(models.Groupable); ok {
				group, err := GroupableToGroup(grp)
				if err != nil {
					return []Group{}, errors.New("failed to convert groupable to group")
				}
				groups = append(groups, group)
			}
		}
		return groups, nil
	}
	return []Group{}, errors.New("empty gets")
}

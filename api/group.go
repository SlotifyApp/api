package api

import (
	"fmt"
	"strconv"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/oapi-codegen/runtime/types"
)

func GroupableToGroup(g models.Groupable) (Group, error) {
	if g.GetId() == nil || g.GetDisplayName() == nil {
		return Group{}, fmt.Errorf("no id or name for group")
	}

	id, err := strconv.ParseUint(*g.GetId(), 10, 32)
	if err != nil {
		return Group{}, fmt.Errorf("error converting group id '%s' to uint32: %v", *g.GetId(), err)
	}
	return Group{
		Id:   uint32(id),
		Name: *g.GetDisplayName(),
	}, nil
}

func UserableToUser(u models.Userable) (User, error) {
	if u.GetId() == nil || u.GetMail() == nil || u.GetGivenName() == nil || u.GetSurname() == nil {
		return User{}, fmt.Errorf("missing required fields")
	}
	id, err := strconv.ParseUint(*u.GetId(), 10, 32)
	if err != nil {
		return User{}, fmt.Errorf("error conveting user id '%s' to uint32: %v", *u.GetId(), err)
	}
	return User{
		Id:        uint32(id),
		Email:     types.Email(*u.GetMail()),
		FirstName: *u.GetGivenName(),
		LastName:  *u.GetSurname(),
	}, nil
}

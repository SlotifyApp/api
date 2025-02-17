package api

import (
	"fmt"
	"strconv"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
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

package api

import (
	"net/http"
)

// Get a group by query params.
// (GET /api/groups)
func (s Server) GetAPIGroups(w http.ResponseWriter, r *http.Request, params GetAPIGroupsParams) {
}

// Get all groups for current user.
// (GET /api/groups/me)
func (s Server) GetAPIGroupsMe(w http.ResponseWriter, r *http.Request) {
}

// Get a group by id.
// (GET /api/groups/{groupID})
func (s Server) GetAPIGroupsGroupID(w http.ResponseWriter, r *http.Request, groupID uint32) {

}

// Get all members of a group.
// (GET /api/groups/{groupID}/users)
func (s Server) GetAPIGroupsGroupIDUsers(w http.ResponseWriter, r *http.Request, groupID uint32) {

}

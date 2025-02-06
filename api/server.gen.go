// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/oapi-codegen/runtime"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CalendarEvent defines model for CalendarEvent.
type CalendarEvent struct {
	EndTime   *string `json:"endTime,omitempty"`
	StartTime *string `json:"startTime,omitempty"`
	Subject   *string `json:"subject,omitempty"`
}

// Notification defines model for Notification.
type Notification struct {
	Created time.Time `json:"created"`
	Id      uint32    `json:"id"`
	Message string    `json:"message"`
}

// Team defines model for Team.
type Team struct {
	Id   uint32 `json:"id"`
	Name string `json:"name"`
}

// TeamCreate defines model for TeamCreate.
type TeamCreate struct {
	Name string `json:"name"`
}

// User defines model for User.
type User struct {
	Email     openapi_types.Email `json:"email"`
	FirstName string              `json:"firstName"`
	Id        uint32              `json:"id"`
	LastName  string              `json:"lastName"`
}

// UserCreate defines model for UserCreate.
type UserCreate struct {
	Email     openapi_types.Email `json:"email"`
	FirstName string              `json:"firstName"`
	LastName  string              `json:"lastName"`
}

// GetAPIAuthCallbackParams defines parameters for GetAPIAuthCallback.
type GetAPIAuthCallbackParams struct {
	Code  string `form:"code" json:"code"`
	State string `form:"state" json:"state"`
}

// GetAPITeamsParams defines parameters for GetAPITeams.
type GetAPITeamsParams struct {
	// Name Team name
	Name *string `form:"name,omitempty" json:"name,omitempty"`
}

// GetAPIUsersParams defines parameters for GetAPIUsers.
type GetAPIUsersParams struct {
	// Email Email of user
	Email *openapi_types.Email `form:"email,omitempty" json:"email,omitempty"`

	// FirstName First name of user
	FirstName *string `form:"firstName,omitempty" json:"firstName,omitempty"`

	// LastName Last name of user
	LastName *string `form:"lastName,omitempty" json:"lastName,omitempty"`
}

// PostAPITeamsJSONRequestBody defines body for PostAPITeams for application/json ContentType.
type PostAPITeamsJSONRequestBody = TeamCreate

// PostAPIUsersJSONRequestBody defines body for PostAPIUsers for application/json ContentType.
type PostAPIUsersJSONRequestBody = UserCreate

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Auth route for authorisation code flow
	// (GET /api/auth/callback)
	GetAPIAuthCallback(w http.ResponseWriter, r *http.Request, params GetAPIAuthCallbackParams)
	// get a user's calendar events
	// (GET /api/calendar/me)
	GetAPICalendarMe(w http.ResponseWriter, r *http.Request)
	// Subscribe to notifications
	// (GET /api/events)
	RenderEvent(w http.ResponseWriter, r *http.Request)
	// Healthcheck route
	// (GET /api/healthcheck)
	GetAPIHealthcheck(w http.ResponseWriter, r *http.Request)
	// CORS preflight for marking a notification as read
	// (OPTIONS /api/notifications/{notificationID}/read)
	OptionsAPINotificationsNotificationIDRead(w http.ResponseWriter, r *http.Request, notificationID uint32)
	// mark a notification as being read
	// (PATCH /api/notifications/{notificationID}/read)
	PatchAPINotificationsNotificationIDRead(w http.ResponseWriter, r *http.Request, notificationID uint32)
	// Refresh Slotify access token and refresh token
	// (POST /api/refresh)
	PostAPIRefresh(w http.ResponseWriter, r *http.Request)
	// Get a team by query params
	// (GET /api/teams)
	GetAPITeams(w http.ResponseWriter, r *http.Request, params GetAPITeamsParams)
	// CORS preflight for teams
	// (OPTIONS /api/teams)
	OptionsAPITeams(w http.ResponseWriter, r *http.Request)
	// Create a new team
	// (POST /api/teams)
	PostAPITeams(w http.ResponseWriter, r *http.Request)
	// Get all joinable teams for a user excluding teams they are already a part of
	// (GET /api/teams/joinable/me)
	GetAPITeamsJoinableMe(w http.ResponseWriter, r *http.Request)
	// Get all teams for user by id passed by JWT
	// (GET /api/teams/me)
	GetAPITeamsMe(w http.ResponseWriter, r *http.Request)
	// Delete a team by id
	// (DELETE /api/teams/{teamID})
	DeleteAPITeamsTeamID(w http.ResponseWriter, r *http.Request, teamID uint32)
	// Get a team by id
	// (GET /api/teams/{teamID})
	GetAPITeamsTeamID(w http.ResponseWriter, r *http.Request, teamID uint32)
	// Get all members of a team
	// (GET /api/teams/{teamID}/users)
	GetAPITeamsTeamIDUsers(w http.ResponseWriter, r *http.Request, teamID uint32)
	// Add current user to a team
	// (POST /api/teams/{teamID}/users/me)
	PostAPITeamsTeamIDUsersMe(w http.ResponseWriter, r *http.Request, teamID uint32)
	// Add a user to a team
	// (POST /api/teams/{teamID}/users/{userID})
	PostAPITeamsTeamIDUsersUserID(w http.ResponseWriter, r *http.Request, teamID uint32, userID uint32)
	// Get a user by query params
	// (GET /api/users)
	GetAPIUsers(w http.ResponseWriter, r *http.Request, params GetAPIUsersParams)
	// Create a new user
	// (POST /api/users)
	PostAPIUsers(w http.ResponseWriter, r *http.Request)
	// Get the user by id passed by JWT
	// (GET /api/users/me)
	GetAPIUsersMe(w http.ResponseWriter, r *http.Request)
	// Logout user
	// (POST /api/users/me/logout)
	PostAPIUsersMeLogout(w http.ResponseWriter, r *http.Request)
	// get user's unread notifications
	// (GET /api/users/me/notifications)
	GetAPIUsersMeNotifications(w http.ResponseWriter, r *http.Request)
	// Delete a user by id
	// (DELETE /api/users/{userID})
	DeleteAPIUsersUserID(w http.ResponseWriter, r *http.Request, userID uint32)
	// Get a user by id
	// (GET /api/users/{userID})
	GetAPIUsersUserID(w http.ResponseWriter, r *http.Request, userID uint32)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetAPIAuthCallback operation middleware
func (siw *ServerInterfaceWrapper) GetAPIAuthCallback(w http.ResponseWriter, r *http.Request) {

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetAPIAuthCallbackParams

	// ------------- Required query parameter "code" -------------

	if paramValue := r.URL.Query().Get("code"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "code"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "code", r.URL.Query(), &params.Code)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "code", Err: err})
		return
	}

	// ------------- Required query parameter "state" -------------

	if paramValue := r.URL.Query().Get("state"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "state"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "state", r.URL.Query(), &params.State)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "state", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPIAuthCallback(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPICalendarMe operation middleware
func (siw *ServerInterfaceWrapper) GetAPICalendarMe(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPICalendarMe(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// RenderEvent operation middleware
func (siw *ServerInterfaceWrapper) RenderEvent(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.RenderEvent(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPIHealthcheck operation middleware
func (siw *ServerInterfaceWrapper) GetAPIHealthcheck(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPIHealthcheck(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// OptionsAPINotificationsNotificationIDRead operation middleware
func (siw *ServerInterfaceWrapper) OptionsAPINotificationsNotificationIDRead(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "notificationID" -------------
	var notificationID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "notificationID", mux.Vars(r)["notificationID"], &notificationID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "notificationID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.OptionsAPINotificationsNotificationIDRead(w, r, notificationID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PatchAPINotificationsNotificationIDRead operation middleware
func (siw *ServerInterfaceWrapper) PatchAPINotificationsNotificationIDRead(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "notificationID" -------------
	var notificationID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "notificationID", mux.Vars(r)["notificationID"], &notificationID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "notificationID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PatchAPINotificationsNotificationIDRead(w, r, notificationID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PostAPIRefresh operation middleware
func (siw *ServerInterfaceWrapper) PostAPIRefresh(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostAPIRefresh(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPITeams operation middleware
func (siw *ServerInterfaceWrapper) GetAPITeams(w http.ResponseWriter, r *http.Request) {

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetAPITeamsParams

	// ------------- Optional query parameter "name" -------------

	err = runtime.BindQueryParameter("form", true, false, "name", r.URL.Query(), &params.Name)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "name", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPITeams(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// OptionsAPITeams operation middleware
func (siw *ServerInterfaceWrapper) OptionsAPITeams(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.OptionsAPITeams(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PostAPITeams operation middleware
func (siw *ServerInterfaceWrapper) PostAPITeams(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostAPITeams(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPITeamsJoinableMe operation middleware
func (siw *ServerInterfaceWrapper) GetAPITeamsJoinableMe(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPITeamsJoinableMe(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPITeamsMe operation middleware
func (siw *ServerInterfaceWrapper) GetAPITeamsMe(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPITeamsMe(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// DeleteAPITeamsTeamID operation middleware
func (siw *ServerInterfaceWrapper) DeleteAPITeamsTeamID(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "teamID" -------------
	var teamID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "teamID", mux.Vars(r)["teamID"], &teamID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "teamID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeleteAPITeamsTeamID(w, r, teamID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPITeamsTeamID operation middleware
func (siw *ServerInterfaceWrapper) GetAPITeamsTeamID(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "teamID" -------------
	var teamID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "teamID", mux.Vars(r)["teamID"], &teamID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "teamID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPITeamsTeamID(w, r, teamID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPITeamsTeamIDUsers operation middleware
func (siw *ServerInterfaceWrapper) GetAPITeamsTeamIDUsers(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "teamID" -------------
	var teamID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "teamID", mux.Vars(r)["teamID"], &teamID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "teamID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPITeamsTeamIDUsers(w, r, teamID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PostAPITeamsTeamIDUsersMe operation middleware
func (siw *ServerInterfaceWrapper) PostAPITeamsTeamIDUsersMe(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "teamID" -------------
	var teamID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "teamID", mux.Vars(r)["teamID"], &teamID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "teamID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostAPITeamsTeamIDUsersMe(w, r, teamID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PostAPITeamsTeamIDUsersUserID operation middleware
func (siw *ServerInterfaceWrapper) PostAPITeamsTeamIDUsersUserID(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "teamID" -------------
	var teamID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "teamID", mux.Vars(r)["teamID"], &teamID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "teamID", Err: err})
		return
	}

	// ------------- Path parameter "userID" -------------
	var userID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "userID", mux.Vars(r)["userID"], &userID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostAPITeamsTeamIDUsersUserID(w, r, teamID, userID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPIUsers operation middleware
func (siw *ServerInterfaceWrapper) GetAPIUsers(w http.ResponseWriter, r *http.Request) {

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetAPIUsersParams

	// ------------- Optional query parameter "email" -------------

	err = runtime.BindQueryParameter("form", true, false, "email", r.URL.Query(), &params.Email)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "email", Err: err})
		return
	}

	// ------------- Optional query parameter "firstName" -------------

	err = runtime.BindQueryParameter("form", true, false, "firstName", r.URL.Query(), &params.FirstName)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "firstName", Err: err})
		return
	}

	// ------------- Optional query parameter "lastName" -------------

	err = runtime.BindQueryParameter("form", true, false, "lastName", r.URL.Query(), &params.LastName)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "lastName", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPIUsers(w, r, params)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PostAPIUsers operation middleware
func (siw *ServerInterfaceWrapper) PostAPIUsers(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostAPIUsers(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPIUsersMe operation middleware
func (siw *ServerInterfaceWrapper) GetAPIUsersMe(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPIUsersMe(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// PostAPIUsersMeLogout operation middleware
func (siw *ServerInterfaceWrapper) PostAPIUsersMeLogout(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostAPIUsersMeLogout(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPIUsersMeNotifications operation middleware
func (siw *ServerInterfaceWrapper) GetAPIUsersMeNotifications(w http.ResponseWriter, r *http.Request) {

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPIUsersMeNotifications(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// DeleteAPIUsersUserID operation middleware
func (siw *ServerInterfaceWrapper) DeleteAPIUsersUserID(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "userID" -------------
	var userID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "userID", mux.Vars(r)["userID"], &userID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeleteAPIUsersUserID(w, r, userID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

// GetAPIUsersUserID operation middleware
func (siw *ServerInterfaceWrapper) GetAPIUsersUserID(w http.ResponseWriter, r *http.Request) {

	var err error

	// ------------- Path parameter "userID" -------------
	var userID uint32

	err = runtime.BindStyledParameterWithOptions("simple", "userID", mux.Vars(r)["userID"], &userID, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "userID", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPIUsersUserID(w, r, userID)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r)
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{})
}

type GorillaServerOptions struct {
	BaseURL          string
	BaseRouter       *mux.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r *mux.Router) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r *mux.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options GorillaServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = mux.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.HandleFunc(options.BaseURL+"/api/auth/callback", wrapper.GetAPIAuthCallback).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/calendar/me", wrapper.GetAPICalendarMe).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/events", wrapper.RenderEvent).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/healthcheck", wrapper.GetAPIHealthcheck).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/notifications/{notificationID}/read", wrapper.OptionsAPINotificationsNotificationIDRead).Methods("OPTIONS")

	r.HandleFunc(options.BaseURL+"/api/notifications/{notificationID}/read", wrapper.PatchAPINotificationsNotificationIDRead).Methods("PATCH")

	r.HandleFunc(options.BaseURL+"/api/refresh", wrapper.PostAPIRefresh).Methods("POST")

	r.HandleFunc(options.BaseURL+"/api/teams", wrapper.GetAPITeams).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/teams", wrapper.OptionsAPITeams).Methods("OPTIONS")

	r.HandleFunc(options.BaseURL+"/api/teams", wrapper.PostAPITeams).Methods("POST")

	r.HandleFunc(options.BaseURL+"/api/teams/joinable/me", wrapper.GetAPITeamsJoinableMe).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/teams/me", wrapper.GetAPITeamsMe).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/teams/{teamID}", wrapper.DeleteAPITeamsTeamID).Methods("DELETE")

	r.HandleFunc(options.BaseURL+"/api/teams/{teamID}", wrapper.GetAPITeamsTeamID).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/teams/{teamID}/users", wrapper.GetAPITeamsTeamIDUsers).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/teams/{teamID}/users/me", wrapper.PostAPITeamsTeamIDUsersMe).Methods("POST")

	r.HandleFunc(options.BaseURL+"/api/teams/{teamID}/users/{userID}", wrapper.PostAPITeamsTeamIDUsersUserID).Methods("POST")

	r.HandleFunc(options.BaseURL+"/api/users", wrapper.GetAPIUsers).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/users", wrapper.PostAPIUsers).Methods("POST")

	r.HandleFunc(options.BaseURL+"/api/users/me", wrapper.GetAPIUsersMe).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/users/me/logout", wrapper.PostAPIUsersMeLogout).Methods("POST")

	r.HandleFunc(options.BaseURL+"/api/users/me/notifications", wrapper.GetAPIUsersMeNotifications).Methods("GET")

	r.HandleFunc(options.BaseURL+"/api/users/{userID}", wrapper.DeleteAPIUsersUserID).Methods("DELETE")

	r.HandleFunc(options.BaseURL+"/api/users/{userID}", wrapper.GetAPIUsersUserID).Methods("GET")

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}

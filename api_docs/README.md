# Documentation for Slotify API

<a name="documentation-for-api-endpoints"></a>
## Documentation for API Endpoints

All URIs are relative to *http://localhost*

| Class | Method | HTTP request | Description |
|------------ | ------------- | ------------- | -------------|
| *DefaultApi* | [**deleteAPITeamsTeamID**](Apis/DefaultApi.md#deleteapiteamsteamid) | **DELETE** /api/teams/{teamID} | Delete a team by id |
*DefaultApi* | [**deleteAPIUsersUserID**](Apis/DefaultApi.md#deleteapiusersuserid) | **DELETE** /api/users/{userID} | Delete a user by id |
*DefaultApi* | [**getAPIAuthCallback**](Apis/DefaultApi.md#getapiauthcallback) | **GET** /api/auth/callback | Auth route for authorisation code flow |
*DefaultApi* | [**getAPICalendarMe**](Apis/DefaultApi.md#getapicalendarme) | **GET** /api/calendar/me | get a user's calendar events |
*DefaultApi* | [**getAPIHealthcheck**](Apis/DefaultApi.md#getapihealthcheck) | **GET** /api/healthcheck | Healthcheck route |
*DefaultApi* | [**getAPITeams**](Apis/DefaultApi.md#getapiteams) | **GET** /api/teams | Get a team by query params |
*DefaultApi* | [**getAPITeamsJoinableMe**](Apis/DefaultApi.md#getapiteamsjoinableme) | **GET** /api/teams/joinable/me | Get all joinable teams for a user excluding teams they are already a part of |
*DefaultApi* | [**getAPITeamsMe**](Apis/DefaultApi.md#getapiteamsme) | **GET** /api/teams/me | Get all teams for user by id passed by JWT |
*DefaultApi* | [**getAPITeamsTeamID**](Apis/DefaultApi.md#getapiteamsteamid) | **GET** /api/teams/{teamID} | Get a team by id |
*DefaultApi* | [**getAPITeamsTeamIDUsers**](Apis/DefaultApi.md#getapiteamsteamidusers) | **GET** /api/teams/{teamID}/users | Get all members of a team |
*DefaultApi* | [**getAPIUsers**](Apis/DefaultApi.md#getapiusers) | **GET** /api/users | Get a user by query params |
*DefaultApi* | [**getAPIUsersMe**](Apis/DefaultApi.md#getapiusersme) | **GET** /api/users/me | Get the user by id passed by JWT |
*DefaultApi* | [**getAPIUsersMeNotifications**](Apis/DefaultApi.md#getapiusersmenotifications) | **GET** /api/users/me/notifications | get user's unread notifications |
*DefaultApi* | [**getAPIUsersUserID**](Apis/DefaultApi.md#getapiusersuserid) | **GET** /api/users/{userID} | Get a user by id |
*DefaultApi* | [**optionsAPINotificationsNotificationIDRead**](Apis/DefaultApi.md#optionsapinotificationsnotificationidread) | **OPTIONS** /api/notifications/{notificationID}/read | CORS preflight for marking a notification as read |
*DefaultApi* | [**optionsAPITeams**](Apis/DefaultApi.md#optionsapiteams) | **OPTIONS** /api/teams | CORS preflight for teams |
*DefaultApi* | [**patchAPINotificationsNotificationIDRead**](Apis/DefaultApi.md#patchapinotificationsnotificationidread) | **PATCH** /api/notifications/{notificationID}/read | mark a notification as being read |
*DefaultApi* | [**postAPIRefresh**](Apis/DefaultApi.md#postapirefresh) | **POST** /api/refresh | Refresh Slotify access token and refresh token |
*DefaultApi* | [**postAPITeams**](Apis/DefaultApi.md#postapiteams) | **POST** /api/teams | Create a new team |
*DefaultApi* | [**postAPITeamsTeamIDUsersMe**](Apis/DefaultApi.md#postapiteamsteamidusersme) | **POST** /api/teams/{teamID}/users/me | Add current user to a team |
*DefaultApi* | [**postAPITeamsTeamIDUsersUserID**](Apis/DefaultApi.md#postapiteamsteamidusersuserid) | **POST** /api/teams/{teamID}/users/{userID} | Add a user to a team |
*DefaultApi* | [**postAPIUsers**](Apis/DefaultApi.md#postapiusers) | **POST** /api/users | Create a new user |
*DefaultApi* | [**postAPIUsersMeLogout**](Apis/DefaultApi.md#postapiusersmelogout) | **POST** /api/users/me/logout | Logout user |
*DefaultApi* | [**renderEvent**](Apis/DefaultApi.md#renderevent) | **GET** /api/events | Subscribe to notifications |


<a name="documentation-for-models"></a>
## Documentation for Models

 - [Attendee](./Models/Attendee.md)
 - [CalendarEvent](./Models/CalendarEvent.md)
 - [Location](./Models/Location.md)
 - [Notification](./Models/Notification.md)
 - [Team](./Models/Team.md)
 - [TeamCreate](./Models/TeamCreate.md)
 - [User](./Models/User.md)
 - [UserCreate](./Models/UserCreate.md)
 - [renderEvent_200_response](./Models/renderEvent_200_response.md)


<a name="documentation-for-authorization"></a>
## Documentation for Authorization

All endpoints do not require authorization.

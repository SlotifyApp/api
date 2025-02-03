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
*DefaultApi* | [**getAPITeamsMe**](Apis/DefaultApi.md#getapiteamsme) | **GET** /api/teams/me | Get all teams for user by id passed by JWT |
*DefaultApi* | [**getAPITeamsTeamID**](Apis/DefaultApi.md#getapiteamsteamid) | **GET** /api/teams/{teamID} | Get a team by id |
*DefaultApi* | [**getAPITeamsTeamIDUsers**](Apis/DefaultApi.md#getapiteamsteamidusers) | **GET** /api/teams/{teamID}/users | Get all members of a team |
*DefaultApi* | [**getAPIUsers**](Apis/DefaultApi.md#getapiusers) | **GET** /api/users | Get a user by query params |
*DefaultApi* | [**getAPIUsersMe**](Apis/DefaultApi.md#getapiusersme) | **GET** /api/users/me | Get the user by id passed by JWT |
*DefaultApi* | [**getAPIUsersUserID**](Apis/DefaultApi.md#getapiusersuserid) | **GET** /api/users/{userID} | Get a user by id |
*DefaultApi* | [**postAPIRefresh**](Apis/DefaultApi.md#postapirefresh) | **POST** /api/refresh | Refresh Slotify access token and refresh token |
*DefaultApi* | [**postAPITeams**](Apis/DefaultApi.md#postapiteams) | **POST** /api/teams | Create a new team |
*DefaultApi* | [**postAPITeamsTeamIDUsersUserID**](Apis/DefaultApi.md#postapiteamsteamidusersuserid) | **POST** /api/teams/{teamID}/users/{userID} | Add a user to a team |
*DefaultApi* | [**postAPIUsers**](Apis/DefaultApi.md#postapiusers) | **POST** /api/users | Create a new user |
*DefaultApi* | [**postAPIUsersMeLogout**](Apis/DefaultApi.md#postapiusersmelogout) | **POST** /api/users/me/logout | Logout user |


<a name="documentation-for-models"></a>
## Documentation for Models

 - [CalendarEvent](./Models/CalendarEvent.md)
 - [Team](./Models/Team.md)
 - [TeamCreate](./Models/TeamCreate.md)
 - [User](./Models/User.md)
 - [UserCreate](./Models/UserCreate.md)


<a name="documentation-for-authorization"></a>
## Documentation for Authorization

<a name="bearerAuth"></a>
### bearerAuth

- **Type**: HTTP Bearer Token authentication (JWT)


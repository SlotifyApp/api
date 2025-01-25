# Documentation for Slotify Backend API

<a name="documentation-for-api-endpoints"></a>
## Documentation for API Endpoints

All URIs are relative to *http://localhost*

| Class | Method | HTTP request | Description |
|------------ | ------------- | ------------- | -------------|
| *DefaultApi* | [**deleteTeamsTeamID**](Apis/DefaultApi.md#deleteteamsteamid) | **DELETE** /teams/{teamID} | Delete a team by id |
*DefaultApi* | [**deleteUsersUserID**](Apis/DefaultApi.md#deleteusersuserid) | **DELETE** /users/{userID} | Delete a user by id |
*DefaultApi* | [**getTeamsTeamID**](Apis/DefaultApi.md#getteamsteamid) | **GET** /teams/{teamID} | Get a team by id |
*DefaultApi* | [**getTeamsTeamIDUsers**](Apis/DefaultApi.md#getteamsteamidusers) | **GET** /teams/{teamID}/users | Get all members of a team |
*DefaultApi* | [**getUsersUserID**](Apis/DefaultApi.md#getusersuserid) | **GET** /users/{userID} | Get a user by id |
*DefaultApi* | [**healthcheckGet**](Apis/DefaultApi.md#healthcheckget) | **GET** /healthcheck | Healthcheck route |
*DefaultApi* | [**postTeamsTeamIDUsersUserID**](Apis/DefaultApi.md#postteamsteamidusersuserid) | **POST** /teams/{teamID}/users/{userID} | Add a user to a team |
*DefaultApi* | [**teamsGet**](Apis/DefaultApi.md#teamsget) | **GET** /teams | Get a team by query params |
*DefaultApi* | [**teamsPost**](Apis/DefaultApi.md#teamspost) | **POST** /teams | Create a new team |
*DefaultApi* | [**usersGet**](Apis/DefaultApi.md#usersget) | **GET** /users | Get a user by query params |
*DefaultApi* | [**usersPost**](Apis/DefaultApi.md#userspost) | **POST** /users | Create a new user |


<a name="documentation-for-models"></a>
## Documentation for Models

 - [Team](./Models/Team.md)
 - [TeamCreate](./Models/TeamCreate.md)
 - [User](./Models/User.md)
 - [UserCreate](./Models/UserCreate.md)


<a name="documentation-for-authorization"></a>
## Documentation for Authorization

All endpoints do not require authorization.

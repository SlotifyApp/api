# DefaultApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**deleteAPITeamsTeamID**](DefaultApi.md#deleteAPITeamsTeamID) | **DELETE** /api/teams/{teamID} | Delete a team by id |
| [**deleteAPIUsersUserID**](DefaultApi.md#deleteAPIUsersUserID) | **DELETE** /api/users/{userID} | Delete a user by id |
| [**getAPIAuthCallback**](DefaultApi.md#getAPIAuthCallback) | **GET** /api/auth/callback | Auth route for authorisation code flow |
| [**getAPICalendarMe**](DefaultApi.md#getAPICalendarMe) | **GET** /api/calendar/me | get a user&#39;s calendar events |
| [**getAPIHealthcheck**](DefaultApi.md#getAPIHealthcheck) | **GET** /api/healthcheck | Healthcheck route |
| [**getAPITeams**](DefaultApi.md#getAPITeams) | **GET** /api/teams | Get a team by query params |
| [**getAPITeamsMe**](DefaultApi.md#getAPITeamsMe) | **GET** /api/teams/me | Get all teams for user by id passed by JWT |
| [**getAPITeamsTeamID**](DefaultApi.md#getAPITeamsTeamID) | **GET** /api/teams/{teamID} | Get a team by id |
| [**getAPITeamsTeamIDUsers**](DefaultApi.md#getAPITeamsTeamIDUsers) | **GET** /api/teams/{teamID}/users | Get all members of a team |
| [**getAPIUsers**](DefaultApi.md#getAPIUsers) | **GET** /api/users | Get a user by query params |
| [**getAPIUsersMe**](DefaultApi.md#getAPIUsersMe) | **GET** /api/users/me | Get the user by id passed by JWT |
| [**getAPIUsersUserID**](DefaultApi.md#getAPIUsersUserID) | **GET** /api/users/{userID} | Get a user by id |
| [**postAPIRefresh**](DefaultApi.md#postAPIRefresh) | **POST** /api/refresh | Refresh Slotify access token and refresh token |
| [**postAPITeams**](DefaultApi.md#postAPITeams) | **POST** /api/teams | Create a new team |
| [**postAPITeamsTeamIDUsersUserID**](DefaultApi.md#postAPITeamsTeamIDUsersUserID) | **POST** /api/teams/{teamID}/users/{userID} | Add a user to a team |
| [**postAPIUsers**](DefaultApi.md#postAPIUsers) | **POST** /api/users | Create a new user |
| [**postAPIUsersMeLogout**](DefaultApi.md#postAPIUsersMeLogout) | **POST** /api/users/me/logout | Logout user |


<a name="deleteAPITeamsTeamID"></a>
# **deleteAPITeamsTeamID**
> String deleteAPITeamsTeamID(teamID)

Delete a team by id

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **teamID** | **Integer**| Numeric ID of the team to delete | [default to null] |

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="deleteAPIUsersUserID"></a>
# **deleteAPIUsersUserID**
> String deleteAPIUsersUserID(userID)

Delete a user by id

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **userID** | **Integer**| Numeric ID of the user to delete | [default to null] |

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPIAuthCallback"></a>
# **getAPIAuthCallback**
> getAPIAuthCallback(code, state)

Auth route for authorisation code flow

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **code** | **String**|  | [default to null] |
| **state** | **String**|  | [default to null] |

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

<a name="getAPICalendarMe"></a>
# **getAPICalendarMe**
> List getAPICalendarMe()

get a user&#39;s calendar events

### Parameters
This endpoint does not need any parameter.

### Return type

[**List**](../Models/CalendarEvent.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPIHealthcheck"></a>
# **getAPIHealthcheck**
> String getAPIHealthcheck()

Healthcheck route

### Parameters
This endpoint does not need any parameter.

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPITeams"></a>
# **getAPITeams**
> List getAPITeams(name)

Get a team by query params

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **name** | **String**| Team name | [optional] [default to null] |

### Return type

[**List**](../Models/Team.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPITeamsMe"></a>
# **getAPITeamsMe**
> List getAPITeamsMe()

Get all teams for user by id passed by JWT

### Parameters
This endpoint does not need any parameter.

### Return type

[**List**](../Models/Team.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPITeamsTeamID"></a>
# **getAPITeamsTeamID**
> Team getAPITeamsTeamID(teamID)

Get a team by id

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **teamID** | **Integer**| Numeric ID of the team to get | [default to null] |

### Return type

[**Team**](../Models/Team.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPITeamsTeamIDUsers"></a>
# **getAPITeamsTeamIDUsers**
> List getAPITeamsTeamIDUsers(teamID)

Get all members of a team

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **teamID** | **Integer**| ID of the team | [default to null] |

### Return type

[**List**](../Models/User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPIUsers"></a>
# **getAPIUsers**
> List getAPIUsers(email, firstName, lastName)

Get a user by query params

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **email** | **String**| Email of user | [optional] [default to null] |
| **firstName** | **String**| First name of user | [optional] [default to null] |
| **lastName** | **String**| Last name of user | [optional] [default to null] |

### Return type

[**List**](../Models/User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPIUsersMe"></a>
# **getAPIUsersMe**
> User getAPIUsersMe()

Get the user by id passed by JWT

### Parameters
This endpoint does not need any parameter.

### Return type

[**User**](../Models/User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="getAPIUsersUserID"></a>
# **getAPIUsersUserID**
> User getAPIUsersUserID(userID)

Get a user by id

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **userID** | **Integer**| Numeric ID of the user to get | [default to null] |

### Return type

[**User**](../Models/User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="postAPIRefresh"></a>
# **postAPIRefresh**
> String postAPIRefresh()

Refresh Slotify access token and refresh token

### Parameters
This endpoint does not need any parameter.

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="postAPITeams"></a>
# **postAPITeams**
> Team postAPITeams(TeamCreate)

Create a new team

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **TeamCreate** | [**TeamCreate**](../Models/TeamCreate.md)|  | |

### Return type

[**Team**](../Models/Team.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="postAPITeamsTeamIDUsersUserID"></a>
# **postAPITeamsTeamIDUsersUserID**
> String postAPITeamsTeamIDUsersUserID(userID, teamID)

Add a user to a team

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **userID** | **Integer**| ID of the user | [default to null] |
| **teamID** | **Integer**| ID of the team | [default to null] |

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="postAPIUsers"></a>
# **postAPIUsers**
> User postAPIUsers(UserCreate)

Create a new user

### Parameters

|Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **UserCreate** | [**UserCreate**](../Models/UserCreate.md)|  | |

### Return type

[**User**](../Models/User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="postAPIUsersMeLogout"></a>
# **postAPIUsersMeLogout**
> String postAPIUsersMeLogout()

Logout user

### Parameters
This endpoint does not need any parameter.

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


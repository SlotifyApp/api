# DefaultApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**deleteTeamsTeamID**](DefaultApi.md#deleteTeamsTeamID) | **DELETE** /teams/{teamID} | Delete a team by id |
| [**deleteUsersUserID**](DefaultApi.md#deleteUsersUserID) | **DELETE** /users/{userID} | Delete a user by id |
| [**getAPIAuthCallback**](DefaultApi.md#getAPIAuthCallback) | **GET** /api/auth/callback | Auth route for authorisation code flow |
| [**getTeamsTeamID**](DefaultApi.md#getTeamsTeamID) | **GET** /teams/{teamID} | Get a team by id |
| [**getTeamsTeamIDUsers**](DefaultApi.md#getTeamsTeamIDUsers) | **GET** /teams/{teamID}/users | Get all members of a team |
| [**getUsersUserID**](DefaultApi.md#getUsersUserID) | **GET** /users/{userID} | Get a user by id |
| [**healthcheckGet**](DefaultApi.md#healthcheckGet) | **GET** /healthcheck | Healthcheck route |
| [**postTeamsTeamIDUsersUserID**](DefaultApi.md#postTeamsTeamIDUsersUserID) | **POST** /teams/{teamID}/users/{userID} | Add a user to a team |
| [**refreshPost**](DefaultApi.md#refreshPost) | **POST** /refresh | Create new Slotify access token and refresh token |
| [**teamsGet**](DefaultApi.md#teamsGet) | **GET** /teams | Get a team by query params |
| [**teamsPost**](DefaultApi.md#teamsPost) | **POST** /teams | Create a new team |
| [**userGet**](DefaultApi.md#userGet) | **GET** /user | Get a user by id passed by JWT |
| [**userLogoutPost**](DefaultApi.md#userLogoutPost) | **POST** /user/logout | Logout user |
| [**usersGet**](DefaultApi.md#usersGet) | **GET** /users | Get a user by query params |
| [**usersPost**](DefaultApi.md#usersPost) | **POST** /users | Create a new user |


<a name="deleteTeamsTeamID"></a>
# **deleteTeamsTeamID**
> String deleteTeamsTeamID(teamID)

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

<a name="deleteUsersUserID"></a>
# **deleteUsersUserID**
> String deleteUsersUserID(userID)

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

<a name="getTeamsTeamID"></a>
# **getTeamsTeamID**
> Team getTeamsTeamID(teamID)

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

<a name="getTeamsTeamIDUsers"></a>
# **getTeamsTeamIDUsers**
> List getTeamsTeamIDUsers(teamID)

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

<a name="getUsersUserID"></a>
# **getUsersUserID**
> User getUsersUserID(userID)

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

<a name="healthcheckGet"></a>
# **healthcheckGet**
> String healthcheckGet()

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

<a name="postTeamsTeamIDUsersUserID"></a>
# **postTeamsTeamIDUsersUserID**
> String postTeamsTeamIDUsersUserID(userID, teamID)

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

<a name="refreshPost"></a>
# **refreshPost**
> String refreshPost()

Create new Slotify access token and refresh token

### Parameters
This endpoint does not need any parameter.

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="teamsGet"></a>
# **teamsGet**
> List teamsGet(name)

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

<a name="teamsPost"></a>
# **teamsPost**
> Team teamsPost(TeamCreate)

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

<a name="userGet"></a>
# **userGet**
> User userGet()

Get a user by id passed by JWT

### Parameters
This endpoint does not need any parameter.

### Return type

[**User**](../Models/User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="userLogoutPost"></a>
# **userLogoutPost**
> String userLogoutPost()

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

<a name="usersGet"></a>
# **usersGet**
> List usersGet(email, firstName, lastName)

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

<a name="usersPost"></a>
# **usersPost**
> User usersPost(UserCreate)

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


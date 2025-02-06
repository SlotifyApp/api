-- name: CreateUser :execlastid
INSERT INTO User (email, first_name, last_name) VALUES (?, ?, ?);

-- name: CountUserByID :one
SELECT COUNT(*) FROM User WHERE id=?;

-- name: CountUserByEmail :one
SELECT COUNT(*) FROM User WHERE email=?;

-- name: DeleteUserByID :execrows
DELETE FROM User WHERE id=?;

-- name: GetUserByID :one
SELECT * FROM User WHERE id=?;

-- name: GetUserByEmail :one
SELECT * FROM User WHERE email=?;

-- name: UpdateUserHomeAccountID :execrows
UPDATE User SET msft_home_account_id=? WHERE id=?;

-- name: GetUsersTeams :many
SELECT t.* FROM UserToTeam utt
JOIN Team t ON utt.team_id=t.id 
WHERE utt.user_id=?;

-- name: ListUsers :many
SELECT id, email, first_name, last_name FROM User
WHERE email = ifnull(sqlc.arg('email'), email)
   AND first_name = ifnull(sqlc.arg('firstName'), first_name)
   AND last_name = ifnull(sqlc.arg('lastName'), last_name);




-- name: AddUserToTeam :execrows
INSERT INTO UserToTeam (user_id, team_id) VALUES (?, ?);

-- name: CountTeamByID :one
SELECT COUNT(*) FROM Team WHERE id=?;

-- name: GetAllTeamMembers :many
SELECT u.id, u.email, u.first_name, u.last_name FROM Team t
JOIN UserToTeam utt ON t.id=utt.team_id
JOIN User u ON u.id=utt.user_id 
WHERE t.id=?;

-- name: GetJoinableTeams :many
SELECT t.* FROM Team t
LEFT JOIN UserToTeam utt ON
     t.id = utt.team_id AND utt.user_id = ? 
WHERE utt.user_id IS NULL;

-- name: GetTeamByID :one
SELECT * FROM Team WHERE id=?;

-- name: DeleteTeamByID :execrows
DELETE FROM Team WHERE id=?;

-- name: ListTeams :many
SELECT * FROM Team
WHERE name = ifnull(sqlc.arg('name'), name);

-- name: AddTeam :execlastid
INSERT INTO Team (name) VALUES (?);




-- name: CreateRefreshToken :execrows
REPLACE INTO RefreshToken (user_id, token) VALUES (?, ?);

-- name: GetRefreshTokenByUserID :one
SELECT * FROM RefreshToken WHERE user_id=?;

-- name: DeleteRefreshTokenByUserID :execrows
DELETE FROM RefreshToken WHERE user_id=?;

-- name: CreateNotification :execlastid
INSERT INTO Notification (message, created) VALUES(?, ?);

-- name: CreateUserNotification :execrows
INSERT INTO UserToNotification (user_id, notification_id, is_read) VALUES(?, ?, FALSE);

-- name: GetUnreadUserNotifications :many
SELECT n.* FROM UserToNotification utn
JOIN Notification n ON n.id=utn.notification_id 
WHERE utn.user_id=? AND utn.is_read=FALSE
ORDER BY n.created DESC;

-- name: MarkNotificationAsRead :execrows
UPDATE UserToNotification SET is_read=TRUE
WHERE user_id=? AND notification_id=?;

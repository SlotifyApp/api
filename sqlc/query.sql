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

-- name: GetUsersSlotifyGroups :many
SELECT sg.* FROM UserToSlotifyGroup utsg
JOIN SlotifyGroup sg ON utsg.slotify_group_id=sg.id 
WHERE utsg.user_id=?;

-- name: SearchUsersByName :many
SELECT id, email, first_name, last_name FROM User
WHERE LOWER(CONCAT(first_name, ' ', last_name)) LIKE LOWER(CONCAT('%', sqlc.arg('name'), '%'))
LIMIT ?;

-- name: SearchUsersByEmail :many
SELECT id, email, first_name, last_name FROM User
WHERE LOWER(email) LIKE LOWER(CONCAT('%', sqlc.arg('email'), '%'))
LIMIT ?;

-- name: ListUsers :many
SELECT id, email, first_name, last_name FROM User;





-- name: AddUserToSlotifyGroup :execrows
INSERT INTO UserToSlotifyGroup (user_id, slotify_group_id) VALUES (?, ?);

-- name: CountSlotifyGroupByID :one
SELECT COUNT(*) FROM SlotifyGroup WHERE id=?;

-- name: GetAllSlotifyGroupMembers :many
SELECT u.id, u.email, u.first_name, u.last_name FROM SlotifyGroup sg
JOIN UserToSlotifyGroup utsg ON sg.id=utsg.slotify_group_id
JOIN User u ON u.id=utsg.user_id 
WHERE sg.id=?;

-- name: CountSlotifyGroupMembers :one
SELECT COUNT(*) FROM SlotifyGroup sg
JOIN UserToSlotifyGroup utsg ON sg.id=utsg.slotify_group_id
JOIN User u ON u.id=utsg.user_id 
WHERE sg.id=?;

-- name: GetAllSlotifyGroupMembersExcept :many
SELECT u.id FROM SlotifyGroup sg
JOIN UserToSlotifyGroup utsg ON sg.id=utsg.slotify_group_id
JOIN User u ON u.id=utsg.user_id 
WHERE sg.id=sqlc.arg('slotifyGroupID') AND u.id!=sqlc.arg('userID');

-- name: GetSlotifyGroupByID :one
SELECT * FROM SlotifyGroup WHERE id=?;

-- name: DeleteSlotifyGroupByID :execrows
DELETE FROM SlotifyGroup WHERE id=?;

-- name: ListSlotifyGroups :many
SELECT * FROM SlotifyGroup
WHERE name = ifnull(sqlc.arg('name'), name);

-- name: AddSlotifyGroup :execlastid
INSERT INTO SlotifyGroup (name) VALUES (?);

-- name: RemoveSlotifyGroupMember :execrows
DELETE FROM UserToSlotifyGroup
WHERE user_id=? AND slotify_group_id=?;

-- name: RemoveSlotifyGroup :execrows
DELETE FROM SlotifyGroup
WHERE id=?;

-- name: CheckMemberInSlotifyGroup :one
SELECT COUNT(*) FROM UserToSlotifyGroup
WHERE user_id=? AND slotify_group_id=?;




-- name: CreateRefreshToken :execrows
REPLACE INTO RefreshToken (user_id, token) VALUES (?, ?);

-- name: GetRefreshTokenByUserID :one
SELECT * FROM RefreshToken WHERE user_id=?;

-- name: DeleteRefreshTokenByUserID :execrows
DELETE FROM RefreshToken WHERE user_id=?;

-- name: CreateNotification :execlastid
INSERT INTO Notification (message, created) VALUES(?, ?);

-- name: CountWeekOldNotifications :one
SELECT COUNT(*) FROM Notification
WHERE DATE(created) <= CURDATE() - INTERVAL 1 WEEK;

-- name: BatchDeleteWeekOldNotifications :execrows
DELETE FROM Notification
WHERE created <= CURDATE() - INTERVAL 1 WEEK
  AND id >= (SELECT MIN(id) FROM Notification WHERE created <= CURDATE() - INTERVAL 1 WEEK)
ORDER BY id
LIMIT ?;

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




-- name: GetInviteByID :one
SELECT * FROM Invite
WHERE id=?;

-- name: CreateInvite :execlastid
INSERT INTO Invite (slotify_group_id, from_user_id, to_user_id, message, status, expiry_date, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?);

-- name: UpdateInviteStatus :execrows
UPDATE Invite SET status=?
WHERE id=?;

-- name: DeleteInviteByID :execrows
DELETE FROM Invite WHERE id=?;

-- name: UpdateInviteMessage :execrows
UPDATE Invite SET message=?
WHERE id=? AND from_user_id=?;

-- name: ListInvitesMe :many
SELECT i.id AS invite_id, i.message, i.status,i.created_at, i.expiry_date, fu.email AS from_user_email, fu.first_name AS from_user_first_name, fu.last_name AS from_user_last_name, sg.name AS slotify_group_name FROM Invite i
JOIN User fu ON fu.id=i.from_user_id
JOIN SlotifyGroup sg ON sg.id=i.slotify_group_id
WHERE i.status = ifnull(sqlc.arg('status'), i.status) 
AND i.to_user_id=?;

-- name: ListInvitesByGroup :many
SELECT 
   i.id AS invite_id, i.message, i.status, i.created_at, i.expiry_date, fu.email AS from_user_email, fu.first_name AS from_user_first_name, fu.last_name AS from_user_last_name, tu.email AS to_user_email, tu.first_name AS to_user_first_name, tu.last_name AS to_user_last_name FROM Invite i
JOIN User fu ON fu.id=i.from_user_id
JOIN User tu ON tu.id=i.to_user_id
WHERE i.status = ifnull(sqlc.arg('status'), i.status) 
AND i.slotify_group_id=?;

-- name: BatchDeleteWeekOldInvites :execrows
DELETE FROM Invite
WHERE created_at <= CURDATE() - INTERVAL 1 WEEK
  AND id >= (SELECT MIN(id) FROM Invite WHERE DATE(created_at) <= CURDATE() - INTERVAL 1 WEEK)
ORDER BY id
LIMIT ?;

-- name: CountExpiredInvites :one
SELECT COUNT(*) FROM Invite
WHERE status='expired';

-- name: BatchExpireInvites :execrows
UPDATE Invite SET status = 'expired'
WHERE expiry_date <= CURDATE()
  AND status != 'expired'
  AND id >= (SELECT MIN(id) FROM Invite WHERE expiry_date <= CURDATE() AND status != 'expired')
ORDER BY id
LIMIT ?;

-- name: CountWeekOldInvites :one
SELECT COUNT(*) FROM Invite
WHERE DATE(created_at) <= CURDATE() - INTERVAL 1 WEEK;




-- name: GetMeetingByID :one
SELECT * FROM Meeting
WHERE id=?;

-- name: GetMeetingByMSFTID :one
SELECT * FROM Meeting
WHERE msft_meeting_id=?;

-- name: GetMeetingPreferences :one
SELECT * FROM MeetingPreferences
WHERE id=?;

-- name: CreateMeetingPreferences :execlastid
INSERT INTO MeetingPreferences (meeting_start_time, start_date_range, end_date_range) VALUES (?,?,?);

-- name: CreateMeeting :execlastid
INSERT INTO Meeting (meeting_pref_id, owner_email, msft_meeting_id) VALUES (?,?,?);

-- name: CreateReschedulingRequest :execlastid
INSERT INTO ReschedulingRequest (requested_by) VALUES (?);

-- name: CreatePlaceholderMeeting :execlastid
INSERT INTO PlaceholderMeeting (request_id, title, start_time, end_time, location, duration, start_date_range, end_date_range) VALUES (?,?,?,?,?,?,?,?);

-- name: CreatePlaceholderMeetingAttendee :execlastid
INSERT INTO PlaceholderMeetingAttendee (meeting_id, user_id) VALUES (?,?);

-- name: CreateRequestToMeeting :execlastid
INSERT INTO RequestToMeeting (request_id, meeting_id) VALUES (?,?);

-- name: GetAllRequestsForUser :many
SELECT rr.*, m.msft_meeting_id, m.id, pm.meeting_id, pm.title, pm.start_time, pm.end_time, pm.duration, pm.location  
FROM ReschedulingRequest rr 
JOIN RequestToMeeting rtm ON rr.request_id = rtm.request_id 
JOIN Meeting m ON rtm.meeting_id = m.id 
LEFT JOIN PlaceholderMeeting pm ON rr.request_id = pm.request_id 
WHERE m.owner_email = ?;

-- name: GetRequestByID :one
SELECT rr.*, m.msft_meeting_id, m.id, pm.meeting_id, pm.title, pm.start_time, pm.end_time, pm.duration, pm.location 
FROM ReschedulingRequest rr 
JOIN RequestToMeeting rtm ON rr.request_id = rtm.request_id 
JOIN Meeting m ON rtm.meeting_id = m.id 
LEFT JOIN PlaceholderMeeting pm ON rr.request_id = pm.request_id 
WHERE rr.request_id = ?;

-- name: UpdateRequestStatusAsRejected :execrows
UPDATE ReschedulingRequest rr SET status = 'declined' WHERE rr.request_id IN (
  SELECT request_id FROM RequestToMeeting WHERE meeting_id = ?
);

-- name: UpdateRequestStatusAsAccepted :execrows
UPDATE ReschedulingRequest rr SET status = 'accepted' WHERE rr.request_id IN (
  SELECT request_id FROM RequestToMeeting WHERE meeting_id = ?
);

-- name: UpdateMeetingStartTime :execlastid
UPDATE MeetingPreferences mp SET mp.meeting_start_time=?
WHERE mp.id IN (
  SELECT m.meeting_pref_id FROM Meeting m WHERE m.id = ?
);
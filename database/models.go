// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusDeclined InviteStatus = "declined"
	InviteStatusExpired  InviteStatus = "expired"
)

func (e *InviteStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = InviteStatus(s)
	case string:
		*e = InviteStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for InviteStatus: %T", src)
	}
	return nil
}

type NullInviteStatus struct {
	InviteStatus InviteStatus `json:"inviteStatus"`
	Valid        bool         `json:"valid"` // Valid is true if InviteStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullInviteStatus) Scan(value interface{}) error {
	if value == nil {
		ns.InviteStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.InviteStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullInviteStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.InviteStatus), nil
}

type ReschedulingrequestStatus string

const (
	ReschedulingrequestStatusPending  ReschedulingrequestStatus = "pending"
	ReschedulingrequestStatusAccepted ReschedulingrequestStatus = "accepted"
	ReschedulingrequestStatusDeclined ReschedulingrequestStatus = "declined"
)

func (e *ReschedulingrequestStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ReschedulingrequestStatus(s)
	case string:
		*e = ReschedulingrequestStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ReschedulingrequestStatus: %T", src)
	}
	return nil
}

type NullReschedulingrequestStatus struct {
	ReschedulingrequestStatus ReschedulingrequestStatus `json:"reschedulingrequestStatus"`
	Valid                     bool                      `json:"valid"` // Valid is true if ReschedulingrequestStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullReschedulingrequestStatus) Scan(value interface{}) error {
	if value == nil {
		ns.ReschedulingrequestStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ReschedulingrequestStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullReschedulingrequestStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ReschedulingrequestStatus), nil
}

type Invite struct {
	ID             uint32       `json:"id"`
	SlotifyGroupID uint32       `json:"slotifyGroupID"`
	FromUserID     uint32       `json:"fromUserID"`
	ToUserID       uint32       `json:"toUserID"`
	Message        string       `json:"message"`
	Status         InviteStatus `json:"status"`
	ExpiryDate     time.Time    `json:"expiryDate"`
	CreatedAt      time.Time    `json:"createdAt"`
}

type Meeting struct {
	ID            uint32 `json:"id"`
	MeetingPrefID uint32 `json:"meetingPrefID"`
	OwnerEmail    string `json:"ownerEmail"`
	MsftMeetingID string `json:"msftMeetingID"`
}

type Meetingpreferences struct {
	ID               uint32    `json:"id"`
	MeetingStartTime time.Time `json:"meetingStartTime"`
	StartDateRange   time.Time `json:"startDateRange"`
	EndDateRange     time.Time `json:"endDateRange"`
}

type Notification struct {
	ID      uint32    `json:"id"`
	Message string    `json:"message"`
	Created time.Time `json:"created"`
}

type Placeholdermeeting struct {
	MeetingID      uint32    `json:"meetingID"`
	RequestID      uint32    `json:"requestID"`
	Title          string    `json:"title"`
	Location       string    `json:"location"`
	Duration       time.Time `json:"duration"`
	StartDateRange time.Time `json:"startDateRange"`
	EndDateRange   time.Time `json:"endDateRange"`
}

type Placeholdermeetingattendee struct {
	MeetingID uint32 `json:"meetingID"`
	UserID    uint32 `json:"userID"`
}

type RefreshToken struct {
	ID      uint32 `json:"id"`
	UserID  uint32 `json:"userID"`
	Token   string `json:"token"`
	Revoked bool   `json:"revoked"`
}

type Requesttomeeting struct {
	RequestID uint32 `json:"requestID"`
	MeetingID uint32 `json:"meetingID"`
}

type Reschedulingrequest struct {
	RequestID   uint32                    `json:"requestID"`
	RequestedBy uint32                    `json:"requestedBy"`
	Status      ReschedulingrequestStatus `json:"status"`
	CreatedAt   time.Time                 `json:"createdAt"`
}

type SlotifyGroup struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID                uint32         `json:"id"`
	Email             string         `json:"email"`
	FirstName         string         `json:"firstName"`
	LastName          string         `json:"lastName"`
	MsftHomeAccountID sql.NullString `json:"msftHomeAccountID"`
}

type Userpreferences struct {
	UserID         uint32    `json:"userID"`
	LunchStartTime time.Time `json:"lunchStartTime"`
	LunchEndTime   time.Time `json:"lunchEndTime"`
}

type Usertonotification struct {
	UserID         uint32 `json:"userID"`
	NotificationID uint32 `json:"notificationID"`
	IsRead         bool   `json:"isRead"`
}

type Usertoslotifygroup struct {
	UserID         uint32 `json:"userID"`
	SlotifyGroupID uint32 `json:"slotifyGroupID"`
}

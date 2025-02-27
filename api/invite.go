package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/logger"
	"github.com/SlotifyApp/slotify-backend/notification"
	"github.com/avast/retry-go"
	"go.uber.org/zap"
)

// InviteStatusSet represents a set of invite statuses.
type InviteStatusSet map[InviteStatus]struct{}

// nolint: gochecknoglobals // immutable map, wont change at runtime
var allowedInviteStatusTransitions = map[InviteStatus]InviteStatusSet{
	InviteStatusPending: {
		InviteStatusAccepted: {},
		InviteStatusDeclined: {},
		InviteStatusExpired:  {},
	},
	InviteStatusAccepted: {},
	InviteStatusDeclined: {},
	InviteStatusExpired:  {},
}

func validateInviteStatusTransition(oldInviteStatus InviteStatus, newInviteStatus InviteStatus) bool {
	possibleNextStates := allowedInviteStatusTransitions[oldInviteStatus]

	var ok bool
	_, ok = possibleNextStates[newInviteStatus]
	return ok
}

type checkIfUsersInGroupParams struct {
	ctx              context.Context
	db               *database.Database
	fromUserID       uint32
	toUserFirstName  string
	toUserLastName   string
	slotifyGroupID   uint32
	slotifyGroupName string
	toUserID         uint32
}

func checkIfUsersInGroup(p checkIfUsersInGroupParams) error {
	var fromUserInGroup bool
	var err error // check if user creating the invite is in the group
	if fromUserInGroup, err = database.CheckMemberInSlotifyGroupWrapper(p.ctx, p.db,
		database.CheckMemberInSlotifyGroupParams{
			UserID:         p.fromUserID,
			SlotifyGroupID: p.slotifyGroupID,
		}); err != nil {
		return fmt.Errorf("failed to see if the user creating the invite is in group: %w", err)
	}

	if !fromUserInGroup {
		return fmt.Errorf("you are not a part of the group %s, cannot send invite", p.slotifyGroupName)
	}

	var toUserInGroup bool
	// check if user creating the invite is in the group
	if toUserInGroup, err = database.CheckMemberInSlotifyGroupWrapper(p.ctx, p.db,
		database.CheckMemberInSlotifyGroupParams{
			UserID:         p.toUserID,
			SlotifyGroupID: p.slotifyGroupID,
		}); err != nil {
		return fmt.Errorf("failed to see if the user being sent the invite is already in group: %w", err)
	}
	if toUserInGroup {
		return fmt.Errorf("user %s %s is already a part of the group, invite not sent", p.toUserFirstName,
			p.toUserLastName)
	}

	return nil
}

type sendPostInviteNotificationParams struct {
	ctx             context.Context
	toUserID        uint32
	fromUserID      uint32
	notifService    notification.Service
	logger          *logger.Logger
	db              *database.Database
	groupName       string
	toUserFirstName string
	toUserLastName  string
}

func sendPostInviteNotification(p sendPostInviteNotificationParams) {
	// Create notification to user who has been invited
	if err := p.notifService.SendNotification(p.ctx, p.logger, p.db,
		[]uint32{p.toUserID}, database.CreateNotificationParams{
			Message: fmt.Sprintf("You have a new invite to team %s!", p.groupName),
			Created: time.Now(),
		}); err != nil {
		p.logger.Errorf("invite api: failed to send notification to toUser",
			zap.Error(err))
	}

	// Create notification to user who created the invite
	if err := p.notifService.SendNotification(p.ctx, p.logger, p.db,
		[]uint32{p.fromUserID}, database.CreateNotificationParams{
			Message: fmt.Sprintf(
				"You successfully created an invite on behalf of team %s to %s %s!",
				p.groupName, p.toUserFirstName, p.toUserLastName), Created: time.Now(),
		}); err != nil {
		p.logger.Errorf("invite api: failed to send notification to fromUser",
			zap.Error(err))
	}
}

type validateAndUpdateInviteStatusParams struct {
	ctx       context.Context
	qtx       *database.Queries
	inviteID  uint32
	l         *logger.Logger
	userID    uint32
	newStatus InviteStatus
}

// validateAndUpdateInviteStatus is used for declining/updating an invite, returns the invite and an error.
func validateAndUpdateInviteStatus(p validateAndUpdateInviteStatusParams) (database.Invite, error) {
	var err error
	var invite database.Invite

	if invite, err = p.qtx.GetInviteByID(p.ctx, p.inviteID); err != nil {
		return database.Invite{}, fmt.Errorf("failed to get invite details from invite id: %w", err)
	}

	if invite.ToUserID != p.userID {
		return database.Invite{}, errors.New("only the user the invite is sent to can edit the status")
	}

	if ok := validateInviteStatusTransition(InviteStatus(invite.Status), p.newStatus); !ok {
		return database.Invite{}, fmt.Errorf(
			"invalid invite state transition, cannot go from %s to %s", invite.Status, p.newStatus,
		)
	}

	err = retry.Do(func() error {
		var rows int64
		rows, err = p.qtx.UpdateInviteStatus(p.ctx,
			database.UpdateInviteStatusParams{
				ID:     p.inviteID,
				Status: database.InviteStatus(p.newStatus),
			})

		if rows != 1 {
			return database.WrongNumberSQLRowsError{ActualRows: rows, ExpectedRows: []int64{1}}
		}

		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				return fmt.Errorf("context cancelled updating invite status: %w",
					err)
			case errors.Is(err, context.DeadlineExceeded):
				return fmt.Errorf("deadline exceeded during updating invite status: %w", err)
			default:
				return fmt.Errorf("failed to update invite status: %w", err)
			}
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Millisecond*500))
	if err != nil {
		p.l.Error("failed to update invite status", zap.Error(err))
		return database.Invite{}, fmt.Errorf("failed to set invite status to %s", p.newStatus)
	}

	return invite, nil
}

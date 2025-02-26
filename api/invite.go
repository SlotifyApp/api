package api

import (
	"context"
	"fmt"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/logger"
	"github.com/SlotifyApp/slotify-backend/notification"
	"go.uber.org/zap"
)

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
	var err error
	// check if user creating the invite is in the group
	if fromUserInGroup, err = database.CheckMemberIsInSlotifyGroup(p.ctx, p.db,
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
	if toUserInGroup, err = database.CheckMemberIsInSlotifyGroup(p.ctx, p.db,
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

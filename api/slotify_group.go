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

type AddUserToSlotifyGroupParams struct {
	ctx            context.Context
	userID         uint32
	slotifyGroupID uint32
	l              *logger.Logger
	qtx            *database.Queries
	notifService   notification.Service
}

func AddUserToSlotifyGroup(p AddUserToSlotifyGroupParams) error {
	userID := p.userID
	slotifyGroupID := p.slotifyGroupID
	qtx := p.qtx
	l := p.l
	ctx := p.ctx

	addUserToGroupParams := database.AddUserToSlotifyGroupParams{
		UserID:         p.userID,
		SlotifyGroupID: p.slotifyGroupID,
	}

	err := database.AddUserToSlotifyGroupWrapper(ctx, qtx, addUserToGroupParams)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	sg, err := qtx.GetSlotifyGroupByID(ctx, slotifyGroupID)
	if err != nil {
		return fmt.Errorf("failed to get group by id: %w", err)
	}

	dbParams := database.GetAllSlotifyGroupMembersExceptParams{
		SlotifyGroupID: slotifyGroupID,
		UserID:         userID,
	}

	members, err := qtx.GetAllSlotifyGroupMembersExcept(ctx, dbParams)
	if err != nil {
		return fmt.Errorf("failed to get slotify group members except new member: %w", err)
	}

	u, err := qtx.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user by id: %w", err)
	}

	allMemberNotif := database.CreateNotificationParams{
		Message: fmt.Sprintf("Say hi to %s, he just joined SlotifyGroup %s", u.FirstName+" "+u.LastName, sg.Name),
		Created: time.Now(),
	}

	newMemberNotif := database.CreateNotificationParams{
		Message: fmt.Sprintf("You were added to SlotifyGroup %s!", sg.Name),
		Created: time.Now(),
	}

	if err = p.notifService.SendNotification(ctx, l, qtx, members, allMemberNotif); err != nil {
		// Don't return error, attempt to send individual notification too
		l.Errorf(
			"slotifyGroup api: failed to send notification to all existing users of slotifyGroup, adding slotifyGroup member",
			zap.Error(err))
	}

	if err = p.notifService.SendNotification(ctx, l, qtx, []uint32{userID}, newMemberNotif); err != nil {
		l.Errorf(
			"slotifyGroup api: failed to send notification to user that just joined slotifyGroup",
			zap.Error(err))
	}

	return nil
}

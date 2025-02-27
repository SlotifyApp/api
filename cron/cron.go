package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/SlotifyApp/slotify-backend/database"
	"github.com/SlotifyApp/slotify-backend/logger"
	"github.com/avast/retry-go"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const (
	BATCHSIZE = 50
)

// RegisterDBCronJobs registers db functions to run at midnight everyday.
func RegisterDBCronJobs(ctx context.Context, db *database.Database, l *logger.Logger) error {
	// Max time for all cron jobs is 2 hours
	ctx, cancel := context.WithTimeout(ctx, 2*time.Hour)
	defer cancel()

	var err error
	c := cron.New()
	if _, err = c.AddFunc("@midnight", func() {
		removeWeekOldNotifications(ctx, db, l)
	}); err != nil {
		return fmt.Errorf("failed to register daily remove week old notifications cron job: %w", err)
	}
	if _, err = c.AddFunc("@midnight", func() {
		removeWeekOldInvites(ctx, db, l)
	}); err != nil {
		return fmt.Errorf("failed to register daily remove week old invites cron job: %w", err)
	}

	if _, err = c.AddFunc("@midnight", func() {
		expireInvites(ctx, db, l)
	}); err != nil {
		return fmt.Errorf("failed to register expire invites cron job: %w", err)
	}

	c.Start()

	return nil
}

// removeWeekOldInvites will delete invites that are a week old from the db.
// nolint: dupl// It's ok to duplicate this, it's a batch and not much code
func removeWeekOldInvites(ctx context.Context, db *database.Database, l *logger.Logger) {
	l.Info("running remove week old invites cron job")

	var affectedRows int64 = 1
	// stop when there are no more affected rows
	for affectedRows != 0 {
		tx, err := db.DB.Begin()
		if err != nil {
			l.Error("failed to start db transaction", zap.Error(err))
			return
		}

		defer func() {
			if err = tx.Rollback(); err != nil {
				l.Error("failed to rollback db transaction", zap.Error(err))
			}
		}()

		qtx := db.WithTx(tx)

		err = retry.Do(func() error {
			affectedRows, err = qtx.BatchDeleteWeekOldInvites(ctx, BATCHSIZE)
			if err != nil {
				return fmt.Errorf("failed to batch delete week old invites: %w", err)
			}

			return nil
		}, retry.Attempts(5), retry.Delay(time.Second))
		if err != nil {
			l.Error("failed to batch delete week old invites AFTER 5 retries", zap.Error(err))
			return
		}

		if err = tx.Commit(); err != nil {
			l.Error("failed to commit db transaction", zap.Error(err))
			return
		}
	}
}

// removeWeekOldNotifications will delete notifications from the db that are a week old.
// nolint: dupl// It's ok to duplicate this, it's a batch and not much code
func removeWeekOldNotifications(ctx context.Context, db *database.Database, l *logger.Logger) {
	l.Info("running remove week old notifications cron job")

	var affectedRows int64 = 1
	// stop when there are no more affected rows
	for affectedRows != 0 {
		tx, err := db.DB.Begin()
		if err != nil {
			l.Error("failed to start db transaction", zap.Error(err))
			return
		}

		defer func() {
			if err = tx.Rollback(); err != nil {
				l.Error("failed to rollback db transaction", zap.Error(err))
			}
		}()

		qtx := db.WithTx(tx)

		err = retry.Do(func() error {
			affectedRows, err = qtx.BatchDeleteWeekOldNotifications(ctx, BATCHSIZE)
			if err != nil {
				return fmt.Errorf("failed to batch delete week old notifications: %w", err)
			}

			return nil
		}, retry.Attempts(5), retry.Delay(time.Second))
		if err != nil {
			l.Error("failed to batch delete week old notifications AFTER 5 retries", zap.Error(err))
			return
		}

		if err = tx.Commit(); err != nil {
			l.Error("failed to commit db transaction", zap.Error(err))
			return
		}
	}
}

// expireInvites will expire all invites that have passed their expiry date.
func expireInvites(ctx context.Context, db *database.Database, l *logger.Logger) {
	l.Info("running expire invites cron job")

	var affectedRows int64 = 1
	for affectedRows != 0 {
		tx, err := db.DB.Begin()
		if err != nil {
			l.Error("failed to start db transaction", zap.Error(err))
			return
		}

		defer func() {
			if err = tx.Rollback(); err != nil {
				l.Error("failed to rollback db transaction", zap.Error(err))
			}
		}()

		qtx := db.WithTx(tx)

		err = retry.Do(func() error {
			if affectedRows, err = qtx.BatchExpireInvites(ctx, BATCHSIZE); err != nil {
				return fmt.Errorf("failed to batch expire invites: %w", err)
			}

			return nil
		}, retry.Attempts(5), retry.Delay(time.Second))
		if err != nil {
			l.Error("failed to batch expire invites AFTER 5 retries", zap.Error(err))
			return
		}

		if err = tx.Commit(); err != nil {
			l.Error("failed to commit db transaction", zap.Error(err))
			return
		}
	}
}

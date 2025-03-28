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
func RegisterDBCronJobs(db *database.Database, l *logger.Logger) error {
	// Max time for all cron jobs is 5 hours
	var err error
	c := cron.New()
	if _, err = c.AddFunc("@midnight", func() {
		RemoveWeekOldNotifications(context.Background(), db, l)
	}); err != nil {
		return fmt.Errorf("failed to register daily remove week old notifications cron job: %w", err)
	}
	if _, err = c.AddFunc("@midnight", func() {
		RemoveWeekOldInvites(context.Background(), db, l)
	}); err != nil {
		return fmt.Errorf("failed to register daily remove week old invites cron job: %w", err)
	}

	if _, err = c.AddFunc("@midnight", func() {
		ExpireInvites(context.Background(), db, l)
	}); err != nil {
		return fmt.Errorf("failed to register expire invites cron job: %w", err)
	}

	c.Start()

	return nil
}

// RemoveWeekOldInvites will delete invites that are a week old from the db.
// nolint: dupl// It's ok to duplicate this, it's a batch and not much code
func RemoveWeekOldInvites(ctx context.Context, db *database.Database, l *logger.Logger) {
	ctx, cancel := context.WithTimeout(ctx, time.Hour*2)
	defer cancel()

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

// RemoveWeekOldNotifications will delete notifications from the db that are a week old.
// nolint: dupl// It's ok to duplicate this, it's a batch and not much code
func RemoveWeekOldNotifications(ctx context.Context, db *database.Database, l *logger.Logger) {
	ctx, cancel := context.WithTimeout(ctx, time.Hour*2)
	defer cancel()

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

// ExpireInvites will expire all invites that have passed their expiry date.
func ExpireInvites(ctx context.Context, db *database.Database, l *logger.Logger) {
	ctx, cancel := context.WithTimeout(ctx, time.Hour*2)
	defer cancel()

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

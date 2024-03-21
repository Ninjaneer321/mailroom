// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/user"
)

// DefaultNotifier is the default implementation of the Notifier interface
// It routes notifications to the appropriate transport based on the user's preferences
type DefaultNotifier struct {
	userStore  user.Store
	transports []Transport
}

func (d *DefaultNotifier) Push(ctx context.Context, notification common.Notification) error {
	var errs []error

	recipientUser, err := d.userStore.Find(notification.Recipient)
	if err != nil {
		slog.Debug("failed to find user", "user", notification.Recipient, "error", err)
		return fmt.Errorf("failed to find recipient user: %w", err)
	}

	for _, transport := range d.transports {
		if !recipientUser.Wants(notification.Type, transport.ID()) {
			slog.Debug("user does not want this notification this way", "user", recipientUser, "transport", transport.ID())
			continue
		}

		slog.Info("pushing notification", "user", recipientUser, "transport", transport.ID())
		// TODO: We should decorate these with some retry logic
		if err = transport.Push(ctx, notification); err != nil {
			slog.Error("failed to push notification", "user", recipientUser, "transport", transport.ID(), "error", err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

var _ Notifier = &DefaultNotifier{}

func New(userStore user.Store, transports ...Transport) *DefaultNotifier {
	return &DefaultNotifier{
		userStore:  userStore,
		transports: transports,
	}
}

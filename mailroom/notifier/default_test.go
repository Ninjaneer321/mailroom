// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/seatgeek/mailroom/mailroom/common"
	"github.com/seatgeek/mailroom/mailroom/identifier"
	"github.com/seatgeek/mailroom/mailroom/notification"
	"github.com/seatgeek/mailroom/mailroom/notifier"
	"github.com/seatgeek/mailroom/mailroom/user"
	"github.com/stretchr/testify/assert"
)

type wantSent struct {
	event     common.EventType
	transport common.TransportKey
}

func TestDefaultNotifier_Push(t *testing.T) {
	t.Parallel()

	unknownUser := identifier.NewCollection()

	knownUser := user.New(
		"rufus",
		user.WithIdentifier(identifier.New(identifier.GenericUsername, "rufus")),
		user.WithPreference("com.example.one", "email", true),
		user.WithPreference("com.example.one", "slack", true),
		user.WithPreference("com.example.two", "email", true),
		user.WithPreference("com.example.two", "slack", false),
	)

	store := user.NewInMemoryStore(knownUser)

	tests := []struct {
		name         string
		notification common.Notification
		transports   []notifier.Transport
		wantSent     []wantSent
		wantErrs     []error
	}{
		{
			name:         "user wants all",
			notification: notificationFor("com.example.one", knownUser.Identifiers),
			transports: []notifier.Transport{
				&fakeTransport{key: "email"},
				&fakeTransport{key: "slack"},
			},
			wantSent: []wantSent{
				{event: "com.example.one", transport: "email"},
				{event: "com.example.one", transport: "slack"},
			},
		},
		{
			name:         "user wants some",
			notification: notificationFor("com.example.two", knownUser.Identifiers),
			transports: []notifier.Transport{
				&fakeTransport{key: "email"},
				&fakeTransport{key: "slack"},
			},
			wantSent: []wantSent{
				{event: "com.example.two", transport: "email"},
			},
		},
		{
			name:         "user wants none",
			notification: notificationFor("com.example.three", knownUser.Identifiers),
			transports: []notifier.Transport{
				&fakeTransport{key: "slack"},
			},
			wantSent: []wantSent{},
		},
		{
			name:         "unknown user",
			notification: notificationFor("com.example.one", unknownUser),
			wantSent:     []wantSent{},
			wantErrs:     []error{user.ErrUserNotFound, user.ErrUserNotFound},
		},
		{
			name:         "transport fails",
			notification: notificationFor("com.example.one", knownUser.Identifiers),
			transports: []notifier.Transport{
				&fakeTransport{key: "email", returns: errSomethingFailed},
			},
			wantSent: []wantSent{},
			wantErrs: []error{errSomethingFailed},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notifier := notifier.New(store, tc.transports...)

			errs := notifier.Push(context.Background(), tc.notification)

			if len(tc.wantErrs) == 0 {
				assert.NoError(t, errs)
			} else {
				for _, wantErr := range tc.wantErrs {
					assert.ErrorIs(t, errs, wantErr)
				}
			}

			assertSent(t, tc.wantSent, tc.transports)
		})
	}
}

func assertSent(t *testing.T, want []wantSent, transports []notifier.Transport) {
	t.Helper()

	for _, w := range want {
		matched := false
		for _, transport := range transports {
			if transport.Key() == w.transport {
				matched = true
				assert.Contains(t, transport.(*fakeTransport).sent, w.event)
				continue
			}
		}

		if !matched {
			assert.Failf(t, "expected transport not found", "transport %q not found", w.transport)
		}
	}
}

var errSomethingFailed = fmt.Errorf("some transport error occurred")

type fakeTransport struct {
	key     common.TransportKey
	sent    []common.EventType
	returns error
}

var _ notifier.Transport = (*fakeTransport)(nil)

func (f *fakeTransport) Key() common.TransportKey {
	return f.key
}

func (f *fakeTransport) Push(_ context.Context, notification common.Notification) error {
	if f.returns != nil {
		return f.returns
	}

	f.sent = append(f.sent, notification.Type())

	return nil
}

func notificationFor(eventType common.EventType, identifiers identifier.Collection) common.Notification {
	return notification.NewBuilder(eventType).
		WithRecipient(identifiers.ToList()...).
		WithDefaultMessage("hello world").
		Build()
}

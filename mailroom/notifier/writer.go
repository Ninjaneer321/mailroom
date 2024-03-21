// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package notifier

import (
	"context"
	"fmt"
	"io"

	"github.com/seatgeek/mailroom/mailroom/common"
)

// WriterNotifier is a notifier that simply writes notifications somewhere, like a file or stdout
// It is primarily used for testing and debugging
type WriterNotifier struct {
	id     common.TransportID
	writer io.Writer
}

var _ Transport = &WriterNotifier{}

func (c *WriterNotifier) ID() common.TransportID {
	return c.id
}

func (c *WriterNotifier) Push(ctx context.Context, notifications ...*common.Notification) error {
	for _, n := range notifications {
		_, err := fmt.Fprintf(
			c.writer,
			"notification: type=%s, from=%s, to=%s, message=%s\n",
			n.Type,
			n.Initiator,
			n.Recipient,
			n.Message.Render("writer"),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func NewWriterNotifier(id common.TransportID, writer io.Writer) *WriterNotifier {
	return &WriterNotifier{id: id, writer: writer}
}
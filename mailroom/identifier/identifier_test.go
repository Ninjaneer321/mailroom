// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaceAndKind_Split(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input         NamespaceAndKind
		wantNamespace string
		wantKind      string
	}{
		{
			input:         NamespaceAndKind("email"),
			wantNamespace: "",
			wantKind:      "email",
		},
		{
			input:         NamespaceAndKind("gitlab.com/username"),
			wantNamespace: "gitlab.com",
			wantKind:      "username",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(string(tc.input), func(t *testing.T) {
			t.Parallel()

			namespace, kind := tc.input.Split()

			assert.Equal(t, tc.wantNamespace, namespace)
			assert.Equal(t, tc.wantKind, kind)

			// We'll also test the Namespace() and Kind() methods here since they are simple wrappers around Split()
			assert.Equal(t, tc.wantNamespace, tc.input.Namespace())
			assert.Equal(t, tc.wantKind, tc.input.Kind())
		})
	}
}

func TestNewNamespaceAndKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		namespace string
		kind      any
		want      NamespaceAndKind
	}{
		{
			kind: KindEmail,
			want: "email",
		},
		{
			namespace: "gitlab.com",
			kind:      "username",
			want:      "gitlab.com/username",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(string(tc.want), func(t *testing.T) {
			t.Parallel()

			var got NamespaceAndKind

			switch kind := tc.kind.(type) {
			case Kind:
				got = NewNamespaceAndKind(tc.namespace, kind)
			case string:
				got = NewNamespaceAndKind(tc.namespace, kind)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("no namespace", func(t *testing.T) {
		t.Parallel()

		got := New("email", "codell@seatgeek.com")

		assert.Equal(t, NamespaceAndKind("email"), got.NamespaceAndKind)
		assert.Equal(t, "codell@seatgeek.com", got.Value)
	})

	t.Run("with namespace", func(t *testing.T) {
		t.Parallel()

		got := New("gitlab.com/username", "codell")

		assert.Equal(t, NamespaceAndKind("gitlab.com/username"), got.NamespaceAndKind)
		assert.Equal(t, "codell", got.Value)
	})

	t.Run("int64 value", func(t *testing.T) {
		t.Parallel()

		got := New("gitlab.com/id", int64(123456))

		assert.Equal(t, NamespaceAndKind("gitlab.com/id"), got.NamespaceAndKind)
		assert.Equal(t, "123456", got.Value)
	})

	t.Run("namespace and kind already a NamespaceAndKind type", func(t *testing.T) {
		t.Parallel()

		nsAndKind := NamespaceAndKind("slack.com/id")

		got := New(
			nsAndKind,
			"U1234567",
		)

		assert.Equal(t, nsAndKind, got.NamespaceAndKind)
		assert.Equal(t, "U1234567", got.Value)
	})
}

func TestCollection_ToList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		collection Collection
		want       []Identifier
	}{
		{
			name:       "empty",
			collection: Collection{},
			want:       nil,
		},
		{
			name: "non-empty",
			collection: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			want: []Identifier{
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.collection.ToList()

			assert.ElementsMatch(t, tc.want, got)
		})
	}
}

func TestCollection_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original Collection
		add      Collection
		want     Collection
	}{
		{
			name: "adds and overwrites",
			original: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@seatgeek.com"),
			),
			add: NewCollection(
				New("id", "123"),
				New("email", "rufus@example.com"),
			),
			want: NewCollection(
				New("username", "rufus"),
				New("email", "rufus@example.com"),
				New("id", "123"),
			),
		},
		{
			name:     "nil original",
			original: nil,
			add:      NewCollection(New("username", "rufus")),
			want:     NewCollection(New("username", "rufus")),
		},
		{
			name:     "empty original",
			original: Collection{},
			add:      NewCollection(New("username", "rufus")),
			want:     NewCollection(New("username", "rufus")),
		},
		{
			name:     "nil add",
			original: NewCollection(New("username", "rufus")),
			add:      nil,
			want:     NewCollection(New("username", "rufus")),
		},
		{
			name:     "empty add",
			original: NewCollection(New("username", "rufus")),
			add:      Collection{},
			want:     NewCollection(New("username", "rufus")),
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.original.Add(tc.add)

			assert.Equal(t, tc.want, tc.original)
		})
	}
}

func TestNewCollection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []Identifier
		want Collection
	}{
		{
			name: "empty",
			args: []Identifier{},
			want: Collection{},
		},
		{
			name: "non-empty",
			args: []Identifier{
				New(KindUsername, "rufus"),
				New(KindEmail, "rufus@seatgeek.com"),
			},
			want: Collection{
				"username": "rufus",
				"email":    "rufus@seatgeek.com",
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewCollection(tc.args...)

			assert.Equal(t, tc.want, got)
		})
	}
}

package database

import (
	"context"
	"io"
)

type Storage interface {
	// Save saves the link to database.
	Save(ctx context.Context, link string, name string) (int64, error)

	// List returns the links in map. The key is number and the value is the link.
	List(ctx context.Context) []Record

	// DeleteByID deletes the links by given IDs.
	DeleteByID(ctx context.Context, id int64) (bool, error)

	// UpdateByID updates the link for given id.
	UpdateByID(ctx context.Context, id int64, link string) (bool, error)

	LinkByID(ctx context.Context, id int64) (string, error)

	io.Closer
}

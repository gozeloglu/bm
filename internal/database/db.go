package database

import (
	"context"
)

type Storage interface {
	// Save saves the link to database.
	Save(ctx context.Context, link string, name string, categoryName string) (int64, error)

	// List returns the links in map. The key is number and the value is the link.
	List(ctx context.Context) ([]Record, error)

	// DeleteByID deletes the links by given IDs.
	DeleteByID(ctx context.Context, id int64) (bool, error)

	Close() error
}

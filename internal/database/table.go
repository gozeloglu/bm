package database

import "time"

type Record struct {
	ID        int64
	Link      string
	CreatedAt time.Time
}

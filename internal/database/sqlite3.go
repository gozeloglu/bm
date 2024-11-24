package database

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
)

var errCreateTable = errors.New("error while creating table")

type SQLite3 struct {
	db     *sql.DB
	logger *slog.Logger
}

type OptFunc func(*SQLite3) *SQLite3

func NewSQLite3(opts ...OptFunc) *SQLite3 {
	sqlite3 := &SQLite3{}
	for _, opt := range opts {
		opt(sqlite3)
	}
	return sqlite3
}

// WithLogger sets custom logger.
func WithLogger(logger *slog.Logger) OptFunc {
	return func(sqlite3 *SQLite3) *SQLite3 {
		sqlite3.logger = logger
		return sqlite3
	}
}

// Open opens database connections.
func (s *SQLite3) Open(ctx context.Context, filename string) error {
	s.logger.InfoContext(ctx, "opening database")
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		s.logger.Info("failed to open database ---", "filename", err.Error())
		return err
	}
	s.db = db

	err = s.createTable(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLite3) Close() error {
	return s.db.Close()
}

func (s *SQLite3) Save(ctx context.Context, link string) (int64, error) {
	res, err := s.db.ExecContext(ctx, "INSERT INTO links (link) VALUES (?)", link)
	if err != nil {
		return -1, err
	}
	linkID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return linkID, nil
}

func (s *SQLite3) List(ctx context.Context) []Record {
	rows, err := s.db.QueryContext(ctx, "SELECT id, link FROM links")
	if err != nil {
		return nil
	}
	defer rows.Close()

	records := make([]Record, 0)

	for rows.Next() {
		var id int64
		var link string
		if err := rows.Scan(&id, &link); err != nil {
			s.logger.ErrorContext(ctx, "error scanning row")
			continue
		}
		records = append(records, Record{ID: id, Link: link})
	}

	return records
}

func (s *SQLite3) DeleteByID(ctx context.Context, id int64) (bool, error) {
	res, err := s.db.ExecContext(ctx, "DELETE FROM links WHERE id=(?)", id)
	if err != nil {
		s.logger.ErrorContext(ctx, "error deleting link id:", id)
		return false, err

	}
	n, _ := res.RowsAffected()
	if n > 0 {
		s.logger.InfoContext(ctx, "deleted link id:", id)
		return true, nil
	}
	return false, nil
}

func (s *SQLite3) UpdateByID(ctx context.Context, id int64, link string) (bool, error) {
	res, err := s.db.ExecContext(ctx, "UPDATE links SET link=? WHERE id=?", link, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "error updating link", "id", id)
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		s.logger.ErrorContext(ctx, "error updating link", "id", id)
		return false, err
	}
	if n > 0 {
		s.logger.InfoContext(ctx, "updated link", "id:", id)
		return true, nil
	}
	return false, nil
}

func (s *SQLite3) LinkByID(ctx context.Context, id int64) (string, error) {
	row := s.db.QueryRowContext(ctx, "SELECT link FROM links WHERE id=?", id)
	if row.Err() != nil {
		s.logger.ErrorContext(ctx, "error getting link id:", id)
		return "", row.Err()
	}
	var link string
	if err := row.Scan(&link); err != nil {
		s.logger.ErrorContext(ctx, "error scanning row")
		return "", err
	}
	return link, nil
}

func (s *SQLite3) createTable(ctx context.Context) error {
	createTableQuery := `CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    link TEXT NOT NULL
                                 );`

	_, err := s.db.ExecContext(ctx, createTableQuery)
	if err != nil {
		return errCreateTable
	}

	return nil
}
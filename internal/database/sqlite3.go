package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/mattn/go-sqlite3"
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

	err = s.createTables(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLite3) Close() error {
	return s.db.Close()
}

func (s *SQLite3) Save(ctx context.Context, link string, name string, categoryName string) (int64, error) {
	// insert categoryName to the categories table
	insertCategoryQuery := `
INSERT OR IGNORE INTO categories (name)
VALUES (?);
`
	categoryName = strings.TrimSpace(categoryName)
	categoryName = strings.ToUpper(categoryName)
	res, err := s.db.ExecContext(ctx, insertCategoryQuery, categoryName)
	if err != nil {
		fmt.Println("error happened while inserting to categories table:", err)
		return -1, err
	}
	categoryID, _ := res.LastInsertId()

	// If the category name already exists.
	if categoryID == 0 {
		categoryID = s.fetchCategoryID(ctx, categoryName)
	}

	// insert link, name, and category id to links table
	insertLinkQuery := `
INSERT INTO links (link, name, category_id) VALUES (?, ?, ?)
`
	link = strings.TrimSpace(link)
	name = strings.TrimSpace(name)
	res, err = s.db.ExecContext(ctx, insertLinkQuery, link, name, categoryID)
	if err != nil {
		fmt.Println("error happened while inserting to links table")
		return -1, err
	}
	linkID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return linkID, nil
}

func (s *SQLite3) List(ctx context.Context) []Record {
	query := `
	SELECT links.id, links.link, links.name, categories.name
	FROM links
	INNER JOIN categories
	ON links.category_id = categories.id;
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	records := make([]Record, 0)

	for rows.Next() {
		var id int64
		var link string
		var name string
		var categoryName string
		if err := rows.Scan(&id, &link, &name, &categoryName); err != nil {
			s.logger.ErrorContext(ctx, "error scanning row")
			continue
		}
		records = append(records, Record{ID: id, Link: link, Name: name, CategoryName: categoryName})
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

func (s *SQLite3) createTables(ctx context.Context) error {
	createLinksTableQuery := `CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link TEXT NOT NULL,
	name TEXT,
	category_id INTEGER
                                 );`

	createCategoryTableQuery := `CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE
                                      );`

	_, err := s.db.ExecContext(ctx, createLinksTableQuery)
	if err != nil {
		return errCreateTable
	}

	_, err = s.db.ExecContext(ctx, createCategoryTableQuery)
	if err != nil {
		return errCreateTable
	}

	return nil
}

func (s *SQLite3) fetchCategoryID(ctx context.Context, categoryName string) int64 {
	row := s.db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", categoryName)
	var id int64
	_ = row.Scan(&id)
	return id
}

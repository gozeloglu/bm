package database

import (
	"context"
	"database/sql"
	"errors"
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
	s.logger.Info("opening database")
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		s.logger.Error("failed to open database ---", "filename", err.Error())
		return err
	}
	s.db = db

	err = s.createTables(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the database. It is important to call before terminating the
// application.
func (s *SQLite3) Close() error {
	return s.db.Close()
}

// Save inserts the given link, name, and categories to links and categories tables.
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
		s.logger.Error("failed to insert category", "category", categoryName, "error", err.Error())
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
		s.logger.Error("failed to insert link", "link", link, "error", err.Error())
		return -1, err
	}
	linkID, err := res.LastInsertId()
	if err != nil {
		s.logger.Error("failed to insert link", "link", link, "error", err.Error())
		return -1, err
	}

	return linkID, nil
}

// List fetches the Record and returns slice of Record.
func (s *SQLite3) List(ctx context.Context) ([]Record, error) {
	query := `
	SELECT links.id, links.link, links.name, categories.name
	FROM links
	INNER JOIN categories
	ON links.category_id = categories.id;
	`
	rows, err := s.db.QueryContext(ctx, query)
	records := make([]Record, 0)
	if err != nil {
		s.logger.Error("failed to fetch record", "error", err.Error())
		return records, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var link string
		var name string
		var categoryName string
		if err := rows.Scan(&id, &link, &name, &categoryName); err != nil {
			s.logger.Error("error scanning row")
			continue
		}
		records = append(records, Record{ID: id, Link: link, Name: name, CategoryName: categoryName})
	}

	return records, nil
}

// DeleteByID deletes the given id from the links table.
func (s *SQLite3) DeleteByID(ctx context.Context, id int64) (bool, error) {
	res, err := s.db.ExecContext(ctx, "DELETE FROM links WHERE id=(?)", id)
	if err != nil {
		return false, err

	}
	n, _ := res.RowsAffected()
	if n > 0 {
		return true, nil
	}
	return false, nil
}

// createTables creates the links and categories tables if they do not exist.
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
		s.logger.ErrorContext(ctx, "error while creating links table", "error", err.Error())
		return errCreateTable
	}

	_, err = s.db.ExecContext(ctx, createCategoryTableQuery)
	if err != nil {
		s.logger.ErrorContext(ctx, "error while creating categories table", "error", err.Error())
		return errCreateTable
	}

	return nil
}

// fetchCategoryID fetches and returns given categoryName's id.
func (s *SQLite3) fetchCategoryID(ctx context.Context, categoryName string) int64 {
	row := s.db.QueryRowContext(ctx, "SELECT id FROM categories WHERE name = ?", categoryName)
	var id int64
	_ = row.Scan(&id)
	return id
}

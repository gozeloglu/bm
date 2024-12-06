package tui

import (
	"context"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/gozeloglu/bm/internal/database"
	"github.com/gozeloglu/bm/internal/file"
	"log"
	"log/slog"
	"path/filepath"
)

const (
	filename    = "bm.db"
	logFilename = "bm.log"
	bmDir       = ".bm"
	bmLogDir    = "log"
	dbDir       = "db"
	logLevel    = slog.LevelError
)

type App struct {
	Logger     *slog.Logger
	Ctx        context.Context
	db         database.Storage
	loggerFile *file.File
}

// NewApp creates files, folders if not exist and creates database connections.
func NewApp() *App {
	f := file.New()
	logDir, err := f.CreateDir(bmDir, bmLogDir)
	if err != nil {
		log.Fatalln("Failed to run app.")
	}

	logFile, err := f.OpenFile(logDir, logFilename)
	if err != nil {
		log.Fatalln("Failed to run app.")
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: logLevel,
	}))

	db := database.NewSQLite3(database.WithLogger(logger))
	bmDBDir, err := f.CreateDir(bmDir, dbDir)
	if err != nil {
		log.Fatalln("Failed to run app.")
	}

	err = db.Open(ctx, filepath.Join(bmDBDir, filename))
	if err != nil {
		logger.Error("failed to open database:", "error", err.Error())
		return nil
	}
	logger.Info("created connection to database")

	return &App{
		Logger:     logger,
		Ctx:        ctx,
		db:         db,
		loggerFile: f,
	}
}

// Close closes the application by closing the logger file and database.
func (a *App) Close() error {
	// close log file after closing the database
	defer a.loggerFile.Close()

	if err := a.db.Close(); err != nil {
		a.Logger.Error("failed to close database", "error", err.Error())
		return err
	}

	a.Logger.Info("database closed successfully")
	return nil
}

// Save saves the link, name, and category.
func (a *App) Save(ctx context.Context, link string, name string, categoryName string) {
	id, err := a.db.Save(ctx, link, name, categoryName)
	if err != nil {
		a.Logger.Error("failed to save link: %+w", err)
		return
	}
	a.Logger.Info("saved link with", "id", id)
	fmt.Printf("\n\n%v saved %s with id: %v\n", emoji.CheckMarkButton, link, id)
	return
}

// List returns list of database.Record.
func (a *App) List(ctx context.Context) []database.Record {
	records, err := a.db.List(ctx)
	if err != nil {
		a.Logger.Error("failed to fetch records", "error", err.Error())
		return records
	}
	return records
}

// Delete deletes the link by using the given id. It returns result as bool.
func (a *App) Delete(ctx context.Context, id int64) bool {
	ok, err := a.db.DeleteByID(ctx, id)
	if err != nil || !ok {
		a.Logger.Error("failed to delete link", "error", err)
		return false
	}
	a.Logger.Info("deleted link", "id", id)
	return true
}

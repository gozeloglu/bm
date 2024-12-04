package tui

import (
	"context"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/gozeloglu/bm-go/internal/database"
	"github.com/gozeloglu/bm-go/internal/file"
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
	logLevel    = slog.LevelInfo
)

type App struct {
	Logger *slog.Logger
	db     database.Storage
	ctx    context.Context
}

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
	defer f.Close()

	//flag.Parse()
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
		logger.ErrorContext(ctx, "failed to open database:", err.Error())
		return nil
	}
	logger.InfoContext(ctx, "created connection to database")

	return &App{
		Logger: logger,
		db:     db,
		ctx:    ctx,
	}
}

func (a *App) Close() error {
	if err := a.db.Close(); err != nil {
		a.Logger.Error("failed to close database:", err.Error())
		return err
	}

	a.Logger.Info("database closed successfully")
	return nil
}

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

func (a *App) List(ctx context.Context) []database.Record {
	return a.db.List(ctx)
}

func (a *App) Delete(ctx context.Context, id int64) bool {
	ok, err := a.db.DeleteByID(ctx, id)
	if err != nil || !ok {
		a.Logger.Error("failed to delete link: %+w", err.Error())
		return false
	}
	return true
}

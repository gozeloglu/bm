package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/gozeloglu/bm-go/internal/database"
	"github.com/gozeloglu/bm-go/internal/file"

	"github.com/enescakir/emoji"
)

const (
	filename    = "bm.db"
	logFilename = "bm.log"
	bmDir       = ".bm"
	bmLogDir    = "log"
	dbDir       = "db"
	logLevel    = slog.LevelError
	appVersion  = "0.1.0"
)

var (
	save    = flag.String("save", "", "Save new link to bm.")
	list    = flag.Bool("list", false, "List all links.")
	del     = flag.Int64("delete", 0, "Delete existing link with given link ID.")
	update  = flag.Int64("update", 0, "Update existing link with given link ID.")
	open    = flag.Int64("open", 0, "Open existing link with given link ID.")
	export  = flag.String("export", "", "Export existing links.")
	version = flag.Bool("version", false, "Show version.")
)

type App struct {
	db     database.Storage
	logger *slog.Logger
}

func Run() {
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

	flag.Parse()
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
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.ErrorContext(ctx, "failed to close database:", err.Error())
			return
		}
		logger.InfoContext(ctx, "database closed successfully")
	}()
	logger.InfoContext(ctx, "created connection to database")

	app := &App{
		db:     db,
		logger: logger,
	}
	if *save != "" {
		app.Save(ctx, *save)
		return
	}
	if *list {
		app.List(ctx)
		return
	}
	if *del > 0 {
		app.Del(ctx, *del)
		return
	}
	if *update > 0 {
		app.Update(ctx, *update)
		return
	}
	if *open > 0 {
		app.Open(ctx, *open)
		return
	}
	if *export != "" {
		app.Export(ctx, *export)
		return
	}
	if *version {
		app.Version(ctx)
		return
	}

	fmt.Printf("%v please provide correct arguments id\n", emoji.CrossMarkButton)
	fmt.Printf("For more information, type bm --help\n")
}

func (a *App) Save(ctx context.Context, link string) {
	id, err := a.db.Save(ctx, link)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to save link: %+w", err)
		return
	}
	a.logger.InfoContext(ctx, "saved link with", "id", id)
	fmt.Printf("%v saved %s with id: %v\n", emoji.CheckMarkButton, link, id)
	return
}

func (a *App) List(ctx context.Context) {
	links := a.db.List(ctx)
	fmt.Printf("ID\t\tLink\n")
	fmt.Printf("----\t\t-------------\n")
	for _, l := range links {
		fmt.Printf("%d\t\t%s\n", l.ID, l.Link)
	}
	return
}

func (a *App) Del(ctx context.Context, id int64) {
	ok, err := a.db.DeleteByID(ctx, id)
	if !ok || err != nil {
		a.logger.ErrorContext(ctx, "failed to delete link with", "id", id)
		fmt.Printf("%v failed to delete given link\n", emoji.CrossMarkButton)
		return
	}
	a.logger.InfoContext(ctx, "deleted link with id: %d", id)
	fmt.Printf("%v deleted link with id: %v\n", emoji.CheckMarkButton, id)
	return
}

func (a *App) Update(ctx context.Context, id int64) {
	if flag.NArg() == 0 {
		a.logger.ErrorContext(ctx, "no links to update")
		fmt.Printf("%v no links to update\n", emoji.CheckMarkButton)
		return
	}
	newLink := flag.Arg(0)
	ok, err := a.db.UpdateByID(ctx, id, newLink)
	if !ok || err != nil {
		a.logger.ErrorContext(ctx, "failed to update link with", "id", id)
		fmt.Printf("%v failed to update link for id=%d\n", emoji.CrossMarkButton, id)
		return
	}
	a.logger.InfoContext(ctx, "updated link with", "id", id)
	fmt.Printf("%v updated link with id: %v\n", emoji.CheckMarkButton, id)
	return
}

func (a *App) Open(ctx context.Context, id int64) {
	link, err := a.db.LinkByID(ctx, id)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to open link with", "id", id)
		fmt.Printf("%v failed to open link for #%d\n", emoji.CrossMarkButton, id)
		return
	}
	//if strings.Trim(link, " ") == "" {
	//	a.logger.ErrorContext(ctx, "failed to open link for #%d\n", id)
	//	fmt.Printf("%v invalid id #%d\n", emoji.CrossMarkButton, id)
	//	return
	//}
	ok := openBrowser(link)
	if !ok {
		a.logger.ErrorContext(ctx, "failed to open link for", "id", id, "link", link)
		fmt.Printf("%v failed to open link for #%d(%s)\n", emoji.CrossMarkButton, id, link)
		return
	}
	fmt.Printf("%v opened link for #%v(%s)\n", emoji.CheckMarkButton, id, link)
	return
}

func (a *App) Export(ctx context.Context, path string) {
	// Check if the provided path is a directory
	fileInfo, err := os.Stat(path)
	if err == nil && fileInfo.IsDir() {
		// If it's a directory, append the default filename
		path = filepath.Join(path, "bm.db")
	} else if err != nil && !os.IsNotExist(err) {
		// Handle unexpected errors during stat
		a.logger.ErrorContext(ctx, "failed to access destination path", "path", path, "error", err.Error())
		fmt.Printf("%v failed to export file: invalid destination path\n", emoji.CrossMarkButton)
		return
	}

	dbFile, err := os.Open("./bm.db")
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to open database file", "error", err.Error())
		fmt.Printf("%v failed to export file\n", emoji.CrossMarkButton)
		return
	}
	defer dbFile.Close()

	destFile, err := os.Create(path)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to export", "file", path)
		fmt.Printf("%v failed to export file\n", emoji.CrossMarkButton)
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, dbFile)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to export file", "file", path)
		fmt.Printf("%v failed to export file\n", emoji.CrossMarkButton)
		return
	}
	fmt.Printf("%v exported file to %s\n", emoji.CheckMarkButton, path)
	return
}

func (a *App) Version(_ context.Context) {
	fmt.Printf("%s\n", appVersion)
}

// openBrowser opens the given url in default browser.
func openBrowser(url string) bool {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}
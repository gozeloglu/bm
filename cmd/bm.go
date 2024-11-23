package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/gozeloglu/bm-go/internal/database"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	filename = "bm.db"
	logLevel = slog.LevelError
)

var (
	save   = flag.String("save", "", "Save new link to bm.")
	list   = flag.Bool("list", false, "List all links.")
	del    = flag.Int64("delete", 0, "Delete existing link with given link ID.")
	update = flag.Int64("update", 0, "Update existing link with given link ID.")
	open   = flag.Int64("open", 0, "Open existing link with given link ID.")
)

func main() {
	logFile, err := os.OpenFile("bm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	flag.Parse()
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: logLevel,
	}))

	db := database.NewSQLite3(database.WithLogger(logger))
	err = db.Open(ctx, filename)
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

	if *save != "" {
		link := *save
		id, err := db.Save(ctx, link)
		if err != nil {
			logger.ErrorContext(ctx, "failed to save link: %+w", err)
			return
		}
		logger.InfoContext(ctx, "saved link with", "id", id)
		fmt.Printf("%v saved %s with id: %v\n", emoji.CheckMarkButton, link, id)
		return
	}
	if *list {
		links := db.List(ctx)
		fmt.Printf("ID\t\tLink\n")
		fmt.Printf("----\t\t-------------\n")
		for _, l := range links {
			fmt.Printf("%d\t\t%s\n", l.ID, l.Link)
		}
		return
	}
	if *del > 0 {
		id := *del
		ok, err := db.DeleteByID(ctx, id)
		if !ok || err != nil {
			logger.ErrorContext(ctx, "failed to delete link with", "id", id)
			fmt.Printf("%v failed to delete given link\n", emoji.CrossMarkButton)
			return
		}
		logger.InfoContext(ctx, "deleted link with id: %d", id)
		fmt.Printf("%v deleted link with id: %v\n", emoji.CheckMarkButton, id)
		return
	}
	if *update > 0 {
		if flag.NArg() == 0 {
			logger.ErrorContext(ctx, "no links to update")
			fmt.Printf("%v no links to update\n", emoji.CheckMarkButton)
			return
		}
		id := *update
		newLink := flag.Arg(0)
		ok, err := db.UpdateByID(ctx, id, newLink)
		if !ok || err != nil {
			logger.ErrorContext(ctx, "failed to update link with", "id", id)
			fmt.Printf("%v failed to update link for id=%d\n", emoji.CrossMarkButton, id)
			return
		}
		logger.InfoContext(ctx, "updated link with", "id", id)
		fmt.Printf("%v updated link with id: %v\n", emoji.CheckMarkButton, id)
		return
	}
	if *open > 0 {
		id := *open
		link, err := db.LinkByID(ctx, id)
		if err != nil {
			logger.ErrorContext(ctx, "failed to open link with", "id", id)
			fmt.Printf("%v failed to open link for #%d\n", emoji.CrossMarkButton, id)
			return
		}
		ok := openBrowser(link)
		if !ok {
			logger.ErrorContext(ctx, "failed to open link for", "id", id, "link", link)
			fmt.Printf("%v failed to open link for #%d(%s)\n", emoji.CrossMarkButton, id, link)
			return
		}
		fmt.Printf("%v opened link for #%v(%s)\n", emoji.CheckMarkButton, id, link)
		return
	}

	fmt.Printf("%v please provide correct arguments id\n", emoji.CrossMarkButton)
	fmt.Printf("For more information, type bm --help\n")
}

// openBrowser opens the given url in default browser.
func openBrowser(url string) bool {
	if !strings.HasPrefix(url, "https://") || !strings.HasPrefix(url, "http://") {
		url = "https://" + url
	}
	fmt.Println(url)
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

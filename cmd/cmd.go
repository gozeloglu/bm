package main

import (
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/enescakir/emoji"
	"github.com/gozeloglu/bm/tui"
	tuiList "github.com/gozeloglu/bm/tui/list"
	"github.com/gozeloglu/bm/tui/textinput"
	"os"
)

const (
	appVersion = "0.2.0"
)

var (
	save    = flag.Bool("save", false, "Save new link to bm.")
	list    = flag.Bool("list", false, "List all links.")
	del     = flag.Bool("delete", false, "Delete existing link with given link ID.")
	version = flag.Bool("version", false, "Show version.")
)

func Run() {
	flag.Parse()
	app := tui.NewApp()
	defer app.Close()

	if *save {
		saveFlag(app)
		return
	}
	if *list {
		listFlag(app)
		return
	}
	if *del {
		delFlag(app)
		return
	}
	if *version {
		Version()
		return
	}

	if flag.NArg() == 0 {
		listFlag(app)
		return
	}
	fmt.Printf("%v please provide correct arguments id\n", emoji.CrossMarkButton)
	fmt.Printf("For more information, type 'bm --help'\n")
}

func Version() {
	fmt.Printf("%s\n", appVersion)
}

func saveFlag(app *tui.App) {
	if _, err := tea.NewProgram(textinput.New(app)).Run(); err != nil {
		app.Logger.Error("failed to run save program:", err.Error())
		os.Exit(1)
	}
}

func listFlag(app *tui.App) {
	if _, err := tea.NewProgram(tuiList.New(app, false), tea.WithAltScreen()).Run(); err != nil {
		app.Logger.Error("failed to run save program:", err.Error())
		os.Exit(1)
	}
}

func delFlag(app *tui.App) {
	if _, err := tea.NewProgram(tuiList.New(app, true), tea.WithAltScreen()).Run(); err != nil {
		app.Logger.Error("failed to run save program:", err.Error())
		os.Exit(1)
	}
}

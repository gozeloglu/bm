package list

import (
	"context"
	"github.com/gozeloglu/bm-go/internal/database"
	"github.com/gozeloglu/bm-go/tui"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	id         int64
	link, name string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.link }
func (i item) FilterValue() string { return i.link + i.name }
func (i item) ID() int64           { return i.id }

type Model struct {
	list            list.Model
	items           []list.Item
	bmList          []database.Record
	app             *tui.App
	deletionEnabled bool
}

func New(app *tui.App, deletionEnabled bool) Model {
	ctx := context.Background()

	// fetch the all bookmarks
	bmList := app.List(ctx)

	// convert records to []list.Item
	items := make([]list.Item, len(bmList))
	for i, it := range bmList {
		items[i] = item{
			id:   it.ID,
			link: it.Link,
			name: it.Name,
		}
	}
	m := Model{
		items:           items,
		list:            list.New(items, list.NewDefaultDelegate(), 0, 0),
		bmList:          bmList,
		app:             app,
		deletionEnabled: deletionEnabled,
	}
	m.list.Title = "Bookmarks"
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//fmt.Println(m.deletionEnabled)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" && !m.deletionEnabled {
			openBrowser(m.list.SelectedItem().(list.DefaultItem).Description())
		}
		if msg.String() == tea.KeyBackspace.String() && m.deletionEnabled {
			if len(m.bmList) > 0 {
				idx := m.list.Index()
				id := m.bmList[idx].ID
				ok := m.app.Delete(context.Background(), id)
				if ok {
					m.list.RemoveItem(m.list.Index())
					m.bmList = append(m.bmList[:idx], m.bmList[idx+1:]...)
				}
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return docStyle.Render(m.list.View())
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
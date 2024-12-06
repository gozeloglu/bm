package list

import (
	"github.com/gozeloglu/bm/internal/database"
	"github.com/gozeloglu/bm/tui"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	id       int64
	link     string
	name     string
	category string
}

func (i item) Title() string {
	title := strings.Builder{}
	if len(i.category) != 0 {
		title.WriteString(i.category)
		title.WriteString(" | ")
	}
	title.WriteString(i.name)
	return title.String()
}
func (i item) Description() string { return i.link }
func (i item) FilterValue() string { return i.category + i.link + i.name }
func (i item) ID() int64           { return i.id }

type Model struct {
	list            list.Model
	items           []list.Item
	bmList          []database.Record
	app             *tui.App
	deletionEnabled bool
}

func New(app *tui.App, deletionEnabled bool) Model {
	// fetch the all bookmarks
	bmList := app.List(app.Ctx)

	// convert records to []list.Item
	items := make([]list.Item, len(bmList))
	for i, it := range bmList {
		items[i] = item{
			id:       it.ID,
			link:     it.Link,
			name:     it.Name,
			category: it.CategoryName,
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" || msg.String() == tea.KeyEsc.String() {
			return m, tea.Quit
		}
		if msg.String() == tea.KeyEnter.String() && !m.deletionEnabled {
			openBrowser(m.list.SelectedItem().(list.DefaultItem).Description())
		}
		if msg.String() == tea.KeyBackspace.String() && m.deletionEnabled {
			if len(m.bmList) > 0 {
				idx := m.list.Index()
				id := m.bmList[idx].ID
				ok := m.app.Delete(m.app.Ctx, id)
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

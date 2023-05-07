package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"pomodoro-tui/spotify"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle  = lipgloss.NewStyle().Margin(1, 2)
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
)

const (
	SearchView = iota
	SearchTypeView
	SearchResultsView
)

type keymap struct {
	reset key.Binding
	quit  key.Binding
	enter key.Binding
	back  key.Binding
}

type model struct {
	textInput          textinput.Model
	textInputValue     string
	help               help.Model
	keymap             keymap
	searchResultsList  list.Model
	searchResultsTable table.Model
	searchTypeList     list.Model
	searchType         string
	spotify            spotify.SpotifyTokenRequest
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.reset,
		m.keymap.quit,
		m.keymap.enter,
	})
}

func (m model) GetCurrentView() int {
	// right no there are 3 views:
	// 1. text input view rename to search view
	// 2. search type list view rename to search type view
	// 3. search results list view rename to search results view
	switch {
	case m.textInputValue == "":
		return SearchView
	case m.searchType == "" && m.textInputValue != "":
		return SearchTypeView
	case m.searchType != "" && m.textInputValue != "":
		return SearchResultsView
	default:
		return SearchView
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.reset):
			m.textInput.Reset()
		case m.GetCurrentView() == SearchView && key.Matches(msg, m.keymap.enter):
			m.textInputValue = m.textInput.Value()
		case m.GetCurrentView() == SearchTypeView && key.Matches(msg, m.keymap.enter):
			m.searchType = m.searchTypeList.Items()[m.searchTypeList.Index()].(Item).title
			searchedItems := spotify.Search(m.textInputValue, m.searchType, m.spotify.AccessToken, &http.Client{})
			m.searchResultsList.Title = "Search results for " + m.textInputValue
			rows := []table.Row{}
			for _, item := range searchedItems.Artists.Items {
				rows = append(rows, table.Row{item.Name, fmt.Sprint(item.Popularity), strings.Join(item.Genres, ",  "), item.ExternalUrls.Spotify})
			}
			var columns []table.Column
			switch m.searchType {
			case spotify.Artist:
				columns = []table.Column{
					{Title: "Name", Width: 10},
					{Title: "Popularity", Width: 10},
					{Title: "Genres", Width: 30},
					{Title: "Link", Width: 60},
				}
			}
			m.searchResultsTable.SetColumns(columns)
			m.searchResultsTable.SetRows(rows)
		case m.GetCurrentView() == SearchResultsView && key.Matches(msg, m.keymap.enter):
			index := m.searchResultsTable.Cursor()
			item := m.searchResultsTable.Rows()[index]
			err := exec.Command("xdg-open", item[3]).Start()
			if err != nil {
				fmt.Println(err)
			}
		case m.GetCurrentView() == SearchResultsView && key.Matches(msg, m.keymap.back):
			m.searchType = ""
		case m.GetCurrentView() == SearchTypeView && key.Matches(msg, m.keymap.back):
			m.textInputValue = ""
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.searchTypeList.SetSize(msg.Width-h, msg.Height-v)
	}
	switch m.GetCurrentView() {
	case SearchView:
		m.textInput, cmd = m.textInput.Update(msg)
	case SearchTypeView:
		m.searchTypeList, cmd = m.searchTypeList.Update(msg)
	case SearchResultsView:
		m.searchResultsTable, cmd = m.searchResultsTable.Update(msg)

	}
	return m, cmd
}

func (m model) View() string {
	switch {
	case m.GetCurrentView() == SearchView:
		s := m.textInput.View() + "\n\n"
		s += m.helpView()
		return s
	case m.GetCurrentView() == SearchTypeView:
		s := m.searchTypeList.View() + "\n\n"
		s += m.helpView()
		return s
	case m.GetCurrentView() == SearchResultsView:
		return baseStyle.Render(m.searchResultsTable.View()) + "\n"
	default:
		return "Invalid View"
	}
}

func main() {
	textInputModel := textinput.NewModel()
	textInputModel.Placeholder = "Search..."
	textInputModel.Focus()
	textInputModel.CharLimit = 255
	textInputModel.Width = 20
	textInputModel.Prompt = "Search: \n\n"

	searchresults := NewListModel("Search results", []list.Item{})
	searchType := NewListModel("Search type", []list.Item{
		Item{title: spotify.Artist, desc: spotify.Artist + " description"},
		Item{title: spotify.Album, desc: spotify.Album + " description"},
		Item{title: spotify.Playlist, desc: spotify.Playlist + " description"},
		Item{title: spotify.Track, desc: spotify.Track + " description"},
	})

	t := table.New(
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		keymap: keymap{
			reset: key.NewBinding(
				key.WithKeys("ctrl+r"),
				key.WithHelp("ctrl + r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl + c", "quit"),
			),
			enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "submit"),
			),
			back: key.NewBinding(
				key.WithKeys("backspace"),
				key.WithHelp("backspace", "back"),
			),
		},
		help:               help.NewModel(),
		textInput:          textInputModel,
		searchResultsList:  searchresults,
		searchResultsTable: t,
		spotify:            spotify.RequestSpotifyToken(),
		searchTypeList:     searchType,
		searchType:         "",
	}

	m.searchResultsList.Title = "Search Results"
	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}

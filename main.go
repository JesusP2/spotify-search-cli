package main

import (
	"fmt"
	"net/http"
	"pomodoro-tui/spotify"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type keymap struct {
	reset key.Binding
	quit  key.Binding
	enter key.Binding
}

type model struct {
	textInput         textinput.Model
	textInputValue    string
	help              help.Model
	keymap            keymap
	searchResultsList list.Model
	searchTypeList    list.Model
	searchType        string
	spotify           spotify.SpotifyTokenRequest
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.reset):
			m.textInput.Reset()
		case key.Matches(msg, m.keymap.enter) && m.textInputValue == "":
			m.textInputValue = m.textInput.Value()
		case key.Matches(msg, m.keymap.enter) && m.textInputValue != "" && m.searchType == "":
			m.searchType = m.searchTypeList.Items()[m.searchTypeList.Index()].(Item).title
			searchedItems := spotify.Search(m.textInputValue, m.searchType, m.spotify.AccessToken, &http.Client{})
			m.searchResultsList.Title = "Search results for " + m.textInputValue
			items := []list.Item{}
			for _, item := range searchedItems.Artists.Items {
				items = append(items, Item{title: item.Name, desc: fmt.Sprint(item.Popularity)})
			}
			m.searchResultsList.SetItems(items)
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.searchResultsList.SetSize(msg.Width-h, msg.Height-v)
		m.searchTypeList.SetSize(msg.Width-h, msg.Height-v)
	}
	m.textInput, cmd = m.textInput.Update(msg)
	m.searchResultsList, cmd = m.searchResultsList.Update(msg)
	m.searchTypeList, cmd = m.searchTypeList.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch {
	case m.textInputValue == "":
		s := m.textInput.View() + "\n\n"
		s += m.helpView()
		return s
	case m.searchType == "" && m.textInputValue != "":
		s := m.searchTypeList.View() + "\n\n"
		s += m.helpView()
		return s
	default:
		m.searchResultsList.Title = "Search results for " + m.textInputValue
		return docStyle.Render(m.searchResultsList.View())
	}
}

func main() {
	textInputModel := textinput.NewModel()
	textInputModel.Placeholder = "Michael Jackson"
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
		},
		help:              help.NewModel(),
		textInput:         textInputModel,
		searchResultsList: searchresults,
		spotify:           spotify.RequestSpotifyToken(),
		searchTypeList:    searchType,
		searchType:        "",
	}

	m.searchResultsList.Title = "Search Results"
	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}

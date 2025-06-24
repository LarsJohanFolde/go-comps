package main

import (
	"fmt"
	"go-comps/db"
	"go-comps/internal/models"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
)

type model struct {
	input           textinput.Model
	allPersons      []models.Person
	filteredPersons []models.Person
	cursor          int
	selectedPerson  models.Person
	shouldClear     bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search competitors..."
	ti.Focus()
	ti.CharLimit = 54
	ti.Width = 30

	persons := db.GetPersons()

	return model{
		input:           ti,
		allPersons:      persons,
		filteredPersons: persons,
		cursor:          0,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.shouldClear = true
			return m, tea.Quit
		case "up", "shift+tab":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = 20
			}
		case "down", "tab":
			if m.cursor < len(m.filteredPersons)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "enter":
			if len(m.filteredPersons) > 0 {
				m.shouldClear = true
				m.selectedPerson = m.filteredPersons[m.cursor]
				return m, tea.Quit
			}
		}
	}

	m.input, cmd = m.input.Update(msg)

	// Filter names based on input
	query := strings.ToLower(m.input.Value())
	m.filteredPersons = nil
	for _, person := range m.allPersons {
		if strings.Contains(strings.ToLower(person.Name), query) {
			m.filteredPersons = append(m.filteredPersons, person)
		}
	}
	if m.cursor >= len(m.filteredPersons) {
		m.cursor = len(m.filteredPersons) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	return m, cmd
}

func (m model) View() string {
	if m.shouldClear {
		return ""
	}

	s := "Type to search. Use ↑/↓ to navigate, Enter to select, esc to quit.\n\n"
	s += m.input.View() + "\n\n"

	for i, person := range m.filteredPersons {
		if i > 20 {
			break
		}
		cursor := "\033[0m " // no cursor
		if i == m.cursor {
			cursor = "\033[32m >" // current selection
		}
		s += fmt.Sprintf("%s %s\033[0m\n", cursor, fmt.Sprintf("%s, %s (%s)", person.Name, person.CountryId, person.WcaId))
	}

	if len(m.filteredPersons) == 0 {
		s += "\nNo results."
	}

	return s
}

func main() {
	if len(os.Args) > 1 {
		wcaId := os.Args[1]
		competitions := db.GetUpcomingCompetitions(wcaId)
		for _, competition := range competitions {
			fmt.Printf(
				"%s, %s\n\t%s\n\n",
				competition.Name,
				competition.CountryId,
				competition.Duration(),
			)
		}
		os.Exit(0)
	}
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	if m, ok := finalModel.(model); ok {
		if m.selectedPerson.WcaId == "" {
			os.Exit(0)
		}
		fmt.Printf("\n\033[32m> %s\033[0m\n\n", m.selectedPerson.Name)
		competitions := db.GetUpcomingCompetitions(m.selectedPerson.WcaId)
		for _, competition := range competitions {
			fmt.Printf(
				"%s[%s] %s, %s\n\t%s\033[0m\n\n",
				competition.StatusColor(),
				competition.CompetingStatus,
				competition.Hyperlink(),
				competition.CountryId,
				competition.Duration(),
			)
		}
	}
}

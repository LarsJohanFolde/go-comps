package main

import (
	"flag"
	"fmt"
	"go-comps/db"
	"go-comps/internal/models"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/rodaine/table"
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
		if strings.Contains(strings.ToLower(person.NormalizedName), query) {
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
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	if m, ok := finalModel.(model); ok {
		if m.selectedPerson.WcaId == "" {
			os.Exit(1)
		}

		registrationTimeStamps := flag.Bool("t", false, "Render the registration timestamps as well as registration opening")
		listAllCompetitions := flag.Bool("a", false, "List new and old competitions")
		flag.Parse()

		competitions := db.GetUpcomingCompetitions(m.selectedPerson.WcaId)

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		if !*listAllCompetitions {
			allCompetitions := competitions
			competitions = []models.Competition{}
			for _, c := range allCompetitions {
				if c.Upcoming {
					competitions = append(competitions, c)
				}
			}
		}

		fmt.Println("\033[32m" + m.selectedPerson.Name)

		if *registrationTimeStamps {
			tbl := table.New("Status", "Competition", "Registered At", "Registration Open", "Registration Close", "Registration Timing", "Country", "Start Date", "End Date")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

			for _, competition := range competitions {
				tbl.AddRow(
					competition.CompetingStatus,
					competition.Name,
					competition.RegisteredAt.Format("2006-01-02 15:04:05"),
					competition.RegistrationOpen.Format("2006-01-02 15:04:05"),
					competition.RegistrationClose.Format("2006-01-02 15:04:05"),
					competition.RegistrationTiming(),
					competition.CountryId,
					competition.StartDate.Format("2006-01-02"),
					competition.EndDate.Format("2006-01-02"),
				)
			}
			tbl.Print()
		} else {
			tbl := table.New("Status", "Competition", "Country", "Start Date", "End Date")
			tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

			for _, competition := range competitions {
				tbl.AddRow(
					competition.CompetingStatus,
					competition.Name,
					competition.CountryId,
					competition.StartDate.Format("2006-01-02"),
					competition.EndDate.Format("2006-01-02"),
				)
			}
			tbl.Print()
		}
	}
}

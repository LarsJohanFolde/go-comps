package models

import (
	"fmt"
	"golang.org/x/text/unicode/norm"
	"strings"
	"time"
	"unicode"
)

type Person struct {
	WcaId          string
	Name           string
	NormalizedName string
	CountryId      string
}

func (p Person) NormalizeName() string {
	t := norm.NFD.String(p.Name)
	result := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		switch r {
		case 'æ':
			result = append(result, 'e')
		case 'Æ':
			result = append(result, 'E')
		default:
			result = append(result, r)
		}
	}
	return strings.ToLower(string(result))
}

type Competition struct {
	ID                string
	Name              string
	CountryId         string
	StartDate         time.Time
	EndDate           time.Time
	CompetingStatus   string
	Upcoming          bool
	RegisteredAt      time.Time
	RegistrationOpen  time.Time
	RegistrationClose time.Time
}

func (c Competition) Hyperlink() string {
	url := fmt.Sprintf("https://worldcubeassociation.org/competitions/%s", c.ID)
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, c.Name)
}

func (c Competition) RegistrationTiming() string {
	if c.RegisteredAt.Before(c.RegistrationOpen) {
		return "Early"
	} else if c.RegisteredAt.After(c.RegistrationClose) {
		return "Late"
	} else {
		return ""
	}
}

func (c Competition) StatusColor() string {
	switch c.CompetingStatus {
	case "accepted":
		return "\033[32m"
	case "waiting_list":
		return "\033[33m"
	case "rejected":
		return "\033[31m"
	case "cancelled":
		return "\033[31m"
	case "pending":
		return "\033[33m"
	}
	return "\033[0m"
}

func (c Competition) Duration() string {
	if c.StartDate == c.EndDate {
		return c.StartDate.Format("2006-01-02")
	}
	return fmt.Sprintf("%s -> %s", c.StartDate.Format("2006-01-02"), c.EndDate.Format("2006-01-02"))
}

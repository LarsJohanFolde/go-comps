package models

import (
	"fmt"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

type Person struct {
	WcaId     string
	Name      string
	CountryId string
}

func (p Person) NormalizeName() string {
	t := norm.NFD.String(p.Name)
	result := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

type Competition struct {
	ID              string
	Name            string
	CountryId       string
	StartDate       string
	EndDate         string
	CompetingStatus string
	Upcoming        bool
}

func (c Competition) Hyperlink() string {
	url := fmt.Sprintf("https://worldcubeassociation.org/competitions/%s", c.ID)
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, c.Name)
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
		return c.StartDate
	}
	return fmt.Sprintf("%s -> %s", c.StartDate, c.EndDate)
}

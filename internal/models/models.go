package models

import (
	"fmt"
)

type Person struct {
	WcaId     string
	Name      string
	CountryId string
}

type Competition struct {
	ID              string
	Name            string
	CountryId       string
	StartDate       string
	EndDate         string
	CompetingStatus string
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

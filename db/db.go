package db

import (
	"database/sql"
	"fmt"
	"go-comps/internal/models"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

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
	case "deleted":
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

func LoadDSN() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	dsn := os.Getenv("DSN")
	return dsn
}

func GetPersons() []person.Person {
	dsn := LoadDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT wca_id, name, countryId FROM Persons WHERE subId = '1'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var persons []person.Person

	for rows.Next() {
		var wcaId string
		var name string
		var countryId string
		if err := rows.Scan(&wcaId, &name, &countryId); err != nil {
			log.Fatal(err)
		}
		p := person.Person{
			WcaId:     wcaId,
			Name:      strings.Split(name, " (")[0],
			CountryId: countryId,
		}
		persons = append(persons, p)
	}

	return persons
}

func GetUpcomingCompetitions(wcaId string) []Competition {
	query := fmt.Sprintf(`
    SELECT
        c.name AS Competition,
        c.countryid AS CountryId,
        c.start_date AS StartDate,
        c.end_date AS EndDate,
        COALESCE((SELECT r1.competing_status FROM registrations r1 WHERE r1.competition_id = c.id AND r1.user_id = (
        SELECT id from users WHERE wca_id = '%s'
        )), "accepted"),
        CASE WHEN start_date > CURRENT_DATE() THEN true ELSE false END AS Upcoming
    FROM Competitions c
    WHERE c.id IN (SELECT competitionId FROM Results WHERE personid = '%s')
        OR c.id IN (
            SELECT r.competition_id
            FROM registrations r
            JOIN Competitions c1 ON r.competition_id = c.id
            WHERE r.user_id = (SELECT id FROM users WHERE wca_id = '%s')
		   )
    GROUP BY c.id
    ORDER BY c.start_date DESC
    `, wcaId, wcaId, wcaId)

	dsn := LoadDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var competitions []Competition

	for rows.Next() {
		var name string
		var countryId string
		var startDate string
		var endDate string
		var competingStatus string
		var upcoming bool
		if err := rows.Scan(&name, &countryId, &startDate, &endDate, &competingStatus, &upcoming); err != nil {
			log.Fatal(err)
		}

		if upcoming {
			c := Competition{Name: name, CountryId: countryId, StartDate: startDate, EndDate: endDate, CompetingStatus: competingStatus}
			competitions = append(competitions, c)
		}
	}
	return competitions
}

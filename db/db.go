package db

import (
	"fmt"
	"database/sql"
	"go-comps/internal/models"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Competition struct {
	ID        string
	Name      string
	CountryId string
	StartDate string
	EndDate   string
}

func (c Competition) Duration() string {
    if c.StartDate == c.EndDate {
        return c.StartDate
    }
    return fmt.Sprintf("%s -> %s", c.StartDate, c.EndDate)
}

func GetPersons() []person.Person {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	dsn := os.Getenv("DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

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
    godotenv.Load()
    query := fmt.Sprintf(`
    SELECT
        c.name AS Competition,
        c.countryid AS CountryId,
        c.start_date AS StartDate,
        c.end_date AS EndDate,
        CASE WHEN start_date > CURRENT_DATE() THEN true ELSE false END AS Upcoming
    FROM Competitions c
    WHERE c.id IN (SELECT competitionId FROM Results WHERE personid = '%s')
        OR c.id IN (
            SELECT r.competition_id
            FROM registrations r
            JOIN Competitions c1 ON r.competition_id = c.id
            WHERE 
                r.user_id = (SELECT id FROM users WHERE wca_id = '%s')
                AND (c1.start_date > CURRENT_DATE() OR (c1.results_posted_at IS NULL AND c1.cancelled_at IS NULL))
                AND r.deleted_at IS NULL
		   )
    GROUP BY c.id
    ORDER BY c.start_date DESC
    `, wcaId, wcaId)

    dsn := os.Getenv("DSN")
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }

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
        var upcoming bool
        if err := rows.Scan(&name, &countryId, &startDate, &endDate, &upcoming); err != nil {
            log.Fatal(err)
        }
        
        if upcoming {
            c := Competition{Name: name, CountryId: countryId, StartDate: startDate, EndDate: endDate}
            competitions = append(competitions, c)
        }
    }
    return competitions
}

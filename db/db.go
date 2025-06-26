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

func LoadDSN() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	dsn := os.Getenv("DSN")
	return dsn
}

func GetPersons() []models.Person {
	dsn := LoadDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT wca_id, name, country_id FROM persons WHERE sub_id = '1'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var persons []models.Person

	for rows.Next() {
		var wcaId string
		var name string
		var countryId string
		if err := rows.Scan(&wcaId, &name, &countryId); err != nil {
			log.Fatal(err)
		}
		p := models.Person{
			WcaId:     wcaId,
			Name:      strings.Split(name, " (")[0],
			CountryId: countryId,
		}
		persons = append(persons, p)
	}

	return persons
}

func GetUpcomingCompetitions(wcaId string) []models.Competition {
	query := fmt.Sprintf(`
    SELECT
        c.id AS Id,
        c.name AS Competition,
        c.country_id AS CountryId,
        c.start_date AS StartDate,
        c.end_date AS EndDate,
        COALESCE((SELECT r1.competing_status FROM registrations r1 WHERE r1.competition_id = c.id AND r1.user_id = (
        SELECT id from users WHERE wca_id = '%s'
        )), "accepted"),
        CASE WHEN end_date > CURRENT_DATE() THEN true ELSE false END AS Upcoming
    FROM competitions c
    WHERE c.id IN (
            SELECT r.competition_id
            FROM registrations r
            JOIN competitions c1 ON r.competition_id = c.id
            WHERE r.user_id = (SELECT id FROM users WHERE wca_id = '%s')
		   )
    GROUP BY c.id
    ORDER BY c.start_date DESC
    `, wcaId, wcaId)

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

	var competitions []models.Competition

	for rows.Next() {
		var id string
		var name string
		var countryId string
		var startDate string
		var endDate string
		var competingStatus string
		var upcoming bool
		if err := rows.Scan(&id, &name, &countryId, &startDate, &endDate, &competingStatus, &upcoming); err != nil {
			log.Fatal(err)
		}

		c := models.Competition{ID: id, Name: name, CountryId: countryId, StartDate: startDate, EndDate: endDate, CompetingStatus: competingStatus, Upcoming: upcoming}
		competitions = append(competitions, c)
	}
	return competitions
}

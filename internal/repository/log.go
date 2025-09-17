package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

func (m *postgresDBRepo) AllLogs(minDate time.Time, maxDate time.Time, page int, limit int) ([]models.Log, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var logs []models.Log
	var count int

	countQuery := "SELECT COUNT(*) FROM log WHERE created_on BETWEEN $1 AND $2 ;"

	// set count
	row := m.DB.QueryRowContext(ctx, countQuery, minDate, maxDate)

	err := row.Scan(&count)
	if err != nil {
		return logs, count, err
	}

	// calculate offset
	offset := limit * (page - 1)

	// build query
	query := `
	SELECT
		l.id,
		l.level,
		l.location,
		l.message,
		v.first_name,
		v.family_name,
		l.created_on
	FROM
		log l
		LEFT JOIN visitors v on v.id = l.visitor_id
	WHERE
	  l.created_on::DATE BETWEEN $1::DATE AND $2::DATE
	ORDER BY
		created_on DESC
	OFFSET
		$3
	LIMIT
		$4
	`

	rows, err := m.DB.QueryContext(ctx, query, minDate, maxDate, offset, limit)
	if err != nil {
		return logs, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var l models.Log

		// start handling null types
		var firstName sql.NullString
		var familyName sql.NullString

		err := rows.Scan(
			&l.ID,
			&l.Level,
			&l.Location,
			&l.Message,
			&firstName,
			&familyName,
			&l.CreatedOn,
		)

		// handling null types
		if firstName.Valid {
			l.Visitor.FirstName = firstName.String
		}
		if familyName.Valid {
			l.Visitor.FamilyName = familyName.String
		}

		if err != nil {
			return logs, 0, err
		}
		logs = append(logs, l)
	}

	return logs, count, nil
}

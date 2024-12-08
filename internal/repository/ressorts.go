package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// RessortByID gets a ressort by id
func (m *postgresDBRepo) RessortByID(id int) (models.Ressort, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ressort models.Ressort

	query := `SELECT id, ressort_name, slack_channel, created_on, updated_on FROM ressorts`
	row := m.DB.QueryRowContext(ctx, query)

	// handle null types
	var slack sql.NullString

	err := row.Scan(
		&ressort.ID,
		&ressort.RessortName,
		&slack,
		&ressort.CreatedOn,
		&ressort.UpdatedOn,
	)

	if err != nil {
		return ressort, err
	}

	// second part of null types

	if slack.Valid {
		ressort.SlackChannel = slack.String
	}

	return ressort, nil
}

// AllRessorts gets all ressorts
func (m *postgresDBRepo) AllRessorts() ([]models.Ressort, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ressorts []models.Ressort

	query := `SELECT id, ressort_name, slack_channel, created_on, updated_on FROM ressorts`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return ressorts, err
	}

	defer rows.Close()

	for rows.Next() {
		var r models.Ressort

		// handle null types
		var slack sql.NullString

		err := rows.Scan(
			&r.ID,
			&r.RessortName,
			&slack,
			&r.CreatedOn,
			&r.UpdatedOn,
		)

		if err != nil {
			return ressorts, err
		}

		// second part of null types
		if slack.Valid {
			r.SlackChannel = slack.String
		}

		ressorts = append(ressorts, r)
	}

	if err = rows.Err(); err != nil {
		return ressorts, err
	}

	return ressorts, nil
}

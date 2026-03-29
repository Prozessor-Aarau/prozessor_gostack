package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/lib/pq"
)

func (m *postgresDBRepo) CreateOpeningHours(openingHours []models.OpeningHour) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, oh := range openingHours {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO opening_hours (start_date, end_date) VALUES ($1, $2)",
			oh.StartTime,
			oh.EndTime,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AllOpeningHours returns all opening hours and their supervisors
func (m *postgresDBRepo) AllOpeningHours(startDate time.Time, dayCount int) ([]models.OpeningHour, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// calculate EndDate
	endDate := startDate.AddDate(0, 0, dayCount)

	var openingHours []models.OpeningHour

	query := `
		SELECT
			oh.id,
			oh.start_date,
			oh.end_date,
			array_agg(vis.id) AS sup_id,
			array_agg(vis.first_name) AS sup_first_name,
			array_agg(vis.family_name) AS sup_family_name
		FROM
			opening_hours oh
			LEFT JOIN opening_hours_supervision sup ON oh.id = sup.opening_hour_id
			LEFT JOIN visitors vis ON sup.visitor_id = vis.id
		WHERE
			DATE(oh.start_date) >= DATE ($1)
			AND DATE(oh.start_date) <= DATE($2)
		GROUP BY
			oh.id
		ORDER BY
			oh.start_date ASC
	`
	rows, err := m.DB.QueryContext(ctx, query, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return openingHours, err
	}

	defer rows.Close()

	for rows.Next() {
		var oh models.OpeningHour

		var supId []sql.NullString
		var supFirstName []sql.NullString
		var supFamilyName []sql.NullString

		err := rows.Scan(
			&oh.ID,
			&oh.StartTime,
			&oh.EndTime,
			pq.Array(&supId),
			pq.Array(&supFirstName),
			pq.Array(&supFamilyName),
		)
		if err != nil {
			return openingHours, err
		}

		if supId[0].Valid {
			for i := range supId {
				oh.Supervisors = append(oh.Supervisors, models.Visitor{
					ID:         supId[i].String,
					FirstName:  supFirstName[i].String,
					FamilyName: supFamilyName[i].String,
				})
			}
		}

		openingHours = append(openingHours, oh)
	}

	if err = rows.Err(); err != nil {
		return openingHours, err
	}

	return openingHours, nil
}

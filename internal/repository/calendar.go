package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/lib/pq"
)

// CalendarOpeningHours returns all opening hours and the available expertise
func (m *postgresDBRepo) CalendarOpeningHours(startDate time.Time, dayCount int) ([]models.OpeningHour, error) {
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
			array_agg(DISTINCT a.atelier_name) FILTER (WHERE a.atelier_name IS NOT NULL)
		FROM
			opening_hours oh
			LEFT JOIN opening_hours_supervision sup ON oh.id = sup.opening_hour_id
			LEFT JOIN visitors vis ON sup.visitor_id = vis.id
			LEFT JOIN visitor_experts vex ON vex.visitor_id = vis.id
			LEFT JOIN ateliers a ON a.id = vex.atelier_id
		WHERE
			oh.start_date >= $1::timestamptz
			AND oh.start_date <= $2::timestamptz
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

		var expertise []sql.NullString

		err := rows.Scan(
			&oh.ID,
			&oh.StartTime,
			&oh.EndTime,
			pq.Array(&expertise),
		)
		if err != nil {
			return openingHours, err
		}

		if len(expertise) > 0 {
			for _, e := range expertise {
				if e.Valid {
					oh.Expertise = append(oh.Expertise, e.String)
				}
			}
		}

		openingHours = append(openingHours, oh)
	}

	if err = rows.Err(); err != nil {
		return openingHours, err
	}

	return openingHours, nil
}

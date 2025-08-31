package repository

import (
	"context"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

func (m *postgresDBRepo) GetStatisticsVisitsByRange() ([]models.ChartVisitCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visits []models.ChartVisitCount

	startDate, _ := time.Parse("2006-01-02", "2024-04-01")
	endDate, _ := time.Parse("2006-01-02", "2024-04-30")

	query := `
		SELECT
			DATE (date_in),
			COUNT(*) AS total,
			COUNT(*) FILTER (
				WHERE
					ressort_id IS NULL
					AND abonnement = false
			) AS daily,
			COUNT(*) FILTER (
				WHERE
					ressort_id IS NULL
					AND abonnement = true
			) AS abonnement,
			COUNT(*) FILTER (
				WHERE
					ressort_id IS NOT NULL
			) AS member
		FROM
			visits
		WHERE
			date_in BETWEEN $1 AND $2
		GROUP BY
			DATE (date_in)
		ORDER BY
			DATE (date_in) ASC;
	`

	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return visits, err
	}

	defer rows.Close()

	for rows.Next() {
		var v models.ChartVisitCount

		err := rows.Scan(
			&v.Date,
			&v.Total,
			&v.Daily,
			&v.Abonnement,
			&v.Member,
		)

		if err != nil {
			return visits, err
		}

		visits = append(visits, v)
	}

	if err = rows.Err(); err != nil {
		return visits, err
	}

	return visits, nil
}

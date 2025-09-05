package repository

import (
	"context"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

func (m *postgresDBRepo) GetStatisticsVisitsByRange() (models.ChartData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visits models.ChartData

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

	// name Y-Axis

	visits.AxisY = []string{"Tagesnutzung", "Abonnement", "Mitglied"}

	var arrDaily []int
	var arrAbo []int
	var arrMember []int

	for rows.Next() {
		var date time.Time
		var total int
		var daily int
		var abonnement int
		var member int

		err := rows.Scan(
			&date,
			&total,
			&daily,
			&abonnement,
			&member,
		)

		if err != nil {
			return visits, err
		}

		visits.AxisX = append(visits.AxisX, date.Format("02.01.2006"))
		arrDaily = append(arrDaily, daily)
		arrAbo = append(arrAbo, abonnement)
		arrMember = append(arrMember, member)
	}

	visits.Data = append(visits.Data, arrDaily)
	visits.Data = append(visits.Data, arrAbo)
	visits.Data = append(visits.Data, arrMember)

	if err = rows.Err(); err != nil {
		return visits, err
	}

	return visits, nil
}

package repository

import (
	"context"
	"math"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// GetStatisticsVisitsByRange gets the number of checkins per date by type
func (m *postgresDBRepo) GetStatisticsVisitsByType(startDate, endDate time.Time) (models.ChartData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visits models.ChartData

	query := `
		SELECT
			DATE (date_in),
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

	var arrDaily []float64
	var arrAbo []float64
	var arrMember []float64

	for rows.Next() {
		var date time.Time
		var daily float64
		var abonnement float64
		var member float64

		err := rows.Scan(
			&date,
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

// GetStatisticsAverageWeekday gets the average checkins for every day of the week
func (m *postgresDBRepo) GetStatisticsAverageWeekday(startDate, endDate time.Time) (models.ChartData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visits models.ChartData

	query := `
	SELECT
		TO_CHAR(DATE_TRUNC('day', date_in), 'Day') as day_name,
		AVG(daily_count) as avg_checkins_per_day
	FROM
		(
			SELECT
				DATE_TRUNC('day', date_in) as date_in,
				COUNT(*) as daily_count
			FROM
				visits
			WHERE
				date_in BETWEEN $1 and $2
			GROUP BY
				DATE_TRUNC('day', date_in)
		) daily_counts
	GROUP BY
		EXTRACT(
			DOW
			FROM
				date_in
		),
		TO_CHAR(DATE_TRUNC('day', date_in), 'Day')
	ORDER BY
		EXTRACT(
			DOW
			FROM
				date_in
		);
	`

	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return visits, err
	}

	defer rows.Close()

	// name Y-Axis

	visits.AxisY = []string{"Durchschn. Besuche"}

	var arrAvg []float64

	for rows.Next() {
		var weekday string
		var average float64

		err := rows.Scan(
			&weekday,
			&average,
		)

		if err != nil {
			return visits, err
		}

		visits.AxisX = append(visits.AxisX, weekday)
		arrAvg = append(arrAvg, math.Round(average*100)/100)
	}

	visits.Data = append(visits.Data, arrAvg)

	if err = rows.Err(); err != nil {
		return visits, err
	}

	return visits, nil
}

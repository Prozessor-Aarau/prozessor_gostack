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
			date_in BETWEEN $3 AND $2
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
	visits.AxisY = []string{"Mitglied", "Abonnement", "Tagesnutzung"}

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

	visits.Data = append(visits.Data, arrMember)
	visits.Data = append(visits.Data, arrAbo)
	visits.Data = append(visits.Data, arrDaily)

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
		AVG(daily_count) as avg_daily,
		AVG(daily_abonnement) as avg_abonnement,
		AVG(daily_member) as avg_member
	FROM
		(
			SELECT
				DATE_TRUNC('day', date_in) as date_in,
				COUNT(*) FILTER (
					WHERE
						ressort_id IS NULL
						AND abonnement = false
				) as daily_count,
				COUNT(*) FILTER (
					WHERE
						ressort_id IS NULL
						AND abonnement = true
				) as daily_abonnement,
				COUNT(*) FILTER (
					WHERE
						ressort_id IS NOT NULL
				) as daily_member
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
	visits.AxisY = []string{"Mitglied", "Abonnement", "Tagesnutzung"}

	var arrDaily []float64
	var arrAbo []float64
	var arrMember []float64

	for rows.Next() {
		var weekday string
		var daily float64
		var abo float64
		var member float64

		err := rows.Scan(
			&weekday,
			&daily,
			&abo,
			&member,
		)

		if err != nil {
			return visits, err
		}

		visits.AxisX = append(visits.AxisX, weekday)
		arrDaily = append(arrDaily, math.Round(daily*100)/100)
		arrAbo = append(arrAbo, math.Round(abo*100)/100)
		arrMember = append(arrMember, math.Round(member*100)/100)
	}

	visits.Data = append(visits.Data, arrMember)
	visits.Data = append(visits.Data, arrAbo)
	visits.Data = append(visits.Data, arrDaily)

	if err = rows.Err(); err != nil {
		return visits, err
	}

	return visits, nil
}

// GetStatisticsAteliers gets the usage number for all active ateliers
func (m *postgresDBRepo) GetStatisticsAteliers(startDate, endDate time.Time) (models.ChartData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visits models.ChartData

	query := `
	SELECT
		a.atelier_name,
		count(*)
	FROM
		visit_ateliers va
		JOIN ateliers a ON a.id = va.atelier_id
		JOIN visits v ON v.id = va.visit_id
	WHERE
		v.date_in BETWEEN $1 AND $2
		AND a.active = true
	GROUP BY
		a.atelier_name;
	`

	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return visits, err
	}

	defer rows.Close()

	// name Y-Axis

	visits.AxisY = []string{"Besuche"}

	var arrAvg []float64

	for rows.Next() {
		var atelier string
		var count float64

		err := rows.Scan(
			&atelier,
			&count,
		)

		if err != nil {
			return visits, err
		}

		visits.AxisX = append(visits.AxisX, atelier)
		arrAvg = append(arrAvg, count)
	}

	visits.Data = append(visits.Data, arrAvg)

	if err = rows.Err(); err != nil {
		return visits, err
	}

	return visits, nil
}

// GetStatisticsAges gets the ages of the visitors by visits
func (m *postgresDBRepo) GetStatisticsAges(startDate, endDate time.Time) (models.ChartData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ages models.ChartData

	query := `
	WITH
		visitor_ages AS (
			SELECT
				EXTRACT(
					YEAR
					FROM
						AGE (CURRENT_DATE, vi.birthdate)
				) AS age
			FROM
				visits v
				JOIN visitors vi ON vi.id = v.visitor_id
			WHERE
				v.date_in BETWEEN $1 AND $2
		)
	SELECT
		ROUND(
			COUNT(*) FILTER (
				WHERE
					age < 20
			) * 100.0 / COUNT(*),
			2
		) AS "-20",
		ROUND(
			COUNT(*) FILTER (
				WHERE
					age BETWEEN 21 AND 30
			) * 100.0 / COUNT(*),
			2
		) AS "20-30",
		ROUND(
			COUNT(*) FILTER (
				WHERE
					age BETWEEN 31 AND 40
			) * 100.0 / COUNT(*),
			2
		) AS "31-40",
		ROUND(
			COUNT(*) FILTER (
				WHERE
					age BETWEEN 41 AND 50
			) * 100.0 / COUNT(*),
			2
		) AS "41-50",
		ROUND(
			COUNT(*) FILTER (
				WHERE
					age BETWEEN 51 AND 65
			) * 100.0 / COUNT(*),
			2
		) AS "51-65",
		ROUND(
			COUNT(*) FILTER (
				WHERE
					age > 65
			) * 100.0 / COUNT(*),
			2
		) AS "65+"
	FROM
		visitor_ages;	
	`

	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return ages, err
	}

	defer rows.Close()

	// name Y-Axis

	ages.AxisY = []string{"Alter"}
	ages.AxisX = []string{"<20", "21-30", "31-40", "41-50", "51-65", ">65"}

	var arrAges [6]float64

	for rows.Next() {

		err := rows.Scan(
			&arrAges[0],
			&arrAges[1],
			&arrAges[2],
			&arrAges[3],
			&arrAges[4],
			&arrAges[5],
		)

		if err != nil {
			return ages, err
		}
	}

	ages.Data = append(ages.Data, arrAges[:])

	if err = rows.Err(); err != nil {
		return ages, err
	}

	return ages, nil
}

// GetStatisticsKeyFigures gets the key figures of the visitors and visits
func (m *postgresDBRepo) GetStatisticsKeyFigures(startDate, endDate time.Time) (models.ChartData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var figures models.ChartData

	query := `
	WITH
		active_visitors AS (
			SELECT
				v.*,
				(
					a.buydate + interval '1 year' BETWEEN $1 AND $2
					OR a.buydate BETWEEN $1 AND $2
				) AS abo_active
			FROM
				visitors v
				LEFT JOIN abonnements a ON a.visitor_id = v.id
		)
	SELECT
		(
			SELECT
				COUNT(*)
			FROM
				visits
			WHERE
				visits.date_in BETWEEN $1 AND $2
		) visits,
		COUNT(*) FILTER (
			WHERE
				created_on <= $2
		) AS total_registered,
		COUNT(*) FILTER (
			WHERE
				created_on BETWEEN $1 AND $2 
		) AS new_registered,
		COUNT(*) FILTER (
			WHERE
				abo_active
				AND ressort_id IS NULL
		) AS active_abonnements,
		COUNT(*) FILTER (
			WHERE
				abo_active
				AND ressort_id IS NOT NULL
		) AS active_members
	FROM
		active_visitors;
	`

	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return figures, err
	}

	defer rows.Close()

	// name Y-Axis

	figures.AxisY = []string{"Kennzahlen"}
	figures.AxisX = []string{"Besuche", "Total Registrationen", "neue Registrationen", "gültige Abonnements", "aktive Mitglieder"}

	var arrFigures [5]float64

	for rows.Next() {

		err := rows.Scan(
			&arrFigures[0],
			&arrFigures[1],
			&arrFigures[2],
			&arrFigures[3],
			&arrFigures[4],
		)

		if err != nil {
			return figures, err
		}
	}

	figures.Data = append(figures.Data, arrFigures[:])

	if err = rows.Err(); err != nil {
		return figures, err
	}

	return figures, nil
}

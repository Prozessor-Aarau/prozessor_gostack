package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// CreateReport puts a new report into the database
func (m *postgresDBRepo) CreateReport(r models.Report) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int

	stmt := `INSERT INTO reports (ressort_id, machine_id, fatal, report_info, report_reporter, created_on, updated_on)
  VALUES($1, $2, $3, $4, $5, $6, $7)
  RETURNING id`

	err := m.DB.QueryRowContext(ctx, stmt,
		r.RessortID,
		r.MachineID,
		r.Fatal,
		r.ReportInfo,
		NewNullString(r.ReportReporter),
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// DeleteReport deletes a report by id
func (m *postgresDBRepo) DeleteReport(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "DELETE FROM reports WHERE id = $1"

	_, err := m.DB.ExecContext(ctx, stmt, id)

	if err != nil {
		return err
	}

	return nil
}

// UpdateReportStatus updates the solved status by id and status
func (m *postgresDBRepo) UpdateReportStatus(status bool, id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "UPDATE reports SET solved = $1 WHERE id = $2"

	_, err := m.DB.ExecContext(ctx, stmt, status, id)

	if err != nil {
		return err
	}

	return nil
}

// ReportsByMachineID returns all reports belonging to a machine
func (m *postgresDBRepo) ReportsByMachineID(id string) ([]models.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reports []models.Report

	query := `SELECT id, fatal, report_info, report_reporter, solved, machine_id, updated_on
  FROM reports
  WHERE machine_id = $1`

	rows, err := m.DB.QueryContext(ctx, query, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var r models.Report

		var reporter sql.NullString

		err := rows.Scan(
			&r.ID,
			&r.Fatal,
			&r.ReportInfo,
			&reporter,
			&r.Solved,
			&r.MachineID,
			&r.UpdatedOn,
		)

		if err != nil {
			return nil, nil
		}
		if reporter.Valid {
			r.ReportReporter = reporter.String
		}

		reports = append(reports, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}

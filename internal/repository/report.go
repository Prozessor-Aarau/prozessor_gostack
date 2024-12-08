package repository

import (
	"context"
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

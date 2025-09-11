package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CreateBadge puts a new badge into the database
func (m *postgresDBRepo) CreateMachine(machine models.Machine) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	id, err := uuid.NewUUID()

	if err != nil {
		return "", err
	}

	//Convert id to string and set it on the user
	machine.ID = id.String()

	stmt := `INSERT INTO machines (id, ressort_id, machine_name, created_on, updated_on)
  VALUES($1, $2, $3, $4, $5)
  RETURNING id`

	_, err = m.DB.ExecContext(ctx, stmt,
		machine.ID,
		machine.RessortID,
		machine.MachineName,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return "", err
	}

	return machine.ID, nil
}

// UpdateMachine updates a machine in the db
func (m *postgresDBRepo) UpdateMachine(machine models.Machine) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE machines 
  SET machine_name = $1,
  ressort_id = $2,
  updated_on = $3 
  WHERE id = $4
  `

	_, err := m.DB.ExecContext(
		ctx,
		query,
		machine.MachineName,
		machine.RessortID,
		time.Now(),
		machine.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// MachineByID gets a badge by id
func (m *postgresDBRepo) MachineByID(id string) (models.Machine, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var machine models.Machine
	query := `SELECT m.id, 
    m.machine_name, 
    m.created_on, 
    m.updated_on, 
    r.id, 
    r.ressort_name, 
    r.slack_channel, 
    m.qr_code, 
    array_agg(rp.id) AS report_id, 
    array_agg(rp.fatal) AS report_fatal, 
    array_agg(rp.solved) AS report_solved, 
    array_agg(rp.report_info) AS report_info, 
    array_agg(rp.report_reporter) AS report_reporter,
    array_agg(DATE(rp.updated_on)) AS updated_on
  FROM machines m
  LEFT OUTER JOIN ressorts r ON r.id = m.ressort_id
  LEFT OUTER JOIN reports rp ON rp.machine_id = m.id
  WHERE m.id = $1
  GROUP BY m.id, r.id
  `
	row := m.DB.QueryRowContext(ctx, query, id)

	var qr sql.NullString

	// null slice vars for potential reports
	var reportID []sql.NullInt32
	var fatal []sql.NullBool
	var solved []sql.NullBool
	var info []sql.NullString
	var reporter []sql.NullString
	var updated []sql.NullString

	err := row.Scan(
		&machine.ID,
		&machine.MachineName,
		&machine.CreatedOn,
		&machine.UpdatedOn,
		&machine.RessortID,
		&machine.RessortName,
		&machine.SlackChannel,
		&qr,
		pq.Array(&reportID),
		pq.Array(&fatal),
		pq.Array(&solved),
		pq.Array(&info),
		pq.Array(&reporter),
		pq.Array(&updated),
	)

	if err != nil {
		return machine, err
	}

	if qr.Valid {
		machine.QR = qr.String
	}

	if reportID[0].Valid {
		for i := range reportID {
			var reportString string

			updatedOn, err := time.Parse("2006-01-02", updated[i].String)
			if err != nil {
				return models.Machine{}, err
			}

			if reporter[i].Valid {
				reportString = reporter[i].String
			}
			machine.Reports = append(machine.Reports, models.Report{
				ID:             int(reportID[i].Int32),
				MachineID:      machine.ID,
				Fatal:          fatal[i].Bool,
				Solved:         solved[i].Bool,
				ReportInfo:     info[i].String,
				ReportReporter: reportString,
				UpdatedOn:      updatedOn,
			})
		}

	}

	return machine, nil
}

// AllBadges gets all badges
func (m *postgresDBRepo) AllMachines(filter models.MachineFilter) ([]models.Machine, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var machines []models.Machine

	var count int

	// TODO: change filter so it ignores upper or lowercase
	search := "%" + filter.MachineName + "%"
	ressort := ""

	// check if there are filters and set where clauses
	if filter.RessortID > 0 {
		ressort = fmt.Sprintf("AND m.ressort_id = %d", filter.RessortID)
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM machines m
	WHERE m.machine_name LIKE $1
  %s
`,
		ressort,
	)

	// set count to 0
	row := m.DB.QueryRowContext(ctx, countQuery, search)

	err := row.Scan(&count)

	if err != nil {
		return machines, 0, err
	}
	query := fmt.Sprintf(
		`SELECT 
      m.id, 
      m.machine_name, 
      m.created_on, 
      m.updated_on, 
      r.id, 
      r.ressort_name, 
      r.slack_channel, 
      m.qr_code,
      array_agg(rp.id) AS report_id, 
      array_agg(rp.fatal) AS report_fatal, 
      array_agg(rp.solved) AS report_solved, 
      array_agg(DATE(rp.updated_on)) AS updated_on
    FROM machines m
    LEFT OUTER JOIN ressorts r ON r.id = m.ressort_id
    LEFT OUTER JOIN reports rp ON rp.machine_id = m.id
    WHERE m.machine_name LIKE $1
    %s
    GROUP BY 
    m.id, r.id
    ORDER BY 
    r.ressort_name ASC,
    m.machine_name ASC
  `,
		ressort,
	)

	rows, err := m.DB.QueryContext(ctx, query, search)
	if err != nil {
		return machines, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var machine models.Machine

		var qr sql.NullString
		var reportID []sql.NullInt32
		var fatal []sql.NullBool
		var solved []sql.NullBool
		var updated []sql.NullString

		err := rows.Scan(
			&machine.ID,
			&machine.MachineName,
			&machine.CreatedOn,
			&machine.UpdatedOn,
			&machine.RessortID,
			&machine.RessortName,
			&machine.SlackChannel,
			&qr,
			pq.Array(&reportID),
			pq.Array(&fatal),
			pq.Array(&solved),
			pq.Array(&updated),
		)

		if err != nil {
			return machines, 0, err
		}

		if qr.Valid {
			machine.QR = qr.String
		}

		if reportID[0].Valid {
			for i := range reportID {

				updatedOn, err := time.Parse("2006-01-02", updated[i].String)
				if err != nil {
					return machines, 0, err
				}

				machine.Reports = append(machine.Reports, models.Report{
					ID:        int(reportID[i].Int32),
					MachineID: machine.ID,
					Fatal:     fatal[i].Bool,
					Solved:    solved[i].Bool,
					UpdatedOn: updatedOn,
				})
			}
		}

		machines = append(machines, machine)
	}

	if err = rows.Err(); err != nil {
		return machines, 0, err
	}

	return machines, count, nil
}

// SetMachineQR sets a machines qr code in the db
func (m *postgresDBRepo) SetMachineQR(machine models.Machine) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE machines 
  SET qr_code = $1,
  updated_on = $2 
  WHERE id = $3
  `

	_, err := m.DB.ExecContext(
		ctx,
		query,
		machine.QR,
		time.Now(),
		machine.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// SetMachineQR sets a machines qr code in the db
func (m *postgresDBRepo) DeleteMachineQR(machine models.Machine) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE machines 
  SET qr_code = NULL,
  updated_on = $1 
  WHERE id = $2
  `

	_, err := m.DB.ExecContext(
		ctx,
		query,
		time.Now(),
		machine.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMachineById deletes a machine from the db
func (m *postgresDBRepo) DeleteMachineById(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "DELETE FROM machines WHERE id = $1"

	_, err := m.DB.ExecContext(
		ctx,
		query,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

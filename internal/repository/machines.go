package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/google/uuid"
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
	query := `SELECT m.id, m.machine_name, m.created_on, m.updated_on, r.id, r.ressort_name, r.slack_channel
  FROM machines m
  LEFT OUTER JOIN ressorts r ON r.id = m.ressort_id
  WHERE b.id = $1
  `
	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&machine.ID,
		&machine.MachineName,
		&machine.CreatedOn,
		&machine.UpdatedOn,
		&machine.RessortID,
		&machine.RessortName,
		&machine.SlackChannel,
	)

	if err != nil {
		return machine, err
	}

	return machine, nil
}

// AllBadges gets all badges
func (m *postgresDBRepo) AllMachines(filter models.MachineFilter) ([]models.Machine, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var machines []models.Machine

	var count int

	search := "%" + filter.MachineName + "%"
	ressort := ""

	// check if there are filters and set where clauses
	if filter.RessortID > 0 {
		ressort = "AND m.ressort_id = $2"
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
		`SELECT m.id, m.machine_name, m.created_on, m.updated_on, r.id, r.ressort_name, r.slack_channel
  FROM machines m
  LEFT OUTER JOIN ressorts r ON r.id = m.ressort_id
	WHERE m.machine_name LIKE $1
  %s
  ORDER BY 
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

		err := rows.Scan(
			&machine.ID,
			&machine.MachineName,
			&machine.CreatedOn,
			&machine.UpdatedOn,
			&machine.RessortID,
			&machine.RessortName,
			&machine.SlackChannel,
		)

		if err != nil {
			return machines, 0, err
		}

		machines = append(machines, machine)
	}

	if err = rows.Err(); err != nil {
		return machines, 0, err
	}

	return machines, count, nil
}

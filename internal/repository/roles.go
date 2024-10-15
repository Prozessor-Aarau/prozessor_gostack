package repository

import (
	"context"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// AllRoles gets all roles
func (m *postgresDBRepo) AllRoles() ([]models.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var roles []models.Role

	query := `SELECT id, role_name, created_on, updated_on FROM roles`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return roles, err
	}

	defer rows.Close()

	for rows.Next() {
		var r models.Role

		err := rows.Scan(
			&r.ID,
			&r.RoleName,
			&r.CreatedOn,
			&r.UpdatedOn,
		)

		if err != nil {
			return roles, err
		}
		roles = append(roles, r)
	}

	if err = rows.Err(); err != nil {
		return roles, err
	}

	return roles, nil
}

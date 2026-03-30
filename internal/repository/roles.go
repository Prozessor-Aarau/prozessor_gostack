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

// RoleByID returns a role by id
func (m *postgresDBRepo) RoleByID(id int) (models.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var role models.Role

	query := `SELECT id, 
            role_name, 
            created_on, 
            updated_on,
            view_visitors,
            edit_visitor,
            edit_visitor_membership,
            edit_visitor_role,
            view_machines,
            view_badges,
            assign_badge,
            edit_badge,
            view_abonnements,
            add_abonnement,
            edit_abonnement,
            view_reminders,
            add_reminder,
            view_newsletters,
            add_newsletter,
            view_logistics,
            edit_logistics,
            view_statistics,
            view_logs,
            view_finances,
            add_finances,
            edit_finances,
						edit_opening_hours,
						delete_opening_hours,
						view_supervisions,
						edit_own_supervisions,
						edit_all_supervisions
            FROM roles
            WHERE id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&role.ID,
		&role.RoleName,
		&role.CreatedOn,
		&role.UpdatedOn,
		&role.ViewVisitors,
		&role.EditVisitor,
		&role.EditVisitorMembership,
		&role.EditVisitorRole,
		&role.ViewMachines,
		&role.ViewBadges,
		&role.AssignBadge,
		&role.EditBadge,
		&role.ViewAbonnements,
		&role.AddAbonnement,
		&role.EditAbonnement,
		&role.ViewReminders,
		&role.AddReminder,
		&role.ViewNewsletters,
		&role.AddNewsletter,
		&role.ViewLogistics,
		&role.EditLogistics,
		&role.ViewStatistics,
		&role.ViewLogs,
		&role.ViewFinances,
		&role.AddFinances,
		&role.EditFinances,
		&role.EditOpeningHours,
		&role.DeleteOpeningHours,
		&role.ViewSupervisions,
		&role.EditOwnSupervisions,
		&role.EditAllSupervisions,
	)
	if err != nil {
		return role, err
	}

	return role, nil
}

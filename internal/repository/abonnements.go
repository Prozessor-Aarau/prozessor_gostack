package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// CreateAbonnement puts a new badge into the database
func (m *postgresDBRepo) CreateAbonnement(a models.Abonnement) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int

	stmt := `INSERT INTO abonnements (buydate, partial_payment, visitor_id, member, price, created_on, updated_on)
  VALUES($1, $2, $3, $4, $5, $6, $7)
  RETURNING id`

	err := m.DB.QueryRowContext(ctx, stmt,
		a.BuyDate,
		a.PartialPayment,
		a.VisitorID,
		a.Member,
		a.Price,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// DeleteAbonnement puts a new badge into the database
func (m *postgresDBRepo) DeleteAbonnement(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM abonnements WHERE id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id)

	if err != nil {
		return err
	}

	return nil
}

// AllAbonnements gets all abonnements filtered and returns also count
func (m *postgresDBRepo) AllAbonnements(
	page int,
	filter models.AbonnementFilter,
) ([]models.Abonnement, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var abonnements []models.Abonnement
	var count int

	search := "%" + filter.SearchString + "%"

	// set count to result
	row := m.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM abonnements a
  JOIN visitors v ON v.id = a.visitor_id
  WHERE (LOWER(v.first_name) LIKE LOWER($1)
  OR LOWER(v.family_name) LIKE LOWER($1))
`, search)

	err := row.Scan(&count)

	if err != nil {
		return abonnements, 0, err
	}

	offset := 10 * (page - 1)

	query := `SELECT a.id, a.buydate, a.partial_payment, a.reminder, a.member, a.price, a.created_on, a.updated_on, v.id, v.first_name, v.family_name
  FROM abonnements a
  JOIN visitors v ON v.id = a.visitor_id
  WHERE (LOWER(v.first_name) LIKE LOWER($1)
  OR LOWER(v.family_name) LIKE LOWER($1))
  ORDER BY 
   a.buydate DESC
  OFFSET $2
  LIMIT 10
  `
	rows, err := m.DB.QueryContext(ctx, query, search, offset)
	if err != nil {
		return abonnements, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var a models.Abonnement

		// start handling null types
		var reminder sql.NullTime
		var visitorID sql.NullString
		var visitorFirstName sql.NullString
		var visitorFamilyName sql.NullString
		var buydate sql.NullTime

		err := rows.Scan(
			&a.ID,
			&buydate,
			&a.PartialPayment,
			&reminder,
			&a.Member,
			&a.Price,
			&a.CreatedOn,
			&a.UpdatedOn,
			&visitorID,
			&visitorFirstName,
			&visitorFamilyName,
		)

		// handling null types
		if reminder.Valid {
			a.Reminder = reminder.Time
		}
		if visitorID.Valid {
			a.VisitorID = visitorID.String
		}
		if visitorFirstName.Valid {
			a.VisitorFirstName = visitorFirstName.String
		}
		if visitorFamilyName.Valid {
			a.VisitorFamilyName = visitorFamilyName.String
		}
		if buydate.Valid {
			a.BuyDate = buydate.Time.Format("2006-01-02")
		}

		if err != nil {
			return abonnements, 0, err
		}
		abonnements = append(abonnements, a)
	}

	if err = rows.Err(); err != nil {
		return abonnements, 0, err
	}

	return abonnements, count, nil
}

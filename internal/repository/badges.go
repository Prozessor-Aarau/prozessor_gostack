package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// CreateBadge puts a new badge into the database
func (m *postgresDBRepo) CreateBadge(b models.Badge) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int

	stmt := `INSERT INTO badges (number, uid)
  VALUES($1, $2)
  RETURNING id`

	err := m.DB.QueryRowContext(ctx, stmt,
		b.Number,
		b.UID,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// DeleteBadge puts a new badge into the database
func (m *postgresDBRepo) DeleteBadge(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM badges WHERE id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, id)

	if err != nil {
		return err
	}

	return nil
}

// UpdateBadge assigns a badge to a visitor or changes its values
func (m *postgresDBRepo) UpdateBadge(b models.Badge) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE badges 
  SET number = $1,
  uid = $2,
  visitor_id = $3, 
  extern_first_name = $4, 
  extern_family_name = $5, 
  extern_city = $6, 
  extern_phone = $7, 
  extern_email = $8, 
  extern_purpose = $9
  WHERE id = $10
  `

	_, err := m.DB.ExecContext(
		ctx,
		query,
		b.Number,
		b.UID,
		NewNullString(b.VisitorID),
		NewNullString(b.ExternFirstName),
		NewNullString(b.ExternFamilyName),
		NewNullString(b.ExternCity),
		NewNullString(b.ExternPhone),
		NewNullString(b.ExternEmail),
		NewNullString(b.ExternPurpose),
		b.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// AllBadges gets all badges
func (m *postgresDBRepo) AllBadges(filter models.BadgeFilter) ([]models.Badge, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var badges []models.Badge

	var count int

	search := "%" + filter.Number + "%"
	inUse := ""
	typ := ""

	// check if there are filters and set where clauses
	switch filter.InUse {

	case "gone":
		inUse = "AND (b.visitor_id IS NOT NULL OR b.extern_first_name IS NOT NULL)"

	case "here":
		inUse = "AND (b.visitor_id IS NULL AND b.extern_fist_name IS NULL)"
	}

	switch filter.Type {

	case "new":
		typ = "AND b.uid != 'old_badge'"

	case "old":
		typ = "AND b.uid = 'old_badge'"
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*)  FROM badges b
	WHERE CAST(b.number AS TEXT) LIKE $1
  %s
  %s
`,
		inUse,
		typ,
	)

	// set count to 0
	row := m.DB.QueryRowContext(ctx, countQuery, search)

	err := row.Scan(&count)

	if err != nil {
		return badges, 0, err
	}
	query := fmt.Sprintf(
		`SELECT b.id, b.number, b.uid, b.created_on, b.updated_on, v.id, v.first_name, v.family_name, b.extern_first_name, b.extern_family_name, b.extern_city, b.extern_phone, b.extern_email, b.extern_purpose,
   (case when a.buydate + INTERVAL '1 year' >= CURRENT_DATE
    then 1 
    else 
    (case when exists 
      (SELECT buydate FROM abonnements WHERE visitor_id = v.id) then 0 else NULL end)
    end) as abo_valid
  FROM badges b
  LEFT OUTER JOIN visitors v ON v.id = b.visitor_id
  LEFT JOIN LATERAL 
    (SELECT * FROM abonnements 
    WHERE v.id = abonnements.visitor_id 
    ORDER BY buydate DESC LIMIT 1) 
    as a ON v.id = a.visitor_id
	WHERE CAST(b.number AS TEXT) LIKE $1
  %s
  %s
  ORDER BY 
  b.number ASC
  `,
		inUse,
		typ,
	)

	rows, err := m.DB.QueryContext(ctx, query, search)
	if err != nil {
		return badges, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var b models.Badge

		// start handling null types
		var visID sql.NullString
		var visFirstName sql.NullString
		var visFamilyName sql.NullString
		var extFirstName sql.NullString
		var extFamilyName sql.NullString
		var extCity sql.NullString
		var extPhone sql.NullString
		var extEmail sql.NullString
		var extPurpose sql.NullString
		var aboValid sql.NullBool

		err := rows.Scan(
			&b.ID,
			&b.Number,
			&b.UID,
			&b.CreatedOn,
			&b.UpdatedOn,
			&visID,
			&visFirstName,
			&visFamilyName,
			&extFirstName,
			&extFamilyName,
			&extCity,
			&extPhone,
			&extEmail,
			&extPurpose,
			&aboValid,
		)

		if err != nil {
			return badges, 0, err
		}

		// handling null types
		if visID.Valid {
			b.VisitorID = visID.String
		}
		if visFirstName.Valid {
			b.VisitorFirstName = visFirstName.String
		}
		if visFamilyName.Valid {
			b.VisitorFamilyName = visFamilyName.String
		}
		if extFirstName.Valid {
			b.ExternFirstName = extFirstName.String
		}
		if extFamilyName.Valid {
			b.ExternFamilyName = extFamilyName.String
		}
		if extCity.Valid {
			b.ExternCity = extCity.String
		}
		if extPhone.Valid {
			b.ExternPhone = extPhone.String
		}
		if extEmail.Valid {
			b.ExternEmail = extEmail.String
		}
		if extPurpose.Valid {
			b.ExternPurpose = extPurpose.String
		}
		if aboValid.Valid {
			b.VisitorAboValid = aboValid.Bool
		}

		badges = append(badges, b)
	}

	if err = rows.Err(); err != nil {
		return badges, 0, err
	}

	return badges, count, nil
}

// BadgeByID gets a badge by id
func (m *postgresDBRepo) BadgeByID(id int) (models.Badge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var badge models.Badge
	query := `SELECT b.id, b.number, b.uid, b.created_on, b.updated_on, v.id, v.first_name, v.family_name, b.extern_first_name, b.extern_family_name, b.extern_city, b.extern_phone, b.extern_email, b.extern_purpose,
   (case when a.buydate + INTERVAL '1 year' >= CURRENT_DATE
    then 1 
    else 
    (case when exists 
      (SELECT buydate FROM abonnements WHERE visitor_id = v.id) then 0 else NULL end)
    end) as abo_valid
  FROM badges b
  LEFT OUTER JOIN visitors v ON v.id = b.visitor_id
  LEFT OUTER JOIN abonnements a ON a.visitor_id = v.id
  WHERE b.id = $1
  `

	rows := m.DB.QueryRowContext(ctx, query, id)

	// start handling null types
	var visID sql.NullString
	var visFirstName sql.NullString
	var visFamilyName sql.NullString
	var extFirstName sql.NullString
	var extFamilyName sql.NullString
	var extCity sql.NullString
	var extPhone sql.NullString
	var extEmail sql.NullString
	var extPurpose sql.NullString
	var aboValid sql.NullBool

	err := rows.Scan(
		&badge.ID,
		&badge.Number,
		&badge.UID,
		&badge.CreatedOn,
		&badge.UpdatedOn,
		&visID,
		&visFirstName,
		&visFamilyName,
		&extFirstName,
		&extFamilyName,
		&extCity,
		&extPhone,
		&extEmail,
		&extPurpose,
		&aboValid,
	)

	if err != nil {
		return badge, err
	}

	// handling null types
	if visID.Valid {
		badge.VisitorID = visID.String
	}
	if visFirstName.Valid {
		badge.VisitorFirstName = visFirstName.String
	}
	if visFamilyName.Valid {
		badge.VisitorFamilyName = visFamilyName.String
	}
	if extFirstName.Valid {
		badge.ExternFirstName = extFirstName.String
	}
	if extFamilyName.Valid {
		badge.ExternFamilyName = extFamilyName.String
	}
	if extCity.Valid {
		badge.ExternCity = extCity.String
	}
	if extPhone.Valid {
		badge.ExternPhone = extPhone.String
	}
	if extEmail.Valid {
		badge.ExternEmail = extEmail.String
	}
	if extPurpose.Valid {
		badge.ExternPurpose = extPurpose.String
	}
	if aboValid.Valid {
		badge.VisitorAboValid = aboValid.Bool
	}

	return badge, nil
}

// ResetBadge badge removes a visitor from a badge
func (m *postgresDBRepo) ResetBadge(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE badges 
  SET visitor_id = NULL, 
  extern_first_name = NULL, 
  extern_family_name = NULL, 
  extern_city = NULL, 
  extern_phone = NULL, 
  extern_email = NULL, 
  extern_purpose = NULL
  WHERE id = $1
  `

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

func (m *postgresDBRepo) UniqueBadge(number int, uid string) (bool, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var bNum int
	var bUID int

	queryNum := `SELECT COUNT(*) FROM badges WHERE number = $1 `
	queryUID := `SELECT COUNT(*) FROM badges WHERE uid = $1 `

	err := m.DB.QueryRowContext(ctx, queryNum,
		number,
	).Scan(&bNum)

	if err != nil {
		return false, false, err
	}

	err = m.DB.QueryRowContext(ctx, queryUID,
		uid,
	).Scan(&bUID)

	if err != nil {
		return false, false, err
	}

	boolNum := bNum < 1
	boolUID := bUID < 1

	return boolNum, boolUID, nil
}

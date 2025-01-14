package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// create a new visitor and return the new ID
func (m *postgresDBRepo) CreateVisitor(v models.Visitor) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	id, err := uuid.NewUUID()

	if err != nil {
		return "", err
	}

	//Convert id to string and set it on the user
	v.ID = id.String()

	stmt := `INSERT INTO visitors (id, first_name, family_name, birthdate, phone, email, address, postcode, city, infos, newsletter, created_on, updated_on)
  VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err = m.DB.ExecContext(ctx, stmt,
		v.ID,
		v.FirstName,
		v.FamilyName,
		v.Birthdate,
		NewNullString(v.Phone),
		NewNullString(strings.ToLower(v.Email)),
		v.Address,
		v.Postcode,
		v.City,
		NewNullString(v.Infos),
		v.Newsletter,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return "", err
	}

	return v.ID, nil
}

// GetAllVisitors gets all visitors and returns also count
func (m *postgresDBRepo) AllVisitors(
	// TODO if someone has more than 1 badge, that visitor will be more than once in the list
	// TODO if someone has more than 1 abonnement, that visitor will be more than once in the list
	page int,
	limit int,
	filter models.VisitorFilter,
) ([]models.Visitor, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visitors []models.Visitor
	var count int

	// set limit to very high if limit is 0, so there is no limit
	if limit == 0 {
		limit = 1000000
	}

	search := "%" + filter.SearchString + "%"
	buydate := ""
	badges := ""
	members := ""
	partners := ""

	// check if there are filters and set where clauses
	if filter.Abo {
		// check if validity is set
		switch filter.AboValidity {

		case "valid":
			buydate = "AND buydate + INTERVAL '1 year' >= CURRENT_DATE"

		case "invalid":
			buydate = "AND buydate + INTERVAL '1 year' < CURRENT_DATE"

		default:
			buydate = "AND buydate IS NOT NULL"
		}
	}

	if filter.Badge {
		badges = "AND b.number IS NOT NULL"
	}

	if filter.Member {
		members = "AND ressort_id IS NOT NULL"
	}

	if filter.Partner {
		partners = "AND partner_id IS NOT NULL"
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM visitors v
  LEFT OUTER JOIN abonnements a ON a.visitor_id = v.id
  LEFT OUTER JOIN badges b ON b.visitor_id = v.id
  LEFT OUTER JOIN visitor_partners vp ON vp.visitor_id = v.id
  WHERE (LOWER(first_name) LIKE LOWER($1)
  OR LOWER(family_name) LIKE LOWER($1) 
  OR LOWER(city) LIKE LOWER($1))
  %s
  %s
  %s
  %s
`,
		buydate,
		badges,
		partners,
		members,
	)

	// set count to 0
	row := m.DB.QueryRowContext(ctx, countQuery, search)

	err := row.Scan(&count)

	if err != nil {
		return visitors, count, err
	}

	offset := 10 * (page - 1)

	query := fmt.Sprintf(
		`SELECT v.id, first_name, family_name, birthdate, phone, v.email, v.address, v.postcode, v.city, infos, newsletter, ressort_id, v.created_on, v.updated_on, a.id, buydate, b.number, b.id, array_agg(partner_id) as partners, array_agg(partner_name) as partner_names
  FROM visitors v
  LEFT JOIN LATERAL 
    (SELECT * FROM abonnements 
    WHERE v.id = abonnements.visitor_id 
    ORDER BY buydate DESC LIMIT 1) 
    as a ON v.id = a.visitor_id
  LEFT OUTER JOIN badges b ON b.visitor_id = v.id
  LEFT OUTER JOIN visitor_partners vp ON vp.visitor_id = v.id
  LEFT OUTER JOIN partners p ON p.id = vp.partner_id
  WHERE (LOWER(first_name) LIKE LOWER($1)
  OR LOWER(family_name) LIKE LOWER($1) 
  OR LOWER(v.city) LIKE LOWER($1))
  %s
  %s
  %s
  %s
   GROUP BY v.id, a.id, a.buydate, b.number, b.id
  ORDER BY 
   %s %s
  OFFSET $2
  LIMIT $3  
  `,
		buydate,
		badges,
		partners,
		members,
		filter.Sort,
		filter.SortDir,
	)

	rows, err := m.DB.QueryContext(ctx, query, search, offset, limit)
	if err != nil {
		return visitors, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var v models.Visitor

		// start handling null types
		var phone sql.NullString
		var email sql.NullString
		var infos sql.NullString
		var ressortID sql.NullInt32
		var buydate sql.NullTime
		var aboID sql.NullInt32
		var badge sql.NullInt32
		var badgeID sql.NullInt32

		// create empty array for partners and partner_name
		var partners []sql.NullInt64
		var partnerNames []sql.NullString

		err := rows.Scan(
			&v.ID,
			&v.FirstName,
			&v.FamilyName,
			&v.Birthdate,
			&phone,
			&email,
			&v.Address,
			&v.Postcode,
			&v.City,
			&infos,
			&v.Newsletter,
			&ressortID,
			&v.CreatedOn,
			&v.UpdatedOn,
			&aboID,
			&buydate,
			&badge,
			&badgeID,
			pq.Array(&partners),
			pq.Array(&partnerNames),
		)

		// handling null types
		if phone.Valid {
			v.Phone = phone.String
		}
		if email.Valid {
			v.Email = email.String
		}
		if infos.Valid {
			v.Infos = infos.String
		}
		if ressortID.Valid {
			v.Ressort.ID = int(ressortID.Int32)
		}
		if aboID.Valid {
			v.AboLast.ID = int(aboID.Int32)
		}
		if buydate.Valid {
			v.AboLast.BuyDate = buydate.Time.Format("2006-01-02")
		}
		if badge.Valid {
			v.Badge.Number = int(badge.Int32)
		}
		if badgeID.Valid {
			v.Badge.ID = int(badgeID.Int32)
		}

		// get partners and push each id to partner model and slice of partners
		var pSlice []models.Partner
		for i, v := range partners {
			var p models.Partner
			if v.Valid {
				p.ID = int(v.Int64)
				p.PartnerName = partnerNames[i].String
				pSlice = append(pSlice, p)
			}
		}
		if len(pSlice) > 0 {
			v.Partner = pSlice
		}

		if err != nil {
			return visitors, 0, err
		}
		visitors = append(visitors, v)
	}

	if err = rows.Err(); err != nil {
		return visitors, count, err
	}

	return visitors, count, nil
}

// VisitorByID returns a user from the database
func (m *postgresDBRepo) VisitorByID(id string) (models.Visitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT v.id, first_name, family_name, birthdate, v.phone, v.email, v.address, v.postcode, v.city, infos, newsletter, ressort_id, role_id, t.ressort_name, v.created_on, v.updated_on, a.id, a.buydate, b.number, b.id, password, avatar, array_agg(p.id) AS partners, array_agg(p.partner_name) AS partner_names
  FROM visitors v
  LEFT JOIN LATERAL 
    (SELECT * FROM abonnements 
    WHERE v.id = abonnements.visitor_id 
    ORDER BY buydate DESC LIMIT 1) 
    as a ON v.id = a.visitor_id
  LEFT OUTER JOIN badges b ON b.visitor_id = v.id
  LEFT OUTER JOIN ressorts t ON t.id = v.ressort_id
  LEFT OUTER JOIN visitor_partners vp ON vp.visitor_id = v.id 
  LEFT OUTER JOIN partners p ON p.id = vp.partner_id
  WHERE v.id = $1
  GROUP BY v.id, t.ressort_name, a.id, a.buydate, b.number, b.id`

	row := m.DB.QueryRowContext(ctx, query, id)

	var v models.Visitor

	// start handling null types
	var phone sql.NullString
	var email sql.NullString
	var infos sql.NullString
	var ressortID sql.NullInt32
	var roleID sql.NullInt32
	var ressortName sql.NullString
	var buydate sql.NullTime
	var aboID sql.NullInt32
	var badge sql.NullInt32
	var badgeID sql.NullInt32
	var password sql.NullString
	var avatar sql.NullString

	// create empty array for partners and partner_name
	var partners []sql.NullInt64
	var partnerNames []sql.NullString

	err := row.Scan(
		&v.ID,
		&v.FirstName,
		&v.FamilyName,
		&v.Birthdate,
		&phone,
		&email,
		&v.Address,
		&v.Postcode,
		&v.City,
		&infos,
		&v.Newsletter,
		&ressortID,
		&roleID,
		&ressortName,
		&v.CreatedOn,
		&v.UpdatedOn,
		&aboID,
		&buydate,
		&badge,
		&badgeID,
		&password,
		&avatar,
		pq.Array(&partners),
		pq.Array(&partnerNames),
	)

	if err != nil {
		return v, err
	}

	// handling null types
	if phone.Valid {
		v.Phone = phone.String
	}
	if email.Valid {
		v.Email = email.String
	}
	if infos.Valid {
		v.Infos = infos.String
	}
	if ressortID.Valid {
		v.Ressort.ID = int(ressortID.Int32)
	}
	if roleID.Valid {
		v.RoleID = int(roleID.Int32)
	}
	if ressortName.Valid {
		v.Ressort.RessortName = ressortName.String
	}
	if aboID.Valid {
		v.AboLast.ID = int(aboID.Int32)
	}
	if buydate.Valid {
		v.AboLast.BuyDate = buydate.Time.Format("2006-01-02")
	}
	if badge.Valid {
		v.Badge.Number = int(badge.Int32)
	}
	if badgeID.Valid {
		v.Badge.ID = int(badgeID.Int32)
	}
	if password.Valid {
		v.Password = password.String
	}
	if avatar.Valid {
		v.Avatar = avatar.String
	}

	// check if abo is still valid
	if len(v.AboLast.BuyDate) > 5 {
		compareDate, err := time.Parse("2006-01-02", v.AboLast.BuyDate)
		if err != nil {
			return v, err
		}
		compareDate = compareDate.AddDate(1, 0, 0)

		if compareDate.After(time.Now()) || compareDate.Equal(time.Now()) {
			v.AboLastValid = true
		} else {
			v.AboLastValid = false
		}
	}

	// get partners and push each id to partner model and slice of partners
	var pSlice []models.Partner
	for i, v := range partners {
		var p models.Partner
		if v.Valid {
			p.ID = int(v.Int64)
			p.PartnerName = partnerNames[i].String
			pSlice = append(pSlice, p)
		}
	}
	if len(pSlice) > 0 {
		v.Partner = pSlice
	}

	return v, nil
}

// UpdateVisitor updates a visitor
func (m *postgresDBRepo) UpdateVisitor(v models.Visitor) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE visitors 
  SET first_name = $1, 
  family_name = $2, 
  birthdate = $3, 
  phone = $4,
  email = $5, 
  address = $6, 
  postcode = $7, 
  city = $8, 
  infos = $9, 
  newsletter = $10, 
  password = $11, 
  avatar = $12,
  ressort_id = $13,
  role_id = $14,
  updated_on = $15
  WHERE id = $16`

	_, err := m.DB.ExecContext(ctx, stmt,
		v.FirstName,
		v.FamilyName,
		v.Birthdate,
		NewNullString(v.Phone),
		NewNullString(strings.ToLower(v.Email)),
		v.Address,
		v.Postcode,
		v.City,
		NewNullString(v.Infos),
		v.Newsletter,
		NewNullString(v.Password),
		NewNullString(v.Avatar),
		NewNullInt(v.Ressort.ID),
		NewNullInt(v.RoleID),
		time.Now(),
		v.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// checks if email is unique
func (m *postgresDBRepo) CheckEmailUnique(email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM visitors WHERE email = $1`

	rows, err := m.DB.QueryContext(ctx, query, email)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	// return true, if there is no record set with the same email
	if !rows.Next() {
		return true, nil
	}

	return false, nil
}

// GetAllVisitors gets all visitors and returns also count
func (m *postgresDBRepo) CheckinAllVisitors(
	page int,
	filter models.VisitorFilter,
) ([]models.Visitor, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visitors []models.Visitor
	var count int

	search := "%" + filter.SearchString + "%"

	// set count to result
	row := m.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM visitors
  WHERE (LOWER(first_name) LIKE LOWER($1)
  OR LOWER(family_name) LIKE LOWER($1) 
  OR LOWER(city) LIKE LOWER($1))
  AND visitors.id NOT IN (
    SELECT visitor_id
    FROM visits
    WHERE DATE(date_in) = CURRENT_DATE)
`, search)

	err := row.Scan(&count)

	if err != nil {
		return visitors, count, err
	}

	offset := 10 * (page - 1)

	query := `SELECT visitors.id, first_name, family_name, city
  FROM visitors
  WHERE (LOWER(first_name) LIKE LOWER($1)
  OR LOWER(family_name) LIKE LOWER($1) 
  OR LOWER(city) LIKE LOWER($1))
  AND visitors.id NOT IN (
    SELECT visitor_id
    FROM visits
    WHERE DATE(date_in) = CURRENT_DATE)
  ORDER BY 
   family_name ASC
  OFFSET $2
  LIMIT 10
  `
	rows, err := m.DB.QueryContext(ctx, query, search, offset)
	if err != nil {
		return visitors, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var v models.Visitor

		err := rows.Scan(
			&v.ID,
			&v.FirstName,
			&v.FamilyName,
			&v.City,
		)

		if err != nil {
			return visitors, 0, err
		}
		visitors = append(visitors, v)
	}

	if err = rows.Err(); err != nil {
		return visitors, count, err
	}

	return visitors, count, nil
}

// CheckCreateLogin checks an email against a uid to make sure this visitor
// can create a new login
func (m *postgresDBRepo) CheckCreateLogin(email, uid string) (models.Visitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visitor models.Visitor

	query := `SELECT first_name, email
  FROM visitors WHERE email = $1 AND register_code = $2`

	row := m.DB.QueryRowContext(ctx, query, email, uid)

	err := row.Scan(&visitor.FirstName, &visitor.Email)
	if err != nil {
		return visitor, err
	}

	return visitor, nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// AllLogisticCategories returns all logistic categories
func (m *postgresDBRepo) AllLogisticCategories() ([]models.LogisticCategory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var categories []models.LogisticCategory

	query := `SELECT id, cat_name, price, price_member, created_on, updated_on FROM logistics_categories`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return categories, err
	}

	defer rows.Close()

	for rows.Next() {
		var l models.LogisticCategory

		err = rows.Scan(
			&l.ID,
			&l.CatName,
			&l.Price,
			&l.PriceMember,
			&l.CreatedOn,
			&l.UpdatedOn,
		)

		if err != nil {
			return categories, err
		}
		categories = append(categories, l)
	}

	if err = rows.Err(); err != nil {
		return categories, err
	}

	return categories, nil
}

// LogisticsByID returns a logistic place by id
func (m *postgresDBRepo) LogisticsByID(id int) (models.Logistic, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT l.id, l.number, l.start_date, l.end_date, l.reminder, l.member, l.unit_price, l.bill_id, b.paid, l.price, l.label, c.id, c.cat_name, v.id, v.first_name, v.family_name, v.phone, v.email, l.created_on, l.updated_on 
    FROM logistics l
    LEFT OUTER JOIN visitors v ON v.id = l.visitor_id
    LEFT OUTER JOIN logistics_categories c ON c.id = l.category
    LEFT OUTER JOIN billing b ON b.id = l.bill_id
    WHERE l.id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var l models.Logistic

	// start handling null types
	var startDate sql.NullTime
	var endDate sql.NullTime
	var reminder sql.NullTime
	var member sql.NullBool
	var unitPrice sql.NullFloat64
	var billID sql.NullInt32
	var billPaid sql.NullBool
	var price sql.NullFloat64
	var label sql.NullString
	var visID sql.NullString
	var firstName sql.NullString
	var familyName sql.NullString
	var phone sql.NullString
	var email sql.NullString

	err := row.Scan(
		&l.ID,
		&l.Number,
		&startDate,
		&endDate,
		&reminder,
		&member,
		&unitPrice,
		&billID,
		&billPaid,
		&price,
		&label,
		&l.Category.ID,
		&l.Category.CatName,
		&visID,
		&firstName,
		&familyName,
		&phone,
		&email,
		&l.CreatedOn,
		&l.UpdatedOn,
	)

	// handling null types
	if startDate.Valid {
		l.StartDate = startDate.Time
	}
	if endDate.Valid {
		l.EndDate = endDate.Time
	}
	if reminder.Valid {
		l.Reminder = reminder.Time
	}
	if member.Valid {
		l.Member = member.Bool
	}
	if unitPrice.Valid {
		l.UnitPrice = unitPrice.Float64
	}
	if billID.Valid {
		l.BillID = int(billID.Int32)
	}
	if billPaid.Valid {
		l.BillPaid = billPaid.Bool
	}
	if price.Valid {
		l.Price = price.Float64
	}
	if label.Valid {
		l.Label = label.String
	}
	if visID.Valid {
		l.Visitor.ID = visID.String
	}
	if firstName.Valid {
		l.Visitor.FirstName = firstName.String
	}
	if familyName.Valid {
		l.Visitor.FamilyName = familyName.String
	}
	if phone.Valid {
		l.Visitor.Phone = phone.String
	}
	if email.Valid {
		l.Visitor.Email = email.String
	}

	if err != nil {
		return l, err
	}

	return l, nil
}

// AllLogistics filters and returns all logistics from db
func (m *postgresDBRepo) AllLogistics(page int, filter models.LogisticFilter) ([]models.Logistic, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var logistics []models.Logistic
	var count int

	// prepare filters
	search := "%" + filter.Filter + "%"
	inuse := ""
	validity := ""
	lType := ""

	// check if there are filters and set where clauses
	switch filter.InUse {
	case "used":
		inuse = "AND visitor_id IS NOT NULL"
	case "available":
		inuse = "AND visitor_id IS NULL"
	}

	switch filter.Validity {
	case "valid":
		inuse = "AND end_date >= CURRENT_DATE"
	case "invalid":
		inuse = "AND end_date < CURRENT_DATE"
	}

	if filter.Type > 0 {
		lType = fmt.Sprintf("AND category = %d", filter.Type)
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM logistics l
  LEFT OUTER JOIN visitors v ON v.id = l.visitor_id
  WHERE (LOWER(v.first_name) LIKE LOWER($1)
  OR LOWER(v.family_name) LIKE LOWER($1) 
  OR l.number::varchar(20) LIKE LOWER($1))
  %s
  %s
  %s
  `,
		inuse,
		validity,
		lType,
	)

	// set count to 0
	row := m.DB.QueryRowContext(ctx, countQuery, search)

	err := row.Scan(&count)

	if err != nil {
		return logistics, count, err
	}

	offset := 10 * (page - 1)

	query := fmt.Sprintf(
		`SELECT l.id, l.number, l.start_date, l.end_date, l.reminder, l.member, l.bill_id, l.price, c.id, c.cat_name, v.id, v.first_name, v.family_name, v.phone, v.email, b.paid
    FROM logistics l
    LEFT OUTER JOIN visitors v ON v.id = l.visitor_id
    LEFT OUTER JOIN logistics_categories c ON c.id = l.category
    LEFT OUTER JOIN billing b ON b.id = l.bill_id
    WHERE (LOWER(v.first_name) LIKE LOWER($1)
    OR LOWER(v.family_name) LIKE LOWER($1) 
    OR l.number::varchar(20) LIKE LOWER($1))
    %s
    %s
    %s
    ORDER BY 
    l.number ASC
    OFFSET $2
    LIMIT 10  
    `,
		inuse,
		validity,
		lType,
	)

	rows, err := m.DB.QueryContext(ctx, query, search, offset)
	if err != nil {
		return logistics, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var l models.Logistic

		// start handling null types
		var startDate sql.NullTime
		var endDate sql.NullTime
		var reminder sql.NullTime
		var member sql.NullBool
		var billID sql.NullInt32
		var price sql.NullFloat64
		var visID sql.NullString
		var firstName sql.NullString
		var familyName sql.NullString
		var phone sql.NullString
		var email sql.NullString
		var billPaid sql.NullBool

		err := rows.Scan(
			&l.ID,
			&l.Number,
			&startDate,
			&endDate,
			&reminder,
			&member,
			&billID,
			&price,
			&l.Category.ID,
			&l.Category.CatName,
			&visID,
			&firstName,
			&familyName,
			&phone,
			&email,
			&billPaid,
		)

		// handling null types
		if startDate.Valid {
			l.StartDate = startDate.Time
		}
		if endDate.Valid {
			l.EndDate = endDate.Time
		}
		if reminder.Valid {
			l.Reminder = reminder.Time
		}
		if member.Valid {
			l.Member = member.Bool
		}
		if billID.Valid {
			l.BillID = int(billID.Int32)
		}
		if price.Valid {
			l.Price = price.Float64
		}
		if visID.Valid {
			l.Visitor.ID = visID.String
		}
		if firstName.Valid {
			l.Visitor.FirstName = firstName.String
		}
		if familyName.Valid {
			l.Visitor.FamilyName = familyName.String
		}
		if phone.Valid {
			l.Visitor.Phone = phone.String
		}
		if email.Valid {
			l.Visitor.Email = email.String
		}
		if billPaid.Valid {
			l.BillPaid = billPaid.Bool
		}

		if err != nil {
			return logistics, 0, err
		}
		logistics = append(logistics, l)
	}

	return logistics, count, nil
}

// CreateLogistic creates a logistic place with or whithout assignment
func (m *postgresDBRepo) CreateLogistic(l models.Logistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO logistics (number, start_date, end_date, category, visitor_id, member, unit_price, bill_id, price, created_on, updated_on)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := m.DB.ExecContext(
		ctx,
		stmt,
		l.Number,
		NewNullTime(l.StartDate),
		NewNullTime(l.EndDate),
		l.Category.ID,
		NewNullString(l.Visitor.ID),
		l.Member,
		l.UnitPrice,
		NewNullInt(l.BillID),
		NewNullFloat64(l.Price),
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// CreateLogistic creates a logistic place with or whithout assignment
func (m *postgresDBRepo) UpdateLogistic(l models.Logistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE logistics
  SET number = $1, 
  start_date = $2, 
  end_date = $3, 
  category = $4, 
  visitor_id = $5, 
  member = $6, 
  unit_price = $7, 
  bill_id = $8, 
  price = $9, 
  updated_on = $10
  WHERE id = $11`

	_, err := m.DB.ExecContext(
		ctx,
		stmt,
		l.Number,
		NewNullTime(l.StartDate),
		NewNullTime(l.EndDate),
		l.Category.ID,
		NewNullString(l.Visitor.ID),
		l.Member,
		l.UnitPrice,
		NewNullInt(l.BillID),
		NewNullFloat64(l.Price),
		time.Now(),
		l.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// DeleteLogistic deletes a logistic place from db
func (m *postgresDBRepo) DeleteLogistic(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM logistics WHERE id = $1`

	_, err := m.DB.ExecContext(
		ctx,
		stmt,
		id,
	)

	if err != nil {
		return err
	}

	return nil
}

// SetLogisticLabel sets a label for a logistic place in the db
func (m *postgresDBRepo) SetLogisticLabel(l models.Logistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE logistics 
  SET label = $1,
  updated_on = $2 
  WHERE id = $3
  `

	_, err := m.DB.ExecContext(
		ctx,
		query,
		l.Label,
		time.Now(),
		l.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteLogisticLabel deletes a logistic label from DB
func (m *postgresDBRepo) DeleteLogisticLabel(l models.Logistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE logistics 
  SET label = NULL,
  updated_on = $1 
  WHERE id = $2
  `

	_, err := m.DB.ExecContext(
		ctx,
		query,
		time.Now(),
		l.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// LogisticsByVisitorID returns al Logistics belonging to one visitor that have a price higher than 0
func (m *postgresDBRepo) LogisticsByVisitorID(vID string) ([]models.Logistic, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var logistics []models.Logistic

	if len(vID) < 1 {
		return logistics, nil
	}

	query := `SELECT l.id, l.number, l.start_date, l.end_date, l.reminder, l.member, l.bill_id, l.price, c.id, c.cat_name, c.price, c.price_member 
    FROM logistics l
    LEFT OUTER JOIN logistics_categories c ON c.id = l.category
    WHERE l.visitor_id = $1
    AND l.bill_id IS NULL
    AND l.price > 0
    ORDER BY 
    l.number ASC`

	rows, err := m.DB.QueryContext(ctx, query, vID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var l models.Logistic

		// start handling null types
		var startDate sql.NullTime
		var endDate sql.NullTime
		var reminder sql.NullTime
		var member sql.NullBool
		var billID sql.NullInt32
		var price sql.NullFloat64

		err := rows.Scan(
			&l.ID,
			&l.Number,
			&startDate,
			&endDate,
			&reminder,
			&member,
			&billID,
			&price,
			&l.Category.ID,
			&l.Category.CatName,
			&l.Category.Price,
			&l.Category.PriceMember,
		)

		// handling null types
		if startDate.Valid {
			l.StartDate = startDate.Time
		}
		if endDate.Valid {
			l.EndDate = endDate.Time
		}
		if reminder.Valid {
			l.Reminder = reminder.Time
		}
		if member.Valid {
			l.Member = member.Bool
		}
		if billID.Valid {
			l.BillID = int(billID.Int32)
		}
		if price.Valid {
			l.Price = price.Float64
		}

		if err != nil {
			return nil, err
		}
		logistics = append(logistics, l)
	}

	return logistics, nil
}

// UpdateLogisticsBillID updates logistics with their corresponding bill id
func (m *postgresDBRepo) UpdateLogisticsBillID(billID int, logistics []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE logistics 
   SET bill_id = $1
   WHERE id = ANY($2)
  `
	_, err := m.DB.ExecContext(ctx, query, billID, logistics)

	if err != nil {
		return err
	}

	return nil
}

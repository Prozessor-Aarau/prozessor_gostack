package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/google/uuid"
)

// CreateBill puts a new bill into the database and returns the id
func (m *postgresDBRepo) CreateBill(b models.BillData) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int

	stmt := `INSERT INTO billing (
    purpose, 
    visitor_id, 
    with_address, 
    firma,
    first_name, 
    family_name, 
    street, 
    number, 
    postcode,  
    city,
    amount,
    infos,
    date,
    with_message,
    message,
    with_reference,
    reference,
    custom,
    created_on, 
    updated_on)
  VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
  RETURNING id`

	err := m.DB.QueryRowContext(ctx, stmt,
		b.Purpose,
		NewNullString(b.Visitor.ID),
		b.WithAddress,
		NewNullString(b.Firma),
		NewNullString(b.FirstName),
		NewNullString(b.FamilyName),
		NewNullString(b.Street),
		NewNullString(b.Number),
		NewNullString(b.Postcode),
		NewNullString(b.City),
		NewNullFloat64(b.Amount),
		NewNullString(b.Infos),
		b.Date,
		b.WithMessage,
		NewNullString(b.Message),
		b.WithReference,
		NewNullString(b.Reference),
		b.Custom,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	// handle items and push them to the database
	if len(b.Items) > 0 {

		// create slices
		var ids []uuid.UUID
		var billIDs []int
		var positions []int
		var descriptions []string
		var unitPrices []float64
		var counts []float64

		for _, i := range b.Items {
			uid, err := uuid.Parse(i.ID)
			if err != nil {
				return 0, err
			}
			ids = append(ids, uid)
			billIDs = append(billIDs, id)
			positions = append(positions, i.Position)
			descriptions = append(descriptions, i.Description)
			unitPrices = append(unitPrices, i.UnitPrice)
			counts = append(counts, i.Count)
		}

		bStmt := `
    INSERT INTO bill_items (id, bill_id, position, description, unit_price, count)
    VALUES (
      UNNEST($1::UUID[]), 
      UNNEST($2::INT[]), 
      UNNEST($3::INT[]), 
      UNNEST($4::TEXT[]), 
      UNNEST($5::REAL[]), 
      UNNEST($6::REAL[]) 
    )`

		_, err := m.DB.ExecContext(
			ctx,
			bStmt,
			ids,
			billIDs,
			positions,
			descriptions,
			unitPrices,
			counts,
		)

		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

// SetBillPath sets the path for the given bill by id
func (m *postgresDBRepo) SetBillPath(path string, id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "UPDATE billing SET file_name= $1 WHERE id = $2"

	_, err := m.DB.ExecContext(ctx, stmt, path, id)

	if err != nil {
		return err
	}

	return nil
}

// GetAllBills gets all bills
func (m *postgresDBRepo) GetAllBills(
	page int,
	filter models.BillFilter,
) ([]models.BillData, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var bills []models.BillData
	var count int

	// prepare filters
	search := "%" + filter.Filter + "%"
	payment := ""
	bType := ""

	// check if there are filters and set where clauses
	switch filter.Payment {
	case "paid":
		payment = "AND b.paid = true"
	case "unpaid":
		payment = "AND b.paid = false"
	}

	if filter.Type != "all" {
		bType = fmt.Sprintf("AND b.purpose = '%s'", filter.Type)
	}

	countQuery := fmt.Sprintf(`SELECT count(*) 
  FROM billing b
    WHERE (b.id::varchar(20) LIKE LOWER($1)
    OR LOWER(b.family_name) LIKE LOWER($1) 
    OR LOWER(b.first_name) LIKE LOWER($1))
    %s
    %s
  `,
		payment,
		bType,
	)

	// set count to 0
	row := m.DB.QueryRowContext(ctx, countQuery, search)

	err := row.Scan(&count)

	if err != nil {
		return bills, count, err
	}

	offset := 10 * (page - 1)

	query := fmt.Sprintf(
		`SELECT b.id, b.purpose, b.with_address, b.firma, b.first_name, b.family_name, b.street, b.number, b.postcode, b.city, b.infos, b.date, b.with_message, b.message, b.with_reference, b.reference, b.custom, b.file_name, b.paid, b.amount, b.created_on, b.updated_on, v.id, v.first_name, v.family_name, v.address, v.postcode, v.city, v.ressort_id
  FROM billing b
  LEFT OUTER JOIN visitors v ON v.id = b.visitor_id
    WHERE (b.id::varchar(20) LIKE LOWER($1)
    OR LOWER(b.family_name) LIKE LOWER($1) 
    OR LOWER(b.first_name) LIKE LOWER($1))
    %s
    %s
    ORDER BY 
    b.date DESC
    OFFSET $2
    LIMIT 10  
  `,
		payment,
		bType,
	)

	rows, err := m.DB.QueryContext(ctx, query, search, offset)
	if err != nil {
		return bills, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var b models.BillData

		// start handling null types
		var aFirma sql.NullString
		var aFirstName sql.NullString
		var aFamilyName sql.NullString
		var aStreet sql.NullString
		var aNumber sql.NullString
		var aPostcode sql.NullString
		var aCity sql.NullString
		var message sql.NullString
		var reference sql.NullString
		var filename sql.NullString
		var visitorID sql.NullString
		var vFirstName sql.NullString
		var vFamilyName sql.NullString
		var vAddress sql.NullString
		var vPostcode sql.NullString
		var vCity sql.NullString
		var vRessortID sql.NullInt32

		err := rows.Scan(
			&b.ID,
			&b.Purpose,
			&b.WithAddress,
			&aFirma,
			&aFirstName,
			&aFamilyName,
			&aStreet,
			&aNumber,
			&aPostcode,
			&aCity,
			&b.Infos,
			&b.Date,
			&b.WithMessage,
			&message,
			&b.WithReference,
			&reference,
			&b.Custom,
			&filename,
			&b.Paid,
			&b.Amount,
			&b.CreatedOn,
			&b.UpdatedOn,
			&visitorID,
			&vFirstName,
			&vFamilyName,
			&vAddress,
			&vPostcode,
			&vCity,
			&vRessortID,
		)

		if err != nil {
			return bills, 0, err
		}

		// handling null types
		if aFirma.Valid {
			b.Firma = aFirma.String
		}
		if aFirstName.Valid {
			b.FirstName = aFirstName.String
		}
		if aFamilyName.Valid {
			b.FamilyName = aFamilyName.String
		}
		if aStreet.Valid {
			b.Street = aStreet.String
		}
		if aNumber.Valid {
			b.Number = aNumber.String
		}
		if aPostcode.Valid {
			b.Postcode = aPostcode.String
		}
		if aCity.Valid {
			b.City = aPostcode.String
		}
		if message.Valid {
			b.Message = message.String
		}
		if reference.Valid {
			b.Reference = reference.String
		}
		if filename.Valid {
			b.FileName = filename.String
		}
		if visitorID.Valid {
			b.Visitor.ID = visitorID.String
		}
		if vFirstName.Valid {
			b.Visitor.FirstName = vFirstName.String
		}
		if vFamilyName.Valid {
			b.Visitor.FamilyName = vFamilyName.String
		}
		if vAddress.Valid {
			b.Visitor.Address = vAddress.String
		}
		if vPostcode.Valid {
			b.Visitor.Postcode = vPostcode.String
		}
		if vCity.Valid {
			b.Visitor.City = vCity.String
		}
		if vRessortID.Valid {
			b.Visitor.Ressort.ID = int(vRessortID.Int32)
		}

		bills = append(bills, b)
	}

	if err = rows.Err(); err != nil {
		return bills, 0, err
	}

	return bills, count, nil
}

// GetBillByID gets a bill by id
func (m *postgresDBRepo) GetBillByID(id int) (models.BillData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT b.id, b.purpose, b.with_address, b.firma, b.first_name, b.family_name, b.street, b.number, b.postcode, b.city, b.infos, b.date, b.with_message, b.message, b.with_reference, b.reference, b.file_name, b.paid, b.amount, b.created_on, b.updated_on, v.id, v.first_name, v.family_name, v.address, v.postcode, v.city, v.ressort_id
  FROM billing b
  LEFT OUTER JOIN visitors v ON v.id = b.visitor_id
  WHERE b.id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var b models.BillData

	// start handling null types
	var aFirma sql.NullString
	var aFirstName sql.NullString
	var aFamilyName sql.NullString
	var aStreet sql.NullString
	var aNumber sql.NullString
	var aPostcode sql.NullString
	var aCity sql.NullString
	var message sql.NullString
	var reference sql.NullString
	var filename sql.NullString
	var visitorID sql.NullString
	var vFirstName sql.NullString
	var vFamilyName sql.NullString
	var vAddress sql.NullString
	var vPostcode sql.NullString
	var vCity sql.NullString
	var vRessortID sql.NullInt32

	err := row.Scan(
		&b.ID,
		&b.Purpose,
		&b.WithAddress,
		&aFirma,
		&aFirstName,
		&aFamilyName,
		&aStreet,
		&aNumber,
		&aPostcode,
		&aCity,
		&b.Infos,
		&b.Date,
		&b.WithMessage,
		&message,
		&b.WithReference,
		&reference,
		&filename,
		&b.Paid,
		&b.Amount,
		&b.CreatedOn,
		&b.UpdatedOn,
		&visitorID,
		&vFirstName,
		&vFamilyName,
		&vAddress,
		&vPostcode,
		&vCity,
		&vRessortID,
	)

	if err != nil {
		return b, err
	}

	// handling null types
	if aFirma.Valid {
		b.Firma = aFirma.String
	}
	if aFirstName.Valid {
		b.FirstName = aFirstName.String
	}
	if aFamilyName.Valid {
		b.FamilyName = aFamilyName.String
	}
	if aStreet.Valid {
		b.Street = aStreet.String
	}
	if aNumber.Valid {
		b.Number = aNumber.String
	}
	if aPostcode.Valid {
		b.Postcode = aPostcode.String
	}
	if aCity.Valid {
		b.City = aPostcode.String
	}
	if message.Valid {
		b.Message = message.String
	}
	if reference.Valid {
		b.Reference = reference.String
	}
	if filename.Valid {
		b.FileName = filename.String
	}
	if visitorID.Valid {
		b.Visitor.ID = visitorID.String
	}
	if vFirstName.Valid {
		b.Visitor.FirstName = vFirstName.String
	}
	if vFamilyName.Valid {
		b.Visitor.FamilyName = vFamilyName.String
	}
	if vAddress.Valid {
		b.Visitor.Address = vAddress.String
	}
	if vPostcode.Valid {
		b.Visitor.Postcode = vPostcode.String
	}
	if vCity.Valid {
		b.Visitor.City = vCity.String
	}
	if vRessortID.Valid {
		b.Visitor.Ressort.ID = int(vRessortID.Int32)
	}

	// get bill items
	billQuery := `SELECT id, position, description, unit_price, count, created_on, updated_on
	 FROM bill_items
	 WHERE bill_id = $1`

	rows, err := m.DB.QueryContext(ctx, billQuery, id)
	if err != nil {
		return b, err
	}

	defer rows.Close()

	var items []models.BillItem

	for rows.Next() {
		var i models.BillItem

		err := rows.Scan(
			&i.ID,
			&i.Position,
			&i.Description,
			&i.UnitPrice,
			&i.Count,
			&i.CreatedOn,
			&i.UpdatedOn,
		)

		if err != nil {
			return b, err
		}

		items = append(items, i)
	}

	if err = rows.Err(); err != nil {
		return b, err
	}

	// add items to the bill
	b.Items = items

	return b, nil
}

// GetBillCategories gets all categories of bills
func (m *postgresDBRepo) GetBillCategories() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var categories []string

	query := `SELECT purpose
    FROM billing
    GROUP BY purpose`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return categories, err
	}

	defer rows.Close()

	for rows.Next() {
		var c string

		err := rows.Scan(&c)

		if err != nil {
			return categories, err
		}

		categories = append(categories, c)
	}

	if err = rows.Err(); err != nil {
		return categories, err
	}

	return categories, nil
}

// SetBillPayment sets the payment status of a bill
func (m *postgresDBRepo) SetBillPayment(b models.BillData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "UPDATE billing SET paid = $1, updated_on = NOW() WHERE id = $2"

	_, err := m.DB.ExecContext(ctx, stmt, b.Paid, b.ID)

	if err != nil {
		return err
	}

	return nil
}

// DeleteBill deletes a bill by id
func (m *postgresDBRepo) DeleteBill(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "DELETE FROM billing WHERE id = $1"

	_, err := m.DB.ExecContext(ctx, stmt, id)

	if err != nil {
		return err
	}

	return nil
}

// UpdateBill updates a bill and its items in the database
func (m *postgresDBRepo) UpdateBill(b models.BillData) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var filename string

	stmt := `UPDATE billing SET
    purpose = $1, 
    visitor_id = $2, 
    with_address = $3, 
    firma = $4,
    first_name = $5, 
    family_name = $6, 
    street = $7, 
    number = $8, 
    postcode = $9,  
    city = $10,
    amount = $11,
    infos = $12,
    date = $13,
    with_message = $14,
    message = $15,
    with_reference = $16,
    reference = $17,
    custom = $18,
    updated_on = NOW()
  WHERE id = $19
  RETURNING file_name`

	err := m.DB.QueryRowContext(ctx, stmt,
		b.Purpose,
		NewNullString(b.Visitor.ID),
		b.WithAddress,
		NewNullString(b.Firma),
		NewNullString(b.FirstName),
		NewNullString(b.FamilyName),
		NewNullString(b.Street),
		NewNullString(b.Number),
		NewNullString(b.Postcode),
		NewNullString(b.City),
		NewNullFloat64(b.Amount),
		NewNullString(b.Infos),
		b.Date,
		b.WithMessage,
		NewNullString(b.Message),
		b.WithReference,
		NewNullString(b.Reference),
		b.Custom,
		b.ID,
	).Scan(&filename)

	if err != nil {
		return "", err
	}

	// handle items if there are any
	// first delete items that have been deleted

	// create slices for items
	var ids []uuid.UUID
	var billIDs []int
	var positions []int
	var descriptions []string
	var unitPrices []float64
	var counts []float64

	for _, i := range b.Items {
		uid, err := uuid.Parse(i.ID)
		if err != nil {
			return filename, err
		}
		ids = append(ids, uid)
		billIDs = append(billIDs, b.ID)
		positions = append(positions, i.Position)
		descriptions = append(descriptions, i.Description)
		unitPrices = append(unitPrices, i.UnitPrice)
		counts = append(counts, i.Count)
	}

	// delete what was deleted
	deleteStmt := `DELETE FROM bill_items WHERE bill_id = $1 AND NOT (id = ANY($2))`

	_, err = m.DB.ExecContext(ctx, deleteStmt, b.ID, ids)

	if err != nil {
		return filename, err
	}

	// add whats new

	bStmt := `
    INSERT INTO bill_items (id, bill_id, position, description, unit_price, count)
    VALUES (
      UNNEST($1::UUID[]), 
      UNNEST($2::INT[]), 
      UNNEST($3::INT[]), 
      UNNEST($4::TEXT[]), 
      UNNEST($5::REAL[]), 
      UNNEST($6::REAL[]) 
    )
    ON CONFLICT (id) DO UPDATE
    SET
        position = EXCLUDED.position, 
        description = EXCLUDED.description,
        unit_price = EXCLUDED.unit_price,
        count = EXCLUDED.count, 
        updated_on = NOW()
    `

	_, err = m.DB.ExecContext(ctx, bStmt, ids, billIDs, positions, descriptions, unitPrices, counts)

	if err != nil {
		return filename, err
	}

	return filename, nil
}

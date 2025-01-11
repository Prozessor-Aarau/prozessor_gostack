package repository

import (
	"context"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
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
    first_name, 
    family_name, 
    street, 
    number, 
    postcode,  
    city,
    amount,
    infos,
    with_message,
    message,
    with_reference,
    reference,
    created_on, 
    updated_on)
  VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
  RETURNING id`

	err := m.DB.QueryRowContext(ctx, stmt,
		b.Purpose,
		NewNullString(b.Visitor.ID),
		b.WithAddress,
		NewNullString(b.Visitor.FirstName),
		NewNullString(b.Visitor.FamilyName),
		NewNullString(b.Street),
		NewNullString(b.Number),
		NewNullString(b.Visitor.Postcode),
		NewNullString(b.Visitor.City),
		NewNullFloat64(b.Amount),
		NewNullString(b.Infos),
		b.WithMessage,
		NewNullString(b.Message),
		b.WithReference,
		NewNullString(b.Reference),
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, err
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

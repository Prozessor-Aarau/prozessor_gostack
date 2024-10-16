package repository

import (
	"context"
	"errors"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateRegisterCode creates and stores a uuid in the DB for later reference
// when registering a new login
func (m *postgresDBRepo) CreateRegisterCode(v models.Visitor) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var uidString string

	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	uidString = uuid.String()

	stmt := `UPDATE visitors SET register_code = $1 WHERE email = $2`

	_, err = m.DB.ExecContext(ctx, stmt, uidString, v.Email)
	if err != nil {
		return "", err
	}

	return uidString, nil
}

func (m *postgresDBRepo) CreatePassword(v models.Visitor) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE visitors 
  SET password = $1, role_id = 1, updated_on = $2, register_code = NULL
  WHERE email = $3`

	_, err := m.DB.ExecContext(ctx, stmt,
		v.Password,
		time.Now(),
		v.Email,
	)

	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (string, models.Visitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id string
	var hashedPassword string
	var visitor models.Visitor

	row := m.DB.QueryRowContext(ctx, `SELECT id, password FROM visitors WHERE email = $1`, email)

	err := row.Scan(&id, &hashedPassword)

	if err != nil {
		return id, visitor, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return "", visitor, errors.New("incorrect password")
	} else if err != nil {
		return "", visitor, err
	}

	visitor, err = m.VisitorByID(id)
	if err != nil {
		return "", visitor, err
	}

	return id, visitor, nil
}

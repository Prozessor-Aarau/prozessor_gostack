package repository

import (
	"context"
	"time"
)

// UpdateExpertise updates the expertises for a user by id
func (m *postgresDBRepo) UpdateExpertise(userID string, expertises []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// delete existing expertise first
	_, err = tx.ExecContext(ctx,
		"DELETE FROM visitor_experts WHERE visitor_id = $1",
		userID,
	)

	// insert chosen expertise
	for _, ex := range expertises {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO visitor_experts (visitor_id, atelier_id) VALUES ($1, $2)",
			userID,
			ex,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

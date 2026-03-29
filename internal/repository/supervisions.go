package repository

import (
	"context"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

func (m *postgresDBRepo) CreateOpeningHours(openingHours []models.OpeningHour) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, oh := range openingHours {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO opening_hours (start_date, end_date) VALUES ($1, $2)",
			oh.StartTime,
			oh.EndTime,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

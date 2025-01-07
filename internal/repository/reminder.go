package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// AllToRemindVisitors gets all visitors that should get a reminder for the abonnement from the db
func (m *postgresDBRepo) AllToRemindVisitors() ([]models.ReminderVisitor, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT v.id, first_name, family_name, v.email, v.ressort_id, b.number, a.buydate, a.buydate + INTERVAL '1 year' AS aboenddate, a.buydate + INTERVAL '1 year' >= CURRENT_DATE AS abovalid, DATE_PART('day', a.buydate + INTERVAL '1 year' - CURRENT_DATE) AS DateDifference, a.reminder
  FROM visitors v
  LEFT JOIN LATERAL 
    (SELECT * FROM abonnements 
    WHERE v.id = abonnements.visitor_id
    ORDER BY buydate DESC LIMIT 1) 
    as a ON v.id = a.visitor_id
  LEFT OUTER JOIN badges b ON b.visitor_id = v.id
    LEFT OUTER JOIN visitor_partners vp ON vp.visitor_id = v.id 
  WHERE b.id > 0
  AND (buydate + INTERVAL '1 year' <= CURRENT_DATE + INTERVAL '30 days' OR buydate IS NULL)
  AND vp.partner_id IS NULL
  AND (a.reminder + INTERVAL '45 days' <= CURRENT_DATE OR a.reminder IS NULL)
  GROUP BY v.id, a.buydate, b.number, a.reminder
  ORDER BY buydate ASC`

	var visitors []models.ReminderVisitor

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return visitors, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var v models.ReminderVisitor

		// start handling null types
		var ressortID sql.NullInt16
		var buydate sql.NullTime
		var enddate sql.NullTime
		var abovalid sql.NullBool
		var datediff sql.NullInt16
		var reminder sql.NullTime

		err := rows.Scan(
			&v.Visitor.ID,
			&v.Visitor.FirstName,
			&v.Visitor.FamilyName,
			&v.Visitor.Email,
			&ressortID,
			&v.BadgeNumber,
			&buydate,
			&enddate,
			&abovalid,
			&datediff,
			&reminder,
		)

		// handling null types
		if ressortID.Valid {
			v.RessortID = int(ressortID.Int16)
		}
		if buydate.Valid {
			v.BuyDate = buydate.Time
		}
		if enddate.Valid {
			v.EndDate = enddate.Time
		}
		if abovalid.Valid {
			v.Valid = abovalid.Bool
		}
		if datediff.Valid {
			v.Difference = int(datediff.Int16)
		}
		if reminder.Valid {
			v.Reminder = reminder.Time
		}

		if err != nil {
			return visitors, 0, err
		}

		visitors = append(visitors, v)
	}

	if err = rows.Err(); err != nil {
		return visitors, 0, err
	}

	return visitors, len(visitors), nil
}

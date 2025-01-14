package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/lib/pq"
)

func (m *postgresDBRepo) GetVisitsByDate(date time.Time) ([]models.Visit, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var visits []models.Visit
	var count int

	queryDate := date.Format("2006-01-02")

	row := m.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM visits WHERE DATE(date_in) = $1",
		queryDate,
	)

	err := row.Scan(&count)

	if err != nil {
		return visits, 0, err
	}

	query := `SELECT v.id, v.date_in, v.abonnement, v.ressort_id, v.partner_id, vt.id, vt.first_name, vt.family_name, vt.city, array_agg(va.atelier_id) AS ateliers, array_agg(a.icon) AS icons
  FROM visits v
  JOIN visitors vt ON vt.id = v.visitor_id
  JOIN visit_ateliers va on va.visit_id = v.id
  JOIN ateliers a on a.id = va.atelier_id
  WHERE DATE(date_in) = $1
  GROUP BY v.id, vt.id, vt.first_name, vt.family_name, vt.city
  ORDER BY v.date_in DESC`

	rows, err := m.DB.QueryContext(ctx, query, queryDate)
	if err != nil {
		return visits, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var v models.Visit

		// start handling null types
		var ressortId sql.NullInt32
		var partnerId sql.NullInt32

		// create empty array for ateliers and icons
		var ateliers pq.Int32Array
		var icons pq.StringArray

		err := rows.Scan(
			&v.ID,
			&v.DateIn,
			&v.Abonnement,
			&ressortId,
			&partnerId,
			&v.Visitor.ID,
			&v.Visitor.FirstName,
			&v.Visitor.FamilyName,
			&v.Visitor.City,
			&ateliers,
			&icons,
		)

		if err != nil {
			return visits, 0, err
		}

		// handling null types
		if ressortId.Valid {
			v.Ressort.ID = int(ressortId.Int32)
		}
		if partnerId.Valid {
			v.Partner.ID = int(partnerId.Int32)
		}

		// get ateliers and push each id to atelier model and slice of ateliers
		var ateliersArray []models.Atelier
		for i, id := range ateliers {
			var a models.Atelier
			a.Icon = icons[i]

			a.ID = int(id)
			ateliersArray = append(ateliersArray, a)
		}

		v.Ateliers = ateliersArray

		visits = append(visits, v)
	}

	if err = rows.Err(); err != nil {
		return visits, count, err
	}

	return visits, count, nil
}

// GetVisitByID gets a single visit by date
func (m *postgresDBRepo) GetVisitByID(id int) (models.Visit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var v models.Visit

	query := `SELECT v.id, v.date_in, v.abonnement, v.ressort_id, v.partner_id, vt.id, vt.first_name, vt.family_name, vt.city, array_agg(va.atelier_id) AS ateliers, array_agg(a.icon) AS icons
  FROM visits v
  JOIN visitors vt ON vt.id = v.visitor_id
  JOIN visit_ateliers va on va.visit_id = v.id
  JOIN ateliers a on a.id = va.atelier_id
  WHERE v.id = $1
  GROUP BY v.id, vt.id, vt.first_name, vt.family_name, vt.city`

	row := m.DB.QueryRowContext(ctx, query, id)

	// start handling null types
	var ressortId sql.NullInt32
	var partnerId sql.NullInt32

	// create empty array for ateliers and icons
	var ateliers pq.Int32Array
	var icons pq.StringArray

	err := row.Scan(
		&v.ID,
		&v.DateIn,
		&v.Abonnement,
		&ressortId,
		&partnerId,
		&v.Visitor.ID,
		&v.Visitor.FirstName,
		&v.Visitor.FamilyName,
		&v.Visitor.City,
		&ateliers,
		&icons,
	)

	if err != nil {
		return v, err
	}

	// handling null types
	if ressortId.Valid {
		v.Ressort.ID = int(ressortId.Int32)
	}
	if partnerId.Valid {
		v.Partner.ID = int(partnerId.Int32)
	}

	// get ateliers and push each id to atelier model and slice of ateliers
	var ateliersArray []models.Atelier
	for i, id := range ateliers {
		var a models.Atelier
		a.Icon = icons[i]

		a.ID = int(id)
		ateliersArray = append(ateliersArray, a)
	}

	v.Ateliers = ateliersArray

	return v, nil
}

// PostVisit posts a new visit
func (m *postgresDBRepo) PostVisit(visit models.Visit, ateliers []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// get user from DB
	user, err := m.VisitorByID(visit.Visitor.ID)
	if err != nil {
		return err
	}
	visit.Visitor = user

	visit.Abonnement = visit.Visitor.AboLastValid

	// insert visit into DB and return new ID
	stmt := `INSERT INTO visits (abonnement, date_in, partner_id, visitor_id, ressort_id, created_on, updated_on)
    VALUES($1, $2, $3, $4, $5, $6, $7) 
    RETURNING id
  `
	var visitID int
	err = m.DB.QueryRowContext(
		ctx,
		stmt,
		visit.Abonnement,
		time.Now(),
		NewNullInt(visit.Partner.ID),
		visit.Visitor.ID,
		NewNullInt(visit.Ressort.ID),
		time.Now(),
		time.Now(),
	).Scan(&visitID)

	if err != nil {
		return err
	}

	//create recordset and convert to json
	recordset := make([]map[string]int, 0, len(ateliers))
	for a := range ateliers {
		b := map[string]int{
			"visit_id":   visitID,
			"atelier_id": ateliers[a],
		}
		recordset = append(recordset, b)
	}

	json_records, err := json.Marshal(recordset)
	if err != nil {
		return err
	}

	// insert ateliers into linking table
	stmt = `INSERT INTO visit_ateliers (visit_id, atelier_id)
  SELECT visit_id, atelier_id FROM json_populate_recordset(null::visit_ateliers, $1)`

	_, err = m.DB.ExecContext(ctx, stmt, string(json_records))
	if err != nil {
		return err
	}

	return nil
}

// UpdateVisit posts a new visit
func (m *postgresDBRepo) UpdateVisit(visit models.Visit, ateliers []int, visitID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// get user from DB
	user, err := m.VisitorByID(visit.Visitor.ID)
	if err != nil {
		return err
	}
	visit.Visitor = user

	visit.Abonnement = visit.Visitor.AboLastValid

	// update visit in DB
	stmt := `UPDATE visits 
  SET abonnement = $1, 
  date_in = $2, 
  partner_id = $3, 
  ressort_id = $4,
  updated_on = $5
  WHERE id = $6`

	_, err = m.DB.ExecContext(
		ctx,
		stmt,
		visit.Abonnement,
		visit.DateIn,
		NewNullInt(visit.Partner.ID),
		NewNullInt(visit.Ressort.ID),
		time.Now(),
		visitID,
	)

	if err != nil {
		return err
	}

	//create recordset and convert to json
	recordset := make([]map[string]int, 0, len(ateliers))
	for a := range ateliers {
		b := map[string]int{
			"visit_id":   visitID,
			"atelier_id": ateliers[a],
		}
		recordset = append(recordset, b)
	}

	json_records, err := json.Marshal(recordset)
	if err != nil {
		return err
	}

	// TODO first delete all entries for this visit in the linking table
	_, err = m.DB.ExecContext(ctx, "DELETE FROM visit_ateliers WHERE visit_id = $1", visitID)
	if err != nil {
		return err
	}

	// insert ateliers into linking table
	stmt = `INSERT INTO visit_ateliers (visit_id, atelier_id)
  SELECT visit_id, atelier_id FROM json_populate_recordset(null::visit_ateliers, $1)`

	_, err = m.DB.ExecContext(ctx, stmt, string(json_records))
	if err != nil {
		return err
	}

	return nil
}

// DeleteVisit deletes a visit by id
func (m *postgresDBRepo) DeleteVisit(visitID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, "DELETE FROM visits WHERE id = $1", visitID)

	if err != nil {
		return err
	}

	return nil
}

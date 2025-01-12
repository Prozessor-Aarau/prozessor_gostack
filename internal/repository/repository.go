package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/lib/pq"
)

type DatabaseRepo interface {
	CreateVisitor(v models.Visitor) (string, error)
	UpdateVisitor(v models.Visitor) error
	CheckEmailUnique(email string) (bool, error)
	VisitorByID(id string) (models.Visitor, error)
	AllVisitors(page int, limit int, filter models.VisitorFilter) ([]models.Visitor, int, error)
	CheckinAllVisitors(page int, filter models.VisitorFilter) ([]models.Visitor, int, error)
	GetVisitsByDate(date time.Time) ([]models.Visit, int, error)
	GetVisitByID(id int) (models.Visit, error)
	GetAllPartners() ([]models.Partner, error)
	GetAllAteliers() ([]models.Atelier, error)
	PostVisit(visit models.Visit, ateliers []int) error
	UpdateVisit(visit models.Visit, ateliers []int, visitID int) error
	DeleteVisit(visitID int) error
	// Badges
	CreateBadge(b models.Badge) (int, error)
	DeleteBadge(id int) error
	UpdateBadge(b models.Badge) error
	AllBadges(filter models.BadgeFilter) ([]models.Badge, int, error)
	BadgeByID(id int) (models.Badge, error)
	ResetBadge(id int) error
	UniqueBadge(number int, uid string) (bool, bool, error)
	// abonnements
	CreateAbonnement(a models.Abonnement) (int, error)
	DeleteAbonnement(id int) error
	AllAbonnements(page int, filter models.AbonnementFilter) ([]models.Abonnement, int, error)
	CreateLog(l models.Log)
	// ressorts
	AllRessorts() ([]models.Ressort, error)
	RessortByID(id int) (models.Ressort, error)
	// roles
	AllRoles() ([]models.Role, error)
	RoleByID(id int) (models.Role, error)
	// registering and login
	CreateRegisterCode(v models.Visitor) (string, error)
	CreatePassword(v models.Visitor) error
	CheckCreateLogin(email, uid string) (models.Visitor, error)
	Authenticate(email, testPassword string) (string, models.Visitor, error)
	// machines
	CreateMachine(machine models.Machine) (string, error)
	UpdateMachine(machine models.Machine) error
	MachineByID(id string) (models.Machine, error)
	AllMachines(filter models.MachineFilter) ([]models.Machine, int, error)
	SetMachineQR(machine models.Machine) error
	DeleteMachineQR(machine models.Machine) error
	DeleteMachineById(id string) error
	// logistics
	AllLogisticCategories() ([]models.LogisticCategory, error)
	LogisticsByID(id int) (models.Logistic, error)
	AllLogistics(page int, filter models.LogisticFilter) ([]models.Logistic, int, error)
	CreateLogistic(l models.Logistic) error
	UpdateLogistic(l models.Logistic) error
	DeleteLogistic(id int) error
	SetLogisticLabel(l models.Logistic) error
	DeleteLogisticLabel(l models.Logistic) error
	// reports
	CreateReport(r models.Report) (int, error)
	// reminder
	AllToRemindVisitors() ([]models.ReminderVisitor, int, error)
	SetReminderDate(visitor models.ReminderVisitor) error
	// billing
	CreateBill(b models.BillData) (int, error)
	SetBillPath(path string, id int) error
	GetAllBills() ([]models.BillData, int, error)
	GetBillByID(id int) (models.BillData, error)
	GetBillCategories() ([]string, error)
	SetBillPayment(b models.BillData) error
}

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

// returns Nullstring if string is empty instead of empty string
func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// returns Nullstring if string is empty instead of empty string
func NewNullInt(d int) sql.NullInt32 {
	if d > 0 {
		return sql.NullInt32{
			Int32: int32(d),
			Valid: true,
		}
	}
	return sql.NullInt32{}
}

// returns Nullstring if string is empty instead of empty string
func NewNullFloat64(f float64) sql.NullFloat64 {
	if f > 0 {
		return sql.NullFloat64{
			Float64: float64(f),
			Valid:   true,
		}
	}
	return sql.NullFloat64{}
}

// returns Nullstring if string is empty instead of empty string
func NewNullTime(t time.Time) sql.NullTime {
	if !t.IsZero() {
		return sql.NullTime{
			Time:  t,
			Valid: true,
		}
	}
	return sql.NullTime{}
}

// CreateLog creates a log in the database
func (m *postgresDBRepo) CreateLog(l models.Log) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO log (type, module, message, visitor_id, created_on)
  VALUES($1, $2, $3, $4, $5)`

	_, err := m.DB.ExecContext(
		ctx,
		stmt,
		l.Typ,
		l.Module,
		l.Message,
		NewNullString(l.VisitorID),
		time.Now(),
	)

	if err != nil {
		log.Println(err)
	}
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

func (m *postgresDBRepo) GetAllAteliers() ([]models.Atelier, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ateliers []models.Atelier

	query := `SELECT id, atelier_name, description, active, icon, created_on, updated_on 
  FROM ateliers
  ORDER BY atelier_name ASC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return ateliers, err
	}

	defer rows.Close()

	for rows.Next() {
		var a models.Atelier

		err := rows.Scan(
			&a.ID,
			&a.AtelierName,
			&a.Description,
			&a.Active,
			&a.Icon,
			&a.CreatedOn,
			&a.UpdatedOn,
		)

		if err != nil {
			return ateliers, err
		}

		ateliers = append(ateliers, a)
	}

	if err = rows.Err(); err != nil {
		return ateliers, err
	}
	return ateliers, nil
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

// GetAllPartners gets all partners
func (m *postgresDBRepo) GetAllPartners() ([]models.Partner, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var partners []models.Partner

	query := `SELECT id, partner_name, contact_firstname, contact_familyname, email, address, postcode, city, description, active, created_on, updated_on 
  FROM partners
  ORDER BY partner_name ASC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return partners, err
	}

	defer rows.Close()

	for rows.Next() {
		var p models.Partner

		// start handling null types
		var address sql.NullString
		var description sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.PartnerName,
			&p.ContactFirstName,
			&p.ContactFamilyName,
			&p.Email,
			&address,
			&p.Postcode,
			&p.City,
			&description,
			&p.Active,
			&p.CreatedOn,
			&p.UpdatedOn,
		)

		if err != nil {
			return partners, err
		}

		if address.Valid {
			p.Address = address.String
		}

		if description.Valid {
			p.Description = description.String
		}

		partners = append(partners, p)
	}

	if err = rows.Err(); err != nil {
		return partners, err
	}
	return partners, nil
}

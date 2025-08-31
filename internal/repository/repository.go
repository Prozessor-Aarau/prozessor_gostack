package repository

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/faebut/prozessor_gostack/internal/config"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/newsletter"
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
	DeleteReport(id int) error
	UpdateReportStatus(status bool, id int) error
	ReportsByMachineID(id string) ([]models.Report, error)
	// reminder
	AllToRemindVisitors() ([]models.ReminderVisitor, int, error)
	SetReminderDate(visitor models.ReminderVisitor) error
	// billing
	CreateBill(b models.BillData) (int, error)
	SetBillPath(path string, id int) error
	GetAllBills(page int, filter models.BillFilter) ([]models.BillData, int, error)
	GetBillByID(id int) (models.BillData, error)
	GetBillCategories() ([]string, error)
	SetBillPayment(b models.BillData) error
	DeleteBill(id int) error
	LogisticsByVisitorID(vID string) ([]models.Logistic, error)
	UpdateLogisticsBillID(billID int, logistics []int) error
	UpdateBill(b models.BillData) (string, error)
	// newsletter
	GetAllNewsletters() ([]models.Newsletter, int, error)
	GetNewsletterTemplates() ([]models.NewsletterTemplate, error)
	GetAllNewsletterElementsByID(id int) ([]models.NewsletterElement, error)
	CreateNewsletter(n models.Newsletter) error
	UpdateNewsletter(n models.Newsletter) error
	GetNewsletterByID(id int) (models.Newsletter, error)
	DeleteNewsletterImage(id string) (models.NewsletterElement, error)
	NewsletterRecipients(recipientGroup string) ([]models.Visitor, error)
	NewsletterUnsubscribe(id string, member bool) (models.Visitor, error)
	NewsletterSubscribe(recipient newsletter.RecipientForm) error
	CheckSubscriptionUnique(email string) (bool, error)
	// statistics
	GetStatisticsVisitsByRange() ([]models.ChartVisitCount, error)
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

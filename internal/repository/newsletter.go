package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/view/newsletter"
	"github.com/google/uuid"
)

// GetAllNewsletters gets all newsletters
func (m *postgresDBRepo) GetAllNewsletters() ([]models.Newsletter, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newsletters []models.Newsletter

	query := `SELECT
    nl.id,
    nl.title,
    nl.description,
    nl.recipient_group,
    nl.template_id,
    nt.name,
    nl.sent_count,
    nl.sent_on,
    nl.updated_on,
    nl.created_on
  FROM
    newsletters nl
    LEFT OUTER JOIN newsletter_templates nt ON nl.template_id = nt.id
  ORDER BY
    nl.updated_on DESC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return newsletters, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var n models.Newsletter

		// handling null types
		var sentCount sql.NullInt32
		var sentOn sql.NullTime

		err := rows.Scan(
			&n.ID,
			&n.Title,
			&n.Description,
			&n.RecipientGroup,
			&n.Template.ID,
			&n.Template.Name,
			&sentCount,
			&sentOn,
			&n.UpdatedOn,
			&n.CreatedOn,
		)

		if err != nil {
			return newsletters, 0, err
		}

		// handling null types
		if sentCount.Valid {
			n.SentCount = int(sentCount.Int32)
		}
		if sentOn.Valid {
			n.SentOn = sentOn.Time
		}

		newsletters = append(newsletters, n)
	}

	if err = rows.Err(); err != nil {
		return newsletters, 0, err
	}

	return newsletters, len(newsletters), nil
}

// GetNewsletterByID get a newsletter by id
func (m *postgresDBRepo) GetNewsletterByID(id int) (models.Newsletter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var n models.Newsletter

	query := `SELECT
    nl.id,
    nl.title,
    nl.description,
    nl.recipient_group,
    nl.template_id,
		nl.color_base,
		nl.color_accent,
    nt.name,
    nt.template,
    nl.sent_count,
    nl.sent_on,
    nl.updated_on,
    nl.created_on
  FROM
    newsletters nl
    LEFT OUTER JOIN newsletter_templates nt ON nl.template_id = nt.id
	WHERE
		nl.id = $1`

	rows := m.DB.QueryRowContext(ctx, query, id)

	// handling null types
	var sentCount sql.NullInt32
	var sentOn sql.NullTime

	err := rows.Scan(
		&n.ID,
		&n.Title,
		&n.Description,
		&n.RecipientGroup,
		&n.Template.ID,
		&n.ColorBase,
		&n.ColorAccent,
		&n.Template.Name,
		&n.Template.Template,
		&sentCount,
		&sentOn,
		&n.UpdatedOn,
		&n.CreatedOn,
	)

	if err != nil {
		return n, err
	}

	// handling null types
	if sentCount.Valid {
		n.SentCount = int(sentCount.Int32)
	}
	if sentOn.Valid {
		n.SentOn = sentOn.Time
	}

	if err = rows.Err(); err != nil {
		return n, err
	}

	return n, nil
}

// GetNewsletterTemplates gets all templates for newsletters
func (m *postgresDBRepo) GetNewsletterTemplates() ([]models.NewsletterTemplate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var templates []models.NewsletterTemplate

	query := `SELECT
    id,
    name,
    description,
    template,
    updated_on,
    created_on
  FROM
    newsletter_templates
  ORDER BY
    name ASC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return templates, err
	}

	defer rows.Close()

	for rows.Next() {
		var t models.NewsletterTemplate

		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.Description,
			&t.Template,
			&t.UpdatedOn,
			&t.CreatedOn,
		)

		if err != nil {
			return templates, err
		}

		templates = append(templates, t)
	}

	if err = rows.Err(); err != nil {
		return templates, err
	}

	return templates, nil
}

// GetAllNewsletterElementsByID gets all elements by newsletter id
func (m *postgresDBRepo) GetAllNewsletterElementsByID(id int) ([]models.NewsletterElement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var elements []models.NewsletterElement

	query := `SELECT
    id,
    position,
    type,
    content,
    url,
    upload,
    alt_text,
    updated_on,
    created_on
  FROM
    newsletter_elements
  WHERE
    newsletter_id = $1
  ORDER BY
    position ASC`

	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return elements, err
	}

	defer rows.Close()

	for rows.Next() {
		var e models.NewsletterElement

		// handle null types
		var url sql.NullString
		var upload sql.NullString
		var alttext sql.NullString

		err := rows.Scan(
			&e.ID,
			&e.Position,
			&e.Type,
			&e.Content,
			&url,
			&upload,
			&alttext,
			&e.UpdatedOn,
			&e.CreatedOn,
		)

		// handle null types
		if url.Valid {
			e.Url = url.String
		}
		if upload.Valid {
			e.Upload = upload.String
		}
		if alttext.Valid {
			e.AltText = alttext.String
		}

		if err != nil {
			return elements, err
		}

		elements = append(elements, e)
	}

	if err = rows.Err(); err != nil {
		return elements, err
	}

	return elements, nil
}

// CreateNewsletter puts a new newsletter into the database
func (m *postgresDBRepo) CreateNewsletter(n models.Newsletter) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// var for new newsletter id
	var id int

	stmt := `INSERT INTO newsletters (
	title,
	description,
	template_id,
	recipient_group,
	color_base,
	color_accent)
	VALUES($1, $2, $3, $4, $5, $6)
	RETURNING id`

	err := m.DB.QueryRowContext(ctx, stmt,
		n.Title,
		n.Description,
		n.Template.ID,
		n.RecipientGroup,
		n.ColorBase,
		n.ColorAccent,
	).Scan(&id)

	if err != nil {
		return err
	}

	// handle elements and push them to the database
	// create slices
	var ids []uuid.UUID
	var newsletterIDs []int
	var positions []int
	var types []string
	var contents []string
	var urls []string
	var uploads []string
	var alttexts []string

	for _, e := range n.Elements {
		uid, err := uuid.Parse(e.ID)
		if err != nil {
			return err
		}
		ids = append(ids, uid)
		newsletterIDs = append(newsletterIDs, id)
		positions = append(positions, e.Position)
		types = append(types, e.Type)
		contents = append(contents, e.Content)
		urls = append(urls, e.Url)
		uploads = append(uploads, e.Upload)
		alttexts = append(alttexts, e.AltText)
	}

	eStmt := `
		INSERT INTO newsletter_elements (
		id,
		newsletter_id,
		position,
		type,
		content,
		url,
		upload,
		alt_text)
		VALUES(
			UNNEST($1::UUID[]),
			UNNEST($2::INT[]),
			UNNEST($3::INT[]),
			UNNEST($4::TEXT[]),
			UNNEST($5::TEXT[]),
			UNNEST($6::TEXT[]),
			UNNEST($7::TEXT[]),
			UNNEST($8::TEXT[])
		)`

	_, err = m.DB.ExecContext(
		ctx,
		eStmt,
		ids,
		newsletterIDs,
		positions,
		types,
		contents,
		urls,
		uploads,
		alttexts,
	)

	if err != nil {
		return err
	}

	return nil
}

// UpdateNewsletter puts a new newsletter into the database
func (m *postgresDBRepo) UpdateNewsletter(n models.Newsletter) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE newsletters SET
		title = $1,
		description = $2,
		template_id = $3,
		recipient_group = $4,
		color_base = $5,
		color_accent = $6
	WHERE id = $7`

	_, err := m.DB.ExecContext(ctx, stmt,
		n.Title,
		n.Description,
		n.Template.ID,
		n.RecipientGroup,
		n.ColorBase,
		n.ColorAccent,
		n.ID,
	)

	if err != nil {
		return err
	}

	// handle elements
	// first delete elements that have been deleted

	// create slices
	var ids []uuid.UUID
	var newsletterIDs []int
	var positions []int
	var types []string
	var contents []string
	var urls []string
	var uploads []string
	var alttexts []string

	for _, e := range n.Elements {
		uid, err := uuid.Parse(e.ID)
		if err != nil {
			return err
		}
		ids = append(ids, uid)
		newsletterIDs = append(newsletterIDs, n.ID)
		positions = append(positions, e.Position)
		types = append(types, e.Type)
		contents = append(contents, e.Content)
		urls = append(urls, e.Url)
		uploads = append(uploads, e.Upload)
		alttexts = append(alttexts, e.AltText)
	}

	// delete what was deleted
	deleteStmt := `DELETE FROM newsletter_elements WHERE newsletter_id = $1 AND NOT (id = ANY($2))`

	_, err = m.DB.ExecContext(ctx, deleteStmt, n.ID, ids)

	if err != nil {
		return err
	}

	// add whats new and update old
	eStmt := `
		INSERT INTO newsletter_elements (
		id,
		newsletter_id,
		position,
		type,
		content,
		url,
		upload,
		alt_text)
		VALUES(
			UNNEST($1::UUID[]),
			UNNEST($2::INT[]),
			UNNEST($3::INT[]),
			UNNEST($4::TEXT[]),
			UNNEST($5::TEXT[]),
			UNNEST($6::TEXT[]),
			UNNEST($7::TEXT[]),
			UNNEST($8::TEXT[])
		)
    ON CONFLICT (id) DO UPDATE
    SET
        position = EXCLUDED.position, 
				content = EXCLUDED.content,
				url = EXCLUDED.url,
				upload = EXCLUDED.upload,
				alt_text = EXCLUDED.alt_text
		`
	_, err = m.DB.ExecContext(
		ctx,
		eStmt,
		ids,
		newsletterIDs,
		positions,
		types,
		contents,
		urls,
		uploads,
		alttexts,
	)

	if err != nil {
		return err
	}

	return nil
}

// DeleteNewsletterImage deletes the image from the database and returns the element
func (m *postgresDBRepo) DeleteNewsletterImage(id string) (models.NewsletterElement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var element models.NewsletterElement

	// delete content of upload field and return the other contents
	stmt := `UPDATE
		newsletter_elements
	SET
		upload = ''
	WHERE
		id = $1
	RETURNING
		content,
		alt_text
	`

	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		id,
	).Scan(
		&element.Content,
		&element.AltText,
	)

	if err != nil {
		return element, err
	}

	// set element id to given id
	element.ID = id

	return element, nil
}

func (m *postgresDBRepo) NewsletterRecipients(recipientGroup string) ([]models.Visitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var recipients []models.Visitor

	groupQuery := ""

	if recipientGroup == "mitglied" {
		groupQuery = "AND v.ressort_id > 0"
	} else if recipientGroup == "newsletter" {
		groupQuery = "AND v.newsletter = true"
	} else if recipientGroup == "abonnement" {
		groupQuery = "AND b.number IS NOT NULL"
	}

	query := fmt.Sprintf(`SELECT
		v.id,
		v.first_name,
		v.family_name,
		v.email,
		v.city	
	FROM
		visitors v
		LEFT JOIN LATERAL (
			SELECT
				*
			FROM
				badges
			WHERE
				v.id = badges.visitor_id
			ORDER BY
				badges.number DESC
			LIMIT
				1
		) as b ON v.id = b.visitor_id
	WHERE
		v.email IS NOT NULL
		%s`,
		groupQuery,
	)

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return recipients, err
	}

	defer rows.Close()

	for rows.Next() {
		var r models.Visitor

		err := rows.Scan(
			&r.ID,
			&r.FirstName,
			&r.FamilyName,
			&r.Email,
			&r.City,
		)

		if err != nil {
			return recipients, err
		}
		recipients = append(recipients, r)
	}

	// also get normal newsletter recipients that aren't registered
	if recipientGroup == "newsletter" {
		queryNewsletter := `SELECT
  		id,
  		first_name,
  		family_name,
  		email
		FROM
  		newsletter_recipients
		`

		rows, err := m.DB.QueryContext(ctx, queryNewsletter)
		if err != nil {
			return recipients, err
		}

		defer rows.Close()

		for rows.Next() {
			var r models.Visitor

			err := rows.Scan(
				&r.ID,
				&r.FirstName,
				&r.FamilyName,
				&r.Email,
			)

			if err != nil {
				return recipients, err
			}
			recipients = append(recipients, r)
		}
	}

	return recipients, nil
}

func (m *postgresDBRepo) NewsletterUnsubscribe(id string, member bool) (models.Visitor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var recipient models.Visitor

	var stmt string

	if member {
		stmt = `UPDATE
		  visitors
		SET
  		newsletter = false
		WHERE
  		id = $1
		RETURNING
  		first_name, email;`
	} else {
		stmt = `DELETE FROM
			visitors 
		WHERE 
			id = $1 
		RETURNING 
			first_name, email`
	}

	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		id,
	).Scan(
		&recipient.FirstName,
		&recipient.Email,
	)

	if err != nil {
		return recipient, err
	}

	return recipient, nil
}

func (m *postgresDBRepo) NewsletterSubscribe(recipient newsletter.RecipientForm) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	query := `INSERT INTO 
		newsletter_recipients(
			id,
			email,
			first_name,
			family_name
		) VALUES ($1, $2, $3, $4)`

	_, err = m.DB.ExecContext(ctx, query, uid.String(), recipient.Email, recipient.FirstName, recipient.FamilyName)

	if err != nil {
		return err
	}

	return nil
}

// checks if email is unique in subscription
func (m *postgresDBRepo) CheckSubscriptionUnique(email string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT * FROM newsletter_recipients WHERE email = $1`

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

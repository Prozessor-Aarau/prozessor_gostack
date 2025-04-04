package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/google/uuid"
)

// AllBadges gets all badges
func (m *postgresDBRepo) GetAllNewsletters() ([]models.Newsletter, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newsletters []models.Newsletter

	query := `SELECT
    nl.id,
    nl.title,
    nl.description,
    nl.members,
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
    nl.updated_on ASC`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return newsletters, 0, err
	}

	defer rows.Close()

	for rows.Next() {
		var n models.Newsletter

		// handling null types
		var sentOn sql.NullTime

		err := rows.Scan(
			&n.ID,
			&n.Title,
			&n.Description,
			&n.Members,
			&n.Template.ID,
			&n.Template.Name,
			&n.SentCount,
			&sentOn,
			&n.UpdatedOn,
			&n.CreatedOn,
		)

		if err != nil {
			return newsletters, 0, err
		}

		// handling null types
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
	members,
	color_base,
	color_accent,
	updated_on,
	created_on)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id`

	err := m.DB.QueryRowContext(ctx, stmt,
		n.Title,
		n.Description,
		n.Template.ID,
		n.Members,
		n.ColorBase,
		n.ColorAccent,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return err
	}

	// handle elements and push them to the database
	if len(n.Elements) > 0 {

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

		_, err := m.DB.ExecContext(
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
	}

	return nil
}

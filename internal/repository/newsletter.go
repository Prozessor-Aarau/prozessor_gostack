package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
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

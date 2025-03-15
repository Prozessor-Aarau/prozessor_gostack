package models

import (
	"time"
)

type NewsletterElement struct {
	ID        int
	Position  int
	Type      string
	Content   string
	UploadUrl string
	CreatedOn time.Time
	UpdatedOn time.Time
}

type NewsletterTemplate struct {
	ID          int
	Name        string
	Description string
	Template    string
	CreatedOn   time.Time
	UpdatedOn   time.Time
}

type Newsletter struct {
	ID          int
	Title       string
	Description string
	Members     bool
	Template    NewsletterTemplate
	Elements    []NewsletterElement
	SentCount   int
	SentOn      time.Time
	CreatedOn   time.Time
	UpdatedOn   time.Time
}

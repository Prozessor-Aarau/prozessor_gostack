package models

// MailData holds an email message
type MailData struct {
	Reason     string
	Count      int
	Progress   int
	To         string
	FirstName  string
	FamilyName string
	Subject    string
	Content    string
	File       string
	// Template string
}

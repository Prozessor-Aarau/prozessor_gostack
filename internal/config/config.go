package config

import (
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/slack"
)

// appConfig holds the application configuration
type AppConfig struct {
	UseCache     bool
	InfoLog      *log.Logger
	ErrorLog     *log.Logger
	InProduction bool
	Session      *scs.SessionManager
	MailChan     chan models.MailData
	Reminder     struct {
		WaitList chan models.MailData
		Active   bool
	}
	Newsletter struct {
		WaitList chan models.MailData
		Active   bool
	}
	Slack *slack.Slack
}

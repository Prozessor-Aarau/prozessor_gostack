package slack

import (
	"fmt"
	"log"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/slack-go/slack"
)

type Slack struct {
	SlackClient *slack.Client
}

func NewSlack(token string) *Slack {
	return &Slack{
		SlackClient: slack.New(token, slack.OptionDebug(false)),
	}
}

func (s *Slack) SendSlackReport(report models.Report) error {
	// deal with the fatality of the report
	var fatality string
	if report.Fatal {
		fatality = "Die Maschine/das Gerät kann *nicht* mehr benutzt werden!"
	} else {
		fatality = "Die Maschine/das Gerät kann weiterhin benutzt werden"
	}

	// get timestamp and format
	timestamp := time.Now().Format("02.01.2006 15:04")

	// set contact info
	contact := "keine Angaben"
	if len(report.ReportReporter) > 0 {
		contact = report.ReportReporter
	}

	// construct message string out of all information
	message := fmt.Sprintf("<!channel> *Neue Fehlermeldung via QR-Code:*\n\n*Zeitpunkt*\n%s\n\n*Maschine/Gerät*\n%s\n\n*Schweregrad*\n%s\n\n*Informationen*\n%s\n\n*Kontaktangaben*\n%s", timestamp, report.MachineName, fatality, report.ReportInfo, contact)

	err := s.SendSlack(report.SlackChannel, message)
	if err != nil {
		return err
	}

	return nil
}

func (s *Slack) SendSlack(channelID, message string) error {
	cID, timestamp, err := s.SlackClient.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		return err
	}

	log.Printf("Message sent successfully to %s channel at %s", cID, timestamp)
	return nil
}

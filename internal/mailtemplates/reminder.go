package mailtemplates

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
)

// Reminder creates the template for the reminder mail
func Reminder(visitor models.ReminderVisitor) (string, error) {

	byteTemplate, err := os.ReadFile("./email-templates/reminder.html")
	if err != nil {
		return "", err
	}

	emailTemplate := string(byteTemplate)

	// here replace the template-strings with the values from the database

	endDate := visitor.EndDate
	bringbackDate := endDate.AddDate(0, 0, 30)
	bringbackDateTooLate := time.Now().AddDate(0, 0, 30)

	var msgAbo string
	var callToAction string
	var bringBack string

	// check if there ever was an abonnement
	if endDate.IsZero() {
		msgAbo = `Du warst bis anhin oder zu einem früheren Zeitpunkt aktiv im Prozessor oder hast aus einem anderen Grund freien Zugang zu den Werkstätten erhalten. An der Vereins-GV vom 5. Mai 2024 wurde beschlossen, dass per 1. Juni 2024 alle Vereinsmitglieder beitragspflichtig sind, sofern sie die Werkstatt ausserhalb der Öffnungszeiten für eigene Zwecke nutzen möchten.
    `

		if visitor.RessortID > 0 {
			callToAction = `Wir möchten uns an dieser Stelle ganz herzilch für deinen Einsatz im Prozessor bedanken! Es würde uns riesig freuen, dich weiterhin als aktives Mitglied in einem der Ressorts beim Prozessor zu haben. 
      <br><br>
      Wenn du weiterhin aktiv im Verein aktiv sein möchtest, hast du die Möglichkeit vergünstigtes Abonnement zum Betrag von 320.- pro Jahr zu erstehen. Einen entsprechenden Einzahlungsschein ist der Email angehängt. 
      <br><br>
      Möchtest du den Prozessor mit deinem Einsatz unterstützen, nutzt die Werkstätten aber nicht privat, so bist du nicht beitragspflichtig. In diesem Fall darfst du dich gerne bei uns melden, wir passen das gerne in unserem System so an.`
		} else {
			callToAction = `Es würde uns freuen, dich weiterhin beim Prozessor zu haben. Du kannst für CHF 400.- ein Abonnement erstehen und behältst deinen Zugangsbadge. In diesem Fall bitten wir dich, den angehängten Einzahlungsschein innert 30 Tagen zu bezahlen.
      <br><br>
      Noch viel toller wärs, wenn du Teil vom Prozessor wirst und dich als aktives Vereinsmitglied in einem der Ressorts engagierst. In diesem Fall profitierst du von einem vergünstigten Abonnement für CHF 320.-. Nimm mit uns Kontakt auf, falls du den Prozessor aktiv mitgestalten möchtest. Wir senden dir dann gerne den Einzahlungsschein für das vergünstigte Abonnement zu.`
		}

		bringBack = fmt.Sprintf(
			"bis spätestens am <strong>%s</strong>",
			bringbackDateTooLate.Format("02.01.2006"),
		)
	} else {
		// check if is part of a ressort
		if visitor.RessortID > 0 {
			msgAbo = `Du hast vor einiger Zeit ein Abonnement im Prozessor gelöst, dich aktiv in einem Ressort eingesetzt und einen Badge erhalten, der dir uneingeschränkten Zugang zu den Werkstätten gewährt.
    <br><br>`

			callToAction = `Wir möchten uns an dieser Stelle ganz herzilch für deinen Einsatz im Prozessor bedanken! Es würde uns riesig freuen, dich weiterhin als aktives Mitglied in einem der Ressorts beim Prozessor zu haben. 
      <br><br>
      Du kannst dein Abonnement zum vergünstigten Preis von CHF 320.- um ein weiteres Jahr verlängern und behältst deinen Zugangsbadge. Wir bitten dich, den angehängten Einzahlungsschein innert 30 Tagen zu bezahlen.`
		} else {
			msgAbo = `Du hast vor einiger Zeit ein Abonnement im Prozessor gelöst und einen Badge erhalten, der dir uneingeschränkten Zugang zu den Werkstätten gewährt.
    <br><br>`

			callToAction = `Es würde uns freuen, dich weiterhin beim Prozessor zu haben. Du kannst dein Abonnement für CHF 400.- um ein weiteres Jahr verlängern und behältst deinen Zugangsbadge. Wir bitten dich, den angehängten Einzahlungsschein innert 30 Tagen zu bezahlen.
      <br><br>
      Noch viel toller wärs, wenn du Teil vom Prozessor wirst und dich als aktives Vereinsmitglied in einem der Ressorts engagierst. In diesem Fall profitierst du von einem vergünstigten Abonnement für CHF 320.-. Nimm mit uns Kontakt auf, falls du den Prozessor aktiv mitgestalten möchtest. Wir senden dir dann gerne den Einzahlungsschein für das vergünstigte Abonnement zu.`
		}

		// check if abo is still valid
		if endDate.Before(time.Now()) {
			msgAbo = msgAbo + fmt.Sprintf(
				"Dein Abonnement ist am <strong>%s</strong> abgelaufen.",
				endDate.Format("02.01.2006"),
			)

			bringBack = fmt.Sprintf(
				"bis spätestens am <strong>%s</strong>",
				bringbackDateTooLate.Format("02.01.2006"),
			)
		} else {
			msgAbo = fmt.Sprintf(
				"Dein Abonnement läuft schon bald ab. Bis am <strong>%s</strong> kannst du die Werkstätten noch uneingeschränkt nutzen.",
				endDate.Format("02.01.2006"),
			)
			bringBack = fmt.Sprintf(
				"bis spätestens am <strong>%s</strong>",
				bringbackDate.Format("02.01.2006"),
			)
		}
	}

	// replace placeholders with generated strings
	msgToSend := strings.Replace(emailTemplate, "[%firstname%]", visitor.Visitor.FirstName, 1)
	msgToSend = strings.Replace(msgToSend, "[%reminder%]", msgAbo, 1)
	msgToSend = strings.Replace(msgToSend, "[%calltoaction%]", callToAction, 1)
	msgToSend = strings.Replace(msgToSend, "[%bringback%]", bringBack, 1)

	return msgToSend, nil
}

// Newsletter creates the template for a newsletter
func Newsletter(newsletter models.Newsletter) (string, error) {

	byteTemplate, err := os.ReadFile("./email-templates/newsletters/" + newsletter.Template.Template)
	if err != nil {
		return "", err
	}

	emailTemplate := string(byteTemplate)

	// here replace the template-strings with the values from the database

	// replace placeholders with generated strings
	// msgToSend := strings.Replace(emailTemplate, "[%firstname%]", visitor.Visitor.FirstName, 1)
	// msgToSend = strings.Replace(msgToSend, "[%reminder%]", msgAbo, 1)
	// msgToSend = strings.Replace(msgToSend, "[%calltoaction%]", callToAction, 1)
	// msgToSend = strings.Replace(msgToSend, "[%bringback%]", bringBack, 1)

	return emailTemplate, nil
}

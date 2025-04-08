package mailtemplates

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/goodsign/monday"
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

	// model html from the elements
	var content string
	for _, e := range newsletter.Elements {

		switch e.Type {

		case "text":
			content += string(mdToHTML([]byte(e.Content)))

		case "button":
			content += fmt.Sprintf(`<div style="display: flex; justify-content: center;">
				<a href="https://%s" target="_blank" style="text-decoration: none; color: #FFFFFF;">
				<div class="mailbutton" align="center" style="background-color: [%%basecolor%%]; padding: 10px 20px; border-radius: 5px; color: #ffffff; font-weight: bold;">
				%s
                  </div>
</a>
				</div>
            `, e.Url, e.Content)

		case "picture":
			content += fmt.Sprintf(`<img src="/static/newsletter/%s" alt="%s" width="100%%"><p style="margin-top: 4px; font-size: 0.8rem;">%s</p>`, e.Upload, e.AltText, e.Content)

		case "events":
			// unmarshal json to event slice
			var events []models.Event
			var month string
			err = json.Unmarshal([]byte(e.Content), &events)
			if err != nil {
				panic(err)
			}

			var eventHTML string
			// open grid tag
			eventHTML = `<h2>Kalender</h2>
			<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 15px;">`

			for _, e := range events {

				// TODO: wrong locale for mont and weekday
				if monday.Format(e.EventTime.Time, "Jan", monday.LocaleDeDE) != month {
					month = monday.Format(e.EventTime.Time, "Jan", monday.LocaleDeDE)
					eventHTML += fmt.Sprintf(`<div style="font-weight: 700; font-size: 1.25rem; padding: 6px; padding-left: 25px; background-color: [%%accentcolor%%]; text-transform: uppercase; grid-column: 1 / -1;">%s</div>`, month)
				}

				eventHTML += fmt.Sprintf(`
					<a class="calenderlink" style="display: flex; gap: 10px; text-decoration: none; margin-left: 25px; margin-right: 25px; align-items: center;" href="%s" target="_blank">
					<div style="width: 30px; font-weight: 600; background-color: [%%basecolor%%]; color: white; font-size: 1rem; padding: 5px 10px; display: flex; flex-direction: column; align-items: center;">
									<span>%s</span>
									<span>%s</span>
								</div>
					<div style="font-size: 16pt; font-weight: 700; padding-top: 10px; padding-bottom: 15px; line-height: 1.1;">
									%s
								</div>
							</a>
					`, e.URL,
					strings.ToUpper(monday.Format(e.EventTime.Time, "Monday", monday.LocaleDeDE)[:2]),
					e.EventTime.Format("02"),
					e.Title,
				)
			}

			// close div tag
			eventHTML += "</div>"

			content += eventHTML
		}

	}

	// TODO: change color for logo

	// replace placeholders with generated strings
	msgToSend := strings.Replace(emailTemplate, "[%content%]", content, 1)

	// replace colors
	msgToSend = strings.Replace(msgToSend, "[%basecolor%]", newsletter.ColorBase, -1)
	msgToSend = strings.Replace(msgToSend, "[%accentcolor%]", newsletter.ColorAccent, -1)

	// apply styles
	msgToSend = strings.Replace(msgToSend, `<h1`, fmt.Sprintf(`<h1 style="color: %s"`, newsletter.ColorBase), -1)
	msgToSend = strings.Replace(msgToSend, `<h2`, fmt.Sprintf(`<h2 style="color: %s"`, newsletter.ColorBase), -1)
	msgToSend = strings.Replace(msgToSend, `<h3`, fmt.Sprintf(`<h3 style="color: %s"`, newsletter.ColorBase), -1)

	return msgToSend, nil
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

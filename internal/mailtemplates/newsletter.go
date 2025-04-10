package mailtemplates

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/goodsign/monday"
)

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
			textFields := string(mdToHTML([]byte(e.Content)))
			content += fmt.Sprintf(`
          <tr>
            <td align="left" class="body" style="padding: 0 40px; font-size: 16px; line-height: 1.6;">
              %s
            </td>
          </tr>
				`, textFields)

		case "button":
			content += fmt.Sprintf(`
          <tr>
            <td style="padding: 0px 40px;" align="center">
              <table cellspacing="0" cellpadding="0" style="margin: auto;">
                <tr>
                  <td align="center" style="background-color: [%%basecolor%%]; padding: 10px 20px; border-radius: 5px;">
											<a href="https://%s" target="_blank"
                      style="color: #ffffff; text-decoration: none;"><b>%s</b></a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
            `, e.Url, e.Content)

		case "picture":
			content += fmt.Sprintf(`
          <tr>
            <td align="left" class="body" style="padding: 0 40px; font-size: 16px; line-height: 1.6;">
				<img style="margin-top: 20px;" src="https://db.prozessor-aarau.ch/static/newsletter/%s" alt="%s" width="100%%"><p style="margin-top: 0px; font-size: 13px;">%s</p>
            </td>
          </tr>
				`, e.Upload, e.AltText, e.Content)

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
			eventHTML = `
					<tr>
						<td class="body" style="padding: 0px 40px;" align="left">
						<h2 style="margin-top: auto;">Kalender</h2>
							<table width="100%" cellspacing="0" cellpadding="0" align="left" style="border-spacing: 0 10px;">
						`
			for _, e := range events {

				// TODO: get black or white if basecolor is dark or light

				if monday.Format(e.EventTime.Time, "Jan", monday.LocaleDeDE) != month {
					month = monday.Format(e.EventTime.Time, "January", monday.LocaleDeDE)
					eventHTML += fmt.Sprintf(`
                <tr>
                  <td colspan="2" style="font-size: 20px; padding: 6px 6px 6px 25px; background-color: [%%accentcolor%%]; color: [%%basecolor%%];">
										<b>%s</b>
                  </td>
                </tr>
					`, strings.ToUpper(month))
				}

				eventHTML += fmt.Sprintf(`
					<tr>
						<td class="datefield" style="width: 40px; background-color: [%%basecolor%%];">
							<table width="100%%" cellspacing="0" cellpadding="0" style="color: #FFFFFF" align="center">
								<tr>
									<td align="center">
										<b>%s</b>
									</td>
								</tr>
								<tr>
									<td align="center">
										<b>%s</b>
									</td>
								</tr>
							</table>
						</td>
						<td class="calenderlink linkfield" style="padding-left: 10px; font-size: 16pt; padding-top: 10px; padding-bottom: 15px; line-height: 1.1;">
					<a class="calenderlink" href="%s" target="_blank" style="color: black; text-decoration: none;"><b>%s</b></a>
						</td>
					</tr>
					`,
					strings.ToUpper(monday.Format(e.EventTime.Time, "Monday", monday.LocaleDeDE)[:2]),
					e.EventTime.Format("02"),
					e.URL,
					e.Title,
				)
			}

			// close div tag
			eventHTML += `
              </table>
            </td>
          </tr>
			`

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

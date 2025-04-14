package mailtemplates

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
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
                  <td class="button" align="center" style="background-color: [%%basecolor%%]; padding: 10px 20px; border-radius: 5px;">
											<a style="color: [%%colorfont%%]; text-decoration: none;" href="https://%s" target="_blank"
                      ><b>%s</b></a>
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
						<td class="body" style="line-height: 1.6; font-size: 16px; padding: 0px 40px;" align="left">
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
							<table width="100%%" cellspacing="0" cellpadding="0" style="color: [%%colorfont%%]" align="center">
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
	var colorfont string
	isLight, err := isLightColor(newsletter.ColorBase)
	if err != nil {
		return "", err
	}
	if isLight {
		colorfont = "#000000"
	} else {
		colorfont = "#FFFFFF"
	}

	msgToSend = strings.Replace(msgToSend, "[%basecolor%]", newsletter.ColorBase, -1)
	msgToSend = strings.Replace(msgToSend, "[%accentcolor%]", newsletter.ColorAccent, -1)
	msgToSend = strings.Replace(msgToSend, "[%colorfont%]", colorfont, -1)

	// apply styles
	msgToSend = strings.Replace(msgToSend, `<h1`, fmt.Sprintf(`<h1 style="color: %s"`, newsletter.ColorBase), -1)
	msgToSend = strings.Replace(msgToSend, `<h2`, fmt.Sprintf(`<h2 style="color: %s"`, newsletter.ColorBase), -1)
	msgToSend = strings.Replace(msgToSend, `<h3`, fmt.Sprintf(`<h3 style="color: %s"`, newsletter.ColorBase), -1)
	msgToSend = strings.Replace(msgToSend, `<a href`, fmt.Sprintf(`<a style="text-decoration: none; color: %s; font-style: italic;" href`, newsletter.ColorBase), -1)

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

func isLightColor(s string) (bool, error) {
	c, err := parseHexColorFast(s)
	if err != nil {
		return false, err
	}

	brightness := (int32(c.R)*299 + int32(c.G)*587 + int32(c.B)*114) / 1000

	return brightness > 127, nil
}

var errInvalidFormat = errors.New("invalid format")

func parseHexColorFast(s string) (c color.RGBA, err error) {
	c.A = 0xff

	if s[0] != '#' {
		return c, errInvalidFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = errInvalidFormat
	}
	return
}

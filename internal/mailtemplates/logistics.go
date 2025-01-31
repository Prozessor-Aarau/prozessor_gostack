package mailtemplates

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/faebut/prozessor_gostack/internal/models"
)

func LogisticsBill(visitor models.Visitor, logistics []models.Logistic) (string, error) {
	// read template
	byteTemplate, err := os.ReadFile("./email-templates/logistic.html")
	if err != nil {
		return "", err
	}
	emailTemplate := string(byteTemplate)

	// get total price
	var totalPrice float64
	for _, l := range logistics {
		totalPrice += l.Price
	}

	var listRows string
	for _, l := range logistics {
		listing := fmt.Sprintf(`
    <tr>
      <td style="padding-left: 5px;">%s</td>
      <td>%s</td>
      <td>%s</td>
      <td>%s</td>
      <td style="text-align: right; padding-right: 5px;">%s</td>
    </tr>
   `, strconv.Itoa(l.Number), l.Category.CatName, l.StartDate.Format("02.01.2006"), l.EndDate.Format("02.01.2006"), strconv.FormatFloat(l.Price, 'f', 2, 64))

		listRows += listing
	}

	totalRow := fmt.Sprintf(`
    <tr style="background: #EEE;">
      <td></td>
      <td></td>
      <td></td>
      <td style="text-align: right;"><strong>TOTAL</strong></td>
      <td style="text-align: right; padding-right: 5px;"><strong>%s</strong></td>
    </tr>
  `, strconv.FormatFloat(totalPrice, 'f', 2, 64))

	// replace placeholders with generated strings
	msgToSend := strings.Replace(emailTemplate, "[%firstname%]", visitor.FirstName, 1)
	msgToSend = strings.Replace(msgToSend, "[%listrows%]", listRows, 1)
	msgToSend = strings.Replace(msgToSend, "[%totalrow%]", totalRow, 1)

	return msgToSend, nil
}

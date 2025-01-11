package billing

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/qrbill"
	"github.com/faebut/prozessor_gostack/internal/repository"
	"github.com/joho/godotenv"
)

// CreateBill creates a bill and returns the billdata including the path to the file
func CreateBill(data models.BillData, m repository.DatabaseRepo) (string, error) {

	// create initial bill
	bill, err := qrbill.New(false)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// read .env file
	err = godotenv.Load()
	if err != nil {
		return "", err
	}
	// Set receiver options
	bill.IsQrIBAN = false
	bill.ReceiverIBAN = os.Getenv("BILL_IBAN")
	bill.ReceiverName = os.Getenv("BILL_NAME")
	bill.ReceiverStreet = os.Getenv("BILL_STREET")
	bill.ReceiverNumber = os.Getenv("BILL_NUMBER")
	bill.ReceiverZIPCode = os.Getenv("BILL_ZIP")
	bill.ReceiverPlace = os.Getenv("BILL_PLACE")
	bill.ReceiverCountry = os.Getenv("BILL_COUNTRY")
	bill.Currency = os.Getenv("BILL_CURRENCY")

	// set bill data
	// address
	if len(data.Visitor.Address) > 0 && data.Visitor.Address != "no address" {
		data.WithAddress = true
		exploded := strings.Split(data.Visitor.Address, " ")
		data.Number = exploded[len(exploded)-1]

		// put all strings back together
		for i := 0; i < len(exploded)-1; i++ {
			if i == 0 {
				data.Street = exploded[i]
			} else {
				data.Street = data.Street + " " + exploded[i]
			}
		}

		bill.PayeeName = data.Visitor.FirstName + " " + data.Visitor.FamilyName
		bill.PayeeStreet = data.Street
		bill.PayeeNumber = data.Number
		bill.PayeeZIPCode = data.Visitor.Postcode
		bill.PayeePlace = data.Visitor.City
		bill.PayeeCountry = "CH"
	}

	// message
	if len(data.Message) > 0 {
		data.WithMessage = true
		bill.AdditionalInfo = data.Message
	}

	// reference
	if len(data.Reference) > 0 {
		data.WithReference = true
		bill.Reference = data.Reference
	}

	// amount
	bill.Amount = strconv.FormatFloat(data.Amount, 'f', 2, 64)

	// create bill and return id
	data.ID, err = m.CreateBill(data)
	if err != nil {
		return "", err
	}

	// set message to RG Number
	bill.AdditionalInfo = fmt.Sprintf("RG%d: ", data.ID) + data.Message

	// check if dirs exist, if not create them
	filepath := "./static/bills/" + data.Purpose + "/"
	if _, err := os.Stat("static/bills"); err != nil {
		os.Mkdir("static/bills/", os.ModePerm)
	}
	if _, err := os.Stat("static/bills/" + data.Purpose); err != nil {
		os.Mkdir("static/bills/"+data.Purpose, os.ModePerm)
	}

	// generate file name
	// set time string for default filename
	timeString := time.Now().Format("060102")
	data.FileName = fmt.Sprintf("%s_%d_%s.pdf", timeString, data.ID, data.Purpose)

	combinedPath := filepath + data.FileName

	// Directly save invoice as PDF.
	err = bill.SaveAsPDF(combinedPath)
	if err != nil {
		return "", err
	}

	// update filename in database
	err = m.SetBillPath(data.FileName, data.ID)
	if err != nil {
		return "", err
	}

	return combinedPath, nil
}

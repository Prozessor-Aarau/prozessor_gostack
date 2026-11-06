package billing

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/faebut/prozessor_gostack/internal/qrbill"
	"github.com/faebut/prozessor_gostack/internal/repository"
	"github.com/goodsign/monday"
	"github.com/joho/godotenv"
	"github.com/signintech/gopdf"
)

// CreateBill creates a bill and returns the billdata including the path to the file
func CreateBill(data models.BillData, m repository.DatabaseRepo) (string, int, error) {

	// create initial bill
	bill, err := qrbill.New(false)
	if err != nil {
		return "", 0, err
	}

	// read .env file
	err = godotenv.Load()
	if err != nil {
		return "", 0, err
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

		data.FirstName = data.Visitor.FirstName
		data.FamilyName = data.Visitor.FamilyName
		data.Postcode = data.Visitor.Postcode
		data.City = data.Visitor.City
	}

	// now if any address exists, write it to the bill struct
	if len(data.Street) > 0 {
		data.WithAddress = true
		if len(data.Firma) > 0 {
			bill.PayeeName = data.Firma
		} else {
			bill.PayeeName = data.FirstName + " " + data.FamilyName
		}
		bill.PayeeZIPCode = data.Postcode
		bill.PayeePlace = data.City
		bill.PayeeStreet = data.Street
		bill.PayeeNumber = data.Number
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
	if data.ID > 0 {
		data.FileName, err = m.UpdateBill(data)
		// delete old file
		DeleteBill(data)
	} else {
		data.ID, err = m.CreateBill(data)
	}
	if err != nil {
		return "", 0, err
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
	timeString := data.Date.Format("060102")
	data.FileName = fmt.Sprintf("%s_%d_%s.pdf", timeString, data.ID, data.Purpose)

	combinedPath := filepath + data.FileName

	// Directly save invoice as PDF.
	err = bill.SaveAsPDF(combinedPath)
	if err != nil {
		return "", 0, err
	}

	// only to this for custom bills
	if len(data.Items) > 0 {
		letter, err := LetterWithItems(data)
		if err != nil {
			return "", 0, err
		}

		// put invoice to next page
		letter.AddPage()
		// Import page 1
		tpl1 := letter.ImportPage(combinedPath, 1, "/MediaBox")
		// Draw pdf onto page
		letter.UseImportedTemplate(tpl1, 0, 0, 210, 297)

		letter.WritePdf(combinedPath)
		letter.Close()
	}

	// update filename in database
	err = m.SetBillPath(data.FileName, data.ID)
	if err != nil {
		return "", 0, err
	}

	return combinedPath, data.ID, nil
}

func LetterWithItems(data models.BillData) (*gopdf.GoPdf, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4, Unit: gopdf.UnitMM})
	var err error

	// add fonts
	err = pdf.AddTTFFont("fira-extralight", "./static/assets/FiraSans-ExtraLight.ttf")
	if err != nil {
		return &pdf, err
	}

	err = pdf.AddTTFFont("fira-regular", "./static/assets/FiraSans-Regular.ttf")
	if err != nil {
		return &pdf, err
	}

	err = pdf.AddTTFFont("fira-bold", "./static/assets/FiraSans-Bold.ttf")
	if err != nil {
		return &pdf, err
	}

	// set header
	pdf.AddHeader(func() {
		err := pdf.Image("./static/assets/logo.png", 184, 10.5, &gopdf.Rect{W: 18.3, H: 18.3})
		if err != nil {
			return
		}

		const (
			xHeaderRight float64 = 109

			yTitleOneLineOne, yTitleOneLineTwo, yTitleOneLineThree float64 = 12.3, 17.6, 22.9
		)

		err = pdf.SetFont("fira-extralight", "", 10)
		if err != nil {
			slog.Error(err.Error())
			return
		}

		pdf.SetXY(xHeaderRight, yTitleOneLineOne)
		pdf.CellWithOption(&gopdf.Rect{W: 70.4, H: 35.2},
			"Verein Prozessor | Erlinsbacherstrasse 34 | 5000 Aarau",
			gopdf.CellOption{Align: gopdf.Right})

		pdf.SetXY(xHeaderRight, yTitleOneLineTwo)
		pdf.CellWithOption(&gopdf.Rect{W: 70.4, H: 35.2},
			"info@prozessor-aarau.ch | www.prozessor-aarau.ch",
			gopdf.CellOption{Align: gopdf.Right})

		pdf.SetXY(xHeaderRight, yTitleOneLineThree)
		pdf.CellWithOption(&gopdf.Rect{W: 70.4, H: 35.2},
			"IBAN "+os.Getenv("BILL_IBAN"),
			gopdf.CellOption{Align: gopdf.Right})
	})

	// add page with header
	pdf.AddPage()

	// set positions and fontsize
	const (
		X          = 25
		width      = 210 - 45
		maxX       = X + width
		addressY   = 50
		infoblockY = 45
		titleY     = 100

		fontsize   = 11
		titlesize  = 13
		lineheight = 1.5 * 0.3527777777777777
		newline    = fontsize * lineheight
		paragraph  = 2 * newline

		cellfont   = 9
		cellheight = 8

		tabletop = titleY + 7*newline
	)

	var (
		bottomcell  = tabletop + float64((len(data.Items)+1)*cellheight+fontsize/5)
		tablebottom = bottomcell + newline
	)

	// address
	err = pdf.SetFont("fira-regular", "", fontsize)
	if err != nil {
		return &pdf, err
	}

	var address []string

	if len(data.Firma) > 0 {
		address = append(address, data.Firma)
	}
	if len(data.FirstName) > 0 {
		address = append(address, data.FirstName+" "+data.FamilyName)
	}
	address = append(address, data.Street+" "+data.Number)
	address = append(address, data.Postcode+" "+data.City)

	for i := 0; i < len(address); i++ {
		pdf.SetXY(X, addressY+float64(i)*newline) //move to address
		pdf.Cell(nil, address[i])                 //write line
	}

	// right items
	err = pdf.SetFont("fira-extralight", "", fontsize)
	if err != nil {
		return &pdf, err
	}
	pdf.SetXY(maxX-77, infoblockY)
	pdf.CellWithOption(&gopdf.Rect{W: 35, H: 3.5},
		"Nummer:",
		gopdf.CellOption{Align: gopdf.Right})

	pdf.SetXY(maxX-77, infoblockY+newline)
	pdf.CellWithOption(&gopdf.Rect{W: 35, H: 3.5},
		"Zweck:",
		gopdf.CellOption{Align: gopdf.Right})

	// values
	err = pdf.SetFont("fira-regular", "", fontsize)
	if err != nil {
		return &pdf, err
	}
	pdf.SetXY(maxX-40.5, infoblockY)
	pdf.Cell(nil, "RG"+strconv.Itoa(data.ID))

	pdf.SetXY(maxX-40.5, infoblockY+newline)
	pdf.Cell(nil, data.Message)

	// date
	pdf.SetXY(maxX-35, addressY+float64(len(address)-1)*(newline))
	pdf.CellWithOption(&gopdf.Rect{W: 35, H: 3.5},
		"Aarau, "+monday.Format(time.Now(), "2. January 2006", monday.LocaleDeDE),
		gopdf.CellOption{Align: gopdf.Right})

	// title
	err = pdf.SetFont("fira-bold", "", titlesize)
	if err != nil {
		return &pdf, err
	}

	// title
	pdf.SetXY(X, titleY)
	pdf.Cell(nil, fmt.Sprintf("Rechnung: RG%s %s", strconv.Itoa(data.ID), data.Message))

	// content
	err = pdf.SetFont("fira-regular", "", fontsize)
	if err != nil {
		return &pdf, err
	}

	// body
	// greeting formula
	pdf.SetXY(X, titleY+3*newline)
	pdf.Cell(nil, "Sehr geehrte Damen und Herren")
	// first textblock
	pdf.SetXY(X, titleY+5*newline)
	pdf.Cell(nil, "Besten Dank für den Auftrag. Wir erlauben uns folgende Positionen in Rechnung zu stellen:")

	// table

	// set up table
	table := pdf.NewTableLayout(X, tabletop, cellheight, len(data.Items))
	table.AddColumn("POS", 10, "center")
	table.AddColumn("BEZEICHNUNG", 92, "left")
	table.AddColumn("MENGE", 16, "right")
	table.AddColumn("PREIS", 22, "right")
	table.AddColumn("TOTAL", 25, "right")

	// add rows
	for _, i := range data.Items {
		table.AddRow([]string{
			strconv.Itoa(i.Position),
			i.Description,
			strconv.FormatFloat(i.Count, 'f', 2, 64),
			strconv.FormatFloat(i.UnitPrice, 'f', 2, 64) + " CHF",
			strconv.FormatFloat(i.Count*i.UnitPrice, 'f', 2, 64) + " CHF",
		})
	}

	// Set the style for table cells
	table.SetTableStyle(gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			Top:    true,
			Left:   true,
			Bottom: true,
			Right:  true,
			Width:  0.2,
		},
		FillColor: gopdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: gopdf.RGBColor{R: 0, G: 0, B: 0},
		FontSize:  cellfont,
	})

	// Set the style for table header
	table.SetHeaderStyle(gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			Bottom:   true,
			Right:    true,
			Width:    0.2,
			RGBColor: gopdf.RGBColor{R: 0, G: 0, B: 0},
		},
		FillColor: gopdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: gopdf.RGBColor{R: 0, G: 0, B: 0},
		Font:      "fira-bold",
		FontSize:  cellfont,
	})

	table.SetCellStyle(gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			Right:    true,
			Bottom:   true,
			Width:    0.2,
			RGBColor: gopdf.RGBColor{R: 0, G: 0, B: 0},
		},
		FillColor: gopdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: gopdf.RGBColor{R: 0, G: 0, B: 0},
		Font:      "fira-regular",
		FontSize:  cellfont,
	})

	// Draw the table
	table.DrawTable()

	// TOTAL bottom
	err = pdf.SetFont("fira-bold", "", cellfont)
	if err != nil {
		return &pdf, err
	}
	pdf.SetXY(maxX-45, bottomcell)
	pdf.Cell(nil, "TOTAL")

	pdf.SetXY(maxX-25, bottomcell)
	pdf.CellWithOption(&gopdf.Rect{W: 23, H: 10},
		strconv.FormatFloat(data.Amount, 'f', 2, 64)+" CHF",
		gopdf.CellOption{Align: gopdf.Right})

	pdf.SetLineWidth(0.2)
	pdf.Line(X+118, bottomcell+0.5+cellfont/3+(cellheight-cellfont/3)/2, maxX, bottomcell+0.5+cellfont/3+(cellheight-cellfont/3)/2)
	pdf.Line(X+118, bottomcell+0.5+1+cellfont/3+(cellheight-cellfont/3)/2, maxX, bottomcell+0.5+1+cellfont/3+(cellheight-cellfont/3)/2)

	// content
	err = pdf.SetFont("fira-regular", "", fontsize)
	if err != nil {
		return &pdf, err
	}

	// longer text
	greetLines, err := pdf.SplitTextWithWordWrap(
		fmt.Sprintf(
			"Wir bitten Sie, die Rechnung bis %s mit dem beigefügten Einzahlungsschein zu begleichen.",
			monday.Format(
				time.Now().Add(30*time.Hour*24),
				"2. January 2006",
				monday.LocaleDeDE,
			),
		),
		width,
	)
	if err != nil {
		return &pdf, err
	}

	for i := 0; i < len(greetLines); i++ {
		pdf.SetXY(X, tablebottom+newline+float64(i)*newline)
		pdf.Cell(nil, greetLines[i])
	}

	pdf.SetXY(X, tablebottom+4*newline)
	pdf.Cell(nil, "Mit freundlichen Grüssen")
	pdf.SetXY(X, tablebottom+6*newline)
	pdf.Cell(nil, "Verein Prozessor")

	return &pdf, nil
}

// DeleteBill deletes a bill from the disk
func DeleteBill(data models.BillData) error {
	// delete file
	if err := os.Remove(fmt.Sprintf("./static/bills/%s/%s", data.Purpose, data.FileName)); err != nil {
		return fmt.Errorf("warning: failed to delete bill: %v", err)
	}

	return nil
}

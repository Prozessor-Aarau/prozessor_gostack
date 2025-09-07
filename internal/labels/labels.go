package labels

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/fogleman/gg"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// GenerateLabel generates a logistic label and returns the filename
func GenerateLabel(l models.Logistic) (string, error) {
	var filename string

	// get contact
	var contact string
	if len(l.Visitor.Phone) > 0 {
		contact = l.Visitor.Phone
	} else if len(l.Visitor.Email) > 0 {
		contact = l.Visitor.Email
	}

	// dimensions
	x := float64(2160)
	y := float64(2960)
	border := float64(200)
	starttext := float64(1400)
	gap := float64(240)
	smallgap := float64(50)
	startframe := starttext - border/2 - 120
	// set new drawing plane
	dc := gg.NewContext(int(x), int(y))
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	// Number
	if err := dc.LoadFontFace("./static/assets/MoskBold.ttf", 244); err != nil {
		return filename, fmt.Errorf("could not load font: %v", err.Error())
	}
	dc.DrawCircle(float64(x)/2, 500, 300)
	dc.SetLineWidth(40)
	dc.Stroke()
	dc.DrawStringAnchored(strconv.Itoa(l.Number), x/2, 500, 0.5, 0.4)

	// Infos
	if err := dc.LoadFontFace("./static/assets/MoskBold.ttf", 200); err != nil {
		return filename, fmt.Errorf("could not load font: %v", err.Error())
	}
	dc.DrawStringAnchored(l.Category.CatName, x/2, 1050, 0.5, 0)

	// rectangle
	dc.DrawRoundedRectangle(border, startframe, x-2*border, border/2+6*gap, 100)
	dc.SetLineWidth(20)
	dc.Stroke()

	// labels
	if err := dc.LoadFontFace("./static/assets/MoskNormal.ttf", 100); err != nil {
		return filename, fmt.Errorf("could not load font: %v", err.Error())
	}
	dc.DrawString("Gehört:", border+border/2, starttext)
	dc.DrawString("Reserviert bis:", border+border/2, starttext+2*gap)
	dc.DrawString("Kontakt:", border+border/2, starttext+4*gap)

	// info
	if err := dc.LoadFontFace("./static/assets/MoskNormal.ttf", 180); err != nil {
		return filename, fmt.Errorf("could not load font: %v", err.Error())
	}
	dc.DrawString(fmt.Sprintf("%s %s", l.Visitor.FirstName, l.Visitor.FamilyName), border+border/2, starttext+gap-smallgap)
	dc.DrawString(l.EndDate.Format("02.01.2006"), border+border/2, starttext+3*gap-smallgap)
	dc.DrawString(contact, border+border/2, starttext+5*gap-smallgap)

	dc.Clip()

	// create filename
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	filename = "Label_" + uuid.String() + ".png"

	// // check if uploads directory exists, if not create it
	if _, err := os.Stat("static/labels"); err != nil {
		os.Mkdir("static/labels/", os.ModePerm)
	}

	// save final file
	err = dc.SavePNG(fmt.Sprintf("./static/labels/%s", filename))
	if err != nil {
		return "", err
	}

	return filename, nil
}

// DeleteLabel deletes a logistic label
func DeleteLabel(l models.Logistic) error {
	// delete file
	if err := os.Remove(fmt.Sprintf("./static/labels/%s", l.Label)); err != nil {
		return fmt.Errorf("warning: failed to delete label: %v", err)
	}

	return nil
}

// PrintLabel prints a label on the label printer
func PrintLabel(l models.Logistic) error {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	api := os.Getenv("QR_PRINT_SERVER")

	apiURL := fmt.Sprintf("http://%s/api/print/image", api)

	fileName := fmt.Sprintf("./static/labels/%s", l.Label)

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", fileName)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	contentType := writer.FormDataContentType()
	fmt.Println(contentType)

	writer.Close()

	resp, err := http.Post(apiURL, contentType, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Status Code:", resp.StatusCode)

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

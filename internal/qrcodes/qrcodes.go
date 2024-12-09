package qrcodes

import (
	"fmt"
	"log"
	"os"

	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/fogleman/gg"
	"github.com/joho/godotenv"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

func GenerateQRCode(machine models.Machine) error {

	// load env file to read base URL
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	baseUrl := os.Getenv("BASE_URL")

	qrc, err := qrcode.NewWith(
		fmt.Sprintf("%s/report?machine=%s", baseUrl, machine.ID),
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest),
	)
	if err != nil {
		return fmt.Errorf("could not generate QRCode: %v", err.Error())
	}

	w, err := standard.New(
		fmt.Sprintf("./static/qr_codes/tmp_%s.jpg", machine.ID),
		standard.WithQRWidth(40),
		standard.WithBorderWidth(500, 100, 100, 100),
	)
	if err != nil {
		return fmt.Errorf("standard.New failed: %v", err.Error())
	}

	// save temp file
	if err = qrc.Save(w); err != nil {
		return fmt.Errorf("could not save image: %v", err.Error())
	}

	// load temp file
	im, err := gg.LoadJPG(fmt.Sprintf("./static/qr_codes/tmp_%s.jpg", machine.ID))
	if err != nil {
		log.Fatal(err)
	}
	dimensions := im.Bounds().Size()

	// set new drawing plane
	dc := gg.NewContext(dimensions.X, dimensions.Y+100)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	if err := dc.LoadFontFace("./static/assets/MoskNormal.ttf", 144); err != nil {
		return fmt.Errorf("could not load font: %v", err.Error())
	}
	dc.DrawRoundedRectangle(0, 0, float64(dimensions.X), float64(dimensions.Y), 0)
	dc.DrawImage(im, 0, 0)
	dc.DrawStringAnchored("Maschine defekt?", float64(dimensions.X)/2, float64(dimensions.Y)/11.5, 0.5, 0.5)

	if err := dc.LoadFontFace("./static/assets/MoskNormal.ttf", 96); err != nil {
		return fmt.Errorf("could not load font: %v", err.Error())
	}
	dc.DrawStringAnchored(machine.MachineName, float64(dimensions.X)/2, float64(dimensions.Y)/6.5, 0.5, 0.5)

	if err := dc.LoadFontFace("./static/assets/MoskNormal.ttf", 60); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored("Bitte QR-Code scannen und Formular ausfüllen", float64(dimensions.X)/2, float64(dimensions.Y)/1, 0.5, 0.5)
	dc.Clip()

	// save final file
	dc.SavePNG(fmt.Sprintf("./static/qr_codes/QR_%s.png", machine.ID))

	// delete temp file
	if err := os.Remove(fmt.Sprintf("./static/qr_codes/tmp_%s.jpg", machine.ID)); err != nil {
		return fmt.Errorf("Warning: failed to delete temporary qr code: %v", err)
	}

	return nil
}

func RemoveQRCode(machine models.Machine) error {
	// delete file
	if err := os.Remove(fmt.Sprintf("./static/qr_codes/QR_%s.png", machine.ID)); err != nil {
		return fmt.Errorf("Warning: failed to delete qr code: %v", err)
	}

	return nil
}

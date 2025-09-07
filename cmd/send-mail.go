package main

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/faebut/prozessor_gostack/internal/logger"
	"github.com/faebut/prozessor_gostack/internal/models"
	"github.com/joho/godotenv"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	go func() {
		for {
			msg := <-app.MailChan
			// send mail here
			sendMsg(msg)
		}
	}()
}

func sendMsg(m models.MailData) {
	// get .env file
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("loading .env file", slog.String("error", err.Error()))
	}

	server := mail.NewSMTPClient()
	server.Host = os.Getenv("SMTP_HOST")
	server.Port, err = strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		slog.Error(err.Error())
		return
	}

	from := os.Getenv("SMTP_USERNAME")
	server.Username = from
	server.Password = os.Getenv("SMTP_PASSWORD")

	encryption, err := strconv.ParseBool(os.Getenv("SMTP_ENCRYPTION"))
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if encryption {
		server.Encryption = mail.EncryptionTLS
	} else {
		server.Encryption = mail.EncryptionNone
	}

	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(m.To).SetSubject(m.Subject)
	email.SetBody(mail.TextHTML, m.Content)

	if len(m.File) > 0 {
		email.Attach(&mail.File{FilePath: m.File})
	}

	err = email.Send(client)
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

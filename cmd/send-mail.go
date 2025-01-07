package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
		log.Fatal("error loading .env file")
	}

	server := mail.NewSMTPClient()
	server.Host = os.Getenv("SMTP_HOST")
	server.Port, err = strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		errorLog.Println(err)
		return
	}

	server.Username = os.Getenv("SMTP_USERNAME")
	server.Password = os.Getenv("SMTP_PASSWORD")

	encryption, err := strconv.ParseBool(os.Getenv("SMTP_ENCRYPTION"))
	if err != nil {
		errorLog.Println(err)
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
		errorLog.Println(err)
		return
	}

	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		data, err := os.ReadFile(fmt.Sprintf("./email-templates/%s", m.Template))
		if err != nil {
			errorLog.Println(err)
			return
		}
		mailTemplate := string(data)

		msgToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	err = email.Send(client)
	if err != nil {
		log.Println(err)
		return
	} else {
		log.Println("Email sent")
	}
}

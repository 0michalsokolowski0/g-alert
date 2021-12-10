package main

import (
	"context"
	"crypto/tls"
	"log"
	"os/exec"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/xhit/go-simple-mail/v2"
)

func main() {
	// TODO: move it to env var to make it configurable
	scheduler := gocron.NewScheduler(time.Local)

	// TODO: cron expression should be passed as env var
	_, err := scheduler.Cron("*/1 * * * *").Do(func() {
		out, err := exec.CommandContext(context.Background(), "googler", "--json", "--np", "--exact", "search_phrase").Output()
		if err != nil {
			panic(err)
		}

		server := mail.NewSMTPClient()

		// TODO: move to config
		server.Host = "smtp.gmail.com"
		server.Port = 587
		server.Username = "!!!!@gmail.com"
		server.Password = "!!!!"
		server.Encryption = mail.EncryptionSTARTTLS
		server.ConnectTimeout = 10 * time.Second
		server.SendTimeout = 10 * time.Second

		server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		// SMTP client
		smtpClient, err := server.Connect()

		if err != nil {
			log.Fatal(err)
		}

		// New email simple html with inline and CC
		email := mail.NewMSG()
		email.SetFrom("From Example <example@example.com>").
			AddTo("TO_EMAIL").
			SetSubject("SUBJECT")

		email.SetBody(mail.TextPlain, string(out))

		err = email.Send(smtpClient)

	})
	if err != nil {
		panic(err)
	}

	scheduler.StartBlocking()
}

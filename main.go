package main

import (
	"context"
	"github.com/0michalsokolowski0/g-alert/internal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/xhit/go-simple-mail/v2"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	logger := logrus.New()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	config, err := internal.NewConfig(configPath)
	if err != nil {
		logger.WithError(err).Error("Failed to load configuration")
		os.Exit(1)
	}

	location, err := time.LoadLocation(config.TimeLocation)
	if err != nil {
		logger.WithError(err).Error("Failed to parse time location")
		os.Exit(1)
	}

	stopCh := make(chan struct{}, 1)

	scheduler := gocron.NewScheduler(location)
	for _, alert := range config.Alerts {
		_, err := scheduler.Cron(alert.CronExpression).Do(func() {
			logger := logger.WithField("alert", alert)

			out, err := exec.CommandContext(context.Background(), "googler", "--json", "--np", "--exact", alert.SearchPhrase).Output()
			if err != nil {
				logger.WithError(err).Errorf("Failed to execute command with search phrase '%s'", alert.SearchPhrase)
				stopCh <- struct{}{}
				return
			}

			server, err := newSMTPServer(config.SMTPClient)
			if err != nil {
				logger.WithError(err).Error("Failed to create SMTP server")
				stopCh <- struct{}{}
				return
			}

			smtpClient, err := server.Connect()
			if err != nil {
				logger.WithError(err).Error("Failed to connect to SMTP server")
				stopCh <- struct{}{}
				return
			}

			email := mail.NewMSG()
			email.AddTo(alert.EmailTo).SetSubject(alert.EmailSubject)
			email.SetBody(mail.TextPlain, string(out))
			err = email.Send(smtpClient)
			if err != nil {
				logger.WithError(err).Errorf("Failed to send email")
				stopCh <- struct{}{}
				return
			}
		})
		if err != nil {
			logger.WithError(err).Errorf("Failed to create job")
			os.Exit(1)
		}
	}

	scheduler.StartAsync()

	for {
		select {
		case <-stopCh:
			scheduler.Stop()
			os.Exit(1)
		case <-sigs:
			scheduler.Stop()
			os.Exit(0)
		}
	}
}

func newSMTPServer(config internal.SMTPClientConfig) (*mail.SMTPServer, error) {
	server := mail.NewSMTPClient()

	server.Host = config.Host
	server.Port = config.Port
	server.Username = config.Username
	server.Password = config.Password
	server.Encryption = mail.EncryptionSTARTTLS
	connectTimeout, err := time.ParseDuration(config.ConnectTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse duration from connect timeout")
	}
	server.ConnectTimeout = connectTimeout
	sendTimeout, err := time.ParseDuration(config.SendTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse duration from connect timeout")
	}
	server.SendTimeout = sendTimeout
	//server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return server, nil
}

package internal

import (
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
	"net/mail"
	"os"
	"time"
)

type Config struct {
	// TimeLocation allows configuring what timezone should be used for scheduling.
	// If the TimeLocation is "" or "UTC" UTC is used.
	// If the TimeLocation is "Local", Local is used.
	// Otherwise, the location corresponding to a file
	// in the IANA Time Zone database, such as "America/New_York" is used.
	TimeLocation string           `yaml:"time_location"`
	SMTPClient   SMTPClientConfig `yaml:"smtp_client"`

	Alerts []Alert `yaml:"alerts"`
}

func (c Config) validate() error {
	_, err := time.LoadLocation(c.TimeLocation)
	if err != nil {
		return errors.Wrap(err, "could not parse time location")
	}

	err = c.SMTPClient.validate()
	if err != nil {
		return err
	}

	for _, alert := range c.Alerts {
		if err = alert.validate(); err != nil {
			return err
		}
	}

	return nil
}

type SMTPClientConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	ConnectTimeout string `yaml:"connect_timeout"`
	SendTimeout    string `yaml:"send_timeout"`
}

func (c SMTPClientConfig) validate() error {
	if c.Host == "" {
		return errors.New("host must not be empty")
	}

	if 0 > c.Port || c.Port > 65535 {
		return errors.New("port must in range <0, 65535>")
	}

	if c.Username == "" {
		return errors.New("username must not be empty")
	}

	if c.Password == "" {
		return errors.New("password must not be empty")
	}

	_, err := time.ParseDuration(c.ConnectTimeout)
	if err != nil {
		return errors.Wrap(err, "could not parse connect timeout")
	}

	_, err = time.ParseDuration(c.SendTimeout)
	if err != nil {
		return errors.Wrap(err, "could not parse send timeout")
	}

	return nil
}

type Alert struct {
	CronExpression string `yaml:"cron_expression"`
	SearchPhrase   string `yaml:"search_phrase"`

	EmailTo      string `yaml:"email_to"`
	EmailSubject string `yaml:"email_subject"`
}

func (a Alert) validate() error {
	_, err := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month).Parse(a.CronExpression)
	if err != nil {
		return errors.Wrapf(err, "cron expression '%s' not valid", a.CronExpression)
	}

	if a.SearchPhrase == "" {
		return errors.New("search phrase must not be empty")
	}

	if _, err := mail.ParseAddress(a.EmailTo); err != nil {
		return errors.Wrap(err, "could not parse email to")
	}

	if a.EmailSubject == "" {
		return errors.New("email subject must not be empty")
	}

	return nil
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open config with path")
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	if err = config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

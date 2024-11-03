package main

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

var auth smtp.Auth

func InitMail() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("failed to load .env: %w", err)
	}

	auth = smtp.PlainAuth(
		"",
		os.Getenv("EMAIL_USER"),
		os.Getenv("EMAIL_PASS"),
		os.Getenv("EMAIL_HOST"),
	)

	return nil
}

func SendEmail(message string) error {
	contents := fmt.Sprintf(
		"Subject: Portfolio Site Message\n"+
			"To: %s\n"+
			"From: Portfolio Site\n\n"+
			"%s",
		os.Getenv("EMAIL_DEST"),
		message,
	)

	err := smtp.SendMail(
		os.Getenv("EMAIL_HOST")+":587",
		auth,
		os.Getenv("EMAIL_USER"),
		[]string{os.Getenv("EMAIL_DEST")},
		[]byte(contents),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

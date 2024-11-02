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

func SendEmail(message []byte) error {
	err := smtp.SendMail(
		os.Getenv("EMAIL_HOST")+":587",
		auth,
		os.Getenv("EMAIL_DEST"),
		[]string{os.Getenv("EMAIL_USER")},
		message,
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

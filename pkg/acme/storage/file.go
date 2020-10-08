package storage

import "fmt"

func privateKeyFileName(email string) string {
	return fmt.Sprintf("%s-acme-privateKey.pem", email)
}

func registrationFileName(email string) string {
	return fmt.Sprintf("%s-acme-account-registration.json", email)
}

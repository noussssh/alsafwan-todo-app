package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateStrongPassword() (string, error) {
	const (
		upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowerChars   = "abcdefghijklmnopqrstuvwxyz"
		numberChars  = "0123456789"
		specialChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)
	
	var password strings.Builder
	
	appendRandomChar := func(charset string) error {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return err
		}
		password.WriteByte(charset[n.Int64()])
		return nil
	}
	
	if err := appendRandomChar(upperChars); err != nil {
		return "", err
	}
	if err := appendRandomChar(lowerChars); err != nil {
		return "", err
	}
	if err := appendRandomChar(numberChars); err != nil {
		return "", err
	}
	if err := appendRandomChar(specialChars); err != nil {
		return "", err
	}
	
	allChars := upperChars + lowerChars + numberChars + specialChars
	for i := 0; i < 4; i++ {
		if err := appendRandomChar(allChars); err != nil {
			return "", err
		}
	}
	
	result := []rune(password.String())
	for i := len(result) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		result[i], result[j.Int64()] = result[j.Int64()], result[i]
	}
	
	return string(result), nil
}

func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	return nil
}

func ValidateName(name string) error {
	if len(name) < 2 || len(name) > 100 {
		return fmt.Errorf("name must be between 2 and 100 characters")
	}
	return nil
}

func ValidateCompany(company *string) error {
	if company == nil {
		return nil
	}
	
	validCompanies := []string{
		"Al Safwan Marine",
		"Louis Safety",
		"Data Grid Labs",
	}
	
	for _, valid := range validCompanies {
		if *company == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid company name")
}
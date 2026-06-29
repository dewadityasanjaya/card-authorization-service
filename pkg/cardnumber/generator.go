package cardnumber

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Generate creates a random 16-digit Visa card number starting with 4
func Generate() string {
	// Visa cards start with 4
	number := "4"

	// Generate 14 random digits
	for i := 0; i < 14; i++ {
		number += fmt.Sprintf("%d", rand.Intn(10))
	}

	// Add Luhn checksum digit
	number += fmt.Sprintf("%d", luhnCheckDigit(number))

	return number
}

// luhnCheckDigit calculates the Luhn check digit
func luhnCheckDigit(number string) int {
	sum := 0
	alternate := true

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return (10 - (sum % 10)) % 10
}

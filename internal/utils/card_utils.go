package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// LuhnCheck проверяет номер карты по алгоритму Луна
func LuhnCheck(cardNumber string) bool {
	var sum int
	var alternate bool

	for i := len(cardNumber) - 1; i >= 0; i-- {
		n, _ := strconv.Atoi(string(cardNumber[i]))
		if alternate {
			n *= 2
			if n > 9 {
				n = n%10 + 1
			}
		}
		sum += n
		alternate = !alternate
	}

	return sum%10 == 0
}

// GenerateCardNumber генерирует валидный номер карты (алгоритм Луна)
func GenerateCardNumber(bin string) string {
	rand.Seed(time.Now().UnixNano())

	if bin == "" {
		bin = "4" // Visa prefix
	}

	// Генерируем 15 цифр
	var number strings.Builder
	number.WriteString(bin)

	for i := len(bin); i < 15; i++ {
		number.WriteString(strconv.Itoa(rand.Intn(10)))
	}

	// Вычисляем контрольную цифру по алгоритму Луна
	numStr := number.String()
	var sum int
	for i := 0; i < len(numStr); i++ {
		n, _ := strconv.Atoi(string(numStr[i]))
		if (len(numStr)-i)%2 == 0 {
			n *= 2
			if n > 9 {
				n = n%10 + 1
			}
		}
		sum += n
	}

	checkDigit := (10 - (sum % 10)) % 10
	fullNumber := numStr + strconv.Itoa(checkDigit)

	return fullNumber
}

// GenerateCVV генерирует CVV код
func GenerateCVV() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%03d", rand.Intn(1000))
}

// MaskCardNumber маскирует номер карты (показывает только последние 4 цифры)
func MaskCardNumber(cardNumber string) string {
	if len(cardNumber) < 4 {
		return "****"
	}
	return "****" + cardNumber[len(cardNumber)-4:]
}

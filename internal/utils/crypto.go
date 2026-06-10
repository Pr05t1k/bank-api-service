package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash проверяет пароль
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashCVV хеширует CVV код
func HashCVV(cvv string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	return string(bytes), err
}

// ComputeHMAC вычисляет HMAC для данных
func ComputeHMAC(data string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC проверяет HMAC
func VerifyHMAC(data, hmacString string, secret []byte) bool {
	expectedHMAC := ComputeHMAC(data, secret)
	return hmac.Equal([]byte(expectedHMAC), []byte(hmacString))
}

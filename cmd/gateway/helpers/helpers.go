package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	Dockerize *bool
)

func ParseFlags() {
	Dockerize = flag.Bool("docker", false, "checks if program is started with docker")
	flag.Parse()
}

func IsAny(arr []string, s string) bool {
	for _, slice := range arr {
		if slice == s {
			return true
		}
	}
	return false
}

func IsPrintable(s string) bool {
	for _, char := range s {
		cond := char >= 32 && char <= 126
		if !cond {
			return false
		}
	}
	return true
}

func IsNameLenOk(s string) bool {
	if len(s) > 10 || len(s) < 3 {
		return false
	}
	return true
}

func CreateCookie() *http.Cookie {
	cookie := &http.Cookie{
		Name:    "potato_batat_bulba",
		Value:   generateSessionID(),
		Expires: time.Now().Add(time.Hour * 168),
		Path:    "/",
	}
	return cookie
}

func GetPasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func generateSessionID() string {
	// Generate a random byte slice with 32 bytes of entropy
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		// If there was an error generating random bytes, panic
		panic(err)
	}

	// Encode the byte slice as a base64 string to create the session ID
	sessionID := base64.URLEncoding.EncodeToString(randomBytes)

	return sessionID
}

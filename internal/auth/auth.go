package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(pw string) (string, error) {
	hash, err := argon2id.CreateHash(pw, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(pw, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(pw, hash)
	if err != nil {
		return false, err
	}
	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signingKey := []byte(tokenSecret)
	signed, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(jwt *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != "chirpy" {
		return uuid.Nil, errors.New("Wrong JWT issuer!")
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, err
	}

	return userUUID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")
	token, found := strings.CutPrefix(bearer, "Bearer ")
	if !found {
		return "", errors.New("Invalid Authorization header!")
	}
	return token, nil
}

func MakeRefreshToken() (string, error) {
	buf := make([]byte, 32)
	n, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	if n != 32 {
		return "", errors.New("Couldn't create refresh token!")
	}
	token := hex.EncodeToString(buf)
	return token, nil
}

func GetApiKey(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")
	apiKey, found := strings.CutPrefix(bearer, "ApiKey ")
	if !found {
		return "", errors.New("Invalid Authorization header!")
	}
	return apiKey, nil
}

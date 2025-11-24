package auth

import (
	"github.com/alexedwards/argon2id"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"fmt"
	"net/http"
	"crypto/rand"
	"encoding/hex"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, err
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer : "chirpy",
		IssuedAt : jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt : jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
		Subject : userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signed, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(
		tokenString, 
		claims, 
		func(t *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})
	if err != nil {
		var zeroUUID uuid.UUID
		return zeroUUID, err
	}

	uuidString, err := claims.GetSubject()
	if err != nil {
		var zeroUUID uuid.UUID
		return zeroUUID, err
	}
	uuidParsed, err := uuid.Parse(uuidString)
	if err != nil {
		var zeroUUID uuid.UUID
		return zeroUUID, err
	}
	return uuidParsed, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	v, ok := headers["Authorization"]
	if !ok {
		return "", fmt.Errorf("No authorization header in the request")
	}
	return v[0][7:], nil
}

func MakeRefreshToken() (string, error) {
	bytes := make([]byte, 256)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	s := hex.EncodeToString(bytes)
	return s, nil
}
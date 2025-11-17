package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	hash1, err1 := HashPassword("testpassword1122334455")
	hash2, err2 := HashPassword("testpassword1122334455")
	if err1 != nil || err2 != nil || hash1 == hash2 {
		t.Errorf("hash password not working")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	pw := "testpassword1122334455"
	hash1, _ := HashPassword(pw)
	hash2, _ := HashPassword(pw)
	check1, err1 := CheckPasswordHash(pw, hash1)
	check2, err2 := CheckPasswordHash(pw, hash2)
	if err1 != nil || err2 != nil || hash1 == hash2 || check1 != true || check2 != true {
		t.Errorf("check hash password not working")
	}
}

func TestMakeJWT(t *testing.T) {
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	uuid3 := uuid.New()
	secret1 := "test-secret-1"
	secret2 := "test-secret-2"
	secret3 := "test-secret-3"
	expire1 := 120 * time.Second
	expire2 := 180 * time.Second
	expire3 := 180 * time.Second
	sign1, err1 := MakeJWT(uuid1, secret1, expire1)
	sign2, err2 := MakeJWT(uuid2, secret2, expire2)
	sign3, err3 := MakeJWT(uuid3, secret3, expire3)
	if err1 != nil || err2 != nil || err3 != nil || sign1 == sign2 || sign1 == sign3 || sign2 == sign3 {
		t.Errorf("MakeJWT not working")
	}
}

func TestValidateJWT(t *testing.T) {
	uuid1 := uuid.New()
	secret1 := "test-secret-1"
	expire1 := 120 * time.Second
	sign1, _ := MakeJWT(uuid1, secret1, expire1)
	valid1, err1 := ValidateJWT(sign1, secret1)

	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}
	if valid1 != uuid1 {
		t.Fatalf("expected %v, got %v", uuid1, valid1)
	}

	signExpired, _ := MakeJWT(uuid.New(), "s", -1*time.Second)
	_, err := ValidateJWT(signExpired, "s")
	if err == nil {
		t.Fatal("expected error for expired token")
	}

	got, err := ValidateJWT(sign1, "wrong-secret")
	if err == nil || got != uuid.Nil {
		t.Fatal("expected error and zero UUID for wrong secret")
	}

	_, err = ValidateJWT("not.a.jwt.signature", "test-secret-1")
    if err == nil {
        t.Fatal("expected error for malformed token")
    }
}
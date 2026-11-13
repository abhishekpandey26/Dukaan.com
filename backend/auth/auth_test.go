package auth

import "testing"

// Run with: go test ./...
func TestTokenRoundTrip(t *testing.T) {
	token, err := GenerateToken("secret", 42, "admin")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ParseToken("secret", token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != 42 || claims.Role != "admin" {
		t.Errorf("got user=%d role=%q, want 42 admin", claims.UserID, claims.Role)
	}

	if _, err := ParseToken("wrong-secret", token); err == nil {
		t.Error("expected error when using the wrong secret")
	}
}

func TestPassword(t *testing.T) {
	hash, err := HashPassword("hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "hunter2") {
		t.Error("correct password was rejected")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("wrong password was accepted")
	}
}

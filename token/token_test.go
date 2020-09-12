package token

import "testing"

const secret = "hgde34jnbvcdewscvbhytrewq5678kncxcnbvcxswqw34fvbkuytr"

func TestGetToken(t *testing.T) {
	authService := New(secret)
	email := "abuarquemf@gmail.com"
	if _, err := authService.GetToken(email); err != nil {
		t.Errorf("want error nil, got %q", err)
	}
}

func TestIsValid(t *testing.T) {
	authService := New(secret)
	email := "abuarquemf@gmail.com"
	token, err := authService.GetToken(email)
	if err != nil {
		t.Errorf("want error nil, got %q", err)
	}
	isValid := authService.IsValid(token)
	if isValid == false {
		t.Errorf("expected to have a valid token")
	}
}

func TestGetClaims(t *testing.T) {
	authService := New(secret)
	email := "abuarquemf@gmail.com"
	token, err := authService.GetToken(email)
	if err != nil {
		t.Errorf("want error nil, got %q", err)
	}
	claims, err := GetClaims(token)
	if err != nil {
		t.Errorf("want err nil when getting claims")
	}
	if email != claims["email"] {
		t.Errorf("want email %s, got %s", email, claims["email"])
	}
}

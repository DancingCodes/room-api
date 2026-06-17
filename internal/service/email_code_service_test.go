package service

import "testing"

func TestNormalizeEmailForCode(t *testing.T) {
	got, err := normalizeEmailForCode("  Alex@Example.COM  ")
	if err != nil {
		t.Fatalf("normalizeEmailForCode() error = %v", err)
	}

	want := "alex@example.com"
	if got != want {
		t.Fatalf("normalizeEmailForCode() = %q, want %q", got, want)
	}
}

func TestNormalizeEmailForCodeRejectsInvalidEmail(t *testing.T) {
	if _, err := normalizeEmailForCode("not-email"); err == nil {
		t.Fatal("normalizeEmailForCode() error is nil, want error")
	}
}

func TestGenerateEmailCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := generateEmailCode()
		if err != nil {
			t.Fatalf("generateEmailCode() error = %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("generateEmailCode() length = %d, want 6", len(code))
		}
		for _, r := range code {
			if r < '0' || r > '9' {
				t.Fatalf("generateEmailCode() = %q, want digits only", code)
			}
		}
	}
}

package service

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got := normalizeEmail("  Alex@Example.COM  ")
	want := "alex@example.com"
	if got != want {
		t.Fatalf("normalizeEmail() = %q, want %q", got, want)
	}
}

func TestNicknameBase(t *testing.T) {
	tests := []struct {
		email string
		want  string
	}{
		{email: "alex@example.com", want: "alex"},
		{email: "abcdefghijk@example.com", want: "abcdefgh"},
		{email: "@example.com", want: "用户"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := nicknameBase(tt.email); got != tt.want {
				t.Fatalf("nicknameBase() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRuneLen(t *testing.T) {
	got := runeLen("Room一二")
	want := 6
	if got != want {
		t.Fatalf("runeLen() = %d, want %d", got, want)
	}
}

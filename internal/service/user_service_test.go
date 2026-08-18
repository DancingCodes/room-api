package service

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got := normalizeEmail("  Alex@Example.COM  ")
	want := "alex@example.com"
	if got != want {
		t.Fatalf("normalizeEmail() = %q, want %q", got, want)
	}
}

func TestValidateUserFields(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		nickname  string
		avatarURL string
		wantErr   bool
	}{
		{
			name:      "valid",
			email:     "alex@example.com",
			nickname:  "Alex",
			avatarURL: "https://example.com/avatar.png",
		},
		{
			name:      "invalid email",
			email:     "not-email",
			nickname:  "Alex",
			avatarURL: "https://example.com/avatar.png",
			wantErr:   true,
		},
		{
			name:      "nickname too long by rune",
			email:     "alex@example.com",
			nickname:  "一二三四五六七八九",
			avatarURL: "https://example.com/avatar.png",
			wantErr:   true,
		},
		{
			name:      "avatar required",
			email:     "alex@example.com",
			nickname:  "Alex",
			avatarURL: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserFields(tt.email, tt.nickname, tt.avatarURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateUserFields() error = %v, wantErr %v", err, tt.wantErr)
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

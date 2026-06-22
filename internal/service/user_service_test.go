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
		account   string
		email     string
		password  string
		nickname  string
		avatarURL string
		wantErr   bool
	}{
		{
			name:      "valid",
			account:   "alex_001",
			email:     "alex@example.com",
			password:  "123456",
			nickname:  "Alex",
			avatarURL: "https://example.com/avatar.png",
		},
		{
			name:      "account too short",
			account:   "abc",
			email:     "alex@example.com",
			password:  "123456",
			nickname:  "Alex",
			avatarURL: "https://example.com/avatar.png",
			wantErr:   true,
		},
		{
			name:      "account has invalid char",
			account:   "alex-001",
			email:     "alex@example.com",
			password:  "123456",
			nickname:  "Alex",
			avatarURL: "https://example.com/avatar.png",
			wantErr:   true,
		},
		{
			name:      "invalid email",
			account:   "alex_001",
			email:     "not-email",
			password:  "123456",
			nickname:  "Alex",
			avatarURL: "https://example.com/avatar.png",
			wantErr:   true,
		},
		{
			name:      "password too short",
			account:   "alex_001",
			email:     "alex@example.com",
			password:  "12345",
			nickname:  "Alex",
			avatarURL: "https://example.com/avatar.png",
			wantErr:   true,
		},
		{
			name:      "nickname too long by rune",
			account:   "alex_001",
			email:     "alex@example.com",
			password:  "123456",
			nickname:  "一二三四五六七八九",
			avatarURL: "https://example.com/avatar.png",
			wantErr:   true,
		},
		{
			name:      "avatar required",
			account:   "alex_001",
			email:     "alex@example.com",
			password:  "123456",
			nickname:  "Alex",
			avatarURL: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserFields(tt.account, tt.email, tt.password, tt.nickname, tt.avatarURL)
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

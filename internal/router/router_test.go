package router

import (
	"strings"
	"testing"

	"room-api/internal/config"
)

func TestNewRequiresJWTSecret(t *testing.T) {
	_, err := New(config.Config{}, nil)
	if err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("New() error = %v, want JWT_SECRET validation error", err)
	}
}

package service

import "testing"

func TestNormalizeMessageLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{
			name:  "keeps valid value",
			limit: 10,
			want:  10,
		},
		{
			name:  "uses default for zero",
			limit: 0,
			want:  defaultMessageLimit,
		},
		{
			name:  "uses default for negative",
			limit: -1,
			want:  defaultMessageLimit,
		},
		{
			name:  "caps large value",
			limit: 100,
			want:  maxMessageLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMessageLimit(tt.limit)
			if got != tt.want {
				t.Fatalf("normalizeMessageLimit() = %d, want %d", got, tt.want)
			}
		})
	}
}

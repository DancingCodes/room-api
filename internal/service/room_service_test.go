package service

import (
	"math"
	"testing"
)

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{
			name:         "keeps valid values",
			page:         2,
			pageSize:     30,
			wantPage:     2,
			wantPageSize: 30,
		},
		{
			name:         "uses defaults for non-positive values",
			page:         0,
			pageSize:     0,
			wantPage:     defaultPage,
			wantPageSize: defaultPageSize,
		},
		{
			name:         "caps large page size",
			page:         1,
			pageSize:     100,
			wantPage:     1,
			wantPageSize: maxPageSize,
		},
		{
			name:         "resets overflow page",
			page:         math.MaxInt,
			pageSize:     2,
			wantPage:     defaultPage,
			wantPageSize: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotPageSize := normalizePage(tt.page, tt.pageSize)
			if gotPage != tt.wantPage || gotPageSize != tt.wantPageSize {
				t.Fatalf("normalizePage() = (%d, %d), want (%d, %d)", gotPage, gotPageSize, tt.wantPage, tt.wantPageSize)
			}
		})
	}
}

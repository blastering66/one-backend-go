package unit

import (
	"testing"

	"github.com/one-backend-go/internal/pkg/pagination"
)

func TestPaginationClamp(t *testing.T) {
	tests := []struct {
		name         string
		input        pagination.Params
		wantPage     int64
		wantPageSize int64
	}{
		{"defaults", pagination.Params{Page: 0, PageSize: 0}, 1, 10},
		{"negative page", pagination.Params{Page: -5, PageSize: 20}, 1, 20},
		{"page_size above max", pagination.Params{Page: 1, PageSize: 100}, 1, 50},
		{"valid values", pagination.Params{Page: 3, PageSize: 25}, 3, 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.input
			p.Clamp()
			if p.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
			}
			if p.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.wantPageSize)
			}
		})
	}
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		pageSize int64
		want     int64
	}{
		{"exact division", 20, 10, 2},
		{"remainder", 21, 10, 3},
		{"single page", 5, 10, 1},
		{"zero items", 0, 10, 0},
		{"zero page size", 10, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pagination.TotalPages(tt.total, tt.pageSize)
			if got != tt.want {
				t.Errorf("TotalPages(%d, %d) = %d, want %d", tt.total, tt.pageSize, got, tt.want)
			}
		})
	}
}

func TestPaginationSkip(t *testing.T) {
	p := pagination.Params{Page: 3, PageSize: 10}
	if got := p.Skip(); got != 20 {
		t.Errorf("Skip() = %d, want 20", got)
	}
}

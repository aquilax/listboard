package main

import (
	"reflect"
	"testing"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name string
		pc   PaginationConfig
		want Pages
	}{
		{
			name: "returns empty list if total = 0",
			pc: PaginationConfig{
				ipp:   10,
				page:  0,
				total: 0,
			},
			want: Pages{},
		},
		{
			name: "returns empty list if total < ipp",
			pc: PaginationConfig{
				ipp:   10,
				page:  0,
				total: 3,
			},
			want: Pages{},
		},
		{
			name: "returns correct number of pages",
			pc: PaginationConfig{
				ipp:   10,
				page:  0,
				total: 13,
				url:   "http://example.com",
				param: "page",
			},
			want: Pages{
				Page{1, ""},
				Page{2, "http://example.com?page=2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Pagination(tt.pc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Pagination() = %v, want %v", got, tt.want)
			}
		})
	}
}

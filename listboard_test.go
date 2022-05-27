package main

import (
	"testing"
)

func TestNewListBoard(t *testing.T) {
	t.Run("ListBoard is not nil", func(t *testing.T) {
		got := NewListBoard()
		if got == nil {
			t.Errorf("NewListBoard() = %v", got)
		}
	})
}

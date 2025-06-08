package random

import (
	"testing"
)

func TestIntWeak(t *testing.T) {
	tests := []struct {
		name string
		from int
		to   int
	}{
		{"t1", 0, 1},
		{"t1", 0, 2},
		{"t1", 10, 15},
		{"t1", -5, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntWeak(tt.from, tt.to)

			// fmt.Printf("IntWeak() = %v, from = %v, to = %v\n", got, tt.from, tt.to)

			if got > tt.to || got < tt.from {
				t.Errorf("IntWeak() = %v, from = %v, to = %v", got, tt.from, tt.to)
			}
		})
	}
}

func TestIntStrong(t *testing.T) {
	tests := []struct {
		name string
		from int
		to   int
	}{
		{"t1", 0, 1},
		{"t1", 0, 2},
		{"t1", 10, 15},
		{"t1", -5, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntStrong(tt.from, tt.to)

			// fmt.Printf("IntWeak() = %v, from = %v, to = %v\n", got, tt.from, tt.to)

			if got > tt.to || got < tt.from {
				t.Errorf("IntStrong() = %v, from = %v, to = %v", got, tt.from, tt.to)
			}
		})
	}
}

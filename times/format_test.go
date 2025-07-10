package times

import (
	"testing"
	"time"
)

func TestFmtDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "zero hours",
			d:    15 * time.Minute,
			want: "00h15m",
		},
		{
			name: "one hour",
			d:    time.Hour + 15*time.Minute,
			want: "01h15m",
		},
		{
			name: "negative hours",
			d:    -(time.Hour + 15*time.Minute),
			want: "-01h15m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FmtDuration(tt.d); got != tt.want {
				t.Errorf("FmtDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

package times

import (
	"testing"
	"time"
)

func TestFmtDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		secs []bool
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
		{
			name: "seconds",
			d:    15*time.Minute + 10*time.Second,
			secs: []bool{true},
			want: "00h15m10s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FmtDuration(tt.d, tt.secs...); got != tt.want {
				t.Errorf("FmtDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

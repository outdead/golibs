package times

import (
	"testing"
	"time"
)

func TestSQLTime_Scan(t *testing.T) {
	sourceTime, _ := time.Parse("2006-01-02 15:04:05", "2025-07-10 02:09:00")

	tests := []struct {
		name      string
		t         SQLTime
		v         interface{}
		wantValue string
		wantErr   bool
	}{
		{
			name:      "nil",
			t:         SQLTime{},
			v:         nil,
			wantValue: "0001-01-01 00:00:00",
			wantErr:   false,
		},
		{
			name:      "time.Time",
			t:         SQLTime{},
			v:         sourceTime,
			wantValue: "2025-07-10 02:09:00",
			wantErr:   false,
		},
		{
			name:      "string",
			t:         SQLTime{},
			v:         "2025-07-10 02:09:00",
			wantValue: "2025-07-10 02:09:00",
			wantErr:   false,
		},
		{
			name:      "byte",
			t:         SQLTime{},
			v:         []byte("2025-07-10 02:09:00"),
			wantValue: "2025-07-10 02:09:00",
			wantErr:   false,
		},
		{
			name:      "error 1",
			t:         SQLTime{},
			v:         1 * time.Second,
			wantValue: "0001-01-01 00:00:00",
			wantErr:   true,
		},
		{
			name:      "error 2",
			t:         SQLTime{},
			v:         "",
			wantValue: "0001-01-01 00:00:00",
			wantErr:   true,
		},
		{
			name:      "error 3",
			t:         SQLTime{},
			v:         []byte("qqq"),
			wantValue: "0001-01-01 00:00:00",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.t.Scan(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			if time.Time(tt.t).Format(DateTimeLayout) != tt.wantValue {
				t.Errorf("Scan() got = %s, want %s", time.Time(tt.t).Format(DateTimeLayout), tt.wantValue)
			}
		})
	}
}

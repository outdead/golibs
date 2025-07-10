package times

import "testing"

func TestParseOnlyTime(t *testing.T) {
	tests := []struct {
		name       string
		strtime    string
		wantHour   uint
		wantMinute uint
		wantSecond uint
		wantErr    bool
	}{
		{
			name:       "hour, minutes and seconds",
			strtime:    "13:10:15",
			wantHour:   13,
			wantMinute: 10,
			wantSecond: 15,
			wantErr:    false,
		},
		{
			name:       "hour and minutes",
			strtime:    "03:12",
			wantHour:   3,
			wantMinute: 12,
			wantSecond: 0,
			wantErr:    false,
		},
		{
			name:       "hour only 1",
			strtime:    "03",
			wantHour:   3,
			wantMinute: 0,
			wantSecond: 0,
			wantErr:    false,
		},
		{
			name:       "hour only 2",
			strtime:    "3",
			wantHour:   3,
			wantMinute: 0,
			wantSecond: 0,
			wantErr:    false,
		},
		{
			name:       "invalid hour 1",
			strtime:    "25",
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
			wantErr:    true,
		},
		{
			name:       "invalid hour 2",
			strtime:    "s",
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
			wantErr:    true,
		},
		{
			name:       "invalid minutes 1",
			strtime:    "03:66",
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
			wantErr:    true,
		},
		{
			name:       "invalid minutes 2",
			strtime:    "03:g",
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
			wantErr:    true,
		},
		{
			name:       "invalid seconds 1",
			strtime:    "03:10:66",
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
			wantErr:    true,
		},
		{
			name:       "invalid seconds 2",
			strtime:    "03:10:q",
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHour, gotMinute, gotSecond, err := ParseOnlyTime(tt.strtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOnlyTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotHour != tt.wantHour {
				t.Errorf("ParseOnlyTime() gotHour = %v, want %v", gotHour, tt.wantHour)
			}

			if gotMinute != tt.wantMinute {
				t.Errorf("ParseOnlyTime() gotMinute = %v, want %v", gotMinute, tt.wantMinute)
			}

			if gotSecond != tt.wantSecond {
				t.Errorf("ParseOnlyTime() gotSecond = %v, want %v", gotSecond, tt.wantSecond)
			}
		})
	}
}

func TestParseOnlyTimeSafe(t *testing.T) {
	tests := []struct {
		name       string
		strtime    string
		wantHour   uint
		wantMinute uint
		wantSecond uint
	}{
		{
			name:       "hour, minutes and seconds",
			strtime:    "13:10:15",
			wantHour:   13,
			wantMinute: 10,
			wantSecond: 15,
		},
		{
			name:       "invalid hour",
			strtime:    "25",
			wantHour:   0,
			wantMinute: 0,
			wantSecond: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHour, gotMinute, gotSecond := ParseOnlyTimeSafe(tt.strtime)
			if gotHour != tt.wantHour {
				t.Errorf("ParseOnlyTimeSafe() gotHour = %v, want %v", gotHour, tt.wantHour)
			}

			if gotMinute != tt.wantMinute {
				t.Errorf("ParseOnlyTimeSafe() gotMinute = %v, want %v", gotMinute, tt.wantMinute)
			}

			if gotSecond != tt.wantSecond {
				t.Errorf("ParseOnlyTimeSafe() gotSecond = %v, want %v", gotSecond, tt.wantSecond)
			}
		})
	}
}

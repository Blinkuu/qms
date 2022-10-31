package timeunit

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	type args struct {
		unit string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ReturnsNoErrorAndSecondUnitForLowercaseSecondString",
			args: args{
				unit: "second",
			},
			want:    time.Second,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndSecondUnitForLowercaseMinuteString",
			args: args{
				unit: "minute",
			},
			want:    time.Minute,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndSecondUnitForLowercaseHourString",
			args: args{
				unit: "hour",
			},
			want:    time.Hour,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndSecondUnitForLowercaseDayString",
			args: args{
				unit: "day",
			},
			want:    24 * time.Hour,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsErrorAndZeroValueForUppercaseSecondString",
			args: args{
				unit: "SECOND",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForUppercaseMinuteString",
			args: args{
				unit: "MINUTE",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForUppercaseHourString",
			args: args{
				unit: "HOUR",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForUppercaseDayString",
			args: args{
				unit: "DAY",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForMixedCaseSecondString",
			args: args{
				unit: "SeCoNd",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForMixedCaseMinuteString",
			args: args{
				unit: "MiNuTe",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForMixedCaseHourString",
			args: args{
				unit: "HoUr",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForMixedCaseDayString",
			args: args{
				unit: "DaY",
			},
			want:    0,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.unit)
			if !tt.wantErr(t, err, fmt.Sprintf("Parse(%v)", tt.args.unit)) {
				return
			}

			assert.Equalf(t, tt.want, got, "Parse(%v)", tt.args.unit)
		})
	}
}

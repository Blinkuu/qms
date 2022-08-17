package timeunit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	type args struct {
		unit string
	}
	tests := []struct {
		name    string
		args    args
		want    Unit
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ReturnsNoErrorAndSecondUnitForLowercaseSecondString",
			args: args{
				unit: "second",
			},
			want:    Second,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndSecondUnitForLowercaseMinuteString",
			args: args{
				unit: "minute",
			},
			want:    Minute,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndSecondUnitForLowercaseHourString",
			args: args{
				unit: "hour",
			},
			want:    Hour,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsNoErrorAndSecondUnitForLowercaseDayString",
			args: args{
				unit: "day",
			},
			want:    Day,
			wantErr: assert.NoError,
		},
		{
			name: "ReturnsErrorAndZeroValueForUppercaseSecondString",
			args: args{
				unit: "SECOND",
			},
			want:    Unit(0),
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForUppercaseMinuteString",
			args: args{
				unit: "MINUTE",
			},
			want:    Unit(0),
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForUppercaseHourString",
			args: args{
				unit: "HOUR",
			},
			want:    Unit(0),
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForUppercaseDayString",
			args: args{
				unit: "DAY",
			},
			want:    Unit(0),
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForMixedCaseSecondString",
			args: args{
				unit: "SeCoNd",
			},
			want:    Unit(0),
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForMixedCaseMinuteString",
			args: args{
				unit: "MiNuTe",
			},
			want:    Unit(0),
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForMixedCaseHourString",
			args: args{
				unit: "HoUr",
			},
			want:    Unit(0),
			wantErr: assert.Error,
		},
		{
			name: "ReturnsErrorAndZeroValueForMixedCaseDayString",
			args: args{
				unit: "DaY",
			},
			want:    Unit(0),
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

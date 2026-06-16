package repository

import (
	"fmt"
	"testing"

	models "github.com/Happy-skills/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCheckMetric(t *testing.T) {
	type args struct {
		mType  string
		mValue string
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "positive gauge",
			args: args{mValue: "1.23588", mType: models.Gauge},
			want: nil,
		},
		{
			name: "positive counter",
			args: args{mValue: "56", mType: models.Counter},
			want: nil,
		},
		{
			name: "negative gauge",
			args: args{mValue: "-1.23588", mType: models.Gauge},
			want: nil,
		},
		{
			name: "negative counter",
			args: args{mValue: "-56", mType: models.Counter},
			want: fmt.Errorf("metric value is negative for %s", models.Counter),
		},
		{
			name: "empty gauge",
			args: args{mValue: "", mType: models.Gauge},
			want: fmt.Errorf("metric value is empty for %s", models.Gauge),
		},
		{
			name: "empty counter",
			args: args{mValue: "", mType: models.Counter},
			want: fmt.Errorf("metric value is empty for %s", models.Counter),
		},
		{
			name: "string in gauge",
			args: args{mValue: "test", mType: models.Gauge},
			want: fmt.Errorf("metric value is invalid for %s, can't parse float", models.Gauge),
		},
		{
			name: "string in counter",
			args: args{mValue: "test", mType: models.Counter},
			want: fmt.Errorf("metric value is invalid for %s, can't parse int", models.Counter),
		},
		{
			name: "int in gauge",
			args: args{mValue: "10", mType: models.Gauge},
			want: nil,
		},
		{
			name: "float in counter",
			args: args{mValue: "56.5489742", mType: models.Counter},
			want: fmt.Errorf("metric value is invalid for %s, can't parse int", models.Counter),
		},
		{
			name: "unknown type",
			args: args{mValue: "56", mType: "string"},
			want: fmt.Errorf("metric value with unknown type for string"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkError := CheckMetric(tt.args.mType, tt.args.mValue)
			assert.Equal(t, checkError, tt.want, "CheckMetric() = %v, want %v", checkError, tt.want)
		})
	}
}

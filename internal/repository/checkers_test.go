package repository

import (
	"testing"

	models "github.com/Happy-skills/metrics/internal/model"
)

func TestCheckMetric(t *testing.T) {
	type args struct {
		mType  string
		mValue string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "positive gauge",
			args: args{mValue: "1.23588", mType: models.Gauge},
			want: 1,
		},
		{
			name: "positive counter",
			args: args{mValue: "56", mType: models.Counter},
			want: 1,
		},
		{
			name: "negative gauge",
			args: args{mValue: "-1.23588", mType: models.Gauge},
			want: 0,
		},
		{
			name: "negative counter",
			args: args{mValue: "-56", mType: models.Counter},
			want: 0,
		},
		{
			name: "empty gauge",
			args: args{mValue: "", mType: models.Gauge},
			want: 0,
		},
		{
			name: "empty counter",
			args: args{mValue: "", mType: models.Counter},
			want: 0,
		},
		{
			name: "string in gauge",
			args: args{mValue: "test", mType: models.Gauge},
			want: 0,
		},
		{
			name: "string in counter",
			args: args{mValue: "test", mType: models.Counter},
			want: 0,
		},
		{
			name: "int in gauge",
			args: args{mValue: "10", mType: models.Gauge},
			want: 1,
		},
		{
			name: "float in counter",
			args: args{mValue: "56.5489742", mType: models.Counter},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckMetric(tt.args.mType, tt.args.mValue); got != tt.want {
				t.Errorf("CheckMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

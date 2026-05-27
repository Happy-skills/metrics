package repository

import (
	"reflect"
	"strconv"

	models "github.com/Happy-skills/metrics/internal/model"
)

func CheckMetric(mType string, mValue string) int {
	if mType == models.Gauge {
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return 0
		}
		if reflect.TypeOf(val).Kind() != reflect.Float64 {
			return 0
		}
		return 1
	}
	if mType == models.Counter {
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return 0
		}
		if reflect.TypeOf(val).Kind() != reflect.Int64 {
			return 0
		}
		return 1
	}

	return 0
}

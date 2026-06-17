package repository

import (
	"fmt"
	"strconv"

	models "github.com/Happy-skills/metrics/internal/model"
)

func CheckMetric(mType string, mValue string) error {
	if len(mValue) == 0 {
		return fmt.Errorf("metric value is empty for %s", mType)
	}
	if mType == models.Gauge {
		_, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return fmt.Errorf("metric value is invalid for %s, can't parse float", mType)
		}
		return nil
	}
	if mType == models.Counter {
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return fmt.Errorf("metric value is invalid for %s, can't parse int", mType)
		}
		if val < 0 {
			return fmt.Errorf("metric value is negative for %s", mType)
		}
		return nil
	}

	return fmt.Errorf("metric value with unknown type for %s", mType)
}

package repository

import (
	"fmt"
	"strconv"
)

func toGauge(input any) (float64, error) {
	switch v := input.(type) {
	case int:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		result, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}

		return result, nil
	}

	return 0, fmt.Errorf("cannot convert to gauge, unexpected type: %T", input)
}

func toCounter(input any) (int64, error) {
	switch v := input.(type) {
	case int:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case string:
		result, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}

		return result, nil
	}

	return 0, fmt.Errorf("cannot convert to counter, unexpected type: %T", input)
}

package retrier

import (
	"fmt"
	"time"

	"github.com/Happy-skills/metrics/internal/logger"
)

var maxRetries = 3

func SetRetries(value int) {
	maxRetries = value
}

type Retry func() (isRetriable bool, err error)

func Retrier(fn Retry) error {
	var err error
	var isRetriable bool

	for attempt := 0; attempt <= maxRetries; attempt++ {
		isRetriable, err = fn()
		if err == nil {
			return nil
		}

		if !isRetriable {
			return err
		}

		if attempt+1 < maxRetries { // skip last iteration
			logger.Sugar.Debugf("retrying after %d attempts", attempt+1)

			// i=0 -> 0*2+1=1
			// i=1 -> 1*2+1=3
			// i=2 -> 2*2+1=5
			time.Sleep(time.Duration(attempt*2+1) * time.Second)
		}
	}

	return fmt.Errorf("max retries exceeded (%d): %w", maxRetries, err)
}

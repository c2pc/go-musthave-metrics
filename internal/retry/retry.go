package retry

import (
	"errors"
	"fmt"
	"time"
)

var MaxAttemptsError = fmt.Errorf("max attempts exceeded")

func Retry(fn func() error, needRetry func(error) bool, delays []time.Duration) error {
	for attempts := 0; attempts < len(delays); attempts++ {
		err := fn()
		if err == nil {
			return nil
		}
		if needRetry(err) {
			if attempts < len(delays)-1 {
				time.Sleep(delays[attempts])
				continue
			}
			return errors.Join(MaxAttemptsError, err)
		} else {
			return err
		}
	}

	return MaxAttemptsError
}

package nimble

import (
	"fmt"
	"os"
	"time"
)

type RetryConfig struct {
	NumAttempts         int
	DelayBetweenRetries time.Duration
	LogEnabled          bool
	LogOperationName    string
}

// Retry will execute the provided executor function multiple times with an interim delay, reporting the final error status in the event of a terminal failure
func Retry(config *RetryConfig, executor func() error) error {
	if config.NumAttempts == 0 {
		config.NumAttempts = 3
	}
	if config.DelayBetweenRetries == 0 {
		config.DelayBetweenRetries = time.Second * 5
	}

	var err error
	for i := 0; i < config.NumAttempts; i++ {
		err = executor()
		if err == nil {
			return nil
		}

		if i < config.NumAttempts-1 {
			if config.LogEnabled {
				opName := config.LogOperationName
				if opName == "" {
					opName = "operation"
				}
				fmt.Fprintf(os.Stderr, "%s failed, retrying in %gs (%d of %d)...", opName, config.DelayBetweenRetries.Seconds(), i+1, config.NumAttempts)
			}
			time.Sleep(config.DelayBetweenRetries)
		}
	}
	return err
}

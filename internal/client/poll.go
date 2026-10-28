package client

import (
	"context"
	"fmt"
	"time"
)

// pollInterval is how often Poll calls check.
var pollInterval = 5 * time.Second

// Poll calls check until it reports done, returns an error, the context is
// cancelled, or timeout elapses.
func Poll(ctx context.Context, timeout time.Duration, check func() (done bool, err error)) error {
	deadline := time.Now().Add(timeout)
	for {
		done, err := check()
		if err != nil || done {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s", timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

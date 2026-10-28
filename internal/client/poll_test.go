package client

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPoll(t *testing.T) {
	pollInterval = time.Millisecond
	defer func() { pollInterval = 5 * time.Second }()

	calls := 0
	err := Poll(context.Background(), time.Minute, func() (bool, error) {
		calls++
		return calls == 3, nil
	})
	if err != nil || calls != 3 {
		t.Errorf("expected success after 3 calls, got %d calls, err %v", calls, err)
	}

	boom := errors.New("boom")
	if err := Poll(context.Background(), time.Minute, func() (bool, error) { return false, boom }); !errors.Is(err, boom) {
		t.Errorf("expected check error, got %v", err)
	}

	if err := Poll(context.Background(), 0, func() (bool, error) { return false, nil }); err == nil {
		t.Error("expected timeout error")
	}
}

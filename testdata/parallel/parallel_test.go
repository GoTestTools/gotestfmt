package parallel

import (
	"testing"
	"time"
)

func TestParallel1(t *testing.T) {
	t.Parallel()
	t.Logf("Test message 1")
	time.Sleep(5 * time.Second)
	t.Logf("Test message 2")
}

func TestParallel2(t *testing.T) {
	t.Parallel()
	time.Sleep(5 * time.Second)
	t.Logf("Test message 1")
	time.Sleep(5 * time.Second)
	t.Logf("Test message 2")
}

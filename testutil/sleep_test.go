package testutil

import (
	"testing"
	"time"
)

// TestSleep is an empty test for the purposes of seeing a long-running test in the output.
func TestSleep(t *testing.T) {
	t.Logf("Now sleeping for 5 seconds...")
	time.Sleep(5 * time.Second)
	t.Logf("Welcome back, now finishing test.")
}

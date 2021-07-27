package gomod

import (
	"testing"
)

func TestSubtest(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		t.Logf("Hello world!")
	})
	t.Run("test2", func(t *testing.T) {
		t.Fatalf("Here's an error.")
	})
	t.Run("test3", func(t *testing.T) {
		t.Skipf("Let's skip this one...")
	})
}

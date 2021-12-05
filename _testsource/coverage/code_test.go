package example_test

import (
	"testing"

	"github.com/haveyoudebuggedit/example"
)

func TestHello(t *testing.T) {
	if example.Hello() != "Hello world!" {
		t.Fail()
	}
}

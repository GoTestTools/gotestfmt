package example_test

import (
	"testing"

	"github.com/gotesttools/example"
)

func TestHello(t *testing.T) {
	if example.Hello() != "Hello world!" {
		t.Fail()
	}
}

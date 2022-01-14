package logging_test

import (
	"testing"

	klog "k8s.io/klog/v2"
)

func TestGoLogging(t *testing.T) {
	t.Parallel()
	t.Logf("Hello world!")
}

func TestKLog(t *testing.T) {
	t.Parallel()
	klog.Info("This is an info message")
	klog.Warning("This is a warning message")
	klog.Error("This is an error message")
}

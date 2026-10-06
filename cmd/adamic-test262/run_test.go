package main

import (
	"strings"
	"testing"
	"time"
)

func TestOutputOverflowTimeout(t *testing.T) {
	endless := runCommand(time.Second, nil, "sh", "-c", "yes")
	if !endless.TimedOut || !crashedExecution(endless) {
		t.Errorf("endless printer: timedOut=%v crashed=%v", endless.TimedOut, crashedExecution(endless))
	}
	got := decide(verdictInput{Node: execution{Stdout: "x", Exit: 0}, Native: endless})
	if got.Kind != outcomeCrashed {
		t.Errorf("endless verdict: %+v", got)
	}
}

func TestOutputOverflowSanitizer(t *testing.T) {
	for _, stream := range []string{"", " >&2"} {
		t.Run(stream, func(t *testing.T) {
			sanitizer := runCommand(10*time.Second, nil, "sh", "-c", "yes | head -c 300000"+stream+"; echo '==1==ERROR: AddressSanitizer: heap-use-after-free' >&2; exit 1")
			if !strings.Contains(sanitizer.Stderr, "AddressSanitizer") || !crashedExecution(sanitizer) {
				t.Errorf("sanitizer lost: stderr bytes=%d crashed=%v", len(sanitizer.Stderr), crashedExecution(sanitizer))
			}
			if len(sanitizer.Stderr) > outputLimit+len("command output exceeded capture limit\n") {
				t.Errorf("unbounded stderr: %d", len(sanitizer.Stderr))
			}
			got := decide(verdictInput{Node: execution{Exit: 0}, Native: sanitizer})
			if got.Kind != outcomeCrashed {
				t.Errorf("sanitizer verdict: %+v", got)
			}
		})
	}
}

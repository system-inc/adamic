package childguard

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type capturePrefix struct {
	kept    []byte
	total   int
	limit   int
	failure error
}

func (w *capturePrefix) Write(p []byte) (int, error) {
	w.total += len(p)
	if w.failure != nil {
		return 0, w.failure
	}
	remaining := w.limit - len(w.kept)
	w.kept = append(w.kept, p[:min(remaining, len(p))]...)
	return len(p), nil
}

func TestCaptureToBoundedWriters(test *testing.T) {
	test.Parallel()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, diagnostic := capturePrefix{limit: 4096}, capturePrefix{limit: 4096}
	if captureToError := CaptureTo(timeoutContext, child("large"), &out, &diagnostic); captureToError != nil {
		test.Fatal(captureToError)
	}
	want := strings.Repeat(strings.Repeat("x", 1023)+"\n", 4)
	for name, writer := range map[string]*capturePrefix{"stdout": &out, "stderr": &diagnostic} {
		if writer.total != 1<<20 || string(writer.kept) != want {
			test.Fatalf(
				"%s: drained %d bytes, kept %d, want 1 MiB drained and its 4096-byte prefix",
				name,
				writer.total,
				len(writer.kept),
			)
		}
	}
}

func TestCaptureToDrainsAfterWriterError(test *testing.T) {
	test.Parallel()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	failure := errors.New("destination failed")
	out := capturePrefix{failure: failure}
	cmd := child("large")
	captureToError := CaptureTo(timeoutContext, cmd, &out, nil)
	if !errors.Is(captureToError, failure) ||
		timeoutContext.Err() != nil ||
		cmd.ProcessState == nil ||
		!cmd.ProcessState.Success() {
		test.Fatalf(
			"destination failure must still drain and reap child: err=%v context=%v state=%v",
			captureToError,
			timeoutContext.Err(),
			cmd.ProcessState,
		)
	}
}

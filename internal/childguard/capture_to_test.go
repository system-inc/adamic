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

func TestCaptureToBoundedWriters(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, diagnostic := capturePrefix{limit: 4096}, capturePrefix{limit: 4096}
	if err := CaptureTo(ctx, child("large"), &out, &diagnostic); err != nil {
		t.Fatal(err)
	}
	want := strings.Repeat(strings.Repeat("x", 1023)+"\n", 4)
	for name, writer := range map[string]*capturePrefix{"stdout": &out, "stderr": &diagnostic} {
		if writer.total != 1<<20 || string(writer.kept) != want {
			t.Fatalf("%s: drained %d bytes, kept %d, want 1 MiB drained and its 4096-byte prefix", name, writer.total, len(writer.kept))
		}
	}
}

func TestCaptureToDrainsAfterWriterError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	failure := errors.New("destination failed")
	out := capturePrefix{failure: failure}
	cmd := child("large")
	err := CaptureTo(ctx, cmd, &out, nil)
	if !errors.Is(err, failure) || ctx.Err() != nil || cmd.ProcessState == nil || !cmd.ProcessState.Success() {
		t.Fatalf("destination failure must still drain and reap child: err=%v context=%v state=%v", err, ctx.Err(), cmd.ProcessState)
	}
}

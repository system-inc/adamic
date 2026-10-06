package yaml

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestSharedSliceAppendGap(t *testing.T) {
	entry, err := filepath.Abs("gaps/sharedSliceAppend.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	expected := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "0")
	if string(expected) != "a\nx\n" {
		t.Fatalf("Node %q", expected)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "shared")
	if err := native.Build(native.C(lowered), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	actual := run(t, "", nil, binary, "0")
	if string(actual) != "x\nx\n" {
		t.Fatalf("gap changed or closed: native %q Node %q; remove the workaround", actual, expected)
	}
	sanitized := filepath.Join(t.TempDir(), "shared-sanitized")
	if err := native.Build(native.C(lowered), sanitized, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, sanitized, "48")
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	if err == nil || !strings.Contains(stderr.String(), "ERROR: AddressSanitizer: heap-buffer-overflow") {
		t.Fatalf("gap changed or closed: %v %s", err, stderr.Bytes())
	}
	t.Logf("Node stdout %q, release native stdout %q; ASan caught heap-buffer-overflow for end slice", expected, actual)
}

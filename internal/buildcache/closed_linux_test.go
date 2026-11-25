//go:build linux && !mips && !mipsle && !mips64 && !mips64le

package buildcache

import (
	"os"
	"path/filepath"
	"testing"
)

// The kernel's own answer, asked live: a directory this process made is writable to it (the writable-directory case),
// and a file it made read-only, in a directory it made read-only, isn't. Skipped as root, whose access ignores mode
// bits.
// Mutant: writableByThisProcess's `case err == nil:` answering `return false, nil`: "a directory it made" goes red.
func TestTheKernelAnswersWhatThisProcessMayWrite(t *testing.T) {
	t.Parallel()
	if os.Geteuid() == 0 {
		t.Skip("running as root, whose access ignores mode bits, so nothing here is closed to it")
	}
	directory := t.TempDir()
	closed := filepath.Join(directory, "closed")
	write(t, closed, "file.c", "int file;\n")
	file := filepath.Join(closed, "file.c")
	if err := os.Chmod(file, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(closed, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(closed, 0o755) })
	for _, each := range []struct {
		name, path string
		writable   bool
	}{
		{"a directory it made", directory, true},
		{"a read-only file in a read-only directory", file, false},
	} {
		if writable, err := writableByThisProcess(each.path); err != nil || writable != each.writable {
			t.Errorf("%s: writable %t (%v), want %t", each.name, writable, err, each.writable)
		}
	}
}

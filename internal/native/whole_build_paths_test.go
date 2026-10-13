package native

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"testing"
)

// A sanitized whole build's binary names no temporary object of clang's (Tests, from Compiler, Oct 12): on macOS ld's
// debug map names each object it links with debug info by its path, and a compile and link in one call linked main.c's
// object from <TMPDIR>/main-<random>.o, random per build. Compiled first to main.o and linked by that name, the debug
// map names it main.o alone. On Linux the linker names no input, so this holds there either way. Mutant (on macOS):
// main.c compiled and linked in one call again (the debug map names clang's temporary object).
func TestAWholeBuildNamesNoTemporaryObject(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "whole")
	if err := Build(splitProbe, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	if temporary := regexp.MustCompile(`main-[0-9A-Za-z]+\.o`).Find(content); temporary != nil {
		t.Errorf("the binary names clang's temporary object %s", temporary)
	}
	if goruntime.GOOS == "darwin" && !bytes.Contains(content, []byte("\x00main.o\x00")) {
		t.Error("the binary's debug map doesn't name main.o by that name alone")
	}
}

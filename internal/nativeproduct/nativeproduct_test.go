package nativeproduct

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

const hello = "#include <stdio.h>\nint main(void) { puts(\"hello\"); return 0; }\n"

// The same C with the same options is one product however it is asked for, another source or other options is
// another, and the binary runs what the source says.
func TestBuildKeysTheSourceAndTheOptions(t *testing.T) {
	t.Parallel()
	binary := Build(t, hello, native.Options{})
	output, err := exec.Command(binary).Output()
	if err != nil || string(output) != "hello\n" {
		t.Fatalf("the product ran %q, %v", output, err)
	}
	if again := Build(t, hello, native.Options{}); again != binary {
		t.Fatalf("the same source and options built twice: %s and %s", binary, again)
	}
	name := func(source string, options native.Options) string {
		return strings.Join(Inputs(source, options).Flags, "\n")
	}
	if name(hello, native.Options{}) == name(strings.Replace(hello, "hello", "howdy", 1), native.Options{}) {
		t.Fatal("another source left the key as it was")
	}
	if name(hello, native.Options{}) == name(hello, native.Options{Sanitize: true}) {
		t.Fatal("sanitizing left the key as it was")
	}
}

package native

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Both expression and array initializer contexts must preserve every byte and the terminator.
func TestLongCNames(t *testing.T) {
	t.Parallel()
	value := strings.Repeat("n", longestLiteral+1)
	checkLongCNames(t, value, cString(value), cArray(value))
}

func TestLongCNamesKeepUnsignedBytes(t *testing.T) {
	t.Parallel()
	value := strings.Repeat("é", longestLiteral)
	checkLongCNames(t, value, cString(value), cArray(value))
}

func checkLongCNames(t *testing.T, value, expression, initializer string) {
	t.Helper()
	directory := t.TempDir()
	source := fmt.Sprintf(`#include <stdio.h>
#include <string.h>
static const char message[] = %s;
static const char *name = %s;
int main(void) {
 const char *local = %s;
 if (sizeof message != %d || message[sizeof message - 1] != 0 || memcmp(message, name, sizeof message) || memcmp(message, local, sizeof message)) return 1;
 return fwrite(message, 1, sizeof message - 1, stdout) != sizeof message - 1;
}
`, initializer, expression, expression, len(value)+1)
	path := filepath.Join(directory, "names.c")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	binary := filepath.Join(directory, "names")
	command := exec.CommandContext(ctx, "clang", "-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", path, "-o", binary)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, output)
	}
	output, err := exec.CommandContext(ctx, binary).Output()
	if err != nil || string(output) != value {
		t.Fatalf("name bytes differ: exit %v, got %d bytes, want %d", err, len(output), len(value))
	}
}

// Inspect before clang so an overlong literal is caught by this guard itself.
func TestLongCNameEmission(t *testing.T) {
	t.Parallel()
	checked, err := load.Load([]string{"../oracle/testdata/long_field_name.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	if strings.Contains(source, "static const char message[] = ((const char[])") {
		t.Fatal("message array needs an initializer, not a compound expression")
	}
	for _, literal := range regexp.MustCompile(`"(?:\\.|[^"\\])*"`).FindAllString(source, -1) {
		if len(literal)-2 > longestLiteral {
			t.Fatalf("emitted a %d-byte C literal for a long name or message", len(literal)-2)
		}
	}
}

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
func TestLongCNames(test *testing.T) {
	test.Parallel()
	value := strings.Repeat("n", longestLiteral+1)
	checkLongCNames(test, value, cString(value), cArray(value))
}

func TestLongCNamesKeepUnsignedBytes(test *testing.T) {
	test.Parallel()
	value := strings.Repeat("é", longestLiteral)
	checkLongCNames(test, value, cString(value), cArray(value))
}

func checkLongCNames(test *testing.T, value, expression, initializer string) {
	test.Helper()
	directory := test.TempDir()
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
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	binary := filepath.Join(directory, "names")
	command := exec.CommandContext(
		timeoutContext,
		"clang",
		"-std=c11",
		"-Wall",
		"-Wextra",
		"-Werror",
		"-pedantic",
		path,
		"-o",
		binary,
	)
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("compile: %v\n%s", combinedOutputError, output)
	}
	output, outputError := exec.CommandContext(timeoutContext, binary).Output()
	if outputError != nil || string(output) != value {
		test.Fatalf("name bytes differ: exit %v, got %d bytes, want %d", outputError, len(output), len(value))
	}
}

// Inspect before clang so an overlong literal is caught by this guard itself.
func TestLongCNameEmission(test *testing.T) {
	test.Parallel()
	checked, loadError := load.Load([]string{"../oracle/testdata/long_field_name.a"})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	source := C(program)
	if strings.Contains(source, "static const char message[] = ((const char[])") {
		test.Fatal("message array needs an initializer, not a compound expression")
	}
	for _, literal := range regexp.MustCompile(`"(?:\\.|[^"\\])*"`).FindAllString(source, -1) {
		if len(literal)-2 > longestLiteral {
			test.Fatalf("emitted a %d-byte C literal for a long name or message", len(literal)-2)
		}
	}
}

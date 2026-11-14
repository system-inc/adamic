package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Probe only, never for landing (#g4arkr4 step 1): which construct of a union-of-array rest emits C that
// clang refuses. Each program prints Node's output, whether it lowers, and where native compile fails,
// clang's first error lines and the emitted C around the line clang names.
func TestG4aProbe(t *testing.T) {
	directory := "internal/oracle/testdata/g4a_probe"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	mainLine := regexp.MustCompile(`main\.c:([0-9]+):`)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".ts" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			name := strings.TrimSuffix(entry.Name(), ".ts")
			absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			t.Logf("PROBE %s node exit=%d stdout=%q stderr=%q", name, truth.exitCode, cut(string(truth.stdout), 200), cut(string(truth.stderr), 200))
			program, err := lowered(t, absolute)
			if err != nil {
				t.Logf("PROBE %s lower stops: %s", name, cut(strings.ReplaceAll(err.Error(), "\n", " / "), 400))
				return
			}
			t.Logf("PROBE %s lowers", name)
			code := native.C(program)
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				message := err.Error()
				clangErrors := 0
				for _, line := range strings.Split(message, "\n") {
					if strings.Contains(line, "error:") && clangErrors < 3 {
						clangErrors++
						t.Logf("PROBE %s clang: %s", name, cut(strings.TrimSpace(line), 400))
					}
				}
				if clangErrors == 0 {
					t.Logf("PROBE %s clang (no error: line): %s", name, cut(strings.ReplaceAll(message, "\n", " / "), 400))
				}
				match := mainLine.FindStringSubmatch(message)
				if match == nil {
					t.Logf("PROBE %s clang names no main.c line", name)
					return
				}
				number, err := strconv.Atoi(match[1])
				if err != nil {
					t.Fatal(err)
				}
				lines := strings.Split(code, "\n")
				for index := max(number-4, 0); index < min(number+3, len(lines)); index++ {
					t.Logf("PROBE %s C %d: %s", name, index+1, cut(lines[index], 240))
				}
				return
			}
			result := execute(t, binary)
			t.Logf("PROBE %s sanitized exit=%d stdout=%q stderr=%q", name, result.exitCode, cut(string(result.stdout), 200), cut(string(result.stderr), 200))
			script := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s javascript exit=%d stdout=%q", name, script.exitCode, cut(string(script.stdout), 200))
		})
	}
}

// cut keeps a probe line compact.
func cut(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}

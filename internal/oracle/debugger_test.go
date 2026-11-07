package oracle

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/debugger_fail.a", true, false})
}

func TestDebuggerNativeEmitsNothing(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/debugger_fail.a"))
	if readError != nil {
		test.Fatal(readError)
	}
	traps := regexp.MustCompile(`\b(?:__builtin_(?:debug)?trap|raise|breakpoint|DebugBreak|__debugbreak)\s*\(`)
	for _, extension := range []string{".a", ".ts"} {
		test.Run(extension, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), "main"+extension)
			if writeError := os.WriteFile(path, source, 0o644); writeError != nil {
				test.Fatal(writeError)
			}
			checked, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			program, lowerError := lower.Lower(context.Background(), checked)
			if lowerError != nil {
				test.Fatal(lowerError)
			}
			code := native.C(program)
			if trap := traps.FindString(code); trap != "" {
				test.Fatalf("debugger emitted a trap or breakpoint call: %s", trap)
			}
			withoutDebugger := strings.ReplaceAll(string(source), "debugger;", "")
			checked, loadOverlayError := load.LoadOverlay([]string{path}, map[string]string{path: withoutDebugger})
			if loadOverlayError != nil {
				test.Fatal(loadOverlayError)
			}
			baseline, lowerError := lower.Lower(context.Background(), checked)
			if lowerError != nil {
				test.Fatal(lowerError)
			}
			if code != native.C(baseline) {
				test.Fatal("debugger emitted native code: C differs from the source without debugger")
			}
			if count := strings.Count(javascript.JavaScript(program), "debugger;"); count != 1 {
				test.Fatalf("JavaScript kept %d debugger statements, want 1", count)
			}
		})
	}
}

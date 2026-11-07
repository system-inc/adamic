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
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/debugger_fail.a"))
	if err != nil {
		test.Fatal(err)
	}
	traps := regexp.MustCompile(`\b(?:__builtin_(?:debug)?trap|raise|breakpoint|DebugBreak|__debugbreak)\s*\(`)
	for _, extension := range []string{".a", ".ts"} {
		test.Run(extension, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), "main"+extension)
			if err := os.WriteFile(path, source, 0o644); err != nil {
				test.Fatal(err)
			}
			checked, err := load.Load([]string{path})
			if err != nil {
				test.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), checked)
			if err != nil {
				test.Fatal(err)
			}
			code := native.C(program)
			if trap := traps.FindString(code); trap != "" {
				test.Fatalf("debugger emitted a trap or breakpoint call: %s", trap)
			}
			withoutDebugger := strings.ReplaceAll(string(source), "debugger;", "")
			checked, err = load.LoadOverlay([]string{path}, map[string]string{path: withoutDebugger})
			if err != nil {
				test.Fatal(err)
			}
			baseline, err := lower.Lower(context.Background(), checked)
			if err != nil {
				test.Fatal(err)
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

package lower

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// Keep the rejected programs as .a files so each call can also be tried with the CLI.
func TestInputSpreadCoverage(test *testing.T) {
	test.Parallel()
	paths, err := filepath.Glob("testdata/input-spread/*.a")
	if err != nil || len(paths) != 47 {
		test.Fatalf("fixtures: %d, %v", len(paths), err)
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			source, err := os.ReadFile(path)
			if err != nil {
				test.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if strings.HasSuffix(path, "_array.a") {
				var check *load.CheckError
				if !errors.As(err, &check) || !strings.Contains(
					err.Error(),
					"A spread argument must either have a tuple type or be passed to a rest parameter",
				) {
					test.Fatalf("got %v, want the checker's array spread diagnostic", err)
				}
				return
			}
			if err != nil {
				test.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			name := strings.Split(filepath.Base(path), "_")[0]
			var notYet *NotYet
			if strings.HasSuffix(path, "_local.a") {
				if !errors.As(err, &notYet) || strings.Contains(err.Error(), "a spread argument to ") {
					test.Fatalf("got %v, want a general NotYet without blaming a prelude function", err)
				}
				return
			}
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "a spread argument to "+name) {
				test.Fatalf("%s: got %v, want NotYet naming %s", source, err, name)
			}
			line := strings.Count(string(source[:strings.LastIndex(string(source), "\n")]), "\n") + 1
			if !strings.Contains(err.Error(), fmt.Sprintf("%s:%d:", filepath.Base(path), line)) {
				test.Fatalf("missing call location: %v", err)
			}
		})
	}
}

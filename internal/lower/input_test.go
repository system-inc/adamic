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
	paths, globError := filepath.Glob("testdata/input-spread/*.a")
	if globError != nil || len(paths) != 47 {
		test.Fatalf("fixtures: %d, %v", len(paths), globError)
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			source, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			program, loadError := load.Load([]string{path})
			if strings.HasSuffix(path, "_array.a") {
				var check *load.CheckError
				if !errors.As(loadError, &check) || !strings.Contains(
					loadError.Error(),
					"A spread argument must either have a tuple type or be passed to a rest parameter",
				) {
					test.Fatalf("got %v, want the checker's array spread diagnostic", loadError)
				}
				return
			}
			if loadError != nil {
				test.Fatal(loadError)
			}
			_, lowerError := Lower(context.Background(), program)
			name := strings.Split(filepath.Base(path), "_")[0]
			var notYet *NotYet
			if strings.HasSuffix(path, "_local.a") {
				if !errors.As(lowerError, &notYet) || strings.Contains(lowerError.Error(), "a spread argument to ") {
					test.Fatalf("got %v, want a general NotYet without blaming a prelude function", lowerError)
				}
				return
			}
			if !errors.As(lowerError, &notYet) || !strings.Contains(lowerError.Error(), "a spread argument to "+name) {
				test.Fatalf("%s: got %v, want NotYet naming %s", source, lowerError, name)
			}
			line := strings.Count(string(source[:strings.LastIndex(string(source), "\n")]), "\n") + 1
			if !strings.Contains(lowerError.Error(), fmt.Sprintf("%s:%d:", filepath.Base(path), line)) {
				test.Fatalf("missing call location: %v", lowerError)
			}
		})
	}
}

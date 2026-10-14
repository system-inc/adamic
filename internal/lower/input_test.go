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
func TestInputSpreadCoverage(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("testdata/input-spread/*.a")
	if err != nil || len(paths) != 47 {
		t.Fatalf("fixtures: %d, %v", len(paths), err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if strings.HasSuffix(path, "_array.a") {
				var check *load.CheckError
				if !errors.As(err, &check) || !strings.Contains(err.Error(), "A spread argument must either have a tuple type or be passed to a rest parameter") {
					t.Fatalf("got %v, want the checker's array spread diagnostic", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			name := strings.Split(filepath.Base(path), "_")[0]
			var notYet *NotYet
			if strings.HasSuffix(path, "_local.a") {
				if !errors.As(err, &notYet) || strings.Contains(err.Error(), "a spread argument to ") {
					t.Fatalf("got %v, want a general NotYet without blaming a prelude function", err)
				}
				return
			}
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "a spread argument to "+name) {
				t.Fatalf("%s: got %v, want NotYet naming %s", source, err, name)
			}
			line := strings.Count(string(source[:strings.LastIndex(string(source), "\n")]), "\n") + 1
			if !strings.Contains(err.Error(), fmt.Sprintf("%s:%d:", filepath.Base(path), line)) {
				t.Fatalf("missing call location: %v", err)
			}
		})
	}
}

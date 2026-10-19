package oracle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"paths.a", "errors.a", "closure.a", "nul.a", "walk-paths.a", "walk-errors.a"} {
		key := "internal/oracle/testdata/host_realpath/" + name
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{key, func(t *testing.T) (string, func() string) {
			path, err := filepath.Abs(filepath.Join(repository, key))
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := load.LoadDisplayed([]string{path}, func(file string) string {
				if file == path {
					return "<source>.a"
				}
				return countPortable(t, file, "")
			})
			if err != nil {
				t.Fatal(err)
			}
			program, err := lower.Lower(context.Background(), loaded)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			harness, err := os.ReadFile("library_node_realpath_test.go")
			if err != nil {
				t.Fatal(err)
			}
			// Key the actual filesystem/cwd/environment builder as well as emitted C;
			// temporary directory spelling changes no value count.
			inputs := countPortable(t, source, path) + "\nrealpath harness\n" + string(harness)
			return inputs, func() string {
				binary := filepath.Join(t.TempDir(), "counted")
				if err := native.Build(source, binary, native.Options{Count: true}); err != nil {
					t.Fatal(err)
				}
				got := realpathRun(t, realpathDirectory(t), binary)
				match := countsLine.FindSubmatch(got.stderr)
				if got.exitCode != 0 || match == nil {
					t.Fatalf("counted realpath fixture: exit %d stderr %q", got.exitCode, got.stderr)
				}
				return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", key, match[1], match[2], match[3], match[4], match[5], match[6])
			}
		}})
	}
}

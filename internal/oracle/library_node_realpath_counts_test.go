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
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{key, func(test *testing.T) (string, func() string) {
				path, err := filepath.Abs(filepath.Join(repository, key))
				if err != nil {
					test.Fatal(err)
				}
				loaded, err := load.LoadDisplayed([]string{path}, func(file string) string {
					if file == path {
						return "<source>.a"
					}
					return countPortable(test, file, "")
				})
				if err != nil {
					test.Fatal(err)
				}
				program, err := lower.Lower(context.Background(), loaded)
				if err != nil {
					test.Fatal(err)
				}
				source := native.C(program)
				harness, err := os.ReadFile("library_node_realpath_test.go")
				if err != nil {
					test.Fatal(err)
				}
				// Key the actual filesystem/cwd/environment builder as well as emitted C;
				// temporary directory spelling changes no value count.
				inputs := countPortable(test, source, path) + "\nrealpath harness\n" + string(harness)
				return inputs, func() string {
					binary := filepath.Join(test.TempDir(), "counted")
					if err := native.Build(source, binary, native.Options{Count: true}); err != nil {
						test.Fatal(err)
					}
					got := realpathRun(test, realpathDirectory(test), binary)
					match := countsLine.FindSubmatch(got.stderr)
					if got.exitCode != 0 || match == nil {
						test.Fatalf("counted realpath fixture: exit %d stderr %q", got.exitCode, got.stderr)
					}
					return fmt.Sprintf(
						"| %s | %s | %s | %s | %s | %s | %s |",
						key,
						match[1],
						match[2],
						match[3],
						match[4],
						match[5],
						match[6],
					)
				}
			}},
		)
	}
}

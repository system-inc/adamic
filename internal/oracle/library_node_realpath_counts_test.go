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
				path, absolutePathError := filepath.Abs(filepath.Join(repository, key))
				if absolutePathError != nil {
					test.Fatal(absolutePathError)
				}
				loaded, loadDisplayedError := load.LoadDisplayed([]string{path}, func(file string) string {
					if file == path {
						return "<source>.a"
					}
					return countPortable(test, file, "")
				})
				if loadDisplayedError != nil {
					test.Fatal(loadDisplayedError)
				}
				program, lowerError := lower.Lower(context.Background(), loaded)
				if lowerError != nil {
					test.Fatal(lowerError)
				}
				source := native.C(program)
				harness, readError := os.ReadFile("library_node_realpath_test.go")
				if readError != nil {
					test.Fatal(readError)
				}
				// Key the actual filesystem/cwd/environment builder as well as emitted C;
				// temporary directory spelling changes no value count.
				inputs := countPortable(test, source, path) + "\nrealpath harness\n" + string(harness)
				return inputs, func() string {
					binary := filepath.Join(test.TempDir(), "counted")
					if buildError := native.Build(source, binary, native.Options{Count: true}); buildError != nil {
						test.Fatal(buildError)
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

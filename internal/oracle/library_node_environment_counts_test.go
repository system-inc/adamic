package oracle

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func processEnvironmentFixtureEnvironment(name string) []string {
	if name == "nul.a" {
		return []string{"A=set", "ADAMIC_ENV_NUL_CODE=0"}
	}
	if name == "prototype.a" {
		return []string{"toString=own environment string"}
	}
	return []string{"ADAMIC_ENV_READ_SET=set", "ADAMIC_ENV_READ_EMPTY=", "ADAMIC_ENV_READ_UNICODE=héllo 世界 🌍"}
}

func init() {
	for _, name := range []string{"reads.a", "prototype.a", "nul.a", "mutations.a"} {
		key := "internal/oracle/testdata/process_env/" + name
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
			environment := processEnvironmentFixtureEnvironment(name)
			// Environment bytes are inputs to the measurement, never inherited state.
			runs := countPortable(t, source, path) + "\nenvironment " + strings.Join(environment, "\n")
			return runs, func() string {
				binary := filepath.Join(t.TempDir(), "program")
				if err := native.Build(source, binary, native.Options{Count: true}); err != nil {
					t.Fatal(err)
				}
				got := processEnvironmentRun(t, environment, binary)
				match := countsLine.FindSubmatch(got.stderr)
				if got.exitCode != 0 || match == nil {
					t.Fatalf("counted environment fixture: exit %d stderr %q", got.exitCode, got.stderr)
				}
				return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", key, match[1], match[2], match[3], match[4], match[5], match[6])
			}
		}})
	}
}

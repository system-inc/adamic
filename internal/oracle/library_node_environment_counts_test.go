package oracle

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/oracle/countproducts"
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
	for _, name := range []string{"reads.a", "prototype.a", "nul.a", "mutations.a", "mutation_values.a", "mutation_special_names.a"} {
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
			// It runs with exactly these variables, as processEnvironmentRun runs it, and must finish.
			row := countproducts.Row{Key: key, C: source, Run: countproducts.Environment, Variables: environment, Checked: true}
			countRows.Store(key, row)
			return runs, func() string { return countMeasuredHere(t, row) }
		}})
	}
}

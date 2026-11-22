package oracle

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/oracle/countproducts"
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
			// The filesystem, working directory and environment it runs in are countproducts' (its realpath layout, as
			// realpathDirectory lays one out), which key its shard's product with countproducts' files; temporary
			// directory spelling changes no value count. It must finish.
			row := countproducts.Row{Key: key, C: source, Run: countproducts.Realpath, Checked: true}
			countRows.Store(key, row)
			return countPortable(t, source, path) + "\nrealpath", func() string { return countMeasuredHere(t, row) }
		}})
	}
}

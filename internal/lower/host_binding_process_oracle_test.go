package lower

import (
	"encoding/gob"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Ordinary runs check that the real table refuses all three typed process declarations. The row
// matches unknown exactly, so these fixtures must stop at their declarations. With the oracle
// environment set, this test runs alone in a subprocess with a synthetic row to exercise the model
// judgment and read representation. Isolation keeps that private replacement out of parallel tests.
func TestHostBindingProcessOracleIR(t *testing.T) {
	t.Parallel()
	of := os.Getenv("ADAMIC_PROCESS_ORACLE_TYPE")
	if of == "" {
		for _, of := range []string{"boolean", "string", "number"} {
			t.Run(of, func(t *testing.T) {
				path, err := filepath.Abs(filepath.Join("..", "oracle", "testdata", "host_bindings", "process_browser_"+of, "main.ts"))
				if err != nil {
					t.Fatal(err)
				}
				loaded, err := load.Load([]string{path})
				if err != nil {
					t.Fatal(err)
				}
				_, err = Lower(t.Context(), loaded)
				var notYet *NotYet
				var refused *Refused
				var where, what string
				switch {
				case errors.As(err, &notYet):
					where, what = notYet.Where, notYet.What
				case errors.As(err, &refused):
					where, what = refused.Where, refused.What
				}
				want := filepath.Join(filepath.Dir(path), "core.ts") + ":3:1"
				if where != want || what != "an ambient value declaration" {
					t.Fatalf("want an ambient value declaration stop at %s, got %v", want, err)
				}
			})
		}
		return
	}
	if flag.Lookup("test.run").Value.String() != "^TestHostBindingProcessOracleIR$" {
		t.Fatal("the typed-row builder must be the only selected test")
	}
	switch of {
	case "boolean", "string", "number":
	default:
		t.Fatalf("no synthetic process row for %q", of)
	}
	output := os.Getenv("ADAMIC_PROCESS_ORACLE_IR")
	if output == "" {
		t.Fatal("no oracle IR destination")
	}
	rows, err := load.HostBindings()
	if err != nil {
		t.Fatal(err)
	}
	for index := range rows {
		if rows[index].Name == "process" {
			rows[index].Declaration.Type = "{ nextTick?: unknown; browser?: " + of + "; } | undefined"
		}
	}
	previous := hostBindingRows
	hostBindingRows = func() ([]load.HostBinding, error) { return rows, nil }
	t.Cleanup(func() { hostBindingRows = previous })
	path, err := filepath.Abs(filepath.Join("..", "oracle", "testdata", "host_bindings", "process_browser_"+of, "main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := Lower(t.Context(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(output)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	ir.RegisterGob()
	if err := gob.NewEncoder(file).Encode(program); err != nil {
		t.Fatal(err)
	}
}

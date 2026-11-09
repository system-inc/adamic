package lower

import (
	"encoding/gob"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// The oracle runs this test alone in a subprocess. The real row matches unknown exactly; these
// three synthetic rows exercise the model judgment and read representation without changing that
// production boundary. Isolation keeps the private hostBindingRows replacement out of parallel tests.
func TestHostBindingProcessOracleIR(t *testing.T) {
	t.Parallel()
	of := os.Getenv("ADAMIC_PROCESS_ORACLE_TYPE")
	if of == "" {
		t.Skip("only the oracle's isolated typed-row builder runs this test")
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

package split_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func compilingOracleFixtures(t *testing.T) []string {
	t.Helper()
	inputs, err := buildcache.GoTestInputs("", "oracle.test", "./internal/oracle", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "oracle compiling fixtures"
	directory := buildcache.BuiltElsewhere(t, "oracle's compiling fixture registry", inputs,
		"internal/oracle TestProduct_OracleCompilingFixtures", func(string) error {
			return fmt.Errorf("run internal/oracle TestProduct_OracleCompilingFixtures ahead")
		})
	data, err := os.ReadFile(filepath.Join(directory, "fixtures.json"))
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		t.Fatal(err)
	}
	var selected []string
	for _, path := range paths {
		// Keep the snapshot family's original corpus directory. Supporting modules
		// and unregistered review witnesses aren't compiling oracle entry points.
		// A .ts entry is a program as an .a is, its downcasts checked rather than proven.
		if strings.HasPrefix(path, "internal/oracle/testdata/") && (filepath.Ext(path) == ".a" || filepath.Ext(path) == ".ts") {
			selected = append(selected, filepath.Join("../oracle/testdata", strings.TrimPrefix(path, "internal/oracle/testdata/")))
		}
	}
	if len(selected) == 0 {
		t.Fatal("oracle registered no compiling .a or .ts fixtures")
	}
	return selected
}

func TestProduct_SplitOracleFixtures(t *testing.T) { t.Parallel(); compilingOracleFixtures(t) }

package estree

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Preserve the shared library checks while using immutable oracle products in
// these owned tests; other workers retain their shared helper unchanged.
func estreeCheckOriginalLibraries(t *testing.T, cases []string, postprocessedGaps int) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	estreeCheckOriginalManifest(t, manifest(t, cases), cases, postprocessedGaps)
}
func estreeCheckOriginalManifest(t *testing.T, list string, cases []string, postprocessedGaps int) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	oracle := estreeOracleProduct(t)
	for _, mode := range []string{"raw", "postprocessed"} {
		flag := "--raw-json"
		if mode != "raw" {
			flag = "--json"
		}
		want := strings.Split(strings.TrimSpace(string(estreeOracleOutput(t, oracle, flag, list))), "\n")
		script, err := filepath.Abs("testdata/library.mjs")
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Split(strings.TrimSpace(string(execute(t, "", "node", script, library, mode, list))), "\n")
		if len(got) != len(want) {
			t.Fatalf("library %s: %d versus %d files", mode, len(got), len(want))
		}
		differences := 0
		for index := range want {
			var left, right any
			if err := json.Unmarshal([]byte(want[index]), &left); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(got[index]), &right); err != nil {
				t.Fatal(err)
			}
			leftBytes, _ := json.Marshal(left)
			rightBytes, _ := json.Marshal(right)
			if !bytes.Equal(leftBytes, rightBytes) {
				source := cases[index]
				known := mode == "postprocessed" && (source == "x\u00a0;" || source == "x\u2028;" || source == "x\r\n;")
				if !known {
					t.Errorf("unrecorded %s case %d: Go %s; library %s", mode, index, leftBytes, rightBytes)
					continue
				}
				ast := left.(map[string]any)["ast"].(map[string]any)
				statement := ast["body"].([]any)[0].(map[string]any)
				if source == "x\r\n;" {
					if !bytes.Equal(mustJSON(t, ast["range"]), []byte("[0,4]")) || !bytes.Equal(mustJSON(t, statement["range"]), []byte("[0,4]")) {
						t.Fatal("CRLF gap changed")
					}
					ast["range"] = []int{0, 3}
					statement["range"] = []int{0, 3}
				} else {
					if statement["__contentEnd"] != float64(2) {
						t.Fatal("multibyte whitespace gap changed")
					}
					statement["__contentEnd"] = 1
				}
				if !bytes.Equal(mustJSON(t, left), rightBytes) {
					t.Fatalf("known gap has additional differences: %s case %d: Go after exact known delta %s; library %s", mode, index, mustJSON(t, left), rightBytes)
				}
				differences++
				t.Logf("proved %s case %d known gap for %q", mode, index, source)
			}
		}
		t.Logf("%s: %d files, %d identical, %d differing", mode, len(want), len(want)-differences, differences)
		expected := 0
		if mode == "postprocessed" {
			expected = postprocessedGaps
		}
		if differences != expected {
			t.Errorf("expected exactly %d documented gaps, got %d", expected, differences)
		}
	}
}

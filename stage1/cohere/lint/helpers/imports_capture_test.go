package helpers

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"os"
	"testing"
)

func importsCapturedInputs(t *testing.T, path string) []map[string]string {
	t.Helper()
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var rows []map[string]string
	s := bufio.NewScanner(z)
	s.Buffer(make([]byte, 65536), 16<<20)
	for s.Scan() {
		var row map[string]string
		if e = json.Unmarshal(s.Bytes(), &row); e != nil {
			t.Fatal(e)
		}
		rows = append(rows, row)
	}
	if e = s.Err(); e != nil {
		t.Fatal(e)
	}
	return rows
}

// A missing consumer cannot earn agreement through controls alone.
func TestImportsCaptureCompleteness(t *testing.T) {
	var ledger struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	data, err := os.ReadFile("readiness.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	for _, set := range []struct{ name, symbol string }{{"call", "github.com/system-inc/cohere/internal/lint/ecmascript/imports.CallExpressionSource"}, {"segment", "github.com/system-inc/cohere/internal/lint/ecmascript/imports.HasPathSegment"}} {
		observed := map[string]int{}
		for _, row := range importsCapturedInputs(t, "imports/testdata/"+set.name+"/sources.jsonl.gz") {
			observed[row["rule"]]++
		}
		consumers := 0
		for _, row := range ledger.Remaining {
			for _, helper := range row.Helpers {
				if helper == set.symbol {
					consumers++
					if observed[row.Rule] == 0 {
						t.Fatalf("%s omitted consumer %s", set.symbol, row.Rule)
					}
				}
			}
		}
		if consumers == 0 {
			t.Fatal("consumer ledger drift", set.symbol)
		}
		t.Logf("%s: all %d consuming rules have captured upstream inputs", set.symbol, consumers)
	}
}

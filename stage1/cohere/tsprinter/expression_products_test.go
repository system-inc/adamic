package tsprinter

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func TestTSPrinterCachedMetadataRelocation(test *testing.T) {
	test.Parallel()
	first, second, product, copy := test.TempDir(), test.TempDir(), test.TempDir(), test.TempDir()
	original := oracles.PrinterCase{
		Label:  first + "/case.ts:3:expression",
		Source: "'repo/keep.ts'",
		Want:   "typescript/keep.ts",
	}
	data, _ := json.Marshal([]oracles.PrinterCase{original})
	if err := os.WriteFile(product+"/cases.json", data, 0644); err != nil {
		test.Fatal(err)
	}
	report, _ := json.Marshal(
		map[string]any{"file_refusals": map[string]string{first + "/case.ts": first + "/case.ts: refused"}, "cases": 1},
	)
	if err := os.WriteFile(product+"/coverage.json", report, 0644); err != nil {
		test.Fatal(err)
	}
	if err := oracles.TSPrinterRewriteMetadata(product, first, "", true); err != nil {
		test.Fatal(err)
	}
	canonical, err := os.ReadFile(product + "/cases.json")
	if err != nil {
		test.Fatal(err)
	}
	for _, name := range []string{"cases.json", "coverage.json"} {
		data, err := os.ReadFile(product + "/" + name)
		if err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(copy+"/"+name, data, 0644); err != nil {
			test.Fatal(err)
		}
	}
	if err := oracles.TSPrinterRewriteMetadata(copy, second, "", false); err != nil {
		test.Fatal(err)
	}
	data, _ = os.ReadFile(copy + "/cases.json")
	var relocated []oracles.PrinterCase
	if err := json.Unmarshal(data, &relocated); err != nil {
		test.Fatal(err)
	}
	want := original
	want.Label = second + "/case.ts:3:expression"
	if len(relocated) != 1 || relocated[0] != want {
		test.Fatalf("relocation changed an oracle case: %+v", relocated)
	}
	unchanged, _ := os.ReadFile(product + "/cases.json")
	if !bytes.Equal(unchanged, canonical) {
		test.Fatal("materialization changed shared product")
	}
	if err := oracles.TSPrinterRewriteMetadata(copy, second, "", true); err != nil {
		test.Fatal(err)
	}
	roundtrip, _ := os.ReadFile(copy + "/cases.json")
	if !bytes.Equal(roundtrip, canonical) {
		test.Fatal("metadata differs after relocating and canonicalizing")
	}
}

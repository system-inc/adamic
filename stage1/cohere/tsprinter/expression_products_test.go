package tsprinter

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

func TestTsPrinterCachedMetadataRelocation(test *testing.T) {
	test.Parallel()
	first, second, product, copy := test.TempDir(), test.TempDir(), test.TempDir(), test.TempDir()
	original := oracles.PrinterCase{
		Label:  first + "/case.ts:3:expression",
		Source: "'repo/keep.ts'",
		Want:   "typescript/keep.ts",
	}
	data, _ := json.Marshal([]oracles.PrinterCase{original})
	if writeError := os.WriteFile(product+"/cases.json", data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	report, _ := json.Marshal(
		map[string]any{"file_refusals": map[string]string{first + "/case.ts": first + "/case.ts: refused"}, "cases": 1},
	)
	if writeError := os.WriteFile(product+"/coverage.json", report, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if tsPrinterRewriteMetadataError := oracles.TSPrinterRewriteMetadata(
		product,
		first,
		"",
		true,
	); tsPrinterRewriteMetadataError != nil {
		test.Fatal(tsPrinterRewriteMetadataError)
	}
	canonical, readError := os.ReadFile(product + "/cases.json")
	if readError != nil {
		test.Fatal(readError)
	}
	for _, name := range []string{"cases.json", "coverage.json"} {
		data, readError := os.ReadFile(product + "/" + name)
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(copy+"/"+name, data, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if tsPrinterRewriteMetadataError := oracles.TSPrinterRewriteMetadata(
		copy,
		second,
		"",
		false,
	); tsPrinterRewriteMetadataError != nil {
		test.Fatal(tsPrinterRewriteMetadataError)
	}
	data, _ = os.ReadFile(copy + "/cases.json")
	var relocated []oracles.PrinterCase
	if decodeError := json.Unmarshal(data, &relocated); decodeError != nil {
		test.Fatal(decodeError)
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
	if tsPrinterRewriteMetadataError := oracles.TSPrinterRewriteMetadata(
		copy,
		second,
		"",
		true,
	); tsPrinterRewriteMetadataError != nil {
		test.Fatal(tsPrinterRewriteMetadataError)
	}
	roundtrip, _ := os.ReadFile(copy + "/cases.json")
	if !bytes.Equal(roundtrip, canonical) {
		test.Fatal("metadata differs after relocating and canonicalizing")
	}
}

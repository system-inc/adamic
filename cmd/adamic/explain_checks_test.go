package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplainChecksCountsIntegratedContracts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for name, source := range map[string]string{
		"tsconfig.json": `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"exactOptionalPropertyTypes":false,"noUncheckedIndexedAccess":false,"lib":["es2024"],"types":[]},"files":["main.ts"]}`,
		"main.ts":       `const items: string[] = ['ready']; const first: string = items[0]; const text: string = JSON.stringify(undefined); try { throw {message: "ready"}; } catch(error) { const message: string = error.message; console.log(message); } console.log(first);`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	lowered, code := compile(filepath.Join(directory, "main.ts"))
	if lowered == nil || code != 0 {
		t.Fatalf("integrated contracts did not lower: exit %d", code)
	}
	var report bytes.Buffer
	explainChecks(lowered, &report)
	want := "checked: indexed-presence=1 catch-error=0 json-stringify-defined=1 optional-write=0 caught-type=1\ntrusted: 0\n"
	if !strings.Contains(report.String(), want) {
		t.Fatalf("missing actual guard counts: %s", report.String())
	}
}

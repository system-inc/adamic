package load

import (
	"errors"
	"strings"
	"testing"
)

func TestJSONStringifyContract(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.a", `const result: string = JSON.stringify(undefined);`})
	_, err := Load(paths)
	if err == nil || !strings.Contains(err.Error(), "Fix: JSON.stringify can return undefined") || !strings.Contains(err.Error(), "fallback with ??") {
		t.Fatalf("missing truthful .a error and fix: %v", err)
	}
	project := projectProgram(t, `{"strict":true,"lib":["es2024"],"types":[]}`, `const result: string = JSON.stringify(undefined); const unrelated: number = 'bad';`)
	_, err = Load(project)
	var rejected *CheckError
	if !errors.As(err, &rejected) || len(rejected.OptionSites) != 1 || rejected.OptionSites[0].Options[0] != "JSON.stringify" || len(rejected.Diagnostics) != 1 || !strings.Contains(rejected.Diagnostics[0], "Type 'string' is not assignable to type 'number'") {
		t.Fatalf("ordinary error must survive JSON conversion: %v", err)
	}
	project = projectProgram(t, `{"strict":true,"lib":["es2024"],"types":[]}`, `const result: string | undefined = JSON.stringify(undefined);`)
	loaded, err := Load(project)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.OptionSites()) != 0 {
		t.Fatalf("observation must not require a defined result: %v", loaded.OptionSites())
	}
}

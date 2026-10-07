package checker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func TestCallbackParameterFacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "file.ts")
	source := "declare const yes:{then(...callbacks:((x:number)=>void)[]):void};yes;\n"
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	start := uint64(strings.LastIndex(source, "yes;"))
	end := start + 3
	ask := func(question string) []string {
		t.Helper()
		wire, err := program.Inspect(file, start, end, "Identifier", question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	own := ask("raw-shape")
	property := ask("property-shape\n" + own[5] + "\nthen")
	callback := ask("callback-parameters\n" + property[5])
	old := ask("call-parameters\n" + property[5])
	if callback[4] != "1" || old[4] != "1" {
		t.Fatal("missing callback roots", callback, old)
	}
	id, err := strconv.Atoi(callback[5])
	if err != nil {
		t.Fatal(err)
	}
	oldID, err := strconv.Atoi(old[5])
	if err != nil {
		t.Fatal(err)
	}
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), program.Compiler.GetSourceFile(file))
	defer release()
	expected := checker.Checker_getIndexTypeOfType(c, program.typesByID[oldID-1], checker.Checker_numberType(c))
	if program.typesByID[id-1] != expected {
		t.Fatal("callback fact differs from direct rest index query")
	}
	// The old fact returns the rest array; the new one returns its callable element.
	if program.typesByID[id-1].Flags()&checker.TypeFlagsObject == 0 || !checker.IsNonDeferredTypeReference(program.typesByID[oldID-1]) || callback[5] == old[5] {
		t.Fatal("rest element not selected", callback, old)
	}
	for _, question := range []string{"callback-parameters", "callback-parameters\n0", "callback-parameters\n01", "callback-parameters\n999999999", "callback-parameters\n1\nextra"} {
		if _, err := program.Inspect(file, start, end, "Identifier", question); err == nil {
			t.Fatal("malformed question accepted", question)
		}
	}
}

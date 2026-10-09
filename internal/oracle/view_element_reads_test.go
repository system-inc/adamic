package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewElementP01(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "p01", "field read failed: view['text'] is not a string; expected string, found number")
}

func TestCheckedViewElementP05(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "p05", "field read failed: view['count'] is not a number; expected number, found string")
}

func TestCheckedViewElementP06(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "p06", "field read failed: view['count'] is not a number; expected number, found boolean")
}

func TestCheckedViewElementP07(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "p07", "field read failed: view['value'] matches no member of Target; expected Target, found object")
}

func TestCheckedViewLiteralTypedKey(t *testing.T) {
	t.Parallel()
	// The literal-typed key has the same declared slot and checked read boundary.
	checkedViewElementFixture(t, "literal_key", "field read failed: view[key] is not a number; expected number, found string")
}

func TestCheckedViewDestructuredMember(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "destructured", "field read failed: count (field count) is not a number; expected number, found string")
}

func TestCheckedViewDestructuredUnion(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "destructured_union", "field read failed: value (field value) matches no member of Target; expected Target, found object")
}

func checkedViewElementFixture(t *testing.T, name, diagnostic string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/view_element_reads", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: " + diagnostic + "\n")}
	// P0 rules a misfit view to stop at its read, while source Node erases the view.
	if err := reviewAgreement(t, path, &want); err != nil {
		t.Fatal(err)
	}
}

func TestCheckedViewSpread(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "spread", "field read failed: ...view (field count) is not a number; expected number, found string")
}

func TestCheckedViewIn(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "in", "field read failed: 'count' in view (field count) is not a number; expected number, found string")
}

func TestCheckedViewKeys(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "keys", "field read failed: Object.keys(view) (field count) is not a number; expected number, found string")
}

func TestCheckedViewKeysAlias(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "keys_alias", "field read failed: keys(view) (field count) is not a number; expected number, found string")
}

func TestCheckedViewValues(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "values", "field read failed: Object.values(view) (field count) expected 7, found number 9")
}

func TestCheckedViewEntries(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "entries", "field read failed: Object.entries(view) (field count) expected 7, found number 9")
}

func TestCheckedViewElementP19(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "p19", "field read failed: view['value'] expected Target, found function with incompatible parameter representations")
}
func TestCheckedViewElementP72(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "p72", "field read failed: viewed['value'] matches no member of string | number; expected string | number, found boolean")
}

func TestCheckedViewCorrectSpread(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "spread", "count: 'seven'", "count: 7", false)
}

func TestCheckedViewCorrectIn(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "in", "count: 'seven'", "count: 7", false)
}

func TestCheckedViewCorrectKeys(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "keys", "count: 'seven'", "count: 7", false)
}

func TestCheckedViewCorrectKeysAlias(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "keys_alias", "count: 'seven'", "count: 7", false)
}

func TestCheckedViewCorrectValues(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "values", "count: 9", "count: 7", false)
}

func TestCheckedViewCorrectEntries(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "entries", "count: 9", "count: 7", false)
}

func TestCheckedViewCorrectP19NumberRead(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "p19", "(value: string): number => value.length", "(value: number): number => value + 1", true)
}

func TestCheckedViewCorrectP19StringRead(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "p19", "(value: string): number => value.length", "(value: number): string => `value${value}`", true)
}

func TestCheckedViewCorrectP72String(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "p72", "value: true", "value: 'hello'", false)
}

func TestCheckedViewCorrectP72Number(t *testing.T) {
	t.Parallel()
	checkedViewCorrectFixture(t, "p72", "value: true", "value: 5", false)
}

// Successful reads must preserve the source answer on both compiler backends.
// p19's invocation ABI is a separate recorded gap; these controls observe the
// checked callable identity without invoking a synthesized union signature.
func checkedViewCorrectFixture(t *testing.T, name, before, after string, readOnly bool) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "internal/lower/testdata/view_element_reads", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	control := strings.Replace(string(source), before, after, 1)
	if control == string(source) {
		t.Fatal("control replacement did not change its input")
	}
	if readOnly {
		control = strings.Replace(control, "console.log(`${callable(5)}`);", "console.log(`${typeof callable}`);", 1)
	}
	path := filepath.Join(t.TempDir(), "control.a")
	if err := os.WriteFile(path, []byte(control), 0644); err != nil {
		t.Fatal(err)
	}
	if err := reviewAgreement(t, path, nil); err != nil {
		t.Fatal(err)
	}
}

func TestCheckedViewElementP19Read(t *testing.T) {
	t.Parallel()
	checkedViewElementFixture(t, "p19_read", "field read failed: view['value'] expected Target, found function with incompatible parameter representations")
}

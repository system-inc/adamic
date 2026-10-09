package oracle

import (
	"path/filepath"
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

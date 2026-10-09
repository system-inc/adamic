package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The function is lowered before the assertion registers View's certificates.
// Removing an apparently unchecked helper during syntax lowering would miscompile.
func forwardSpreadFixture(t *testing.T, valid bool) {
	t.Helper()
	source := `interface Root { readonly kind: 1; }
interface View extends Root { readonly count: number; }
function copy(value: View): View { return {...value}; }
const raw = { kind: 1 as const, count: 'seven' };
const base: Root = raw;
const view = base as View;
console.log('count=' + copy(view).count);
`
	var want *run
	if valid {
		source = strings.Replace(source, "count: 'seven'", "count: 7", 1)
	} else {
		want = &run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: ...value (field count) is not a number; expected number, found string\n")}
	}
	path := filepath.Join(t.TempDir(), "forward-spread.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if err := reviewAgreement(t, path, want); err != nil {
		t.Fatal(err)
	}
}

func TestCheckedViewSpreadForwardCertificate(t *testing.T) {
	t.Parallel()
	forwardSpreadFixture(t, false)
}

func TestCheckedViewSpreadForwardCorrect(t *testing.T) {
	t.Parallel()
	forwardSpreadFixture(t, true)
}

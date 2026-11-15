package native_test

import (
	"bytes"
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"testing"
)

// The golden is main 6a9ed29661b74858c086e4cd3a37ff019066299e's emitter output.
// Dropping optionalFunctionPresence's program gate reds this byte comparison.
func TestRegExpUnservedEmissionMatchesMain(t *testing.T) {
	t.Parallel()
	checked, err := load.Load([]string{"../oracle/testdata/regexp_served_values/no_relation_presence.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	if program.RegExpServedView {
		t.Fatal("no relation acquired the served-program flag")
	}
	want, err := os.ReadFile("testdata/regexp-structural/no-relation.golden")
	if err != nil {
		t.Fatal(err)
	}
	got := []byte(native.C(program))
	if !bytes.Equal(got, want) {
		path := filepath.Join(t.TempDir(), "actual.c")
		if err := os.WriteFile(path, got, 0600); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("emitted bytes differ from main: got %d, want %d; actual %s", len(got), len(want), path)
	}
	t.Logf("byte-identical to main: %d bytes", len(got))
}

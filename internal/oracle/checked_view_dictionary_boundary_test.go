package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{
		"stage3/interface-downcasts/dictionaries/source/map-primitive-rejected.a", true, true,
	})
}

// No dictionary read follows the cast, so only the inserted boundary check can
// catch the primitive. Removing it must compile and lose that refusal.
func TestCheckedViewDictionaryPrimitiveBoundary(t *testing.T) {
	program, path := interfaceFixture(t, "dictionaries/source/map-primitive-rejected")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "number\n" {
		t.Fatalf("Node: %#v", truth)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: a union value does not match its narrowed type\n")}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("boundary: %s; %#v", diff, got)
		}
	}
	c := native.C(program)
	if calls := strings.Count(c, "adamic_union_narrow("); calls != 1 {
		t.Fatalf("inserted boundary checks=%d, want 1", calls)
	}
	c = "#include \"adamic.h\"\nstatic adamic_heap *dictionary_narrow_mutant(adamic_heap *value, unsigned char wanted) { (void)wanted; return value; }\n" + strings.ReplaceAll(c, "adamic_union_narrow(", "dictionary_narrow_mutant(")
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(c, binary, native.Options{}); err != nil {
		t.Fatal("semantic mutant must compile: ", err)
	}
	code := javascript.JavaScript(program)
	if !strings.Contains(code, "adamicNarrow(") {
		t.Fatal("missing JavaScript boundary check")
	}
	code = "const dictionaryNarrowMutant = (value, wanted) => value;\n" + strings.ReplaceAll(code, "adamicNarrow(", "dictionaryNarrowMutant(")
	js := filepath.Join(t.TempDir(), "mutant.mjs")
	if err := os.WriteFile(js, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{execute(t, binary), onNode(t, js)} {
		if got.exitCode != 0 || len(got.stdout) == 0 || len(got.stderr) != 0 {
			t.Fatalf("boundary mutant did not lose refusal: %#v", got)
		}
		t.Logf("removed boundary check caught: mutant exit=%d stdout=%q", got.exitCode, got.stdout)
	}
}

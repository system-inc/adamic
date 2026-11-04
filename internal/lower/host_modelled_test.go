package lower

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tsc's two throwing requires, each in the form sys.ts writes it, lower whole: 1597's cast of
// require("source-map-support") before install, which checks nothing and emits no check, and 1619's
// computed require returned as a ModuleImportResult's module, where no member typed any reaches the
// return. Each holds its throw once, as head 2's fixtures do. Mutant: hostModelledNever answering
// false, under which 1597 stops at the cast and 1619 at the return (host_modelled.mutant.md).
func TestHostModelledNeverCrossesTheRules(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, holds string }{
		{"source-map-support-cast", `Operation:"host_require_missing"`},
		{"computed-return", `Name:"require_javascript"`},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata/commonjs_reasons", one.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSourceAs(t, string(source), ".ts")
			if err != nil {
				t.Fatal(err)
			}
			dump := fmt.Sprintf("%#v", program)
			if held := strings.Count(dump, one.holds); held != 1 {
				t.Fatalf("want %s once, got %d", one.holds, held)
			}
			if strings.Contains(dump, "CheckedCast") || strings.Contains(dump, `Name:"checked_class_cast"`) {
				t.Fatal("a cast of a value modelled never emits a check")
			}
		})
	}
}

// The rules read the audit's answer for the expression, never require's name or type, so a counted
// require keeps its any: a cast of require("fs") still stops at the cast, and a member typed any beside
// a modelled-never one still stops the return, at that member (Language's condition on #3sh9398).
func TestHostModelledNeverLaundersNoAny(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, want string }{
		{"fs-cast", "a cast the runtime can't check"},
		{"computed-return-any-sibling", "a member typed any (name) related to string at the return site"},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata/commonjs_reasons", one.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSourceAs(t, string(source), ".ts")
			var refused *Refused
			var pending *NotYet
			if !errors.As(err, &refused) && !errors.As(err, &pending) || !strings.Contains(err.Error(), one.want) {
				t.Fatalf("want a stop with %q, got %v", one.want, err)
			}
		})
	}
}

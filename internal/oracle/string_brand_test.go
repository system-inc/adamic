package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"escape", "unescape", "map", "optional", "plain", "value"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/string_brand_" + name + ".a", true, false})
	}
}

func stringBrandOracle(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_brand_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	oracle := onNode(t, path)
	if oracle.exitCode != 0 {
		t.Fatalf("source Node: %#v", oracle)
	}
	sanitized, binary := natively(t, program)
	for _, got := range []run{sanitized, released(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(oracle, got); difference != "" {
			t.Fatalf("%s; Node %#v; backend %#v", difference, oracle, got)
		}
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
}

func TestStringBrandEscape(t *testing.T)   { t.Parallel(); stringBrandOracle(t, "escape") }
func TestStringBrandUnescape(t *testing.T) { t.Parallel(); stringBrandOracle(t, "unescape") }
func TestStringBrandMap(t *testing.T)      { t.Parallel(); stringBrandOracle(t, "map") }
func TestStringBrandOptional(t *testing.T) { t.Parallel(); stringBrandOracle(t, "optional") }
func TestStringBrandPlain(t *testing.T)    { t.Parallel(); stringBrandOracle(t, "plain") }

func TestStringBrandReadRefused(t *testing.T) { t.Parallel(); stringBrandRefused(t, "read", "__brand") }
func TestStringBrandIndexRefused(t *testing.T) {
	t.Parallel()
	stringBrandRefused(t, "index", "__brand")
}
func TestStringBrandRealRefused(t *testing.T) { t.Parallel(); stringBrandRefused(t, "real", "actual") }

func stringBrandRefused(t *testing.T, name, member string) {
	t.Helper()
	_, err := lowered(t, filepath.Join(repository, "internal/oracle/testdata/string_brand_refused", name+".a"))
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.What, member) {
		t.Fatalf("want named refusal for %s, got %v", member, err)
	}
}

func TestStringBrandValue(t *testing.T) { t.Parallel(); stringBrandOracle(t, "value") }

func TestStringBrandDestructureRefused(t *testing.T) {
	t.Parallel()
	stringBrandRefused(t, "destructure", "__brand")
}

func TestStringBrandUnknownCastRefused(t *testing.T) {
	t.Parallel()
	_, err := lowered(t, filepath.Join(repository, "internal/oracle/testdata/string_brand_refused/unknown.a"))
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "cast") {
		t.Fatalf("want unknown cast refused, got %v", err)
	}
}

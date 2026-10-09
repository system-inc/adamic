package tailwind

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// surfacesOf is the compiled surfaces a rule reads from its decoded options, as its Run reads them.
func surfacesOf(t *testing.T, decoded any) *ClassLiteralSurfaces {
	t.Helper()
	carrier, carries := decoded.(interface{ ClassLiteralSurfaces() *ClassLiteralSurfaces })
	if !carries {
		t.Fatalf("%T carries no class literal surfaces", decoded)
	}
	return carrier.ClassLiteralSurfaces()
}

// TestClassLiteralSurfacesAreKeyedByTheMergedSettings pins that every decoder compiles its rule's
// surfaces, and that the compiled surfaces belong to the settings' value rather than to whoever asked:
// twelve rules, written in settings or in their own options, share one entry when they read the same
// surfaces, and never share one when they do not.
func TestClassLiteralSurfacesAreKeyedByTheMergedSettings(t *testing.T) {
	t.Parallel()
	handBuilt := ClassLiteralSurfacesFor(TailwindClassLiteralOptions{Attributes: []string{"tw"}}.ClassLiteralSettings())
	if handBuilt == DefaultClassLiteralSurfaces() {
		t.Fatal("settings reading the tw attribute share the defaults' surfaces")
	}

	for ruleName, decode := range tailwindDecoders(t) {
		for _, testCase := range []struct {
			name string
			raw  string
			base rule.OptionsBase
			want *ClassLiteralSurfaces
		}{
			{"a bare severity", "", rule.OptionsBase{ConfigDirectory: "/repo"}, DefaultClassLiteralSurfaces()},
			{"its own options", `{"attributes":["tw"]}`, rule.OptionsBase{ConfigDirectory: "/repo"}, handBuilt},
			{"settings", "", settingsBase(`{"attributes":["tw"]}`), handBuilt},
			{"options over settings", `{"attributes":["tw"]}`, settingsBase(`{"attributes":["other"]}`), handBuilt},
		} {
			var raw []byte
			if testCase.raw != "" {
				raw = []byte(testCase.raw)
			}
			decoded, err := decode(raw, testCase.base)
			if err != nil {
				t.Fatalf("%s, %s: %v", ruleName, testCase.name, err)
			}
			if surfaces := surfacesOf(t, decoded); surfaces != testCase.want {
				t.Errorf("%s, %s: decoded to surfaces of their own rather than the ones their settings "+
					"share, so each selection compiles and memoizes apart", ruleName, testCase.name)
			}
		}
	}
}

// Not parallel: testing.AllocsPerRun refuses to run in a parallel test, because other tests'
// allocations would be counted as this one's.
//
// TestClassLiteralSurfacesAreReadWithoutAllocating pins that a rule asking for its surfaces on every
// file reads what its decoder compiled, configured or not. Building them on every ask cost an ahra
// cold run 41 MB in 456K objects (#y2nj5ex).
func TestClassLiteralSurfacesAreReadWithoutAllocating(t *testing.T) {
	decoded, err := tailwindDecoders(t)[NoDuplicateClasses.Name]([]byte(`{"attributes":["tw"]}`), settingsBase(`{"callees":["merge"]}`))
	if err != nil {
		t.Fatal(err)
	}
	var options any = decoded
	allocations := testing.AllocsPerRun(100, func() {
		surfaces := DefaultClassLiteralSurfaces()
		if configured, isConfigured := rule.OptionsAs[NoDuplicateClassesOptions](options); isConfigured {
			surfaces = configured.ClassLiteralSurfaces()
		}
		_ = surfaces
	})
	if allocations != 0 {
		t.Errorf("reading a rule's class literal surfaces allocated %.0f times per file, want 0", allocations)
	}
}

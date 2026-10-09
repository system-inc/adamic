package tailwind

import (
	"path/filepath"
	"testing"

	tailwindengine "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
)

// The public twins of the engine assertions that needed a shape only a repository's own stylesheet had
// (#f598zk0, tier B). Each live original stays where it is, opt-in through its corpus; the twin asks the
// same question of collapse/testdata/public_theme, which declares the shape on purpose, so it runs on
// every machine and fails where the original would have skipped.

// publicThemeSystem is the public theme's design system, over the vendored tailwindcss.
func publicThemeSystem(t *testing.T) DesignSystemResult {
	t.Helper()
	entryPoint, err := filepath.Abs(filepath.Join("collapse", "testdata", "public_theme", "theme.css"))
	if err != nil {
		t.Fatalf("resolving the public theme: %v", err)
	}
	system, err := tailwindengine.LoadDesignSystem(tailwindengine.LoadOptions{
		EntryPoint:          entryPoint,
		TailwindPackageRoot: vendored.TailwindPackageRoot(),
	})
	if err != nil {
		t.Fatalf("loading the public theme: %v", err)
	}
	return DesignSystemResult{System: system, Table: tailwindengine.NewTable(system), EntryPoint: entryPoint}
}

// TestClassOrderReadsTheRepositoryOnThePublicTheme is TestClassOrderLiveReadsTheRepositoryRatherThanATable's
// twin: a static `@utility` only the stylesheet declares gets a key, from the design system and not a table.
func TestClassOrderReadsTheRepositoryOnThePublicTheme(t *testing.T) {
	t.Parallel()
	designSystem := publicThemeSystem(t)
	const ownClass = "surface-card"
	if !designSystem.System.HasUtility(ownClass, tailwindengine.UtilityKindStatic) {
		t.Fatalf("the public theme declares `@utility %s` and the design system does not know it", ownClass)
	}
	keys, unplaceable, resolved := classOrderKeys([]string{ownClass, "flex"}, designSystem.System, designSystem.Table)
	if !resolved {
		t.Fatalf("could not place %q", unplaceable)
	}
	if _, placed := keys[ownClass]; !placed {
		t.Errorf("%s got no key, so the order came from a table that cannot know a repository's own utilities", ownClass)
	}
}

// TestExistenceIsNotReadFromPropertyTablesOnThePublicTheme is TestExistenceIsNotReadFromPropertyTables'
// twin. `glow-soft` declares only a custom property, so it exists and no property table can say so.
func TestExistenceIsNotReadFromPropertyTablesOnThePublicTheme(t *testing.T) {
	t.Parallel()
	system := publicThemeSystem(t).System
	lostByShortcut := 0
	for _, className := range []string{"container", "from-black/70", "to-transparent", "glow-soft"} {
		if !classExistsIn(className, system) {
			t.Errorf("%s is real and the rule reports it unknown", className)
			continue
		}
		if _, canResolve := resolveClassFactsIn(className, DesignSystemResult{System: system}); !canResolve {
			lostByShortcut++
		}
	}
	if lostByShortcut == 0 {
		t.Fatal("every class resolved through the property tables, so this fixture cannot tell existence from them")
	}
}

// TestExistenceComesFromTheRepositoryOnThePublicThemes is TestExistenceComesFromTheRepositoryRatherThanATable's
// twin, with the public theme in ahra's place: two design systems that declare different `@utility`
// blocks must disagree about them, which a generated table cannot do.
func TestExistenceComesFromTheRepositoryOnThePublicThemes(t *testing.T) {
	t.Parallel()
	public := publicThemeSystem(t).System
	independent := independentLiveSystem(t)
	testCases := []struct {
		className               string
		onPublic, onIndependent bool
	}{
		{"surface-card", true, false},
		{"pane-wide", true, false},
		{"synthetic-static", false, true},
		{"synthetic-fn-small", false, true},
		{"flex", true, true},
		{"px-4", true, true},
	}
	for _, testCase := range testCases {
		if got := classExistsIn(testCase.className, public); got != testCase.onPublic {
			t.Errorf("%s: the public theme says exists=%v, want %v", testCase.className, got, testCase.onPublic)
		}
		if got := classExistsIn(testCase.className, independent); got != testCase.onIndependent {
			t.Errorf("%s: the independent theme says exists=%v, want %v", testCase.className, got, testCase.onIndependent)
		}
	}
}

// TestRepositoryNamesAreNoLongerVouchedForOnThePublicThemes is
// TestRepositoryNamesAreNoLongerVouchedForByTheFramework's twin: names one stylesheet declares are
// known there and unknown on a system that does not declare them.
func TestRepositoryNamesAreNoLongerVouchedForOnThePublicThemes(t *testing.T) {
	t.Parallel()
	public := publicThemeSystem(t).System
	independent := independentLiveSystem(t)
	for _, name := range []string{"surface-card", "glow-soft", "pane", "rail"} {
		if !classExistsIn(name, public) {
			t.Errorf("%s is declared by the public theme and the rule reports it unknown there", name)
		}
		if classExistsIn(name, independent) {
			t.Errorf("%s is known on the independent theme, which never declares it, so something vouches for it "+
				"beyond the repository", name)
		}
	}
}

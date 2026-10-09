package tailwind

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
)

// The public twins of the engine assertions that needed a shape only a repository's own stylesheet had
// (#f598zk0, tier B). The live originals stay opt-in through their corpus; these ask the same question of
// testdata/public_theme, which declares each shape on purpose, so they run everywhere and fail where an
// original would have skipped.

// publicThemeSystem is testdata/public_theme's design system, over the vendored tailwindcss.
func publicThemeSystem(t *testing.T) *LoadedDesignSystem {
	t.Helper()
	entryPoint, err := filepath.Abs(filepath.Join("testdata", "public_theme", "theme.css"))
	if err != nil {
		t.Fatalf("resolving the public theme: %v", err)
	}
	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: vendored.TailwindPackageRoot()})
	if err != nil {
		t.Fatalf("loading the public theme: %v", err)
	}
	return system
}

// TestARootDeclaredBothWaysKeepsBothKindsOnThePublicTheme is TestARootDeclaredBothWaysKeepsBothKinds'
// twin, in both orders: `pane` is declared static first and `rail` functional first.
func TestARootDeclaredBothWaysKeepsBothKindsOnThePublicTheme(t *testing.T) {
	t.Parallel()
	system := publicThemeSystem(t)
	for _, root := range []string{"pane", "rail"} {
		staticForm := system.HasUtility(root, UtilityKindStatic)
		functionalForm := system.HasUtility(root, UtilityKindFunctional)
		if !staticForm || !functionalForm {
			t.Errorf("`@utility %s` and `@utility %s-*` are both declared and the system reports static=%v functional=%v",
				root, root, staticForm, functionalForm)
		}
	}
}

// TestTableCarriesTheThemeOnThePublicTheme is TestLiveTableCarriesTheRepositoryTheme's twin. The public
// theme declares four static roots and two functional ones, so a composition that drops either kind fails.
func TestTableCarriesTheThemeOnThePublicTheme(t *testing.T) {
	t.Parallel()
	system := publicThemeSystem(t)
	table := NewTable(system)
	checkTableCarriesTheTheme(t, "public", system, table)
	for _, root := range []string{"surface-card", "glow-soft", "pane", "rail"} {
		if _, found := table.Statics[root]; !found {
			t.Errorf("static `@utility %s` is missing from the table's statics", root)
		}
	}
	for _, root := range []string{"pane", "rail"} {
		if _, found := table.Descriptors[root]; !found {
			t.Errorf("functional `@utility %s-*` is missing from the table's descriptors", root)
		}
	}
}

package tailwind

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestFindPathRecursiveIsUpstreamsSearch pins the port of async-utils/fs.js: the path against start,
// then the same file name a directory up, stopping at the filesystem root or once the directory
// searched is cwd.
func TestFindPathRecursiveIsUpstreamsSearch(t *testing.T) {
	t.Parallel()
	present := map[string]bool{
		"/repo/styles/theme.css":   true,
		"/repo/theme.css":          true,
		"/outside/brand.css":       true,
		"/repo/packages/web/a.css": true,
	}
	exists := func(path string) bool { return present[path] }
	testCases := []struct {
		name, cwd, entry, want string
	}{
		{"found where written", "/repo", "styles/theme.css", "/repo/styles/theme.css"},
		{"found one directory up, keeping the name", "/repo", "missing/theme.css", "/repo/theme.css"},
		{"not searched above cwd", "/repo/packages", "x/brand.css", ""},
		{"an absolute path stands", "/repo", "/outside/brand.css", "/outside/brand.css"},
		{"nested", "/repo/packages/web", "a.css", "/repo/packages/web/a.css"},
	}
	for _, testCase := range testCases {
		if got := findPathRecursive(testCase.cwd, testCase.cwd, testCase.entry, exists); got != testCase.want {
			t.Errorf("%s: found %q, want %q", testCase.name, got, testCase.want)
		}
	}
}

// runLocationFixture runs no-unknown-classes on a project whose default stylesheet (app/globals.css,
// one of FindEntryPoint's candidates) declares nothing of its own, and which holds a second stylesheet
// declaring `brand-glow` under packages/web/design/brand.css. Which one the rule reads decides whether
// `brand-glow` is known.
func runLocationFixture(t *testing.T, options NoUnknownClassesOptions) rule_testing.Result {
	t.Helper()
	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so there is no design system to read")
	}
	files := map[string]string{
		"Component.tsx":                 `export const element = <div className="brand-glow flex" />;`,
		"app/globals.css":               `@import "tailwindcss";`,
		"packages/web/design/brand.css": "@import \"tailwindcss\";\n@utility brand-glow { color: red; }",
	}
	return rule_testing.RunTypedFilesWithSetupAndOptions(t, NoUnknownClasses, files, "Component.tsx", options, func(root string) {
		modules := filepath.Join(root, "node_modules")
		if err := os.MkdirAll(modules, 0o755); err != nil {
			t.Fatalf("creating node_modules: %v", err)
		}
		if err := os.Symlink(packageRoot, filepath.Join(modules, "tailwindcss")); err != nil {
			t.Fatalf("linking tailwindcss: %v", err)
		}
	})
}

// TestTailwindLocationOptions is upstream's entryPoint, tailwindConfig and cwd, each changing which
// stylesheet the rule reads. With none, the project's own stylesheet is read and brand-glow is unknown.
func TestTailwindLocationOptions(t *testing.T) {
	t.Parallel()
	known := func(name string, options NoUnknownClassesOptions) {
		t.Helper()
		rule_testing.ExpectClean(t, runLocationFixture(t, options))
	}
	unknown := func(name string, options NoUnknownClassesOptions) {
		t.Helper()
		if result := runLocationFixture(t, options); len(result.Diagnostics) != 1 {
			t.Errorf("%s: want brand-glow reported once, got %d findings", name, len(result.Diagnostics))
		}
	}
	location := func(entryPoint, tailwindConfig, cwd string) NoUnknownClassesOptions {
		return NoUnknownClassesOptions{TailwindLocationOptions: TailwindLocationOptions{
			EntryPoint: entryPoint, TailwindConfig: tailwindConfig, Cwd: cwd,
		}}
	}

	unknown("none: the project's own stylesheet", NoUnknownClassesOptions{})
	known("entryPoint", location("packages/web/design/brand.css", "", ""))
	known("tailwindConfig stands in for entryPoint", location("", "packages/web/design/brand.css", ""))
	known("entryPoint wins over tailwindConfig", location("packages/web/design/brand.css", "app/globals.css", ""))
	known("cwd anchors the entry point", location("design/brand.css", "", "packages/web"))
	// Not found from cwd, so upstream's fallback: Tailwind's default theme, where brand-glow is unknown.
	unknown("an entry point not found falls back", location("design/brand.css", "", ""))
}

// TestDesignSystemIsCachedPerLocation asks one program for two locations and requires two stylesheets.
// A cache keyed on the program alone would hand the second rule the first rule's design system.
func TestDesignSystemIsCachedPerLocation(t *testing.T) {
	t.Parallel()
	packageRoot := unknownFixturePackageRoot()
	if packageRoot == "" {
		t.Skip("no installed tailwindcss on this machine, so there is no design system to read")
	}
	var entryPoints []string
	probe := rule.Rule{
		Name:         "probe/design-system-per-location",
		ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDesignSystem,
		NoListener:   rule.NoListenerAnswersInRun,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			for _, location := range []DesignSystemLocation{{}, {ConfigPath: "packages/web/design/brand.css"}, {}} {
				entryPoints = append(entryPoints, filepath.Base(DesignSystemForProgramAt(ctx.Program, location).EntryPoint))
			}
			return nil
		},
	}
	files := map[string]string{
		"Component.tsx":                 `export const element = <div className="flex" />;`,
		"app/globals.css":               `@import "tailwindcss";`,
		"packages/web/design/brand.css": `@import "tailwindcss";`,
	}
	rule_testing.RunTypedFilesWithSetup(t, probe, files, "Component.tsx", func(root string) {
		modules := filepath.Join(root, "node_modules")
		if err := os.MkdirAll(modules, 0o755); err != nil {
			t.Fatalf("creating node_modules: %v", err)
		}
		if err := os.Symlink(packageRoot, filepath.Join(modules, "tailwindcss")); err != nil {
			t.Fatalf("linking tailwindcss: %v", err)
		}
	})
	if want := []string{"globals.css", "brand.css", "globals.css"}; !slices.Equal(entryPoints, want) {
		t.Fatalf("entry points per location %v, want %v", entryPoints, want)
	}
}

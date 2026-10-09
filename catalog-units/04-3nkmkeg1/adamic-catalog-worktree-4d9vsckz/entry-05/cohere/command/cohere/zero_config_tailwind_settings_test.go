package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
)

// tailwindProjectConfig is a tsconfig for a project with TSX in it.
const tailwindProjectConfig = `{
    "compilerOptions": {"target": "ES2022", "module": "esnext", "moduleResolution": "bundler", "strict": true,
        "noEmit": true, "jsx": "preserve"},
    "include": ["**/*.ts", "**/*.tsx"]
}`

// zeroConfigTailwindProject writes a project whose CohereSettings.json, when it has one, names no
// cohere: set, so the run takes the house stack, with tailwindcss linked into its node_modules.
func zeroConfigTailwindProject(t *testing.T, files map[string]string) string {
	t.Helper()
	// The vendored snapshot of tailwindcss, since cohere installs no npm packages, so this runs on any
	// machine (#sycrdr6). It used to be ahra's installed copy, through the ahra corpus.
	packageRoot := vendored.TailwindPackageRoot()
	root := t.TempDir()
	files["tsconfig.json"] = tailwindProjectConfig
	writeTree(t, root, files)
	if err := os.MkdirAll(filepath.Join(root, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(packageRoot, filepath.Join(root, "node_modules", "tailwindcss")); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestZeroConfigReadsTailwindSettingsOrRefusesThem is the case both reviewers of #gj5nm6e asked for: a
// project with no cohere: set in its chain, so zero config decides which sets apply, that writes
// settings["better-tailwindcss"]. The settings reach the Tailwind rules, or the run refuses and says
// why. Dropping them silently is the defect the unit exists to end.
func TestZeroConfigReadsTailwindSettingsOrRefusesThem(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	environment := []string{"COHERE_RUN_CACHE=off"}

	t.Run("settings naming a stylesheet the default probes miss are read", func(t *testing.T) {
		t.Parallel()
		root := zeroConfigTailwindProject(t, map[string]string{
			"CohereSettings.json": `{"settings": {"better-tailwindcss": {"entryPoint": "./design/brand.css", ` +
				`"callees": ["mergeClassNames"]}}}`,
			"design/brand.css": `@import "tailwindcss";`,
			"Component.tsx": "declare function mergeClassNames(...classes: string[]): string;\n" +
				"declare function cn(...classes: string[]): string;\n" +
				"export const ours = mergeClassNames('flex flex');\n" +
				"export const theirs = cn('p-2 p-2');\n",
		})
		output, _ := runCohereWithEnvironment(t, binary, root, environment, "--no-fix", "--no-cache")
		if !strings.Contains(output, "design/brand.css is a Tailwind stylesheet") {
			t.Fatalf("zero config did not apply cohere:tailwind from the stylesheet the settings name:\n%s", output)
		}
		if !strings.Contains(output, `"flex" is listed more than once`) {
			t.Fatalf("the settings' callee was not read, so mergeClassNames' repeat went unreported:\n%s", output)
		}
		if strings.Contains(output, `"p-2" is listed more than once`) {
			t.Fatalf("cn was read though the settings' callees replace the defaults:\n%s", output)
		}
	})

	t.Run("settings the house would not apply are refused by name", func(t *testing.T) {
		t.Parallel()
		root := zeroConfigTailwindProject(t, map[string]string{
			"CohereSettings.json": `{"settings": {"better-tailwindcss": {"callees": ["mergeClassNames"]}}}`,
			"index.ts":            "export const value = 1;\n",
		})
		output, exitCode := runCohereWithEnvironment(t, binary, root, environment, "--no-fix", "--no-cache")
		if exitCode == 0 || !strings.Contains(output, `writes settings["better-tailwindcss"], and zero config does not apply cohere:tailwind`) {
			t.Fatalf("settings nothing would read loaded (exit %d):\n%s", exitCode, output)
		}
	})

	t.Run("an outsider with no settings has cn and clsx read", func(t *testing.T) {
		t.Parallel()
		root := zeroConfigTailwindProject(t, map[string]string{
			"app/globals.css": `@import "tailwindcss";`,
			"Component.tsx": "declare function cn(...classes: string[]): string;\n" +
				"declare function clsx(...classes: string[]): string;\n" +
				"export const a = cn('flex flex');\n" +
				"export const b = clsx('gap-2 gap-2');\n",
		})
		output, _ := runCohereWithEnvironment(t, binary, root, environment, "--no-fix", "--no-cache")
		for _, repeated := range []string{`"flex" is listed more than once`, `"gap-2" is listed more than once`} {
			if !strings.Contains(output, repeated) {
				t.Errorf("an outsider's repeat went unreported (%s):\n%s", repeated, output)
			}
		}
	})
}

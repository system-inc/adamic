package adamic

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * Every adamic rule is tested under Adamic's own compiler options (adamic docs/0.1.md, "Compiler
 * options"), because a hole is a hole only under them: without exactOptionalPropertyTypes or
 * noUncheckedIndexedAccess several of tsc's answers change. The fixtures are the probes on #drbrp8c, each
 * accepted by tsc 6.0.3 and failing on Node 24.14.1 when it fires, and each control running right when it
 * stays clean. Adamic's ten programs, which Adamic 0.1 must compile, are the clean corpus every rule
 * shares (testdata/programs, byte for byte from the tarball on #8vpywxp).
 */

// adamicConfiguration is Adamic 0.1's compiler options, with es2024's library and a console declaration
// standing in for Adamic's own library, as Adamic's verification did.
const adamicConfiguration = `{
	"compilerOptions": {
		"strict": true,
		"noUncheckedIndexedAccess": true,
		"exactOptionalPropertyTypes": true,
		"noImplicitReturns": true,
		"noFallthroughCasesInSwitch": true,
		"erasableSyntaxOnly": true,
		"verbatimModuleSyntax": true,
		"allowImportingTsExtensions": true,
		"noEmit": true,
		"module": "nodenext",
		"target": "es2024",
		"lib": ["es2024"],
		"types": []
	},
	"include": ["**/*.ts"]
}`

// adamicSupport is what Adamic's verification put beside the programs: a console, and the runtime's
// `panic` as the package `adamic`.
var adamicSupport = map[string]string{
	"tsconfig.json":                    adamicConfiguration,
	"package.json":                     `{ "type": "module" }`,
	"console.d.ts":                     `declare const console: { log(message: string): void; error(message: string): void };`,
	"node_modules/adamic/package.json": `{ "name": "adamic", "type": "module", "types": "index.d.ts" }`,
	"node_modules/adamic/index.d.ts":   `export declare function panic(message: string): never;`,
}

// writeAdamicSupport writes adamicSupport into a fixture's directory, over the harness's default tsconfig.
func writeAdamicSupport(t *testing.T) func(directory string) {
	return func(directory string) {
		for name, contents := range adamicSupport {
			path := filepath.Join(directory, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// runAdamic runs one rule over one fixture, Case.ts, under Adamic's options.
func runAdamic(t *testing.T, subject rule.Rule, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFilesWithSetup(t, subject, map[string]string{"Case.ts": source}, "Case.ts",
		writeAdamicSupport(t))
}

// spans is the text each finding covers, in order, from the fixture as the harness wrote it.
func spans(source string, result rule_testing.Result) []string {
	text := rule_testing.FixtureText(source)
	found := make([]string, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		found = append(found, text[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	sort.Strings(found)
	return found
}

// expectSpans asserts exactly the spans the findings cover, so a finding at the wrong node fails as surely
// as a missing one. The ids are asserted beside it, with rule_testing.ExpectFindings, at each call.
func expectSpans(t *testing.T, source string, result rule_testing.Result, want ...string) {
	t.Helper()
	sort.Strings(want)
	if got := spans(source, result); strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("spans %q, want %q", got, want)
	}
}

// repeated is id once per finding, for ExpectFindings.
func repeated(id string, count int) []string {
	ids := make([]string, count)
	for index := range ids {
		ids[index] = id
	}
	return ids
}

// adamicPrograms reads Adamic's ten programs, keyed by their path under testdata/programs.
func adamicPrograms(t *testing.T) map[string]string {
	t.Helper()
	programs := map[string]string{}
	root := filepath.Join("testdata", "programs")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		programs[filepath.ToSlash(relative)] = string(contents)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	// Eleven files: ten programs, the seventh in two modules.
	if len(programs) != 11 {
		t.Fatalf("read %d program files, want 11, so the clean corpus is not the one Adamic must compile", len(programs))
	}
	return programs
}

// expectProgramsClean runs a rule over each of Adamic's programs and expects nothing: each is a program
// Adamic 0.1 must compile, so a finding on one is a rule refusing real Adamic.
func expectProgramsClean(t *testing.T, subject rule.Rule) {
	t.Helper()
	programs := adamicPrograms(t)
	for name := range programs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFilesWithSetup(t, subject, programs, name, writeAdamicSupport(t))
			rule_testing.ExpectClean(t, result)
		})
	}
}

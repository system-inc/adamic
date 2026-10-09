package main

import (
	"maps"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// adamicProject is a project whose tsconfig claims Adamic's `.a` (#6mhafvb): two `.a` files that import each
// other with a type error between them, one formatted badly, one whose `<T,>` the same file named `.ts` drops,
// and a static library outside the include that nothing may read.
func adamicProject(t *testing.T, extra map[string]string) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"tsconfig.json": `{
    "sourceExtensions": [".a"],
    "compilerOptions": {
        "target": "ES2022",
        "module": "esnext",
        "moduleResolution": "bundler",
        "strict": true,
        "noEmit": true,
        "allowImportingTsExtensions": true
    },
    "include": ["src"]
}
`,
		"CohereSettings.json":      `{"extends":"./NexusCohereSettings.json","rules":{}}`,
		"NexusCohereSettings.json": `{"format":{"ignore":["tsconfig.json","*.json"]}}`,
		".gitignore":               ".cache/\n",
		"src/geometry.a":           "export function area(width: number, height: number): number {\n  return width * height;\n}\n",
		"src/main.a":               "import { area } from \"./geometry.a\";\n\nexport const total: string = area(2, 3);\n",
		"src/messy.a":              "export const   messy  =  1;\n",
		"src/generic.a":            "export const identity = <T,>(value: T): T => value;\n",
		"lib/libfoo.a":             "!<arch>\nnot TypeScript at all {{{\n",
	}
	maps.Copy(files, extra)
	writeTree(t, root, files)
	return root
}

// adamicAliasPattern is any name an `.a` file could be printed under that is not its own: X.a.ts, X.a.js and
// the like.
var adamicAliasPattern = regexp.MustCompile(`[\w/.-]+\.a\.[a-z]+\b`)

// A project that claims `.a` is checked, linted and formatted under the files' own names (#6mhafvb). Every path
// the run prints for an Adamic file ends in `.a`, never in an alias; the type error between two `.a` files is
// found, so the import resolved; the formatter rewrites them exactly as it would the same text named `.ts`; and
// a static library named `.a` outside the program is never read, named or written.
func TestAnAdamicProjectIsCheckedUnderItsOwnNames(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := adamicProject(t, nil)

	output, code := runCohere(t, binary, root, "--no-fix")
	if code == 0 {
		t.Fatalf("the planted type error and formatting passed:\n%s", output)
	}
	for _, want := range []string{"src/main.a:3:14", "TS2322", "src/messy.a", "src/generic.a"} {
		if !strings.Contains(output, want) {
			t.Errorf("the check does not name %q:\n%s", want, output)
		}
	}
	if alias := adamicAliasPattern.FindString(output); alias != "" {
		t.Errorf("an Adamic file was printed as %q, a name that is not its own:\n%s", alias, output)
	}
	if strings.Contains(output, "libfoo") || strings.Contains(output, "geometry.a:") || strings.Contains(output, "main.a:1:1") {
		t.Errorf("a file with nothing to report was named, or the static library was read:\n%s", output)
	}

	if output, _ := runCohere(t, binary, root, "--format-only"); adamicAliasPattern.MatchString(output) {
		t.Errorf("the writing run printed an alias:\n%s", output)
	}
	for name, want := range map[string]string{
		"src/messy.a":   "export const messy = 1;\n",
		"src/generic.a": "export const identity = <T>(value: T): T => value;\n",
		"src/geometry.a": "export function area(width: number, height: number): number {\n" +
			"  return width * height;\n}\n",
		"lib/libfoo.a": "!<arch>\nnot TypeScript at all {{{\n",
	} {
		if got := readForTest(t, filepath.Join(root, name)); got != want {
			t.Errorf("%s after the writing run:\n%s\nwant:\n%s", name, got, want)
		}
	}
	if output, code := runCohere(t, binary, root, "--no-fix", "--format-only"); code != 0 {
		t.Errorf("the formatted tree does not pass its own format check, exit %d:\n%s", code, output)
	}
}

// An `.a` file the program holds and git ignores is checked and never formatted, and the run says so on the
// footer and by file, with the ignore line (#6mhafvb). C and C++ ignore templates list `*.a`, and a silent green
// over Adamic source nobody formatted is the failure this exists for. The same tree without the line formats
// them, the control that the walk and not the formatter is what left them out.
func TestAnAdamicFileGitIgnoresIsANamedGap(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := adamicProject(t, map[string]string{".gitignore": ".cache/\n*.a\n"})

	output, _ := runCohere(t, binary, root, "--no-fix")
	if !strings.Contains(output, "src/messy.a is in the program, but git's ignore rules leave it out of the format walk (.gitignore line 2, `*.a`): checked, not formatted") {
		t.Errorf("the ignored Adamic file is not named with its ignore line:\n%s", output)
	}
	if !strings.Contains(output, "⚠ 4 .a files git ignores, not formatted") {
		t.Errorf("the footer does not count the gap:\n%s", output)
	}
	if !strings.Contains(output, "TS2322") {
		t.Errorf("the ignored files were not type-checked:\n%s", output)
	}
	if output, _ := runCohere(t, binary, root, "--format-only"); readForTest(t, filepath.Join(root, "src/messy.a")) != "export const   messy  =  1;\n" {
		t.Errorf("a file git ignores was formatted:\n%s", output)
	}

	writeTree(t, root, map[string]string{".gitignore": ".cache/\n"})
	if output, _ := runCohere(t, binary, root, "--no-fix"); strings.Contains(output, "git ignores") {
		t.Errorf("with the ignore line gone, the gap is still reported:\n%s", output)
	}
}

// claimAdamic takes into a scope exactly the `.a` files the program holds and the walk admitted (#6mhafvb): all
// of them into a scope drawn from the walk, less those on record for the default scope, and in a named scope it
// keeps a named `.a` only when both say so, so a named libfoo.a never reaches the printer or the parse guard.
func TestTheScopeClaimsTheAdamicFilesTheProgramHolds(t *testing.T) {
	t.Parallel()
	admitted := []string{"/repo/geometry.a", "/repo/messy.a", "/repo/lib/libfoo.a"}
	held := map[string]struct{}{"/repo/geometry.a": {}, "/repo/messy.a": {}, "/repo/ignored.a": {}}
	scopeOf := func(scope formatScope) string {
		names := []string{}
		for _, fileName := range scope.FileNames {
			if !scope.includes(fileName) {
				t.Errorf("%s is listed and not indexed", fileName)
			}
			names = append(names, fileName)
		}
		return strings.Join(names, " ")
	}

	walked := formatScope{FileNames: []string{"/repo/a.ts"}, index: map[string]struct{}{"/repo/a.ts": {}}, adamic: admitted, adamicJoins: true}
	claimedScope, claimed := walked.claimAdamic(held)
	if got := scopeOf(claimedScope); got != "/repo/a.ts /repo/geometry.a /repo/messy.a" {
		t.Errorf("a scope drawn from the walk became %q", got)
	}
	if strings.Join(claimed, " ") != "/repo/geometry.a /repo/messy.a" {
		t.Errorf("claimed %v, want the held files the walk admitted", claimed)
	}

	walked.adamicUnrecorded = func(files []string) []string { return files[1:] }
	if got := scopeOf(func() formatScope { scope, _ := walked.claimAdamic(held); return scope }()); got != "/repo/a.ts /repo/messy.a" {
		t.Errorf("the default scope took a file on record: %q", got)
	}

	named := formatScope{
		FileNames: []string{"/repo/a.ts", "/repo/ignored.a", "/repo/lib/libfoo.a", "/repo/messy.a"},
		index:     map[string]struct{}{"/repo/a.ts": {}, "/repo/ignored.a": {}, "/repo/lib/libfoo.a": {}, "/repo/messy.a": {}},
		adamic:    admitted,
	}
	namedScope, _ := named.claimAdamic(held)
	if got := scopeOf(namedScope); got != "/repo/a.ts /repo/messy.a" {
		t.Errorf("a named scope became %q: an .a the walk left out or the program does not hold stays out", got)
	}
	if namedScope.includes("/repo/lib/libfoo.a") || namedScope.includes("/repo/ignored.a") {
		t.Error("an unclaimed .a is still in the named scope's index")
	}
}

// An editor's save formats an Adamic `.a` buffer the program holds as the gate would, and hands back a `.a`
// buffer the program does not hold as typed: whether a `.a` file is source is the program's to say, on a save
// as on a run (#6mhafvb).
func TestAnEditorSaveFormatsOnlyTheAdamicFilesTheProgramHolds(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := adamicProject(t, map[string]string{"scratch/notes.a": "export const   notes  =  1;\n"})

	stdout, stderr, code := runCohereWithStdin(t, binary, root, "export const   messy  =  1;\n", "--fix", "--format", "--stdin-filepath", "src/messy.a")
	if code != 0 || stdout != "export const messy = 1;\n" {
		t.Errorf("the save of a held .a answered %q, exit %d:\n%s", stdout, code, stderr)
	}
	stdout, stderr, code = runCohereWithStdin(t, binary, root, "export const   notes  =  1;\n", "--fix", "--format", "--stdin-filepath", "scratch/notes.a")
	if code != 0 || stdout != "export const   notes  =  1;\n" {
		t.Errorf("the save of a .a outside the program answered %q, exit %d:\n%s", stdout, code, stderr)
	}
}

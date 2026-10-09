package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

/*
 * What a run checks and what it may write are two different sets.
 *
 * A named path is checked together with everything that imports it, because a change to an exported
 * type breaks its consumers and a check that stopped at the named file would miss those findings.
 * Above the closure limit the whole tree is checked. Both widenings are about seeing more, and both
 * are honest when the coverage line says so.
 *
 * Writing is not seeing. The fix phase used to rewrite every file the checked set handed it, so
 * `cohere --fix app` in www-phi-health reached past 500 files, fell back to the whole tree, and
 * rewrote 15 files inside libraries/structure and 2 inside its nested nexus, none of them named.
 * The caller said which files this run is about; widening the check must never widen that.
 *
 * Every case puts the same fixable violation in a file that was named and in one that was not, so
 * the named file changing is the control that proves the fixer ran, and the other staying
 * byte-identical is the boundary.
 */

const fixScopeTsconfig = `{
    "compilerOptions": {
        "target": "ES2022",
        "module": "esnext",
        "moduleResolution": "bundler",
        "strict": true,
        "noEmit": true
    },
    "include": ["**/*.ts"]
}`

// fixScopeProject writes a project where the named file, a file that imports it, and an unrelated
// sibling all carry the same `debugger` statement, which `no-debugger` repairs.
func fixScopeProject(t *testing.T, root string, extra map[string]string) {
	t.Helper()
	files := map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      `{"extends":"./NexusCohereSettings.json","rules":{"no-debugger":"error"}}`,
		"NexusCohereSettings.json": `{"format":{"ignore":[]}}`,
		"Producer.ts":              "export function value(): number {\n    debugger;\n    return 1;\n}\n",
		"Consumer.ts":              "import { value } from './Producer';\n\nexport function doubled(): number {\n    debugger;\n    return value() * 2;\n}\n",
		"Sibling.ts":               "export function alone(): number {\n    debugger;\n    return 3;\n}\n",
	}
	for name, contents := range extra {
		files[name] = contents
	}
	writeTree(t, root, files)
}

func readForTest(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

// assertWroteOnlyNamed checks the control and the boundary together.
func assertWroteOnlyNamed(t *testing.T, root string, before map[string]string, written string, output string) {
	t.Helper()
	if after := readForTest(t, filepath.Join(root, written)); after == before[written] || strings.Contains(after, "debugger") {
		t.Fatalf("the control failed: %s was named and still carries its violation, so the fixer never ran:\n%s\n%s",
			written, after, output)
	}
	for name, contents := range before {
		if name == written {
			continue
		}
		if after := readForTest(t, filepath.Join(root, name)); after != contents {
			t.Errorf("%s was not named and was rewritten:\nbefore:\n%s\nafter:\n%s\noutput:\n%s", name, contents, after, output)
		}
	}
}

func TestAFixRunWritesOnlyWhatWasNamed(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)

	// Within the closure limit: the consumer is checked because it imports the named file, and it
	// is not written because nobody named it.
	t.Run("a dependent is checked and not written", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		fixScopeProject(t, root, nil)
		before := map[string]string{
			"Producer.ts": readForTest(t, filepath.Join(root, "Producer.ts")),
			"Consumer.ts": readForTest(t, filepath.Join(root, "Consumer.ts")),
			"Sibling.ts":  readForTest(t, filepath.Join(root, "Sibling.ts")),
		}

		output, code := runCohere(t, binary, root, "--fix", "Producer.ts")
		if code != 0 {
			t.Fatalf("exit %d:\n%s", code, output)
		}
		assertWroteOnlyNamed(t, root, before, "Producer.ts", output)
		// The withheld repair is named rather than dropped, so a reader knows a fixable finding sits
		// outside what they asked about.
		if !strings.Contains(output, "Consumer.ts") || !strings.Contains(output, "not rewritten") {
			t.Errorf("the withheld repair was not reported:\n%s", output)
		}
	})

	// Past the closure limit the whole tree is checked, which used to make the whole tree writable.
	t.Run("the whole-tree fallback checks everything and writes only the named file", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		consumers := map[string]string{}
		for index := 0; index <= program.DependentClosureLimit; index++ {
			consumers[filepath.Join("consumers", "Consumer"+itoaForTest(index)+".ts")] =
				"import { value } from '../Producer';\nexport const used" + itoaForTest(index) + " = value();\n"
		}
		fixScopeProject(t, root, consumers)
		before := map[string]string{
			"Producer.ts": readForTest(t, filepath.Join(root, "Producer.ts")),
			"Consumer.ts": readForTest(t, filepath.Join(root, "Consumer.ts")),
			"Sibling.ts":  readForTest(t, filepath.Join(root, "Sibling.ts")),
		}

		output, code := runCohere(t, binary, root, "--fix", "Producer.ts")
		if code != 0 {
			t.Fatalf("exit %d:\n%s", code, output)
		}
		// The premise: this run really did take the fallback. Without it the case would pass on a
		// closure that stayed inside the limit and prove nothing about the fallback.
		if !strings.Contains(output, "so the whole tree is checked") {
			t.Fatalf("the fixture did not reach the whole-tree fallback:\n%s", output)
		}
		assertWroteOnlyNamed(t, root, before, "Producer.ts", output)
	})

	// The control on the boundary: with nothing named, the whole project is the caller's, and every
	// file with a repair is written as it always was.
	t.Run("with nothing named, every repair lands", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		fixScopeProject(t, root, nil)

		output, code := runCohere(t, binary, root, "--fix")
		if code != 0 {
			t.Fatalf("exit %d:\n%s", code, output)
		}
		for _, name := range []string{"Producer.ts", "Consumer.ts", "Sibling.ts"} {
			if strings.Contains(readForTest(t, filepath.Join(root, name)), "debugger") {
				t.Errorf("%s kept its violation on a run that named nothing:\n%s", name, output)
			}
		}
	})
}

// A whole-tree fix run does not write into a repository of its own below the root, the boundary the
// format walk already kept (#sqc145r). `cohere --fix --format --format-all` in api-phi-health applied
// six return-await fixes inside Base, a nested repository nobody named, and they had to be reverted.
//
// Each case carries the project's own violations beside the nested one, so the project's files being
// fixed is the control that proves the fixer ran, and the nested file staying byte-identical is the
// boundary. Both shapes a nested repository takes are covered: a clone, whose `.git` is a directory,
// and a submodule, whose `.git` is a gitlink file.
func TestAFixRunDoesNotWriteIntoANestedRepository(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	nestedViolation := "export function inside(): number {\n    debugger;\n    return 4;\n}\n"

	shapes := []struct {
		name  string
		plant func(t *testing.T, nested string)
	}{
		{"a clone", plantClone},
		{"a submodule's gitlink", func(t *testing.T, nested string) {
			writeTree(t, nested, map[string]string{".git": "gitdir: ../.git/modules/nested\n"})
		}},
	}
	for _, shape := range shapes {
		t.Run(shape.name+": a whole-tree run leaves it alone and says so", func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			fixScopeProject(t, root, map[string]string{"nested/Inside.ts": nestedViolation})
			shape.plant(t, filepath.Join(root, "nested"))

			output, code := runCohere(t, binary, root, "--fix")
			if code != 0 {
				t.Fatalf("exit %d:\n%s", code, output)
			}
			for _, name := range []string{"Producer.ts", "Consumer.ts", "Sibling.ts"} {
				if strings.Contains(readForTest(t, filepath.Join(root, name)), "debugger") {
					t.Fatalf("the control failed: the project's own %s kept its violation, so the fixer never ran:\n%s", name, output)
				}
			}
			if after := readForTest(t, filepath.Join(root, "nested", "Inside.ts")); after != nestedViolation {
				t.Errorf("a file in a nested repository nobody named was rewritten:\n%s\noutput:\n%s", after, output)
			}
			if !strings.Contains(output, "1 fix not applied: in nested repository nested") {
				t.Errorf("the withheld fix was not reported:\n%s", output)
			}
		})
	}

	t.Run("naming a path inside it writes it", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		fixScopeProject(t, root, map[string]string{"nested/Inside.ts": nestedViolation})
		plantClone(t, filepath.Join(root, "nested"))

		output, code := runCohere(t, binary, root, "--fix", "nested/Inside.ts")
		if code != 0 {
			t.Fatalf("exit %d:\n%s", code, output)
		}
		if strings.Contains(readForTest(t, filepath.Join(root, "nested", "Inside.ts")), "debugger") {
			t.Errorf("a named file inside a nested repository was not written:\n%s", output)
		}
		if strings.Contains(output, "not applied: in nested repository") {
			t.Errorf("a named file was reported as withheld:\n%s", output)
		}
	})
}

// plantClone makes a directory look like a clone the way cohere looks: a `.git` directory, which is all
// formatfiles.HasOwnRepository asks. No git runs, so the fixture holds where git is absent.
func plantClone(t *testing.T, directory string) {
	t.Helper()
	writeTree(t, directory, map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
}

// A format candidate the walk never linted is formatted and never fixed (#xn1k1gz). angular/angular keeps a
// 2.8 MB esbuild bundle beside a tsconfig that does not include it, so no rule runs over it and no finding is
// reported for it. The format pass printed it, and every later fix pass re-linted the printed text with every
// rule and applied what they proposed: a vendored file rewritten by repairs nobody was told about, and 26
// minutes of re-linting on one core.
//
// The project's own files losing their violation is the control that proves the fixer ran; the bundle being
// reformatted proves the format pass reached it; its `debugger` surviving is the boundary.
func TestAFileTheWalkNeverLintedIsFormattedAndNeverFixed(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	fixScopeProject(t, root, map[string]string{"bundle.js": "var first=1\nfunction run(){debugger;return first}\n"})

	output, code := runCohere(t, binary, root, "--fix")
	if code != 0 {
		t.Fatalf("exit %d:\n%s", code, output)
	}
	for _, name := range []string{"Producer.ts", "Consumer.ts", "Sibling.ts"} {
		if strings.Contains(readForTest(t, filepath.Join(root, name)), "debugger") {
			t.Fatalf("the control failed: the project's own %s kept its violation, so the fixer never ran:\n%s", name, output)
		}
	}
	bundle := readForTest(t, filepath.Join(root, "bundle.js"))
	if !strings.Contains(bundle, "var first = 1;") {
		t.Fatalf("the bundle was not formatted, so the format pass never reached it and the boundary is untested:\n%s\noutput:\n%s", bundle, output)
	}
	if !strings.Contains(bundle, "debugger") {
		t.Errorf("a file no rule was run over had a rule's fix applied to it:\n%s\noutput:\n%s", bundle, output)
	}
}

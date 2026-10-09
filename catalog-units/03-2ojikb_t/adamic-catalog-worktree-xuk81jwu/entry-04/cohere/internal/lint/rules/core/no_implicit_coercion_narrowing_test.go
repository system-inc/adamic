package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/internal/types/program"
)

/*
 * The `!!x` fix must not break TypeScript's narrowing.
 *
 * TypeScript narrows a reference through `!!x` and through an explicit comparison, including when the
 * result is aliased into a const, and it does not narrow through a call. So rewriting `!!x` to
 * `Boolean(x)` changes what the program means to the compiler while preserving what it does at run
 * time:
 *
 *	const hasData = !!data;        hasData ? data.energy : 0    compiles
 *	const hasData = Boolean(data); hasData ? data.energy : 0    TS18048, data is possibly undefined
 *
 * That shape is www-phi-health `app/(app-layout)/_layout/sidebar/AppSidebarEnergyBalance.tsx`, whose
 * hand repair was `energyBalanceRequestData !== undefined`, and it broke seven ahra sites the same
 * morning. So the fixer emits the comparison the operand's type admits when the operand is a
 * reference whose narrowing a later line could depend on, keeps `Boolean(x)` where narrowing cannot
 * change anything, and declines (still reporting) when no comparison means the same thing.
 *
 * Every fixed shape is type-checked after the fix, and the phi shape is also type-checked with the
 * old `Boolean` output, which must fail: that control is what proves the check can see the defect.
 */

const narrowingDeclarations = "declare const request: { data: { energy: { credits: number } } | undefined };\n" +
	"declare const node: { name: string } | null;\n" +
	"declare const either: { name: string } | null | undefined;\n" +
	"declare const fallback: boolean | undefined;\n" +
	"declare const text: string;\n" +
	"declare const count: number | undefined;\n" +
	"declare const flag: boolean;\n" +
	"declare function load(): { name: string } | undefined;\n"

// compileErrors type-checks one file under the harness's strict config and returns its diagnostics.
func compileErrors(t *testing.T, source string) []string {
	t.Helper()
	directory := t.TempDir()
	config := `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"moduleDetection":"force","types":[],"noEmit":true},"include":["*.ts"]}`
	if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	fileName := filepath.Join(directory, "fixture.ts")
	if err := os.WriteFile(fileName, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	graph, err := program.Build(program.Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building the program: %v", err)
	}
	messages := []string{}
	for _, sourceFile := range graph.ProjectFiles() {
		for _, diagnostic := range graph.Diagnostics(context.Background(), sourceFile) {
			messages = append(messages, diagnostic.String())
		}
	}
	return messages
}

func TestNoImplicitCoercionFixKeepsNarrowing(t *testing.T) {
	t.Parallel()

	// The control first. If the old output compiled, the check below would pass for any fix.
	t.Run("the Boolean rewrite of the phi shape does not compile", func(t *testing.T) {
		t.Parallel()
		broken := narrowingDeclarations + "const data = request.data;\nconst hasData = Boolean(data);\n" +
			"export const total = hasData ? data.energy.credits : 0;\n"
		if errors := compileErrors(t, broken); len(errors) == 0 {
			t.Fatal("the Boolean rewrite compiled, so this test cannot tell a narrowing-safe fix from a broken one")
		}
	})

	for _, testCase := range []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "AppSidebarEnergyBalance.tsx: undefined only, aliased into a const",
			source: "const data = request.data;\nconst hasData = !!data;\nexport const total = hasData ? data.energy.credits : 0;\n",
			want:   "const data = request.data;\nconst hasData = data !== undefined;\nexport const total = hasData ? data.energy.credits : 0;\n",
		},
		{
			name:   "null only",
			source: "const hasNode = !!node;\nexport const name = hasNode ? node.name : '';\n",
			want:   "const hasNode = node !== null;\nexport const name = hasNode ? node.name : '';\n",
		},
		{
			name:   "null and undefined",
			source: "const hasEither = !!either;\nexport const name = hasEither ? either.name : '';\n",
			want:   "const hasEither = either !== null && either !== undefined;\nexport const name = hasEither ? either.name : '';\n",
		},
		// The two comparisons joined by `&&` cannot sit bare beside `??`, which is a syntax error,
		// so the replacement keeps its own parentheses there.
		{
			name:   "null and undefined beside a nullish coalescing operator",
			source: "export const known = fallback ?? !!either;\n",
			want:   "export const known = fallback ?? (either !== null && either !== undefined);\n",
		},
		{
			name:   "a property reference, used directly as a condition",
			source: "export const credits = !!request.data ? request.data.energy.credits : 0;\n",
			want:   "export const credits = request.data !== undefined ? request.data.energy.credits : 0;\n",
		},
		// The plain shapes keep upstream's fix, because narrowing either cannot happen or changes
		// nothing: a call result is not a reference, and truthiness does not narrow `string`.
		{
			name:   "a call result",
			source: "export const loaded = !!load();\n",
			want:   "export const loaded = Boolean(load());\n",
		},
		{
			name:   "a string",
			source: "export const hasText = !!text;\n",
			want:   "export const hasText = Boolean(text);\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				narrowingDeclarations+testCase.source, nil)
			rule_testing.ExpectFindings(t, result, "implicitCoercion")
			rule_testing.ExpectFixedSource(t, result, narrowingDeclarations+testCase.want)
			if errors := compileErrors(t, narrowingDeclarations+testCase.want); len(errors) > 0 {
				t.Errorf("the fixed source does not compile:\n%s", strings.Join(errors, "\n"))
			}
		})
	}

	// No comparison means what `!!x` means here, because the type holds a falsy value that is not
	// nullish (`0`, `false`), and `Boolean(x)` would lose a narrowing the type makes visible. The
	// finding stays, offered as a suggestion for a person to judge, never applied unattended.
	for _, testCase := range []struct {
		name   string
		source string
	}{
		{name: "a number that may be undefined", source: "const hasCount = !!count;\nexport const next = hasCount ? count + 1 : 0;\n"},
		// A plain `boolean` is no longer here: it is already a boolean, so `!!flag` converts nothing
		// and the rule is silent (#vsy2eym). Beside undefined, `false` is still a falsy value no
		// comparison removes.
		{name: "a boolean that may be undefined", source: "export const on = !!fallback;\n"},
	} {
		t.Run("declines: "+testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoImplicitCoercion, noImplicitCoercionFile,
				narrowingDeclarations+testCase.source, nil)
			rule_testing.ExpectFindings(t, result, "implicitCoercion")
			for _, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) > 0 {
					t.Errorf("a fix was proposed where no rewrite is safe: %q", diagnostic.Fixes[0].Text)
				}
				if len(diagnostic.Suggestions) == 0 {
					t.Error("the declined fix was not offered as a suggestion")
				}
			}
		})
	}
}

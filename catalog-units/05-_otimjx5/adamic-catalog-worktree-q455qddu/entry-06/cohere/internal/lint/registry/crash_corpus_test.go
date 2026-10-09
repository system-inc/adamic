package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/internal/types/program"
)

// crashShapes are source snippets whose optional nodes are absent.
//
// Every one of these is ordinary code a rule meets on a real tree, or code someone is mid-edit
// through. A declaration with no initializer crashed react-hook-require-result-naming on 42 real
// files; a binding pattern with no initializer is a syntax error that still reaches the walk while
// someone is typing.
//
// Kept short deliberately. This corpus is not trying to be a fuzzer: it is the specific shapes that
// have crashed a rule here, plus the neighbours of those shapes, and each one earns its line by
// having caught something or by being one edit away from a shape that did.
var crashShapes = []string{
	"declare const memo: unknown;\nexport const value = 1;\n",
	"declare const Field: { displayName: string };\nlet other;\nexport const value = other;\n",
	"export function Page() { return null; }\n",

	// A declaration with no initializer inside a component.
	"export function Panel() {\n    let pending;\n    return pending;\n}\n",
	"export function Panel() {\n    declare const value: unknown;\n    return value;\n}\n",
	"export default function ThingPageRoute() {\n    let pending;\n    return pending;\n}\n",

	// A binding pattern with no initializer. A syntax error, and it reaches the walk anyway.
	"export function Panel() {\n    let { label };\n    return label;\n}\n",

	// A shorthand property, whose initializer is nil where a longhand one is not.
	"import { networkService } from './NetworkService.ts';\ndeclare const Document: unknown;\n" +
		"declare const identifier: string;\n" +
		"export function useThingRequest(variables: { identifier: string }) {\n" +
		"    return networkService.useGraphQlQuery(Document, { identifier });\n}\n",

	// A call with no arguments where the rule expects some.
	"import { networkService } from './NetworkService.ts';\n" +
		"export function useThingRequest() { return networkService.useGraphQlQuery(); }\n",

	// An empty file, and a file that is only a comment. Both are real, and both leave every
	// optional node absent at once.
	"",
	"// nothing here\n",

	// A node that is present and of a kind the rule did not expect, which is a different class from
	// every shape above. A shared `case A, B:` arm reaching for one kind's accessor panics on the
	// other, because `As*()` is an interface conversion rather than a cast: `AsCallExpression()` on a
	// NewExpression crashed prefer-arrow-callback on 71 real files (#qa9nttp). Its neighbours are the
	// other pairs a porter would put in one arm.
	"declare class Foo { constructor(callback: () => void); }\nexport const made = new Foo(function () {});\n",
	"declare class Foo { constructor(callback: () => void); }\nexport const made = new Foo(() => {});\n",
	"declare function tag(strings: TemplateStringsArray): string;\nexport const tagged = tag`text`;\n",
	"export class Base {}\nexport class Derived extends Base { constructor() { super(); } }\n",
}

// typedCorpusConfig is the corpus's tsconfig. The harness's own, plus what the shapes ask of it: JSX
// for the .tsx paths, and imports written with their .ts extension.
const typedCorpusConfig = `{
	"compilerOptions": {
		"strict": true,
		"target": "ES2022",
		"lib": ["ES2022"],
		"module": "ESNext",
		"moduleResolution": "Bundler",
		"allowImportingTsExtensions": true,
		"noEmit": true,
		"jsx": "react-jsx",
		"moduleDetection": "force",
		"types": []
	},
	"include": ["**/*.ts", "**/*.tsx"]
}`

// networkServiceStub is what the shapes import as './NetworkService.ts'.
const networkServiceStub = "export const networkService = {\n" +
	"    useGraphQlQuery(...parameters: unknown[]): unknown { return parameters; },\n};\n"

// No registered rule may crash on a shape whose optional nodes are absent.
//
// This runs every rule in the registry rather than a list someone maintains. The list this replaces
// held 21 entries with one duplicated, covering 20 of the 77 registered rules, and it could not
// cover more: it lived in `package structure`, which cannot import the registry that imports it. So
// 57 rules were never offered a crash shape, and nothing said so.
//
// A rule that panics is worse than a rule that is wrong. `fdcf866` added a per-file boundary so one
// panic loses one file rather than the run, and that boundary is what keeps a tree verifiable while
// this guard finds the panics. Neither replaces the other.
func TestNoRegisteredRuleCrashesOnAbsentOptionalNodes(t *testing.T) {
	t.Parallel()
	rules := All()
	if len(rules) == 0 {
		// A sweep that finds nothing to check passes for the wrong reason, which is the same shape
		// as the defect it guards against.
		t.Fatal("the registry is empty, so this test proved nothing")
	}

	// Both extensions and a route path, because a React-gated rule declines a .ts file before
	// reaching any shape below it. Running only .ts made the previous version of this guard silent
	// for every rule with an IsReactFile check, and it is why that guard missed a nil dereference
	// that crashed 42 real files: it had the right shape in its list and never delivered it.
	fileNames := []string{
		"/repository/source/api/ThingRequest.ts",
		"/repository/source/components/ThingRequest.tsx",
		"/repository/app/thing/page.tsx",
	}

	// One panic ends the whole run, because Go's test binary does not survive one: the first crash
	// is reported and every rule after it goes unmeasured. So a failing run names one real defect
	// and says nothing about the rest, and a reader must not read "one crash" as "one crash exists".
	//
	// Not worked around by recovering here. A recover would let the run continue and report every
	// crash at once, and it would also make this guard pass on a rule that panics, because the
	// panic it exists to surface would be caught by the thing surfacing it. The per-file boundary in
	// the walk is where recovery belongs; a test that recovers from the failure it tests for is the
	// vacuous shape this repository keeps finding.
	for _, subject := range rules {
		for index, source := range crashShapes {
			t.Run(fmt.Sprintf("%s/shape-%d", subject.Name, index), func(t *testing.T) {
				t.Parallel()
				// A panic fails the test. The only assertion is that this returns at all: what each
				// rule concludes about these shapes belongs in its own fixture pair, and asserting
				// it here would make this guard fail for reasons that are not crashes.
				for _, fileName := range fileNames {
					rule_testing.Run(t, subject, fileName, source)
				}
			})
		}
	}
}

// No registered rule may crash on the corpus when it has types.
//
// The sweep above runs every rule through rule_testing.Run, which walks with no checker. 154 of the
// registered rules then return early or skip every path that asks a type question, so the sweep
// passed while their typed half never ran. prefer-arrow-callback's shared arm was the proof: its own
// regression test panicked, all 111 upstream fixtures passed, and the sweep above passed, because the
// arm sat behind a checker. Adding a shape could not fix that; the harness was the gap (#qa9nttp).
//
// So the corpus is also built into one typed program, every shape at every path, and one walk runs
// every registered rule over it. One build rather than one per rule and shape is what makes the typed
// half affordable. The walk contains a rule's panic to that rule and names it (cf5afb6), so this reads
// the crashes the walk names rather than recovering on its own, and a run with several crashing
// rules names every one.
func TestNoRegisteredRuleCrashesOnTheCorpusWithTypes(t *testing.T) {
	t.Parallel()
	rules := All()
	directory := t.TempDir()
	files := map[string]string{"tsconfig.json": typedCorpusConfig}
	for index, source := range crashShapes {
		for _, folder := range []string{"source/api", "source/components", fmt.Sprintf("app/thing%d", index)} {
			files[folder+"/NetworkService.ts"] = networkServiceStub
		}
		files[fmt.Sprintf("source/api/ThingRequest%d.ts", index)] = source
		files[fmt.Sprintf("source/components/ThingRequest%d.tsx", index)] = source
		files[fmt.Sprintf("app/thing%d/page.tsx", index)] = source
	}
	for name, contents := range files {
		path := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(directory, "tsconfig.json"), CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building the corpus program: %v", err)
	}
	result, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules)
	if err != nil {
		t.Fatalf("walking the corpus: %v", err)
	}

	// Every type-aware rule must have been handed a file, or this proved nothing about it.
	typed := 0
	for _, subject := range rules {
		if !subject.NeedsTypeChecker {
			continue
		}
		typed++
		if result.Coverage.RulesOffered[subject.Name] == 0 {
			t.Errorf("type-aware rule %s was offered no file of the corpus, so its typed half went unswept", subject.Name)
		}
	}
	if typed == 0 {
		t.Fatal("no registered rule declares NeedsTypeChecker, so the typed sweep proved nothing")
	}

	for _, crash := range result.Coverage.RulesCrashed {
		t.Errorf("rule %s crashed on %s with types: %v", crash.RuleName, filepath.Base(crash.FileName), crash.Cause)
	}
	for _, crash := range result.Coverage.FilesCrashed {
		t.Errorf("%s crashed outside any rule with types: %v", filepath.Base(crash.FileName), crash.Cause)
	}
}

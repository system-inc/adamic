package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestUnextractableDependencyAbandonsTheWholeList pins that a list we cannot read is not an empty
// list.
//
// Upstream calls `env.recordError` on an entry that is not a simple access path
// (`DropManualMemoization.ts:356`), which is a compile error: the function bails and
// `ValidatePreservedManualMemoization` never runs on it. Dropping the entry and keeping the rest
// turns "this list could not be read" into "this list was empty", and those are different claims --
// an empty list is the developer promising no dependencies, so every inferred one is then reported
// as a disagreement they never made.
//
// A nil `Deps` is the closest faithful equivalent available here, and this pass already documents it
// as the signal that disables the comparison.
//
// Measured on `useMemo-dep-array-literal-access.ts`, whose deps list is `[x[0]]` and whose own
// comment says upstream recognises only "hoistable" values there: we extracted zero of one entry,
// handed the validator an empty list, and it reported `props` against nothing. Upstream compiles
// that fixture with no error at all. Fixing it took ungated false positives from 35 to 34 with
// golden unchanged.
func TestUnextractableDependencyAbandonsTheWholeList(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		source      string
		wantNilDeps bool
		wantEntries int
		wantDropped int
	}{
		{
			name: "a computed access cannot be extracted",
			source: `
				import {useMemo} from 'react';
				function Foo(props) {
					const x = [props];
					return useMemo(() => [x[0]], [x[0]]);
				}
			`,
			wantNilDeps: true,
			wantEntries: 0,
			wantDropped: 1,
		},
		{
			name: "a plain list is extracted whole",
			source: `
				import {useMemo} from 'react';
				function Foo(props) {
					return useMemo(() => [props.a], [props.a]);
				}
			`,
			wantNilDeps: false,
			wantEntries: 1,
			wantDropped: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			deps, isNil, result := memoDepsFor(t, testCase.source)
			if result.Recognised == 0 {
				t.Fatal("no memo call was recognised, so this test asserts nothing about its " +
					"dependency list")
			}
			if isNil != testCase.wantNilDeps {
				t.Errorf("nil deps = %v, want %v; a nil list disables the comparison and an empty "+
					"one asserts the developer promised nothing", isNil, testCase.wantNilDeps)
			}
			if len(deps) != testCase.wantEntries {
				t.Errorf("extracted %d entries, want %d", len(deps), testCase.wantEntries)
			}
			if result.UnextractableDeps != testCase.wantDropped {
				t.Errorf("dropped %d entries, want %d", result.UnextractableDeps,
					testCase.wantDropped)
			}
		})
	}
}

// Upstream's `findOptionalPlaces` derives optionality from Optional terminals before it collects
// the source dependency list. Optional lowering has the same shape here, so the written path must
// retain exactly the steps guarded by `?.`, even though the PropertyLoad instructions themselves
// are non-optional.
func TestDropManualMemoizationRecoversOptionalPlacesFromControlFlow(t *testing.T) {
	t.Parallel()

	deps, isNil, result := memoDepsFor(t, `
		import {useMemo} from 'react';
		function Component(props) {
			return useMemo(() => props.items.edges.nodes, [props.items?.edges?.nodes]);
		}
	`)
	if result.Recognised != 1 {
		t.Fatalf("recognised %d memo calls, want 1", result.Recognised)
	}
	if isNil || len(deps) != 1 {
		t.Fatalf("dependency list is nil=%v with %d entries, want one extracted dependency",
			isNil, len(deps))
	}
	path := deps[0].Path
	want := []DependencyPathEntry{
		{Property: "items", Optional: false},
		{Property: "edges", Optional: true},
		{Property: "nodes", Optional: true},
	}
	if len(path) != len(want) {
		t.Fatalf("path = %v, want %v", path, want)
	}
	for index := range want {
		if path[index].Property != want[index].Property ||
			path[index].Optional != want[index].Optional {
			t.Errorf("path[%d] = %+v, want %+v", index, path[index], want[index])
		}
	}
}

// The inference sidemap intentionally rejects LoadGlobal roots, but source dependency extraction
// must not: upstream accepts globals in dependency arrays and derives their optionality with a
// separate terminal walk.
func TestDropManualMemoizationRecoversOptionalGlobalDependency(t *testing.T) {
	t.Parallel()

	deps, isNil, result := memoDepsFor(t, `
		import {useMemo} from 'react';
		declare const GLOBAL: {field?: {leaf: number}} | undefined;
		function Component() {
			return useMemo(() => GLOBAL?.field?.leaf, [GLOBAL?.field?.leaf]);
		}
	`)
	if result.Recognised != 1 {
		t.Fatalf("recognised %d memo calls, want 1", result.Recognised)
	}
	if isNil || len(deps) != 1 {
		t.Fatalf("dependency list is nil=%v with %d entries, want one extracted dependency",
			isNil, len(deps))
	}
	dependency := deps[0]
	if !dependency.Root.IsGlobal || dependency.Root.Name != "GLOBAL" {
		t.Fatalf("root = %+v, want global GLOBAL", dependency.Root)
	}
	want := []DependencyPathEntry{
		{Property: "field", Optional: true},
		{Property: "leaf", Optional: true},
	}
	if !equalPaths(dependency.Path, want) {
		t.Fatalf("path = %+v, want %+v", dependency.Path, want)
	}
}

// A phi with one optional operand is not necessarily the optional chain's own join. Projecting
// through the enclosing ternary would accept a dependency expression upstream classifies as
// unextractable and silently compare against only one of its two possible values.
func TestDropManualMemoizationDoesNotProjectOptionalThroughTernaryPhi(t *testing.T) {
	t.Parallel()

	deps, isNil, result := memoDepsFor(t, `
		import {useMemo} from 'react';
		function Component({condition, value}) {
			return useMemo(() => value, [condition ? value?.field : value + 1]);
		}
	`)
	if result.Recognised != 1 {
		t.Fatalf("recognised %d memo calls, want 1", result.Recognised)
	}
	if !isNil || len(deps) != 0 {
		t.Fatalf("dependency list is nil=%v with %d entries, want abandoned unextractable list",
			isNil, len(deps))
	}
	if result.UnextractableDeps != 1 {
		t.Fatalf("unextractable dependencies = %d, want 1", result.UnextractableDeps)
	}
}

// memoDepsFor returns the first memo marker's written dependencies, whether they are nil, and the
// pass result.
func memoDepsFor(t *testing.T, source string) ([]ManualMemoDependency, bool, ManualMemoization) {
	t.Helper()
	var deps []ManualMemoDependency
	isNil := true
	var result ManualMemoization
	probe := rule.Rule{
		Name:             "memo-deps",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the typed harness handed this probe a nil checker; without one " +
							"no memo call is recognised and every assertion below is vacuous")
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						if result.Recognised > 0 {
							return
						}
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						InferReactive(function, ctx.TypeChecker)
						result = DropManualMemoization(function)
						for _, block := range function.Blocks {
							if block == nil {
								continue
							}
							for _, instructionId := range block.Instructions {
								instruction := function.Instructions[instructionId]
								if instruction == nil {
									continue
								}
								if marker, ok := instruction.Value.(*StartMemoize); ok {
									deps = marker.Deps
									isNil = marker.Deps == nil
									return
								}
							}
						}
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")
	return deps, isNil, result
}

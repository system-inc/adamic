# Task #39wtqdj — lower-39w-6-1 measurement

The `export default items` false-side read does not reach a successfully lowered program in either A or B, or in the requested B variant: Lower refuses the export at items.ts:7:1. The separate B CLI build confirms that refusal. Node prints `object`; the pinned TypeScript 6.0.3 strict check passes when run from the external probe directory.

A marks both the named-export and label probes' false directions checked. B marks both unobservable. Removing `ast.IsExpressionNode(node)` at B makes the label probe checked again: both label identifiers match the checker reference, carry flow reaching the predicate's false side, and become `predicateNameReads yes`, even though neither reads the value. The named-export probe stays unobservable because its export specifier is a declaration name and `predicateNameReads` still rejects it. This measurement demonstrates what the broader filter admits; passing the selected suites does not establish that label positions are value reads.

The supplied test compiled unchanged at every revision. No throwaway-file fix was needed. It logs without assertions, so PASS alone is not the agreement result; the logged lowering outcomes and read classifications are the evidence.

The test does not log an export-assignment expression or identifiers in items.ts in the default-export probe. `load.Load` receives only main.ts, and `Program.Files()` returns the explicit roots, not imported dependencies (internal/load/load.go:187). The main.ts importer operand is logged as `readonly string[] | undefined` in all three measurements. No export-expression type or default-export read-filter admission is inferred from a missing log line.

A = `63eb56cd701c744a603f3c694295d749ba9d46dc` (`compiler/predicate-false-read-4`).
B = `5ff2d7f5701a2fd32e4d9a1406bf3058ea0b2617` (`compiler/predicate-false-read-5`).

Repository: `/workspace/adamic`. Loaded `/workspace/adamic-tools/env.sh` before repository work. The initial worktree was clean. Fetched with `git fetch origin compiler/predicate-false-read-4 compiler/predicate-false-read-5` (exit 0). Checked out each revision detached and ran `git submodule update --init --recursive` successfully at each. Pinned submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Created `/tmp/adamic-gate` outside the repository before Go ran because env.sh sets TMPDIR to it. All heavy commands ran sequentially; none ran concurrently. The logging test, overlay, probes, command captures, upstream checkout and package install were outside the repository. No test was placed in internal/lower.

Provisioned upstream with `git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/measure39w6/TypeScript`; verified HEAD `050880ce59e30b356b686bd3144efe24f875ebc8`. Exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/measure39w6/TypeScript`. Copied stage3/api/package.json and package-lock.json to `/tmp/measure39w6/api` and ran `npm ci --prefix /tmp/measure39w6/api`. Put `/tmp/measure39w6/api/node_modules/.bin` first on PATH. Versions: tsc 6.0.3, Node v24.19.0, Go go1.27.1 linux/amd64.

Exported these variables for npm and every measurement command, inherited by admission-gate children:

```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```

Overlay `/tmp/measure39w6/overlay.json`:

```json
{"Replace": {"/workspace/adamic/internal/lower/measure39w6_test.go": "/tmp/measure39w6/measure39w6_test.go"}}
```

The following captures preserve complete stdout and stderr, including every t.Logf line, temporary path, diagnostic, and each selected subtest result. Empty fenced blocks denote empty streams. Exit codes are those of the measured commands, captured directly by subprocess.run. The capture wrapper itself returns successfully after recording either outcome.

**External pinned package installation**

Working directory: `/workspace/adamic`.

```sh
npm ci --prefix /tmp/measure39w6/api
```

Exit: 0.

stdout (verbatim):

```text

added 3 packages in 1s
```

stderr (verbatim):

```text
npm notice
npm notice New major version of npm available! 11.9.0 -> 12.2.2
npm notice Changelog: https://github.com/npm/cli/releases/tag/v12.2.2
npm notice To update run: npm install -g npm@12.2.2
npm notice
```

**Step 2 — A overlay measurement**

Working directory: `/workspace/adamic`.

```sh
go test -overlay /tmp/measure39w6/overlay.json ./internal/lower -run '^TestMeasure39w6$' -count=1 -v -timeout 20m
```

Exit: 0.

stdout (verbatim):

```text
=== RUN   TestMeasure39w6
    measure39w6_test.go:57: export_default: Lower error: /tmp/adamic-gate/TestMeasure39w6164252343/001/items.ts:7:1: Adamic 0.1 refuses export default; export by name: one name for one thing
    measure39w6_test.go:84: export_default: typeof operand at main.ts:2:20 typed readonly string[] | undefined
    measure39w6_test.go:59: export_named: Lower ok, predicate sites [{Where:/tmp/adamic-gate/TestMeasure39w6164252343/002/main.ts:6:5 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] | undefined -> readonly string[]}]}]
    measure39w6_test.go:111: export_named: call at main.ts:6:5 argument at main.ts:6:10
    measure39w6_test.go:123:   identifier at main.ts:5:7 parent KindVariableDeclaration: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:6:10 parent KindCallExpression: IsExpressionNode yes, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:7:10 parent KindExportSpecifier: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches yes
    measure39w6_test.go:130:   predicateUseDirections false no, predicateFalseRead yes, checkedPredicateDirections 3
    measure39w6_test.go:59: label: Lower ok, predicate sites [{Where:/tmp/adamic-gate/TestMeasure39w6164252343/003/main.ts:5:65 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] | undefined -> readonly string[]}]}]
    measure39w6_test.go:111: label: call at main.ts:5:65 argument at main.ts:5:70
    measure39w6_test.go:123:   identifier at main.ts:5:14 parent KindParameter: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:5:70 parent KindCallExpression: IsExpressionNode yes, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:5:80 parent KindLabeledStatement: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches yes
    measure39w6_test.go:123:   identifier at main.ts:5:104 parent KindBreakStatement: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches yes
    measure39w6_test.go:130:   predicateUseDirections false yes, predicateFalseRead yes, checkedPredicateDirections 3
--- PASS: TestMeasure39w6 (0.13s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.144s
```

stderr (verbatim):

```text
```

**Step 2 — B overlay measurement**

Working directory: `/workspace/adamic`.

```sh
go test -overlay /tmp/measure39w6/overlay.json ./internal/lower -run '^TestMeasure39w6$' -count=1 -v -timeout 20m
```

Exit: 0.

stdout (verbatim):

```text
=== RUN   TestMeasure39w6
    measure39w6_test.go:57: export_default: Lower error: /tmp/adamic-gate/TestMeasure39w6625820161/001/items.ts:7:1: Adamic 0.1 refuses export default; export by name: one name for one thing
    measure39w6_test.go:84: export_default: typeof operand at main.ts:2:20 typed readonly string[] | undefined
    measure39w6_test.go:59: export_named: Lower ok, predicate sites [{Where:/tmp/adamic-gate/TestMeasure39w6625820161/002/main.ts:6:5 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
    measure39w6_test.go:111: export_named: call at main.ts:6:5 argument at main.ts:6:10
    measure39w6_test.go:123:   identifier at main.ts:5:7 parent KindVariableDeclaration: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:6:10 parent KindCallExpression: IsExpressionNode yes, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:7:10 parent KindExportSpecifier: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches yes
    measure39w6_test.go:130:   predicateUseDirections false no, predicateFalseRead no, checkedPredicateDirections 1
    measure39w6_test.go:59: label: Lower ok, predicate sites [{Where:/tmp/adamic-gate/TestMeasure39w6625820161/003/main.ts:5:65 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
    measure39w6_test.go:111: label: call at main.ts:5:65 argument at main.ts:5:70
    measure39w6_test.go:123:   identifier at main.ts:5:14 parent KindParameter: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:5:70 parent KindCallExpression: IsExpressionNode yes, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:5:80 parent KindLabeledStatement: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches yes
    measure39w6_test.go:123:   identifier at main.ts:5:104 parent KindBreakStatement: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches yes
    measure39w6_test.go:130:   predicateUseDirections false no, predicateFalseRead no, checkedPredicateDirections 1
--- PASS: TestMeasure39w6 (0.13s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.145s
```

stderr (verbatim):

```text
```

**Step 3 — exact external TypeScript probes**

`/tmp/measure39w6/export_default/items.ts`:

```typescript
function some<T>(array: readonly T[] | undefined): array is readonly T[];
function some<T>(array: readonly T[] | undefined, predicate: (value: T) => boolean): boolean;
function some<T>(array: readonly T[] | undefined, predicate?: (value: T) => boolean): boolean { if (array === undefined) return false; if (predicate === undefined) return array.length > 0; for (const value of array) { if (predicate(value)) return true; } return false; }
function load(): readonly string[] | undefined { return []; }
const items = load();
if (some(items)) throw new Error("x");
export default items;
```

`/tmp/measure39w6/export_default/main.ts`:

```typescript
import items from "./items.ts";
console.log(typeof items);
```

**Step 3 — Node oracle at B**

Working directory: `/workspace/adamic`.

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/measure39w6/export_default/main.ts
```

Exit: 0.

stdout (verbatim):

```text
object
```

stderr (verbatim):

```text
```

**Step 3 — tsc initial invocation at repository root**

Working directory: `/workspace/adamic`.

```sh
tsc --noEmit --strict --target es2022 --module nodenext --allowImportingTsExtensions /tmp/measure39w6/export_default/main.ts
```

Exit: 1.

stdout (verbatim):

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

stderr (verbatim):

```text
```

The root invocation reports TS5112 because the repository has tsconfig.json. Repeated the exact same command without changing its flags from the external probe directory, where that config guard does not apply. Both attempts are retained.

**Step 3 — unchanged tsc command from external probe directory**

Working directory: `/tmp/measure39w6/export_default`.

```sh
tsc --noEmit --strict --target es2022 --module nodenext --allowImportingTsExtensions /tmp/measure39w6/export_default/main.ts
```

Exit: 0.

stdout (verbatim):

```text
```

stderr (verbatim):

```text
```

**Step 3 — B Adamic CLI build**

Working directory: `/workspace/adamic`.

```sh
go run ./cmd/adamic build /tmp/measure39w6/export_default/main.ts -o /tmp/measure39w6/export_default/out
```

Exit: 1.

stdout (verbatim):

```text
```

stderr (verbatim):

```text
adamic: /tmp/measure39w6/export_default/items.ts:7:1: Adamic 0.1 refuses export default; export by name: one name for one thing
exit status 1
```

**Step 4 — temporary B variant**

Asserted the exact original return expression occurred exactly once in internal/lower/predicates_proof.go, then replaced only that occurrence:

```diff
-return ast.IsExpressionNode(node) && !ast.IsRightSideOfPropertyAccess(node)
+return !ast.IsRightSideOfPropertyAccess(node)
```

No other compiler change was made. The original bytes were saved outside the repository for restoration.

**Step 4 — B variant overlay measurement**

Working directory: `/workspace/adamic`.

```sh
go test -overlay /tmp/measure39w6/overlay.json ./internal/lower -run '^TestMeasure39w6$' -count=1 -v -timeout 20m
```

Exit: 0.

stdout (verbatim):

```text
=== RUN   TestMeasure39w6
    measure39w6_test.go:57: export_default: Lower error: /tmp/adamic-gate/TestMeasure39w64015022458/001/items.ts:7:1: Adamic 0.1 refuses export default; export by name: one name for one thing
    measure39w6_test.go:84: export_default: typeof operand at main.ts:2:20 typed readonly string[] | undefined
    measure39w6_test.go:59: export_named: Lower ok, predicate sites [{Where:/tmp/adamic-gate/TestMeasure39w64015022458/002/main.ts:6:5 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:unobservable Reason:readonly string[] | undefined -> readonly string[]}]}]
    measure39w6_test.go:111: export_named: call at main.ts:6:5 argument at main.ts:6:10
    measure39w6_test.go:123:   identifier at main.ts:5:7 parent KindVariableDeclaration: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:6:10 parent KindCallExpression: IsExpressionNode yes, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:7:10 parent KindExportSpecifier: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches yes
    measure39w6_test.go:130:   predicateUseDirections false no, predicateFalseRead no, checkedPredicateDirections 1
    measure39w6_test.go:59: label: Lower ok, predicate sites [{Where:/tmp/adamic-gate/TestMeasure39w64015022458/003/main.ts:5:65 Function:some Overload:1 Directions:[{Direction:true Status:checked Reason:readonly string[] | undefined -> readonly string[]} {Direction:false Status:checked Reason:readonly string[] | undefined -> readonly string[]}]}]
    measure39w6_test.go:111: label: call at main.ts:5:65 argument at main.ts:5:70
    measure39w6_test.go:123:   identifier at main.ts:5:14 parent KindParameter: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName yes, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads no, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:5:70 parent KindCallExpression: IsExpressionNode yes, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches no
    measure39w6_test.go:123:   identifier at main.ts:5:80 parent KindLabeledStatement: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches yes
    measure39w6_test.go:123:   identifier at main.ts:5:104 parent KindBreakStatement: IsExpressionNode no, IsRightSideOfPropertyAccess no, IsDeclarationName no, IsPartOfTypeNode no, predicateWriteOnly no, checker matches yes, predicateNameReads yes, FlowNode-nil no, reaches yes
    measure39w6_test.go:130:   predicateUseDirections false yes, predicateFalseRead yes, checkedPredicateDirections 3
--- PASS: TestMeasure39w6 (0.16s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.172s
```

stderr (verbatim):

```text
```

Step 4a — selected lower tests: result events {'PASS': 96, 'FAIL': 0, 'SKIP': 0}; subtest result events {'PASS': 78, 'FAIL': 0, 'SKIP': 0}. Every per-subtest PASS/FAIL is retained below. Complete output also preserves every failure's whole got/want diagnostics without shortening.

**Step 4a — selected lower tests**

Working directory: `/workspace/adamic`.

```sh
go test ./internal/lower -run 'TestBrandedSupertypeIsJudgedWhereItIsLowered|PredicateFalseRead|PredicateOverload|KindPredicate|BrandedSupertypeIdentity' -count=1 -v -timeout 40m
```

Exit: 0.

stdout (verbatim):

```text
=== RUN   TestKindPredicateCallIsItsComparison
=== PAUSE TestKindPredicateCallIsItsComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison
=== PAUSE TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateWalkRefusals
=== PAUSE TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateThroughModules
=== PAUSE TestKindPredicateThroughModules
=== RUN   TestPredicateOverloadInPlaceCallback
=== PAUSE TestPredicateOverloadInPlaceCallback
=== RUN   TestPredicateOverloadNamedCallback
=== PAUSE TestPredicateOverloadNamedCallback
=== RUN   TestPredicateOverloadCallbackPushes
=== PAUSE TestPredicateOverloadCallbackPushes
=== RUN   TestPredicateOverloadReturnsAnotherRead
=== PAUSE TestPredicateOverloadReturnsAnotherRead
=== RUN   TestPredicateOverloadLocalWrittenAfterTheCall
=== PAUSE TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestPredicateOverloadWithoutPredicate
=== PAUSE TestPredicateOverloadWithoutPredicate
=== RUN   TestPredicateOverloadCallbackFromParameter
=== PAUSE TestPredicateOverloadCallbackFromParameter
=== RUN   TestPredicateOverloadStaysOutOfA
=== PAUSE TestPredicateOverloadStaysOutOfA
=== RUN   TestPredicateOverloadLanguageLies
=== PAUSE TestPredicateOverloadLanguageLies
=== RUN   TestBrandedSupertypeIdentityNarrowingLowers
=== PAUSE TestBrandedSupertypeIdentityNarrowingLowers
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections
=== PAUSE TestCheckedPredicateFalseReadDirections
=== RUN   TestPredicateOverloadRuntime
=== PAUSE TestPredicateOverloadRuntime
=== RUN   TestIndirectPredicateOverloadIsPending
=== PAUSE TestIndirectPredicateOverloadIsPending
=== CONT  TestKindPredicateCallIsItsComparison
=== CONT  TestPredicateOverloadWithoutPredicate
=== CONT  TestPredicateOverloadNamedCallback
=== RUN   TestKindPredicateCallIsItsComparison/true_and_false
=== PAUSE TestKindPredicateCallIsItsComparison/true_and_false
=== CONT  TestPredicateOverloadInPlaceCallback
=== CONT  TestPredicateOverloadLocalWrittenAfterTheCall
=== RUN   TestKindPredicateCallIsItsComparison/negated
=== PAUSE TestKindPredicateCallIsItsComparison/negated
=== RUN   TestKindPredicateCallIsItsComparison/conditional
=== PAUSE TestKindPredicateCallIsItsComparison/conditional
=== RUN   TestKindPredicateCallIsItsComparison/early_return
=== PAUSE TestKindPredicateCallIsItsComparison/early_return
=== RUN   TestKindPredicateCallIsItsComparison/===_false
=== PAUSE TestKindPredicateCallIsItsComparison/===_false
=== RUN   TestKindPredicateCallIsItsComparison/!==_true
=== PAUSE TestKindPredicateCallIsItsComparison/!==_true
=== RUN   TestKindPredicateCallIsItsComparison/const_arrow
=== PAUSE TestKindPredicateCallIsItsComparison/const_arrow
=== RUN   TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== PAUSE TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== CONT  TestKindPredicateThroughModules
=== RUN   TestKindPredicateThroughModules/import_as
=== PAUSE TestKindPredicateThroughModules/import_as
=== RUN   TestKindPredicateThroughModules/re-export_as
=== PAUSE TestKindPredicateThroughModules/re-export_as
=== RUN   TestKindPredicateThroughModules/namespace_import
=== PAUSE TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateWalkRefusals
=== RUN   TestKindPredicateWalkRefusals/comparison_control
=== PAUSE TestKindPredicateWalkRefusals/comparison_control
=== RUN   TestKindPredicateWalkRefusals/wide_kind_call
=== PAUSE TestKindPredicateWalkRefusals/wide_kind_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_call
=== PAUSE TestKindPredicateWalkRefusals/shared_value_call
=== RUN   TestKindPredicateWalkRefusals/shared_value_comparison
=== PAUSE TestKindPredicateWalkRefusals/shared_value_comparison
=== RUN   TestKindPredicateWalkRefusals/lying_declared_type
=== PAUSE TestKindPredicateWalkRefusals/lying_declared_type
=== RUN   TestKindPredicateWalkRefusals/twin_layouts
=== PAUSE TestKindPredicateWalkRefusals/twin_layouts
=== RUN   TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== PAUSE TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== RUN   TestKindPredicateWalkRefusals/generic,_reduced
=== PAUSE TestKindPredicateWalkRefusals/generic,_reduced
=== RUN   TestKindPredicateWalkRefusals/this_parameter
=== PAUSE TestKindPredicateWalkRefusals/this_parameter
=== RUN   TestKindPredicateWalkRefusals/inferred_predicate
=== PAUSE TestKindPredicateWalkRefusals/inferred_predicate
=== RUN   TestKindPredicateWalkRefusals/overloaded
=== PAUSE TestKindPredicateWalkRefusals/overloaded
=== RUN   TestKindPredicateWalkRefusals/method
=== PAUSE TestKindPredicateWalkRefusals/method
=== RUN   TestKindPredicateWalkRefusals/value_alias
=== PAUSE TestKindPredicateWalkRefusals/value_alias
=== RUN   TestKindPredicateWalkRefusals/let_arrow
=== PAUSE TestKindPredicateWalkRefusals/let_arrow
=== RUN   TestKindPredicateWalkRefusals/parameter_with_a_default
=== PAUSE TestKindPredicateWalkRefusals/parameter_with_a_default
=== RUN   TestKindPredicateWalkRefusals/optional_parameter
=== PAUSE TestKindPredicateWalkRefusals/optional_parameter
=== RUN   TestKindPredicateWalkRefusals/two_statements
=== PAUSE TestKindPredicateWalkRefusals/two_statements
=== RUN   TestKindPredicateWalkRefusals/asserts
=== PAUSE TestKindPredicateWalkRefusals/asserts
=== RUN   TestKindPredicateWalkRefusals/switch_(true)
=== PAUSE TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateMatchesTheWrittenComparison
=== RUN   TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== RUN   TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== PAUSE TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestIndirectPredicateOverloadIsPending
--- PASS: TestIndirectPredicateOverloadIsPending (0.22s)
=== CONT  TestPredicateOverloadRuntime
=== RUN   TestPredicateOverloadRuntime/parser_every_result
--- PASS: TestPredicateOverloadWithoutPredicate (0.24s)
=== CONT  TestCheckedPredicateFalseReadDirections
=== RUN   TestCheckedPredicateFalseReadDirections/empty_branches
=== PAUSE TestCheckedPredicateFalseReadDirections/empty_branches
=== RUN   TestCheckedPredicateFalseReadDirections/semver's_toString
=== PAUSE TestCheckedPredicateFalseReadDirections/semver's_toString
=== RUN   TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== PAUSE TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== PAUSE TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== RUN   TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== PAUSE TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== RUN   TestCheckedPredicateFalseReadDirections/false_branch
=== PAUSE TestCheckedPredicateFalseReadDirections/false_branch
=== RUN   TestCheckedPredicateFalseReadDirections/else
=== PAUSE TestCheckedPredicateFalseReadDirections/else
=== RUN   TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== PAUSE TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== RUN   TestCheckedPredicateFalseReadDirections/capture
=== PAUSE TestCheckedPredicateFalseReadDirections/capture
=== RUN   TestCheckedPredicateFalseReadDirections/renarrowed
=== PAUSE TestCheckedPredicateFalseReadDirections/renarrowed
=== RUN   TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== PAUSE TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== RUN   TestCheckedPredicateFalseReadDirections/getter
=== PAUSE TestCheckedPredicateFalseReadDirections/getter
=== RUN   TestCheckedPredicateFalseReadDirections/spread
=== PAUSE TestCheckedPredicateFalseReadDirections/spread
--- PASS: TestPredicateOverloadLocalWrittenAfterTheCall (0.25s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered
=== RUN   TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== PAUSE TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== RUN   TestCheckedPredicateFalseReadDirections/conditional
=== PAUSE TestCheckedPredicateFalseReadDirections/conditional
=== RUN   TestCheckedPredicateFalseReadDirections/||_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/||_operand
=== RUN   TestCheckedPredicateFalseReadDirections/&&_operand
=== PAUSE TestCheckedPredicateFalseReadDirections/&&_operand
=== RUN   TestCheckedPredicateFalseReadDirections/const_alias
=== PAUSE TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestBrandedSupertypeIdentityNarrowingLowers
--- PASS: TestPredicateOverloadNamedCallback (0.73s)
=== CONT  TestPredicateOverloadLanguageLies
=== RUN   TestPredicateOverloadLanguageLies/field_reassigned
=== PAUSE TestPredicateOverloadLanguageLies/field_reassigned
=== RUN   TestPredicateOverloadLanguageLies/or_beside_the_call
=== PAUSE TestPredicateOverloadLanguageLies/or_beside_the_call
=== RUN   TestPredicateOverloadLanguageLies/lying_named_predicate
=== PAUSE TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadStaysOutOfA
--- PASS: TestPredicateOverloadInPlaceCallback (0.79s)
=== CONT  TestPredicateOverloadCallbackFromParameter
--- PASS: TestPredicateOverloadStaysOutOfA (0.17s)
=== CONT  TestPredicateOverloadReturnsAnotherRead
--- PASS: TestPredicateOverloadCallbackFromParameter (0.16s)
=== CONT  TestPredicateOverloadCallbackPushes
--- PASS: TestPredicateOverloadReturnsAnotherRead (0.10s)
=== CONT  TestKindPredicateCallIsItsComparison/true_and_false
--- PASS: TestBrandedSupertypeIdentityNarrowingLowers (0.82s)
=== CONT  TestKindPredicateThroughModules/import_as
=== CONT  TestKindPredicateCallIsItsComparison/walked_call_beside_filter
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== RUN   TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== PAUSE TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestKindPredicateCallIsItsComparison/const_arrow
--- PASS: TestPredicateOverloadCallbackPushes (0.48s)
=== CONT  TestKindPredicateCallIsItsComparison/!==_true
=== CONT  TestKindPredicateCallIsItsComparison/===_false
=== CONT  TestKindPredicateCallIsItsComparison/early_return
=== CONT  TestKindPredicateCallIsItsComparison/conditional
=== CONT  TestKindPredicateCallIsItsComparison/negated
=== RUN   TestPredicateOverloadRuntime/parser_some_result
=== CONT  TestKindPredicateWalkRefusals/comparison_control
=== CONT  TestKindPredicateThroughModules/namespace_import
=== CONT  TestKindPredicateThroughModules/re-export_as
=== CONT  TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag
--- PASS: TestKindPredicateCallIsItsComparison (0.00s)
    --- PASS: TestKindPredicateCallIsItsComparison/!==_true (21.23s)
    --- PASS: TestKindPredicateCallIsItsComparison/true_and_false (21.85s)
    --- PASS: TestKindPredicateCallIsItsComparison/const_arrow (21.45s)
    --- PASS: TestKindPredicateCallIsItsComparison/walked_call_beside_filter (21.82s)
    --- PASS: TestKindPredicateCallIsItsComparison/===_false (1.36s)
    --- PASS: TestKindPredicateCallIsItsComparison/early_return (1.32s)
    --- PASS: TestKindPredicateCallIsItsComparison/conditional (1.32s)
    --- PASS: TestKindPredicateCallIsItsComparison/negated (1.29s)
=== CONT  TestKindPredicateWalkRefusals/switch_(true)
=== CONT  TestKindPredicateWalkRefusals/asserts
--- PASS: TestKindPredicateThroughModules (0.00s)
    --- PASS: TestKindPredicateThroughModules/import_as (0.09s)
    --- PASS: TestKindPredicateThroughModules/namespace_import (0.16s)
    --- PASS: TestKindPredicateThroughModules/re-export_as (0.19s)
=== CONT  TestKindPredicateWalkRefusals/two_statements
=== RUN   TestPredicateOverloadRuntime/overload_some
=== CONT  TestKindPredicateWalkRefusals/optional_parameter
=== CONT  TestKindPredicateWalkRefusals/parameter_with_a_default
=== CONT  TestKindPredicateWalkRefusals/let_arrow
=== CONT  TestKindPredicateWalkRefusals/value_alias
=== CONT  TestKindPredicateWalkRefusals/method
=== CONT  TestKindPredicateWalkRefusals/overloaded
=== CONT  TestKindPredicateWalkRefusals/inferred_predicate
=== CONT  TestKindPredicateWalkRefusals/this_parameter
=== CONT  TestKindPredicateWalkRefusals/generic,_reduced
=== CONT  TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow
=== CONT  TestKindPredicateWalkRefusals/twin_layouts
=== CONT  TestKindPredicateWalkRefusals/lying_declared_type
=== CONT  TestKindPredicateWalkRefusals/shared_value_comparison
=== CONT  TestKindPredicateWalkRefusals/shared_value_call
=== CONT  TestKindPredicateWalkRefusals/wide_kind_call
=== CONT  TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body
=== CONT  TestKindPredicateMatchesTheWrittenComparison/const_enum_tag
=== CONT  TestCheckedPredicateFalseReadDirections/empty_branches
--- PASS: TestKindPredicateWalkRefusals (0.00s)
    --- PASS: TestKindPredicateWalkRefusals/comparison_control (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/asserts (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/switch_(true) (0.19s)
    --- PASS: TestKindPredicateWalkRefusals/optional_parameter (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/two_statements (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/parameter_with_a_default (0.09s)
    --- PASS: TestKindPredicateWalkRefusals/let_arrow (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/value_alias (0.11s)
    --- PASS: TestKindPredicateWalkRefusals/method (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/overloaded (0.12s)
    --- PASS: TestKindPredicateWalkRefusals/inferred_predicate (0.08s)
    --- PASS: TestKindPredicateWalkRefusals/this_parameter (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/generic,_reduced (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/twin_layouts (0.13s)
    --- PASS: TestKindPredicateWalkRefusals/annotated_const_wider_than_its_arrow (0.16s)
    --- PASS: TestKindPredicateWalkRefusals/lying_declared_type (0.15s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_call (0.10s)
    --- PASS: TestKindPredicateWalkRefusals/shared_value_comparison (0.14s)
    --- PASS: TestKindPredicateWalkRefusals/wide_kind_call (0.11s)
=== CONT  TestCheckedPredicateFalseReadDirections/renarrowed
=== CONT  TestCheckedPredicateFalseReadDirections/const_alias
=== CONT  TestCheckedPredicateFalseReadDirections/&&_operand
=== CONT  TestCheckedPredicateFalseReadDirections/||_operand
=== CONT  TestCheckedPredicateFalseReadDirections/conditional
=== CONT  TestCheckedPredicateFalseReadDirections/the_same_predicate_again
=== CONT  TestCheckedPredicateFalseReadDirections/spread
=== CONT  TestCheckedPredicateFalseReadDirections/getter
=== CONT  TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later
=== CONT  TestCheckedPredicateFalseReadDirections/false_branch
=== CONT  TestCheckedPredicateFalseReadDirections/capture
=== CONT  TestCheckedPredicateFalseReadDirections/after_an_early_exit
=== CONT  TestCheckedPredicateFalseReadDirections/else
=== CONT  TestCheckedPredicateFalseReadDirections/replaced_before_the_read
=== CONT  TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference
=== CONT  TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge
=== RUN   TestPredicateOverloadRuntime/overload_some_empty
=== CONT  TestCheckedPredicateFalseReadDirections/semver's_toString
=== CONT  TestPredicateOverloadLanguageLies/field_reassigned
=== CONT  TestPredicateOverloadLanguageLies/lying_named_predicate
=== CONT  TestPredicateOverloadLanguageLies/or_beside_the_call
--- PASS: TestCheckedPredicateFalseReadDirections (0.00s)
    --- PASS: TestCheckedPredicateFalseReadDirections/empty_branches (0.06s)
    --- PASS: TestCheckedPredicateFalseReadDirections/renarrowed (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/const_alias (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/&&_operand (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/conditional (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/||_operand (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/the_same_predicate_again (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/spread (0.10s)
    --- PASS: TestCheckedPredicateFalseReadDirections/getter (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture_of_a_let_assigned_later (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/false_branch (0.14s)
    --- PASS: TestCheckedPredicateFalseReadDirections/capture (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/after_an_early_exit (0.12s)
    --- PASS: TestCheckedPredicateFalseReadDirections/else (0.09s)
    --- PASS: TestCheckedPredicateFalseReadDirections/replaced_before_the_read (0.13s)
    --- PASS: TestCheckedPredicateFalseReadDirections/a_property_named_as_the_reference (0.11s)
    --- PASS: TestCheckedPredicateFalseReadDirections/semver's_toString (0.08s)
    --- PASS: TestCheckedPredicateFalseReadDirections/read_after_the_sides_merge (0.12s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims
--- PASS: TestKindPredicateMatchesTheWrittenComparison (0.00s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/union-of-literals_tag (0.18s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/declared_type_narrower_than_the_body (0.11s)
    --- PASS: TestKindPredicateMatchesTheWrittenComparison/const_enum_tag (0.97s)
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance
=== CONT  TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim
--- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered (1.18s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/second_instance (0.10s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/closure_after_the_call (0.08s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/generic_arrow (0.09s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2_over_unbranded_claims (0.14s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_honest_instance (0.13s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/orPredicate2's_lying_instance (0.14s)
    --- PASS: TestBrandedSupertypeIsJudgedWhereItIsLowered/read_from_its_claim (0.07s)
--- PASS: TestPredicateOverloadLanguageLies (0.01s)
    --- PASS: TestPredicateOverloadLanguageLies/field_reassigned (0.08s)
    --- PASS: TestPredicateOverloadLanguageLies/or_beside_the_call (0.09s)
    --- PASS: TestPredicateOverloadLanguageLies/lying_named_predicate (0.63s)
=== RUN   TestPredicateOverloadRuntime/overload_some_false_valid
=== RUN   TestPredicateOverloadRuntime/overload_some_false_read
=== RUN   TestPredicateOverloadRuntime/overload_some_objects
=== RUN   TestPredicateOverloadRuntime/overload_some_true_only
=== RUN   TestPredicateOverloadRuntime/overload_callback
=== RUN   TestPredicateOverloadRuntime/overload_erased
=== RUN   TestPredicateOverloadRuntime/overload_checked
=== RUN   TestPredicateOverloadRuntime/overload_false
=== RUN   TestPredicateOverloadRuntime/overload_once
=== RUN   TestPredicateOverloadRuntime/overload_every
=== RUN   TestPredicateOverloadRuntime/overload_array_alias
=== RUN   TestPredicateOverloadRuntime/overload_assertion
=== RUN   TestPredicateOverloadRuntime/overload_nominal
--- PASS: TestPredicateOverloadRuntime (35.27s)
    --- PASS: TestPredicateOverloadRuntime/parser_every_result (23.04s)
    --- PASS: TestPredicateOverloadRuntime/parser_some_result (1.10s)
    --- PASS: TestPredicateOverloadRuntime/overload_some (1.18s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_empty (0.89s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_valid (0.61s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_false_read (0.78s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_objects (0.70s)
    --- PASS: TestPredicateOverloadRuntime/overload_some_true_only (0.74s)
    --- PASS: TestPredicateOverloadRuntime/overload_callback (0.72s)
    --- PASS: TestPredicateOverloadRuntime/overload_erased (0.64s)
    --- PASS: TestPredicateOverloadRuntime/overload_checked (0.59s)
    --- PASS: TestPredicateOverloadRuntime/overload_false (0.70s)
    --- PASS: TestPredicateOverloadRuntime/overload_once (0.65s)
    --- PASS: TestPredicateOverloadRuntime/overload_every (0.80s)
    --- PASS: TestPredicateOverloadRuntime/overload_array_alias (0.77s)
    --- PASS: TestPredicateOverloadRuntime/overload_assertion (0.68s)
    --- PASS: TestPredicateOverloadRuntime/overload_nominal (0.68s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	35.655s
```

stderr (verbatim):

```text
```

Step 4b — uncached selected oracle tests: result events {'PASS': 35, 'FAIL': 0, 'SKIP': 0}; subtest result events {'PASS': 31, 'FAIL': 0, 'SKIP': 0}. Every per-subtest PASS/FAIL is retained below. Complete output also preserves every failure's whole got/want diagnostics without shortening.

Environment: `ADAMIC_GATE_UNCACHED=1`, plus the exported setup above.

**Step 4b — uncached selected oracle tests**

Working directory: `/workspace/adamic`.

```sh
go test ./internal/oracle -run 'TestPredicateFalseRead|TestCheckedPredicateOracle|TestCheckedPredicateCountsAreRecorded|TestCountsAreRecordedTable' -count=1 -v -timeout 40m
```

Exit: 0.

stdout (verbatim):

```text
=== RUN   TestCountsAreRecordedTable
=== PAUSE TestCountsAreRecordedTable
=== RUN   TestPredicateFalseRead
=== PAUSE TestPredicateFalseRead
=== RUN   TestCheckedPredicateOracle
=== PAUSE TestCheckedPredicateOracle
=== RUN   TestCheckedPredicateCountsAreRecorded
=== PAUSE TestCheckedPredicateCountsAreRecorded
=== CONT  TestCountsAreRecordedTable
=== CONT  TestCheckedPredicateOracle
=== CONT  TestPredicateFalseRead
=== CONT  TestCheckedPredicateCountsAreRecorded
=== RUN   TestCheckedPredicateOracle/true
=== RUN   TestPredicateFalseRead/no_read
=== PAUSE TestPredicateFalseRead/no_read
=== RUN   TestPredicateFalseRead/false_read
=== PAUSE TestPredicateFalseRead/false_read
=== RUN   TestPredicateFalseRead/else_read
=== PAUSE TestPredicateFalseRead/else_read
=== RUN   TestPredicateFalseRead/early_exit
=== PAUSE TestPredicateFalseRead/early_exit
=== RUN   TestPredicateFalseRead/capture
=== PAUSE TestPredicateFalseRead/capture
=== RUN   TestPredicateFalseRead/renarrowed
=== PAUSE TestPredicateFalseRead/renarrowed
=== RUN   TestPredicateFalseRead/getter
=== PAUSE TestPredicateFalseRead/getter
=== RUN   TestPredicateFalseRead/getter_capture
=== PAUSE TestPredicateFalseRead/getter_capture
=== RUN   TestPredicateFalseRead/spread
=== PAUSE TestPredicateFalseRead/spread
=== RUN   TestPredicateFalseRead/again
=== PAUSE TestPredicateFalseRead/again
=== RUN   TestPredicateFalseRead/shorthand
=== PAUSE TestPredicateFalseRead/shorthand
=== RUN   TestPredicateFalseRead/logical_assignment
=== PAUSE TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/no_read
=== CONT  TestPredicateFalseRead/getter
--- PASS: TestCountsAreRecordedTable (0.00s)
=== CONT  TestPredicateFalseRead/logical_assignment
=== CONT  TestPredicateFalseRead/shorthand
--- PASS: TestCheckedPredicateCountsAreRecorded (50.73s)
=== CONT  TestPredicateFalseRead/again
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 231.64
=== NAME  TestPredicateFalseRead/logical_assignment
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d miss 238.18
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 238.19
=== NAME  TestPredicateFalseRead/shorthand
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 238.18
=== NAME  TestPredicateFalseRead/no_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 238.19
=== CONT  TestPredicateFalseRead/spread
=== CONT  TestPredicateFalseRead/getter_capture
=== CONT  TestPredicateFalseRead/early_exit
=== CONT  TestPredicateFalseRead/capture
=== NAME  TestCheckedPredicateOracle/true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.07
    predicates_checked_test.go:166: true lying body exits 70; removed-true-check mutant executes Node result "claimed\n" and is caught
=== RUN   TestCheckedPredicateOracle/true_unused
=== NAME  TestPredicateFalseRead/getter_capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.02
=== CONT  TestPredicateFalseRead/else_read
=== NAME  TestPredicateFalseRead/spread
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.04
=== CONT  TestPredicateFalseRead/false_read
=== NAME  TestPredicateFalseRead/capture
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.08
=== NAME  TestPredicateFalseRead/early_exit
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.05
=== CONT  TestPredicateFalseRead/renarrowed
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.07
=== NAME  TestPredicateFalseRead/false_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.02
=== NAME  TestCheckedPredicateOracle/true_unused
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    predicates_checked_test.go:166: true_unused lying body exits 70; removed-true-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false
=== NAME  TestPredicateFalseRead/else_read
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
=== NAME  TestPredicateFalseRead/renarrowed
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
--- PASS: TestPredicateFalseRead (0.00s)
    --- PASS: TestPredicateFalseRead/getter (44.31s)
    --- PASS: TestPredicateFalseRead/shorthand (239.62s)
    --- PASS: TestPredicateFalseRead/logical_assignment (283.96s)
    --- PASS: TestPredicateFalseRead/again (233.25s)
    --- PASS: TestPredicateFalseRead/no_read (284.21s)
    --- PASS: TestPredicateFalseRead/getter_capture (2.59s)
    --- PASS: TestPredicateFalseRead/spread (2.79s)
    --- PASS: TestPredicateFalseRead/capture (3.76s)
    --- PASS: TestPredicateFalseRead/early_exit (4.60s)
    --- PASS: TestPredicateFalseRead/false_read (2.98s)
    --- PASS: TestPredicateFalseRead/else_read (3.50s)
    --- PASS: TestPredicateFalseRead/renarrowed (2.20s)
=== NAME  TestCheckedPredicateOracle/false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.06
=== RUN   TestCheckedPredicateOracle/false_alias
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
=== RUN   TestCheckedPredicateOracle/false_compound_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    predicates_checked_test.go:166: asserts lying body exits 70; removed-asserts-check mutant executes Node result "called\nasserted\n" and is caught
=== RUN   TestCheckedPredicateOracle/valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
=== RUN   TestCheckedPredicateOracle/asserts_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_valid
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    predicates_checked_test.go:166: condition_lie lying body exits 70; removed-asserts-check mutant executes Node result "continued\n" and is caught
=== RUN   TestCheckedPredicateOracle/condition_once
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
=== RUN   TestCheckedPredicateOracle/condition_once_lie
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    predicates_checked_test.go:166: condition_once_lie lying body exits 70; removed-asserts-check mutant executes Node result "argument\ncontinued\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_unchanged
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
=== RUN   TestCheckedPredicateOracle/assert_not_node
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    predicates_checked_test.go:166: assert_not_node lying body exits 70; removed-asserts-check mutant executes Node result "asserting\nnarrowed\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_switch_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    predicates_checked_test.go:166: false_switch_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_eq_false
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    predicates_checked_test.go:166: false_eq_false lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_ne_true
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.03
    predicates_checked_test.go:166: false_ne_true lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_false_eq
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    predicates_checked_test.go:166: false_false_eq lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
=== RUN   TestCheckedPredicateOracle/false_comma
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.01
    oracle_test.go:645: build go_build_./oracle/json_types.go_json-types 148ad9f48b8d hit 0.02
    predicates_checked_test.go:166: false_comma lying body exits 70; removed-false-check mutant executes Node result "undefined\n" and is caught
--- PASS: TestCheckedPredicateOracle (310.68s)
    --- PASS: TestCheckedPredicateOracle/true (285.68s)
    --- PASS: TestCheckedPredicateOracle/true_unused (4.30s)
    --- PASS: TestCheckedPredicateOracle/false (1.28s)
    --- PASS: TestCheckedPredicateOracle/false_alias (1.55s)
    --- PASS: TestCheckedPredicateOracle/false_compound_unchanged (1.19s)
    --- PASS: TestCheckedPredicateOracle/asserts (1.10s)
    --- PASS: TestCheckedPredicateOracle/valid (1.00s)
    --- PASS: TestCheckedPredicateOracle/asserts_valid (0.94s)
    --- PASS: TestCheckedPredicateOracle/condition_valid (1.02s)
    --- PASS: TestCheckedPredicateOracle/condition_lie (1.13s)
    --- PASS: TestCheckedPredicateOracle/condition_once (1.26s)
    --- PASS: TestCheckedPredicateOracle/condition_once_lie (1.80s)
    --- PASS: TestCheckedPredicateOracle/false_unchanged (0.92s)
    --- PASS: TestCheckedPredicateOracle/assert_not_node (1.31s)
    --- PASS: TestCheckedPredicateOracle/false_switch_true (1.19s)
    --- PASS: TestCheckedPredicateOracle/false_eq_false (1.21s)
    --- PASS: TestCheckedPredicateOracle/false_ne_true (1.38s)
    --- PASS: TestCheckedPredicateOracle/false_false_eq (1.13s)
    --- PASS: TestCheckedPredicateOracle/false_comma (1.27s)
PASS
gate cache: native hits=0 misses=19
gate cache: node hits=0 misses=72
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	310.786s
```

stderr (verbatim):

```text
```

**Restoration and report-only publication**

Restored internal/lower/predicates_proof.go byte for byte to B.

`git diff --exit-code`: exit 0; stdout and stderr empty.

`git status --porcelain`: exit 0; stdout and stderr empty (clean).

These clean checks ran after all measurements and before creating the report file.

The report branch is `proof/lower-39w-6-1-report`, created from B after restoring the variant. Only `proof/lower-39w-6-1/report.md` is included in the report commit. The push is restricted to `HEAD:refs/heads/proof/lower-39w-6-1-report`; no other branch is pushed.

**Exact throwaway Go test used in all three overlay measurements**

SHA-256: `e69cbf842b43453b02221060a9c80a166046395abe4e869bdf90dacf4bdcea52`.

```go
package lower

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

const measure39w6Some = `function some<T>(array: readonly T[] | undefined): array is readonly T[];
function some<T>(array: readonly T[] | undefined, predicate: (value: T) => boolean): boolean;
function some<T>(array: readonly T[] | undefined, predicate?: (value: T) => boolean): boolean { if (array === undefined) return false; if (predicate === undefined) return array.length > 0; for (const value of array) { if (predicate(value)) return true; } return false; }
function load(): readonly string[] | undefined { return []; }
`

// TestMeasure39w6 logs, for each probe, Lower's answer and, for the first some call in the probe's own file,
// every identifier spelled like the argument: where it stands, the position tests a read filter could use,
// whether the checker matches it, and what the false-read walk answers from it, plus the checker's type of
// each export default expression and of the importer's read.
func TestMeasure39w6(t *testing.T) {
	for _, probe := range []struct {
		name  string
		files map[string]string
		entry string
		site  string
	}{
		{"export_default", map[string]string{
			"items.ts": measure39w6Some + "const items = load();\nif (some(items)) throw new Error(\"x\");\nexport default items;\n",
			"main.ts":  "import items from \"./items.ts\";\nconsole.log(typeof items);\n",
		}, "main.ts", "items.ts"},
		{"export_named", map[string]string{
			"main.ts": measure39w6Some + "const items = load();\nif (some(items)) throw new Error(\"x\");\nexport { items };\nconsole.log(\"named\");\n",
		}, "main.ts", "main.ts"},
		{"label", map[string]string{
			"main.ts": measure39w6Some + "function use(items: readonly string[] | undefined): void { if (!some(items)) { items: for (;;) { break items; } } console.log(\"label\"); }\nuse([]);\n",
		}, "main.ts", "main.ts"},
	} {
		directory := t.TempDir()
		for name, source := range probe.files {
			if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
		program, err := load.Load([]string{filepath.Join(directory, probe.entry)})
		if err != nil {
			t.Logf("%s: load: %v", probe.name, err)
			continue
		}
		lowered, lowerErr := Lower(context.Background(), program)
		if lowerErr != nil {
			t.Logf("%s: Lower error: %v", probe.name, lowerErr)
		} else {
			t.Logf("%s: Lower ok, predicate sites %+v", probe.name, lowered.PredicateChecks.Sites)
		}
		for _, file := range program.Files() {
			if filepath.Base(program.FileName(file)) != probe.site && filepath.Base(program.FileName(file)) != probe.entry {
				continue
			}
			checked, release := program.Checker(context.Background(), file)
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			where := func(node *ast.Node) string {
				line, column := scanner.GetLineAndCharacterOfPosition(file, scanner.GetTokenPosOfNode(node, file, false))
				return filepath.Base(program.FileName(file)) + ":" + strconv.Itoa(line+1) + ":" + strconv.Itoa(column+1)
			}
			var exports ast.Visitor
			exports = func(node *ast.Node) bool {
				if node.Kind == ast.KindExportAssignment {
					expression := node.AsExportAssignment().Expression
					t.Logf("%s: export assignment expression at %s typed %s", probe.name, where(expression), l.checker.TypeToString(l.checker.GetTypeAtLocation(expression)))
				}
				if node.Kind == ast.KindCallExpression && node.Expression().Kind == ast.KindIdentifier && node.Expression().Text() == "log" {
					for _, argument := range node.Arguments() {
						t.Logf("%s: log argument at %s typed %s", probe.name, where(argument), l.checker.TypeToString(l.checker.GetTypeAtLocation(argument)))
					}
				}
				if node.Kind == ast.KindTypeOfExpression {
					operand := node.AsTypeOfExpression().Expression
					t.Logf("%s: typeof operand at %s typed %s", probe.name, where(operand), l.checker.TypeToString(l.checker.GetTypeAtLocation(operand)))
				}
				node.ForEachChild(exports)
				return false
			}
			file.AsNode().ForEachChild(exports)
			if filepath.Base(program.FileName(file)) != probe.site {
				release()
				continue
			}
			var call *ast.CallExpression
			var find ast.Visitor
			find = func(node *ast.Node) bool {
				if call == nil && node.Kind == ast.KindCallExpression && node.Expression().Kind == ast.KindIdentifier && node.Expression().Text() == "some" {
					call = node.AsCallExpression()
				}
				node.ForEachChild(find)
				return false
			}
			file.AsNode().ForEachChild(find)
			if call == nil {
				t.Logf("%s: no some call", probe.name)
				release()
				continue
			}
			claim := predicateOfSignature(l.checker, l.checker.GetResolvedSignature(call.AsNode()))
			argument := ast.SkipParentheses(call.Arguments.Nodes[claim.ParameterIndex()])
			t.Logf("%s: call at %s argument at %s", probe.name, where(call.AsNode()), where(argument))
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindIdentifier && node.Text() == argument.Text() {
					data := node.FlowNodeData()
					flowNil, reaches := "n/a", "-"
					if data != nil {
						flowNil = measure39w6Yes(data.FlowNode == nil)
						if data.FlowNode != nil {
							reaches = measure39w6Yes(l.predicateFalseReaches(data.FlowNode, call.AsNode(), argument))
						}
					}
					t.Logf("  identifier at %s parent %s: IsExpressionNode %s, IsRightSideOfPropertyAccess %s, IsDeclarationName %s, IsPartOfTypeNode %s, predicateWriteOnly %s, checker matches %s, predicateNameReads %s, FlowNode-nil %s, reaches %s",
						where(node), node.Parent.Kind.String(), measure39w6Yes(ast.IsExpressionNode(node)), measure39w6Yes(ast.IsRightSideOfPropertyAccess(node)), measure39w6Yes(ast.IsDeclarationName(node)), measure39w6Yes(ast.IsPartOfTypeNode(node)), measure39w6Yes(predicateWriteOnly(node)), measure39w6Yes(l.predicateSameReference(node, argument)), measure39w6Yes(l.predicateNameReads(node, argument)), flowNil, reaches)
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			t.Logf("  predicateUseDirections false %s, predicateFalseRead %s, checkedPredicateDirections %d", measure39w6Yes(l.predicateUseDirections(call)&predicateFalse != 0), measure39w6Yes(l.predicateFalseRead(call, argument)), l.checkedPredicateDirections(call, claim))
			release()
		}
	}
}

func measure39w6Yes(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
```

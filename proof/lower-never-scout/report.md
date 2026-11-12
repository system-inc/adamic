# Lowering `never`-read scout report

Tasks #p6vawef and #z5d0jye. Report date: 2026-11-12 UTC.

The completed full retry with a memory cap encountered **10 reads and 88 never-typed calls** in tsc. The oracle corpus union encountered **19 reads and 27 never-typed calls**.

The memory-capped full run completed normally; its counts are the full-mode headline.

These are unique source sites encountered on the requested lowering paths, not AST-wide totals, execution frequencies, or a benchmark of generated checks. An encounter can precede a later refusal. Calls are counted separately and are not added to the read-trap budget.

By the requested enum-identity proxy, **9 tsc reads and 9 corpus reads lack today’s enum trap**. At least one additional tsc read—the lazy assignment-cache comparison—can be live in ordinary checking without a false type promise. These are site counts, not measured runtime overhead.

## Pins and unchanged inputs

- Adamic: `0c7052759671347775615af5b4a8028dd8688af8`, detached after `git fetch origin main`; recursive submodules initialized.
- cohere: `5360753e19e4417675a90e92a1823bea66b13918`.
- checker TypeScript-Go: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Upstream source: `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Go 1.27.1; Node v24.19.0; external lockfile-pinned TypeScript 6.0.3.

The 79 adapted compiler inputs were hashed in `compiler-input-hashes.json`. All recorded census headers report `checker_rejected: false` with no diagnostics; the tool’s fixed “measured on a checker-rejected program” label is not the actual checker status of this input. Both completed streams (first-error and memory-capped full) have 79 file records plus their header and relation-report trailer. All 79 compiler-input hashes still matched after the full retry.

## Counts by class

Classes are exclusive: (a) a read contributing to an argument of a resolved never-returning call; (b) an exhaustiveness binding/return; (c) other narrowing of a non-never declaration; (d) everything else. Class (a) takes precedence over an overlapping switch-default description. Class (c) includes assignment-flow narrowing, including the Boolean `null!` case below.

“Already trapped” is exactly the requested `enumNeverIdentity(node, map[*ast.Node]bool{}) != nil` metric. Function counts distinguish file and enclosing function start position. `<top>` is not counted as a function; its scopes and site counts are shown separately.

### tsc — completed first-error mode

| Class | Reads | Functions | Top-level scopes | Top-level reads | Already trapped |
|---|---:|---:|---:|---:|---:|
| (a) | 1 | 1 | 0 | 0 | 1 |
| (b) | 0 | 0 | 0 | 0 | 0 |
| (c) | 0 | 0 | 0 | 0 | 0 |
| (d) | 0 | 0 | 0 | 0 | 0 |

Total reads: **1**. Never-typed call count: **13**.

### tsc — initial full attempt, incomplete (SIGKILL)

| Class | Reads | Functions | Top-level scopes | Top-level reads | Already trapped |
|---|---:|---:|---:|---:|---:|
| (a) | 1 | 1 | 0 | 0 | 1 |
| (b) | 0 | 0 | 0 | 0 | 0 |
| (c) | 2 | 2 | 0 | 0 | 0 |
| (d) | 0 | 0 | 0 | 0 | 0 |

Total reads: **3**. Never-typed call count: **53**.

### tsc — full retry, complete

| Class | Reads | Functions | Top-level scopes | Top-level reads | Already trapped |
|---|---:|---:|---:|---:|---:|
| (a) | 6 | 6 | 0 | 0 | 1 |
| (b) | 0 | 0 | 0 | 0 | 0 |
| (c) | 4 | 4 | 0 | 0 | 0 |
| (d) | 0 | 0 | 0 | 0 | 0 |

Total reads: **10**. Never-typed call count: **88**.

### Oracle corpus — deduplicated union

| Class | Reads | Functions | Top-level scopes | Top-level reads | Already trapped |
|---|---:|---:|---:|---:|---:|
| (a) | 0 | 0 | 0 | 0 | 0 |
| (b) | 10 | 7 | 0 | 0 | 10 |
| (c) | 9 | 1 | 3 | 8 | 0 |
| (d) | 0 | 0 | 0 | 0 | 0 |

Total reads: **19**. Never-typed call count: **27**.

### Corpus coverage and deduplication

| Stream | Raw rows | Unique reads | Unique calls |
|---|---:|---:|---:|
| corpus_suite | 50 | 19 | 27 |
| corpus_sweep | 45 | 18 | 27 |
| corpus_union | 95 | 19 | 27 |

The three isolated host-row probes logged zero sites. The union uses `(file, pos, end)`, matching the hook’s identity but extending deduplication across processes/runs. Duplicate descriptors agree. No recorded never-read or never-call site had a temporary filename; inline tests were not excluded, and the twelve formerly blocked inline assignment-proof tests were rerun. Different temporary filenames would remain distinct sites and would be labelled by test name, or “inline” if unknown.

All observed sites entered through route `value`; the other route probes added no distinct observed site. This does not mean those routes cannot read references.

## Source classification and reachability

Reachability below is a source-based judgement, not dynamic hit profiling. “tsc self-compilation” refers to executing the compiler on its own sources; compiling a corpus fixture does not execute that fixture inside tsc. Corpus runtime judgements therefore describe running the fixture itself.

### Classes (a) and (b): counts and examples

The tsc `moduleResolution` default requires a false exhaustive-domain promise (an invalid/out-of-domain compiler option). Ordinary self-compilation with valid options does not reach it. `Debug.fail` has a sole signature returning `never` (`debug.ts:174`); interpolation into its message is an argument use.

The classFields and esDecorators `node` fallbacks likewise require a false asserted assignment-element domain: all three allowed element kinds return before it. Their `visitEachChild<T>(node:T,...):T` signature resolves to `never` through T=never; the helper can return normally on an off-contract AST and is not intrinsically a runtime stop. Ordinary self-compilation has valid assignment elements.

The ES2015 `insertCaptureNewTargetIfNeeded` switch covers all FunctionLikeDeclaration kinds; its Debug.failBadSyntaxKind argument requires a wrong-kind AST that violates that declared domain and is not reached by ordinary self-compilation. Both JSX failure arguments also follow complete handling of their JsxChild or JsxAttributeValue domains; wrong-kind ASTs are required, and the compiler’s own .ts sources do not contain JSX.

Corpus enum defaults likewise require the exhaustive-member promise to be false. Several fixtures deliberately supply number 42 through an open numeric-enum input; others call only real members and never enter their defaults. The current binding trap stops execution before a later alias return can run.

**tsc, class (a)**: 6 distinct observed sites across the listed streams.

- `/tmp/never-scout/adapted/src/compiler/moduleNameResolver.ts:1456:67`: `moduleResolution`; enclosing function `resolveModuleName`; already trapped: yes.
- `/tmp/never-scout/adapted/src/compiler/transformers/classFields.ts:3144:31`: `node`; enclosing function `visitObjectAssignmentElement`; already trapped: no.
- `/tmp/never-scout/adapted/src/compiler/transformers/es2015.ts:2225:52`: `node`; enclosing function `insertCaptureNewTargetIfNeeded`; already trapped: no.

**tsc, class (b)**: 0 distinct observed sites across the listed streams.

No observed example exists for this class; three examples cannot be supplied without inventing sites.

**corpus, class (a)**: 0 distinct observed sites across the listed streams.

No observed example exists for this class; three examples cannot be supplied without inventing sites.

**corpus, class (b)**: 10 distinct observed sites across the listed streams.

- `/workspace/adamic/internal/oracle/testdata/enums_flags_never_default.a:8:40`: `value`; enclosing function `describe`; already trapped: yes.
- `/workspace/adamic/internal/oracle/testdata/enums_open_members.a:6:47`: `value`; enclosing function `describe`; already trapped: yes.
- `/workspace/adamic/internal/oracle/testdata/enums_open_never.a:7:40`: `value`; enclosing function `describe`; already trapped: yes.

### Every class-(c) and class-(d) site

Raw checker filenames add `.ts` to `.a` inputs. Locations below map those back to their physical `.a` files; positions and deduplication retain the checker’s original `(file, pos, end)`.

**tsc**

`/tmp/never-scout/adapted/src/compiler/checker.ts:30272:86` — class (c). Node `symbol.lastAssignmentPos`; function `isSymbolAssignedDefinitely`; declared type `number | undefined`; flow type `never`; enum identity absent.

Why never: The early lastAssignmentPos !== undefined branch returns, leaving the property narrowed to undefined. isSymbolAssigned calls isPastLastAssignment, which calls markNodeAssignments and can populate that same property. The checker retains the earlier property narrowing through the helper call; the subsequent !== undefined success edge therefore makes the final property read never.

Reachability: Reachable during ordinary checking, including self-checking inputs with mutable local variables assigned after declaration; no dishonest predicate or fabricated value is required. Static path: checkIdentifier -> isSymbolAssignedDefinitely -> isSymbolAssigned -> isPastLastAssignment -> markNodeAssignments. markNodeAssignments writes the cached number at checker.ts:30336 (and 30351). This measurement does not claim dynamic execution coverage.

`/tmp/never-scout/adapted/src/compiler/moduleNameResolver.ts:450:154` — class (c). Node `bestVersionPaths`; function `readPackageJsonTypesVersionPaths`; declared type `MapLike<string[]>`; flow type `never`; enum identity absent.

Why never: bestVersionPaths is declared MapLike<string[]> by VersionPaths.paths. The branch typeof bestVersionPaths !== "object" excludes its entire declared object type, so the subsequent typeof operand is never. This is defensive validation of a selected typesVersions entry from parsed package.json.

Reachability: Not reached in an ordinary self-compilation using valid pinned package metadata. It can be reached with tracing enabled (traceResolution/state.traceEnabled) and a malformed selected typesVersions mapping, such as a string instead of an object. The unchecked parsed-JSON boundary then violates the declared MapLike<string[]> promise; this read exists to diagnose that malformed input. This is a source-based reachability judgment, not dynamic execution coverage.

`/tmp/never-scout/adapted/src/compiler/sys.ts:1841:64` — class (c). Node `dirent`; function `getAccessibleFileSystemEntries`; declared type `Dirent<string>`; flow type `never`; enum identity absent.

Why never: The true arm of typeof dirent === "string" excludes the entire declared Dirent<string> object type. This is the retained pre-Node-10.10 readdirSync compatibility branch; withFileTypes:true is typed as returning Dirent objects.

Reachability: Not reached during ordinary self-compilation on the provided Node v24.19.0 with withFileTypes:true. A pre-10.10 host that ignores that option and returns strings, or an off-contract host shim, can reach it; that host behavior violates the modern Dirent declaration. This is the full compiler counterpart of the corpus filesystem fragment, not a new normal modern-host path.

`/tmp/never-scout/adapted/src/compiler/transformers/module/esnextAnd2015.ts:111:121` — class (c). Node `importRequireStatements`; function `transformSourceFile`; declared type `[ImportDeclaration, VariableStatement] | undefined`; flow type `never`; enum identity absent.

Why never: transformSourceFile explicitly assigns the captured importRequireStatements variable undefined at line 97. updateExternalModule invokes visitor, which can process an ImportEqualsDeclaration and call createRequireCall; that helper assigns the tuple at line 248. The checker retains the undefined narrowing across this indirect mutation, so the truthy guarded argument read is never despite its declared tuple-or-undefined type.

Reachability: Reachable for valid Node-style ESM inputs with a value import x = require("module") (Node16/NodeNext-family module emission), with no false type promise. The visitor synthesizes createRequire support and then inserts the accumulated statements here. Not required by ordinary self-compilation of the measured src/compiler sources: they have no external import-equals declaration; a plain const inspector = require("inspector") in sys.ts is not that syntax and does not call createRequireCall. This is a source-based judgement, not a dynamic execution trace.

No class-(d) site was observed.

**corpus**

`/workspace/adamic/internal/oracle/testdata/census_boolean_dead_branch.a:5:7` — class (c). Node `flag`; function `<top>`; declared type `boolean`; flow type `never`; enum identity absent.

Why never: The enclosing guard contradicts the known literal value of flag: flag was assigned false and other was assigned true. The nested boolean-expression read is consequently never, despite declared boolean storage.

Reachability: Not reached by this fixture. Its outer if(flag) and if(!other) guards are false, and no intervening mutation can change either value. These are deliberate dead-branch lowering probes, not tsc self-compilation runtime sites.

`/workspace/adamic/internal/oracle/testdata/census_boolean_dead_branch.a:5:15` — class (c). Node `flag`; function `<top>`; declared type `boolean`; flow type `never`; enum identity absent.

Why never: The enclosing guard contradicts the known literal value of flag: flag was assigned false and other was assigned true. The nested boolean-expression read is consequently never, despite declared boolean storage.

Reachability: Not reached by this fixture. Its outer if(flag) and if(!other) guards are false, and no intervening mutation can change either value. These are deliberate dead-branch lowering probes, not tsc self-compilation runtime sites.

`/workspace/adamic/internal/oracle/testdata/census_boolean_dead_branch.a:5:24` — class (c). Node `flag`; function `<top>`; declared type `boolean`; flow type `never`; enum identity absent.

Why never: The enclosing guard contradicts the known literal value of flag: flag was assigned false and other was assigned true. The nested boolean-expression read is consequently never, despite declared boolean storage.

Reachability: Not reached by this fixture. Its outer if(flag) and if(!other) guards are false, and no intervening mutation can change either value. These are deliberate dead-branch lowering probes, not tsc self-compilation runtime sites.

`/workspace/adamic/internal/oracle/testdata/census_boolean_dead_branch.a:11:22` — class (c). Node `other`; function `<top>`; declared type `boolean`; flow type `never`; enum identity absent.

Why never: The enclosing guard contradicts the known literal value of other: flag was assigned false and other was assigned true. The nested boolean-expression read is consequently never, despite declared boolean storage.

Reachability: Not reached by this fixture. Its outer if(flag) and if(!other) guards are false, and no intervening mutation can change either value. These are deliberate dead-branch lowering probes, not tsc self-compilation runtime sites.

`/workspace/adamic/internal/oracle/testdata/census_boolean_dead_branch.a:11:31` — class (c). Node `other`; function `<top>`; declared type `boolean`; flow type `never`; enum identity absent.

Why never: The enclosing guard contradicts the known literal value of other: flag was assigned false and other was assigned true. The nested boolean-expression read is consequently never, despite declared boolean storage.

Reachability: Not reached by this fixture. Its outer if(flag) and if(!other) guards are false, and no intervening mutation can change either value. These are deliberate dead-branch lowering probes, not tsc self-compilation runtime sites.

`/workspace/adamic/internal/oracle/testdata/census_boolean_dead_branch.a:11:41` — class (c). Node `other`; function `<top>`; declared type `boolean`; flow type `never`; enum identity absent.

Why never: The enclosing guard contradicts the known literal value of other: flag was assigned false and other was assigned true. The nested boolean-expression read is consequently never, despite declared boolean storage.

Reachability: Not reached by this fixture. Its outer if(flag) and if(!other) guards are false, and no intervening mutation can change either value. These are deliberate dead-branch lowering probes, not tsc self-compilation runtime sites.

`/workspace/adamic/internal/oracle/testdata/host_rest_union.a:17:48` — class (c). Node `dirent`; function `<top>`; declared type `{ name: string; }`; flow type `never`; enum identity absent.

Why never: typeof dirent === "string" excludes the declared { name: string } object type; the ternary true-arm read is therefore never.

Reachability: Not reached by this fixture: entries contains two object literals with name fields. A string introduced through a false type promise would be required. It is a fixture site, not a tsc self-compilation runtime site.

`/workspace/adamic/internal/oracle/testdata/placeholder_nonnull_null_coalesce.a:12:16` — class (c). Node `flag`; function `<top>`; declared type `boolean`; flow type `never`; enum identity absent.

Why never: Assignment-flow narrowing reduces the boolean union against the never type of the null! initializer. This is the other-narrowing part of class (c), rather than a predicate false side; number and string slots in the same fixture do not produce the same never read.

Reachability: Reached at top level by this fixture: null! supplies the deliberately unset/null placeholder and flag ?? true probes it before flag=false. The non-null assertion makes the declared boolean promise false at that point. This is not a tsc self-compilation runtime site.

`/workspace/adamic/stage3/fixtures/host/08_getDirectories.a:192:56` — class (c). Node `dirent`; function `getAccessibleFileSystemEntries`; declared type `Dirent<string>`; flow type `never`; enum identity absent.

Why never: The true arm of typeof dirent === "string" excludes the entire declared Dirent<string> object type. This is the retained pre-Node-10.10 readdirSync compatibility branch.

Reachability: Not reached with the provided Node v24.19.0 and withFileTypes:true. It can run on a pre-10.10 host that ignores withFileTypes and returns strings, which violates the modern declaration used here. This corpus file is a tsc filesystem fragment; the analogous legacy-host branch can be live on such an old host, not during ordinary self-compilation on the pinned modern Node.

No class-(d) site was observed.

The cache-mutation site in `isSymbolAssignedDefinitely` is present unchanged in pinned upstream TypeScript (upstream `checker.ts:30269`, adapted `checker.ts:30272`). The helper can populate `symbol.lastAssignmentPos` through `markNodeAssignments` at adapted `checker.ts:30336`/`:30351`. A normal cold self-checking path is the uninitialized `result` in `core.ts:321`, assigned at `:323` and read at `:325` in `map`; no fabricated type is needed for lazy assignment marking. This is a source-path argument, not an execution count.

## Existing trap and observed routes

The supplied locations are correct: `enumNeverCheck` (`internal/lower/enum_never.go:84`) is called by `value` (`expression.go:582`), the checked numeric-enum switch default (`object.go:1385`), and `shorthand` (`object.go:2021`). The last is the variable-value read in `{ value }`, not ordinary `object.value`. The helper evaluates its operand once and builds a panic carrying the value and numeric enum name; no new machine representation for never is required. `expression` runs prechecks before `value` and applies allocation/boundary/representation work afterward. An earlier error can keep a node from reaching the hook.

A checked switch default reuses the saved scrutinee, including an implicit default. It has no new source AST reference whose queried flow type is never. It is observed at switch entry; a read-only backstop census does not replace that synthetic default check. Ordinary property/element reads pass through `value`, but an optional-index continuation calls `property(receiver)` directly.

| Route | Original method/file | Observed node |
|---|---|---|
| `value` | `expression.go:576`, `value` | `node` |
| `property` | `object.go:441`, `property` | `node` |
| `shorthand` | `object.go:1990`, `shorthand` | `property.Name()` |
| `switch` | `object.go:1350`, `switchStatement` | `switch scrutinee` |
| `readObjectField` | `object.go:754`, `readObjectField` | `node` |
| `updateIndex` | `object.go:2525`, `updateIndex` | `target` |
| `localRead` | `locals.go:280`, `localRead` | `node` |
| `generatorMade` | `generators_stored.go:182`, `generatorMade` | `node` |
| `tupleLength` | `collections.go:401`, `tupleLength` | `tuple receiver` |
| `methodAliasRead` | `library_method_values.go:167`, `methodAliasRead` | `node` |
| `detachedOwnCall` | `detached_own.go:146`, `detachedOwnCall` | `alias receiver after nil guard` |
| `assignment` | `assignments.go:12`, `assignment` | `target after SkipParentheses` |
| `increment` | `assignments.go:334`, `increment` | `operand after SkipParentheses` |
| `assignmentValue` | `assignment_value.go:10`, `assignmentValue` | `target` |
| `logicalAssignment` | `logical_assignment.go:13`, `logicalAssignment` | `target` |
| `updateProperty` | `class.go:728`, `updateProperty` | `target` |
| `staticConstruct` | `class_static.go:787`, `staticConstruct` | `new-expression class reference` |

These include the returned namespace-assignment local read, stored-generator move, retained library-method/own-property alias checks, static class construction binding, and read-modify-write paths. Assignment targets are filtered as requested. `tupleLength` elides its plain-local receiver and returns a constant length. Direct canonical function targets (for example a named sort comparator), generated parameter/capture/temporary reads, namespace-container construction from declaration names, and super’s implicit this read have no separate eligible runtime source-reference read to add. Direct call callees are explicitly excluded.

## Hook and overlay contract

The hook is `/tmp/never-scout/hook/never_scout.go`, overlaid as `<repository>/internal/lower/never_scout.go`. A call is the first statement of `value`. Other calls sit at the route entry or immediately after the referenced target/receiver is obtained. The plain overlay contains 12 patched lower files plus the new hook (13 entries); there is no census loader/state code in it. The census overlay reuses its generated `expression.go` and `locals.go` copies before adding observation calls. Generated `overlay.json` was retained, with additions written to `overlay-scout.json`.

Skips are exactly `ast.IsPartOfTypeNode`, `ast.IsDeclarationName`, **`ast.IsAssignmentTarget`**, and an immediate call parent whose expression is the node. `ast.SkipParentheses` is applied first. Only the type’s own `checker.TypeFlagsNever` bit is examined; union members are not inspected. Identifier, this, property-access and element-access nodes are reads; never-typed call/new nodes produce `kind: "call"` rows.

The literal declaration-name predicate also identifies shorthand names, so `{ value }` is filtered despite the hooked direct-read route. The assignment-target predicate excludes compound/update targets that semantically read and write. These specified filters mean the result is not an unfiltered count of every semantic read.

A mutex protects append and process-lifetime `(file, pos, end)` deduplication. Each row records the requested kinds, route, filename/positions, first 120 bytes of text, parent/grandparent, enclosing name, enum identity, declared symbol type and flow type. Positions use 1-based UTF-16 columns, matching `Program.Where`. Extra fields retain function position, source context/full inline source, and resolved direct argument-call return types. The hook changes no IR, return expressions, or lowering error branches. Before building, path-to-string conversion and UTF-16 formatting were corrected; the census build had no errors and needed no post-build hook fix.

### Full hook

```go
package lower

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode/utf16"
)

var neverScoutMutex sync.Mutex
var neverScoutSeen = map[string]bool{}

// Observation only: no IR, lowering state, returns, or errors are changed.
func neverScoutNote(l *lowering, node *ast.Node, route string) {
	out := os.Getenv("NEVER_SCOUT_OUT")
	if out == "" || node == nil {
		return
	}
	node = ast.SkipParentheses(node)
	kind := "read"
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression:
		kind = "call"
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
	default:
		return
	}
	if ast.IsPartOfTypeNode(node) || ast.IsDeclarationName(node) || ast.IsAssignmentTarget(node) {
		return
	}
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == node {
		return
	}
	t := l.checker.GetTypeAtLocation(node)
	if t == nil || t.Flags()&checker.TypeFlagsNever == 0 {
		return
	}
	file := ast.GetSourceFileOfNode(node)
	pos, end := node.Pos(), node.End()
	key := fmt.Sprintf("%s:%d:%d", file.FileName(), pos, end)
	neverScoutMutex.Lock()
	defer neverScoutMutex.Unlock()
	if neverScoutSeen[key] {
		return
	}
	start := scanner.GetTokenPosOfNode(node, file, false)
	starts := scanner.GetECMALineStarts(file)
	lineIndex := sort.Search(len(starts), func(i int) bool { return int(starts[i]) > start }) - 1
	if lineIndex < 0 {
		lineIndex = 0
	}
	line := lineIndex + 1
	col := len(utf16.Encode([]rune(file.Text()[int(starts[lineIndex]):start]))) + 1
	text := file.Text()[start:end]
	if len(text) > 120 {
		text = text[:120]
	}
	pk, gpk := "<nil>", "<nil>"
	if parent != nil {
		pk = fmt.Sprint(parent.Kind)
		if parent.Parent != nil {
			gpk = fmt.Sprint(parent.Parent.Kind)
		}
	}
	function := "<top>"
	functionPos := -1
	for p := parent; p != nil; p = p.Parent {
		if ast.IsFunctionLike(p) {
			function = "<anonymous>"
			functionPos = p.Pos()
			if p.Name() != nil {
				function = sourceExpression(p.Name())
			} else if p.Parent != nil && p.Parent.Name() != nil {
				function = sourceExpression(p.Parent.Name())
			}
			break
		}
	}
	declared := ""
	if symbol := l.flagValueSymbol(node); symbol != nil {
		declared = l.checker.TypeToString(l.checker.GetTypeOfSymbol(symbol))
	}
	record := map[string]any{"kind": kind, "route": route, "file": file.FileName(), "pos": pos, "end": end, "line": line, "col": col, "text": text, "node_kind": fmt.Sprint(node.Kind), "parent_kind": pk, "grandparent_kind": gpk, "function": function, "already_trapped": l.enumNeverIdentity(node, map[*ast.Node]bool{}) != nil, "declared_type": declared, "flow_type": l.checker.TypeToString(t)}
	record["function_pos"] = functionPos
	arg := node
	for arg.Parent != nil && arg.Parent.Kind == ast.KindParenthesizedExpression {
		arg = arg.Parent
	}
	if arg.Parent != nil && arg.Parent.Kind == ast.KindCallExpression && arg.Parent.AsCallExpression().Expression != arg {
		if signature := l.checker.GetResolvedSignature(arg.Parent); signature != nil {
			record["argument_call_returns"] = l.checker.TypeToString(l.checker.GetReturnTypeOfSignature(signature))
			record["argument_call"] = sourceExpression(arg.Parent.AsCallExpression().Expression)
		}
	}
	if strings.HasPrefix(string(file.FileName()), "/tmp/") && !strings.Contains(string(file.FileName()), "/adapted/") {
		record["inline_source"] = file.Text()
	}
	// Include source context so inline test inputs remain classifiable after TempDir cleanup.
	lines := strings.Split(file.Text(), "\n")
	lo, hi := line-8, line+7
	if lo < 0 {
		lo = 0
	}
	if hi > len(lines) {
		hi = len(lines)
	}
	record["context_start"] = lo + 1
	record["context"] = strings.Join(lines[lo:hi], "\n")
	f, err := os.OpenFile(out, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	err = json.NewEncoder(f).Encode(record)
	closeErr := f.Close()
	if err != nil {
		panic(err)
	}
	if closeErr != nil {
		panic(closeErr)
	}
	neverScoutSeen[key] = true
}
```

## Runs, wall times, and failures

Heavy root commands ran sequentially. The tool environment was loaded first when present. `TMPDIR`, Go/npm/build caches, external sources, binaries and logs were under `/tmp/never-scout`. `NODE_EXTRA_CA_CERTS` and `npm_config_cafile` both named `/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`, for dependency and measurement commands and their inherited children.

| Step | Child exit | Wall seconds |
|---|---:|---:|
| upstream | 0 | 11.655 |
| packages | 0 | 1.511 |
| apply | 0 | 623.600 |
| overlay | 0 | 13.710 |
| build | 0 | 298.535 |
| tsc-full | -9 | 9000.390 |
| tsc-first-error | 0 | 2811.855 |
| oracle | 1 | 4492.007 |
| corpus-supplement-build | 0 | 308.544 |
| corpus-sweep | 0 | 465.518 |
| corpus-host-build | 0 | 68.294 |
| corpus-host-boolean | 0 | 0.910 |
| corpus-host-string | 0 | 0.737 |
| corpus-host-number | 0 | 0.848 |
| tsc-full-memory-capped | 0 | 15661.328 |

The initial full run received SIGKILL (`-9`; the timing wrapper’s shell exit was 247) after 9000.390 s, not timeout’s 124. It emitted no compiler error text. A 16 GiB cgroup limit and one OOM kill were observed afterward; no before-run OOM baseline was captured, so this is consistent with memory exhaustion rather than a timestamped proof of attribution. Its scout data is retained as `tsc-full-partial.jsonl`. Its latent stream had only the header because `LatentWithRelationChecks` buffers file records until traversal returns; the scout appends immediately.

The first-error run completed. One final full-mode retry used `GOMEMLIMIT=10GiB` with the identical binary/overlay and the original five-hour timeout. This is a GC resource policy, not a lowering feature change. Its output is `tsc-full-retry.jsonl`; the completed fallback remains `tsc.jsonl`. The mode labels in the count tables identify which file supplies each result. The full retry exited 0 in 15,661.328 seconds (4 h 21 min 1.328 s), with its no-output assertion enabled and its final relation-report record present. From the first recorded retry resource baseline (about ten minutes into the retry) to completion, cgroup OOM/OOM-kill counters stayed at 1/1: no additional event was observed.

The primary oracle suite exited 1 without reaching its 150-minute suite timeout. Its failures were:

- `TestStdoutColumnsRefusesWASI`
- `TestAssignmentProofWiderTargetMutant`
- `TestAssignmentProofWiderTarget`
- `TestAssignmentProofCompoundMutant`
- `TestAssignmentProofCompound`
- `TestAssignmentProofChainedMutant`
- `TestAssignmentProofChained`
- `TestAssignmentProofReturnDefinedMutant`
- `TestAssignmentProofReturnDefined`
- `TestAssignmentProofIfNarrowingMutant`
- `TestAssignmentProofIfNarrowing`
- `TestAssignmentProofScannerKeywordMutant`
- `TestAssignmentProofScannerKeyword`
- `TestFamiliesCatchAPlantedCase_CountsAreRecorded`

The first failure needed `WASI_SYSROOT`; the twelve assignment-proof failures overrode `NODE_PATH` to the absent repository-local package install; the count-recording probe hit its own 90-second child limit. The twelve assignment-proof tests all passed when a test-only overlay redirected that one `NODE_PATH` assignment to the external pinned packages. No lowering decision was changed by that bridge.

The oracle’s shared `lowered` helper (`oracle_test.go:660`) uses `load.Load` and in-process `lower.Lower`. Some CLI/mutant controls and the isolated host-row builder use separately built binaries. To cover their current-HEAD corpus inputs, the throwaway sweep applies those same production load/lower operations to every `.a`/`.ts` testdata root and every registered `fixturePaths()` root, with no backend execution. It records loader/lowering rejections unchanged and catches only an outer per-root panic to continue observation. `rememberImports` is Node-runner bookkeeping, not a Lower input. The private host builder was separately built with the scout overlay and run for Boolean, string and number.

The sweep covered 4,001 roots: 2,712 lowered, 1,266 were refused by lowering, 23 were rejected by the loader/checker, and zero per-root panics were recovered. The three host cases passed and recorded no never sites. The sweep plus twelve targeted retests passed. Raw suite and supplemental records are not blindly added: the union is deduplicated by site, and all duplicate descriptors agree.

## Reproduction

The external `stage3` copy was initially byte-for-byte from the pinned repository so its lockfile packages could be installed outside the repository. `apply.sh` ran from that copy; make_overlay and Go builds used the actual pinned repository. The 79 measured compiler inputs were unchanged throughout all tsc attempts. For disk space after the first measurements, unused upstream test fixtures, redundant copied stage3 evidence/directories, and a disposable Go build cache were removed; compiler sources, Git objects, API packages, live helper products, overlays, and observations were retained.

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/tmp/never-scout/tmp
export GOCACHE=/tmp/never-scout/go-cache
export GOPATH=/tmp/never-scout/go-path
export XDG_CACHE_HOME=/tmp/never-scout/cache
export ADAMIC_BUILD_CACHE_DIR=/tmp/never-scout/cache
export STAGE3_CACHE=/tmp/never-scout/stage3-cache
export npm_config_cache=/tmp/never-scout/npm-cache
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/never-scout/TypeScript
export PATH=/tmp/never-scout/stage3/api/node_modules/.bin:$PATH
export NODE_PATH=/tmp/never-scout/stage3/api/node_modules
export PYTHONDONTWRITEBYTECODE=1
# Initial setup (after fetching/detaching and recursive submodule initialization):
# copy stage3 outside the repository, npm ci --prefix /tmp/never-scout/stage3/api --ignore-scripts --no-audit --no-fund
# provision upstream TypeScript HEAD 050880ce59e30b356b686bd3144efe24f875ebc8 under $ADAMIC_TYPESCRIPT_SOURCE
cd /tmp/never-scout
bash stage3/apply.sh /tmp/never-scout/adapted > /tmp/never-scout/apply.log 2>&1
cd /workspace/adamic
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/never-scout/latent-overlay
gofmt -w /tmp/never-scout/latent-overlay/*.go
python3 /tmp/never-scout/patch_overlay.py /tmp/never-scout/latent-overlay
gofmt -w /tmp/never-scout/latent-overlay/scout_*.go
go build -buildvcs=false -overlay=/tmp/never-scout/latent-overlay/overlay-scout.json -o /tmp/never-scout/latent-census ./stage3/census/latent/tool
NEVER_SCOUT_OUT=/tmp/never-scout/tsc.jsonl LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 timeout 5h /tmp/never-scout/latent-census /tmp/never-scout/adapted/src/compiler /tmp/never-scout/latent.jsonl
# Preserve initial partial files, then use LATENT_FULL=0 for the completed fallback.
GOMEMLIMIT=10GiB NEVER_SCOUT_OUT=/tmp/never-scout/tsc-full-retry.jsonl LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 timeout 5h /tmp/never-scout/latent-census /tmp/never-scout/adapted/src/compiler /tmp/never-scout/latent-full-retry.jsonl
python3 /tmp/never-scout/patch_overlay.py /tmp/never-scout/plain-overlay
# Copy plain-overlay/corpus-overlay.json to /tmp/never-scout/corpus-overlay.json.
NEVER_SCOUT_OUT=/tmp/never-scout/corpus.jsonl go test -overlay /tmp/never-scout/corpus-overlay.json ./internal/oracle -count=1 -timeout 150m
# Supplemental oracle overlay adds the test below and the one-line NODE_PATH bridge.
go test -c -overlay=/tmp/never-scout/corpus-supplement-overlay.json -o /tmp/never-scout/corpus-scout.test ./internal/oracle
cd /workspace/adamic/internal/oracle
NEVER_SCOUT_OUT=/tmp/never-scout/corpus-sweep.jsonl /tmp/never-scout/corpus-scout.test '-test.run=^(TestNeverScoutAllTestdata|TestAssignmentProof.*)$' -test.count=1 -test.timeout=150m -test.v
cd /workspace/adamic
go test -c -overlay=/tmp/never-scout/corpus-overlay.json -o /tmp/never-scout/host-lower.test ./internal/lower
cd /workspace/adamic/internal/lower
for host_probe_type in boolean string number; do
  NEVER_SCOUT_OUT=/tmp/never-scout/corpus-host.jsonl ADAMIC_PROCESS_ORACLE_TYPE="$host_probe_type" ADAMIC_PROCESS_ORACLE_IR="/tmp/never-scout/host-$host_probe_type.gob" /tmp/never-scout/host-lower.test -test.run=^TestHostBindingProcessOracleIR$ -test.count=1 -test.timeout=900s
done
```

### Overlay builder

```python
import pathlib,json,sys
repo=pathlib.Path('/workspace/adamic')
out=pathlib.Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
base=json.loads((out/'overlay.json').read_text()) if (out/'overlay.json').exists() else {'Replace':{}}
replace=dict(base['Replace'])
# Every insertion observes only; no branches, result expressions or error handling are rewritten.
patches={
 'expression.go': [('func (l *lowering) value(node *ast.Node) (ir.Expression, error) {','neverScoutNote(l, node, "value")')],
 'object.go': [
 ('func (l *lowering) property(node *ast.Node) (ir.Expression, error) {','neverScoutNote(l, node, "property")'),
 ('func (l *lowering) shorthand(property *ast.Node) (ir.Expression, error) {','neverScoutNote(l, property.Name(), "shorthand")'),
 ('func (l *lowering) switchStatement(node *ast.Node) ([]ir.Statement, error) {','neverScoutNote(l, node.AsSwitchStatement().Expression, "switch")'),
 ('func (l *lowering) readObjectField(node *ast.Node, property ir.Property) ir.Expression {','neverScoutNote(l, node, "readObjectField")'),
 ('func (l *lowering) updateIndex(node *ast.Node, target *ast.Node, operator ast.Kind, valueNode *ast.Node) ([]ir.Statement, error) {','neverScoutNote(l, target, "updateIndex")')],
 'class_static.go':[('func (l *lowering) staticConstruct(node, declaration *ast.Node, constructor int, arguments []ir.Expression) ir.Expression {','neverScoutNote(l, node.AsNewExpression().Expression, "staticConstruct")')],
 'class.go':[('func (l *lowering) updateProperty(node *ast.Node, target *ast.Node, operator ast.Kind, valueNode *ast.Node) ([]ir.Statement, error) {','neverScoutNote(l, target, "updateProperty")')],
 'locals.go':[('func (l *lowering) localRead(node *ast.Node, local int) (ir.Expression, error) {','neverScoutNote(l, node, "localRead")')],
 'generators_stored.go':[('func (l *lowering) generatorMade(node *ast.Node, function int) (ir.Expression, error) {','neverScoutNote(l, node, "generatorMade")')],
 'collections.go':[('func (l *lowering) tupleLength(node *ast.Node, tuple *ast.Node) (ir.Expression, error) {','neverScoutNote(l, tuple, "tupleLength")')],
 'library_method_values.go':[('func (l *lowering) methodAliasRead(node *ast.Node) (ir.Expression, error) {','neverScoutNote(l, node, "methodAliasRead")')],
 'detached_own.go':[('receiver := l.detachedOwnCallReceiver(node)\n\tif receiver == nil {\n\t\treturn nil, false, nil\n\t}', 'neverScoutNote(l, receiver, "detachedOwnCall")')],
 'assignments.go':[('target := ast.SkipParentheses(binary.Left)','neverScoutNote(l, target, "assignment")'),('operand = ast.SkipParentheses(operand)','neverScoutNote(l, operand, "increment")')],
 'assignment_value.go':[('target := ast.SkipParentheses(node.AsBinaryExpression().Left)','neverScoutNote(l, target, "assignmentValue")')],
 'logical_assignment.go':[('target := ast.SkipParentheses(binary.Left)','neverScoutNote(l, target, "logicalAssignment")')],
}
for name, edits in patches.items():
 key=str(repo/'internal/lower'/name)
 text=pathlib.Path(replace.get(key,key)).read_text()
 for needle,call in edits:
  assert text.count(needle)==1,(name,needle,text.count(needle))
  text=text.replace(needle,needle+'\n\t'+call,1)
 dest=out/('scout_'+name);dest.write_text(text);replace[key]=str(dest)
replace[str(repo/'internal/lower/never_scout.go')]='/tmp/never-scout/hook/never_scout.go'
(out/('overlay-scout.json' if (out/'overlay.json').exists() else 'corpus-overlay.json')).write_text(json.dumps({'Replace':replace},indent=2)+'\n')
```

### Throwaway corpus sweep

```go
package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The oracle's program path is load.Load followed by lower.Lower. The sweep
// performs those same operations, without building or executing a backend.
// Checked-invalid and refused fixtures remain rejected; each result is recorded.
func TestNeverScoutAllTestdata(t *testing.T) {
	root, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata"))
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	seenRoots := map[string]bool{}
	add := func(path string) error {
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if !seenRoots[absolute] {
			seenRoots[absolute] = true
			paths = append(paths, absolute)
		}
		return nil
	}
	for _, path := range fixturePaths() {
		if !filepath.IsAbs(path) {
			path = filepath.Join(repository, path)
		}
		if err := add(path); err != nil {
			t.Fatal(err)
		}
	}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && (filepath.Ext(path) == ".a" || filepath.Ext(path) == ".ts") {
			return add(path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	out, err := os.Create("/tmp/never-scout/corpus-coverage.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	encoder := json.NewEncoder(out)
	counts := map[string]int{}
	attempt := func(path string) (record map[string]any) {
		record = map[string]any{"file": path, "status": "lowered"}
		defer func() {
			if recovered := recover(); recovered != nil {
				record["status"] = "panic"
				record["error"] = fmt.Sprint(recovered)
			}
		}()
		program, err := load.Load([]string{path})
		if err != nil {
			record["status"] = "load_rejected"
			record["error"] = err.Error()
			return
		}
		_, err = lower.Lower(context.Background(), program)
		if err != nil {
			record["status"] = "lower_rejected"
			record["error"] = err.Error()
		}
		return
	}
	for i, path := range paths {
		record := attempt(path)
		counts[record["status"].(string)]++
		if err := encoder.Encode(record); err != nil {
			t.Fatal(err)
		}
		if (i+1)%100 == 0 {
			t.Logf("scout coverage %d/%d: %v", i+1, len(paths), counts)
		}
	}
	t.Logf("scout coverage complete: roots=%d, results=%v", len(paths), counts)
}
```

The test-only bridge replaces the assignment-proof helper’s repository-local `NODE_PATH` append with `NODE_PATH=/tmp/never-scout/stage3/api/node_modules`; its version/zero-diagnostic assertion is retained. Both extra test entries exist only through the supplemental overlay.

## Artifact fingerprints

Only this report is committed, as requested. Raw streams, logs and scratch source files remain outside the repository. SHA-256 fingerprints identify the measured artifacts below.

| Artifact | SHA-256 |
|---|---|
| `hook/never_scout.go` | `02cff105ac3153d5127cae658ed06f0c8936134f64352b297138544ab6d6a070` |
| `patch_overlay.py` | `0f06e4be637907cce2650ef934a67e351486d5558967430e7a0f91bcfd87bd2b` |
| `compiler-input-hashes.json` | `e654239c73e6e73c783f0d7245e335c68cdce13d320324d349fea4e4c5ec3be5` |
| `latent-overlay/overlay.json` | `32c82551efd39fe8bd9792114082890207f1987716654353a976a3b9ecbb3201` |
| `latent-overlay/overlay-scout.json` | `af8484e04fb763e78940723a8829d457f4bba0358f99b3025d0e57ec907c3215` |
| `corpus-overlay.json` | `74d850665fc57701dbc0c8cf7687aafb5672be9f28362f5ccb1ea2f84b07411a` |
| `corpus-supplement-overlay.json` | `00f4af92402fb57baea9b00c837fd8037157d1e1c1b9c5b3fa3b08dfc0b6b4bb` |
| `never_scout_corpus_test.go` | `ec037e66c4dbd75da8dec63b5aa51bf1b58c3c5055537236fbeec69bc5b2816f` |
| `assignment_proofs_external_test.go` | `d70695d82aa4fe4834d73430135f0edeb5a6a539b0d6a2a07ca939c782a5637e` |
| `tsc.jsonl` | `190078ff72e4afe623304a7b1572be3d1d4bfa5da9cdee0c6ac82e0162196ecf` |
| `tsc-full-partial.jsonl` | `ee1ca82f4093a3e8bf93647a08dbc215600570d3fa36fee7ec3ae542017c7f2c` |
| `tsc-full-retry.jsonl` | `f3f0005a8f45708f160ed63742ef6c2e8330262f925d1a0e609d986b4e7278d1` |
| `corpus.jsonl` | `9208c1f45757a6ccf2b3d5b4e77392671cbfbd33efc359654c4de40ce2e7591d` |
| `corpus-sweep.jsonl` | `e97cb84a2620fb1bf00422d74e9f8fca2237af4103870694200fdd2baff6265d` |
| `corpus-unique.jsonl` | `06b57f3223f3c83f1d861fdb4135e2d2fd0bdf53ea40ffc49e09904f3cf16eb9` |
| `corpus-coverage.jsonl` | `5ec97c5a4659dbe5b5132f1aca645888d01177d9155cf79044e1516b88d097e9` |
| `latent-census` | `b1d0b5e8c4f4831ad231d79f0c215d05d57e235d3df2f76112c2c4c7954474d0` |
| `corpus-scout.test` | `ab6633ffc2919688e8ccab87aae8ab46d9e29528243c199b647f601b3abbf618` |
| `host-lower.test` | `9724c23c73a7ddc4eeecc758fbf658e655dae03b46cf5007d6a0d5516aca0281` |

## Repository status and scope

`git status --porcelain` was empty before writing this report. All instrumentation was in scratch overlays; no source implementation change was made. The report branch is `proof/lower-never-scout-report`, based on the pinned HEAD, with only `proof/lower-never-scout/report.md` staged/committed. That branch is the only authorized publication target; no other branch is a push target.

The main surprises were the live cache-mutation false-never path, the captured module-import tuple populated by a visitor, the Boolean `null!` placeholder narrowing, generic calls resolving to never without being intrinsic runtime stops, the literal skip predicates excluding shorthand/update reads, and the difference between a completed first-error census and interrupted full coverage. A small static site count is not a frequency estimate: the cache-property read can be on an ordinary checking path, whereas most corpus narrowing probes are dead or deliberately violate a type promise.

# Named nested functions

Eleven standalone `.a` programs extracted from TypeScript 6.0.3, tag v6.0.3,
commit 050880ce59e30b356b686bd3144efe24f875ebc8. The exact census reason is
`a function inside a function (a closure)`: 5,574 sites on 5,574 start lines in
57 original files. This is a syntax inventory, not evidence that those files
passed Adamic's checker. Census commit: 429c1177f0130f785c19cf590d1860513b2ddbfc.

The inventory is concentrated in checker.ts (2,443 sites), nodeFactory.ts (518),
emitter.ts (392), binder.ts (165), and scanner.ts (80). This bucket prioritizes
the scanner milestone and covers a few checker forms. It is not proportional
coverage of all 57 files.

## Source fidelity and scope

01 was built first. It retains createScanner's frame locals, initialization call
before the helper declarations, 24 sibling named helpers, and the original
object's arrow getters and retained named helper properties. It keeps the helper
statements unchanged, including `var`, assertions, comma expressions, generic
lookahead, and optional parameters. The public Scanner annotation is removed so
the cut-down object infers its actual type. Debug-only object instrumentation and
unretained scanner properties are cut. `scanDigits` is exposed for the standalone
driver. Enum constants needed by the slice are copied as scalar object fields
from stock typescript@6.0.3; this bucket does not test enum lowering.

The fixture runs decimal/octal digit scanning, separate scanner instances,
lookahead rollback, scanRange rollback, and text replacement. It does not retain
the main scan dispatch or implement a complete lexer. In particular it is not
proof that scanner.ts runs natively. Its first compile blocker on both tested
branches is the unchanged non-null assertion in codePointAt.

Smaller fixtures keep the named helper bodies verbatim and introduce minimal
standalone frame/driver scaffolding. The parser fixture supplies the scanner's
real token sequence for `const x = 1;` (87, 80, 64, 9, 27, 1); it does not perform
parsing. The binder fixture supplies a minimal Symbol constructor and counts
allocations; it does not bind an AST. The constituent-count fixture uses an
array.reduce adapter for the imported reduceLeft operation. Checker type and
symbol views retain only the fields read by these helpers.

verify-source.cjs checks 41 retained helper-body instances against the independent
upstream source. The enclosing createScanner slice and introduced drivers are
excluded from that body comparison. Source spans in each file identify the
upstream declarations, not the introduced wrapper code.

| Fixture | Form and exercised behavior |
| --- | --- |
| 01_scanner_frame | 24 sibling declarations, shared mutable scanner locals, closure object, hoisted setText/resetTokenState, generic lookahead/range helpers |
| 02_scanner_digits | scanDigits calls sibling charCodeChecked/Unchecked; returned closures share position and tokenValue; independent scanner frames |
| 03_scanner_hoisting | getText called before its declaration line |
| 04_scanner_escaped_text | returned getText outlives its frame; two independent captured string parameters |
| 05_scanner_reassigned_text | a captured let is assigned after the function value is saved |
| 06_parser_token_state | sibling token reader observes writes by nextTokenWithoutCheck through the supplied scanner |
| 07_binder_symbol_count | createSymbol mutates shared symbolCount; separate binder instances |
| 08_checker_symbol_recursion | getSymbolPath recurses along ts.SyntaxKind.Identifier's parent symbols |
| 09_checker_constituent_recursion | mutual call graph between getConstituentCount and getConstituentCountOfTypes through a reducer callback; union, intersection, alias and origin paths |
| 10_checker_arrow_cleanup | original generic cleanup inside an arrow, with save/update/early-cleanup statements retained; restores enclosingDeclaration |
| 11_scanner_method_wrapper | original getText inside an introduced method wrapper with a captured local |

A scan of the pinned src/compiler found zero named declarations whose nearest
function ancestor is a method. 11 therefore explicitly uses a driver method;
it does not claim an upstream method origin. 10's arrow placement is upstream.

## Recorded outcomes

status.json has exactly the requested schema and records current main at
**ef3d907ecdc4c771b016f7d9c52372def057a340**. nested-functions-status.json uses
that same schema for **b15216dabf65ffaa7152f6e64709b7b062ea01a9**, the fetched
origin/codex/nested-functions. Its docs/nested-functions.md was read before the
comparison. Neither compiler was patched for these runs. Node used the stock
oracle/node.mjs runner, without an enum/namespace runner change.

All eleven original-source Node runs exit 0 with empty stderr. Main compiles
none: nine NotYet and two Refused. On the feature branch:

| Fixtures | Observed result |
| --- | --- |
| 01, 09 | Refused: unchanged non-null assertions |
| 02, 03, 04 | Compiler panic: `native: a store into the borrowed parameter text` |
| 05, 07, 11 | Compiles; native stdout, stderr and exit match Node byte for byte |
| 06 | NotYet: assignment expression with a number and a number |
| 08 | Refused: non-boolean condition |
| 10 | NotYet: value of type undefined |

The schema has no compiler-crash outcome. The three panics are stored under
Checker as an unsuccessful build, with their complete actual panic output in
what. They are compiler crashes, not TypeScript checker diagnostics or successful
feature support. Stack addresses and checkout paths are observations from this
run and may differ on reproduction. No silent miscompile was observed in the
three fixtures that reached native execution.

All three compiling feature fixtures also passed native --sanitize builds with
ASan, UBSan and ASAN_OPTIONS=detect_leaks=1. The existing filtered main oracle
check passed for functions.a and generic_functions.a. No full integration gate,
JavaScript backend comparison, or tsc suite was run for this fixture-only unit.
No fixtures_test.go, oracle registry, compiler source, or other bucket was edited.

## Mutants

1. Flatten 03's nested getText and its captured text slot to top level, leaving
   the getText body unchanged. On main the original is NotYet; the mutant
   Compiles, exits 0, and prints `const x = 1;` followed by a newline, exactly
   like Node. The recorded main outcome distinguishes them. The scratch mutant
   is /tmp/nested-flattened.a; its source is reproduced below.
2. Change 03's helper from `return text;` to `return text + "!";` in a scratch
   copy. The independent source-body audit exits 1 with
   `03_scanner_hoisting.a: changed getText`. Production fixtures were not mutated.

```a
let text = '';
function getText(): string { return text; }
function readScanner(input: string): string {
    text = input;
    return getText();
}
console.log(readScanner('const x = 1;'));
```

## Reproduction

Setup succeeded with Go 1.27.1, clang 20.1.8 and Node 24.19.0. Timing lines:
go ready 0s, clang ready 0s, node ready 0s, submodules ready 0s,
build cache warm 113s, done 113s. nproc: 5; cgroup cpu.max: 400000 100000.
Shells sourced /workspace/adamic-tools/env.sh. Tests wrote logs without piping.
The original recursive fetch completed later and its queued checkout reported
that the requested branch already existed. Its queued setup rerun succeeded in
26s: all ready steps 0s, build cache warm 26s, done 26s. The already-created
fixture branch remained selected; both setup logs are retained.

Prepare a detached worktree at the recorded feature commit, with its cohere
submodule available, then run from the fixture branch repository:

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/fixtures/nested-functions/record.py \
  --feature-root /tmp/nested-comparison/feature \
  --logs /tmp/nested-comparison > /tmp/nested-comparison/run.log 2>&1
node stage3/fixtures/nested-functions/verify-source.cjs \
  /tmp/nested-tsc /tmp/nested-api/node_modules/typescript \
  > /tmp/nested-source-audit.log 2>&1
go test ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/functions.a$' \
  -count=1 -timeout 10m -v > /tmp/nested-filtered-oracle.log 2>&1
```

record.py records all top-level NN_*.a files, saves individual Node/build/native
logs, and writes both status arrays. It prints SILENT MISCOMPILE if a compiled
binary differs. It does not turn a compiler panic into a passing native result.
The helper audit requires the independent v6.0.3 source checkout and stock
TypeScript 6.0.3 API. Representative logs are committed under logs/; complete
individual run logs remain under /tmp/nested-comparison.

## Left unfinished

The requested several-hundred-sibling createTypeChecker fixture is not built.
I did not finish isolating its real types, shared initialization and transitive
helper dependencies. Synthetic renamed arithmetic helpers would cover scale but
would not meet the required source-fidelity contract. The small checker fixtures
exercise actual helper bodies, not the scale requirement.

Full parser and binder equivalents, complete scanner token dispatch, factory and
emitter proportional coverage, and a native TypeScript compilation remain
uncovered. Parser and binder here are state-bearing slices only. The source's
var, assertion, truthiness and generic/optional-parameter interactions remain
visible blockers; this unit did not adapt or implement those features. There is
no claim that all October 7 requirements are met.

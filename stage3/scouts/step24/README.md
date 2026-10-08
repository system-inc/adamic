Built: a streaming exact-byte node-dump comparator, parser-source fixtures, and a measured step 24 scout.
Base: origin/main 45487a809f89885a3fc651cd590e7dabf31362dc; branch codex/step24-scout; only this directory changes.
Results: comparator green; fresh latent census and Node recursion measured; context and native frame calibration agree with source Node.
Mutants: comparator, context restoration, speculation rewind, factory header, instrumentation and existing parser mutants all caught, detailed below.
Limits: no native parser acceptance; reconstructed corpus differs from the committed target by 13 bytes; language options remain undecided.

# Evidence and scope

TypeScript source pin is 050880ce59e30b356b686bd3144efe24f875ebc8 (6.0.3).
Node is v24.19.0, Go 1.27.1, clang 20.1.8, nproc 5, CPU quota 4.
The cohere pin is 7945d102a6c18dd36adf9114a758ce646e8b2359; its typescript-go
pin is d92d9bfee114c80be2c375d72edae966176e3a4f. No feature branches were merged.

CLAUDE.md, README.md, docs/0.1.md, docs/memory.md and the parser README and
BLOCKERS informed this scout. The entire prior evidence tree was read and indexed:
2,585 files, 59,388,023 expanded bytes. [prior-evidence-index.json](evidence/prior-evidence-index.json)
records every file's decoded size and SHA256, including gzip contents. This is a
provenance inventory, not a claim of individually revalidating every historical run.
Recent BLOCKERS rows describe scratch integrations and throwing discovery stubs;
they are not current-main native passes.

Setup ran with GOPROXY='https://proxy.golang.org|direct', then sourced the printed
/workspace/adamic-tools/env.sh. Timing lines: Node ready 0.029s; Go ready 0.045s;
clang ready 0.313s; markdown install duration 0.951s, ready 1.063s; submodules
20.294s; Go build ready 278.335s; cache warm 278.448s; done 278.475s.
[setup.log.gz](evidence/setup.log.gz) retains the exact output. No setup workaround was needed.

# What the parser adds to the scanner

There are three different measurements, with different meanings.
The resolved file graph has 79 parser files and 79 scanner files, with zero
exclusive files, because both pass through the compiler barrel. The supplementary
value-reference graph has 26 parser files, 12 scanner files, and 14 exclusive files,
including current main's hostErrors.ts. See [file-closure.json](evidence/file-closure.json)
and [value-closure.json](evidence/value-closure.json). These existing repository tools
retain whole top-level declarations and namespaces; neither produces a validated slice.

The syntax inventory subtracts scanner-reachable declaration spans from parser-reachable
spans, including added declarations in shared files. [constructs.json](evidence/constructs.json)
retains every counted site as file, one-based line/column and source text. Counts are
syntactic, including type-level signatures. For example, a bodyless function signature
includes function-type nodes and methods, not just overload declarations. Retaining
Parser/IncrementalParser namespaces whole conservatively includes code a fresh parse
will not execute. It would be wrong to call every count a necessary runtime feature.

The fresh latent census on this exact main measured a checker-rejected program,
with diagnosed bodies skipped and only the first lowering failure per attempted unit.
Its guarded overlay never writes usable IR and imports no backend. Findings below are
whole-file unique (kind, where, reason, text) sites for files containing declaration-delta
spans, not findings restricted to those spans. [latent.json](evidence/latent.json) preserves
all exact reasons and file:line:column sites; [latent-raw.jsonl.gz](evidence/latent-raw.jsonl.gz)
keeps attempt contexts and eligibility. This is a blocker ledger, not compilation proof.

| File under src/compiler | Fresh NotYet | Fresh Refused |
| --- | ---: | ---: |
| checker.ts | 9 | 26 |
| commandLineParser.ts | 82 | 101 |
| core.ts | 218 | 165 |
| factory/baseNodeFactory.ts | 1 | 0 |
| factory/emitNode.ts | 30 | 10 |
| factory/nodeChildren.ts | 4 | 1 |
| factory/nodeConverters.ts | 1 | 42 |
| factory/nodeFactory.ts | 11 | 418 |
| factory/nodeTests.ts | 0 | 227 |
| factory/parenthesizerRules.ts | 2 | 37 |
| factory/utilities.ts | 53 | 176 |
| factory/utilitiesPublic.ts | 1 | 2 |
| hostErrors.ts | 2 | 2 |
| parser.ts | 55 | 283 |
| path.ts | 38 | 16 |
| performance.ts | 9 | 1 |
| performanceCore.ts | 2 | 7 |
| scanner.ts | 34 | 72 |
| tracing.ts | 4 | 23 |
| types.ts | 1 | 8 |
| utilities.ts | 397 | 760 |
| utilitiesPublic.ts | 64 | 219 |
| visitorPublic.ts | 36 | 24 |

The checked-in latent report is historical main b8fb957a: parser.ts 46 NotYet/121
Refused, nodeFactory.ts 16/5, baseNodeFactory.ts 1/0, factory/utilities.ts 67/118,
nodeTests.ts 227/227, utilities.ts 517/343. These must not replace the fresh counts.
Changes to adaptation input, checker eligibility and compiler refusal rules all affect
counts. No isolated compiler-feature improvement is inferred from that comparison.

Selected declaration-delta syntax counts, with first-site witnesses:

| File | Construct | Count | First site in main-adapted source |
| --- | --- | ---: | --- |
| factory/baseNodeFactory.ts | bodyless function signature | 5 | 14:5 |
| factory/nodeFactory.ts | generic function | 51 | 503:58 |
| factory/nodeFactory.ts | bodyless function signature | 44 | 478:29 |
| factory/nodeFactory.ts | nested function or callback | 593 | 496:40 |
| factory/nodeFactory.ts | assertion | 55 | 1026:20 |
| factory/nodeFactory.ts | non-null assertion | 27 | 1215:23 |
| factory/nodeFactory.ts | field or index write | 971 | 1187:13 |
| factory/nodeFactory.ts | construction | 1 | 7045:39 |
| factory/utilities.ts | generic function | 1 | 678:1 |
| factory/utilities.ts | bodyless function signature | 10 | 1551:1 |
| factory/utilities.ts | assertion | 8 | 608:58 |
| factory/utilities.ts | non-null assertion | 3 | 1560:119 |
| parser.ts | generic function | 181 | 443:1 |
| parser.ts | bodyless function signature | 354 | 443:31 |
| parser.ts | nested function or callback | 681 | 496:33 |
| parser.ts | assertion | 130 | 1161:28 |
| parser.ts | non-null assertion | 41 | 1279:25 |
| parser.ts | field or index write | 82 | 1341:5 |
| parser.ts | array length write | 2 | 2283:17 |
| parser.ts | parser speculation call | 82 | 1674:29 |
| parser.ts | scanner speculation boundary | 4 | 2273:15 |
| parser.ts | namespace | 3 | 1437:1 |
| parser.ts | construction | 12 | 1463:53 |

## Speculation and captured state

parser.ts:2256 saves token, diagnostic length, pending parse error and context flags.
At 2273/2274 it delegates scanner state to lookAhead/tryScan. At 2283 it truncates
parseDiagnostics for failed tries and lookahead, but deliberately keeps diagnostics
for Reparse. A successful try commits; lookahead always rewinds. Context flags are
asserted equal, not blindly overwritten. doOutsideOfContext (2043) and
doInsideOfContext (2064) restore only the bits they changed, preserving cached flags.
The fixture sets another bit inside a nested callback to prove this distinction.

Node's generic callback result uses ECMAScript falsiness, not merely undefined:
false, zero, empty string and missing results all mean failed speculation.
The scout's speculation fixture specializes to boolean and models only a scalar
scanner position; it does not prove scanner token-value/flags rewind, nor generic
result lowering. Identifiers, node counts and cached parse data are not all restored
by this helper. Rewinding every allocation or every shared field would change Node.
Speculative objects returned by lookahead can also escape the speculation call,
so reclaiming a whole region immediately requires an escape proof.

In typescript-go, parser/parser.go:340 defines ParserState, mark at 351 and rewind
at 364; lookAhead at 376 unconditionally restores it. It restores scanner state,
context flags, three diagnostic/JSDoc/reparsed slices and await/error state explicitly.
That is a different state contract, not an interchangeable implementation of tsc's
closure. Parser lives in a sync.Pool (parser.go:119), with node/string arenas at
90/91 and an explicit per-parser factory. Adamic cannot import Go's collector as
its ownership solution. The scanner is UTF-8 byte based (scanner.go:455/459);
tsc positions and dump escaping are UTF-16. Go positions need conversion before
using it as a second witness. No Go runtime node-dump equivalence is claimed here.

## Factory construction, shapes and JSDoc

parser.ts:437 and factory/baseNodeFactory.ts:26 cache constructor values from
objectAllocator. utilities.ts:8505, 8519, 8531 initialize Node/Token/Identifier base
fields; utilities.ts:8552 supplies constructors through casts. nodeFactory.ts:1209
casts a base node to generic Mutable<T>, then createBaseDeclaration at 1213 fills
symbol/localSymbol. parser.ts:2600 finishNode fills range and context/error flags.
A base allocation alone does not establish subtype fields such as Identifier.text.
The factory-cast fixture runs on Node and reads undefined from a field typed string;
current main loudly refuses it with an a-check header, as it should.

Node observations in constructs.json include 24 distinct kind/own-key shapes from
a directed JS parse. SourceFileObject has own absent-valued parent/original/emitNode
and many caches; TokenObject has fewer base fields. NodeArray has indexed keys plus
pos, end, hasTrailingComma and transformFlags. Absence and present-undefined are
observable by key enumeration, so fixed native layouts need presence information,
not a blanket insertion of undefined properties. Shapes observed in the stock
package use its service allocator prototypes; compiler-source objectAllocator uses
the base constructors above. Their dumped semantic fields agree, but prototype and
key-reflection identity is a separate contract, not established by the dump.

Go ast/ast.go:180 stores Kind/Flags/Loc/Parent plus nodeData, with generated typed
accessors; newNode at 75 sets a typed payload and hooks. This avoids generic
half-constructed JS object casts by representing node variants explicitly. It does
not justify treating tsc's unsound generic cast as a proven type in Adamic.

JSDocParser starts at parser.ts:8790; parseJSDocCommentWorker is at 8885. It has
nested scanner/context state, string-or-structured comments, links, tags, types and
JS-only diagnostics. Go parser.go:470/478 attaches diagnostics and a JSDoc cache;
479-481 explicitly enables lazy JSDoc for non-JS files. Tsc's dump visits attached
JSDoc before ordinary children. A Go port must force the corresponding docs before
comparison. Ordinary forEachChild omits attached node.jsDoc; that was a proved
blind spot in the old driver. The existing tag and diagnostic mutants still fail.

# Recursion measured on Node

[measure-depth.cjs](measure-depth.cjs) instruments 496 block-bodied functions in
stock 6.0.3's bundled Parser IIFE, callbacks included, using balanced try/finally
entry/exit. It counts active parser functions, excluding scanner/factory/runtime
frames. It independently traverses both instrumented and uninstrumented trees
iteratively and compares kind/range/flags/text/JSDoc comment/tag and diagnostic
projections per input. It does not compare every hidden cache/property or claim
that instrumented stack thresholds equal uninstrumented V8 thresholds.

All 81 reconstructed compiler inputs and the mandatory 10,406 pinned single-file
cases completed without instrumentation errors. Deepest results:

| Input | Active parser functions | Tree height including JSDoc |
| --- | ---: | ---: |
| src/compiler/transformers/module/module.ts | 205 | 24 |
| tests/cases/compiler/parsingDeepParenthensizedExpression.ts | 859 | 285 |
| tests/cases/conformance/types/specifyingTypes/typeLiterals/parenthesizedTypes.ts | 468 | 45 |
| tests/cases/compiler/deeplyNestedConditionalTypes.ts | 324 | 104 |
| tests/cases/compiler/uncalledFunctionChecksInConditionalPerf.ts | 306 | 54 |

[depth.json.gz](evidence/depth.json.gz) retains each file's maximum, deepest active
call chain and projection check outcome. Multi-file exclusions (2,039 cases),
project/option matrices and unit tests that generate inputs dynamically are outside
this measurement. Thus 859 is the maximum of the mandatory single-file inputs,
not an assertion about every input TypeScript's entire test harness can generate.

[measure-stack.py](measure-stack.py) measures fresh processes at RLIMIT_STACK
8,388,608 bytes, core dumps disabled. Uninstrumented stock Node parses a synthetic
parentheses input through depth 679 and gives RangeError at 680. Native recursion.a
succeeds through 62,803 and exits 70 at 62,804 with
`adamic: panic: RangeError: Maximum call stack size exceeded`. clang -O2 reports
an 88-byte static descend frame. [stack.json](evidence/stack.json) holds every trial.
These are different functions, so their numeric ratio is not a parser speed or
capacity result. The runtime stack.c guard reserves a quarter of the stack plus
256 KiB on an 8 MiB stack, leaving about 5.75 MiB. Even 859 * 88 bytes is only a
calibration: parser C has other frames, callbacks, scanner and factory calls, and
sanitizers enlarge frames. Actual parser frame sizes and peak native stack remain
unmeasured because the parser build does not reach a binary.

# A first shape that needs no language ruling

[dumpdiff](dumpdiff/compare.go) is a separate Go package and CLI, with no compiler
changes. It compares complete records byte for byte, preserving final newlines,
UTF-16 escapes, empty fields, file names, flags and both diagnostic families. Exit
0 means equal; 1 means a first difference; 2 means CLI/I/O failure. It reads only
through the first differing record, retaining two records and buffered readers.
Paths are zero-based preorder indices including JSDoc, not guessed AST parent edges.
The existing dump contains no hierarchy. A future optional oracle sidecar can map
these indices to child-edge paths without changing acceptance bytes.

```
go build -o /tmp/step24-dumpdiff ./stage3/scouts/step24/dumpdiff/cmd
/tmp/step24-dumpdiff reference.dump actual.dump
```

On the actual successful end mutant it reports line 166 and
`_namespaces/ts.ts/preorder/150`: Identifier end 3145 versus 3146. Equal self-comparison
exits 0. Five alternating warm-cache runs against diff -u on the same 36 MB files
gave bests 0.001451s versus 0.056490s. This is an early-mismatch
measurement, not a general speedup claim. [comparator-benchmark.json](evidence/comparator-benchmark.json)
retains all trials and [dumpdiff-0.log](evidence/dumpdiff-0.log) the actual report.

# Fixtures, mutants, commands and uncovered work

[fixtures.json](evidence/fixtures.json) records exact source Node outputs, diagnostics,
backend and native comparisons and mutants. context.a prints 14, 13, 13, 13 on
source Node, emitted JS and sanitized native, exits 0, and has empty stderr.
recursion.a at depth 8 prints 4 and passes those comparisons. Counted builds have
matching stdout, with deterministic counts in [counts.md](counts.md). No fixture
was added to internal/oracle discovery, so its counts.md is not changed.

speculation.a's Node witness completes, including nested rewind and Reparse,
but native refuses at its real diagnostic-array length assignment with NotYet
`assigning a field of a value`. It carries no expected-refusal header: NotYet
is not Refused. factory-cast.a carries the measured `a-check: refused a cast the
runtime can't check` header. No source/proof bypass was used to compile either.

Every mutant run and its catcher:

| Mutant | Observed catcher |
| --- | --- |
| Comparator ignores records | node end/flags/text/header/diagnostic/JSDoc/newline and long-record tests fail |
| Comparator ignores EOF | truncated and extra-file tests fail |
| Comparator ignores final newline | newline test fails |
| Comparator keeps previous file's node index | long-record/file-reset test fails |
| Context restores cleared bit as false | source Node and successful sanitized mutant native stdout differ |
| Context resets all flags, losing cached bit | source Node and successful sanitized mutant native stdout differ |
| Speculation replaces token rewind with 99 | successful source Node stdout differs |
| Speculation restores pending error as true | successful source Node stdout differs |
| Speculation omits diagnostic truncation | successful source Node stdout differs |
| Reparse truncates diagnostics unconditionally | successful source Node stdout differs |
| Factory a-check reason changed | captured real refusal/header comparison rejects it |
| Factory a-check header removed | captured real refusal/header comparison rejects it |
| Instrumentation changes finishNode.end | uninstrumented Node tree projection comparison fails, exit 1 |
| Instrumentation observes no depth | positive-frame check fails, exit 1 |
| Real parser Identifier.end +1 | both Node runs finish; exactly one dump line changes; cmp and dumpdiff exit 1 |
| Real parser drops JSDoc tags at return site | both Node runs finish; 3,846 tags become zero; cmp exits 1 |
| Real parser changes JSDoc diagnostic 1110 to 1111 | both Node runs finish; exactly one diagnostic row changes; cmp exits 1 |
| Census signature/body-range mutant | dedicated audit rejects the wrong body skip |
| Census extra latent finding and misattribution | dedicated audit checks delta 1 at one.a:3:1, delta 0 in two.a, catches attribution mutant |

Comparator mutants all fail behavior assertions, not builds or clang warnings.
Header mutants exercise the local captured-diagnostic contract, not a fresh run
of cohere's Gate.aCheck. These boundaries are explicit in the runner.

Exact successful commands, all output redirected to logs:

```
export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh > /tmp/step24-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/step24-adapted > /tmp/step24-apply.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/step24-latent-overlay > /tmp/step24-latent-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/step24-latent-overlay/overlay.json -o /tmp/step24-latent ./stage3/census/latent/tool > /tmp/step24-latent-build.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/step24-latent /tmp/step24-adapted/src/compiler /tmp/step24-latent.jsonl > /tmp/step24-latent-run.log 2>&1
python3 stage3/census/latent/audit.py /tmp/step24-latent > /tmp/step24-latent-audit.log 2>&1
export PARSER_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
node stage3/drivers/parser/closure.cjs /tmp/step24-adapted stage3/scouts/step24/evidence/file-closure.json > /tmp/step24-file-closure.log 2>&1
node stage3/drivers/parser/value-closure.cjs /tmp/step24-adapted stage3/scouts/step24/evidence/value-closure.json > /tmp/step24-value-closure.log 2>&1
node stage3/scouts/step24/research.cjs /tmp/step24-adapted stage3/scouts/step24/evidence/constructs.json > /tmp/step24-constructs.log 2>&1
node --max-old-space-size=2048 stage3/scouts/step24/measure-depth.cjs /tmp/step24-corpus /tmp/step24-depth-final.json > /tmp/step24-depth-final.log 2>&1
python3 stage3/scouts/step24/check-fixtures.py > /tmp/step24-fixtures.log 2>&1
python3 stage3/scouts/step24/measure-stack.py > /tmp/step24-stack.log 2>&1
bash stage3/drivers/parser/run.sh /tmp/step24-adapted /tmp/step24-parser-final --inputs /tmp/step24-corpus > /tmp/step24-parser-final.log 2>&1
go test ./stage3/scouts/step24/dumpdiff/... -count=1 > /tmp/step24-dumpdiff-final-test.log 2>&1
python3 stage3/scouts/step24/dumpdiff/mutants.py > /tmp/step24-comparator-final-mutants.log 2>&1
python3 stage3/scouts/step24/benchmark-diff.py > /tmp/step24-benchmark.log 2>&1
python3 stage3/scouts/step24/package-evidence.py > /tmp/step24-package-final.log 2>&1
go vet ./stage3/scouts/step24/dumpdiff/... > /tmp/step24-vet.log 2>&1
```

Corpus reconstruction uses the local pinned bare mirror, 00-setup, current
10-type-imports and the upstream diagnostic generator; it is separate from full
apply. Its manifest and per-file hashes are preserved. The first parser attempt
overlapped incomplete apply and is discarded. A census invocation before its
binary finished building exited 127 and was repeated after the build completed.
Initial fixture console arguments were corrected to strings before validation;
emitted JS needed the repository runtime resolver. Initial stack-frame compilation
needed the runtime include path. The instrumentation's first root discovery failed
because the bundle IIFE is nested, and was corrected with AST traversal. Failed
commands are observations, not green claims.

The actual current-main CLI attempt `adamic build /tmp/step24-adapted/parser-proof-main.a
-o /tmp/step24-native-parser` exits 1 with 320 checker diagnostics, first at
builder.ts:1246:69 (TS2345, Path | undefined passed to string). This is the full
adapted driver/file graph, not a new validated declaration-slice blocker order.
No parser C or binary was emitted. [native-parser-build.log.gz](evidence/native-parser-build.log.gz)
retains the complete diagnostics.

The completed run.sh dump is **36,429,244 bytes**, SHA256
8c158aec1978cadfcc0fd078b4e23a748a2badfa63458cfbeda1f252d80d8537, with 81 inputs,
4,149 JSDoc nodes and 3,846 tags. It does **not** equal reference.json's
36,429,231 bytes / 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
The adaptation-10 compiler-source parser on these same reconstructed inputs also
produces the fresh 36,429,244 bytes, byte-identical to the current-main parser.
This separates the mismatch from current full adaptations, but the exact historical
input difference is unresolved without the old full dump/corpus bytes. The reference
is not changed. No native acceptance, new recovery-case hashes, Go differential,
full stage3 upstream lane, full compiler package tests or repository gate is claimed.

# Options within Adamic's rules, undecided

1. **Factory result proofs.** One option is per-constructor/per-kind completion
   analysis proving every promised field initialized before the subtype escapes;
   another is source builders returning a closed discriminated node union, with
   generic dispatch resolved to its actual completed branch. Both must keep
   readiness checks and reject the factory-cast witness. Blanket `as T` admission,
   any and fictional defaults are outside the rules. Open numeric enum equality
   alone is not a unique variant proof. Existing native ownership stays in force.
2. **NodeArray metadata.** One option is a fixed typed array representation whose
   metadata/presence slots are declared at allocation; another is a typed wrapper
   plus an explicit source adaptation proved against all callers. Both must
   preserve holes, alias identity, enumerable metadata and writes in order.
   General dynamic expandos and prototype mutation remain forbidden.
3. **Speculative state and ownership.** One option implements current scalar/slice
   writes and lets ordinary reference counting reclaim abandoned temporaries;
   another introduces an explicit transactional state value plus scoped arenas
   only where escape/freshness analysis proves safe. Reparse diagnostics and
   returned lookahead values must keep Node semantics. Neither permits a GC,
   unchecked widening, or freeing a returned node at rewind. Generic callback
   falsiness needs exact Node behavior or a separately proved specialization.
4. **Recursive ownership.** One option proves fresh child writes and keeps children
   immutable after completion, with any parent/back edges weak and strongly owned
   targets surviving observations; another uses already-designed region lifetimes
   that own complete trees and prove every escape. Removing parent links for the
   disabled-parent acceptance path does not justify changing enabled-parent APIs.
   Generic mutable arrays must remain invariant; structural callback methods may
   not gain bivariance as a shortcut.
5. **Deep recursion.** One option retains recursive C with measured frame budgets
   and the existing loud stack guard; another lowers selected deep recursive grammar
   paths to explicit continuation stacks, preserving evaluation and cleanup order.
   The existing overflow-depth exception allows native and Node thresholds to
   differ, but not silent bad output or a bare crash. No new depth cap or rewritten
   grammar is selected by this scout. JSDoc laziness is a performance representation
   choice only if the externally observed nodes/diagnostics retain Node behavior.

Questions for @system_adamic, left undecided:

- Which proof surface should establish completed factory subtypes: constructor
  completion analysis, closed node-union builders, or another verified contract?
- Should declared NodeArray metadata be part of a fixed array type, or require a
  source wrapper with a proved adapter? How should absent versus present-undefined
  keys be represented without admitting arbitrary expandos?
- Should generic speculation retain exact falsiness through checked result
  specialization, or move to an explicit typed commit/rewind result with a source proof?
- Should parser tree ownership first extend fresh-write/Weak proofs across factory
  helpers, or specialize existing region lifetimes for complete source files?
- Is guarded recursive C the initial target once actual parser frames are measured,
  or should the deepest grammar paths first use explicit continuations? What stack
  budget must hold for release and sanitizer builds before native acceptance?

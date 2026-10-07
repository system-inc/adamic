# Scanner blockers

## October 7: corrected adaptations pass the full baseline

Adaptations 52, 53 and 54 pass the unfiltered stage 3 baseline: **106,367
passing, zero failing, zero pending, zero baseline differences**. Install,
build and tests exit 0; total 259.105 seconds. No changed snapshots accepted.
The generic undefined-entry regression and public API snapshot changes from
the earlier attempt are fixed. Node still matches all 509,014 tokens. See
[member-baseline-revised-report.json](evidence/member-baseline-revised-report.json)
and [member-baseline-revised-tests.log](evidence/member-baseline-revised-tests.log).

The current tracked slice still stops at Debug.fail function widening
(method-signature-style). Untracked probes omit V8 stack capture and the
assertion contract only to discover subsequent diagnostics. After guarding
both languageVersion comparisons and the shebang regex result, the next is:
`scanner.ts:483:16: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case`.
This is `var text = textInitial!`, where omitted initial text is a supported
createScanner input. Census reason: `the non-null assertion !`. There is no
validated native executable. Do not confuse the successful Node slice proof
with native execution.

Independent verifier passes 91 source spans and rejects the declaration
omission mutant. Restoring private diag's return annotation is caught by
TS2375. Clearing the Unicode table is caught by the explicit indexed-read
guard. Off-by-one token end and declaration omission remain caught by the
Node comparison/ReferenceError checks. Source closure and full declaration
ledger are in stage3/slice/evidence/member-scanner-{closure,manifest}.json.

## October 7: shebang optional read needs an explicit bound

Replacing the regex assertion with optional indexed access in the untracked
probe returns to the checker gate:
`scanner.ts:434:17: TS18048: 'shebang' is possibly 'undefined'`.
Census reason: unchecked indexed reads. The subsequent shebang.length needs
the regex match and its first element to be proven present. A successful regex
match is guaranteed by this function's initial #! test, but the checker does
not derive that regex fact. No closing feature branch demonstrated.

## October 7: comparison-preserving languageVersion guard, next non-null site

In the untracked probe, replace the two `languageVersion! >= ES2015`
comparisons with `languageVersion !== undefined && languageVersion >= ES2015`.
Undefined still selects the older Unicode map. Next exact refusal:
`scanner.ts:433:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case`.
This is shebangTriviaRegex.exec(text)![0]. Census reason: `the non-null assertion !`.
No demonstrated closing feature branch. This probe has not been tracked as
an adaptation or presented as a native token proof.

Count correction: the raw slice has 89 declarations (39 function declarations,
22 variable statements, 12 enums, nine interfaces, seven type aliases) and
**two** namespace-wrapper spans, not four. Total remains 91 copied spans,
7,127 lines and eight files. Original counts of 87 plus four were a reporting
error. The final manifest and independent byte verifier use the correct counts.

## October 7: full baseline rejects the first checker adaptations

Full unfiltered baseline: 106,312 passing, 55 failing, zero pending, seven
baseline differences. Install/build pass. Tests exit 1 in 232.685 seconds.
Adaptation 52's guarded generic forEach rejects legitimate undefined entries
in other compiler users. Adaptation 53's explicit undefined unions alter the
public API declaration snapshot. These are adaptation regressions, not a
native silent miscompile; no native executable was produced. The slice token
oracle passed, demonstrating that it alone did not cover these cases.

The revised README-only plans are pushed before corrective implementation:
use entries iteration for forEach without rejecting undefined values; infer
private diag's actual return type and keep the public interface unchanged.
Neither first version is approved by the baseline. Full details are in
[member-baseline-first-report.json](evidence/member-baseline-first-report.json).

## October 7: removing the assertion contract exposes non-null assertions

In the same untracked diagnostic copy, changing only `asserts expression` to
`void` leaves the checker clear and reaches:
`scanner.ts:291:12: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case`.
This is `languageVersion! >= ScriptTarget.ES2015`. Census reason:
`the non-null assertion !`. No closing feature branch was demonstrated. Undefined
languageVersion is legitimate here: JavaScript's comparison is false and selects
the older Unicode map. Replacing it with a panic would change that behavior.

Other non-null uses in the reached scanner include textInitial!, regex exec,
codePointAt, and the intentional assignment `tokenValue = undefined!`. They
are not interchangeable checked-read repairs. In particular, blindly replacing
undefined! with an empty string would change observable scanner state. These
remain unadapted; the scratch assertion-contract and V8 omissions are not
included in the tracked adaptations. No native binary exists yet.

## October 7: scratch omission of V8 capture exposes assertion contract

An untracked scratch copy removes only the captureStackTrace if-block from
Debug.fail after removing debugger. Next exact refusal:
`debug.ts:15:142: Adamic 0.1 refuses a type predicate; narrow where you use it, with ===, typeof or instanceof (adamic/no-type-predicate)`.
This is Debug.assert's `asserts expression` return contract. Census reason:
`a type predicate`. No demonstrated closing feature branch. The V8 block
omission is a diagnostic probe only; no adaptation drops failure-stack behavior
on the deliverable branch. No assertion predicate rewrite is authorized by
adaptation 54's published plan, so it has not been made there.

## October 7: debugger removed, function widening is next

Temporary 54 removes only debugger from the reached Debug.fail member. Next:
`debug.ts:13:67: Adamic 0.1 refuses a function taking string | undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)`.
This is the `stackCrawlMark || fail` value passed to V8's captureStackTrace.
Census reason: method-signature-style. No demonstrated closing feature branch.
Removing this failure-stack-only registration is a scratch probe, not yet a
validated adaptation. The scanner calls Debug.assert/assertEqual without any
stack crawl mark or verbose callback; it never calls formatEnum. Successful
tokenization does not exercise assertion failure stack formatting.

## October 7: import cycles closed, next refusal is debugger

After integrating `codex/import-cycles` at `6b17636` into the unpushed scratch
compiler, the cycle refusal closes. The next exact diagnostic is:
`debug.ts:10:9: Adamic 0.1 refuses debugger; remove it`.
Census reason: debugger. No closing feature branch was found. Temporary 54's
slice-only plan is published with this blocker before implementation. It
covers only the reached Debug.fail member and first removes debugger. V8's
captureStackTrace and type contracts remain separate probes, not assumed fixed.

## October 7: checker clear in the adapted slice, first lowering blocker

Main plus fallthrough alone removes ten TS7029 findings (34 to 24). Newest
four-feature tips plus fallthrough leave eleven checker findings. Adaptations
52 and 53, planned and pushed at `21cad99` before edits, remove those eleven.
Adapted Node output still matches every byte of the 509,014-token full-tree
oracle. Both scripts require slice.json and reject full-tree application.
Full baseline validation is pending; these scripts are provisional probes.

First lowering refusal, exactly:
`core.ts:1:1: Adamic 0.1 refuses an import cycle; move what both modules need into a third that neither imports`.
Census reason: module cycles. Closing feature: `codex/import-cycles`, remote
`6b17636`, now being fetched. This is the type import from Debug back to core's
AnyFunction plus core's value import of Debug. Node initialization passes.
Logs: [member-first-lowering.log](evidence/member-first-lowering.log),
[member-current-checker.log](evidence/member-current-checker.log), and
[member-fallthrough-only.log](evidence/member-fallthrough-only.log).

Scanner omission mutant completed: delete the reached
unicodeESNextIdentifierStart declaration. Node exits 1 with ReferenceError
at isUnicodeIdentifierStart during scanIdentifier; the unmutated control exits
0 and matches the full token oracle. See stage3/slice/evidence/scanner-omission.stderr.

## October 7: namespace member slice is green on Node

Commit `c4e011f` gathers namespace members independently. Eight source files,
89 declarations and two original namespace-wrapper spans, 7,127 copied-span
lines. All spans match source bytes. Node prints exactly 509,014 tokens, with
empty diff against the full-tree oracle and SHA-256
`c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f`.
Debug initialization no longer fails. Debug reaches only isDebugging, fail,
assert and assertEqual; no enum formatter or registration member is needed.

Current-main checker and prior four-feature scratch checker stop before
lowering. Ordered diagnostics are [member-main-checker.json](evidence/member-main-checker.json)
and [member-features-checker.json](evidence/member-features-checker.json).
The latter has 21 findings: bounded indexed reads / TS2532 and generic indexed
reads / TS2345, exact optional declarations / TS2375, and intentional switch
fallthrough / TS7029. Main additionally rejects enums and namespaces / TS1294.
Closing branches: flag-enums and namespaces-tsc for TS1294;
fallthrough-and-implicit-returns for TS7029. Indexed reads and exact optional
properties have no demonstrated closing feature in this probe.

The remote fallthrough feature now exists at `5f77d33`; current main is
`e011f8f`. A main-plus-fallthrough-only scratch compiler is being built before
any adaptation decision. Earlier whole-namespace failure reports below are
superseded for the member slice, but remain historical evidence.

## October 7: declaration slice, latest observation

The checker-driven tool is in `stage3/slice/` and was pushed independently at
`b7dda00` before this follow-up evidence. Final extraction from adaptations 10
and 50, excluding 20, reaches **4,437 declaration/export records, 190,148
copied-span lines, 78 files**. Every copied span passes byte comparison against
its source. Counts include whole namespaces and module export facades.

Ordered observations:

1. Whole `Debug` in `src/compiler/debug.ts` contains `(ts as any)[enumName]`.
   This reaches the namespace barrel and all exported declarations. Keeping
   top-level namespaces whole defeats the intended small scanner slice. No
   compiler feature branch fixes this gathering-granularity issue. Namespace
   member gathering requires clarification; it is not implemented.
2. Node fails before emitting tokens: `src/compiler/binder.ts` creates the
   binder at module initialization, reaching `createFlowNode` and reading
   `Debug.attachFlowNodeDebugInfo` while `Debug` is undefined. The stack's line
   180 is in emitted JavaScript, not an upstream TypeScript location. Census
   category: module cycles / value read at load time. The existing cycle work
   refuses load-time reads; it does not close this new import-order problem.
   This is an oracle failure, not an observed Adamic refusal.
3. Main stops at the checker gate. The four-feature scratch compiler also
   stops, with **2,652 slice-file diagnostics**. Exact returned order and slice
   TypeScript locations are in [slice-feature-checker.json](evidence/slice-feature-checker.json).
   First: `binder.ts:1134:13`, TS2322, `FlowNode | undefined` is not assignable
   to `FlowNode`. This is strict optionality, with no demonstrated closing
   feature branch. TS7030 and TS7029 remain enabled; their proposed closing
   branch is `codex/fallthrough-and-implicit-returns`. Enum TS1294 on main is
   covered by `codex/flag-enums`. Other diagnostics have not been individually
   attributed to census reasons or closing branches.

No slice token stream exists to compare with the full-tree 509,014-token
oracle. No lowering was reached; no ordered Refused/NotYet sequence is claimed.
I did not run the one-feature-at-a-time matrix for this slice or repair the
remaining checker gate. The checked-in diagnostic list is checker evidence,
not a complete feature blocker census.

Mutant: deleting reached `compareComparableValues` from a smaller `compareValues`
slice causes Node `ReferenceError: compareComparableValues is not defined`.
Control output is four lines: -1, 1, 0, -1. This proves declaration omission is
observable for that entry; the scanner-specific omission mutant remains undone
because the scanner control already fails initialization.

No new temporary slice adaptations are planned or implemented in this update.
Previous full-tree adaptation 50 is used only as the requested input baseline;
51 remains deferred. Any future adaptation 50-59 must have its slice-only plan
pushed before implementation. No broad namespace or cycle rewrite was made.

## October 7: full baseline complete, latest result

Temporary 50 passes the unfiltered stage 3 baseline: **106,367 passing, zero
failing, zero pending, zero baseline differences**, total 347.707 seconds.
Install, build and tests all exited 0; tests took 318.641 seconds. See
[baseline50-report.json](evidence/baseline50-report.json) and the adjacent phase
logs. The fresh-tree driver path also ran apply (excluding adaptation 20), Node,
the comparison control, the end mutant and the native build. Node and the
mutant check passed; native build exited 1 with the same 2,652 diagnostics.
[fresh-report.json](evidence/fresh-report.json) records that failure explicitly.

Harness syntax and evidence ordering/count audits pass. `git diff --check`
passes. No native scanner execution occurred.

## October 7: after temporary 50, newest observation

**Stopped at the checker gate. This is not an exhaustive ordered Refused/NotYet
list.** The integrated scratch compiler returns 2,652 checker diagnostics after
temporary 50, down from 2,654 before it. No scanner corpus entry reaches lowering.
All diagnostics, their complete message chains, locations, census code, and
feature attribution are in [after50-checker.json](evidence/after50-checker.json),
numbered in the exact returned order. The first remains:

`src/compiler/binder.ts:1109:17: TS2412: Type 'undefined' is not assignable to type 'FlowNode' with 'exactOptionalPropertyTypes: true'.`

I did not successively repair the remaining 2,652 checker findings, and therefore
did not discover the complete lowering/ownership blocker sequence. Safely
repairing optionality, indexed-read density, generic contracts, and all other
compiler files is beyond this run. These are not interchangeable mechanical
edits. No checker option was weakened, no checker-rejected program was passed
to lowering, and no production compiler change is on the deliverable branch.

Temporary 50 inserts only explicit `return undefined;` into getShebang and
scanIdentifier in scanner.ts. Both original scanner TS7030 findings disappear.
The same scanner outputs 509,014 token lines on the fixed pre-edit corpus,
byte for byte identical to the Node control. Removing only getShebang's added
return in scratch brings back scanner.ts:966:43 TS7030. Reapplying the adapter
reports zero files and zero returns changed.

The required README-only plan was committed and pushed as
`586caf413c8235facd82c9aa63d8c19314fd1029` before either adapter script was written.
Temporary 51 is **deferred**: its runnable placeholder edits no source and reports
that all 13 scanner fallthroughs still require review. It is not a completed
adaptation or a claim of baseline validation for a fallthrough rewrite.

The completed unfiltered stage 3 oracle result is recorded at the top.

## October 7: adaptation 10 and integrated features

[before50-checker.json](evidence/before50-checker.json) retains the 2,654 diagnostics
from the same driver with adaptation 10 only and regenerated diagnostics. An
earlier direct scanner entry saw 2,655 because apply generates diagnostics before
adaptation 10 edits its generator. The runner now regenerates that owned output,
removing the generated-file TS1484; this is upstream generation, not a temporary
source rewrite. Adaptation 20 was excluded after the user's instruction; the
initial adaptation-20 experiment is not used as the delivered proof.

The Node oracle passed on all 81 files: 78 TypeScript sources, including generated
diagnostics, plus diagnosticMessages.json, diagnosticMessages.generated.json and
tsconfig.json. It printed 509,014 token lines. Native compilation exited 1 at the
checker gate. The copied comparison control passed diff (exit 0). Incrementing
only the first Node token's end from 76 to 77 failed that same diff (exit 1):

```text
-ExportKeyword  0  70  76  1  "export"
+ExportKeyword  0  70  77  1  "export"
```

The actual output uses tabs. Full catch:
[end-mutant.diff](evidence/end-mutant.diff). This is comparison sensitivity,
not a native correctness claim. Token output byte counts and SHA256s are in
[token-equivalence.json](evidence/token-equivalence.json).

## Scratch compiler and feature coverage

Started from origin/main `ef3d907ecdc4c771b016f7d9c52372def057a340` on
codex/stage3-scanner-proof. Feature changes were merged only into
`scratch/scanner-proof`, never pushed. The four fetched feature tips were:

| Feature | Fetched commit |
|---|---|
| codex/taste-not-soundness | `aa896b5d5ccc82210184fd01b8fe4d0ce0730a50` |
| codex/flag-enums | `f7d62772fa8e52fcfae754e047ceb66dba88b782` |
| codex/namespaces-tsc | `ce8a2acf14e420a9c82345236845a377cd4c7a50` |
| codex/nested-functions | `b15216dabf65ffaa7152f6e64709b7b062ea01a9` |

Scratch merged them in that order. Enum conflicts retained main's nominal and
accessor checks plus enum checks. Namespace conflicts retained flag-enum
validation and added namespace traversal/parameter properties. Nested-function
conflicts retained namespace-qualified calls and added recursive sibling calls
and captured-binding checks. Scratch head after merges:
`606698b` (these resolutions are not proposed compiler patches).

The shared cohere submodule required `go build -buildvcs=false`. The integrated
compiler build passed. No full feature-integration gate was run; its scanner
checker observations establish only the measured gate.

These are **source candidates**, not encountered Refused/NotYet observations.
The locations remain in stage3/census/data/sites.json, read in full and filtered
by the resolved closure. The feature column identifies the intended owner,
not proof that every real usage already compiles:

| Census reason | Closure sites | Intended closing feature |
|---|---:|---|
| an ExportDeclaration | 77 | codex/taste-not-soundness; export-star barrels still need module integration |
| enum | 164 | codex/flag-enums; individual enum shapes may still block |
| non-boolean control condition | 6697 | codex/taste-not-soundness |
| type assertion sites | 6231 | No closing feature established in this run |
| a function inside a function (a closure) | 5574 | codex/nested-functions; capture/generic cases remain to be encountered |
| the non-null assertion ! | 1123 | No closing feature established in this run |
| ||= | 110 | codex/taste-not-soundness |
| a type predicate | 651 | No closing feature established in this run |
| explicit any | 210 | No closing feature established in this run |
| a namespace | 11 | codex/namespaces-tsc; its documented unsupported shapes remain |
| in | 10 | No closing feature established in this run |
| the void operator | 15 | codex/taste-not-soundness |
| the comma operator | 47 | codex/taste-not-soundness |
| yield (generators) | 16 | No closing feature established in this run |
| a label | 10 | codex/taste-not-soundness |
| an index signature | 11 | records feature, not one of the four merged branches |
| a spread after the first field | 17 | No closing feature established in this run |
| Record<string, T> | 5 | records feature, not one of the four merged branches |
| delete | 2 | No closing feature established in this run |
| debugger | 1 | No closing feature established in this run |
| a definite assignment assertion ! | 12 | No closing feature established in this run |
| &&= | 1 | codex/taste-not-soundness |

Module cycles are a separate unresolved integration requirement. The merged
`internal/lower/modules.go` still returns Refused "an import cycle" at its DFS
back edge. That is an inspected implementation fact, not an observed scanner
lowering result. None of the four fetched branches closes it. The Node direct
scanner-root experiment also observed a real ESM load-time value read in
parser.ts (`textToKeywordObj`), so simply dropping cycle refusal would not prove
initialization correct.

## TS7030 and TS7029 while both options stay on

Before temporary 50 there are 252 TS7030 and 83 TS7029 findings in this closure.
After it there are 250 and 83. Intended closing feature for both:
`codex/fallthrough-and-implicit-returns`. It was not part of the requested four
feature merges. NoImplicitReturns and NoFallthroughCasesInSwitch remained true.
The scanner-local observations before edits, in returned diagnostic order:

| Location | Census reason | Temporary plan |
|---|---|---|
| src/compiler/scanner.ts:1553:13 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:1563:13 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:1686:13 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:2133:17 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:2425:14 | TS7030 | 50, implemented |
| src/compiler/scanner.ts:2765:21 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:2838:21 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:2907:17 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:3300:17 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:3480:17 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:3861:13 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:444:13 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:651:13 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:845:13 | TS7029 | 51, deferred |
| src/compiler/scanner.ts:966:43 | TS7030 | 50, implemented |

Every remaining location across the closure is retained in the ordered checker
JSON, including all full chains and the closing-feature field. Optional-property
codes point at the excluded adaptation-20 branch as a partial prerequisite;
other code buckets have no closing feature established by this run.

## Whole source import closure

Resolved from the adapted scanner.ts with stock TypeScript 6.0.3's resolver.
Includes source imports and re-exports, including type-only module requests;
excludes ambient standard-library and Node declaration files from the source
ownership list. There are 78 source files and 164 edges, all retained with
locations in [adaptation10-closure.json](evidence/adaptation10-closure.json).
The graph reaches all compiler sources through `_namespaces/ts.ts`, so the
scanner and parser source closures overlap. This run edits scanner.ts only.
There is no disjoint "rest of compiler" source closure for the parser worker.

- src/compiler/_namespaces/ts.moduleSpecifiers.ts
- src/compiler/_namespaces/ts.performance.ts
- src/compiler/_namespaces/ts.ts
- src/compiler/binder.ts
- src/compiler/builder.ts
- src/compiler/builderPublic.ts
- src/compiler/builderState.ts
- src/compiler/builderStatePublic.ts
- src/compiler/checker.ts
- src/compiler/commandLineParser.ts
- src/compiler/core.ts
- src/compiler/corePublic.ts
- src/compiler/debug.ts
- src/compiler/diagnosticInformationMap.generated.ts
- src/compiler/emitter.ts
- src/compiler/executeCommandLine.ts
- src/compiler/expressionToTypeNode.ts
- src/compiler/factory/baseNodeFactory.ts
- src/compiler/factory/emitHelpers.ts
- src/compiler/factory/emitNode.ts
- src/compiler/factory/nodeChildren.ts
- src/compiler/factory/nodeConverters.ts
- src/compiler/factory/nodeFactory.ts
- src/compiler/factory/nodeTests.ts
- src/compiler/factory/parenthesizerRules.ts
- src/compiler/factory/utilities.ts
- src/compiler/factory/utilitiesPublic.ts
- src/compiler/moduleNameResolver.ts
- src/compiler/moduleSpecifiers.ts
- src/compiler/parser.ts
- src/compiler/path.ts
- src/compiler/performance.ts
- src/compiler/performanceCore.ts
- src/compiler/program.ts
- src/compiler/programDiagnostics.ts
- src/compiler/resolutionCache.ts
- src/compiler/scanner.ts
- src/compiler/semver.ts
- src/compiler/sourcemap.ts
- src/compiler/symbolWalker.ts
- src/compiler/sys.ts
- src/compiler/tracing.ts
- src/compiler/transformer.ts
- src/compiler/transformers/classFields.ts
- src/compiler/transformers/classThis.ts
- src/compiler/transformers/declarations.ts
- src/compiler/transformers/declarations/diagnostics.ts
- src/compiler/transformers/destructuring.ts
- src/compiler/transformers/es2015.ts
- src/compiler/transformers/es2016.ts
- src/compiler/transformers/es2017.ts
- src/compiler/transformers/es2018.ts
- src/compiler/transformers/es2019.ts
- src/compiler/transformers/es2020.ts
- src/compiler/transformers/es2021.ts
- src/compiler/transformers/esDecorators.ts
- src/compiler/transformers/esnext.ts
- src/compiler/transformers/generators.ts
- src/compiler/transformers/jsx.ts
- src/compiler/transformers/legacyDecorators.ts
- src/compiler/transformers/module/esnextAnd2015.ts
- src/compiler/transformers/module/impliedNodeFormatDependent.ts
- src/compiler/transformers/module/module.ts
- src/compiler/transformers/module/system.ts
- src/compiler/transformers/namedEvaluation.ts
- src/compiler/transformers/taggedTemplate.ts
- src/compiler/transformers/ts.ts
- src/compiler/transformers/typeSerializer.ts
- src/compiler/transformers/utilities.ts
- src/compiler/tsbuild.ts
- src/compiler/tsbuildPublic.ts
- src/compiler/types.ts
- src/compiler/utilities.ts
- src/compiler/utilitiesPublic.ts
- src/compiler/visitorPublic.ts
- src/compiler/watch.ts
- src/compiler/watchPublic.ts
- src/compiler/watchUtilities.ts

## Reproduction and limits

Read CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the complete census REPORT
and stage3 README before edits. Read the census module edges/cycles, sites,
diagnostics, file inventories, string lookups and shape additions in full.
Setup log is retained; Go 1.27.1, clang 20.1.8, Node 24.19.0, all tool readiness
phases 0 seconds, build cache warmup/total 96 seconds, nproc 5, cgroup quota 4.

All command output was saved to files. Main measurements:

```sh
# From the integrated scratch checkout; compiler changes remain unpushed.
source /workspace/adamic-tools/env.sh
go build -buildvcs=false -o /workspace/scratch/scanner-adamic ./cmd/adamic
# From the deliverable checkout; --tree came from apply with adaptation 20 excluded.
stage3/drivers/scanner/run.sh /workspace/scratch/scanner-run10-final --tree /workspace/scratch/scanner-adapted10 --compiler /workspace/scratch/scanner-adamic
stage3/drivers/scanner/run.sh /workspace/scratch/scanner-run50 --tree /workspace/scratch/scanner-adapted50 --inputs /workspace/scratch/scanner-adapted10 --node-only
# From the scratch checkout; no filter, default runners, four workers.
stage3/oracle/run.sh /workspace/scratch/scanner-adapted50 /workspace/scratch/scanner-baseline50
```

The fixed input tree was still pre-edit when scanner-run50 ran. It was modified
only afterward for the native checker and return-removal mutant probes.

Not covered: a native scanner execution, native/Node equality, parser-driven
rescanning, the exhaustive ordered Refused/NotYet traversal, all 13 intentional
fallthrough source rewrites, or the complete uncached Adamic gate. No silent
miscompile was observed; checking prevented native compilation.

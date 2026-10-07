# Overload and marker fixes rerun: red at sameMap

Scratch d1ed400a contains proven-predicates 5588a3b and census-small-families
bde0030, plus all previous integration inputs. The scratch merge is unpushed.
Its one lowering conflict was verifier naming: keep proveFlowPredicate while
adding censusPredicateMarkerContract. No compiler proof bypasses or source
stubs are present in this rerun. SDK and loader compatibility remain scratch-only.

Validated parser-temp65-slice: C emission, split off, split on and repeat all
exit 1 at src/compiler/core.ts:195:37: sameMap's
`array.slice(0, i) as unknown[] as U[]` is a cast the runtime cannot check.
The real first stop is now row 2; rows 1 and 4 are cleared in the actual slice.
C size, clang time, binary size, native timings and both native references are
unavailable. Repeat has zero cache objects, so it is not a warm-cache result.

All eleven unchanged original probes were rerun: 3/11 pass, rows 1, 4 and 9.
Their native exit/stdout/stderr match Node byte for byte, and all three one-byte
native-output mutants are caught (cmp 1). The other eight remain compile refusals.
These three reduced probes admit declarations and print their existing markers;
they are not the full parser runtime witness. Row 10's original cast probe stays
unchanged and refused; adaptation 65 removes that cast in the validated slice.
This is not a regression of adaptation 65. Row 6 remains exactly as upstream
wrote it, pending the compiler SHA for the arguments.length ruling.

Predicate/overload focused tests pass (4.530s). The broader focused selection
has one failure: TestCensusPredicateMarkerKeepsProofBoundaries expects Refused
for a live predicate callback but receives a union-lowering NotYet. The exact
source and standalone reproduction are native-marker-live-boundary.a; its
build refuses at 4:12, and Node exits zero with boundary declaration loaded.
No compiler fix or weaker test expectation was applied. This diagnostic failure
is recorded separately from the parser's first source stop; no full gate claimed.

Evidence and replay: evidence/front11/{inputs,build,probes,boundary-test}.json
and compressed logs. probe-stops.py COMPILER SCRATCH NEW_OUTPUT reruns the eleven
sources, the strict comparisons and byte mutants. measure-builds.py reruns split
modes on the entry path in build.json. No native parser dump is claimed.

---

# Adaptation 65 validated; waiting for compiler implementations

65-temporary-node-process-cast is implemented and checked with library's loader
and @types/node 25.3.3. It removes the process as any cast and replaces the
local unknown-field ambient type with NodeJS.Process plus the existing optional
browser extension. All ten emitted JavaScript files are byte-identical, the
newest driver slice checker has zero diagnostics, parser/scanner references
are unchanged, and the sanctioned full stage 3 oracle passes 106,367 tests
with no baseline differences or new API exceptions. Exact evidence and replay
commands are in stage3/adapt/65-temporary-node-process-cast/README.md.

Row 6 has a ruling: arguments.length read alone is accepted with JavaScript's
exact value. reduceLeft remains unchanged; no adaptation is proposed for it.
Rerun native-arguments-length.a only when the compiler SHA implementing that
ruling arrives (the user's target is October 8 12:00). No new native build
was requested or claimed here. Resting until the compiler SHAs arrive.

---

# Adaptation 65 planned before implementation

65-temporary-node-process-cast removes `(process as any).browser` in
core.isNodeLikeSystem. Adaptation 30 currently adds a local `process` declaration
with unknown fields, so 65 also changes its type to
`(NodeJS.Process & { browser?: unknown }) | undefined`, preserving the optional
browser probe while sourcing the Node contract from pinned @types/node 25.3.3.
The official Process interface has no browser property; no fabricated Node
binding is substituted. This is valid only with library's Node-types loader.
All edits erase, and emitted JavaScript must be byte-identical. This temporary
moves into adaptation 40 once the loader lands on main. User explicitly
requested 65 for this shared-core site. No other discovery stub is sanctioned.

---

# Predicate callback fix rerun: red at the overload result

6feee11b is merged in unpushed scratch c40d2a4e. The previous callback minimal
now builds and runs native `true`, exit 0, empty stderr. Focused predicate and
checked-overload-result tests pass (0.683s). The validated slice still refuses
core.ts:54:116: every's outer `array is readonly U[]` result on a bodyless
overload has no body proving its parameter. The callback at column 101 is cleared.
Reduced native-predicate-overload-result.a reproduces at 3:109.

Unsplit, split and repeat builds all exit 1 at this same refusal, before C.
C bytes/lines, clang wall times, parser binary size, native user time and both
native reference comparisons are unavailable; no native output-byte mutant can
run. Cache contains zero objects; repeat is not warm. nproc 5, perf absent.
Same-input Node runs all match the 81-file reference (cmp 0), best user time
6.162363s. Reference: 36,429,231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
The committed 10,406-case manifest remains the conditional native acceptance
input; no native case run occurred, and its Node run was not repeated this turn.

## Ordered discovery behind stubs

Only row 1 is observed on the validated slice. All subsequent rows are found
behind explicit scratch-only bypasses, listed with exact diagnostics and eleven
Node-checked minimal programs in evidence/front10/build/report.json. The final
source patch and compiler-proof bypass are retained there as compressed evidence;
they are not source adaptations, do not enter the acceptance slice, and were
never committed to the scratch compiler. The real scratch predicate verifier
has been restored. No compiler edits are pushed on this proof branch.

1. `src/compiler/core.ts:54`: every overload result; minimal `native-predicate-overload-result.a`.
2. `src/compiler/core.ts:195`: sameMap generic array cast; minimal `native-generic-array-cast.a` (found behind a stub).
3. `src/compiler/core.ts:230`: flatMap empty fallback; minimal `native-generic-empty-array.a` (found behind a stub).
4. `src/compiler/core.ts:283`: some overload result; minimal `native-some-predicate-overload.a` (found behind a stub).
5. `src/compiler/core.ts:421`: toSorted brand cast; minimal `native-sorted-array-brand.a` (found behind a stub).
6. `src/compiler/core.ts:525`: reduceLeft arguments; minimal `native-arguments-length.a` (found behind a stub).
7. `src/compiler/core.ts:542`: hasOwnProperty detached method; minimal `native-has-own-property.a` (found behind a stub).
8. `src/compiler/core.ts:594`: isArray builtin predicate; minimal `native-array-is-array-predicate.a` (found behind a stub).
9. `src/compiler/core.ts:619`: predicate to AnyFunction view; minimal `native-predicate-function-view.a` (found behind a stub).
10. `src/compiler/core.ts:890`: process any cast; minimal `native-process-any-cast.a` (found behind a stub).
11. `src/compiler/debug.ts:26`: Debug namespace; minimal `native-debug-namespace.a` (found behind a stub).

Discovery stops at Debug's namespace: further continuation needs broad namespace
transformation, beyond the isolated probes used here. This is not an exhaustive
blocker list. The temporary source-only erasure of some's predicate caused
nodeFactory checker errors; those were probe artifacts, not new source blockers.
The original annotation was restored for subsequent probes. Namespace Node
execution uses the merged scratch oracle's transform mode; the strip-only
runner's unsupported-syntax attempt is also retained honestly in the report.

Replay builds with measure-builds.py and native-proof.py using the paths recorded
in report.json; the latter invokes both native references only after a binary
exists. Minimal builds use parser-front10-adamic from c40d2a4e plus the retained
SDK and predicateRefusal integration. Ordinary minima run with oracle/node.mjs;
the namespace minimal uses the scratch transform runner. No full gate claimed.
Native status remains syntax tree and parse diagnostics identical on Node,
JSDoc unverified, error recovery unverified. No new temporaries made.

---

# Newest-tip reference checkpoint: Node green, native red

The fresh full adapted tree and refreshed slice match every committed per-case
record: 10,406 single-file cases, 1,323,111 nodes, 1,996 parse diagnostic rows
and 34 JSDoc diagnostic rows. The 81-file extended reference is unchanged:
36,429,231 bytes, SHA256 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
The node-end, dropped-JSDoc-tags and JSDoc-diagnostic mutants are caught.
The freshly rerun JSX recovery mutant changes exactly two case hashes and
removes two diagnostic rows (1,996 to 1,994), while the 81-file hash stays unchanged.
Evidence: evidence/front9/node/. No dumps or duplicate case manifest committed.

Native remains refused at the bodyless-overload callback below; neither native
reference acceptance nor its output-byte mutant can run without a binary.
Claim remains: syntax tree and parse diagnostics identical on Node,
JSDoc unverified, error recovery unverified (native).
Assumption: reuse the pinned upstream single-file inputs and committed manifest,
with no compiler-option matrix, as the native acceptance inputs.

---

# Newest-tip native build checkpoint: red at an overload callback

Fresh newest-area slice and all ten requested compiler inputs: parser native
build and C-emission requests exit 1 at core.ts:54:101, before C emission:
`a type predicate whose return is not proven (there is no body proving this parameter)`.
This is every's bodyless overload callback signature. Split off, split on and
split repeat fail identically. No temporary signature removal or stub used.
New minimal native-predicate-overload.a reproduces at line 3:45; Node prints
`true`, exit 0. In contrast, the previous native-predicate-callback.a (with an
implementation body) now builds and runs native `true`, exit 0, empty stderr.
Thus the new predicate feature is present, and the remaining gap is specifically
a callback contract on a bodyless overload. The reduced overload returns boolean
on both signatures, so result-covariance ambiguity is not needed to reproduce it.

All three timed Node runs match the 81-file extended dump (cmp 0), 36,429,231
bytes, SHA256 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Best Node user time 6.552528 seconds. nproc 5. Parser C bytes/lines, clang wall
unsplit/split, parser binary bytes, native user time and instruction counts are
unavailable: no parser C or binary exists. Repeat has zero cache objects and is
not a warm-cache measurement. perf is not installed. No native byte mutant or
81-file/10,406-case comparison was possible. Exact failures and metrics attempts:
evidence/front9/build/report.json and logs. No compiler edits pushed.

---

# Newest-tip native rerun: inputs checkpoint

All ten requested fetched tips are in scratch HEAD 46854599d67d4c75bb39900f43e21c56c0538ff0,
verified with merge-base. area/stage3 is 03ccf222, front-2 391b3e9c,
array-holes 0140eed8, proven-predicates 746af2f2; remaining exact SHAs are in
evidence/front9/inputs/report.json. Scratch merge is never pushed.
Stage0 compiler rebuild succeeds. Focused predicate/non-null/checked-overload-result/
array-hole/phantom/require tests pass (3.816s); no complete gate claimed.
The new predicate feature changes verifier APIs: checked-overload flow helpers
use predicateFlowProof/Path; refusal dispatch uses predicateRefusal and preserves
actual callback body checks. These compatibility resolutions are scratch only.
Array-hole dispatch preserves record, phantom and Node host dispatch, plus
module Require roots and readiness behavior. Exact conflicts/diff retained.

Assumption: newest sanctioned source adaptations, including existing validated
60-64, produce the source slice; the old 81-file inputs and 10,406-case manifest
stay fixed as the acceptance targets. stage3/apply.sh succeeds at newest area.
The shared slicer regenerates the driver slice: 1,991 code declarations in 27
files (79 modules including evaluation scaffolding), 41,853 copied-span lines.
Audit: 2,080 exact source spans and 79 ordered imports PASS. No new adaptation,
source stub or second slicer. Next checkpoint is split/off native build and both
reference comparisons if a parser binary builds.

---

# Non-null and checked overload results native rerun: red

Newest fetched non-null-check a02613eff851e07f567ab9934adfc8a2dc98eeca and
census-small-families 33a90f4 are merged into unpushed scratch HEAD
b729fadd78fafb52c190f44d8bd89cbc3bfffd22 (both merge-base checks pass).
The stage0 executable rebuilds. The core.ts:28 indexed-read ! refusal is gone.
The actual first parser slice failure now, identical unsplit/split/split repeat:

```
/tmp/parser-front-driver-slice/src/compiler/core.ts:54:101:
Adamic 0.1 refuses a type predicate whose return is not proven
(there is no body proving this parameter); inline the check where you use it,
or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate)
```

This is every's overload callback type `(element: T, index: number) => element is U`.
The type-only FunctionType has no body. The complete callback of a call site can
have a real proven body; current predicate validation rejects the signature first.
Minimal native-predicate-callback.a (from pristine core.ts:140) checker-passes,
then reproduces the refusal at line 3:45. Raw source Node exits 0, stdout `true`.
No source adaptation, predicate weakening, or stub used to bypass this blocker.

The 81-file Node reference was rerun three times: cmp 0, 36,429,231 bytes,
SHA256 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615,
best user 6.109186 seconds. Parser C generation is unreached; no native binary,
no native 81-file or 10,406-case comparison, and no native parser output mutant.
Case reference still contains 1,323,111 nodes, 1,996 parse diagnostic rows and
34 JSDoc rows. Both references are required by native-proof.py before green.
Unsplit, split and repeat attempts all exit 1 before clang, so C size/lines,
clang time, parser binary size and native runtime/instruction counts are null.
No warm cache claim. nproc 5, perf is not installed.

## Scratch integration and checks

Non-null lowering conflicts preserve both readiness checks and import-cycle
module-read checks. The existing predicate verifier remains; non-null and
uninitialized-read refusal entries come out as the new readiness feature requires.
Small-families conflicts touch class field storage, condition coercion, callable
parameter fitting and invariance. Preserve omitted-argument padding and enum
relations while using the new callable representation; conditions use census ToBoolean.
Native conflicts preserve process exit status, checked field readiness and Node
performance invocation while adding maybe-boolean fields and receiver-aware calls.
Full UTF-16 view uses the new shared header layout, also serving RegExp's view.
The combined closure ABI has three parameters and a receiver field: the old Node
performance built-ins initially fail clang, then scratch-only signature/initializer
updates make the non-null native probe compile. All exact diffs and failures retained.
No scratch merge or compiler change is pushed on the proof branch.

Targeted lower TestNonNull, TestCensusOverloadReturnProof, TestPhantom and TestRequire
pass (2.914s). A broader TestCensusOverloadRelation/result_covariance check fails
because phantom-brand refusal precedes its expected overload-result diagnostic;
this is reported, not claimed green. No full repository gate was run.
The native-nonnull.a valid probe prints `1`, native/Node stdout cmp 0 and empty
native stderr. An empty-array input mutant compiles and exits 70 with
`non-null assertion failed: array[index]! is null or undefined`; the ruled loud
check is observed. This is a minimal feature probe, not a native parser proof.
Evidence: evidence/front8/report.json, compressed logs and scratch diff.

Status: **syntax tree and parse diagnostics identical on Node, JSDoc unverified,
error recovery unverified**.

---

# Recovery mutant and corrected old coverage: third checkpoint

Important correction: the old reference.json coverage.diagnostic_rows 0 counted
JSDoc rows, not parse rows. The 81-file dump has 10,671 parse diagnostic rows:
diagnosticMessages.generated.json 2,130, diagnosticMessages.json 8,535, tsconfig.json 6.
All three were deliberately parsed in TS mode. The 78 TypeScript inputs have zero
parse diagnostics. The broad parseErrorAtCurrentToken no-op mutant is already
caught by the old extended reference (including changed error flags); claiming
otherwise would be false. The original probe report and reference coverage fields
are corrected, with distinct parse/JSDoc counts. Dump bytes and SHA are unchanged.

A real uncovered JSX recovery path is the successful narrow mutant:
parseJsxAttributeValue skips its call to
parseErrorAtCurrentToken(Diagnostics.or_JSX_element_expected). Same source mutation
on the scratch slice exits 0 and leaves all 81 old inputs byte-identical (cmp 0).
The 10,406-case reference rejects it (cmp 1): exactly two changed case hashes,
1,996 -> 1,994 diagnostic rows, identical total 1,323,111 node count. Cases are
conformance/jsx/jsxAttributeInitializer.ts and compiler/jsxAttributeMissingInitializer.tsx.
Both lose their sole parse diagnostic. The committed recovery-mutant.py performs
both comparisons and requires positive diagnostic differences. No mutated source
or complete dumps committed. Evidence: evidence/recovery-mutant/report.json;
broad-mutant.json records the rejected broad attempt honestly.

Native remains red at core.ts:28:37's indexed-read ! before C generation, not
module initialization. No native binary, timings, instruction counts or output
comparison claimed. Status: syntax tree and parse diagnostics identical on Node,
JSDoc unverified, error recovery unverified. Native green requires the extended
81-file reference AND every case hash from cases-reference.json.

---

# Error recovery reference: second checkpoint

10,406 pinned single-file compiler/conformance cases are identical on full-tree
and slice Node runs, for every per-case SHA256: 1,323,111 nodes, 1,996 parse
 diagnostic rows and 34 JSDoc diagnostic rows. Cases by group: parser 800/245 rows,
JSX 175/95, salsa 108/0, remaining 9,323/1,656. The manifest cases-reference.json
records input paths/source hashes, full-tree and slice hashes, counts, and all
excluded multi-file cases. No dumps committed. Sources verified against pin
050880ce59e30b356b686bd3144efe24f875ebc8. All zero/one @filename cases included;
cases with more than one filename directive excluded (2,039). Metadata comments
retained, BOM decoded as upstream IO does, effective filename selects script kind;
Latest target, no compiler-option permutations or semantic checking of test inputs.

Driver handles TSX/JSX modes and prints JSX text. Its recursive traversal overflowed
on an upstream deeply nested tree; iterative preorder now handles it with identical
81-file bytes/SHA. Strict checker exit 0. The permanent node-end/JSDoc mutants
still reject wrong dumps. Native-proof.py requires both the 81-file reference and
cases-reference.json before green; --case-checkout and --full-tree select pinned
checkout/full runtime separately from the fixed compiler corpus.

After relocating retained own artifacts out of the full /tmp filesystem, phantom
and require focused tests pass (4.190s). Earlier bare [build failed] reruns were
at full /tmp; no compiler diagnostic survived. Native non-null refusal remains
real and unchanged. Native status: syntax tree and parse diagnostics identical on
Node, JSDoc unverified, error recovery unverified. The diagnostic-dropping mutant
is the next separately pushed checkpoint. Evidence: evidence/recovery-reference.

---

# Native rerun with module initialization, brands and require: red

Before the error-recovery reference, the unpushed scratch merge adds module-init-order
28e366f, phantom-brands d2d3c77 and require-builtins-2 6439f4c to the prior four-feature,
records/runtime and developer-tools merge. Scratch HEAD 422be912; no scratch merge pushed.
The stage0 executable builds. Native parser build exits 1 at slice core.ts:28:37:
`Adamic 0.1 refuses the non-null assertion !`. This is forEach's array[i]! from
permanent adaptation 30, not temporary 64's addRange read. Module-init-order has
moved the attempt beyond the enum-map stop. Minimal native-nonnull.a checker-passes
then is refused at !; the same source prints 1 on Node.

Extended 81-file Node output was rerun three times, all cmp 0: 36,429,231 bytes,
SHA256 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Best Node user time 5.974716 seconds. No C generated, clang unreached, no native
binary or native comparison or native output mutant possible. Unsplit, split and
split repeat all stop at the same refusal; no warm cache. nproc 5.

Lowering conflicts: phantom cast/dispatch/refusals, then require cast/representation/
dispatch/refusals/statements/exceptions/invariance. Both features and records/enum
checks retained; phantom and performance projection enter castProof before casting.
Typed SDK paths preserved in loader; five new string API calls get explicit string
conversions, SDK modules remain external scratch replacements. No compiler edit is
on the proof branch. Exact scratch diff and merge logs: evidence/front7.
Focused tests initially fail the performance projection and enum-name representation;
after projection resolution, test reruns report [build failed] without a detailed
compiler diagnostic. These are not green checks; logs retained, no full gate claimed.

Native status: syntax tree and parse diagnostics identical on Node, JSDoc unverified,
error recovery unverified. Continuing with the requested error-recovery reference
only after recording this native attempt. Evidence: evidence/front7/report.json.

---

# Error recovery blind spot: first checkpoint

Existing extended driver on Node alone, pinned TypeScript parser cases.
800 cases without any @filename directive, 245 parse diagnostic rows, 63183 nodes; 20 cases with explicit filename directives excluded.
Raw single-file contents and paths preserved. No driver edits for this probe.
The 81 compiler inputs exercise zero parse diagnostic rows, so their equality
does not establish error recovery. Details: evidence/recovery-blind-spot.
Status: syntax tree and parse diagnostics identical on Node, JSDoc unverified,
error recovery unverified.

---

# Extended dump acceptance checks: fourth checkpoint

The full adapted compiler and unchanged validated slice were compared again:
cmp exits 0, empty stderr, 36,429,231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
The same real slice-omission mutant removes Parser.initializeState as a complete
function declaration in a scratch copy. With PARSER_TRACE_ERRORS=1, Node exits 1
with ReferenceError: initializeState is not defined; the partial dump also fails
cmp (exit 1). This is a caught omission, not a completed mutant parse.
The successful control's planted Identifier.end +1 changes exactly one record
at line 166, end 3145 to 3146, and cmp exits 1. The permanent dropped-tags and
JSDoc-only diagnostic mutants were rerun and both fail comparison too.
Evidence: evidence/jsdoc-final/report.json and logs. No mutated compiler source
is committed. Native-proof.py now requires this extended reference, not the old
35,456,964-byte dump. Native comparison awaits compiler's module-init-order fix.
Until then: **syntax tree and parse diagnostics identical on Node, JSDoc unverified**.

---

# Permanent JSDoc mutants: third checkpoint

The permanent jsdoc-mutants.py runs from run.sh. Dropping tags at the parser's
return site still completes on all 81 inputs and retains all 4,149 JSDoc nodes,
but loses all 3,846 tags; cmp exits 1 against the extended reference. This is
the same source mutation that left the legacy SHA unchanged. A separate actual
JSDoc diagnostic code mutation on directed JavaScript input changes exactly
one row (1110 to 1111); trees and parse diagnostics stay identical, cmp exits 1.
Reports and logs: evidence/jsdoc-mutants. Full/slice reference remains
36,429,231 bytes, SHA256 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Native status: syntax tree and parse diagnostics identical on Node, JSDoc unverified.

---

# Extended JSDoc dump: second checkpoint

Status until native proof: **syntax tree and parse diagnostics identical on
Node, JSDoc unverified**. Full adapted compiler and slice extended Node dumps
are byte-identical (cmp 0, empty stderr): **36,429,231 bytes**, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
The old 35,456,964-byte SHA 2014ef06... is no longer the native acceptance target.
reference.json and native-proof.py enforce the v2 reference.

The driver visits attached node.jsDoc before ordinary children. On each JSDoc
and tag it prints kind/pos/end/flags, tagName where applicable, string or
structured comment text, and recursively visits tags/type expressions, link
names and all their children with forEachChild. It prints both parseDiagnostics
and jsDocDiagnostics per file. Full corpus: 4,149 JSDoc nodes, 3,846 tags, five
JSDocTypeExpressions, 56 links. No new slice declarations are needed.

The fixed TS/JSON compiler corpus has zero jsDocDiagnostics. Directed .js
inputs use range-cases' JS mode and produce a nonempty jsDocDiagnostics list:
code 1110, start 13, length 1, Type expected. They also exercise type expressions,
tag comments and a link/structured comment. Same driver, no synthetic diagnostic.
Strict integrated checker accepts the extended driver. Exact full/slice reports,
coverage and directed output are in evidence/jsdoc-extended. The permanent
no-tags mutant and rerun omission/end checks are the next separate checkpoints.

---

# JSDoc blind spot proved: first checkpoint

Status: **syntax tree and parse diagnostics identical on Node, JSDoc unverified**.
The legacy 35,456,964-byte SHA256 2014ef06... dump does not establish JSDoc
correctness: forEachChild does not visit node.jsDoc, and the driver prints only
parseDiagnostics. It emits zero attached JSDoc/tag nodes and no jsDocDiagnostics.

In a scratch slice, parseJSDocCommentWorker.doJSDocScan still consumes all tags
but returns undefined tagsArray to createJSDocComment. A positive TS @param
control changes from one comment/one tag to one comment/zero tags. The full
81-file corpus legacy dump nevertheless remains byte-identical (cmp 0),
35,456,964 bytes, SHA256
2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
This is an effective mutant that the old oracle fails to detect. Exact evidence
is in evidence/jsdoc-blind-spot/report.json. No native JSDoc proof is claimed.

Next checkpoint extends the driver before regenerating full-tree and slice
references; the old SHA remains historical, not the native acceptance target.

---

# Found behind a stub: follow-on Map initialization stops

This is exploratory evidence, not a source adaptation and not a native proof.
The validated slice and committed compiler remain unchanged. Reordered sources
stay untracked under /tmp and are never committed or pushed.

**Actual first blocker remains core.ts:11:52**, emptyMap's new Map call before
enum initialization. To investigate only that declaration, copied the validated
slice into /tmp/parser-after-runtime-enum-stub-slice, removed emptyMap with its
attached @internal comment from core.ts, and inserted the exact declaration
immediately after the final non-const top-level enum, ModuleKind, in types.ts.
The existing compiler barrel exports both modules, so no import, export-facade
or other declaration was rewritten. Only those two source files differ.
This deliberately changes evaluation order and is a diagnostic stub, not an
accepted semantics-preserving repair.

| Measurement | Result |
| --- | --- |
| Full slice, **found behind a stub** | debug.ts:333:29: stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet |
| Minimal probe, **found behind a stub**, moving emptyMap after Pending | /tmp/parser-enum-map-after.a:4:52: stage 0 can't lower a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions yet |

The exact statement is `const enumMemberCache = new Map<Record<string, string | number>, SortedReadonlyArray<[number, string]>>();`.

The full-slice result is still the global enum-initialization guard, now at
Debug's enumMemberCache initializer. No further initializer was reordered and
no compiler check was disabled. The minimal reordered probe exposes unsupported
never-key Map storage after the enum guard clears; that is a separate probe
finding, not a claimed second result on the complete slice.

A first local move after core's existing AssertionLevel const enum merely
relocates the same core failure to line 693. The guard tracks pending ordinary
enums across all modules; a const enum does not decrement that count. This
explains why the one declaration was moved to the runtime enum's module in
the actual exploratory attempt. Evidence/report.json records exact diagnostics,
paths and original source hashes. These findings belong to compiler fixes;
no temporary adaptation 65 or other adaptation is added.

---

# Final six-input retry: same enum-initialization NotYet, both compile modes

Unpushed fresh scratch starts at front-2 860a0d5 (779ff9d ancestor verified),
then area/stage3 4ad53a4, library 8280fd0, records-lowering 70fb62b, runtime-records
754e666 (already included), and developer-tools 2adf65c. Compiler builds.

**First failure:** core.ts:11:52, new Map<never, never>():
`stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet`.
C emission and actual unsplit, split and repeated-split CLI builds all fail
there, before generating C. C bytes/lines, clang wall times, parser binary
size, native parser runtime/diff/mutant are **unreached**, not zero. Split cache
contains zero files; the second split attempt is not a warm-cache measurement.
Attempt wall times 0.767176 / 0.770804 / 0.846200 seconds are not clang timings.
nproc=5; split jobs=5. No source workaround or compiler check suppression.

Type-only MapLike<T> *is admitted on this merge's record integration*: the
minimal native-type-only-map-like.a builds, native and Node print type-only,
exit 0, empty stderr; cmp 0. One-byte change to that actual native output gives
cmp 1. This is not a parser-output mutant and not a claim about all runtime
record forms. native-enum-map.a reproduces the parser's first NotYet.

Final same-input Node comparison: all three dumps match 35,456,964 bytes,
SHA256 2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5;
best user time 5.465565 seconds. perf is absent. Entry remains 1,990 code
declarations, with one extra diagnostic-driver declaration. Import gathering
and five validated temporary adaptations are unchanged.

Lowering conflicts: library cast.go keeps checked tag/class casts and carries
the qualified-const guard into cast_proof.go. Records expression.go keeps enum
and record dispatch; refusals.go keeps Node/enum/predicate/cast checks plus
all record checks. Record filename predicates need typed-path string conversion.
Developer-tools merge is source-conflict-free; only restoring scratch go.mod
paths conflicts afterward. Detailed resolutions and exact pins in front5/front6
README and compressed patches. Focused loader/lower Node/Cast/ImportCycle/Record
tests and native split literal/shared-state/header provenance tests pass.
Oracle counts keep front rows; no full count gate or full compiler gate claimed.

Evidence and scripts are committed only on codex/stage3-parser-proof. Scratch
merges and compiler resolutions are never pushed. Earlier measured stopping
points below are historical and superseded by this section.

---

# Four-input native retry: checker clean, lowering NotYet at emptyMap

Newest fetched area/stage3 4ad53a4, front-2 80fb9b7, explicit cycles 779ff9d,
and library Node loader 8280fd0 are merged only in detached scratch. Compiler
build succeeds against the cycle worker's exact SDK. The driver's erased
node:fs import activates the real pinned Node type seat; TS2591 is cleared.
Five validated parser temporaries remain applied.

**First measured lowering failure:** core.ts:11:52, at
`new Map<never, never>()` in emptyMap:
`stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet`.
Minimal native-enum-map.a has the same NotYet; Node (merged enum-capable runner)
prints 0 and exits 0. No new source workaround is introduced.

Lowering conflict in library merge: cast.go. Retain front castProof and checked
class/tag casts; carry library's qualified-const IsIdentifier guard into
cast_proof.go. Additional library_node.go typed-filename conversion resolves
an SDK compile error, not a textual conflict. Loader conflicts preserve typed
paths, Node globals and exact original-root accounting. Details and compiled
scratch patch are in evidence/front3. Focused load/lower Node, Cast and
ImportCycle tests pass. Oracle-count conflicts keep front's counts; no full
count gate or full compiler integration gate is claimed.

Parser entry remains 1,990 declarations / 26 files. Node dump unchanged,
35,456,964 bytes, SHA256 2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
All three Node comparisons exit 0, best user CPU time 5.727305 seconds.
perf is absent. No native parser binary means no native timing, size, actual
parser diff or parser-output mutant. Separate integrated native comparator
control self-cmp exits 0; one-byte mutant cmp exits 1. Node node-end mutant is
caught. This red supersedes front2's checker-stage stopping point below.

---

# Native attempt on compiler/stage3-front-2: red

October 7 run: area/stage3 4ad53a4 plus compiler/stage3-front-2 80fb9b7,
merged without conflicts as e224e66 in a detached scratch worktree, never
pushed. Compiler source unchanged; only scratch Go module replacement paths
were adjusted to the exact cohere 715ba94f / TypeScript 8d550c83 SDK pins.
Compiler build succeeds.

**First parser build failure:** src/compiler/performanceCore.ts:37:37,
TS2591: Cannot find name 'require'. Six checker diagnostics total: require and
perf_hooks in performanceCore, fs, require and two process references in tracing.
Validated temporaries 60-64 are applied. No fake Node declarations were added.
Minimal native-node-binding.a (`console.log(typeof require)`) produces TS2591.
The actual build diagnostic order and full text are in evidence/front2.

Lowering, native parser execution, native-vs-Node diff, parser binary size,
native user time and the parser-output byte mutant are **unreached** in this
run. The earlier zero-checker result used the six-feature scratch compiler with
its official host-node bindings and import control; this published compiler is
a different integration input. Do not mistake its binding diagnostics for a
regression of temporary 64.

The published front also does not contain import-cycles 779ff9d in ancestry:
its modules.go still refuses all import cycles. That is a source/ancestry
observation, not a newly measured second failure; checking stopped first.

Fresh createSourceFile gather: 26 code files / 1,990 code declarations; driver
roots add one declaration in program.ts, 27 files / 1,991 declarations. Raw
span and evaluation-prefix audit passes, before applying five validated temps.
Node full dump: 35,456,964 bytes, SHA256 2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
All three timing runs compare equal. Best Node user CPU time: 5.179120 seconds
(others 5.531493, 5.444617), same 81-file fixed corpus and manifest as prior proof.
This includes Node source loading/transpilation and printing. perf is absent.
Setup: go/clang/node/submodules ready 0 seconds each, build cache 95 seconds,
total 95 seconds, nproc 5.

A separate real native native-output-control.a prints parser comparison control.
Its output self-cmp exits 0; flipping only the first byte makes cmp exit 1.
This proves the comparator on real native output, not a native parser result.
The existing Node Identifier end mutant also changes exactly one line and is
caught. No green native proof is claimed. Scripts, minimal programs, pins,
counts, exact diagnostics and compressed logs are committed in parser territory.

---

# Current result after temporary 64: zero checker diagnostics, lowering reached

Push of 8411276 succeeded. The adaptation 64 list was separately pushed as
8dd6805 before the checked-read implementation. See its README for all 13
retained callers, ordinary-array producer evidence and the explicit boundary
excluding arbitrary client getters/proxies/prototype mutation.

Strict slice checker findings: **5 -> 4 -> 3 -> 2 -> 1 -> 0**. Removing the
second-read assertion restores exactly the former core TS2345. Native assertion
worker control exits 0 with 23; its invalid-second-read mutant exits 70 loudly.
The complete checked-model upstream suite passes 106,367 tests with only the
sanctioned API lines; parser and scanner remain byte-identical on Node.
Corpus second-read observations are zero; directed TS/JS/JSDoc cases make five.
Source provenance establishes the scoped invariant; corpus coverage alone does
not. Getter/push-getter counterexamples remain real for arbitrary public inputs.

## Ordered measured lowering list

| Order | Checker | Outcome | Exact location and message |
| --- | --- | --- | --- |
| 1 | Zero diagnostics, load.Load accepts | Refused | src/compiler/core.ts:1:1: an import cycle |

The unchanged real probe invokes lower.Lower only after load.Load succeeds.
Evidence: range-checker-lowering.json. No fake checker bypass or manual cycle
suppression was used. No second lowering result is established.

## Attempt to move the cycle blocker

Scratch only, never pushed: merge origin/codex/import-cycles
779ff9d337ed3e1f3b27d71dd3152dcb1ca8614b into the six-feature probe worktree,
resolve conflicts retaining its existing feature semantics, commit d361a72.
The old SDK first lacked cohere/rule_runner. Retry with the worker's exact pins
(cohere 7945d102, TypeScript d92d9bfee) instead exposes loader API integration
errors: sourceFS.AppendFile(string,string) does not satisfy the new typed-path
FS, ComparePathsOptions and UseCaseSensitiveFileNames no longer exist, and
NewCachedFSCompilerHost/NewParsedCommandLine have changed signatures. This is
an integration build failure, **not** an invented second parser lowering blocker.
Compressed build evidence records the failure. I stopped before changing more
compiler implementation or claiming the cycle was cleared.

Native range-check.a is independently tested on origin/codex/non-null-check
a02613eff851e07f567ab9934adfc8a2dc98eeca. That proves a loud absence check, not
successful integration of all features or a native parser proof. Full native
parser compilation and further ordered lowering blockers remain unmeasured.

The older stopping descriptions below are historical and superseded by this
section and the user-authorized scoped checked invariant.

---

# Checked addRange invariant: updated plan before implementation

The user now permits a loud checked read on states TypeScript's own parser
callers cannot produce. The previous whole-public-API counterexamples remain
valid, but do not alone establish reachability from createSourceFile.

**64-temporary-parser-range-read:** at addRange's second read only, change
`to.push(from[i])` to `to.push(from[i]!)`. Preserve both reads and the push
property lookup. This is contingent on the 13 retained direct callers' ordinary
array provenance: no indexed accessor/proxy and no side-effecting push getter.
The default compiler callbacks inspect nodes or assign module-indicator fields;
they do not install array descriptors. Debug NodeArray prototypes add only
__tsDebuggerDisplay, inherit Array.prototype, and do not change push or indices.

Reviewed producers: parseList's []/push statement arrays; initializeState's
[]/push diagnostic array and its saved alias; fresh [] jsDocDiagnostics;
mapDefined/filter/Array.filter tag arrays built from parsed JSDoc; sameFlatMap's
[]/slice result with the internal flattenCommaElements callback's node.elements
or two-element literal; mergeEmitNode's slice destination and compiler-created
synthetic-comment arrays (the latter is not reached for freshly parsed nodes
whose emitNode is undefined). Sparse/undefined entries at the first read remain
valid and are skipped exactly as before. The assertion only rejects absence
at the conditional second read.

This is a parser-entry checked invariant, not a universal addRange promise for
client-created arrays, custom factories, custom module-indicator callbacks or
prototype mutations. Evidence must enumerate all retained call sites, assert
there are no unreviewed flatMap/sameFlatMap callbacks, and retain a checked Node
probe that rejects both previously demonstrated getter states. Observed corpus
and directed case runs supplement the producer review, not replace it.

If these checks reveal a real compiler-owned getter state, no adaptation is
made; report an upstream candidate. Otherwise run the strict checker, exact
parser/scanner oracles and default stage3 suite, then record the first actual
lowering outcome and continue only through justified source adaptations.

---

# Temporary parser repairs: list pushed before implementation

## Current measured result after temporaries 60-63

The feature-integrated scratch checker, with its official Node type import
control, reports **5 -> 4 -> 3 -> 2 -> 1** findings as the four views are applied.
Every source edit is on one of the two parser-exclusive files. Exact sequential
findings are in evidence/temporary-{60,61,62,63}-checker.json. All four complete
Node dumps remain 35,456,964 bytes with SHA256
2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
The planted Identifier end mutant is caught in each run, changing just one line.

The initial Mutable<Declaration> view did not clear localSymbol: its optional
base field disallows explicit undefined under exact optional writes. The
revised structural view explicitly admits undefined and clears that site.
Its plan was pushed at da804b0 before the revised implementation. The obsolete
view's first oracle was intentionally interrupted (test exit -15), not counted
as a pass; that report is retained in evidence.

verify-temporaries.py removes each final view independently: each restores its
own diagnostic alongside core's finding. It also plants a runtime change in
each adapter; all four fail its JavaScript equality assertion before writing.
The original scratch source is restored byte for byte after every mutant.
Evidence: temporary-type-view-mutants.json.

The same-source scanner control produces exactly **509,014 tokens**, 27,879,197
bytes, SHA256 c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f.
Its control comparison exits zero and token-end mutant comparison exits one.
No core change was made. Partition 30's actual core counterexamples were
rerun: both append an own undefined element. The cached-read mutant exits one,
changing two reads to one and the appended value to 23. Evidence is retained.

**Remaining ordered blocker: core.ts:379:21, TS2345**, from[i] is T or undefined
at to.push. Lowering is **unreached**: the probe only calls lower.Lower after
load.Load accepts, and this final input still fails checking. I did not add !,
cache a read, widen away the T[] return promise, or lower rejected source.
No truthful, exact core restructuring was established in this unit. That is
its stopping point, explicitly allowed by the user's latest instruction.

These are temporary type views, not permanent proofs of factory generic
construction or native Node bindings. The compiler is still the recorded
scratch feature merge, not a production native gate. The first final-60 default run against the original API reference had
106,366 passing and one public-API failure. The older 20/40/70 checker rejected
that snapshot because it lacked partition 32's two handoff owners and 11 public
host-method owners. Its rejection and the raw API-only baseline diff are saved.
Partition 32's public-api.cjs reconstructs the exact sanctioned snapshot from
pristine source and owner ledgers, verifies all 60,930 other references unchanged,
and accepts only that snapshot in the disposable tree. No new parser exception
is added. An isolated unowned version declaration mutant (string -> number)
fails that checker; restored controls pass. All four successive sanctioned default runs pass **106,367 tests each**, zero
failures/pending, empty baseline diffs, install/build/test exits zero. No test
filter or weakened upstream compiler option is used; nproc is five and workers
four, with a 1,536 MiB heap budget. Wall seconds: **60: 243.492; 61: 251.273;
62: 244.528; 63: 234.146**. Each post-run exact API proof passes and its JSON hash
is identical, confirming no temporary introduces an API change and all 60,930
other reference baselines remain original. Reports and compressed phase logs
are in evidence/temporary-{60,61,62,63}-oracle*, with the aggregate in
temporary-oracle-summary.json.

This list supersedes the earlier ownership restriction for core.ts only:
the user now permits that shared file if the scanner oracle remains identical.

| Number | Directory | Exact proposed edit | Removal condition |
| --- | --- | --- | --- |
| 60 | 60-temporary-factory-local-symbol | At createBaseDeclaration's localSymbol assignment only, view node as { localSymbol: Declaration["localSymbol"] | undefined }, explicitly admitting absence; retain the assignment and return verbatim. | Factory construction establishes the generic subtype's fields. |
| 61 | 61-temporary-factory-type-expression | At createJSDocTypeLikeTagWorker's typeExpression assignment only, use a structural write view with JSDocTypeExpression or undefined; retain the assignment and return verbatim. | Factory construction establishes the generic subtype's fields. |
| 62 | 62-temporary-tracing-write | At writeSync(typesFd, JSON.stringify(descriptor)) only, give writeSync a type-only call view admitting string or undefined. Undefined still reaches Node and throws as before. | Node binding admission models this call's throwing behavior. |
| 63 | 63-temporary-tracing-legend | At writeFileSync(legendPath, JSON.stringify(legend)) only, give writeFileSync the equivalent type-only data view. Preserve path and data evaluation order. | Node binding admission models this call's throwing behavior. |

The two tracing errors concern serialized data, not paths (columns 35 and
38 point at JSON.stringify). These views leave runtime Node validation in
place; they do not claim native host binding support. The factory views are
explicitly temporary escape hatches: they do not prove the narrower generic
return contract which partition 30 declined.

**Core stopping condition:** addRange must still perform its first indexed
read, the push property lookup, and its second indexed read in that order.
Partition 30's two counterexamples append an own undefined element. Caching,
skipping that element, or asserting presence changes or lies about that
behavior. Widening the target array at the push would leave its T[] return
contract unproved. No core adaptation is proposed until an exact, truthful
contract repair is available. Lowering remains behind that checker gate.

The initial Mutable<Declaration> probe still rejects undefined: that optional base field does not admit an explicit write of undefined. The revised structural view above is pushed before implementing it.

Each implemented temporary must preserve emitted JavaScript byte for byte,
pass the default stage 3 suite with no additional API changes, and preserve
the full parser dump. Its removed type view must restore its diagnostic.
The scanner oracle and both core counterexamples will also be rerun.

---

# Fixed parser slice: ordered checker blockers, October 7

Shared tool 188de02 is merged. The raw slice now runs on Node and exactly
matches the full-tree 35,456,964-byte oracle, SHA256 2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
Both actual mutants are caught: one Identifier end increment changes exactly
one dump record; removing Parser.initializeState causes ReferenceError and
a failing full-dump comparison, with the control passing beforehand.
SLICE.md and evidence/fixed-slice-equivalence.json record these checks.

## Input and slice ownership

Source is the unchanged fully adapted area/stage3 7ad8666 tree used in the
previous run. The shared fix was merged, never rebased. No new adaptation
was applied. createSourceFile itself has 26 code files, 1,990 declarations
and 41,657 code-span lines (namespace wrapper pieces included). The driver
adds one function in program.ts: 27 code files, 1,991 declarations, 41,683
code-span lines. There are 79 evaluation modules, including empty scaffolds;
those modules do not mean the whole compiler's code is retained.

The same-tree scanner slice has eight code files and 89 declarations.
The parser driver has **19 code files outside the scanner slice**; the entry
alone has 18. evidence/fixed-slice-ownership.json lists all file sets.
Temporary 60-69 edits may touch only those exclusive code files.

## Stage0: baseline checker gate

The actual unmodified stage0 CLI build exits 1 with **143 diagnostics**, all
in reached-code slice files; no native binary is produced. The parser-entry
probe reports the exact same diagnostics. In first-code-occurrence order:

| Code | Count | Intended work |
| --- | ---: | --- |
| TS1294 | 85 | enums and namespaces feature branches |
| TS2345 | 1 | generic indexed read in scanner-shared core.ts |
| TS7030 | 26 | fallthrough-and-implicit-returns feature |
| TS7029 | 23 | fallthrough-and-implicit-returns feature |
| TS2412 | 2 | generic field writes in factory/nodeFactory.ts |
| TS2591 | 6 | official Node declarations and remaining host types |

The table groups codes; evidence/fixed-slice-baseline-ordered-checker.json
preserves every diagnostic and message chain in the exact returned order.
Only the absolute scratch-root prefix is removed. The first is
checker.ts:109:19 TS1294; :10:19 follows because the loader lexicographically
sorts complete formatted diagnostics. This is checker order, not execution
order. Strict, NoUncheckedIndexedAccess, ExactOptionalPropertyTypes,
NoImplicitReturns and NoFallthroughCasesInSwitch all remain enabled here.

## Candidate feature closure, scratch only

A never-pushed scratch branch merged taste-not-soundness, flag-enums,
namespaces-tsc, nested-functions, fallthrough-and-implicit-returns and
host-node-types-land in that order. Exact SHAs and source patch/hash are
saved in evidence/fixed-slice-stage0-summary.json and
evidence/fixed-feature-integration.patch.gz. Compiler changes never entered
the deliverable branch. Conflicts were resolved only in that scratch copy.

The scratch probe builds after removing a stale pre-feature enumElement
call; initial failed build and merge logs are retained. No full feature gate
or native compiler integration gate was run; these resolutions are not
proposed compiler patches or evidence of runtime correctness. In particular
this probe never calls lowering, because all measured sources fail checking.

The feature compiler removes all 85 TS1294 and 49 return/fallthrough
diagnostics, leaving **nine**. Return/fallthrough options change only by
merging their actual feature branch, not by a manual loader relaxation.
Strict indexed reads and exact optional writes stay enabled.

The host feature activates its locked official @types/node 25.3.3 on static
node:* imports. The gathered runtime uses require without a static host
import. A separate scratch driver prepends only
import type {} from 'node:fs'; this selects that feature's official declarations.
It changes no gathered statement, option or runtime import: Node still emits
the exact full oracle dump. With this control, the gate has **five** findings:

| Returned order | Slice location | Exact code and leading message | Ownership |
| --- | --- | --- | --- |
| 1 | core.ts:379:21 | TS2345: Argument of type T or undefined is not assignable to parameter of type T | Scanner-shared; outside parser temporary territory |
| 2 | factory/nodeFactory.ts:5037:9 | TS2412: JSDocTypeExpression or undefined is not assignable to T["typeExpression"] | Parser-exclusive |
| 3 | factory/nodeFactory.ts:757:9 | TS2412: undefined is not assignable to T["localSymbol"] | Parser-exclusive |
| 4 | tracing.ts:321:35 | TS2769: No overload matches this call; string or undefined passed where string is required | Parser-exclusive |
| 5 | tracing.ts:340:38 | TS2345: string or undefined is not assignable to string or ArrayBufferView | Parser-exclusive |

Those table messages are summaries; exact text and full chains are in
evidence/fixed-slice-features-host-ordered-checker.json. The nine-finding
profile is separately retained in fixed-slice-features-ordered-checker.json.
Both lists contain slice-file diagnostics only.

## Lowering order and temporary-adaptation boundary

**Lowering remains unreached.** load.Load returns CheckError in every measured
profile, so the probe does not invoke lower.Lower. I did not lower rejected
source, suppress diagnostics, substitute stubs, or claim an ordered
Refused/NotYet list behind that gate. The next required work is a faithful
closure of those five source contracts before real lowering can be measured.

The first remaining blocker is in scanner-shared core.ts:addRange. Its
from[i] read is guarded and then repeated; the generic checker still sees
T or undefined at to.push. It is not eligible for a parser-owned 60-69 edit.
The four exclusive sites are candidates for separately proved source/type
adaptations, not already validated repairs. No 60-69 adaptation is planned
or implemented here. This complete observed list is pushed before any such
edit; an exact README plan must precede an adapter, and its baseline must pass.

No native parser binary, native execution or baseline suite result is claimed.
The shared source-span/evaluation audit still passes after all controls.
No source file inside the delivered slice was edited.

## Reproduction and validation

```sh
source /workspace/adamic-tools/env.sh
go build -buildvcs=false -o /tmp/parser-fixed-adamic ./cmd/adamic > /tmp/parser-fixed-stage0-build.log 2>&1
/tmp/parser-fixed-adamic build /tmp/parser-fixed-driver-slice/parser-proof-main.a -o /tmp/parser-fixed-native > /tmp/parser-fixed-cli.log 2>&1
```

CLI exit 1, 143 diagnostics; probe outcomes Checker. All stdout/stderr/test
output went to files. run.sh bash syntax, changed-report whitespace checks,
full Node equality, both actual mutants and the final shared byte/import
audit were run. Full repository/feature gates and the upstream baseline
suite were not rerun for these driver/tool/evidence changes.

# Previous tool reports, superseded by the fixed-tool measurements above

# Parser slice blockers, October 7

The requested raw createSourceFile slice was generated from area/stage3
7ad8666 with the shared namespace-member slicer at faae0e9. Step 1 was pushed
at ed0fa58. Its exact inventory is 79 files, 5,103 declarations, 190,844
copied-span lines and 5,125 byte-identical spans. Dump helpers add no records.

## Gate order actually observed

1. Node initialization stops before parsing or printing. semver.ts Version.zero
   calls Debug.assert while Debug is undefined. A namespace-facade driver
   control has the same failure; a semver-first driver control instead fails
   at binder.createFlowNode's Debug.attachFlowNodeDebugInfo load. This is a
   real value read at module load time in the gathered import graph. It is
   not an Adamic Refused/NotYet observation. No closing feature was tested.
2. Unmodified stage0 at the integration source rejects the slice at checking:
   **830 diagnostics, all in slice files**. No checker options were changed,
   no loader overlay or diagnostic suppression was used, and the caller only
   invokes lower.Lower after load.Load accepts. The first returned diagnostic is:

```text
src/compiler/binder.ts:1672:18: error TS7030: Not all code paths return a value.
```

3. Lowering was not reached. There is no observed lowering/ownership/native
   blocker order for this raw parser slice. The list below is the complete
   checker gate, not a conjectured Refused/NotYet sequence behind it.

The fully adapted parser, outside the slice, still produces the original
35,456,964-byte dump SHA256 2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5
on the original 81-file input corpus. The raw slice produces zero bytes.
Deleting the exported createSourceFile declaration is caught by the passing
byte audit and the missing-export loader check; runtime dump-equivalence
mutant coverage is still unmet because the slice control cannot load.

## Exact ordered checker list

[evidence/slice-ordered-checker.json](evidence/slice-ordered-checker.json)
contains all 830 exact diagnostics, including message chains, one-based
returned order, file/line/column, code and ownership eligibility. Only the
absolute scratch-root prefix is removed. Ordering is stage0's lexicographic
sort of complete formatted diagnostics, not source execution order.

The following code groups are listed in first returned occurrence order;
this table does not replace or reorder the individual diagnostic artifact.

| Code | Count | First slice location | Candidate closing work |
| --- | ---: | --- | --- |
| TS7030 | 252 | src/compiler/binder.ts:1672:18 | codex/fallthrough-and-implicit-returns; not merged into this measured compiler |
| TS1294 | 180 | src/compiler/binder.ts:171:19 | codex/flag-enums and codex/namespaces-tsc; feature compilers not measured here |
| TS7029 | 83 | src/compiler/binder.ts:2148:13 | codex/fallthrough-and-implicit-returns; not merged into this measured compiler |
| TS2345 | 88 | src/compiler/builder.ts:1171:69 | No closing feature demonstrated in this raw-slice probe |
| TS2488 | 6 | src/compiler/builder.ts:1183:65 | No closing feature demonstrated in this raw-slice probe |
| TS2375 | 14 | src/compiler/builder.ts:2198:9 | exact-optional declarations; no demonstrated closure |
| TS2769 | 6 | src/compiler/builder.ts:2302:13 | No closing feature demonstrated in this raw-slice probe |
| TS18048 | 49 | src/compiler/checker.ts:10914:30 | optional/indexed-read contracts; no demonstrated closure |
| TS2322 | 30 | src/compiler/checker.ts:14025:9 | No closing feature demonstrated in this raw-slice probe |
| TS2412 | 23 | src/compiler/checker.ts:16770:21 | exact-optional write contracts; no demonstrated closure |
| TS2722 | 1 | src/compiler/checker.ts:19475:32 | No closing feature demonstrated in this raw-slice probe |
| TS2379 | 10 | src/compiler/checker.ts:20314:98 | exact-optional arguments; no demonstrated closure |
| TS2556 | 1 | src/compiler/checker.ts:21378:29 | No closing feature demonstrated in this raw-slice probe |
| TS2532 | 15 | src/compiler/checker.ts:49033:121 | indexed-read adaptations; remaining sites need individual proof |
| TS2420 | 1 | src/compiler/checker.ts:53209:7 | No closing feature demonstrated in this raw-slice probe |
| TS18046 | 6 | src/compiler/commandLineParser.ts:2193:91 | No closing feature demonstrated in this raw-slice probe |
| TS2740 | 1 | src/compiler/core.ts:1631:11 | No closing feature demonstrated in this raw-slice probe |
| TS2591 | 50 | src/compiler/performanceCore.ts:36:37 | codex/host-node-types-land; not measured here |
| TS2304 | 6 | src/compiler/sys.ts:1415:121 | No closing feature demonstrated in this raw-slice probe |
| TS2307 | 1 | src/compiler/sys.ts:1514:69 | host declarations/resolution; no demonstrated closure |
| TS7006 | 1 | src/compiler/sys.ts:1613:54 | No closing feature demonstrated in this raw-slice probe |
| TS7031 | 1 | src/compiler/sys.ts:1613:61 | No closing feature demonstrated in this raw-slice probe |
| TS2538 | 2 | src/compiler/transformers/classFields.ts:2043:45 | No closing feature demonstrated in this raw-slice probe |
| TS2339 | 3 | src/compiler/transformers/generators.ts:1798:50 | No closing feature demonstrated in this raw-slice probe |

NoImplicitReturns and NoFallthroughCasesInSwitch remain true: 252 TS7030
and 83 TS7029 findings count as blockers. The feature branch is a candidate
closure, not justification to suppress these diagnostics in the measured
compiler. The existing permanent adaptations have already run; they do not
close all remaining checker findings. No new baseline-suite result is claimed.

## Temporary adaptation ownership and stopping point

The scanner slice reproduced from this same adapted tree has eight files,
89 declarations and 7,126 copied-span lines. The parser slice has **71 files
outside those eight**. evidence/slice-file-ownership.json saves all three file
sets and the exact eligible diagnostic count. Shared files are:

- src/compiler/commandLineParser.ts
- src/compiler/core.ts
- src/compiler/corePublic.ts
- src/compiler/debug.ts
- src/compiler/diagnosticInformationMap.generated.ts
- src/compiler/scanner.ts
- src/compiler/types.ts
- src/compiler/utilities.ts

Temporary 60-69 edits may touch only the 71 parser-exclusive files, never
these shared files. No 60-69 edit is planned or implemented in this commit.
This observed list is pushed before any temporary source edit. Before such
an edit, its exact file/site/feature plan must be pushed; its README must
start "Temporary: comes out when <feature> lands"; its baseline must pass.

The expansion includes reached Debug.formatSyntaxKind using
(ts as Record<"SyntaxKind", Record<string, string | number>>).SyntaxKind.
The slicer's immediate-parent qualification test sees ts below AsExpression
and conservatively gathers the module namespace's exports. This inspected
path explains one source of the large gather, without proving it is the only
path. debug.ts is scanner-shared, so a parser-owned temporary Debug rewrite
is outside the authorized adaptation territory. I did not build a second
slice tool or alter gathered declarations/imports to bypass initialization.

The next proof step needs a faithful shared-tool/import-graph repair that
loads on Node, followed by a new byte audit and dump comparison, then checker
repairs or landed compiler features. Lowering cannot be honestly ordered
until checking accepts the real slice. No checker-rejected source was lowered,
no native parser binary was built, and no native agreement is claimed.

## Commands and validation

```sh
source /workspace/adamic-tools/env.sh
go build -buildvcs=false -o /tmp/parser-area-probe ./stage3/drivers/parser/probe > /tmp/parser-area-probe-build.log 2>&1
/tmp/parser-area-probe /tmp/parser-driver-slice/src/compiler/parser.ts /tmp/parser-slice-parser-entry.json > /tmp/parser-slice-parser-entry.log 2>&1
```

The probe exits zero after serializing the Checker outcome; that is not an
accepted build. The actual stage0 CLI build of the corrected slice driver
was also run: exit 1, exactly the same 830 diagnostics, no native binary.
Its full output is evidence/slice-stage0-cli.log. SLICE.md records the gather, byte audits, full-tree equality,
load-time failures and actual declaration omission. Bash syntax checks and
git diff --check were run on the driver/report changes. The complete Adamic
gate and native stage3 baseline were not run for this blocked raw slice.

# Historical reports below, superseded for the current slice

# Parser proof: slice blockers pending

The proof now targets the declaration slice of createSourceFile, using the
scanner worker's stage3/slice tool. As of fetched scanner commit 46041de, that
tool has not been pushed. No second slice tool is being built here. Once it
lands, the required order is: slice createSourceFile, compare its Node dump
byte for byte with the full-tree dump, then measure checker and lowering
blockers on that slice. Temporary 60-69 adaptations are limited to slice files
outside the scanner slice, and the slice blocker list must be pushed first.

The measurements below are historical whole-file evidence, not the requested
slice blocker list. Neither closure.cjs nor value-closure.cjs emits a slice.

# Historical whole-file checker measurements

The Node driver is complete and its node-end mutant is caught. The native
parser is still blocked at the checker. The entries below are observations
from real parser.ts/scanner.ts loads; no checker-rejected program was lowered.
The earlier stopped integration report is retained below as historical evidence.

## Inputs and reproduction

Base main remains ef3d907ecdc4c771b016f7d9c52372def057a340 (confirmed against
origin on this continuation). Feature SHAs are in each integration JSON and
in the historical input table below. Adaptation source is exactly
 a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e, adaptation 10 only, followed by normal
upstream diagnostic regeneration. Upstream source pin remains
050880ce59e30b356b686bd3144efe24f875ebc8. The Node dump includes 81 files.

The measurement overlay only replaces the Adamic prelude with pinned upstream
Node declarations (@types/node 25.3.3, @types/source-map-support 0.5.10).
All compiler options remain those of each merged feature compiler, including
noUncheckedIndexedAccess, exactOptionalPropertyTypes, noImplicitReturns and
noFallthroughCasesInSwitch. No upstream-tsconfig relaxation is used. The sound
regex library adapter remains in force. Each overlay source is saved in
 evidence/*-load.go.txt. Go VCS stamping is disabled because the scratch
worktrees reuse the pinned cohere checkout through a symlink. Absolute module
replacement paths reference the same pinned source; they change no semantics.

```sh
source /workspace/adamic-tools/env.sh
python3 stage3/drivers/parser/measure.py /tmp/new-parser-profiles /tmp/parser-adapted10 > /tmp/measure.log 2>&1
python3 stage3/drivers/parser/summarize.py /tmp/new-parser-profiles > /tmp/summary.log 2>&1
```

Scratch branches are scratch/parser-main, scratch/parser-taste,
scratch/parser-flags, scratch/parser-namespaces, scratch/parser-nested and
scratch/parser-combined. None was pushed. Initial incoming-whole-file conflict
resolutions dropped main fields and failed Go compilation. Those failures stay
in the logs. Final flag/namespace/combined compilers rebuild successfully after
three-way hunk reconstruction; merge.py records how both sides are retained.
The final scratch resolution diffs are committed as evidence/*-scratch-resolution.patch.
No production compiler file on codex/stage3-parser-proof was edited.

## Closure and ownership boundary

Following resolved imports and re-exports, including type-only edges, parser.ts
and scanner.ts each reach the same 78 TypeScript source files. Both enter the
large SCC through _namespaces/ts.ts. Therefore the literal file-closure
subtraction contains zero files and zero exclusive diagnostics. closure.json
saves every edge and both ordered member lists. This is the graph the current
loader checks, not a claim that scanner executes every compiler function.

A supplementary value-reference walk from createSourceFile/createScanner
reaches 25/12 files and 13 parser-only files. It groups declarations by top-level
statement and retains namespaces whole, so it is a conservative inventory,
not a proven source-slicing transformation or a replacement loader graph.
The definition and declaration spans are in value-closure.json. The 13 files are:

- src/compiler/checker.ts
- src/compiler/factory/baseNodeFactory.ts
- src/compiler/factory/emitNode.ts
- src/compiler/factory/nodeChildren.ts
- src/compiler/factory/nodeConverters.ts
- src/compiler/factory/nodeFactory.ts
- src/compiler/factory/parenthesizerRules.ts
- src/compiler/parser.ts
- src/compiler/path.ts
- src/compiler/performance.ts
- src/compiler/performanceCore.ts
- src/compiler/tracing.ts
- src/compiler/visitorPublic.ts

No temporary adaptation was made while this distinction is unresolved.
These historical inventories do not authorize 60-69 edits. The slice blocker
list will be pushed before any such adaptation.

## What each feature alone moves

Each entry is an actual load and a successful probe executable. All six final
compilers build. Each stops at Checker, so no corpus entry reaches Lower.

| Profile | Closure diagnostics | Removed from main | Added | parser.ts diagnostics |
| --- | ---: | ---: | ---: | ---: |
| main | 2673 | 0 | 0 | 68 |
| taste | 2673 | 0 | 0 | 68 |
| flags | 2493 | 180 | 0 | 58 |
| namespaces | 2493 | 180 | 0 | 58 |
| nested | 2673 | 0 | 0 | 68 |
| combined | 2493 | 180 | 0 | 58 |

Flags and namespaces each remove the same 180 TS1294 diagnostics by enabling
non-erasable syntax in their loader options. This gate movement covers enums,
namespaces and parameter properties; it does not prove flag-enums alone lowers
namespaces. Taste and nested functions move no checker diagnostics. Their
lowering effects are unobserved behind the remaining checker failures.

The combined scratch merge order was taste, flags, namespaces, nested.
Its final result is exactly the same ordered parser-entry diagnostic list as
flags alone: 2,493 diagnostics, no added diagnostics. Intermediate combined
source loads were not run; feature-alone loads and the final combined load are
the measured comparisons. The first reported blocker is shared:

```
src/compiler/binder.ts:1109:17: error TS2412: Type 'undefined' is not assignable to type 'FlowNode' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
```

Parser/scanner entry loads have identical site/code populations and counts,
with one exact-text difference at program.ts:3541:13 TS2375: checker type
printing changes object-field order and abbreviates a function signature.
The complete strings are in feature-summary.json; they were not normalized
away or credited as byte agreement.

## Exact current ordered gate list

The complete diagnostic chains, in the loader's returned order, are in
 evidence/<profile>-ordered-diagnostics.json.gz. Their uncompressed SHA-256s
are in feature-summary.json. This order is the actual checker report order,
including its textual-location sorting; it is not an inferred order of future
lowering refusals. The combined inventory has 58 diagnostics in parser.ts itself.
The full parser.ts lists are also uncompressed in evidence/*-parser-file.json.

The supplementary 13-file value inventory has 1,528 remaining diagnostics
(1,562 on main); all exact strings and their order are in
 evidence/*-value-exclusive-diagnostics.json.gz. Unreachable declarations in
those files are still checked by today's loader; the count is not restricted
to only reachable statement spans.

Combined checker-code counts, in first-reported occurrence order:

| Code | Diagnostics |
| --- | ---: |
| TS2412 | 617 |
| TS7029 | 83 |
| TS2322 | 105 |
| TS2532 | 225 |
| TS18048 | 359 |
| TS7030 | 252 |
| TS2345 | 688 |
| TS2375 | 77 |
| TS2379 | 31 |
| TS2769 | 6 |
| TS2722 | 2 |
| TS2556 | 2 |
| TS2488 | 1 |
| TS2339 | 28 |
| TS2420 | 1 |
| TS18046 | 6 |
| TS2538 | 7 |
| TS2366 | 1 |
| TS2684 | 2 |

## TS7030 and TS7029 remain blockers

noImplicitReturns and noFallthroughCasesInSwitch remain on in every profile.
The combined closure retains 252 TS7030 and 83 TS7029 diagnostics.
parser.ts itself retains the following 16 return and 9 fallthrough diagnostics
(in the loader's actual order). These are candidates for temporary 60-69
adaptations if parser ownership is defined by value dependencies; none was
silently suppressed pending codex/fallthrough-and-implicit-returns.

| Location in parser.ts | Code | Exact message |
| --- | --- | --- |
| 10367:18 | TS7030 | Not all code paths return a value. |
| 1272:199 | TS7030 | Not all code paths return a value. |
| 1682:21 | TS7029 | Fallthrough case in switch. |
| 2904:13 | TS7029 | Fallthrough case in switch. |
| 3791:14 | TS7030 | Not all code paths return a value. |
| 3987:37 | TS7030 | Not all code paths return a value. |
| 4095:14 | TS7030 | Not all code paths return a value. |
| 447:153 | TS7030 | Not all code paths return a value. |
| 4603:13 | TS7029 | Fallthrough case in switch. |
| 4609:13 | TS7029 | Fallthrough case in switch. |
| 4758:14 | TS7030 | Not all code paths return a value. |
| 4924:14 | TS7030 | Not all code paths return a value. |
| 5819:13 | TS7029 | Fallthrough case in switch. |
| 5852:13 | TS7029 | Fallthrough case in switch. |
| 6610:13 | TS7029 | Fallthrough case in switch. |
| 7492:53 | TS7030 | Not all code paths return a value. |
| 7753:14 | TS7030 | Not all code paths return a value. |
| 7766:25 | TS7030 | Not all code paths return a value. |
| 8441:14 | TS7030 | Not all code paths return a value. |
| 8984:25 | TS7029 | Fallthrough case in switch. |
| 9195:80 | TS7030 | Not all code paths return a value. |
| 9272:25 | TS7029 | Fallthrough case in switch. |
| 9346:22 | TS7030 | Not all code paths return a value. |
| 9446:22 | TS7030 | Not all code paths return a value. |
| 9707:44 | TS7030 | Not all code paths return a value. |

## Limit of the measured order

The current gate list is complete, but the exact latent lowering blocker order
is not measured: all actual source entries fail checking. I did not remove
2,493 checker diagnostics in a scratch source copy to reach lowering, did not
make temporary adaptations, and did not run the stage-3 suite or a native
parser dump. Moving enum/namespace syntax through the checker does not show
that their runtime or ownership cases compile. There is no native agreement
claim. The dump's successful mutant proves only that its byte comparison can
catch an actual one-node end change on Node.

The driver command, dump hash and mutant lines are in README.md and
 evidence/node-report.json. Setup from the first pass remains valid: 98 seconds,
nproc 5. Additional checks: bash -n run.sh, probe go vet, Python AST parsing,
and git diff --check excluding saved patch artifacts. The unfiltered check
reports whitespace in unified-diff context lines inside those artifacts;
they preserve the scratch changes verbatim. Test output stayed in logs. No complete repository gate
was run for these driver/probe/report changes.

## Historical preflight, superseded by the measurements above

This is an incomplete integration report, not the requested exact, ordered
parser lowering blocker list. No parser-specific lowering blocker was measured.
No temporary adaptation was made. No native parser proof is claimed.

## Pinned inputs

Started from origin/main ef3d907ecdc4c771b016f7d9c52372def057a340 on
codex/stage3-parser-proof. Fetched the required branches explicitly because the
checkout's default fetch brought down only main.

| Input | Commit |
| --- | --- |
| taste-not-soundness | aa896b5d5ccc82210184fd01b8fe4d0ce0730a50 |
| flag-enums | f7d62772fa8e52fcfae754e047ceb66dba88b782 |
| namespaces-tsc | ce8a2acf14e420a9c82345236845a377cd4c7a50 |
| nested-functions | b15216dabf65ffaa7152f6e64709b7b062ea01a9 |
| stage3-base | 8728405135d329efc12c837a7a6c293234abbe1c |
| tsc-census | 429c1177f0130f785c19cf590d1860513b2ddbfc |
| stage3-type-imports | a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e |
| stage3-optional-declarations | e3535e702e285a5dcb4d06364c4f889b8178ea0c |

## Observed integration sequence

The scratch worktree is /tmp/adamic-parser-scratch, branch
scratch/parser-proof. That branch has never been pushed.

1. An octopus merge of taste, flags, namespaces and nested functions failed
   with exit 2. Git reported "Should not be doing an octopus."
2. A sequential merge of taste fast-forwarded successfully.
3. The sequential flag-enums merge conflicted in four files:
   internal/lower/class_inheritance.go, internal/lower/lower.go,
   internal/lower/refusals.go and internal/oracle/counts.md.
4. I stopped without resolving these conflicts. Namespaces and nested
   functions were not subsequently merged. The scratch tree is still in its
   conflicted merge state; it is not a buildable four-feature compiler.

The shell command which printed the sequential merge logs returned zero
because its final command was git diff, but the flag-enums merge itself
failed. Its log explicitly says "Automatic merge failed". Do not count that
shell exit as a successful merge.

The conflict diff is preserved in evidence/merge-conflicts.diff. The conflicts
join nominal checks with enum checks, accessor discovery with enum
initialization, definite-assignment refusal with enum refusal, and independent
count rows. These are integration issues, not evidence about parser.ts.
They may be resolvable; I did not establish that they are impossible.

## Second witness

The stage1/typescript/parser implementation is a separate indexed-node port
following typescript-go's parser. Its --whole traversal has kind names,
positions, ends, cooked identifier/literal text and additional selected
metadata. main.ts maps its internal UTF-16 positions to UTF-8 byte offsets.
nodes.ts prints only the optional-chain bit as its node-flags column, plus
separate literal flags. Declaration semantics are another column. The node
table has no full context/NodeFlags field. expect and semicolon panic on
unsupported or malformed grammar; the driver does not emit parse diagnostics.

Consequently its current output cannot be reformatted into the requested
createSourceFile dump with full flags and parse diagnostics. Producing that
answer would require adding parser state and diagnostic behavior, beyond a
printer change. A kind/position/text projection could be compared after
normalizing positions and kind spellings, but that would be a weaker witness.
I did not run that projection or a Node corpus comparison. There is no measured
list of corpus differences in this report.

The historical WHOLE_REPORT.md records Go/Node/native agreement for its own
protocol over 77 compiler files. That is prior evidence, not a run tonight,
and does not establish this unit's full-flags dump agreement.

## Setup and limits

bash cloud/setup.sh completed successfully. Source the printed tool environment
at /workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node v24.19.0.
Timing lines: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warm 98s,
total 98s. nproc: 5; cgroup cpu.max: 400000 100000. The complete setup log is
in evidence/setup.log. Its warm run builds test binaries without running tests.

Not done: main.a, run.sh, adapted-tree creation, scanner/parser closure
subtraction, sequential source blocker removal, temporary adaptations, stage-3
baseline oracle, native comparison and the planted Node node.end mutant.
No dump comparison exists yet, so no claim that it catches that mutant is made.
No compiler or scanner-worker file was edited on the deliverable branch.


## Replaying the temporary checks

Use a fresh full tree made by the area/stage3 adaptation pipeline at 7ad8666,
and preserve its adaptation-20 stdout JSON owner ledger. Build a pristine
checkout of TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. After applying
a temporary, build the full tree before checking its emitted API. The existing
sanctioned API preparation is:

```sh
TSC_ADAPT_TYPESCRIPT="$typescript_api" node stage3/adapt/32-indexed-reads-program/public-api.cjs "$pristine" "$full_tree" "$optional_owner_ledger" stage3/adapt/30-indexed-reads/remaining-readonly-owners.json --accept-api
NODE_OPTIONS=--max-old-space-size=1536 bash stage3/oracle/run.sh "$full_tree" "$new_oracle_output" > "$oracle_log" 2>&1
```

Apply 60, 61, 62, then 63 successively, one per gate; rerun the same API proof
without --accept-api after every suite. The only changed reference is the
mechanically sanctioned api/typescript.d.ts. The temporaries themselves change
no published declaration. Scripts accept either the full adapted tree or a
copy of the gathered driver slice; CENSUS_TYPESCRIPT selects stock 6.0.3.

```sh
bash stage3/drivers/parser/run.sh "$driver_slice" "$new_dump_output" --inputs /tmp/parser-adapted10 > "$dump_log" 2>&1
python3 stage3/drivers/parser/verify-temporaries.py "$driver_slice" "$feature_probe" "$feature_compiler_root" "$new_mutant_output" > "$mutant_log" 2>&1
bash stage3/drivers/scanner/run.sh "$new_scanner_output" --tree "$scanner_slice" --inputs /tmp/parser-adapted10 --node-only > "$scanner_log" 2>&1
CENSUS_TYPESCRIPT="$typescript_api" node stage3/adapt/30-indexed-reads/remaining-contracts.cjs "$full_tree" > "$core_control_log" 2>&1
CENSUS_TYPESCRIPT="$typescript_api" node stage3/adapt/30-indexed-reads/remaining-contracts.cjs "$full_tree" --mutant-cache > "$core_mutant_log" 2>&1
```

The last command must exit one. The source feature patch and exact branch
versions needed to rebuild the scratch probe remain in the earlier evidence.
Use its existing type-only node:fs activation control to load official Node
declarations; do not replace that library with fabricated declarations or
relax checker options. The two explicitly listed local call views are temporary
source adaptations, not a substitute host library. No full merged native compiler gate or
native binary result is claimed.

# Scanner blockers

## October 8: reconciled closure restart stops at checked narrowing integration

Fresh scratch `/workspace/scratch/scanner-proof-restart-20261008` starts directly
at front-3 0059e65c1a3674692fe905c03abf9c2159c5c017. The requested closure tip
22fd701a3a4dee5c11cffec3d6b750cfe28ce399 is not its ancestor. Merging that exact
tip leaves 28 conflicted files. No conflict was resolved and scratch is never
pushed. Fetched area/stage3 is 6a8eebf41710e9bb696da9a7872d19765b1e7476,
which contains the required 116036f5. Its merge exits 128 because the closure
merge remains unresolved; no area changes are applied.

Exact compiler decision: `internal/lower/non_null.go`, in `nonNullValue`.
Front-3 retains the result representation of a checked union narrowing and
returns a plain Number/Boolean without a nullish pointer test. Incoming closure
adds cases that unwrap `ir.Narrow` and a single-target `ir.Call` marked
`CheckedUnionNarrow` to the original operand. Front-3 includes narrowed-number
ed6e30e2; implementation commit 2b6b0404 deliberately removed the `ir.Narrow`
unwrapping and added the scalar return. Restoring these incoming cases changes
that newer fix's treatment of checked results. Compiler must reconcile which
checked results may be unwrapped while preserving their checks and scalar
representation. This is a source integration conflict, not an observed runtime
failure or a claim that either branch independently fails.

The user-selected closure convention determines the ABI choice. The count-last,
typed plain/count-bearing code pointers in 22fd701a are not an undecided ABI
order in this handoff. Other conflicts and all automatically merged changes
remain preserved in scratch for compiler review. Complete combined diff and
exact ancestry checks: [closure-restart-conflicts.json](evidence/closure-restart-conflicts.json)
and [closure-restart-stop.json](evidence/closure-restart-stop.json).

Per the explicit stop instruction, no scanner compilation in either split mode,
C emission, binary, clang/warm timing, fresh Node corpus run, native comparison,
native user timing or native-output mutant was run. No minimal program is claimed:
a source program cannot reproduce a Git conflict, and no new compiler diagnostic
was observed. The 1,369,432-token / 466-error Node reference and its SHA-256 remain
historical evidence. No compiler test or full gate ran.

Required setup ran on the clean deliverable checkout, with GOPROXY exported
beforehand, and exited 0. Timing lines: Node 0.017s, Go 0.021s, submodules 0.049s,
markdown 0.050s, clang 0.118s, Go build 21.943s, deferred test binaries 22.006s,
build cache 22.008s, done 22.028s. These are the setup script's elapsed lines.
Sourced `/workspace/adamic-tools/env.sh`; nproc 5. Setup's overflow probe is
caught by AddressSanitizer. The setup log is retained alongside the evidence.

Evidence SHA-256 control passes. XOR of the conflict diff's first byte fails
its recorded SHA-256; this checks evidence integrity only. Staged whitespace
and scanner-only territory audits pass. No native-output sensitivity or scanner
correctness is inferred from those checks.

## October 7: front-3 integration stops at host compiler contracts

This retry stops in step 1, before building the validated scanner slice.
Scratch `/workspace/scratch/scanner-proof-20261007` starts at scanner branch
b3b9e81b and successfully merges area/stage3 b2c4549f and front-3 2db01c2f.
The successful scratch head is 17a219e824f2188de4526547eb87569c712b05fb.
None of the nine requested feature pins is an ancestor of that front-3 tip.
Merging host-blockers cd220dd5 leaves 32 conflicted files. No resolution or
compiler change was made, and the scratch checkout is never pushed.

Compiler handoff decisions:

- `internal/native/runtime/adamic.h` and `emit_functions.go`: front-3 uses
  `(self, argument_count, arguments)`; host-blockers uses
  `(self, arguments, argument_count)`. The call emitter also conflicts between
  front-3 Direct/canonical nested calls and incoming receiver injection.
  Declarations, emitters, runtime callbacks and host builtins need one coherent
  ABI while retaining both features.
- `internal/lower/expression.go`: front-3 checks a narrowed boxed union with
  TypeOf and Panic before Narrow; host-blockers directly narrows the observed
  array-predicate type. Integration must preserve the invalidated-narrowing
  guard while admitting the incoming predicate/storage cases.
- `internal/ir/ir.go`, flow/build.go and both backends: Break.Depth and existing
  breakable cleanup conflict with Break.Label/Continue.Label and labeled jumps.
  The compiler owner must reconcile the IR, CFG and cleanup contract together.

Per this unit's instruction to stop when a conflict needs a compiler decision,
the remaining feature merges are not attempted. The successful merges and the
unresolved host merge are preserved in scratch. The complete conflict diff,
all requested ancestry checks and exact hashes are retained in
[evidence/front3-integration-stop.json](evidence/front3-integration-stop.json)
and [the conflict diff](evidence/front3-host-merge-conflicts.json).

No scanner build in either split mode, generated C, clang timing, binary,
fresh Node corpus run, native comparison, native output mutant or compiler gate
is claimed. The established Node reference remains historical evidence:
1,369,432 tokens, 466 errors, SHA-256
41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc.
Setup initially exited 1 after checking out initial main's newer cohere pin
while the branch fetch was running; cache warming reported compiler-host and
RootedFilePath API mismatches. Rerunning on the scanner branch restored its
recorded cohere pin and passed. Retry timing lines: Go 0.016s, Node 0.018s,
markdown 0.051s, clang 0.105s, submodules 22.625s, Go build 52.044s,
cache warm 52.108s, done 52.130s; nproc 5. Sourced
`/workspace/adamic-tools/env.sh`. Both setup logs are retained with the evidence.

Evidence audit and staged whitespace check pass. XOR of the conflict artifact's
first byte is caught by its recorded SHA-256. This mutant checks only artifact
integrity; it is not a native-output mutant. No compiler tests or full gate ran.

This is a compiler merge blocker, not a new source-level NotYet; no minimal
source program can reproduce a Git conflict. A clean retry from front-3 once
it contains all requested pins is still required.

## October 7: five increment and nested-reference probes fixed

Integrated scanner-value-increments 85cade98 and scanner-nested-references
e35ee2d3 in unpushed scratch 48f092a5. All five unchanged probes build natively
and match source Node stdout, empty stderr and exit 0: postfix-call-argument
prints `0` then `1`; prefix-call-argument prints `1` twice; sibling and generic
sibling callback probes print `x`; ancestor call prints `1`. Each one-byte native
stdout mutant is caught by diff exit 1.

Merge limits and results: `evidence/increments-nested-references.json`.
Nested integration isolated f76331d5 with needed frame-forwarding seams, retained
existing host/readiness/coercion features, and initialized new canonical cache
fields in existing host closure constants. No compiler changes were pushed.
This is five-probe validation, not a full compiler gate or scanner token proof.
Full validated slice was held at this step, pending non-null write retry.

## October 7: string destructuring and nested overload probes fixed

Integrated scanner-string-destructuring cf9fa276 and nested-functions-scanner
8b7b8db9 into unpushed scratch b810a559, retaining scanner-expressions 998fb3eb.
The unchanged string-length-destructuring probe prints `10`; the unchanged
nested-overload-declaration probe prints `x`. Both build natively and match
source Node stdout, empty stderr and exit 0. Each one-byte native-output mutant
is caught by diff exit 1. The bodyless nested declaration guard test passes.
Commands, observations and merge limitations:
`evidence/destructuring-overload-fixes.json`.

String destructuring's lowering auto-merged. The nested tip's divergent
prerequisites conflicted in lowering and native/runtime code; scratch resolution
isolated its overload fix commit 433b2abe (skip signatures only when implemented,
and loudly guard a missing body) while preserving current prerequisites.
This is not validation of every additional prerequisite in that branch.
No compiler changes or scratch merge were pushed.

Validated slice audit passes. Full scanner builds were deliberately held,
as instructed: its last real first stop remains assigning to NonNullExpression.
No source adaptation, scanner native comparison or new metrics are claimed.

## October 7: numeric concatenation fixed; narrow bigint bypass reaches overload panic

Merged scanner-expressions 998fb3eb into unpushed scratch 934af3c1.
The unchanged string-number-concatenation.a builds and prints `7`, matching
source Node stdout, empty stderr and exit 0. A one-byte native stdout mutant
is caught by diff exit 1. No lowering conflicts in this merge; expression.go
auto-merged, while documentation and generated counts retained scratch versions.
Exact evidence: `evidence/scanner-expressions-retry.json`.

In the untracked discovery copy, restored parsePseudoBigInt's complete validated
body and scanner.ts's complete validated contents. Replaced ONLY the two compound
writes with `void shiftedDigit` and `if (residual) void residual`, retaining the
Uint16Array allocation, digit computations, and division reads/writes. The known
Map-using getNameOfScriptTarget stub remains. No other earlier scanner bypasses
remain, including no removed overload signatures.

Discovery now reaches the nested overload nil-pointer panic in
nested_functions.go:99, before any additional parsePseudoBigInt stop appears.
The existing nested-overload-declaration.a is its minimal probe. C output is
empty. This observes lowering order under explicit stubs; it proves neither
stubbed runtime equivalence nor native scanner correctness.

The validated slice byte audit passes. No validated source was changed, no full
slice retry performed here, and no native scanner token diff is available.
The last real slice stop remains the checked non-null typed-array write target.

## October 7: Map iteration and Uint16 constructor fixed; non-null write stops first

Merged host-blockers cd220dd5 into the scratch with runtime 1836fc27's
Uint16Array feature. Scratch SHA: 662f876c. Only generated counts conflicted;
no lowering conflict in this latest merge. Runtime integration's isolated patch
resolution remains explicitly disclosed in `evidence/uint16-runtime-retry.json`.

Both ADAMIC_NATIVE_SPLIT=0 and =1 builds of the untouched validated slice stop
at utilities.ts:76:9 (upstream utilities.ts:10502, with slice checked-index
adaptation on the target):

```text
stage 0 can't lower assigning to a NonNullExpression yet
```

Minimal probe: `probes/typed-array-nonnull-write.a`, `segments[0]! = 1`.
Source Node prints `1`; native refuses at 4:1 with the same diagnostic.
Required feature: lower assignment through the checked non-null element target
without losing its typed-array store or bounds check. This representation-specific
NotYet is not separately named in REPORT.md.

The unchanged map-union-iterator-binding probe now builds and prints `one:1`
and `text:x`, matching source Node stdout, stderr and exit. Its one-byte native
stdout mutant is caught. The unchanged Uint16 constructor probe was already green
in the preceding evidence. Slice audit passes; no stubs in these real builds.

The discovery order does not fully hold: this write stops before the nested
overload panic. Earlier discovery stubbed the entire parsePseudoBigInt body to
bypass Uint16Array, so it hid these adapted assignment targets. The ten later
stops remain discovery observations, not a claim that this new stop is absent.
No scanner C/binary, token diff or metrics are available. Expanded Node reference
is unchanged and reused, with no fresh full scanner Node run claimed.
Exact commands/results: `evidence/map-iterator-runtime-retry.json`.

## October 7: Uint16Array probe fixed; bigint writes stay in range

Runtime 1836fc27 is integrated in unpushed scratch 0c328304, using an isolated
first-parent typed-array patch to preserve existing compiler features. This is
not a claim that all divergent area/runtime prerequisites were merged.
Evidence and merge limitations: `evidence/uint16-runtime-retry.json`.
The unchanged uint16-array-constructor.a prints `1` on native and Node, with
empty stderr and exit 0. Its one-byte native-output mutant is caught by diff.

parsePseudoBigInt allocates L=ceil(N*b/16) segments, where N is digit count and
b is 1, 3 or 4. Digit k writes floor(k*b/16)<L. A nonzero residual means
(segment+1)*16<(k+1)*b<=N*b, so segment+1<L. The later division loop only writes
indices L-1 down to zero. Scanner-validated digits fit b bits; the length never
changes. This establishes in-range writes for valid scanner inputs, assuming
allocatable lengths; malformed direct callers are outside that argument.

`uint16-bounds.cjs` runs the actual sliced function with a bounds-checking Proxy:
4,610 edge inputs, 23,533,762 checked writes, zero bounds failures and results
identical to BigInt. The fixed compiler-source corpus has zero bigint literal
occurrences. Injecting a write at segments.length is caught. No full scanner
rerun occurred in this Uint16-only step, as instructed.

## October 7: ten stops behind Map and Uint16Array discovery stubs

Discovery only, using compiler scratch 04a365a8. Nothing stubbed enters the
validated slice. Its byte audit still passes. Full diagnostics, bypass sequence,
commands and source Node observations: `evidence/discovery-after-union-map.json`.
All probes below run on Node with exit 0 and empty stderr, and reproduce the
listed compiler stop. No scanner C was emitted and no equivalence is claimed.

First, stubbing getNameOfScriptTarget's body bypasses the Map iterator and reaches
utilities.ts:65:11, Uint16Array<ArrayBuffer>. Existing minimal probe:
`uint16-array-constructor.a`. Stubbing parsePseudoBigInt then reaches these ten
ordered stops (discovery locations; shrinking stubs change later line numbers):

| Order | Location | Exact stop | Minimal probe | Required feature |
| --- | --- | --- | --- | --- |
| 1 | nested_functions.go:99 while lowering createScanner | nil-pointer panic at bodyless nested overload signature | nested-overload-declaration.a | Skip overload signatures and lower implementation |
| 2 | scanner.ts:694:32 | BinaryExpression with a string and a number | string-number-concatenation.a | JavaScript string-number concatenation |
| 3 | scanner.ts:752:26 | BinaryExpression with a string and a number | string-number-concatenation.a | Same feature, next coercion site |
| 4 | scanner.ts:776:15 | destructuring a string | string-length-destructuring.a | String property destructuring |
| 5 | scanner.ts:1630:67 | first-class nested function reference from another nested function | nested-sibling-callback.a | Sibling nested callback value |
| 6 | scanner.ts:1805:62 | PostfixUnaryExpression | postfix-call-argument.a | Value-used postfix increment |
| 7 | scanner.ts:2049:25 | first-class nested function reference from another nested function | nested-generic-sibling-call.a | Generic sibling call with callback capturing a sibling |
| 8 | scanner.ts:2093:21 | first-class nested function reference from another nested function | nested-ancestor-call.a | Ancestor nested function call from deeper nesting |
| 9 | scanner.ts:2267:138 | PrefixUnaryExpression on a number | prefix-call-argument.a | Value-used prefix increment |
| 10 | scanner.ts:555:18 | function returning T | generic-function-property.a | Generic arrow wrapper in returned scanner object |

The ten count ordered sites, including repeated diagnostic families, rather than
ten distinct compiler features. Every NotYet above has the usual prefix
`stage 0 can't lower a ... yet`. The panic is a compiler crash, not a refusal.
Source spans are documented in each minimal probe. These representation-specific
reasons are not separately named in REPORT.md; required features in the table are
source-based inferences, with ownership left to compiler.

Bypasses are recorded explicitly: stub the Map-using helper and bigint helper;
remove nested error overload signatures; use String calls at numeric coercions;
read string .length directly; omit conflict-marker error callbacks; split the
postfix argument increment; omit the regex scanRange callback; stub the regex
worker; use a compound assignment for the prefix argument. All are untracked
exploration changes, not proposed adaptations. No mutant is claimed for a native
scanner comparison because every run stops before C emission. Prior probe/output
mutants remain recorded in the preceding real-build evidence.

## October 7: Map storage fixed; union iterator binding is first real stop

Merged host-blockers 2606e8b4 into the unpushed scratch already carrying
imported-const-case c9d885d7. Newest fetched library-array-holes remains f05aec3c.
Compiler builds. Exact pins, conflict resolutions and commands/results are in
`evidence/union-map-fix-retry.json`.

Both validated-slice builds (ADAMIC_NATIVE_SPLIT=0 and =1) stop at
utilities.ts:15:22, upstream utilities.ts:748, the `for (const [key, value]
of iterator)` binding in forEachEntry:

```text
stage 0 can't lower an iterator binding held in more than one word yet
```

Minimal unchanged-form reduction: `probes/map-union-iterator-binding.a`.
Node prints `one:1` and `text:x`; native refuses its binding at 4:18 with
that same diagnostic. Required feature: destructured Map entry iterator binding
whose value is a tagged union (string | number). This representation-specific
NotYet is not separately named in REPORT.md. The Map constructor/get probe now
builds and prints `1`; the case-declaration probe builds and prints `other`.
Both match Node stdout, stderr and exit, and each one-byte native-output mutant
is caught by diff exit 1.

The proof slice audit passes and the real builds use no stubs. Uint16Array's
independent probe still refuses, but the real slice has not reached that stop;
no Uint16Array discovery stub was used. Full scanner C/executable, token diff
and metrics remain unavailable. The established expanded Node reference is
unchanged and reused; no fresh full scanner Node run is claimed.

## October 7: case declaration probe fixed; union-valued Map still pending

Merged imported-const-case c9d885d7 into the unpushed feature scratch.
The unchanged `probes/switch-case-declaration.a` builds natively and prints
`other\n`, identical to source Node stdout, stderr and exit (0).
Changing exactly one byte of native stdout makes diff exit 1.
Evidence, exact pins, commands and conflict resolutions are in
`evidence/case-declaration-fix.json`; the mutant diff is tracked alongside it.

Conflicts touched lowering (collections, locals, object/switch and statements),
IR, JavaScript and native local readiness, documentation and the counts table.
Scratch resolutions keep prior features and combine switch readiness with
existing checks, enum default guards and discriminant evaluation order.
No compiler changes or scratch merge were pushed.

The broader uncached switch oracle failed: fixtures with block-scoped nested
functions stop in lowering, and the dead_zone_direct Node error control encounters
an undefined AdamicPanic in oracle/adamic.mjs. These are recorded failures,
not a green claim for the whole feature gate. The requested unchanged scanner
probe and its one-byte comparison mutant are green.

Per instruction, the full slice was not rebuilt or rerun. Its last observed first
stop is still union-valued Map (stop 2), awaiting host-blockers. No scanner native
diff or metrics are newly available.

## October 7: generic optional result fixed; union-valued Map is first real stop

Merged scanner-generic-optional-result 6ffb8bb8 into the existing unpushed
scratch, keeping all prior features. Only the generated counts table conflicted;
generic.go auto-merged. Exact scratch SHA and pins are recorded in
`evidence/generic-result-fix-retry.json`.

The unchanged `generic-optional-callback-result.a` probe builds natively and
prints `x`, matching Node stdout, stderr and exit. Its one-byte output mutant is
caught. Focused uncached source/native/JavaScript optional-result fixtures pass,
as does the compiler worker's absent-number-to-zero mutant (valid binary,
clean sanitizer execution, wrong stdout caught by Node).

Both ADAMIC_NATIVE_SPLIT=0 and =1 now stop at utilities.ts:14:22:

```text
stage 0 can't lower a Map of V yet
```

The site is `const iterator = map.entries()` in forEachEntry, upstream
utilities.ts:747. The scanner call passes targetOptionDeclaration.type, declared
Map<string, string | number>. The independent `map-union-values.a` probe still
refuses at 2:16 as `a Map of string | number`; source Node prints `1`. Required
feature: union-valued Map storage and operations, preserving the value tag.
This representation-specific NotYet is not separately named in REPORT.md.

The validated slice audit passes; no source or discovery changes were made.
Its established full-tree Node reference remains 1,369,432 tokens and 466 errors,
SHA-256 `41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc`.
This retry reused that unchanged full scanner Node dump and reran the focused
Node differential fixtures. The previously discovered case declarations and
Uint16Array stops remain recorded below; no stubs were used in these real builds.

No scanner C/executable, native token diff, compile/runtime metrics or scanner
native-byte mutant could run because lowering stops first. nproc is 5.

## October 7: never fallback fixed; generic result, Map union, case scope and Uint16 remain

Fetched and merged host-blockers 877ff0c, proven-predicates 6feee11, and newest
Array-holes f05aec3c into the unpushed scratch. All prior features remain.
Exact pins, conflict resolutions, tests and logs: evidence/host-blockers-retry.json.
The unchanged never-fallback probe builds and prints `x` then `missing`, matching
Node. Its one-byte native-output mutation is caught. Focused predicate/overload
lowering and uncached host never-arm, optional-callable and append oracle fixtures pass.

The validated slice is unchanged: its byte audit passes. Node still matches the
full extended reference exactly: 1,369,432 tokens, 466 errors, SHA-256
`41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc`.
The token-end mutant is caught. Both split modes stop at utilities.ts:13:17:

```text
stage 0 can't lower a function returning U | undefined yet
```

This is forEachEntry's result, upstream utilities.ts:746-755. Minimal probe:
`generic-optional-callback-result.a`; Node prints `x`. Constraining U to {} or
using a concrete string result closes the probe, with native/Node equality.
Required feature: infer and represent unconstrained generic optional callback
results. Census reason: this representation-specific NotYet is not separately
named in REPORT.md; the underlying form is a generic callback and optional result.

Continued in an untracked discovery copy, without changing the proof slice:

1. Adding only `U extends {}` exposes utilities.ts:14:22,
   `stage 0 can't lower a Map of V yet`, at map.entries().
2. A concrete specialization using the map's actual declared string-or-number
   value exposes the same site as `a Map of string | number`. Minimal isolated
   probe `map-union-values.a` refuses at 2:16; Node prints `1`. Generic and
   concrete numeric-valued Map controls compile, so those are not blanket Map
   or generic-Map refusals. Required feature: union-valued Map storage/operations.
   The first attempted ScriptTarget specialization was checker-rejected (TS2345);
   it was corrected after inspecting the actual `Map<string, string | number>`.
3. Restore the original generic declaration and stub only getNameOfScriptTarget
   in discovery. This exposes utilities.ts:51:13, `a declaration directly in a
   case (wrap the case in a block)`, from upstream utilities.ts:10477.
   `switch-case-declaration.a` reproduces at 6:13; block-wrapped control builds
   and matches Node's `other`. Required feature: lexical switch-case declarations,
   or a separately validated temporary scope wrapper.
4. An untracked block around that default clause exposes utilities.ts:67:11,
   `a value of type Uint16Array<ArrayBuffer>`, from upstream utilities.ts:10491.
   `uint16-array-constructor.a` reproduces at 2:7. Node prints `1` after storing
   65537, observing unsigned 16-bit truncation. Required feature: Uint16Array
   storage, construction, initialized reads and modulo writes. Fetched typed-arrays
   22b58091 and runtime 86769a34 explicitly leave Uint16Array unsupported; they
   support Uint8Array, Int32Array and Float64Array. Their uncommitted merge was
   abandoned; it is not part of the final compiler or any proof.

These later NotYet reasons are not separately named in REPORT.md. Each is
reported as an observation; no reachable declaration was dropped by the slice
tool. The lookup stub, constraints, specialization and case wrapper remain
untracked discovery-only changes, and no Node equivalence is claimed for them.
No temporary adaptation was introduced by this retry.

All successful native probe/control comparisons were challenged with one-byte
output mutants, each caught. Scanner C, executable, token diff, clang split and
warm-cache timings, binary size and native user time remain unavailable because
the first real blocker prevents C emission. The previous Node best user time is
4.270926 seconds; nproc is 5. No native scanner output mutant could run.

## October 7: adaptation 89 validated; never fallback in ?? is next

Latest integrated scratch includes frontend 391b3e9, proven-predicates 746af2f,
and every requested tip. The Array constructor first passed with 0140eed;
subsequent fetch/merge used newest tip ceec0ca1 (including e669f94). Exact pins
and the unpushed scratch SHA are in evidence/latest-native-retry.json.

Adaptation 89 was planned/pushed first in 9347c1d. Its two numeric type arguments
and six erased non-null markers preserve JavaScript bytes. Numeric write mutant
fails that identity check. A numeric array write/read control prints `3` natively
and on Node, and its one-byte native-output mutant is caught. Full baseline:
106,366 passing, zero pending, one sanctioned API difference, reconstructed
exactly with the reference tree's matching 0b484f8f API-owner proof; all 60,930
other references unchanged. Newer proof tools expect newer adaptation 32 source
forms and fail against this intentionally retained source snapshot; that attempt
and the successful pinned proof are both recorded. No baseline was accepted.

Expanded Node slice remains byte-identical to the established full-tree dump:
1,369,432 tokens, 466 errors, SHA-256
`41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc`.
Token-end mutation is caught. Both split modes stop at core.ts:128:43:

```text
stage 0 can't lower ?? whose sides have different types yet
```

Expression: `(s1[i - 1] ?? Debug.fail("stage3: missing dense element"))!` in
levenshteinWithMax, from upstream core.ts:2220. The checked read is adaptation
52; Debug.fail returns never. Census reason: this representation-specific NotYet
is not separately named in REPORT.md. Required compiler feature: a never-returning
fallback in nullish coalescing, preserving lazy execution and throwing behavior.
The present string determines the expression's value representation; the never
call has no result to store. No replacement of Debug.fail or its error behavior
is made as a source adaptation.

Minimal probe coalesce-never-fallback.a reproduces at 4:12. Node prints `x` then
`missing`; native refuses before C emission. Literal-fallback and explicit-branch
controls both build and match those two Node lines. A one-byte change to each
native control is caught. Focused lowering tests and uncached optional-function,
append-overload and Array scanner differential fixtures pass after scratch-only
merge resolutions. C size, clang timings, scanner binary size, native user time
and native scanner byte mutant remain unavailable because no C was produced.

## October 7: newest tips close Array construction; numeric scratch storage is next

Library-array-holes 0140eed builds the unchanged length-constructor probe,
native stdout `4`. The merged compiler passes focused predicate, overload,
phantom and array-hole lowering tests. Next C-emission stop is core.ts:113:9,
`stage 0 can't lower an array of any yet`, at `previous[i] = i` in
levenshteinWithMax (upstream core.ts:2205). Adaptation 89 is planned and pushed
before implementation: erased numeric element typing and, where required,
erased assertions on reads proved initialized by the original banded loops.
Evidence and all latest pins: evidence/latest-tips-preflight.json.

## October 7: checked overloads and direct Debug.fail pass; Array constructor is next

Scratch includes census-small-families 33a90f4, module-init-order 28e366f,
phantom-brands d2d3c77, and require-builtins-2 6439f4c alongside the previous
integrated compiler features. Scratch SHA and all pins are recorded in
`evidence/three-compiler-fixes.json`; none of this compiler merge is pushed.

Unchanged append probe builds natively and prints `[1,2]`, `[3]`, `undefined`.
Unchanged Debug.fail callable-marker probe builds and prints `ok`. Both match
Node byte for byte, and a one-byte mutation of each native result is caught.
Adaptation 87 is retired from the scanner profile. Removing its erased Function
view emits identical JavaScript; changing the capture argument fails that check.
No claim is made that an individual branch alone closes the marker relation.

Final slice still has 89 code declarations in eight files. Node matches the
full-tree expanded reference: 1,369,432 tokens, 466 errors, 108,019,935 bytes,
SHA-256 `41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc`.
Its token-end mutant is caught (diff exit 1).

Both split modes and C emission stop at slice core.ts:107:20:

```text
stage 0 can't lower new an Identifier yet
```

The expression is `new Array(s2.length + 1)` inside `levenshteinWithMax`, from
upstream core.ts:2199 (second allocation at 2200). Census reason: this length-form Array NotYet is not separately named in
REPORT.md; it was discovered by lowering the checker-clean slice. Required feature: length-form Array construction with JavaScript
holes and observable length, rather than Map/Set/class construction alone.
Minimal probe: `probes/array-length-constructor.a`; Node prints `4`, and native
stops at 4:20 with the same NotYet. No stub or adaptation substitutes an array
with different hole behavior. This requires compiler/library support.

Compiler merge conflicts touched lowering (cast, expressions, record/refusal
seams and host operations), flow, loading and emission. Resolutions stayed in
scratch. Focused overload-proof and phantom tests pass. The attempted uncached
optional-function oracle cannot compile its test package because incoming host
fixtures pass an inputRun to the newer inputLeaks factory API; this is a separate
test-harness merge incompatibility, recorded exactly in the evidence.

No scanner C or executable was produced, so clang split/unsplit/warm timings,
binary size, native user time and the scanner native-output byte mutant could
not run. nproc is 5; perf is not installed. Earlier native probe mutants and the
current Node token-end mutant are green; neither is a native scanner proof.

## October 7: optional function values landed; append overload results are next

Merged census-small-families 8d34357 (including f69bf60) into unpushed scratch
and resolved overlapping lowering/emission features. Final scratch: deb79ea5.
Focused lowering checks and four uncached optional-function differential oracle
fixtures pass after resolving duplicate old/new boolean slot converters.
The first merged optional-boolean tests failed clang, not output comparison;
retaining the incoming tagged representation closes that merge artifact.

Adaptation 87 annotates only Debug.fail's metadata fallback `(fail as Function)`.
Upstream AnyFunction parameters and runtime arguments are unchanged. The marker
probe builds, prints `ok` natively and matches Node; a one-byte native output
mutant is caught. Stock JavaScript stays byte-identical (1,302 bytes), and the
changed-capture-argument mutant is caught.

Next preflight was scanner.ts:3510:67, the two String-as-any casts around
fromCodePoint (adamic/no-unchecked-cast). Removing them exposed 3510:66,
unbound-method, for the availability read. Adaptation 88 removes those casts
and uses `typeof String.fromCodePoint === "function"`, retaining both branches
and their bodies. Callable/absent controls match the original; a flipped
predicate mutant is caught. Its contract is the standard callable-or-absent
intrinsic, excluding truthy non-callable monkey patches.

Both split modes now stop at core.ts:34:1:

```text
Adamic 0.1 refuses overload 1 of append result T[] cannot be served by implementation result T[] | undefined; make the implementation result covariant with every overload result
```

Census reason: overload. Required feature: prove correlated overload results
from the implementation body. Original source: core.ts:921-932. The complete
body returns `to` when value is absent, a new array when to is absent, otherwise
the appended array. The overload promises follow these input conditions;
checking only implementation-result covariance loses that correlation.
Minimal source/body probe: probes/append-overload-result.a. Node prints `[1,2]`,
`[3]`, `undefined`; native refuses at line 3:1 with the same diagnostic.

An untracked experiment removed only the three erased signatures. The upstream
compiler checker went from zero diagnostics to 18. The experiment did not
mutate the validated baseline tree. No adaptation 89 or widened overload
promise was committed. See evidence/fallback-append-signature-experiment.json.

87/88 were planned and pushed first in 11163169. Full baseline with both edits:
106,366 passing, one mismatch limited to api/typescript.d.ts. The composed stock
API proof reconstructs exactly the sanctioned changes; all 60,930 other
references remain identical. No API reference was accepted or edited.
Both expanded slice Node runs match the full-tree reference: 1,369,432 tokens,
466 errors and SHA-256
41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc.
Both token-end mutants are caught. Both native builds and C emission exit 1 at
append before C exists. Clang times, C sizes, scanner binary size and native
user time are unavailable. nproc 5; perf not installed. This remains a red
native result. Complete evidence: evidence/fallback-retry.json.

## October 7: 55 and 80 reconciled; upstream callable marker exposes function widening

Adaptation 55's opaque `{}` rewrite is withdrawn. New slices retain the exact
upstream `stackCrawlMark?: AnyFunction`; old opaque slices can be restored by
55's adapter. Adaptation 80 still removes only the two erased Error casts.
Stock TypeScript 6.0.3 and pinned @types/node 25.3.3 accept the call. Restoring
`{}` fails the stock checker with TS2345 and the integrated checker with TS2740.
The combined 55/80 edit emits the same 1,302 JavaScript bytes as the untouched
gathered Debug source; a changed capture argument fails that identity check.

Both ADAMIC_NATIVE_SPLIT=0 and ADAMIC_NATIVE_SPLIT=1/ADAMIC_NATIVE_JOBS=5 now
stop at debug.ts:15:58 with the same exact diagnostic:

```text
Adamic 0.1 refuses a function taking string | undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
```

Census reason: method-signature-style. Required compiler feature: sound function
widening to AnyFunction. No function-widening branch was found. The named
optional-widening branches concern optional properties. The library-function
tip afd87f6b records existing Function gaps; it still refuses optional/rest
functions as values. Minimal reproduction: probes/debug-fail-callable-marker.a
(Node prints `ok`; integrated native build refuses at line 7:54). The smaller
probes/capture-stack-marker-callable.a isolates the marker independently.

An erased Function cast on the fallback gets past widening in a control but
then reports `stage 0 can't lower a function with an optional or rest parameter,
as a value yet`. That control is tracked as
probes/capture-stack-marker-function-view.a. No runtime rewrite, stub or
unchecked cast was committed to bypass either compiler gap. The capture call
without a marker still builds and prints `ok` natively.

Both Node runs match the extended full-tree reference: 1,369,432 tokens,
466 errors, 108,019,935 bytes, SHA-256
41672da9bab56f9d10ad7d45b5938f96e3f969c6a299e5be1b260cba189893cc.
Each token-end mutant is caught. No native scanner comparison is possible yet.
C emission exits 1 before any C is written; clang is not reached, so C size,
clang times (unsplit, split cold/warm), native binary size and native user time
are unavailable. Node best-of-three user time: 4.270926 seconds; all three
outputs match the reference. nproc: 5. perf is not installed.

Scratch compiler 34accc5c includes area abdcf3db, front-2 860a0d5e, library
8280fd08, records a2572f0b/runtime 754e6667, developer tools 5794c876 and prior
non-null/Error features. The octopus attempt conflicted in lowering expression
and refusal files and rolled back. Sequential records conflicted only in the
refusal pass and counts: retained its recordStorageView check and the existing
predicateRefusal wiring. This scratch merge and its compiler changes are never
pushed. Full evidence and replayable probes: evidence/marker-reconciliation.json.

## October 7: developer-tools split compile on and off

Merged area/developer-tools 2adf65c2 without conflicts into scratch 64d47034,
retaining the previous front-2/records/library integration. Rebuilt compiler
successfully. nproc is **5**. Tried the unchanged scanner slice with
ADAMIC_NATIVE_SPLIT=0 and with ADAMIC_NATIVE_SPLIT=1 ADAMIC_NATIVE_JOBS=5.
Both stop at the same Error-as-any refusal in debug.ts:14:14. Both Node
streams contain 509,014 tokens and match the full-tree oracle (empty diff).

`adamic c` refuses at the same location before writing C (stdout 0 bytes).
Therefore no scanner generated-C byte/line count, clang wall time (unsplit,
split cold or warm), or binary size is available. Zero stdout is a failed
emission, not a zero-sized generated program; clang is not reached.

Added build-metrics.go, explicitly run inside the integrated compiler checkout
(`go run stage3/drivers/scanner/build-metrics.go GENERATED_C NEW_OUTPUT 5`).
It prewarms the release runtime, then times native.Build unsplit, split with
cold object cache, and split with warm object cache. The new private cache
keeps cold/warm observations separate from unrelated programs. Its wall scope
follows CLANG_UNITS.md: splitting, preprocessing, cache checks, clang compilation
and linking; C generation and Go startup are excluded. This includes API
bookkeeping rather than pretending to report pure optimizer phase time.

Validated the helper on the unchanged admitted MapLike probe only. All three
binaries print `ok`, byte-identical to Node; a one-byte mutation of warm-split
native output is caught (diff exit 1). These control measurements do not stand
in for scanner measurements. Raw numbers and scope are in
evidence/split-control-metrics.json. No compiler change or scratch merge was
pushed. Evidence: split-run.json and split-*.log/.diff.


## October 7: records closes MapLike, next real gate is Error's any cast

Scratch 06b7f7b3 contains newest area/stage3 4ad53a4, front-2 860a0d5
(which includes import-cycles 779ff9d), library/qualified-as-const 8280fd08,
records-lowering 70fb62b1 and runtime-records 754e6667, retaining prior scanner
features. All five requested tips are ancestors. Compiler rebuild succeeds.

**MapLike blocker closed:** unchanged probes/index-signature-type-only.a
compiles and prints `ok` natively, identical to Node. The pure mutable
string-indexed type-only declaration is admitted; this probe needs no additional
compiler fix. Changing one byte of this probe's native output is caught by
diff (exit 1), scoped to the probe rather than scanner output.

**RED, exact next real failure:** `debug.ts:14:14: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)`.
The expression is `(Error as any)` in Debug.fail's captureStackTrace condition.
The existing unmodified error-constructor-value.a is the minimal program:
Node prints `captureStackTrace available`; this compiler refuses its cast at
2:6. Error as a value is implemented in the scratch merge; the checked-cast
preflight now refuses the cast to any before the Error value lowering gets it.
The source also has the same cast in the subsequent capture call. No Error
stub, removed capture operation, or unchecked compiler admission was added.
As previously assigned, removal of these casts after admitting the pinned Node
declarations belongs to adaptation 40's unit, not a scanner source bypass.

Records merge lowering conflicts: expression.go combines recordExpression with
enum/namespace/Error dispatch; refusals.go replaces the blanket index-signature
refusal with recordLiteralRead, recordStorageView, recordElement and delete
admission while preserving all existing feature checks. The runtime branch was
already an ancestor of records-lowering. Updating front-2 conflicts only in
oracle counts. Two scratch rooted-filename API conversions were needed in
expression.go and records.go. No compiler edit or scratch merge was pushed.

The unstubbed scanner still emits 509,014 Node tokens over all 81 files; the
full-tree diff is empty (exit 0). Token end+1 mutant is caught (exit 1). Native
scanner binary, token diff, timings and size remain unavailable at this refusal;
perf is not installed. No further stub discovery or full compiler gate was run.
Evidence: records-run.json and records-*.log/.diff.


## October 7: front-2 plus explicit import-cycles and library loader, real run

**RED, exact first failure:** `corePublic.ts:9:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added`.
This is the reached type-only `MapLike<T>` declaration's `[index: string]: T`,
from upstream corePublic.ts:14. No body lowering or ownership pass is reached.

Scratch 6ade8c38 includes newest area/stage3 4ad53a4, front-2 80fb9b75, explicit
import-cycles 779ff9d, library/qualified-as-const 8280fd08, and the previous
scanner feature integration. All four requested refs are ancestors. Compiler
build succeeds; the safe import-cycle probe prints `1` natively. This retires
the previous conservative cyclic-import stopping point. Lowering conflicts and
scratch API resolutions are reported in evidence/front2-conflicts.md; none are
pushed as compiler changes. No stubs or new source adaptations were used.

Minimal reproduction: probes/index-signature-type-only.a. Node prints `ok`;
stage 0 refuses its index signature at 3:5. The control omits only the interface,
prints `ok` natively, and matches Node (diff exit 0). Flipping exactly one byte
of that native control's output (`ok` -> `nk`) is caught by diff (exit 1).
This is a native **control** mutant, not scanner native output. The runner also
contains the requested one-byte scanner native-output mutant, for when it builds.

Node scanner over all 81 files still emits 509,014 tokens, 27,879,197 bytes,
SHA256 c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f.
The full-tree diff is empty (exit 0), and the token-end mutant is caught (exit 1).
Native scanner diff, best-of-three timings, and scanner binary size are unavailable
because compilation stops. perf is not installed here. measure.py will record
all three user CPU times and binary bytes and attempts instruction counting when
a green native comparison exists. Full compiler gate was not run.

This is a compiler admission blocker for a reached type-only string-indexed
record. Replacing it with Map changes TypeScript's declaration; silently dropping
it would violate the slice contract. No further bypass discovery was done.
Evidence: front2-run.json, front2-*.log/.diff, and front2-conflicts.md.


## October 7: real unstubbed run after library Error and non-null tip

Scratch compiler e034d3d merges library-error-value 6469f37 and non-null-check
a02613e, retaining the earlier feature integration. Compiler build passes;
the original error-constructor-value.a probe prints `captureStackTrace available`
natively. No Error capture stub or discovery bypass is present. Selected
validated adaptations remain 52, 53, 54, 55, 56, 57, 59, 81, 82, 85; 58 is retired.

**Next real blocker, before lowering:**
`diagnosticInformationMap.generated.ts:13:45: Adamic 0.1 refuses a conservative refusal of an imported binding read during cyclic module evaluation [nexus/correctness-no-import-cycle-load-time-read]; defer the read until module evaluation finishes or break the import cycle (#xjdce2d)`.
The read is `DiagnosticCategory.Error` in the Diagnostics initializer. The
order-preserving slice retains types.ts -> _namespaces/ts.ts -> generated
diagnostics -> types.ts. The interim guard rejects all top-level imported
reads in a cyclic program, including initialized exporters and enum members.
This is an observation on the scratch merge, not a claim of unsafe execution.

Minimal program: probes/cyclic-initialized-enum/{main,barrel,types,reader}.a.
Stock TypeScript 6.0.3 on Node prints `1`; stage 0 refuses reader.a:2:23 with
the same diagnostic. codex/import-cycles tip 7b2c3f2 was fetched and inspected:
it still installs conservativeLoadTimeReads; the public runner admission seam
is TODO #xjdce2d. Needed compiler change: prove safe imported reads using actual
module evaluation order (including const-enum reads), while refusing premature
value reads. No source adaptation was made: deferring Diagnostics changes
verbatim declarations, and removing barrel cycles discards the evaluation-order
proof the shared slicer now preserves. Native build and native token diff stop
here; no further stub discovery or full compiler gate was run.

Node over all 81 corpus files emits 509,014 tokens, 27,879,197 bytes, exactly
matching the full-tree oracle (diff exit 0). Token end+1 mutant is caught
(diff exit 1); unmodified comparison control exits 0. Evidence: error-real-*.


## October 7: slice evaluation order and reference-chain repair

Shared tool now retains the original runtime graph and binding modules.
Scanner still has 89 code declarations in eight files; 70 import-only/facade
modules preserve the original 78-module evaluation graph. Node matches all
509,014 tokens. Parser driver reaches 1,988 code declarations in 26 files,
41,677 code-span lines; its full 35,456,964-byte dump matches the full tree.
Reverse-import mutants fail Node on both slices and specifically fail the
ordered-import audit while declaration bytes remain intact. --why identifies
the old static Debug enum lookup as a false whole-barrel expansion; no
reachable dynamic namespace member is dropped. See stage3/slice/WHY.md and
evaluation-proof.json. This is a Node/tool proof, not a refreshed native run.
Library Error SHA is still awaited; no new behind-stub discovery is done.

## October 7: silent miscompile isolated behind the discovery stubs

**Found behind a stub, on the unpushed integration compiler, not attributed to
a feature branch alone.** The heavily stubbed scanner built after further
bypasses below. Identical scratch source on Node emits six tokens for
`let x = 1;\n`; native exits 0 but emits only Identifier "l" and EOF at 1.
Output diff catches this silent miscompile (exit 1). The minimal independent
probe discovery-optional-method.a has no scanner stubs: native prints 0,
Node 11. Its trace prints `false false false` natively versus
`false true true` on Node for undefined checks on newText/start/length.
Missing optional numeric method arguments arrive as present zero values.

Controls: discovery-required-method.a prints 11 on both backends.
discovery-explicit-undefined-method.a, with the original optional signature
and explicit undefined arguments, prints 11 on both backends. This isolates omission
handling; it does not prove the integration conflict resolutions are sound.
The scratch driver with explicit undefined arguments emits the same six ASCII
tokens as Node (diff exit 0); an end+1 mutant is rejected (diff exit 1).
No production compiler change or tracked source adaptation is proposed here.
Evidence behind-optional-method-*, behind-required-method-*,
behind-explicit-undefined-method-*, and behind-capture-ascii-38.diff.

Further ordered discovery observations after the previous ten:

| Order | Scratch location | Observed diagnostic / behavior | Scratch-only bypass |
| --- | --- | --- | --- |
| 11 | scanner.ts:775:79, then 871:25 | stage 0 can't lower a function value taking boolean &#124; undefined yet | required Boolean arguments for five nested helpers; adjust scratch interface and calls |
| 12 | scanner.ts:699:32, 707:30 | stage 0 can't lower a BinaryExpression with a string and a number yet | five numeric-text concatenations become templates |
| 13 | scanner.ts:781:15 | stage 0 can't lower destructuring a string yet | read scanIdentifierParts().length directly |
| 14 | scanner.ts:1094:17 | stage 0 can't lower a declaration directly in a case (wrap the case in a block) yet | wrap eight scanner cases |
| 15 | scanner.ts:2053:25 | stage 0 can't lower a first-class nested function reference from another nested function yet | omit scanRange callback's regexp-worker call |
| 16 | scanner.ts:2099:21 | same cross-nested NotYet, charCodeChecked call in scanDisjunction | stub whole regexp worker after retaining its body for this probe |
| 17 | core.ts:38:59 | stage 0 can't lower a value of type T &#124; undefined yet | omit remaining comment-directive append write |
| 18 | scanner.ts:3269:138 | stage 0 can't lower a PrefixUnaryExpression on a number yet | ++pos in an expression becomes pos += 1 |
| 19 | scanner.ts:558:18 | stage 0 can't lower a function returning T yet | remove three generic callback members from scratch Scanner/object |
| 20 | scanner.ts:565:9 | Adamic 0.1 refuses Object.defineProperty; property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores | omit inactive debugging registration |
| 21 | scanner.ts:116:5 | stage 0 can't lower a computed field name yet | direct constructor key with KeywordSyntaxKind cast |
| 22 | scanner.ts:188:46 | stage 0 can't lower Object.entries on a shape not proven by a plain literal or its const binding yet | empty keyword Map |
| 23 | scanner.ts:3513:67 | stage 0 can't lower String as a value outside equality or typeof (overloaded calls and static properties need their own representation) yet | select existing utf16EncodeAsStringFallback |
| 24 | scanner.ts:3523:59 | same Object.entries NotYet on inline Unicode alias literal | empty alias Map |
| 25 | scanner.ts:476:5 | Refused onError, captured variable can be reached from what it holds (adamic/cycle-capable) | empty setOnError body |
| 26 | generated native C | clang rejects 11 Boolean-to-adamic_object* diag argument conversions | private optional metadata parameters become Boolean |
| 27 | scanner.ts:3547, isolated property probe | native exit 70: non-null assertion failed: undefined! is null or undefined; Node property control prints 0 | replace scratch Script_Extensions initializer with empty Set |
| 28 | scanner.ts:3475 resetTokenState | same loud undefined! panic after property bypass | omit absent tokenValue reset assignment |
| 29 | driver setText call | silent bounds miscompile described above | explicit undefined method arguments: ASCII control matches Node |

These are new lowering/runtime observations, not original checker census
counts. Owner candidates: closure/first-class generic cases nested-functions
and generic-maybe-undefined; numeric concatenation/static String and Array
constructors library; computed/entry shapes records-lowering; primitive
optional representation and missing method arguments compiler ABI; undefined
property initializer non-null-check; onError ownership/cycle analysis. None
of those candidates has been proven to close these observations. Latest
nested-functions 70be5ff adds witnesses/docs only, no lowering implementation
diff from integrated e7587d3. Library Error branch is still awaited.

Captured literal initializer control discovery-captured-literal.a passes on
both backends ([] and [source]); adaptation 58 remains retired. Object-field
initializer and reset assignment are separate from that successful local
initializer admission. Generic callback/regexp/body/metadata stubs change
semantics; this scratch executable is not the scanner proof. Initial missing
clang PATH and temporary stub-induced TS2322 diagnostics are excluded from
the blocker count. Full logs behind-capture-14 through -39.

A full-corpus stub run was cancelled: blanking error calls discarded a pos++
argument in the PrivateIdentifier path, so even Node lost progress. This is
a discovery-stub artifact, excluded from the compiler blocker list. Retained
a bounded log; no full-corpus stub comparison or native proof is claimed.


## October 7: discovery behind an untracked capture stub

**Found behind a stub. These are not validated adaptations or native proof.**
Compiler: scratch 32352a2 / scanner-latest-live-adamic. Driver: copied main.a
and token-names.a, adapted symlink points only at the scratch discovery tree
/workspace/scratch/scanner-behind-capture. The deliverable slice is unchanged.
Library owns Error as a value and captureStackTrace: upcoming
codex/library-error-value (unit 01a114a0), due 06:00 MDT; not yet merged/tested.
The host-node-types-land 69c71d5 loader and adaptation 40 are another owner's
work. Capture is removed only in scratch, never as a tracked adaptation.

Observed first diagnostics, in discovery order; all locations refer to scratch
slice files under src/compiler, not original upstream line numbers:

| Order | Location | Exact diagnostic after stage 0 can't lower | Scratch bypass used to reach the next | Owner / census relation |
| --- | --- | --- | --- | --- |
| 1 | debug.ts:13:15 | throwing an Error that isn't made where it's thrown or caught by the catch around it yet | inline new Error at throw; remove local e | Error lowering; library / error-classes candidate, not verified |
| 2 | diagnosticInformationMap.generated.ts:6:64 | a field from a boolean &#124; undefined variable yet | omit optional diagnostic metadata fields | optional representation; census TS2375-related contract, new lowering finding |
| 3 | utilities.ts:13:17 | a function returning U &#124; undefined yet | remove forEachEntry and return undefined from getNameOfScriptTarget | generic optional return; generic-maybe-undefined 86ad3eec candidate, not verified |
| 4 | utilities.ts:51:13 | a declaration directly in a case (wrap the case in a block) yet | wrap the default case in braces | switch lexical declarations; no closing branch demonstrated |
| 5 | utilities.ts:65:22 | new an Identifier yet | stub parsePseudoBigInt to return its input | Uint16Array construction; library owner, no closure demonstrated |
| 6 | core.ts:106:20 | new an Identifier yet | stub levenshteinWithMax to return undefined | Array(length) construction; library-array candidate, not verified |
| 7 | scanner.ts:261:17 | ?? whose sides have different types yet | four dense-index ?? Debug.fail guards become ! | nullish coalescing / never representation; taste / non-null candidate, integrated tips do not close this occurrence |
| 8 | scanner.ts:609:5 | a function without a body yet | blank two nested error overload declarations, keep implementation | overload declarations; nested-functions tip does not close this occurrence |
| 9 | scanner.ts:611:48 | a function value with an optional parameter yet | remove error implementation/reporting calls, suppress conflict callbacks | closure optional/default parameters; census a function inside a function (a closure), nested-functions tip does not close it |
| 10 | scanner.ts:775:79 | a function value with an optional parameter yet | not yet bypassed | checkForIdentifierStartAfterNumericLiteral isScientific?; same closure representation |

The overload-removal stub additionally induced a method-signature-style Refused
at scanner.ts:1635:67: error taking number seen as number | undefined. Widening
errPos to number | undefined advanced to order 9. This is a stub-induced
signature change, not an independent claim about the original slice. Removing
only the five conflict callback arguments did not close order 9. An initial
empty-reporting stub spelled undefined as a statement and produced an Identifier
statement NotYet; corrected to empty blocks, excluded from the blocker count.

Discovery build commands: scanner-latest-live-adamic build SCRATCH_DRIVER/main.a
-o /tmp/scanner-behind-capture, each redirected to its own log. All listed builds
exit 1 before native emission. Logs are evidence/behind-capture-01 through
behind-capture-13. No executable, ownership check, token equality, baseline
validation or adaptation approval is claimed for this stub tree. Live branch
names were inspected; candidate owners above are inferences, not green probes.


## October 7: live feature refresh; native compiler blocker

Explicitly fetched feature heads: the configured ordinary origin fetch updates
only main. Earlier branch-tip observations below are historical. Live refs are
in evidence/native-live-feature-refs.json; scratch integration is
32352a2a46c2313eeeaf56a5ee4f9ebfaa36ccbb (never pushed).

Current profile selects 52-57, 59, 81, 82 and 85. Non-null bc9f5d7
closes 58: literal undefined! and captured tokenValue! controls compile and
match Node. Adaptation 58 is omitted, not silently applied. The original
textInitial! initialization still fails: probes/scanner-lazy-text.a compiles
but native exits 70, read before assignment: variable text in textInitial!.
Node prints [] then [source]. Therefore retain 57's text initialization and
setText(textInitial) rewrite. On non-null alone this exact probe stops earlier
at reading setText (nested-function dependency); the integrated runtime result
is the observed lazy-read failure. No feature-alone runtime result is claimed.

Numeric enum ec67b02 closes 80, 83, 84 and proposed 86. Those adaptations
are omitted; 86 has no delivered source edit. Proven predicates eef541e admits
Debug.assert syntax, but restoring it exposes the next lowering NotYet:
assert-unknown.a:1:17, stage 0 can't lower a value of type unknown yet.
Keep call-site adaptation 56. With it, checker diagnostics in slice files are
clear; lowering stops at debug.ts:12:14, stage 0 can't lower reading Error yet.
Minimal program probes/error-constructor-value.a reproduces it at 2:6.
Node prints captureStackTrace available. No compiler branch found for Error
constructor reflection; omitting capture breaks the stack-marker oracle.
A faithful source adaptation has not been found. This is the stopping point: no
native scanner executable, ownership traversal, or native token diff completed.

Current reduced profile: Node 509,014 tokens, 27,879,197 bytes, SHA256
c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f.
Full baseline 106,367 passing, zero failures/pending/differences, 246.014s.
The end+1 mutant is rejected (diff exit 1); comparison control exit 0.
State mutants (callback start, JSX default, regexp key) all exit normally and
differ from Node controls. Evidence native-current-*, native-state-mutants.json,
native-lazy-text-* and native-captured-marker-*.

## October 7: adaptation 85 guards function-body lexical locals

23 function-body var keywords become let, preserving declaration positions and
initializers. No nested-block or repeated var bindings; AST guard enforces it.
Creation state is declared before setText and scanner construction; regex state
before parsing calls. Full 84-86 baseline 106,367 pass, zero differences.
Census reason var; no branch for general var lowering found. A minimal var
function is Refused; its let control runs natively and matches Node output 1.
The nested-block mutant is rejected by the adaptation guard.

This is a proven known lowering limitation, not the first full-slice diagnostic:
full-slice syntax checking first reports scanner:2034:75 numeric regexp keys;
after closing that, module lowering stops at Debug's Error value before reaching
scanner locals. Do not read preemptive local-var handling as a completed lowering
or ownership traversal. Evidence native85-{scope-mutant,var-refusal,let-control}.

## October 7: adaptation 84 names the zero reset

The lone raw tokenFlags = 0 becomes TokenFlags.None. Both are numeric zero.
Census reason enum-flags, feature zero-assignment inference pending. Restoration
mutant caught at scanner:1864:22. Combined 84-86 suite: 106,367 pass, zero
failures/pending/differences, 252.427s; Node tokens match. The next syntax
refusal observed with lexical locals is scanner:2034:75, numeric code point
passed as CharacterCodes. Evidence native84-86-baseline-* and restoration ledger.

## October 7: adaptation 83 gives four zero branches flag provenance

Four private EscapeSequenceScanningFlags ternaries use String & ReportErrors
for zero. Disjoint bits prove zero; enum and all other expressions unchanged.
Census reason enum-flags; zero inference still missing at flag-enums tip.
Next exact refusal scanner:1864:22, tokenFlags = 0. Complete combined suite:
106,367 pass, zero differences/failures/pending, 230.719s. Node tokens match.
Evidence native83-next-blocker.log and native59-83-baseline-*.

## October 7: adaptation 82 forwards generic scanner callbacks

Three Scanner object properties have explicit generic forwarding arrows.
Public interface and original helpers unchanged. Merely spelling the interface
methods as properties still refused; wrappers close it. Census reason
method-signature-style; generic function widening pending. Next exact refusal:
scanner:949:48, EscapeSequenceScanningFlags.String combined with an unproven
numeric zero branch (enum-flags). Combined suite passes 106,367 tests with zero
differences; Node tokens match. Evidence native82-next-blocker.log.

## October 7: adaptation 81 preserves the JSX default while widening its type

reScanJsxToken parameter is boolean | undefined = true. Omitted/undefined
arguments still receive true; false stays false. The initial README incorrectly
assumed false and was corrected in a separate plan push before implementation.
Census reason method-signature-style; function widening pending.
Next exact refusal: scanner:558:9, function taking () => T seen as () => T,
for tryScan. Combined 59/80-83 baseline passes 106,367 tests without differences;
Node tokens match. Evidence native81-next-blocker.log.

## October 7: adaptation 80 makes the flag return type explicit

getNumericLiteralFlags's arrow now declares : TokenFlags; its AND expression
is unchanged. Census reason enum-flags, branch flag-enums inference gap.
Next exact refusal: scanner:538:9, Boolean parameter seen as Boolean | undefined
(method-signature-style) for reScanJsxToken. Complete 59/80-83 suite passes
106,367 tests with zero differences; Node tokens match. Evidence native80-*.

## October 7: adaptation 59 preserves legitimately missing regex operands

`let operand!: string` becomes `let operand: string | undefined`. Malformed
class-set inputs leave it absent; existing !operand tests handle that absence
and narrow before length/code-point reads. Plain string was rejected by TS2454
at both tests. Census reason: definite assignment assertions. Feature: lazy
initializer/definite-assignment checks. Node 509,014 tokens match; combined
59 and 80-83 baseline 106,367 pass, zero failures/pending/differences, 230.719s.
Next at this snapshot remains scanner:528:33 flag-return inference; its earlier
position in the refusal walk precedes operand. Evidence native59-83-*.

## October 7: adaptation 58 removes the initialization spelling blocker

Per updated instruction, text is now `var text: string;`; its first setText
still receives textInitial. tokenValue is `var tokenValue: string;`. The lazy
initializer feature 01a1130a will remove this adaptation. No eager value added.
Text's first write precedes every read; the intervening-read mutant holds that.
An early getTokenValue or speculation save can legitimately observe absence;
the token driver calls neither. Do not claim universal assignment-before-read
for tokenValue. Plain declarations pass checking, unlike undefined!:never.

Full upstream baseline: 106,367 pass, zero failures/pending/differences,
233.054 seconds. Node: all 509,014 tokens identical. Next actual refusal on
this path is scanner.ts:528:33, flag-return inference at getNumericLiteralFlags
(adamic/enum-flags); explicit TokenFlags return annotation closes it in a probe.
The plain nested operand probe instead produced TS2454 at two legitimate
undefined tests; adaptation 59 uses an optional type. Evidence native58-*.

## October 7: adaptation 57 validated; native still blocked

Actual text initialization rewrite: upstream scanner.ts:1034 becomes
`var text: string = undefined!;`; upstream:1058 becomes
`setText(textInitial, start, length);`. The intervening executable statements
only declare locals; comments mentioning text do not read it. setText assigns
`newText || ""` before scanner use. The later reset setText call is untouched.
The proof guard rejects a planted intervening read. Two optional languageVersion
comparisons now explicitly guard undefined. The shebang regex ! remains a
required check: the preceding test guarantees this non-global regex matches.
Census reason: non-null assertions. non-null-check 6ae58a0 closes required !
expressions, but not the uninitialized literal initializer semantics.

Refreshed unpushed scratch compiler includes main e8ba3d5, the four feature
tips, fallthrough, import cycles and non-null-check (integration c56dae76).
Checker gate has no slice diagnostics. Next actual syntax refusal, in order:
`scanner.ts:498:19: Adamic 0.1 refuses a definite assignment assertion !; remove ! and initialize it where it is declared or in the constructor, or type it T | undefined`.
This is tokenValue, census reason: definite assignment assertions. No available
branch for that syntax or initialization was found; only non-null-check exists.

Independent minimal initializer probe, on non-null-check alone and integration:
`let text: string = undefined!; text = "assigned before read"; console.log(text);`
returns `stage 0 can't lower a value of type never yet`. Node prints the assigned
value. Removing the initializer in a control compiles and prints the same
line; removing the assignment instead is caught by TS2454. The compiler needs
to treat literal undefined!/null! initializers as uninitialized declarations,
rather than lower them as ordinary never-valued expressions. The requested
exact scanner rewrite is retained. I did not replace it with the compiling
control spelling. I stopped here; definite-assignment admission, subsequent
lowering, ownership and native token comparison remain unfinished. Adaptations
58-59 were not made. No native/Node equivalence is claimed.

Validated actual 55-57 edits: full upstream suite 106,367 pass, zero failures,
pending or baseline differences (336.493 seconds, two workers); Node 509,014
tokens and SHA-256 c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f.
The first adaptation-57 guard rejected comment text, so that attempt made no
edits; the corrected implementation and subsequent full retry are the evidence.
Logs: native57-{next-blocker,initializer-minimal,initializer-node,
initializer-control,early-read-mutant,intervening-read-mutant}.log;
feature refs: native57-feature-refs.json. Reproducers are in probes/.

Commands: scanner/run.sh OUT --tree SLICE --inputs FIXED --node-only;
stock upstream oracle/run.sh TREE OUT --workers 2; scratch adamic build
DRIVER -o OUT; oracle/node.mjs probes/uninitialized-non-null.a. Every test
writes a separate log. No compiler implementation was edited on this branch.
Toolchain remains prepared (setup total 96 seconds), nproc 5; the environment
file is /workspace/adamic-tools/env.sh. Pushes use only the own scanner branch.

## October 7: adaptation 56 closes the assertion predicate

Census reason: type predicates. Compiler feature: proven-predicates, admission
seam pending (remote still b4366a7). Eleven reached Debug.assert call sites
become direct Boolean guards with the same failure messages; only the now
unreached slice member is removed. No full-tree predicate contract is erased.
The required resultingToken narrowing and side-effectful regex escape test are
preserved. Exact next refusal on the earlier integrated compiler:
scanner.ts:291:12, non-null assertion !, in optional languageVersion comparison.

Node still matches 509,014 tokens. A real failing reScanQuestionToken call has
the same message; disabling its guard is caught by diff exit 1. Combined 55-57
upstream validation: 106,367 pass, zero failures/pending/baseline differences,
336.493 seconds, two workers. Evidence: native56-57-baseline-{report.json,tests.log},
native56-{next-blocker,assert-failure,assert-mutant}.log.

The initial overlapping baseline attempt lost a worker without diagnostics;
it is not a pass. The container reported an OOM kill, but attribution was not
proven. The isolated two-worker retry above completed the full suite.

## October 7: adaptation 55 closes function widening

Fetched first: no function-widening branch. Per compiler update none lands
before noon. Merged origin/main e8ba3d5 into this delivery branch without
rebasing. Adaptation 55 changes three reached Debug stack marker parameter
types from AnyFunction to {}, preserving every runtime statement and marker
identity. Census reason: method-signature-style. Compiler feature: function
widening, pending; temporary adaptation 55 closes this refusal.

Node: all 509,014 tokens byte-identical; planted token-end offset caught by
diff exit 1. Stack marker mutant caught with "marker was not preserved".
Full upstream baseline: 106,367 pass, zero failures/pending/differences.
Evidence: native55-baseline-report.json, native55-next-blocker.log and
native55-stack-marker{,-mutant}.log in evidence/.

Next actual refusal: debug.ts:18:133, assertion predicate
(adamic/no-type-predicate). proven-predicates remote remains b4366a7;
the promised 01a1143b admission seam is not yet available on that branch.
Plans 56 and 57 have already been pushed, before implementation.

## October 7: final refreshed feature probe and stopping point

Refreshed and merged the four current tips in the **unpushed** scratch branch:
taste `cabb7f1`, flag enums `246ecc0`, namespaces `39c3eca`, nested functions
`59fe216`. Fallthrough `5f77d33` and import cycles `6b17636` are also present.
Scratch conflicts retained already integrated feature implementations where
branches overlapped; scratch compiler changes are never part of this delivery.
Feature hashes and scratch head: [member-final-feature-refs.json](evidence/member-final-feature-refs.json).
The raw slice still has eleven checker diagnostics. The validated adapted
slice clears the checker and still refuses Debug.fail's `stackCrawlMark || fail`
function widening at debug.ts:13:67. Exact final logs:
[member-latest-raw.log](evidence/member-latest-raw.log) and
[member-latest-revised.log](evidence/member-latest-revised.log).

Candidate branches inspected, not merged: proven-predicates `b4366a7` retains
the blanket type-predicate syntax refusal; its admission seam is not wired.
non-null-check `6ae58a0` admits `!` using a nullish panic, which does not close
the scanner's valid undefined comparisons or omitted initial text with Node's
semantics. Applying it blindly would create a native/Node disagreement.
Optional adaptation 20 now has a newer fix at b7e379e; this proof deliberately
retains the requested 10+50 input and the independently validated slice-only
private-diag repair, without accepting any API snapshot differences.

I stopped before a native executable. The actual tracked-path blocker is
function widening in failure-stack capture. Untracked probes reached, in order,
the assertion predicate, languageVersion non-null assertion, shebang non-null
assertion, shebang indexed-read checker diagnostic, and textInitial non-null
assertion. Those probes are not validated adaptations. I did not complete the
remaining lowering/ownership sequence or a one-feature-only matrix beyond
main-plus-fallthrough. The Node scanner slice proof and omission mutant are
complete; the requested native deadline is not yet fulfilled by this branch.

All pushes are to codex/stage3-scanner-proof, without force. No main push,
rebasing of pushed history, PR, production compiler edit or other worker's
fixture change occurred. Final syntax, byte-span audit, idempotence and git
whitespace checks pass. Toolchain is the already prepared Go 1.27.1 / clang
20.1.8 / Node 24.19.0 environment, nproc 5 (original setup total 96 seconds).

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

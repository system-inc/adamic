# Land-area-next 28285421: fresh scratch, native red

records-maplike-library 2648ad33 conflicts in 44 paths, including lowering,
loading and native. Merge aborted as instructed; compiler is land-area-next
28285421 alone. No compiler merge is pushed. Adaptations 65 and 76 remain applied,
sameMap remains unchanged. Both split modes now stop in the checker at
**debug.ts:113:19 TS2339 captureStackTrace missing on ErrorConstructor**, with the
second use at 114:19. The previous library scratch's Error fact is absent here.
No C, clang time, native parser dump or native acceptance corpus is reached.

Earlier families, measured by isolated probes: core.ts:11:52 enum initialization
persists; all three namespace forms persist. Predicate overload and callback
result probes now pass, but the predicate callback-parameter probe still refuses.
arguments.length and its function-value witness both match Node, with byte mutants
caught (the latter prints 1). Six of twenty focused probes pass; every green has
identical stdout, stderr and exit, plus a caught one-byte output mutant.
MapLike and the sameMap/key-array casts still refuse in focused probes.

The fresh byte-audited slice has 27 declaration files, 1993 code declarations,
41857 copied-span lines and 2082 spans across 79 modules. Extended full/slice Node
dumps are identical: 36,429,231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Node-end and both JSDoc mutants are caught on both trees. 76's contract mutants
and 32 local lane tests pass. No adaptation changes or new API sanctions; prior
full landing-lane proof remains the baseline. Error-recovery manifest not rerun.

Ten ordered source stops follow, all after row 1 behind uncommitted throwing
placeholders. Rows 9-10 are observed on helper calls after the enum placeholders;
the unchanged Map-before-namespace minimal reproduces that namespace preflight.
Observed column 109 on row 10 maps to original cache initializer column 29.
Further stubs create TS2454 and an inferred void-result TS2339; an earlier attempt
also duplicated a helper. These artifacts are excluded, and discovery stops here
rather than claiming fifteen source failures. Exact diagnostics, failed attempts,
placeholder patches and merge paths: [front33 evidence](evidence/front33/README.md).

| Order | Original slice location | Exact diagnostic | Minimal program |
|---|---|---|---|
| 1 | src/compiler/debug.ts:113:19 | error TS2339: Property 'captureStackTrace' does not exist on type 'ErrorConstructor'. | [native-error-capture-stack.a](native-error-capture-stack.a) |
| 2 | src/compiler/core.ts:11:52 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) |
| 3 | src/compiler/debug.ts:333:29 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) |
| 4 | src/compiler/performanceCore.ts:37:37 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-call-before-enum.a](native-call-before-enum.a) |
| 5 | src/compiler/performance.ts:15:18 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-call-before-enum.a](native-call-before-enum.a) |
| 6 | src/compiler/performance.ts:16:15 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) |
| 7 | src/compiler/performance.ts:17:16 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) |
| 8 | src/compiler/performance.ts:18:19 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) |
| 9 | src/compiler/core.ts:11:52 | stage 0 can't lower a call before all runtime namespaces are initialized; put namespaces before executable module code yet | [native-map-before-namespace.a](native-map-before-namespace.a) |
| 10 | src/compiler/debug.ts:333:29 | stage 0 can't lower a call before all runtime namespaces are initialized; put namespaces before executable module code yet | [native-map-before-namespace.a](native-map-before-namespace.a) |

Earlier units below are historical evidence, not this scratch’s green claims.

---

# Current unit: library merge and permanent finite-key contract

Native red on scratch 2d7c31ac (main ffe6efc1 + area/library b05a9306).
ErrorConstructor's two checker diagnostics vanish. Both final slice build modes
stop at core.ts:11:52 before enum initialization; no C or native acceptance run.
Unsplit attempt 0.5679s, split 0.5193s (pre-clang wall, not clang time).
The isolated Error probe now refuses ambient host method presence at 4:9.

Namespace-read e1ebac45 conflicted in 16 files, including lower and native. It
was aborted, not manually resolved, as requested. The isolated namespace `this`,
class and merged-function probes all still refuse on the retained scratch.

Permanent 76 narrows AssertionKeys to the six actual private cache keys. An
ownership audit rejects new writes, keys, escapes and altered enumeration. Both
full landing lanes PASS: 106366 passing, one identical sanctioned API failure,
zero pending. The API diff and ten emitted JavaScript artifacts match main by
byte. The key-array cast's finite narrowing is justified; native array-view
support remains a compiler gap. The existing cast is retained, with a truthful
six-key target, and no new narrowing assertion is introduced.

sameMap is a tsc source-contract hole: its callback can mutate the original
input, leaving T in the returned array despite a U-only overload. Its truthful
readonly (T|U)[] proposal erases identically and has a contract mutant, but full
stock checking reveals builder.ts:564 TS2322. Widening the private builder helper
reveals next-field assignment failures at 545, 551 and 555. The proposal is NOT
in apply; no cast hides the mismatch. This part is incomplete pending a truthful
builder/diagnostic contract or approved runtime repair. Exact proposal patches
and compiler diagnostics are in [evidence/front32](evidence/front32/README.md).

The final slice has 27 declaration files, 1993 code declarations, 41857 copied
lines, 2082 audited spans and 79 ordered modules. Full/slice Node identity remains
36,429,231 bytes, SHA256 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Node-end, dropped JSDoc tags and JSDoc diagnostic mutants are caught on both.
32 local lane tests pass. The existing recovery manifest is not rerun this unit.

The ordered walk below uses the prior validated slice, before 76's finite-key
alias. The final 76 slice was separately audited and rebuilt in both modes and
has the same first stop. Rows 2–15 are found behind uncommitted throwing discovery
placeholders. Exact edits (including overload predicate widening to boolean),
messages, stdout and stderr are retained; these are not source adaptations.

| Order | Slice file:line:column | Exact diagnostic | Minimal program | Known feature branch (not admitted by this scratch) |
|---|---|---|---|---|
| 1 | src/compiler/core.ts:11:52 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) | codex/enum-init-reach 2152fc3b (historical covering probe; not merged here) |
| 2 | src/compiler/debug.ts:333:29 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) | codex/enum-init-reach 2152fc3b (historical covering probe; not merged here) |
| 3 | src/compiler/performanceCore.ts:37:37 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-call-before-enum.a](native-call-before-enum.a) | codex/enum-init-reach 2152fc3b (historical covering probe; not merged here) |
| 4 | src/compiler/performance.ts:15:18 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-call-before-enum.a](native-call-before-enum.a) | codex/enum-init-reach 2152fc3b (historical covering probe; not merged here) |
| 5 | src/compiler/performance.ts:16:15 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) | codex/enum-init-reach 2152fc3b (historical covering probe; not merged here) |
| 6 | src/compiler/performance.ts:17:16 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) | codex/enum-init-reach 2152fc3b (historical covering probe; not merged here) |
| 7 | src/compiler/performance.ts:18:19 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](native-enum-map.a) | codex/enum-init-reach 2152fc3b (historical covering probe; not merged here) |
| 8 | src/compiler/corePublic.ts:9:5 | Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added | [native-maplike-index.a](native-maplike-index.a) | codex/records-lowering (historical branch coverage; not merged here) |
| 9 | src/compiler/core.ts:28:37 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [native-nonnull.a](native-nonnull.a) | codex/non-null-check (historical branch coverage; not merged here) |
| 10 | src/compiler/core.ts:42:30 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [native-nonnull.a](native-nonnull.a) | codex/non-null-check (historical branch coverage; not merged here) |
| 11 | src/compiler/core.ts:54:101 | Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate) | [native-predicate-overload.a](native-predicate-overload.a) | codex/proven-predicates (historical branch coverage; not merged here) |
| 12 | src/compiler/core.ts:76:113 | Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate) | [native-predicate-overload.a](native-predicate-overload.a) | codex/proven-predicates (historical branch coverage; not merged here) |
| 13 | src/compiler/core.ts:99:23 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [native-nonnull.a](native-nonnull.a) | codex/non-null-check (historical branch coverage; not merged here) |
| 14 | src/compiler/core.ts:110:34 | Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case | [native-nonnull.a](native-nonnull.a) | codex/non-null-check (historical branch coverage; not merged here) |
| 15 | src/compiler/core.ts:124:65 | Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate) | [native-predicate-overload.a](native-predicate-overload.a) | codex/proven-predicates (historical branch coverage; not merged here) |

The twelve focused probes all execute on Node and all refuse natively on this
scratch (0/12). No byte-output native mutant can run for these refused builds.
Main remains untouched, scratch compiler merges are never pushed. Earlier
measurements below are historical, not this unit's pass claims.

---

# Current main and temporary Node builtins: adaptation green, native red

Driver inputs use readTextFile and programArguments from adamic. The erased
node:fs activation import is removed. Both the full tree and the newly gathered
slice produce 36,429,231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615,
with empty stderr; cmp exits 0. Node-end and both JSDoc mutants are caught.
The audited re-cut has 27 declaration files, 1,994 code declarations,
41,861 copied-span lines and 2,083 spans, across a 79-module evaluation graph.
The archive at slices/current.tar.gz contains the validated tree and driver.

Temporary 65 now groups the erased Node builtin contracts. Current main already
has the minimal local process view, so full apply edits only performanceCore.ts
and tracing.ts: 7 lines added, 2 removed. Both complete stage3 lanes PASS against
main efe9f404: 106,366 passing, 1 sanctioned API failure, 0 pending; exactly
api/typescript.d.ts differs, with identical 222-line composed sanctions. Install
and build exit 0; the expected upstream/oracle exit is 1. All ten newly built
JavaScript artifacts are byte-identical. No new API sanction is added.

On current main both native split modes stop before C at debug.ts:113:19:
TS2339 Property captureStackTrace does not exist on type ErrorConstructor.
The second use at 114:19 has the same error. This is the measured checker result,
even though Error facts were reported as landed. Build-attempt wall times are
0.5710s unsplit, 0.6212s split; these are not clang times. There is no parser C,
binary, native dump comparison, performance measurement or native-output mutant.

The following fifteen source stops were found in a separate uncommitted copy.
All after row 1 are behind throwing discovery placeholders. Two source columns
are mapped back from edited callback types; the raw observed columns and exact
messages are retained in ordered-stops.json. Return-predicate stops were reached
only after widening the preceding overload callback contracts.

| Order | Original slice location | Minimal program | Covered on a branch |
|---|---|---|---|
| 1 | src/compiler/debug.ts:113:19 | [native-error-capture-stack.a](native-error-capture-stack.a) | Error checker fact missing on observed main; no covering clean merge verified |
| 2 | src/compiler/core.ts:11:52 | [native-enum-map.a](native-enum-map.a) | codex/enum-init-reach 2152fc3b, previously reported green; absent from main |
| 3 | src/compiler/debug.ts:333:29 | [native-enum-map.a](native-enum-map.a) | codex/enum-init-reach 2152fc3b, previously reported green; absent from main |
| 4 | src/compiler/performanceCore.ts:37:37 | [native-call-before-enum.a](native-call-before-enum.a) | codex/enum-init-reach 2152fc3b, previously reported green; absent from main |
| 5 | src/compiler/performance.ts:15:18 | [native-call-before-enum.a](native-call-before-enum.a) | codex/enum-init-reach 2152fc3b, previously reported green; absent from main |
| 6 | src/compiler/performance.ts:16:15 | [native-call-before-enum.a](native-call-before-enum.a) | codex/enum-init-reach 2152fc3b, previously reported green; absent from main |
| 7 | src/compiler/performance.ts:17:16 | [native-call-before-enum.a](native-call-before-enum.a) | codex/enum-init-reach 2152fc3b, previously reported green; absent from main |
| 8 | src/compiler/performance.ts:18:19 | [native-call-before-enum.a](native-call-before-enum.a) | codex/enum-init-reach 2152fc3b, previously reported green; absent from main |
| 9 | src/compiler/corePublic.ts:9:5 | [native-maplike-index.a](native-maplike-index.a) | codex/records-lowering 456c981b, absent from main |
| 10 | src/compiler/core.ts:28:37 | [native-nonnull.a](native-nonnull.a) | codex/non-null-check a02613ef, absent from main |
| 11 | src/compiler/core.ts:42:30 | [native-nonnull.a](native-nonnull.a) | codex/non-null-check a02613ef, absent from main |
| 12 | src/compiler/core.ts:54:101 | [native-predicate-overload.a](native-predicate-overload.a) | codex/notyet-predicates c191f97b, attempted merge conflicted and aborted |
| 13 | src/compiler/core.ts:54:116 | [native-some-predicate-overload.a](native-some-predicate-overload.a) | codex/notyet-predicates c191f97b, attempted merge conflicted and aborted |
| 14 | src/compiler/core.ts:56:128 | [native-some-predicate-overload.a](native-some-predicate-overload.a) | codex/notyet-predicates c191f97b, attempted merge conflicted and aborted |
| 15 | src/compiler/core.ts:76:113 | [native-predicate-overload.a](native-predicate-overload.a) | codex/notyet-predicates c191f97b, attempted merge conflicted and aborted |

The discovery helpers replace function bodies with throw, preserve inferred
variable types and explicitly preserve tryGetPerformance's return type. The
MapLike declaration becomes Record<string,T> with an unused throwing typed
helper. Overload predicate parameter contracts are widened and the implementation
is made throwing; later bodyless return overloads are removed. These edits have
no Node-identity or acceptance claim. Trial synthetic inference/syntax errors
and a reintroduced duplicate enum stop are excluded from the fifteen-source
list; raw failed attempts remain evidence. The last find placeholder is recorded
without another build after the limit. No discovery source is an adaptation.

## Requested feature trials

Fresh unpushed feature scratch starts at efe9f404. Generic-function-value
5d1b45e1 merges cleanly. Its isolated empty-array conditional prints `0\n1\n`
on Node and native, stderr empty, cmp 0; a one-byte output mutant gives cmp 1.
The probe binary is 361,104 bytes. Main refuses this probe at array-of-never
lowering. The older whole toSorted witness still stops at its independent
SortedReadonlyArray cast, so it does not test the conditional in isolation.

Predicates c191f97b was found on codex/notyet-predicates and attempted; conflicts
were aborted. Closure-convention 22fd701a also conflicted and was aborted,
including lowering files. No conflicts were resolved. Both full-slice modes on
the clean generic-function-value merge retain the same Error checker stop.
Predicate and arguments.length probes remain refused; no disappearance is
claimed from an unmerged fix. Namespace fixes are still with compiler.

## The two casts

Classify both as source-proof/edit candidates, not permission to admit unchecked
array casts. native-same-map-return-cast.a isolates the T-to-U return view;
native-key-array-cast.a isolates string[] to finite AssertionKeys[]. Both are
refused by main and the clean feature scratch.

There is also a concrete sameMap source counterexample: native-same-map-side-effect.a
keeps its implementation statements and uses a type-correct string-returning
callback that changes the earlier captured input slot to a number. Node prints
`number`, although the result type promises strings. Equality of each mapped
item at its own iteration therefore does not prove the final array view under
callback mutation. This is an upstream source-contract candidate, not a native
miscompile: native compilation stops first at the checked indexed read. The
smaller cast-only witness reaches the actual cast refusal. Actual tsc callers
would need separate invariant evidence for a checked/type-only source repair;
none is implemented in this unit.

The assertion-cache key case has a local finite-key ownership invariant, but
getOwnKeys returns string[]. A local typed enumeration or checked key narrowing
would expose that invariant; the reduced cast admits arbitrary string input
and cannot justify blanket compiler narrowing. No key-array source edit is
implemented here.

## Gate and limitations

The first adapter gate found a missing writeFileSync member; it was added before
any final push. Unbounded upstream worker runs exited without counts; bounded
Node heaps and sequential full lanes pass. No failures were accepted. The
adapter is idempotent, and its JS-equality guard catches a ! to !! mutation of
tsc's actual browser test. All 32 local lane tests pass, including their planted
API/count/failure-title mutants. Fifteen distinct stop/feature probes run on
Node with exit 0; main rejects all fifteen, and the clean feature scratch builds
only the isolated conditional. Its native comparison and byte mutant pass.
The extra sameMap source counterexample runs on Node and remains natively refused.

Tool setup: Go ready 0.029s, Node 0.031s, clang 0.246s, markdown install 0.811s
(ready 0.908s), submodules 239.908s; nproc 5. Initial scratch SDK replacement
used the wrong module name, then the pinned TypeScript/tsc replacement corrected
it; compiler build exits 0. No compiler edits or full compiler gate are claimed.

[Final evidence](evidence/front31/result.json),
[ordered exact diagnostics](evidence/front31/ordered-stops.json),
[branch attempts](evidence/front31/parser-front31-feature-merges.json) and raw
compressed logs include commands, outputs, mutants and the exact discovery diff.

---

# Main-based parser run: fifteen ordered stops, native red

Fresh main ef3141e9 plus clean namespace 47a6fabe, enum reachability 2152fc3b
and front3 a36d1c04 merges yields unpushed scratch cadeac4f. Front3 was added
for the Node-loader stop and merged cleanly, but did not clear it. Remaining
library 51761b0 and callable views 37fa06d3 merges conflicted and were aborted;
no compiler edits or conflict resolutions. Front3 supplies method-presence,
predicates-2 and debugger. Full ancestry/merge logs are retained.

Both split modes exit 1 on the unchanged validated slice with eight checker
diagnostics. First: parser-proof-main.a:3:21 TS2591 Cannot find name node:fs.
Node dependencies are installed; the clean merge lacks their loader. Unsplit
0.4154s, split 0.4679s, repeat 0.4716s are build-attempt wall times, not clang
times. No C, binary, native comparison, performance or native-output mutant.

| Order | Slice location | Stop | Minimal program |
|---|---|---|---|
| 1 | parser-proof-main.a:3:21 | Node filesystem binding | [native-node-fs-import.a](native-node-fs-import.a) |
| 2 | src/compiler/core.ts:883:25 | NodeJS namespace | [native-nodejs-process-type.a](native-nodejs-process-type.a) |
| 3 | src/compiler/performanceCore.ts:37:37 | require / perf_hooks | [native-require-perf-hooks.a](native-require-perf-hooks.a) |
| 4 | src/compiler/tracing.ts:24:27 | fs type import | [native-fs-type-import.a](native-fs-type-import.a) |
| 5 | src/compiler/tracing.ts:47:22 | require / process | [native-require-fs.a](native-require-fs.a) |
| 6 | src/compiler/core.ts:195:37 | sameMap array return cast | [native-same-map-return-cast.a](native-same-map-return-cast.a) |
| 7 | src/compiler/core.ts:421:34 | toSorted readonly empty conditional | [native-sorted-empty-conditional.a](native-sorted-empty-conditional.a) |
| 8 | src/compiler/core.ts:525:17 | arguments.length | [native-arguments-length.a](native-arguments-length.a) |
| 9 | src/compiler/core.ts:594:5 | isArray predicate | [native-array-is-array-predicate.a](native-array-is-array-predicate.a) |
| 10 | src/compiler/core.ts:616:97 | predicate callback parameter | [native-predicate-callback-parameter.a](native-predicate-callback-parameter.a) |
| 11 | src/compiler/debug.ts:429:5 | namespace this receiver | [native-namespace-object-receiver.a](native-namespace-object-receiver.a) |
| 12 | src/compiler/debug.ts:526:5 | namespace this receiver | [native-namespace-object-receiver.a](native-namespace-object-receiver.a) |
| 13 | src/compiler/debug.ts:739:5 | class inside namespace | [native-namespace-class.a](native-namespace-class.a) |
| 14 | src/compiler/debug.ts:50:5 | function merged with namespace | [native-callable-namespace.a](native-callable-namespace.a) |
| 15 | src/compiler/debug.ts:80:31 | key-array cast | [native-key-array-cast.a](native-key-array-cast.a) |

[Ordered evidence](evidence/front30/ordered-stops.json) retains exact messages,
coordinates and placeholder edits. All rows after 1 are found behind stubs;
rows 3 and 5 each remove additional host diagnostics in their throwing bodies.
The five checker sites cover all eight original diagnostic rows.

Fourteen distinct minimal programs run on Node with exit 0 and empty stderr;
all fourteen native builds reject the corresponding family. The shared receiver
minimal covers rows 11 and 12. No silent miscompile is observed in this run.

Discovery changes remain uncommitted in a separate copy. The compressed exact
patch is evidence, not an adaptation. Host placeholders have intentionally any
return types; removed mode initialization is replaced by a throwing typed
initializer. A namespace class becomes an interface plus a throwing value
initializer; the unused merged log function is renamed and made throwing.
These changes have no Node identity or native acceptance claim. Three artificial
TS2454 errors from the first pass were excluded and fixed in the canonical
discovery pass. Row 15 has a recorded placeholder but no subsequent build: the
requested fifteen-stop limit was reached.

[Final run report](evidence/front30/result.json) and compressed raw logs retain
commands, outputs and split-mode results. The prior Node references remain the
acceptance targets; this run cannot compare native output against them.

---

# Main-based parser run: first checker stop, continuing discovery

Fresh scratch uses origin/main ef3141e9 plus clean namespace 47a6fabe and enum
reachability 2152fc3b merges, unpushed ca36518c. Library 51761b0, method-presence
16cb9b10, predicates-2 0f722679, callable views 37fa06d3 and debugger 7d2cbc89
were not ancestors; each conflicted and was aborted without resolutions.

Compiler Go build exits 0. Both native split modes on the unchanged validated
slice exit 1 in checking, before lowering/C/clang. First diagnostic:
parser-proof-main.a:3:21 TS2591, Cannot find name 'node:fs'. Seven more diagnostics
refer to NodeJS, require, perf_hooks, fs and process. Node API dependencies are
installed, but this clean scratch has no Node-types loader implementation.

Minimal native-node-fs-import.a prints `function\n` on Node, exit 0 and empty
stderr; native build exits 1 with the same TS2591. No native parser binary or
byte comparison is available. Initial split/off attempt times are 0.4690 and
0.4681 seconds, not clang times; no warm-cache claim.

[First-stop evidence](evidence/front30/first-stop.json) is pushed early as
requested. Work continues: try front3 only because this stop needs its loader,
and retain it only if the merge is clean. Discovery placeholders will remain
in a separate uncommitted source copy; no adaptation or compiler edit.

---

# Restart from front-3: stopped on convention merge conflicts

Fresh scratch started from front-3 0059e65c. Convention 22fd701a was not an
ancestor and its merge conflicted in 28 files, including
internal/lower/non_null.go, closure emission/runtime, record and array handling.
The merge was resolved in scratch before the subsequent user instruction to
leave these conflicts to compiler/front-3's merger. No compiler source or merge
is pushed. The user instruction now supersedes continuing this scratch.

Area/stage3 6a8eebf4 merged cleanly. Ancestry checks led to namespace 47a6fabe,
enum-init-reach 2152fc3b, host-method-presence 51761b0, method-presence-test
16cb9b10, views-callables 37fa06d3 and debugger 7d2cbc89 merges, with their
conflicts resolved locally. Predicates-2 0f722679 was already included. Scratch
68c9fdbd contains every requested input but remains unvalidated; its integration
resolutions must not be treated as compiler's accepted decisions.

Intermediate merge-boundary Go errors were corrected locally, including missing
incoming view metadata/helpers and duplicate definitions. Final Go compiler
build exited 0 before the stop instruction was observed. This establishes only
a buildable Go artifact, not compiler semantics. Exact logs/scripts and full
scratch diff are retained for review; no native run followed.

I did not build the slice in either split mode, confirm a first source stop,
rerun green probes, run new mutants, or attempt either native acceptance reference
or performance metrics. The validated slice and references remain unchanged.
No new source re-cut was performed. The old quarantined scratch was not used.

[Evidence](evidence/front29/report.json) records original convention conflicts,
ancestry checks, local resolutions, failures and final Go build. Per the user's
latest instruction, stop here and wait for front-3's SHA with the reconciled
convention. The next run starts fresh from that integration, not this scratch.

---

# Front-3 retry: integration merge blocked before build

Fresh unpushed scratch starts at origin/compiler/stage3-front-3 21243ee4.
All eleven requested fix SHAs were checked individually against this base with
merge-base --is-ancestor: none are ancestors. Namespace 47a6fabe was merged
with scratch-only resolutions, yielding ebaa8d7d. These touch
lowering, JS and native readiness handling; gofmt succeeds, but the resolution
has not been built or validated. Exact patch/script retained.

The next ancestry-checked merge, host-blockers b788e96e, exits 1 with 38
unresolved files. It conflicts across lowering, IR, library loading, native
emission and runtime. Two observed representation conflicts explain why this
is an integration stop rather than a source blocker:

- front-3 closure/method ABI is `(self, argument_count, arguments)`;
  host-blockers is `(self, arguments, argument_count)`. Both runtime headers,
  emitted definitions and call sites conflict.
- front-3 IR uses `Break{Depth int}` and `Continue{}`;
  host-blockers uses `Break{Label string}` and `Continue{Label string}`.

No new wrong-output result is claimed. The scratch merge remains unresolved;
no compiler build or probe rerun was performed. Thus I did not confirm the
expected enum-initialization stop on this scratch, run either split mode, or
run new byte mutants. The remaining nine merges have not been attempted.
This avoids claiming a compiled compiler from an incomplete ABI integration.
Arguments-length 9534e8ab is absent from both the committed scratch and incoming
host-blockers branch. Quarantined scratch, published slice and references remain
unchanged. No compiler scratch merge is pushed.

[Evidence](evidence/front28/report.json) records every ancestry result, the
38 paths, merge logs, namespace resolution and unresolved conflict patch.
Holding for compiler's reconciled enum/convention or front-3 update.

---

# Fresh fixes: targeted probes green, earlier integration stops remain

Fresh unpushed scratch 1b382081 plus proven-predicates-2 0f722679,
views-callables 37fa06d3 and debugger-statement 7d2cbc89 yields 9d2696bb.
Records-lowering 456c981b remains an ancestor. Arguments-length 9534e8ab and
quarantined scratch 33bf53ac are absent. No compiler merge is pushed.

Five targeted witnesses match Node, exit 0 with empty stderr, and catch an actual
native stdout one-byte mutation with cmp exit 1: predicate callback declaration,
assertIsDefined declaration, namespace-free callable marker, debugger declaration,
and finite partial-record view. Four of five unchanged probes pass. The original
namespace-bearing native-function-any-view.a is refused at its namespace; the
additional namespace-free native-function-any-view-called-only.a matches the
incoming branch's parser-cache reduction. Predicate/assertion/debugger witnesses
compile unused declarations and print loaded, so they do not establish runtime
predicate semantics or executing throw/debugger behavior.

The untouched re-cut slice stops at core.ts:11:52, not core.ts:525:
`stage 0 can't lower an indirect call or class construction before enum
initialization; declare enums before executable module code yet`.
Unsplit, split and repeated split all exit 1 before C emission, in 0.8231,
0.8247 and 0.8234 seconds respectively. These are compiler attempt times, not
clang times. No C, binary, native parser comparison or warm-cache claim.

The assumption is to keep exactly the requested fresh scratch and merges,
without importing historical fixes from the quarantined scratch to force the
expected 525 stop. The fresh scratch still lacks namespace, generic-empty-array,
Array.isArray-predicate and method-presence support observed below.

Only in a discovery copy, reduceLeft's body is replaced with the previously used
throw stub. Its arguments.length read is therefore removed in that copy only;
it was not reached on the untouched slice. No arguments-length probe was run.
Then these stops were observed in order, ten behind the first real stop:

| Order | Slice site | Observation | Minimal program |
| --- | --- | --- | --- |
| 1 | core.ts:11:52 | Map construction before pending enum | native-enum-map.a |
| 2 | debug.ts:333:29 | enumMemberCache Map construction before pending enum | native-enum-map.a |
| 3 | performanceCore.ts:37:37 | require call before pending enum | native-call-before-enum.a |
| 4 | performance.ts:15:18 | timestamp callable value before pending enum | native-call-before-enum.a |
| 5 | performance.ts:16:15 | marks Map construction before pending enum | native-enum-map.a |
| 6 | performance.ts:17:16 | counts Map construction before pending enum | native-enum-map.a |
| 7 | performance.ts:18:19 | durations Map construction before pending enum | native-enum-map.a |
| 8 | core.ts:230:22 | never seen as mutable generic U in flatMap fallback | native-generic-empty-array.a |
| 9 | core.ts:594:5 | isArray(unknown) predicate return is not trusted | native-array-is-array-predicate.a |
| 10 | core.ts:889:14 | process.nextTick truthiness read refused as unbound method | native-process-next-tick-read.a |
| 11 | debug.ts:26:1 | namespace refused | native-debug-namespace.a |

Rows after 1 are found behind cumulative stubs, plus the upfront reduceLeft stub;
this is discovery order, not the eventual unchanged compiler order. Initializers
are replaced with undefined! while retaining their declared/inferred Map or
number type. Function bodies become throws; tryGetPerformance's inferred return
shape is explicitly preserved to avoid artificial checker errors. Discovery
patch and every exact diagnostic/edit are retained. No discovery source is in
the published slice or a temporary adaptation. The enum-map and typed-call
minimals reproduce their common refusal family; they do not individually model
the Node require host dependency. All listed family minimals run on Node and
are refused by this compiler with the corresponding diagnostics.

Merge conflicts touched lowering: predicate refusals; callable casts, expressions,
invariance, object/statement dispatch and refusals; debugger refusals. Resolution
retains records hooks alongside incoming predicate/callable hooks. Exact scratch
diff is retained. Two incomplete hook returns in the first resolution caused
an intermediate build failure, then were corrected; final compiler build passes.
The callable merge also exhausted disk during checkout. Clearing the rebuildable
21 GB Go cache and recovering that checkout allowed completion. SDK pins remain
cohere 7945d102 and TypeScript d92d9bfee through local absolute module paths.
Missing API-seat/cwd and artificial stub checker failures are retained separately;
they are not counted as real source blockers.

[Evidence](evidence/front27/report.json) and
[ordered diagnostics](evidence/front27/ordered-stops.json) retain commands,
outputs, mutations, conflicts and scope limits. Node references and validated
slice are unchanged. No full gate or native acceptance run. Holding for the
closure convention and a compiler integration that carries the other fixes.

---

# Fresh isolated partial-record probe: green

Fresh scratch starts from parser c109c06d, merges area/stage3 b61e7064 and
records-lowering 456c981b, yielding unpushed 1b382081. Both merges are conflict-free.
Ancestry checks prove arguments-length 9534e8ab and quarantined scratch 33bf53ac
are absent. The existing compiler SDK pins are reused through scratch-only
module paths; no compiler implementation change. The quarantined scratch is
untouched, and no native parser acceptance was attempted.

Unchanged native-partial-record-view.a now builds and runs natively. Node and
native both print `0\n`, exit 0, and have empty stderr. stdout and stderr cmp
each exit 0. A one-byte mutation of actual native stdout (`0` to `1`) is caught
by cmp exit 1. Probe binary is 360,312 bytes. This clears the former finite
Partial<Record> seen as MapLike storage refusal for this exact empty-record
getOwnKeys witness; no broader dictionary semantic coverage is claimed.

[Evidence](evidence/front26/report.json) records the fresh ancestry, commands,
outputs and byte mutant. Only this probe was run, as requested. No arguments
probe, full parser build, feature tests, full gate or corpus/native acceptance.
Published slice and references remain unchanged. Disk-space checkout/link
failures and the intervening wrong-directory build failure are retained; final
correct-directory compiler build succeeds. Cleanup retained compressed case
dumps and removed only completed disposable copies, a clean unused worktree and
six obsolete rebuildable scratch compiler executables.

Evidence only is pushed on the parser branch. Holding for the closure convention.

---

# Re-cut parser slice: Node green, scratch function-value count miscompile

A silent miscompile is detected in the unpushed scratch integration. The minimal
native-arguments-length-value.a prints `9\n` natively instead of Node's `1\n`;
both runs exit 0 with empty stderr, build exits 0, and cmp exits 1. This is a
function value forwarding arguments.length, not a native parser result. The
larger uncached arguments oracle also finds release wrong-output results in
value_count (last line 0 instead of 3) and reduce_left (last calls 2/undefined
instead of 12/112), plus clang ABI/packing failures and sanitizer/exit failures.
The scratch is unsafe for native acceptance pending the closure-convention fix.
No compiler merge or implementation is pushed.

Source is current origin/area/stage3 b61e7064, containing requested 234ab1aa.
Fresh apply carries isArray(value: unknown), void brands and Error adaptations.
It has no adaptation 48 directory and retains memoize's original body; the
assumption is to use this exact tip, without inventing the missing adaptation.
Temporary 80 is absent and not reintroduced. All parser temporaries 60–65 remain:
removing 60/61 separately restores TS2412, 62 TS2769, and 63/64 TS2345. The current
area still contains `(process as any).browser` and its older local process
contract, so 65 was applied before gathering with its existing stock-JavaScript
byte-identity guard. Its original cast probe refuses independently; removing 65
in the whole slice is masked by the earlier callback predicate refusal.
None of the new source makes a parser temporary unnecessary; none is dropped.

The extended-driver slice has 27 declaration files and 1,991 code declarations,
plus 77 facade records and 12 namespace-wrapper spans (2,080 copied spans,
41,853 copied-span lines). The preserved runtime graph has 79 modules, 52
without reached declarations. The scanner-owned verifier passes all 2,080 source
span hashes and 79 ordered import lists. No discovery stub is in this slice.
The published payload is slices/current.tar.gz with summary/hash in current.json;
recut.sh reproduces it using the scanner-owned tool. TypeScript license included.

Full-tree Node and slice Node each match the fixed 81-file reference byte for
byte: 36,429,231 bytes, SHA-256
`686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615`.
Inputs are the same pinned 81-file corpus; the newly adapted tree is the parser
implementation, not a replacement corpus. All 10,406 case records also equal the
existing manifest byte for byte (manifest SHA-256
`eaeca6000e53262767c67b9bedcb4ac94c4af1b5add518bd34b79b56cccaa6de`):
1,996 parse diagnostic rows, 34 JSDoc diagnostic rows, 1,323,111 nodes. Full and
slice output is equal case by case. The slice's node-end, dropped-tags and JSDoc
diagnostic mutants are caught. The full runner hit a disk-full error while
copying unrelated upstream tests for its JSDoc mutant; its successful baseline
and end mutation are retained. After deleting only that failed disposable copy,
the JSDoc checks pass on byte-identical full compiler sources without those tests.

Merged debugger-statement 7d2cbc89 and arguments-length 9534e8ab into unpushed
scratch 33bf53ac. Both split modes on this validated slice now stop at
core.ts:616:97: `a type predicate whose return is not proven (there is no body
proving this parameter)` in cast's callback type. The former arguments.length
stop and isArray(any) stop clear. Unsplit, split and repeated split attempts exit
1 in 0.867, 0.921 and 0.920 seconds, before C. Those are build-attempt times,
not clang measurements; repeat has no warm cache. The existing callback
signature minimal program reproduces this diagnostic family.

The original eleven probes are 10/11 green, with all ten native-output byte
mutants caught. native-arguments-length.a now matches Node `false\ntrue\n`,
both exits 0, equal empty stderr, cmp 0, mutant cmp 1. The remaining original
cast probe is superseded in the actual slice by temporary 65. This small probe
count does not imply the arguments feature or scratch integration is green.
Focused arguments lowering tests pass in 0.710 seconds; debugger flow passes in
0.009 seconds. The lower debugger regex selected no tests, so no lower debugger
pass is claimed. The uncached arguments/debugger oracle exits 1 in 10.887 seconds;
its full failures, including wrong-output cases, are retained separately.

Lowering merge conflicts touched expression dispatch, argument lowering,
function/receiver metadata and checked array views. Existing Node/host paths,
overload-result checks and view reads were retained while incoming count-aware
metadata was combined. The scratch already has three-argument host/runtime
closures; the arguments branch explicitly keeps its legacy two-argument packing
convention. The preserved host dispatch does not populate its packed count slot,
consistent with the reproduced 9-versus-1 failure. This is an integration
inference, not a claim that the upstream arguments branch fails in isolation.
The exact resolution patch and scripts are retained for the forthcoming closure
fix. No compiler bypass or attempt to hide failed expectations.

No parser C or binary was emitted, so native 81-file/case comparisons, parser
output mutant, C size, clang split/unsplit/warm time, binary size and native
best-of-three time remain unavailable. nproc is 5; previous Node user time is
6.162 seconds. No full gate, counts gate, complete oracle or WASI run claimed.
[Evidence](evidence/front25/report.json) and [slice](slices/current.json).

---

# Memoized namespace rerun: arguments.length first, ten stops behind stubs

Merged namespaces-tsc 47a6fabe into unpushed scratch a0aa1e69, yielding
e4cc453a. The only merge conflict was oracle/counts.md; both sides' rows
were retained. No lowering conflict or compiler implementation edit. Compiler
rebuild passes. All focused namespace tests now pass in 1.181 seconds, including
the three expectations that failed on the previous merge. The 24-level two-edge
call graph emits C in 0.064 seconds (previously exceeded 30 seconds); all five
size/control probes emit C and print `true\n` on Node.

The unchanged validated parser slice now stops at core.ts:525:17:
`Adamic 0.1 refuses arguments; name the parameters, or take a rest parameter`.
Unsplit, split and repeated split builds all exit 1 at this exact site, in
0.917, 0.874 and 0.969 seconds. These are pre-C build attempt times, not clang
times. No C, binary or cache objects were produced; repeat is not warm.
The `arguments.length` ruling still needs its compiler implementation; no source
adaptation was made. The eleven probes remain 9/11 green, rows 1–5, 7–9 and 11.
All nine native/Node comparisons match and all nine byte mutants are caught.
Row 8 retains its approved unknown signature; row 10's original process-as-any
probe still refuses, but adaptation 65 removes that site from the actual slice.

Ordered discovery (locations are in the mirrored slice). Only row 1 is observed
on the validated slice. Every later row is found behind cumulative scratch-only
stubs and has no native or Node identity claim:

| Row | Slice site | Exact refusal reason or NotYet detail | Minimal program |
| --- | --- | --- | --- |
| 1 | core.ts:525:17, reduceLeft | refuses arguments | native-arguments-length.a |
| 2 | core.ts:594:5, isArray(value: any) | type predicate return expression is not a trusted check on value | native-array-is-array-any.a |
| 3 | core.ts:616:97, cast callback parameter | type predicate: there is no body proving this parameter | native-predicate-callback-parameter.a |
| 4 | debug.ts:429:5, attachFlowNodeDebugInfoWorker | this in a namespace function; qualified and detached calls have different receivers | native-namespace-object-receiver.a |
| 5 | debug.ts:526:5, enableDebugInfo | same namespace receiver refusal | native-namespace-object-receiver.a |
| 6 | debug.ts:739:5, DebugTypeMapper | class inside a namespace; constructor registration/initialization not proven | native-namespace-class.a |
| 7 | debug.ts:50:5, namespace log | namespace merged with a function; callable properties/identity/receivers not represented | native-callable-namespace.a |
| 8 | debug.ts:80:42, getOwnKeys(assertionCache) | Partial<Record<AssertionKeys, ...>> seen as MapLike; fixed objects and records have different storage | native-partial-record-view.a |
| 9 | debug.ts:102:56, Debug[name] as AnyFunction | invariant-mutable: wider function view can write void where boolean is read | native-function-any-view.a |
| 10 | debug.ts:111:9, fail | refuses debugger | native-debugger-statement.a |
| 11 | debug.ts:161:131, assertIsDefined | predicate normal return has not narrowed value to NonNullable<T> | native-generic-assert-non-nullable.a |

The nine new reductions each run on Node with exit 0 and empty stderr, and each
reproduces its diagnostic family with C emission exit 1. Rows 4/5 share a
reduction; row 1 uses the existing arguments probe. Exact diagnostics, Node
outputs and stub ranges are retained in evidence/front24. The new signature
probe cuts the body to a loud throw while retaining the predicate callback type.
The nine original green probes are distinct from these nine red reductions.

Discovery replaced function bodies with loud throws. To cross the callback
signature refusal, its predicate result was changed to boolean in scratch only.
To cross the class refusal, DebugTypeMapper became an interface plus a typed,
checked-undefined prototype value. To cross the callable namespace refusal,
the unused top-level Debug.log function was renamed parserProofLogStub while its
namespace remained. Later rows depend on these altered source shapes, so they
are discovery candidates, not a claim about the eventual unmodified build order.
The raw diagnostics preserve their exact traversal order. Stopped at ten rows
behind arguments.length, as requested; no compiler proof bypass and no temporary
adaptation. Initial discovery attempts repeated an ineffective body-only callback
stub and then selected a type node incorrectly; those artificial outcomes are
retained in logs but excluded from this ordered list.

Native acceptance on the 81 files and 10,406 cases remains unreached. The extended
Node reference is still 36,429,231 bytes, SHA-256
`686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615`.
No parser-output byte mutant, C size, clang split/unsplit/warm time, parser binary
size, perf or native best-of-three user time can be measured without C/binary.
nproc is 5; previous Node user time is 6.162 seconds, not a fresh measurement.
No full gate, counts gate, complete oracle or WASI run is claimed.

[Full evidence](evidence/front24/report.json) includes both-mode attempts,
ordered stops, all eleven probes, minimal reductions and green namespace tests.
Compiler merge remains scratch-only; only parser-owned evidence and programs
are committed and pushed.

---

# Namespace reachability rerun: red before C, 9 of 11 probes green

Merged namespaces-tsc 023a1a34 into unpushed scratch 16d0caa7, yielding
a0aa1e69. Compiler rebuild succeeds. The old core.ts:11:52 Map-before-namespace
witness now passes natively against Node (`0:false\n`); its output-byte mutant
is caught. All eleven probes were rerun: rows 1–5, 7–9 and 11 pass (9/11),
each with equal stdout/stderr, zero exits and a caught native byte mutant.
Row 6 still refuses `arguments`; row 10 is the original unchecked cast probe,
already removed from the real slice by validated adaptation 65. Row 8 retains
the approved unknown signature. The validated slice was not changed.

The first observed full-build blocker is performance in namespace reachability,
not a new lowering diagnostic. The C attempt produced zero bytes and no diagnostic
for approximately six minutes on one busy CPU. Its retained SIGQUIT stack is in
`lower.namespaceInitialization.func1`, repeatedly following function calls.
Bounded C/build reruns, with split off and on, retain their stacks and outcomes
in evidence/front23. A timeout is not counted as a compiler refusal.

Minimal program: native-namespace-call-graph.a. Twenty-four function levels each
contain two calls to the previous function under `if (false)`. Node prints
`true\n` without taking any of those branches. C emission exceeds 30 seconds;
reducing each level to one call takes 0.064 seconds, with identical Node output.
Two-call depths 12, 16 and 20 take 0.066, 0.217 and 2.675 seconds respectively.
These timings were taken while the full build was also running. Observation:
both the full stack and the size-scaling witness implicate namespace traversal.
Inference from namespaces.go: its active-path map removes a node on return,
so shared DAG targets are revisited through each path. This supports exponential
work; it does not establish an infinite loop. No compiler implementation was
changed or readiness proof bypassed to get past it.

The new reachability/closed-call-graph tests pass in 0.181 seconds. Broader focused
namespace tests fail separately in 1.136 seconds: Parser and IncrementalParser
shape tests expected a bodyless-function NotYet but got nil; early_enum expected
an enum-before-initialization NotYet but got nil. Their expectations were not
changed. Compiler/lowering merge resolutions are listed in the report and kept
as a compressed scratch-only patch. The scratch merge is never pushed.

No parser C or binary means no native 81-file or 10,406-case comparison, clang
split/unsplit/warm timing, binary size or user-time measurement. nproc is 5;
the previous Node user time is 6.162 seconds. The extended Node reference remains
36,429,231 bytes, SHA-256
`686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615`.
No native parser-output mutant can run without that output. The nine probe byte
mutants and the cleared namespace-map witness mutant are all caught. No
arguments.length stub was made: the first observed blocker occurs before a
lowering diagnostic can identify that site. No new temporary adaptation.

Exact evidence: [front23 report](evidence/front23/report.json), bounded build
report/stacks, eleven probe comparisons and namespace call-graph scaling inputs.
Reproduction uses measure-builds.py with `--timeout 60 --jobs 5`,
probe-stops.py and namespace-call-graph.py against the scratch compiler.

---

# Host method descriptor rerun: process.nextTick truthiness green

Merged host-method-presence 51761b0 (including method-presence-test 722a1c5a)
into unpushed scratch 290f216b, yielding 16d0caa7. Compiler rebuild succeeds.
The unchanged native-process-next-tick-read.a now compiles. Native and Node both
print `true\n`, exit 0, empty stderr; stdout/stderr comparisons each exit 0.
A one-byte change to actual native stdout is caught by cmp exit 1. The probe
binary is 441,056 bytes. Node loader types still pass with process: Process and
pinned @types/node 25.3.3. The prior ambient-host runtime descriptor stop clears.

Focused method-presence lower tests pass in 0.229s. The uncached
TestNodeProcessMethodPresence oracle passes in 14.489s: its three fixtures compare
sanitized native and emitted JavaScript with source Node, verify SSA, and run
native/JavaScript descriptor-absent mutants. For each fixture, replacing the
native descriptor anchor with false or the JavaScript !!process.nextTick anchor
with false runs cleanly (exit 0, empty stderr) and is caught only by differing
stdout. Six descriptor mutants are caught, plus the requested output-byte mutant.
Exact commands and output are retained in evidence/front22/report.json and logs.
No full gate, counts gate, complete oracle or WASI run is claimed.

Conflicts span loader/SDK API compatibility, cast and expression dispatch,
refusal observation guards, native compiler selection and string indexing.
Current SDK-compatible loader roots (including require typing), deferred checked
views, phantom/reconciled interface views, records, namespace readiness and
existing Node dispatch were retained. Incoming host descriptor support was
integrated without bypassing the general ambient-host refusal for unsupported
methods. Current empty-allocation/literal string index and compiler selection
paths were kept. Exact ancestry/resolutions and scratch-only patch are retained;
compiler integration was never pushed. No probe source adaptation was made.

The full parser rerun remains held for the namespace fix or front-3. The last
observed first slice stop remains core.ts:11:52; no fresh all-eleven-probe count,
split rerun or native parser reference comparison is asserted. The extended
81-file and 10,406-case native acceptance targets remain unverified. Reproduce
this probe with evidence/front22/run-probe.py.

---

# Method truthiness probe: red at ambient-host runtime descriptor

Merged method-presence-test 722a1c5a into unpushed scratch c7e63804, yielding
290f216b. Compiler rebuild passes. native-process-next-tick-read.a is unchanged.
Node prints `true\n`, exit 0, empty stderr. Native build exits 1 at 6:52:
`stage 0 can't lower presence of an ambient host method without a runtime descriptor yet`.
The previous unbound-method refusal has become this more specific NotYet stop.

This is not lack of the Node loader. `adamic types` on the exact probe exits 0
and reports `variable process: Process`. The activated pinned @types/node is
25.3.3; its resolved package path is in evidence/front21/report.json. Code
inspection: methodPresence explicitly refuses a method whose receiver binding
is declared under an Ambient modifier without a runtime descriptor. Its
TestMethodPresenceHostIsNotYet tests this boundary independently of the loader.
The actual host method presence still lacks that runtime lowering path.

Focused TestMethodPresence lower tests pass in 0.241s, covering ordinary presence,
escapes and this host NotYet boundary. No native binary or output is emitted for
the requested probe, so the native byte comparison and native-output mutant
could not run. No substitute native output or green host claim is made.

The refusal-pass merge conflict preserves Buffer and existing Node/record
observation exceptions while adding incoming truthiness/comparison observations;
count conflicts union incoming rows. No compiler proof bypass or temporary
source adaptation was made. Exact ancestry, scratch-only patch and commands are
in evidence/front21, with run-probe.py for reproduction. Compiler merge remains
unpushed. No full gate, complete oracle, counts gate or sanitizer run occurred.

Full parser rerun remains held for the namespace fix. The last observed first
slice stop remains core.ts:11:52; no all-eleven-probe or split rerun was performed.
Both native parser references, JSDoc and error-recovery acceptance remain
unverified. This run only establishes the new stop for process.nextTick presence.

---

# Corrected host unknown integration: row 8 remains green

Merged host-blockers b788e96e into unpushed scratch e3846087. The merge is
02274b02; a test-only merge correction yields final scratch c7e63804. The corrected
host tip replaces a85a9cb1 as the selected implementation; older commits remain
ancestors. Compiler rebuild passes. The probe retains its sanctioned unknown
signature and its source is unchanged this turn.

native-array-is-array-predicate.a builds and prints `true\nfalse\n` natively
and on Node, exit 0 and empty stderr. stdout/stderr comparisons both exit 0.
The binary is 440,888 bytes. A one-byte change to actual native stdout is caught
(cmp exit 1). Commands, outputs and actual mutant are in evidence/front20/probe.

Initial focused testing could not compile its package: the merge had duplicated
err/refused declarations in two qualified Node tests. Scratch repair removed
only duplicate identical lowerSource calls/declarations, keeping both sets of
refusal assertions. Final focused array-predicate, unknown-reflection and
qualified-Node tests pass in 0.983s. Initial failure and final log are retained;
no expectation was weakened. The test-only correction did not alter compiler
code used for the already-built probe.

The merge replaces old Array.isArray-only unknown restrictions with incoming
general unknown/nonprimitive representation, while keeping record storage,
namespace readiness, checked cast dispatch, Node/detached-own paths and the
shared never-array adapter. Erased unknown-array element reads stay refused.
Exact ancestry, merge resolutions and scratch-only patch are in front20/report.json
and scratch-merge.patch.gz. No compiler merge was pushed.

The user identifies b788e96e as fixing the prior unknown-narrowing regressions in
library fixtures 05 and 13. Those fixtures were not rerun here; the local claim
covers row 8 and focused lower contracts, not their full library runtime behavior.
No full gate, complete oracle, counts gate or sanitizer run is claimed.

Full slice and both native references remain held for the namespace fix. The
last observed first slice stop is still core.ts:11:52, and no fresh cumulative
11-probe count is asserted. Only row 8 was rerun. No source adaptation or compiler
proof bypass was introduced. Native JSDoc/error-recovery acceptance remains
unverified. See evidence/front20/run-probe.py for reproduction.

---

# Nullable-array rerun: nine probes green, slice red at namespace initialization

Before merging, `git ls-remote --heads origin compiler/stage3-front-3` returned
no ref. Used the supplied views-arrays-callables-parser c898009b, merged into
unpushed scratch 975e4d0b, yielding e3846087. Compiler and current native metrics
helper rebuild successfully. The validated parser-temp65-slice was not edited.

Both split modes and repeat exit 1 before C emission at
src/compiler/core.ts:11:52, `new Map<never, never>()`, with NotYet:
`a call before all runtime namespaces are initialized; put namespaces before executable module code`.
This is now the first real stop, earlier than the array sites. Inspection:
namespaceInitialization counts every runtime namespace, then refuses all
new/call expressions while any are pending, even without a namespace read.
The minimal native-map-before-namespace.a reproduces NotYet at 4:45 with an
independent Map construction before a Debug namespace; Node prints `0:false\n`,
exit 0 and empty stderr. No compiler proof bypass or source adaptation was made.

The full current probe count is 9/11, rows 1, 2, 3, 4, 5, 7, 8, 9 and 11 green.
Each matches Node stdout/stderr/exit byte for byte. All nine native-output
one-byte mutants are caught (cmp exit 1). Row 8 has the sanctioned unknown
signature; the other ten sources are unchanged. The remaining probes are row 6,
arguments.length Refused, and row 10, the original process-as-any cast Refused.
Adaptation 65 removes row 10's cast in the actual slice; it does not clear
unrelated reads of process.nextTick.

Both additional array witnesses clear: native-same-map-return-cast.a prints
`2\nundefined\n`, and native-sorted-empty-conditional.a prints `0\n1\n`,
matching Node byte for byte, exit 0, empty stderr. Both additional native-output
byte mutants are caught. These prove the former core.ts:204 and 421 forms as
standalone programs; the full slice preflight stops before reaching those sites.

Focused shared-array-adapter, preflight refusal/consumer/comparer, predicate
interaction and marker-boundary lower tests pass in 0.544s. No full gate,
complete oracle, counts gate or sanitizer run was performed. The merge has two
lowering conflicts: retain equivalent never relation semantics in invariance.go,
and use incoming shared viewNeverArrayElement in object.go. Exact ancestry,
commands, outputs and scratch-only merge patch are in evidence/front19.

C bytes/lines, clang times, native parser binary size and timing remain
unavailable; no parser C was emitted. Split-repeat has an empty cache and is not
warm. nproc is 5. Neither native acceptance reference ran: the extended 81-file
Node reference and the 10,406-case manifest remain targets, not native passes.

Known outstanding integration issues remain documented: user-reported
host-blockers a85a9cb1 regressions in library unknown-narrowing fixtures 05 and
13 (not reproduced here), and earlier broad namespace expectation failures in
front16 (not rerun by these focused tests). No behind-stub discovery was added
at this namespace initialization stop. Wait for the combined front-3 SHA for
future integration. The scratch compiler merge was never pushed.

Reproduce with probe-stops.py and measure-builds.py, compiler
/workspace/scratch/parser-front19-adamic, metrics parser-front19-metrics,
scratch /tmp/parser-front7-scratch and validated parser-temp65-slice entry,
jobs 5. Additional witnesses use evidence/front19/run-array-witnesses.py.

---

# Reconciled census/predicate rerun: all six requested probes remain green

Reported upstream after this run: merged host-blockers a85a9cb1 regresses
unknown narrowing in library fixtures 05 and 13. This is user-supplied evidence,
not reproduced locally. The green direct-call isArray probe and six requested
probes do not cover those cases, so the regression remains outstanding. Wait
for the supplied compiler/stage3-front-3 SHA and use that combined integration
in place of individual compiler feature tips; full slice stays held meanwhile.

Tools are available. Merged census-small-families-3 24a2b287 into unpushed scratch
80aef697, yielding 975e4d0b. This is the reconciled implementation in place of
separate census-small-families/proven-predicates tips; older commits remain
ancestors of the scratch merge. Compiler rebuild succeeds.

Unchanged rows 1, 2, 3, 4, 5 and 9 all compile natively and match Node stdout,
stderr and exit byte for byte. Every exit is 0 and stderr is empty. Row 1 prints
`overload declarations loaded\n`, row 2 `1\n`, row 3 `0\n`, row 4
`overload declarations loaded\n`, row 5 `2\n`, and row 9
`predicate view loaded\n`. A first-byte mutation of each actual native stdout
is caught by cmp exit 1, six of six. Probe sources were not edited.

Focused census predicate interaction/boundary, predicate and checked-overload
result lower tests pass in 12.067s. The command includes
TestCensusPredicateMarkerKeepsProofBoundaries, which failed in the prior merge,
and the new interaction boundary/escape tests. Exact command and output are in
evidence/front18/report.json and tests.log.gz. No test expectation was changed.

The scratch merge conflicts affect cast/proof dispatch, conditions, expression
reads and refusal preflight. Incoming interfaceView proof handling was combined
with the existing deferred checked-view/phantom dispatch. Incoming predicate
argument checks were retained, as were record storage checks and the namespace,
checked module/union and void read paths. No compiler proof bypass was made.
Exact ancestry, resolutions and scratch-only patch are retained in front18.
Only parser evidence is on the pushed branch; compiler integration stays unpushed.

Full slice, both native parser references and the other five probes were not
rerun, as requested. core.ts:204:12 remains the last observed first slice stop,
waiting for lane 2. This run is six of six requested probes, not a fresh full
11-probe count. Declaration-admission probes retain their prior limited coverage.
No complete gate, counts gate, sanitizer or full oracle run is claimed.

See evidence/front18/probes/report.json and run-probes.py for actual commands,
outputs, binary sizes and mutants. Native JSDoc and error-recovery acceptance
remain unverified.

---

# Array.isArray predicate probe: adapted signature green

Merged host-blockers a85a9cb1 into unpushed scratch e54b4179, producing
80aef697. Compiler rebuild succeeds after scratch conflict resolution.
The requested native-array-is-array-predicate.a now uses `isArray(value: unknown)`
instead of `any`, exactly as sanctioned and consistent with adaptation 40.
The function statements and calls are unchanged; validated slice files were not
edited. This is a signature-adapted row-8 rerun, not an unchanged eleven-probe run.

Original source on Node, adapted source on Node and native all print
`true\nfalse\n`, exit 0 with empty stderr. stdout and stderr comparisons exit 0.
The native binary is 436,368 bytes. A one-byte change to its actual stdout is
caught (cmp exit 1). TypeScript 6.0.3's emitted JavaScript is byte-identical for
the original any and adapted unknown signatures. A source mutant negating the
real Array.isArray return changes the emitted JavaScript and that guard catches
it (cmp exit 1). Both erased outputs and both mutants are retained.

Focused array-predicate element-contract, narrowing, tuple-refusal and unknown
observation lower tests pass in 0.356s. No full gate, complete oracle, counts
gate or sanitizers ran. Probe runtime coverage is direct calls with array and
object inputs; no complete parser or general narrowing-runtime claim is made.

The merge also imports truthiness, labels and void paths. Conflicts were resolved
while preserving current checked-cast preflight, namespace and import-cycle
readiness, detached own-property handling, open numeric/flag enums and Node/record
bindings. Existing generic predicate-flow naming and instantiated types were
retained while adding the trusted Array.isArray narrowing. Initial build errors
for the stale enumElement helper and duplicated export case were resolved using
the current enum dispatch and incoming named-export handler. Exact ancestry,
conflict resolutions and scratch merge patch are in evidence/front17/report.json
and scratch-merge.patch.gz. Compiler changes remain unpushed; no proof bypass
or temporary source adaptation was added.

Full slice and all-eleven-probe reruns remain held for lane 2. core.ts:204:12 is
still the last observed first real slice stop. This run establishes row 8 only,
with the approved signature; no fresh cumulative probe count is claimed.
Native extended 81-file and 10,406-case acceptance remain unverified.

See evidence/front17/probe/report.json and run-probe.py for commands, exact
outputs and reproduction (the original Node source is in scratch at the path
recorded by that report). The emitted original JavaScript is retained in evidence.

---

# Detached own-property and Debug namespace probes: both runtime comparisons green

Merged records-lowering 198ff780 and namespaces-tsc f893faf2 into unpushed
scratch 42b2ee13, yielding e54b4179. Compiler rebuild succeeds after scratch
merge resolution; all compiler edits remain unpushed. Both probe sources are
unchanged, with no temporary adaptation or proof bypass.

native-has-own-property.a compiles and prints `true\nfalse\n` natively and on
Node. native-debug-namespace.a compiles and prints `false\n` in both. Both exit
0, with empty stderr; stdout and stderr comparisons each exit 0. A one-byte
change to each actual native output is caught by cmp exit 1. Probe binaries
are 436,672 and 436,328 bytes respectively. Commands, outputs and actual mutants
are in evidence/front16/probes/, with run-probes.py for reproduction.

## Integration tests reported separately

The broader focused lower command exits 1 in 1.177s, with three expectation
failures. IncrementalParser in TestTscNamespaceDeclarationShapes returns nil
instead of the expected NotYet function without a body. The computed_enum and
early_enum cases in TestNamespaceLimitsStayLoud get namespace initialization
NotYet rather than the expected enum diagnostic. These observations do not
establish why declaration admission changed; no expectation or compiler proof
was weakened to turn the tests green. Full output is retained in tests.log.gz.

The selected detached-own, namespace type erasure, receiver-refusal,
returned-assignment and parser-factory hoisting tests pass in 0.713s. The full
gate, counts gate, sanitizers and complete oracle were not run.

Merge conflicts touch lowering, namespace state/module registration, enum
handling and native/JavaScript readiness emission. Existing checked-view cast
preflight, import-cycle checks, open numeric/flag enums, lazy non-null defaults,
records and Node bindings were retained while integrating namespace dispatch.
The localRead helper carries the existing checked union read and its possible
NotYet error through all namespace callers. The initial helper-signature compile
failure and its correction are recorded. Exact ancestry/resolutions and the
scratch-only merge patch are in evidence/front16/report.json and
scratch-merge.patch.gz; no compiler changes are on the proof branch.

The full parser rerun remains held as requested. core.ts:204:12 remains the
last observed first real slice stop, with no new slice check claimed. The prior
full eleven-probe run was 6/11; rows 7 and 11 are now separately green, but a
fresh 8/11 run was not performed. Both native parser references, JSDoc and error
recovery remain unverified. No fresh C/clang/timing metric is claimed.

---

# Deferred cast preflight rerun: six probes green, slice red at sameMap return

Merged views-arrays-callables-parser 4ed5e301 into the unpushed scratch, producing
42b2ee13. It includes proven-predicates 3d48edb, census-small-families bde0030,
library-array-holes f05aec3 and phantom-brands a73f93c5. Compiler rebuild succeeds.
The validated slice still contains only sanctioned 60-65 adaptations.

Rows 2 and 5 clear: native prefix cast prints 1, branded-array cast prints 2,
matching Node byte for byte. Unchanged probe count is now 6/11, rows 1, 2, 3, 4,
5 and 9. All six native-output one-byte mutants are caught (cmp exit 1). Row 10's
original cast probe remains unchanged/refused; adaptation 65 removes that cast
in the actual slice. Probe greens retain their original coverage: some only
prove declaration admission, as their source comments say.

Unsplit, split and repeat builds all exit 1 before C emission at
src/compiler/core.ts:204:12, sameMap's final `array as unknown[] as U[]`.
This is another array-view case: the operand is `readonly T[] | undefined`,
unlike the now-green changed-prefix slice cast. The minimal
native-same-map-return-cast.a reproduces the refusal at 4:12; Node prints
`2\nundefined\n`. No unchecked-cast proof was bypassed.

## Ordered discovery behind scratch-only source stubs

Only row 1 is the real validated-slice failure. Every later row was found behind
cumulative stubs in a separate copied tree. Function bodies were replaced by
throwing bodies; the unbound method initializer became a typed uninitialized
slot. These are discovery stubs, not adaptations, and were never run as proofs.

| Order | Slice location | Stop | Minimal witness |
| --- | --- | --- | --- |
| 1 | core.ts:204:12 | sameMap final nullable readonly array cast refused | native-same-map-return-cast.a |
| 2 | core.ts:421:34 | toSorted conditional emptyArray: never viewed as writable T | native-sorted-empty-conditional.a |
| 3 | core.ts:525:17 | reduceLeft arguments.length refused (original row 6) | native-arguments-length.a |
| 4 | core.ts:542:24 | Object.prototype.hasOwnProperty method read refused (original row 7) | native-has-own-property.a |
| 5 | core.ts:594:5 | Array.isArray predicate return not trusted (original row 8) | native-array-is-array-predicate.a |
| 6 | core.ts:889:14 | process.nextTick truthiness method read refused | native-process-next-tick-read.a |
| 7 | debug.ts:26:1 | Debug namespace refused (original row 11) | native-debug-namespace.a |

Row 2's new minimal reproduces adamic/invariant-mutable and Node prints
`0\n1\n`. Row 6's new minimal activates the pinned Node loader with an otherwise
unused Node import, reproduces unbound-method and Node prints `true\n`.
Arguments.length stays exactly as ruled, with no adaptation. Stubbing it reaches
the hasOwnProperty read; stubbing that read reaches the Array.isArray predicate.
Discovery ends at the namespace, which has no function-body/initializer stub;
no namespace or compiler proof bypass was introduced to reach more stops.
All diagnostics, exact cumulative replacements and discovery patch are retained.

## Validation and merge limits

Focused deferred-preflight, phantom-array, predicate and checked-overload lower
tests pass in 13.283s. Full gate, counts gate and complete oracle were not run.
The scratch merge conflicts touch lowering, native emission, runtime, flow and
loader; their resolutions and exact ancestry are in evidence/front15/report.json
and scratch-merge.patch.gz. Source stubs live only in the discovery scratch copy.

C size, clang times, native parser binary size/timing and both native reference
comparisons remain unavailable because lowering stops before C. nproc is 5.
The repeat cache is empty and is not a warm measurement. The 81-file extended
reference and 10,406-case manifest remain Node-validated acceptance targets;
native JSDoc and error-recovery acceptance remain unverified.

Reproduce the validated run with probe-stops.py and measure-builds.py, compiler
/workspace/scratch/parser-front15-adamic, scratch /tmp/parser-front7-scratch,
and entry /workspace/scratch/parser-temp65-slice/parser-proof-main.a (jobs 5).
See evidence/front15/probes/report.json, build-modes/report.json,
new-probes/report.json and discovery/report.json for exact outputs.

---

# Row 5 phantom-array brand rerun: red before native emission

Merged phantom-brands a73f93c5 into scratch 444d57b1, producing unpushed
4bda8003. Compiler rebuild succeeds. The unchanged minimal
native-sorted-array-brand.a still refuses at 7:12 with adamic/no-unchecked-cast.
Node prints `2\n`, exit 0, empty stderr. No native binary/output was emitted,
so the native comparison and native-output byte mutant could not run.

Focused `go test ./internal/lower -run TestPhantomArray -count=1` fails (1.171s).
Required and optional brand cast tests hit the same legacy cast refusal; other
proof-boundary tests receive that refusal instead of their expected diagnostics.
See evidence/front14/tests.log for every failure. No focused test pass is claimed.

Code inspection: refusals.go invokes castProof before expression lowering;
castProof calls phantomCast, but the new phantomArrayCast dispatch is in cast.go.
The merged preflight still rejects the probe before reaching that dispatch.
The preflight was not bypassed. Lowering merge conflicts and exact ancestry are
recorded in evidence/front14/report.json and scratch-merge.diff. Compiler changes
remain scratch-only. No temporary source adaptation was added.

Full-slice builds and both native references are held until the row-2 castProof
ordering fix arrives, as requested. Row 2 remains the last observed first stop;
this probe-only run does not establish a new slice stop or rerun the eleven-probe
count. Native JSDoc and error-recovery acceptance remain unverified.

---

# Array views and generic empty fallback rerun: red at the legacy cast preflight

Both requested fixes are in unpushed scratch 444d57b1: views-arrays-callables
0b141c26 (including lane 1 609ed395), and parser-generic-empty-array 4e022aae.
Stage0 compiler rebuild succeeds. The validated slice still has only sanctioned
60-65 adaptations; no discovery source stubs or compiler proof bypasses were used.

Final first real failure remains core.ts:195:37, sameMap's
`array.slice(0, i) as unknown[] as U[]`, adamic/no-unchecked-cast. Unsplit,
split and repeat builds all exit 1 before C emission. This is still the array
family, not a tuple or a later array consumer. No such later stop was reached.

Observed: the final diagnostic has the older castRepair message; the unchanged
row-2 minimal still refuses at 4:12 and Node prints 1, exit 0. Code inspection:
refusals.go calls the pre-existing castProof before expression lowering, and
castProof rejects this array assertion instead of dispatching checked-array
views. Incoming cast.go does contain checked-view dispatch after conflict
resolution. Thus the combined compiler still has a preflight/checked-view
integration interaction. The lane branch's separately reported generic-probe
success is not claimed as reproduced here. The old preflight was not bypassed.

Final unchanged probe count: 4/11, rows 1, 3, 4 and 9 compile and match Node
exit/stdout/stderr byte for byte. All four native-output one-byte mutants are
caught (cmp 1). With only the view branch added the count was 3/11; the empty
fallback branch clears row 3 exactly. Row 10's original cast probe remains
unchanged and refused; adaptation 65 removes its source site in the actual slice.

C size, clang times, parser binary size, native timing and both native references
remain unavailable. Repeat cache has zero objects and is not warm. nproc 5.

## Integration validation and limitations

Focused lower empty-array, array/callable-contract and predicate tests pass
(5.076s). Before the empty-array merge, the checked-array oracle command could
not compile its test package: three node_require_test.go calls pass inputRun
where inputLeaks now requires func() inputRun. No fixture execution or checked-
array oracle pass is claimed. No compiler/test repair was made for that API
interaction, and no full gate ran. Initial focused lower contracts/predicates
before the empty merge passed in 10.709s.

The views merge touched lowering: keep incoming checked-view cast dispatch and
old literal/nominal fallback, concrete generic types, Node require projection
and qualified-name guard. Native/JavaScript conflicts retain array-hole paths,
counted closure ABI and current readiness bookkeeping while incorporating view
reads. Joint holes+views runtime semantics were not verified by this stopped
parser run; only the four actual green minimal programs were executed.
All scratch conflict resolutions and exact requested ancestry are recorded in
evidence/front13/inputs.json and compressed patches/logs. They were never pushed
as compiler commits. No second array-view implementation or temporary source
adaptation was made.

Reproduce with measure-builds.py and probe-stops.py using the compiler/slice paths
in evidence/front13/front13-build.json and inputs.json. The existing unchanged
native-generic-array-cast.a is the minimal first-stop witness. Both the view-only
and final two-feature observations are retained. Native JSDoc and error-recovery
acceptance remain unverified.

---

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

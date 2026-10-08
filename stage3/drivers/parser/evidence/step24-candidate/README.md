# Step 24 candidate parser measurement

Evidence branch starts at main 45487a80. Compiler integration happened only in
a never-pushed detached scratch. Candidate 0c99fc54 had not landed on main.
records-maplike-next 8c013f1c merged cleanly, producing scratch 9d09100b.
area-stack c6ae12df conflicted in exactly two paths and was aborted/skipped:
internal/oracle/counts.md and stage3/fixtures/nested-functions/09_checker_constituent_recursion.a.
No conflict was resolved and no scratch/compiler/source adaptation is pushed.

The scratch uses cohere SDK 7945d102. Reusing its existing checkout through a
scratch symlink makes Go VCS stamping fail with exit 128. The retry uses
-buildvcs=false and succeeds; summary.json pins the compiler SHA and binary hash.
The symlink is environmental setup, not a compiler source edit. It also makes
ordinary git status on that scratch complain about the submodule symlink.

Parser source is a fresh full apply using the exact ec71eb48 parser-proof
workflow. 65 and 76 are applied and their checks/mutants pass; sameMap stays
unchanged. The seven-root slice has 27 declaration files, 1,993 code declarations,
41,857 copied lines, 2,082 spans and 79 ordered modules. The byte audit passes.
The fixed 81-file input corpus is freshly reconstructed by scout a661f9f3;
generation uses the relative diagnosticMessages argument. Its reference and
one-byte identifier/input-hash mutant pass. No corpus dump contents are committed.

The slice's complete extended Node dump matches the fixed reference:
36,429,231 bytes, SHA256 686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615,
exit zero, empty stderr. The Node-end and two JSDoc mutants are caught.
This is Node source evidence, not a native parser success.

## Unmodified native attempts

ADAMIC_NATIVE_SPLIT=0 and =1 both exit 1, at debug.ts:113:19 and 114:19:
TS2339, captureStackTrace missing on ErrorConstructor. Their stderr is identical.
Wall times are 0.5659s and 0.5178s, respectively. C/clang, native parser binary,
native parser dump and recovery acceptance are unreached; these are not clang
times. The pinned scout comparator was built, ready to run first on any successful
native output. Since neither parser builds, it ran only on green focused probes
and their one-byte native-output mutants.

## Ordered stops and earlier BLOCKERS.md

The earlier list is pinned verbatim from ec71eb48 in previous-BLOCKERS.md.
Only row 1 below is reached by the unmodified builds. Rows 2 onward are observed
in a separate discovery copy behind recorded throwing function/initializer
placeholders. They are not applied adaptations and were never used for Node or
acceptance. The final row's namespace stop is on the discovery initializer
helper at the original emptyMap location; its independent unchanged Map-before-
namespace witness reproduces that family. All nine families/locations were
present in the earlier ordered list: status **still there**.

| Order | Slice location | Stop | Minimal Node-held program | Owner | Prior status |
| --- | --- | --- | --- | --- | --- |
| 1 | src/compiler/debug.ts:113:19 | error TS2339: Property 'captureStackTrace' does not exist on type 'ErrorConstructor'. | [native-error-capture-stack.a](../../native-error-capture-stack.a) | area/library: ErrorConstructor host contract | still there |
| 2 | src/compiler/core.ts:11:52 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](../../native-enum-map.a) | compiler module initialization (historical codex/enum-init-reach) | still there |
| 3 | src/compiler/debug.ts:333:29 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](../../native-enum-map.a) | compiler module initialization (historical codex/enum-init-reach) | still there |
| 4 | src/compiler/performanceCore.ts:37:37 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-call-before-enum.a](../../native-call-before-enum.a) | compiler module initialization (historical codex/enum-init-reach) | still there |
| 5 | src/compiler/performance.ts:15:18 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-call-before-enum.a](../../native-call-before-enum.a) | compiler module initialization (historical codex/enum-init-reach) | still there |
| 6 | src/compiler/performance.ts:16:15 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](../../native-enum-map.a) | compiler module initialization (historical codex/enum-init-reach) | still there |
| 7 | src/compiler/performance.ts:17:16 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](../../native-enum-map.a) | compiler module initialization (historical codex/enum-init-reach) | still there |
| 8 | src/compiler/performance.ts:18:19 | stage 0 can't lower an indirect call or class construction before enum initialization; declare enums before executable module code yet | [native-enum-map.a](../../native-enum-map.a) | compiler module initialization (historical codex/enum-init-reach) | still there |
| 9 | src/compiler/core.ts:11:52 | stage 0 can't lower a call before all runtime namespaces are initialized; put namespaces before executable module code yet | [native-map-before-namespace.a](../../native-map-before-namespace.a) | compiler namespace initialization (compiler/area-stack skipped) | still there |

Discovery stops at nine, rather than inventing fifteen. Replacing tryGetPerformance
by throw initially changes its inferred return to void; that checker artifact is
excluded. The continuation explicitly preserves the original stock-inferred
`{ shouldWriteNativeEvents: boolean; performance: Performance } | undefined`
return type, recorded in original-return-type.log. Trying to bypass the namespace
stop repeats the emptyMap helper and causes duplicate implementation diagnostics;
that attempt is excluded too. Exact attempts, placeholders and the discovery-only
patch are retained. The edited copy does not claim original span hashes.

## Focused earlier-family comparison

All 21 existing minimal programs run on source Node with empty stderr. Six build
natively and match stdout/stderr/exit. Each of these six has a caught one-byte
native-output mutant using a661f9f3's comparator. No accepted output differs.
MapLike indexing is **gone**. Three formerly green namespace declaration probes
are **new** refusals on this retained candidate; the requested area-stack merge
which formerly admitted them was skipped. Those earlier successes were declaration
admission only, not executed namespace receiver/constructor proofs. Remaining
refusals are still there; five non-namespace green probes remain green.

| Existing minimal | Compared with 2c7d9fd2 | Current native equality |
| --- | --- | --- |
| [native-error-capture-stack.a](../../native-error-capture-stack.a) | still there | build blocked |
| [native-namespace-object-receiver.a](../../native-namespace-object-receiver.a) | new | build blocked |
| [native-namespace-class.a](../../native-namespace-class.a) | new | build blocked |
| [native-callable-namespace.a](../../native-callable-namespace.a) | new | build blocked |
| [native-enum-map.a](../../native-enum-map.a) | still there | build blocked |
| [native-call-before-enum.a](../../native-call-before-enum.a) | still there | build blocked |
| [native-maplike-index.a](../../native-maplike-index.a) | gone | yes, byte mutant caught |
| [native-nonnull.a](../../native-nonnull.a) | still there | build blocked |
| [native-predicate-overload.a](../../native-predicate-overload.a) | still green | yes, byte mutant caught |
| [native-predicate-callback.a](../../native-predicate-callback.a) | still green | yes, byte mutant caught |
| [native-same-map-return-cast.a](../../native-same-map-return-cast.a) | still there | build blocked |
| [native-key-array-cast.a](../../native-key-array-cast.a) | still there | build blocked |
| [native-arguments-length.a](../../native-arguments-length.a) | still green | yes, byte mutant caught |
| [native-arguments-length-value.a](../../native-arguments-length-value.a) | still green | yes, byte mutant caught |
| [native-generic-array-cast.a](../../native-generic-array-cast.a) | still there | build blocked |
| [native-generic-empty-array.a](../../native-generic-empty-array.a) | still there | build blocked |
| [native-sorted-array-brand.a](../../native-sorted-array-brand.a) | still there | build blocked |
| [native-predicate-callback-parameter.a](../../native-predicate-callback-parameter.a) | still there | build blocked |
| [native-some-predicate-overload.a](../../native-some-predicate-overload.a) | still green | yes, byte mutant caught |
| [native-debug-namespace.a](../../native-debug-namespace.a) | still there | build blocked |
| [native-map-before-namespace.a](../../native-map-before-namespace.a) | still there | build blocked |

The namespace regressions belong to compiler/area-stack (skipped); the cleared
MapLike witness belongs to codex/records-maplike-next. Other owner routing is the
same earlier feature routing, not a claim that those branches are merged here.
No new authored .a program was added: this unit measures existing driver witnesses.
counts.md records the fresh 21/6 observations; no central oracle row changes.

## Commands, checks and limitations

Setup timings: Go 0.025s, Node 0.026s, markdown 0.086s, submodules 0.092s,
clang 0.221s, Go build 12.387s, deferred test binaries 12.523s, warm cache
12.525s, done 12.554s; nproc 5, quota 4. Versions: Go 1.27.1, clang 20.1.8,
Node 24.19.0. All commands wrote logs, never test pipes.

Commands: named-ref fetch and detached worktrees; ordered merges and abort;
Go build -buildvcs=false ./cmd/adamic; ec71eb48 stage3/apply.sh; 65/check.cjs
and 76/check.cjs; seven-root slice/run.sh and slice/verify.cjs; scout
reconstruct-corpus.py; parser/run.sh with --inputs fixed corpus;
step24-measure.py, step24-discover.py and step24-continue.py; step24-probes.py.
The archived scripts preserve exact arguments and scratch paths. Their paths
must be supplied/replaced when reproducing in another environment. Compiler
feature source changes are absent from the delivery branch.

verify.py checks packaged identities, merge skips, fixed reference, build
outcomes, ordered coverage and every green probe's byte equality/mutant.
--drop-stop and --wrong-hash each must exit 1 at its own assertion. No full Go
package, upstream landing lane or full gate was run. No native parser acceptance,
performance, sanitizer or recovery corpus result is claimed. The earlier source
adaptation lane results remain historical; this unit changes no adapter.

Built: certified original referenced-map tuple consumers; total 3 pairs / 3 candidate reads.
Commits: prior 49647618880c8953cbb0328044247a0d28f53345; referenced-map commit recorded in git history.
Checks: referenced-map oracle PASS 2.630s; numeric controls PASS 0.216s/0.421s; tuple counts refreshed PASS 25.401s.
Mutants: drop original FileIdListId position check fails the named refusal oracle, printing 7 then false and exiting 0.
Remaining: 7 candidate pairs / 12 reads plus optional/rest Map forms; integrator failures below do not block this worker.

Referenced-map certification imports IncrementalBuildInfoReferencedMap unchanged,
including both required-any numeric brands. No number or void-brand alias is
substituted. The imported numeric carrier hook proves the brand member name
absent from the primitive using the existing inventory, retains number-kind
checks, and leaves demanded any-valued reads refused. Original source hashes
are checked by the existing 78-file manifest. Five fixtures cover valid IDs,
wrong key/value types, tuple length and numeric-keyed records. Node, JavaScript,
native release and sanitized native pass; the successful native fixture passes
leak checks. The numeric position mutant is caught first in JavaScript.

Commands, with output in logs:
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
'^TestCheckedViewTupleOriginalReferencedMap$' -count=1 -v -timeout 3m
(/tmp/views-tuples-references.log).
ADAMIC_TUPLE_NUMERIC_MUTANT=1 on the same oracle with
'^TestCheckedViewTupleOriginalReferencedMap/references-value-wrong$'
(/tmp/views-tuples-references-mutant.log), expected exit 1.
go test ./internal/lower -run
'^(TestPhantomRefusals|TestPhantomPrimitiveNames|TestPhantomLiteralCastsStayRefused)$'
-count=1 -v -timeout 3m (/tmp/views-tuples-numeric-regression.log).
go test ./internal/oracle -run
'^TestCheckedViewTupleOriginalNumericBrandReadRefuses$' -count=1 -v -timeout 3m
(/tmp/views-tuples-numeric-any-read.log).
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$' -count=1
-timeout 3m -args -update-counts (/tmp/views-tuples-references-counts.log).

The previous checkpoint and the named integrator failures follow.

---

Built: original TrackedSymbol optional forEach consumers through the existing tuple certificate; 2 pairs / 2 candidate reads now certified.
Commits: starting f753567dd8837e90914e7c229df88d8536a3bf8b; pair 6b58eb3c02d57ea341985643ae9528b5597115ac.
Checks: outSignature PASS 26.108s; TrackedSymbol PASS 12.566s; tuple counts update/check PASS 32.037s/37.821s; complete counts refresh blocked.
Mutants: original skip/shape/nested and tracked meaning/guard all fail their semantic oracles; JavaScript catches them before native emission.
Remaining: 8 candidate pairs / 13 candidate reads; optional/rest Map forms remain refused; exact reaching-view coverage is unmeasured.

Eight tracked-*.a fixtures import unchanged TrackedSymbol from the 78 hashed
upstream declarations at 050880ce59e30b356b686bd3144efe24f875ebc8. The oracle
also asserts the full original Symbol field set. Node decides fixture behavior;
JavaScript, native release and ASan/UBSan match the valid programs and pinned
refusals. Successful native runs pass the existing leak checker. Controls cover
nested Symbol.flags, tuple meaning, tuple arity, tuple versus record identity,
undefined receivers, receiver evaluation once and conditional callback creation.

The minimal statement hook saves receiver?.forEach(callback)'s receiver once
and guards the existing ArrayVisit; callback creation remains in that branch.
It handles arrays of required-position tuples only, without null receivers,
optional calls on the method itself, or thisArg. Other optional calls retain
refusals. No second tuple representation, constructor or flow graph is added.
The hook and original counts binding are listed in docs/checked-views-plan.md.
Twenty original tuple fixture count rows are now recorded, including the twelve
previous outSignature fixtures. Original counts remeasurement remains opt-in
with the same external declaration inputs as the existing semantic oracle.

Integrator failures, not a worker blocker: the required TestCountsAreRecorded refresh fails on
fixtures outside this unit: ctor_set.a aborts with free(): invalid pointer;
census_overload_contracts.a refuses excess implementation arguments;
census_small_boolean.a and nbody_field_values.a refuse template interpolation;
maybe_number_slots.a has a native argument representation error; and
require_perf_hooks.a panics in Node.Text on a binding pattern. The three named
lowering failures reproduce with statements.go restored from the exact starting
commit using Go's source overlay, which restores the entire production source
of this checkpoint. The broader lower refusal regression also has three stale
expectations for already-supported union operations; all three reproduce with
the starting-source overlay. The relevant optional-call and tuple-storage
regressions pass (0.136s and 0.319s). No prohibited file was edited, no complete
package or full gate was run, and no successful complete counts refresh is claimed.
The tuple-only counts update and check both pass. The user confirms these
unrelated failures belong to the integrator; work continues through the queue.

Remaining production candidates: four original numeric-brand pairs / seven
reads (97898, 97913, 97926, 97931), and four optional-position receiver pairs /
six reads (95604 fields 0/1; 68230 fields 0/1). A scratch unchanged
IncrementalBuildInfoEmitSignature probe still refuses its representation at
stage 0. No required-any numeric brand was substituted or erased. Optional/rest
Map gap controls pass as named compile refusals, not implementation completion;
no rest production pair has been found in the supplied inventory.

Commands actually run (all test output redirected to separate log files):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/views-tuples-setup-retry.log 2>&1
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/tuples/prepare.cjs /tmp/views-tuples-pinned /tmp/views-tuples-original-declarations > /tmp/views-tuples-prepare.log 2>&1
export ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewTupleOriginalOutSignature$' -count=1 -v -timeout 3m > /tmp/views-tuples-original-baseline.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewTupleOriginalTrackedSymbols$' -count=1 -v -timeout 3m > /tmp/views-tuples-tracked-final.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewMapStorageGaps$/^(optional-tuple|rest-tuple)$' -count=1 -v -timeout 3m > /tmp/views-tuples-map-gaps.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 15m -args -update-counts > /tmp/views-tuples-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$' -count=1 -timeout 3m -args -update-counts > /tmp/views-tuples-owned-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewTupleOriginalCounts$' -count=1 -timeout 3m > /tmp/views-tuples-owned-counts-check.log 2>&1
go test ./internal/lower -run '^(TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestATupleSeenAsAnArrayIsNotYet)$' -count=1 -v -timeout 3m > /tmp/views-tuples-lower-regression.log 2>&1
go test ./internal/lower -run '^TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat$/^a_(method|function_value)_called_through' -count=1 -v -timeout 3m > /tmp/views-tuples-optional-call-regression.log 2>&1
go test ./internal/lower -run '^TestATupleSeenAsAnArrayIsNotYet$' -count=1 -v -timeout 3m > /tmp/views-tuples-storage-regression.log 2>&1
```

Every mutant is a separate go test ./internal/oracle -count=1 -v -timeout 3m:

- ADAMIC_TUPLE_ORIGINAL_MUTANT=skip with
  ^TestCheckedViewTupleOriginalOutSignature/emit-wrong-kind-noread$:
  exits 0 and prints boolean instead of the expected exit-70 named refusal.
- ADAMIC_TUPLE_ORIGINAL_MUTANT=shape with
  ^TestCheckedViewTupleOriginalOutSignature/emit-record-noread$:
  accepts a record and prints object instead of the named tuple refusal.
- ADAMIC_TUPLE_ORIGINAL_MUTANT=nested with
  ^TestCheckedViewTupleOriginalOutSignature/emit-helper-wrong$:
  prints false instead of the named nested-position refusal.
- ADAMIC_TUPLE_TRACKED_MUTANT=meaning with
  ^TestCheckedViewTupleOriginalTrackedSymbols/tracked-meaning-wrong$:
  prints 1 then false and exits 0 instead of the pinned SymbolFlags refusal.
- ADAMIC_TUPLE_TRACKED_MUTANT=guard with
  ^TestCheckedViewTupleOriginalTrackedSymbols/tracked-evaluation-undefined$:
  improperly creates the callback and exits 70 with a TypeError; Node and the
  unchanged program print only receiver and exit 0. The callback stdout and
  exit-code oracle catch the missing guard; no clang or sanitizer failure is
  credited as a mutant kill.

Logs: /tmp/views-tuples-mutant-{skip,shape,nested}.log and
/tmp/views-tuples-tracked-mutant-{meaning,guard}.log. Starting-source checks:
/tmp/views-tuples-start-counts-failures.log and
/tmp/views-tuples-start-lower-failures.log, using
/tmp/views-tuples-start-overlay.json.

Setup finished successfully after working around the submodule fetch. Its
initial automatic TypeScript checkout fetched full history and was interrupted
(exit 143). A racing manual shallow fetch failed with 'shallow file has changed
since we read it'; sequential exact shallow fetch and checkout of
cohere/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f then succeeded.
Retry timing lines: Node ready 0.033s; Go 0.043s; submodules 0.088s;
markdown-width skipped, ready 0.090s (step 0.007s); clang 0.258s; Go build
795.430s; test binaries deferred 795.533s; cache warm 795.535s; done 795.618s.
nproc=5; cgroup quota=4; Go 1.27.1, clang 20.1.8, Node 24.19.0. The long cold
build also included redundant initial compilations, which were stopped before
sequential scoped reruns. No timeout cutoff was used for this checkpoint.

Previous checkpoint report follows unchanged.

---

Built: original EmitSignature field views using lane 1's single tuple certificate path and lane 2 array read hooks.
Commits: territory b3bc7138; lane 1 merge 13486c7c (includes 65d8a138).
Checks: original pair oracle PASS 11.507s; scoped IR/lower/native/JavaScript PASS 0.011s/4.593s/39.285s/0.875s.
Mutants: skip root, accept record, drop helper position each fail the pinned refusal oracle; none is killed by clang.
Remaining: 9 candidate pairs / 14 candidate reads after this push; optional/rest Map forms are additional fixture obligations.

The combined candidate queue is 10 pairs / 15 reads. Lane 2 supplies 6/9;
lane 1's nullish table supplies four optional-position receiver pairs / six
reads. No rest receiver is identifiable in that table. The two optional/rest
Map gap fixtures are tracked separately from production counts. Exact reaching
view coverage is not measured. This push covers pair 97934, outSignature, one
candidate read, against unreduced declarations from upstream commit
050880ce59e30b356b686bd3144efe24f875ebc8.

Preparation reuses lane 4b's original declaration emitter, verifies all nine
lane 2 read sites, and hashes all 78 generated declaration files. Tests require
the original IncrementalBundleEmitBuildInfo field set; no reduced interface is
substituted. All twelve fixtures run their unchanged source on Node, generated
JavaScript, native release and native ASan/UBSan. Successful native runs also
pass the shared leak checker. Undefined is admitted; the required field being
missing refuses. Wrong primitive, tuple length, record identity and nested
position diagnostics are pinned literally in the oracle.

Lane 1's constructor is generalized only to accept the existing recursive child
builders. Map producers still require complete descriptors and reject callable
tuple descendants. The array adapter retains lazy descendants. Empty tuples
have an explicit layout marker; Map schemas distinguish them from ordinary
records. Native producers carry tuple identity independently of casts, and
ordinary object copies do not acquire it. These hooks are listed in the plan.

Reproduction (source /workspace/adamic-tools/env.sh first):

```sh
node stage3/interface-downcasts/tuples/prepare.cjs /tmp/untagged-typescript /tmp/views-tuples-original-declarations > /tmp/views-tuples-prepare.log 2>&1
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/views-tuples-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewTupleOriginalOutSignature$' -count=1 -v > /tmp/views-tuples-pair1-final.log 2>&1
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'View|Map|Contract|Tuple' -count=1 -timeout 10m > /tmp/views-tuples-packages-handoff.log 2>&1
```

With the same original-declaration environment, each mutant is a separate
`go test ./internal/oracle -count=1 -v` run. `ADAMIC_TUPLE_ORIGINAL_MUTANT=skip`
on `^TestCheckedViewTupleOriginalOutSignature/emit-wrong-kind-noread$` prints
boolean and exits 0 instead of the named exit-70 refusal. `shape` on
`.../emit-record-noread$` admits an ordinary record and exits 0. `nested` on
`.../emit-helper-wrong$` returns the wrong position value and exits 0. Logs:
/tmp/views-tuples-mutant-skip-noread.log,
/tmp/views-tuples-mutant-shape-handoff.log,
/tmp/views-tuples-mutant-nested-handoff.log. IR mutations do not alter production
sources. The earlier skip probe was masked by secondary narrowing and was
replaced by the consumer that does not narrow.

Setup: GOPROXY=https://proxy.golang.org|direct; submodules 0.077s; markdown-width
skipped (ready 0.090s); clang ready 0.182s; Go build 37.033s; deferred test
binaries 37.214s; cache ready 37.216s; done 37.247s. nproc=5, cgroup quota=4.
Go 1.27.1, clang 20.1.8, Node 24.19.0. No complete repository gate is claimed.

Target for the optional/rest adapters is October 8. A completion date for the
whole original-declaration family is contingent on certifying original branded
numeric IDs: the highest-ranked tuple uses IncrementalBuildInfoFileId, whose
original required brand is any. The unchanged original read currently reports
NotYet for that numeric intersection. No number alias or void-brand replacement
will be counted as certification of that original declaration.

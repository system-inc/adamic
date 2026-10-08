# Views V4 pre-build plan

Roadmap step 11, task #1p9ste9, under #a03mesg. This is preparation, not a landed slice.

## Current integration method

The compiler's net-change ruling supersedes the historical cherry-pick method
below. For each lane, compare its pinned final tree with its actual fork point,
apply that single diff to the slice, and resolve each conflicted file once.
Produce one commit per lane, listing its original own commit SHAs and subjects.
Apply the final ruled merge judgments once per touched file. Use that lane's
tests and fixtures to verify the resolution, followed by the unchanged slice
comparison, counts, stage3, gap, a-check and mutation sweeps.

The pinned endpoints and net changed paths are in
views-v4-evidence/net-lanes.json. The lane5 fork is 2de6d1ba; shares a, b and c
fork at 926a1d39, share d at 18d1c9da, factory at a62778110, and code, Set receiver
and optional host at 926a1d39. Shared ancestors are not applied a second time.
The Set receiver lane contributes the dependency once; the code net diff retains
f339ca8d and 2efc8949 without overwriting that reconciled receiver work.

A raw final-tree diff can include other slices introduced by lane5's historical
merges. Review shared files by V4 function/hook, not merely by pathname, and retain
the V3 version for excluded admissions. Netting history does not waive the cyclic
ownership, fresh_refused, graph_regions or other-slice exclusions. The net manifest
records candidates, not an authorization to apply every listed path.

Push a partial compiler/views-v4 as soon as the real V3-based compiler builds,
even with red tests. Report exactly which lanes and resolutions it contains,
the failed comparisons and remaining work. This supersedes the older green-only,
single-final-push text below. A successful build is not a completed slice.

## Base and observed state

The required origin/compiler/views-v3 branch is absent. Every origin check in this
preparation returned no matching ref. No compiler/views-v4 branch is created or
pushed. The planning branch is scratch/views-v4-plan from origin/main 54cbc1254.
The isolated dry-run worktree is /workspace/scratch/views-v4-dry-run.

The main dry run attempted 104 candidate own non-merge commits. Eleven picked
cleanly and 93 conflicted, across 1,898 distinct paths. A conflicted pick was
aborted before continuing. Later picks therefore see only earlier successful
picks: this records encountered conflicts, not the complete conflict set of a
resolved integration. Three early callable-helper commits discovered through
merged ancestry were probed after the first pass; the real replay orders them
before the admissions. Full SHAs, changed paths, exit status and conflicting
paths are in views-v4-evidence/commits.json. Complete command output is in
views-v4-evidence/cherry-pick-output.txt. No dry-run compiler is claimed valid.

Git infers many directory-renaming conflicts into stage3/a-check-headers.
Preserve original lane5 fixture paths; do not accept these inferred moves.
Recheck origin at every further preparation or replay pass. When V3 appears,
fetch it, create compiler/views-v4 from its exact tip, and record that immutable
base before any edits or tests. Do not merge main or any other slice into V4.

## Scope and prerequisite cut

Admit only callable members and function values read/called through checked views.
Preserve every other base admission, Refused and NotYet result. Read descriptors
are lazy; an unread unsupported callable does not reject the cast. Actual producer
identity and the complete declared signature decide a callable read. Preserve
this, identity, evaluation order and single receiver evaluation.

Main has neither internal/ir/views.go nor internal/lower/view_callables.go.
The callable scaffold originates in 8346b28c1 and 9fc5c01ab, both mixed array/callable
commits inherited by lane5. Do not cherry-pick either entire mixed commit. V3 is
expected to carry shared view/array infrastructure. If callable scaffold is absent
there, extract only transferred lane5 callable files and minimal named callable
hooks from those commits, recording the cut and showing ordinary arrays unchanged.
The V4-owned helper commits be6e4f33, f13be3e29 and 9c4d09044 then precede 37fa06d39.
No intrinsic array admission is authorized by this extraction.

Never take 2724dabd, fresh_refused or graph_regions fixtures, or their admission
expectation/count changes. Do not import V2 union, V3 array, V5 dictionary or
intersection, or V6 erasure admissions through historical merges. A callable's
argument/result can consume the base's existing checked payload descriptor, but
V4 cannot create an unrelated payload admission. If a named source fixture needs
one, keep its base refusal and report the prerequisite rather than weakening it.

Take each own commit once, never a source merge. Share d inherits share a and the
factory inherits early share a: their shared ancestors appear once in the inventory.
360084ad and 6f7902e0 are equivalent cherry-picks of 09d5525f and 058635b9: take
360084ad and 6f7902e0 once, omit the original copies from the actual replay.
The inventory has 104 candidates, not 104 unique semantic changes.
Follow them with f339ca8d and 2efc8949, preserving both higher-order logical
producer identity and Set logical element domains. The later Set report
0960d58e adds no second receiver conversion.

Factory d48d2ad01 and f339ca8d overlap nested signature recognition and producer
checks. Preserve nested complete descriptors, the immutable logical producer gate,
and per-instantiation identity, not one alternative selected merely to resolve
text. Test the factory negative nested producer and all higher-order negatives.
Capturing generic functions, arbitrary predicates and unsupported callback effects
retain the base boundaries. Optional-host 61d77dbee is limited to its certified
immutable source/const binding subset; it cannot invent fields on plain Error.

## Ruled conflicts to preserve

Use the final applied sections of the two documents at c923d1cc, not their earlier
pending observations. Existing user authorization explicitly includes the Set
producer-domain fix and f339ca8d, resolving that ledger's inclusion question.

- Original judgments for typed closure layout: preserve counted/plain typed code
  pointers, optional/rest/count arguments, readiness and the receiver flag. Never
  cast a function pointer. Keep both existing argument guards and producer masks.
- Original IR tag judgment: preserve the base's unique semantic/physical tags and
  C/Go agreement; do not reintroduce old numeric tag collisions.
- Judgment 8: only independently producer-certified checked callable Union boxing
  escapes the ordinary unknown-function refusal.
- Judgment 10: retain marker result-erasure and stored never-rest call refusals;
  only the separately certified immediate discarded marker operation can pass.
- Judgment 11: plain Error cannot acquire optional NodeJS.ErrnoException fields.
- Judgment 12: retain symbol/body identity proof for overloaded shorthand values.
- Judgment 13: computed finite keys require a registered receiver view contract;
  ordinary unviewed finite selection keeps its refusal.
- Judgment 14: preserve the base nullable-reference/unknown boundary. V4 does not
  replace the base null and undefined representation.
- Judgment 15: normalize callable masks only through already proven phantomBase
  and phantomArrayBase; other intersections retain refusal.
- Judgment 16: retain the scalar literal length control and the Union field guard.
- Judgment 17: preserve the base ProcessCall exception edges; no process admission.
- Judgment 18: boxed adapters allocate/populate the producer fixed/optional/rest/count
  layout, including missing optional slots and actual argument count. Include
  string.h for emitted memset. Caller-sized packing is the required mutant.
- Judgment 19: compare executable IR independently while separately retaining
  FunctionTypeTargets/CallTargets proof; never clear nonempty target evidence.
- Judgment 20: certified void result uses the existing wildcard mask, while real
  undefined retains its distinct representation. Unknown producers remain refused.
- Judgments 21 to 23: retain the base ArrayRecord contained-reference proof and
  existing logical/String/ordinary boxed-array boundaries. They are other-slice
  behavior, not new V4 admissions.
- Judgment 24: retain ordinary boxed Union Map callbacks only without erasing a
  registered checked-view contract; unknown checked callbacks remain refused.
- Judgment 25: retain const intrinsic hasOwnProperty.apply with a literal one-key
  tuple, source evaluation order and all escaping/unknown signature refusals.
- Judgment 26: preserve fixed-object Object values/entries/keys intrinsic dispatch
  without admitting fixed-object-to-Record storage conversion.
- Judgment 27: preserve certified direct overload extras and evaluate ignored actual
  arguments in order, preserving implementation relation and callback proofs.
- Set reconciliation: retain argument-index guard and counted/plain method entry
  wrapper convention around allocation identity and the complete producer domain.

Every actual conflict resolution commit must name the applicable judgment and
its fixture. If one of these files is untouched by the V4 cut, retain it from V3.
Do not cherry-pick unrelated judgment implementations merely because they exist
in the integration history. The original first-table hunk decisions also apply to
all touched shared expression/function/generic/collection/emitter functions:
keep base contextual/readiness/rest/count handling alongside the minimal V4 hook.

## Validation and moved results

The detailed unit's full comparisons override its generic no-whole-package
instruction for the specifically named packages. Use the same installed tools,
Node typings, cohere revision and machine for the immutable V3 base and V4.
Capture all test output directly in logs. Mutants run serially and restore source;
never run a production comparison concurrently with a mutating runner.

1. Run npm ci in stage3/api in both comparison worktrees.
2. Record go test internal/lower, internal/ir, internal/flow, internal/javascript
   and the full internal/oracle on V3, then on V4 with identical uncached settings.
   Compare each leaf by status and diagnostic/stdout, not just aggregate counts.
   Every moved result must be a V4 admission or a new V4 fixture, individually named
   in a Moved-result trailer. Any unrelated move blocks delivery.
3. Run stage3/fixtures on both, then run again with its platform guard lifted
   locally. Restore the guard before committing. Compare the Linux-only host
   records too; do not relabel their expected diagnostics merely to make it green.
4. Run go test ./stage1/... -run 'Gap|Gaps|Probes' on both. Any V4-closed gap moves
   to that package's closed form and updates GAPS.md, with a Moved-result trailer.
5. Refresh TestCountsAreRecorded only for V4 runtime rows. Preserve every unrelated
   row byte-for-byte and separately assert any deliberately removed V4 refusal row.
   A failing full updater is not evidence that the table is refreshed.
6. Audit every changed/new .a outside a Go package under cloud/fast-gate/run.py
   aCheck rules. A negative starts with its precise refusal/type-error header and
   must have a test asserting it. Keep source Node observations unchanged.
7. For each newly operative check, run a semantic check-removal mutant and record
   its catching fixture on V4. Historical lane logs are prerequisites, not new
   validation. Reject build failures and -Werror as mutation evidence.
8. Confirm no cyclic-ownership admissions, other-slice changes, local platform
   guard changes or cohere copies enter the diff. Push the partial V3-based slice
   when it builds, reporting reds; continue the full comparison and mutation sweep
   before calling V4 complete.

New checks require their existing lane mutation runners plus independent ruled
packing, marker, plain-Error, producer identity, overload-set membership, nested
callback, generic instantiation/result, Set storage/domain, optional presence and
single evaluation controls. Build a check-to-fixture manifest from every new
production condition. A runner's success is not a substitute for that manifest.

## Observations and open language questions

No new language ruling is inferred from textual conflicts. Existing lane reports
contain unsupported generic/predicate/rest/brand cases; V4 retains them unless an
included callable proof closes the exact boundary. There are no newly executed
program/Node language comparisons in this preparation. The prior Set unsafe writer
and higher-order programs remain documented source evidence, not measurements on
V3 or this dry run. If an inseparable slice change appears on V3, report the exact
commits and program, measured Node behavior and proposed boundary before stopping.
The current obstacle is the missing required base, not a proved inseparability.

## Commands and setup

Fetched the nine named source branches and the ruled-judgment branch. Inspected
pinned commit histories, excluded merges, deduplicated shared SHAs and attempted
cherry-picks in the isolated main worktree. No package tests, fixture tests, counts
updater or mutants are run on this intentionally incomplete dry-run compiler.

Setup succeeded on the previous recorded code tip before the planning checkout:
Node 0.022s, Go 0.022s, submodules 0.063s, markdown 0.065s, clang 0.155s,
build 35.974s, deferred tests 36.139s, cache 36.140s, done 36.166s.
nproc=5; cpu.max=400000 100000. Source /workspace/adamic-tools/env.sh.
Full setup lines are in views-v4-evidence/setup.txt.

## Candidate own commits and changed compiler files

Each row names what the commit changes. Full fixture/document paths and conflict
paths remain in the JSON inventory. This is a replay inventory, not a list of
commits successfully integrated or pushed.

| Source | Commit | Change | Compiler/test files |
| --- | --- | --- | --- |
| lane5-scaffold | `9c4d09044` | Prepare reached callable reads and code identity signature certificates | `internal/javascript/view_callables.go`, `internal/javascript/view_callables_signature.go`, `internal/javascript/view_callables_signature_test.go`, `internal/lower/view_callables_read.go`, `internal/lower/view_callables_read_test.go`, `internal/native/view_callables_signature.go`, `internal/native/view_callables_signature_test.go` |
| lane5-scaffold | `f13be3e29` | Record callable mutant evidence and lane 2 merge verification | Fixture, count or report evidence |
| lane5-scaffold | `be6e4f339` | Add deferred callable shape helpers and reconcile conservative read demand | `internal/javascript/view_callables_contract.go`, `internal/javascript/view_callables_contract_test.go`, `internal/lower/view_callables_contract.go`, `internal/lower/view_callables_contract_test.go`, `internal/native/runtime/view_callables_contract.h`, `internal/native/view_callables_contract.go`, `internal/native/view_callables_contract_test.go`, `internal/oracle/checked_views_callable_contract_test.go` |
| lane5 | `37fa06d39` | Check fixed scalar callable views at each read | `internal/javascript/view_callables.go`, `internal/javascript/view_callables_signature.go`, `internal/lower/cast_proof.go`, `internal/lower/expression.go`, `internal/lower/interface_cast.go`, `internal/lower/invariance.go`, `internal/lower/object.go`, `internal/lower/view_callables.go`, `internal/lower/view_callables_marker.go`, `internal/lower/view_callables_read.go`, `internal/native/emit_functions.go`, `internal/native/runtime/view_callables.c`, `internal/native/runtime/view_callables.h`, `internal/native/view_callables.go`, `internal/native/view_callables_methods.go`, `internal/native/view_callables_signature.go`, `internal/native/view_fields.go`, `internal/oracle/checked_views_arrays_test.go`, `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_source_test.go`, `internal/oracle/checked_views_callables_test.go`, `internal/oracle/checked_views_lazy_test.go`, `internal/oracle/interface_cast_counts_test.go` |
| lane5 | `946b35c5c` | Rank mixed union callable adapter candidates and retain ABI refusal | Fixture, count or report evidence |
| lane5 | `b6e53fb46` | Check stored erased marker calls and release discarded results | `internal/ir/ir.go`, `internal/ir/views.go`, `internal/javascript/javascript.go`, `internal/javascript/view_callables.go`, `internal/javascript/view_callables_contract.go`, `internal/javascript/view_callables_discard.go`, `internal/javascript/view_callables_signature.go`, `internal/lower/expression.go`, `internal/lower/view_callables.go`, `internal/lower/view_callables_marker.go`, `internal/lower/view_callables_read.go`, `internal/native/emit_functions.go`, `internal/native/runtime/view_callables_contract.h`, `internal/native/view_callables.go`, `internal/native/view_callables_discard.go`, `internal/native/view_callables_signature.go`, `internal/native/view_fields.go`, `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_source_test.go` |
| lane5 | `d1153e5d2` | Certify watcher and canonical name contracts with original witnesses | `internal/lower/view_callables_read.go`, `internal/native/view_callables_methods.go`, `internal/oracle/checked_views_callable_canonical_test.go`, `internal/oracle/checked_views_callable_counts_test.go` |
| lane5 | `c33f71bb1` | Certify original scanner and performance callable witnesses | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_scalar_witness_test.go` |
| lane5 | `15b30747f` | Pin callable mixed result boxing boundary against Node | `internal/oracle/checked_views_callable_scalar_witness_test.go` |
| lane5 | `cb6adb25e` | Adapt callable scalar and boxed union boundaries with directional checks | `internal/ir/ir.go`, `internal/ir/views.go`, `internal/javascript/javascript.go`, `internal/javascript/view_callables_boxing.go`, `internal/javascript/view_callables_contract.go`, `internal/javascript/view_callables_signature.go`, `internal/lower/census_small.go`, `internal/lower/expression.go`, `internal/lower/functions.go`, `internal/lower/view_callables_boxing.go`, `internal/lower/view_callables_read.go`, `internal/native/emit_arrays.go`, `internal/native/emit_expressions.go`, `internal/native/emit_functions.go`, `internal/native/emit_maps.go`, `internal/native/from.go`, `internal/native/library_array_holes.go`, `internal/native/reuse.go`, `internal/native/runtime/view_callables_contract.h`, `internal/native/view_arrays.go`, `internal/native/view_callables_boxing.go`, `internal/native/view_callables_contract_test.go`, `internal/native/view_callables_signature.go`, `internal/oracle/checked_views_callable_boxing_test.go`, `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_scalar_witness_test.go` |
| lane5 | `3e1a7f6fb` | Certify six aggregate callable members and transitive payload reads | `internal/ir/views.go`, `internal/javascript/view_callables_signature.go`, `internal/lower/collections.go`, `internal/lower/functions.go`, `internal/lower/interface_cast.go`, `internal/lower/lower.go`, `internal/lower/view_callables_aggregate.go`, `internal/lower/view_callables_aggregate_test.go`, `internal/lower/view_callables_boxing.go`, `internal/lower/view_callables_read.go`, `internal/lower/view_lazy.go`, `internal/native/view_callables_methods.go`, `internal/native/view_callables_signature.go`, `internal/oracle/checked_views_callable_aggregate_test.go`, `internal/oracle/checked_views_callable_counts_test.go` |
| lane5 | `ebb1f8a61` | Certify four mixed aggregate callable members and optional numeric arguments | `internal/lower/object.go`, `internal/lower/view_callables_boxing.go`, `internal/lower/view_callables_contract.go`, `internal/lower/view_callables_contract_test.go`, `internal/lower/view_callables_read.go`, `internal/native/runtime/object.c`, `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_element_test.go`, `internal/oracle/checked_views_callable_numeric_test.go`, `internal/oracle/checked_views_callable_property_test.go` |
| lane5 | `164dac2c2` | Certify callable array parameters with deferred element read checks | `internal/lower/view_callables_aggregate.go`, `internal/lower/view_callables_boxing.go`, `internal/oracle/checked_views_callable_arrays_test.go`, `internal/oracle/checked_views_callable_counts_test.go` |
| lane5 | `7175993b6` | Certify original optional aggregate callable members | `internal/lower/view_callables.go`, `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_optional_aggregates_test.go` |
| lane5 | `5d0ad1586` | Record optional callable certification and mutant evidence | Fixture, count or report evidence |
| lane5 | `21e31b4ea` | Certify original optional return statement callable | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_ranked_test.go` |
| lane5 | `5ecc56cfe` | Certify original transformation notification callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_ranked_test.go` |
| lane5 | `ea341202e` | Certify ranked aggregate and optional boolean callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_ranked_test.go` |
| lane5 | `f04e467b0` | Certify original local name callable and retain union gap | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `a9fea5756` | Certify original parenthesizer scanner and true literal callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `375160dcd` | Certify exit null and parenthesized callables and retain binding refusal | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `ccdf894e3` | Certify type reference options and write callables with inherited declarations | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `0515a19d9` | Certify declaration access and source file callables and retain method refusal | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `619b65937` | Certify path measure and binary callables and record Set producer boundary | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `685888b2b` | Certify helper hoist and modifier callable reads and defer intrinsic Set members | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `868325297` | Certify scanner token end and full start member reads | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `baa72933a` | Certify export computed source file and diagnostic callable member reads | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `69883a0e2` | Certify source path and resolver callables and retain original read refusals | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `c0e852823` | Certify optional writer line and watcher close callable reads | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `85d285204` | Certify logical named export and string scanner callables and retain arrow union refusal | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `1a3c524a5` | Certify helper case and newline callable reads and record skipped arrow prerequisite | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `fe969d268` | Certify tagged variable conditional binary diagnostic and scanner text callable reads | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `ad2e41d59` | Certify lexical module false and parenthesized type callable reads | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `de5826ab6` | Certify original tagged export property and type tag callable reads | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `004075230` | Certify prefix syntax and static block callable reads and retain for initializer refusal | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `8a5b672e6` | Certify stored substitution internal name call and class update member reads | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `428f10a1b` | Check original resolution settings scanner lexical entry and writer callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `433e1ec09` | Check original tagged parameter and property declaration update callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `c9017a85b` | Check original import call partial emission and inequality member contracts | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `490b24478` | Record callable declaration and expression batch validation | Fixture, count or report evidence |
| lane5 | `19b8c74ab` | Check original tagged property declaration and qualified name callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `4c2cc83ea` | Check original outer accessor parenthesized and host source file callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `f241a5593` | Check original cancellation writer path iteration and type expression callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `77723f55e` | Check original module reflect type function import and directory callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `554cb1122` | Record original name accessor and module callable batch validation | Fixture, count or report evidence |
| lane5 | `343df8fbb` | Check original canonical emit options and package cache callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `70010b531` | Check original conditional constructor export and function type callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `bb275f0f5` | Check original class constructor template scanner and context callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `e7297f111` | Record canonical type and template callable batch validation | Fixture, count or report evidence |
| lane5 | `7d406f6c1` | Check original array bundle export import spread and type callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `d5d8a5eb2` | Check original initializer import result and catch update callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `3729c616e` | Check original reflect try update and System callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `ae16a6605` | Check original Info canonical name and performance duration callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `4810c561d` | Check original class and function creation and update callables | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go` |
| lane5 | `c8aafdab2` | Record original callable continuation batch validation | Fixture, count or report evidence |
| lane5 | `926a1d39d` | Record integrated callable checks and remaining receiver dependencies | Fixture, count or report evidence |
| share-a | `a3dd5fd9c` | Certify 17 lane 5 share a callable pairs and record generic frontiers | `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_share_a_test.go` |
| share-a | `a62778110` | Certify 13 more share a callable pairs and record higher-order frontier | Fixture, count or report evidence |
| share-a | `c07e0670e` | Certify 13 callable pairs and record ranked intrinsic frontiers | `internal/oracle/checked_views_callable_share_a_test.go` |
| share-a | `14a80a604` | Certify 11 more callable pairs and record constructor frontiers | `internal/oracle/checked_views_callable_share_a_test.go` |
| share-a | `18d1c9da9` | Certify 11 callable pairs including the required mapped tracker view | Fixture, count or report evidence |
| share-b | `473904f4d` | Certify 20 share b callable pairs with original-member fixtures | `internal/oracle/checked_views_callable_share_b_test.go` |
| share-b | `8086754d1` | Fix share b evidence verifier and record certification logs | Fixture, count or report evidence |
| share-b | `3df4c66ea` | Certify 14 more share b callable pairs and record receiver dependencies | `internal/oracle/checked_views_callable_share_b_test.go` |
| share-b | `edf888d7c` | Certify 10 ranked share b callable pairs and pin collection blockers | `internal/oracle/checked_views_callable_share_b_test.go` |
| share-b | `51762724d` | Certify 19 share b callable pairs and record six code dependencies | `internal/oracle/checked_views_callable_share_b_test.go` |
| share-b | `8dae78805` | Certify 17 share b callable pairs and record 19 code dependencies | `internal/oracle/checked_views_callable_share_b_test.go` |
| share-b | `83429e12e` | Certify 20 share b callable pairs and record 28 code dependencies | `internal/oracle/checked_views_callable_share_b_test.go` |
| share-b | `5ebb80871` | Certify 20 share b callable pairs and record two code dependencies | `internal/oracle/checked_views_callable_share_b_test.go` |
| share-b | `8d92aaf03` | Certify 20 share b callable pairs and record 14 code dependencies | `internal/oracle/checked_views_callable_share_b_test.go` |
| share-c | `803524917` | Certify first 20 lane 5 share c callable pairs | `internal/oracle/checked_views_callable_share_c_test.go` |
| share-c | `3466a0865` | Certify lane 5 share c callable batch 2 | Fixture, count or report evidence |
| share-c | `6174e76ed` | Certify lane 5 share c callable batch 3 | Fixture, count or report evidence |
| share-c | `076205018` | Certify lane 5 share c callable batch 2 | Fixture, count or report evidence |
| share-c | `e02577901` | Certify lane 5 share c callable batch 3 | Fixture, count or report evidence |
| share-c | `c99da2a64` | Certify lane 5 share c callable batch 4 | Fixture, count or report evidence |
| share-c | `23cabbb41` | Certify lane 5 share c callable batch 5 | Fixture, count or report evidence |
| share-c | `726008abc` | Certify lane 5 share c callable batch 6 | Fixture, count or report evidence |
| share-c | `a2418ec3f` | Certify lane 5 share c callable batch 7 | Fixture, count or report evidence |
| share-c | `572f7fa3b` | Certify lane 5 share c callable batch 8 | Fixture, count or report evidence |
| share-c | `cd22130dd` | Certify 17 original Array push callable pairs through views | `internal/oracle/checked_views_callable_share_c_intrinsics_test.go`, `internal/oracle/checked_views_callable_share_c_test.go` |
| share-c | `4d018fcfa` | Certify lane 5 share c callable batch 12 | Fixture, count or report evidence |
| share-c | `1ca056841` | Certify lane 5 share c callable batch 14 | Fixture, count or report evidence |
| share-c | `8283591a4` | Certify lane 5 share c callable batch 15 | Fixture, count or report evidence |
| share-c | `6e2e4878d` | Record callable share c final certification and code boundaries | `internal/oracle/checked_views_callable_share_c_test.go` |
| share-c | `bd953abe3` | Certify ten lane 5 share c pairs for step 09 | `internal/oracle/checked_views_callable_share_c_step09_test.go` |
| share-c | `edee6e6cd` | Certify eleven more lane 5 share c callable pairs | `internal/oracle/checked_views_callable_share_c_pop_test.go`, `internal/oracle/checked_views_callable_share_c_step09_batch02_test.go`, `internal/oracle/checked_views_callable_share_c_test.go` |
| share-c | `85381238d` | Certify twenty more share c array and Map receiver pairs | `internal/oracle/checked_views_callable_share_c_continuation_counts_test.go`, `internal/oracle/checked_views_callable_share_c_maps_test.go` |
| share-c | `f4d80337a` | Certify twenty more share c Map receiver pairs | Fixture, count or report evidence |
| share-c | `3ce3b1017` | Certify seventeen more share c Map receiver pairs | `internal/oracle/checked_views_callable_share_c_map_boundaries_test.go`, `internal/oracle/checked_views_callable_share_c_maps_test.go` |
| share-d | `d25bdcd0f` | Certify bottom callable ranks for views share d | `internal/oracle/checked_views_callable_share_d_test.go` |
| factory | `1da34a72a` | Certify the first three NodeFactory overload contracts | `internal/ir/views.go`, `internal/javascript/view_callables_contract.go`, `internal/javascript/view_callables_signature.go`, `internal/lower/expression.go`, `internal/lower/view_callable_calls.go`, `internal/lower/view_callable_calls_test.go`, `internal/lower/view_callables_contract.go`, `internal/lower/view_callables_contract_test.go`, `internal/lower/view_callables_read.go`, `internal/lower/view_callables_read_test.go`, `internal/lower/view_maps_callables.go`, `internal/native/runtime/view_callables_contract.h`, `internal/native/view_callables_methods.go`, `internal/native/view_callables_overloads.go`, `internal/native/view_callables_signature.go`, `internal/oracle/checked_views_callable_factory_test.go`, `internal/oracle/checked_views_callable_share_a_test.go` |
| factory | `c1f80ec5f` | Certify import clause and yield overload contracts | `internal/oracle/checked_views_callable_factory_test.go` |
| factory | `d48d2ad01` | Add module generic callable contracts and nested callback descriptors | `internal/ir/call_targets.go`, `internal/ir/call_targets_test.go`, `internal/ir/ir.go`, `internal/ir/views.go`, `internal/javascript/javascript.go`, `internal/javascript/view_callables_boxing.go`, `internal/javascript/view_callables_contract.go`, `internal/javascript/view_callables_nested.go`, `internal/javascript/view_callables_signature.go`, `internal/lower/expression.go`, `internal/lower/generic.go`, `internal/lower/view_callables_boxing.go`, `internal/lower/view_callables_generic.go`, `internal/lower/view_callables_generic_test.go`, `internal/lower/view_callables_nested.go`, `internal/lower/view_callables_read.go`, `internal/native/emit_functions.go`, `internal/native/runtime/view_callables_contract.h`, `internal/native/view_callables_boxing.go`, `internal/native/view_callables_contract_test.go`, `internal/native/view_callables_generic.go`, `internal/native/view_callables_methods.go`, `internal/native/view_callables_nested.go`, `internal/native/view_callables_signature.go`, `internal/oracle/checked_views_callable_factory_generics_test.go`, `internal/oracle/checked_views_callable_factory_higher_test.go`, `internal/oracle/checked_views_callable_factory_test.go`, `internal/oracle/checked_views_callable_share_a_test.go` |
| factory | `5d30aca96` | Record the eight callable contract certificates and limits | Fixture, count or report evidence |
| code | `360084ada` | Preserve Set allocation identity across structural receiver calls | `internal/javascript/javascript.go`, `internal/javascript/readiness.go`, `internal/native/emit_expressions.go`, `internal/native/emit_functions.go`, `internal/native/runtime/adamic.h`, `internal/native/runtime/map.c`, `internal/native/runtime/object.c` |
| code | `6f7902e05` | Certify Set add and has through callable views | `internal/javascript/view_callables.go`, `internal/javascript/view_callables_signature.go`, `internal/native/runtime/view_callables.c`, `internal/native/runtime/view_set_intrinsics.c`, `internal/native/runtime/view_set_intrinsics.h`, `internal/native/view_callables_methods.go`, `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go`, `internal/oracle/checked_views_set_intrinsics_test.go` |
| code | `f339ca8d3` | Check complete higher order callable producer signatures | `internal/javascript/view_callables_higher_order.go`, `internal/javascript/view_callables_signature.go`, `internal/lower/view_callables_boxing.go`, `internal/lower/view_callables_higher_order.go`, `internal/lower/view_callables_read.go`, `internal/native/view_callables_higher_order.go`, `internal/native/view_callables_methods.go`, `internal/native/view_callables_signature.go`, `internal/oracle/checked_views_callable_code_counts_test.go`, `internal/oracle/checked_views_callable_code_test.go` |
| code | `2efc8949c` | Keep Set callable domains separate from physical storage | `internal/ir/ir.go`, `internal/javascript/javascript.go`, `internal/lower/library_map_set.go`, `internal/lower/set.go`, `internal/lower/view_set_domains.go`, `internal/native/emit_expressions.go`, `internal/native/runtime/adamic.h`, `internal/native/runtime/map.c`, `internal/native/runtime/view_set_intrinsics.c`, `internal/native/runtime/view_set_intrinsics.h`, `internal/oracle/checked_views_callable_code_counts_test.go`, `internal/oracle/checked_views_callable_code_set_test.go` |
| set-receiver | `09d5525f1` | Preserve Set allocation identity across structural receiver calls | `internal/javascript/javascript.go`, `internal/javascript/readiness.go`, `internal/native/emit_expressions.go`, `internal/native/emit_functions.go`, `internal/native/runtime/adamic.h`, `internal/native/runtime/map.c`, `internal/native/runtime/object.c` |
| set-receiver | `058635b99` | Certify Set add and has through callable views | `internal/javascript/view_callables.go`, `internal/javascript/view_callables_signature.go`, `internal/native/runtime/view_callables.c`, `internal/native/runtime/view_set_intrinsics.c`, `internal/native/runtime/view_set_intrinsics.h`, `internal/native/view_callables_methods.go`, `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_callable_later_ranked_test.go`, `internal/oracle/checked_views_set_intrinsics_test.go` |
| set-receiver | `0960d58e7` | Report Set receiver verification and integration limits | Fixture, count or report evidence |
| optional-host | `61d77dbee` | Admit certified optional host methods through callable views | `internal/lower/cast_proof.go`, `internal/lower/expression.go`, `internal/lower/library_node_fs_file_test.go`, `internal/lower/refusals.go`, `internal/lower/view_cast_preflight.go`, `internal/lower/view_objects.go`, `internal/lower/view_optional_host.go`, `internal/lower/view_optional_host_test.go`, `internal/oracle/checked_views_callable_counts_test.go`, `internal/oracle/checked_views_optional_host_test.go` |
| optional-host | `3716d57c7` | Record optional host method verification | Fixture, count or report evidence |

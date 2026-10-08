Built: predicate direction reporting in the driver and oracle counts, plus a seven-configuration latent census rerun.
Commits: reporting 460beeca; tested main merge and feature push e58c805469e3cfa36b2d1e37b0ada2a6ae2ade6a; this artifact records that code.
Commands/results: affected packages, vet and filtered Node oracle pass; seven census builds/runs and independent recount pass.
Mutants: dropped report direction, trusted inferred callback, census extra finding/body scope/attribution/output guards and report-artifact mutations caught.
Limits: observed declaration refusals 581 to 518; callback refusals 0 to 451; checked-view field gaps 0 to 246; three census panics and existing readiness failures remain.

# Predicate direction reporting

`adamic c|js|build FILE --explain-checks` reports each emitted overload predicate
call from `ir.Program.PredicateChecks`, with source location, function, overload,
and each true/false direction's proven, checked or unobservable status and reason.
Assertions have only a true direction. Explanations go to stderr; generated source
continues to go to stdout. Unobservable directions count in Proven and separately
in Unobservable. The count writer appends an independent proof-count table to
`internal/oracle/counts.md` without changing unrelated memory baselines.

The 15 overload fixtures emit 17 call sites: 24 proven directions, including
20 unobservable directions, and 9 checked directions. Golden output covers five
fixtures in C, JavaScript and sanitized native build modes. The runtime suite runs
the original source on Node, generated JavaScript on Node and native with ASan/UBSan.
The dropped-false-direction report mutant fails the complete stderr comparison.
The inferred-callback mutant bypasses independent body proof and fails the pinned
callback-argument refusal. Both mutants were restored before the feature push.

A full count regeneration exposed three stock inferred callback contracts. Equality
arrows and named functions are now body-proven independently of the checker's inferred
claim. No body or annotation alone is trusted. Existing initialized-variable/field
readiness failures reproduce at the previous pushed head `3d48edbb`; their memory
baselines were retained. This is not a green full-oracle claim. Logs are in evidence/.
Setup took 28.017 seconds; `nproc` reported 5, with four CPUs available through the
cgroup. Go 1.27.1, Node 24.19.0 and clang 20.1.8 were used.

# Census method

**Every census count is measured on a checker-rejected program.**
The original method is `70456b7:stage3/census/latent/README.md` and REPORT.md.
This rerun uses TypeScript 6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`, with adaptations 10 and 20 applied
from the original prepared pipeline. All 78 compiler root hashes exactly match
the original census. No upstream compiler source is copied into this artifact.

The overlay retains all checker diagnostics and exposes rejected programs only
through LatentLoad. Ordinary Load/LoadOverlay and lower.Lower are disabled; every
run enables LATENT_ASSERT_NO_OUTPUT. The driver imports no backend and emits no IR.
It scans refusals, then attempts each eligible top-level function and statement in
fresh lowering state, skipping diagnosed function bodies by raw diagnostic spans.
Signature-only diagnostics do not skip bodies. Diagnosed dependency bodies are
separate SkippedDependency boundaries. Generic declarations receive no invented
specialization. Each lowering attempt records its first error or recovered panic.

Counts deduplicate `(kind, location, reason, complete diagnostic text)` across
attempts. REPORT.json includes every finding and per-file/per-reason totals;
compressed JSONL retains caller/unit context, phases, diagnostics and eligibility.
FAMILIES.json contains exact diagnostic-message allowlists including complete fix
text, and recounts the named families by exact equality. Callback argument refusals
are separate from predicate declaration refusals. Checked-view gaps are separate
NotYet diagnostics and are not counted as proven admissions.

Six original configurations were reproduced with pinned preparation and feature
SHAs and the original cohere/checker dependencies. Their complete normalized
finding sets match the original report exactly. The seventh adds the pushed
predicate branch to the original cumulative feature configuration in a never-pushed
scratch worktree. Its main/dependency integrations are also present, so global
changes cannot be attributed solely to the predicate verifier.

The scratch-only compatibility changes preserve the method: pin feature SHAs,
resume existing preparation trees, select each configuration's dependency checkout,
remove the newer early-stop callback preflight from the continuing refusal scan
(the same call contracts remain checked by that visitor), record checking pragmas
without stopping, and skip diagnosed callback dependency bodies. The cumulative
conflict recipe left two duplicate closing lines; a scratch repair removed them.
The after merge retains predicate/interface/main enum and cast safeguards plus
original namespace/nested/taste hooks. The exact merge diff (`evidence/feature-integration.patch.gz`), recipes and logs are
in evidence/. Scratch resolution and backend integration are not claimed as native
validated production code. No scratch branch was pushed.

# Family counts

All values in the following tables are **measured on a checker-rejected program**.
The wholesale predicate syntax diagnostic is 581 before and 0 after. Its replacement
proof refusals still matter: declarations are 518 after, and argument contracts 451.
These observations do not establish that 63 declarations became executable.

| Configuration | Predicate declaration refusals | Predicate argument refusals | Checked-view field NotYet |
|---|---:|---:|---:|
| main | 581 | 0 | 0 |
| taste | 581 | 0 | 0 |
| flags | 581 | 0 | 0 |
| namespaces | 581 | 0 | 0 |
| nested | 581 | 0 | 0 |
| cumulative | 581 | 0 | 0 |
| predicates | 518 | 451 | 246 |

The cumulative/after comparison has 4275 common eligible units,
32 newly eligible units and 0 no-longer-eligible units.
On common eligible units, declaration refusals are 581 to 517, argument refusals
0 to 419 and checked-view field gaps 0 to 194. Changed eligibility and moved first
errors remain visible in the raw evidence. A view may add multiple field gaps across
instantiations; these counts are diagnostic sites, not accepted syntax-node counts.

# Full measurement totals

| Configuration | Checker diagnostics | Units | Bodies skipped | NotYet | Refused | Dependency skips | Errors | Panics |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| main | 2165 | 4530 | 256 | 1938 | 1813 | 23 | 0 | 0 |
| taste | 2165 | 4530 | 256 | 1864 | 1708 | 25 | 0 | 0 |
| flags | 1985 | 4530 | 255 | 1177 | 2106 | 23 | 0 | 0 |
| namespaces | 1985 | 4530 | 255 | 1188 | 2129 | 23 | 1 | 0 |
| nested | 2165 | 4530 | 256 | 1943 | 1809 | 23 | 0 | 0 |
| cumulative | 1985 | 4530 | 255 | 1032 | 1983 | 25 | 1 | 0 |
| predicates | 1650 | 4530 | 223 | 1005 | 2651 | 25 | 1 | 3 |

The after run records two refusal-scan panics: `builder.ts` (QualifiedName.Text)
and `transformers/declarations/diagnostics.ts` (ComputedPropertyName.Text). Those
scans stop at the panic, so later refusal syntax in those files is not exhaustively
observed. Their top-level lowering attempts still run. The lowering panic is at
`emitter.ts:685:1` (slice bound `[:-1]`). The ordinary unlocated parser no-symbol error
is retained in outside_roots. All three panics and their unit/phase contexts are in
the raw ledger. They were not erased or converted into proof/admission counts.

# Provenance

| Configuration | Scratch head | cohere | checker |
|---|---|---|---|
| main | 65be00ae33fe4d394de1cb9468ddb472c3184e8f | 715ba94f3608a6500086b1076ce5cb7e51b836db | 8d550c837c90bd1805b047b7eeccc2baac2d5e7a |
| taste | b556f77ab985384d4822b2a7399498972e14116e | 715ba94f3608a6500086b1076ce5cb7e51b836db | 8d550c837c90bd1805b047b7eeccc2baac2d5e7a |
| flags | f1b33ccef61d71d299f139d3569024a1d9a24c75 | 715ba94f3608a6500086b1076ce5cb7e51b836db | 8d550c837c90bd1805b047b7eeccc2baac2d5e7a |
| namespaces | 5c65c6cae00a03a18e4f49c2fafff94001de1989 | 715ba94f3608a6500086b1076ce5cb7e51b836db | 8d550c837c90bd1805b047b7eeccc2baac2d5e7a |
| nested | e5e7073ae6bcb4a13d22ea7536f30f6c3de6bc8f | 715ba94f3608a6500086b1076ce5cb7e51b836db | 8d550c837c90bd1805b047b7eeccc2baac2d5e7a |
| cumulative | 9396f8c212bd68f72b291fbe60b7c751fbbd928c | 715ba94f3608a6500086b1076ce5cb7e51b836db | 8d550c837c90bd1805b047b7eeccc2baac2d5e7a |
| predicates | b209456a2e7743bf2dca47a8cefa97e7b3a9c002 | 7945d102a6c18dd36adf9114a758ce646e8b2359 | d92d9bfee114c80be2c375d72edae966176e3a4f |

REPORT.json also records every preparation/feature SHA, never-pushed branch,
binary hash, overlay hashes and build/run duration. Source hashes are independent
of the checkout paths. All six original baselines retain their original dependency
pins. The after dependency pins come from the main merge, with its typed file paths.

# Commands and mutants

Test outputs were written directly to log files. The reporting release commands were:

```sh
go test ./cmd/adamic ./internal/lower ./internal/ir ./internal/load -count=1
go vet ./cmd/adamic ./internal/lower ./internal/ir ./internal/load ./internal/oracle
go test ./internal/oracle -run 'TestPredicateDirectionCountsAreRecorded|TestNativeAgreesWithNode/internal/oracle/testdata/(weak_narrowed|maybe_number_slots|visits|census_append_overload|census_overload_lie|census_small_overload|census_overload_contracts).a$' -count=1
python3 tools/report_mutants.py
```

The census commands, using scratch trees and the same adapted bytes, were:

```sh
python3 tools/run_comparisons.py REPOSITORY ADAPTED OUTPUT WORKTREES_JSON
python3 tools/audit.py OUTPUT/main-census
python3 tools/audit_predicate_seam.py OUTPUT/predicates-census
python3 tools/audit_corpus.py OUTPUT ADAPTED
python3 tools/audit_output_guards.py MAIN_SCRATCH OUTPUT/main-overlay NEW_TMP
python3 tools/summarize.py OUTPUT ADAPTED
python3 tools/audit_report.py ADAPTED
python3 tools/audit_families.py
```

The predicate-seam audit pins multiple independent refusal sites, a positive body
proof, an argument contract and a checker-diagnosed body skip. Census mutants add
one NotYet to one probe and to `binder.ts:330:1`; the corpus recount requires +1 in
binder and unchanged records in the other 77 files. Signature/body range and
misattribution mutants fail their specific checks. Non-nil IR and permissive ordinary
loader mutants fail the measurement output guards. Report mutants alter source
hashes, file coverage, checker/finding/reason/per-file/unit/outside-root counts,
labels, common-unit counts, deltas and raw body status; every one is caught by the
independent recount. `evidence/census-report-audit.log` records the failures.

# Remaining work

The full oracle/gate is not claimed green; the reproduced readiness failures remain.
The `.a` over-promising-declaration refusal mode remains a TODO hook, as ruled.
Broader object/array/union checked-field contracts belong to the parallel lanes.
Opaque/recursive and unsupported body shapes remain refused. Indirect checked
predicate overload calls remain NotYet. The census panics need separate reduced
reproductions before attributing them to production behavior on a checked program.
This measurement is a blocker observation ledger, not proof of native tsc execution.

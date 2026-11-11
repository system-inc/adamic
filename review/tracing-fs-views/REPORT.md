# Baobab: retire tracing fs views

Implementation: `9ad919ace`, one commit on the fetched origin/main tip
`838809f7a`. The only implementation paths are adaptations 62, 63,
99-correctness-tracing-dumptypes, and stage3/upstream/LEDGER.md. The implementation
checkpoint is typescript/tracing-fs-views-checkpoint. Evidence is kept separately
on typescript/tracing-fs-views-evidence.

Adamic's #baobab ruling, Nov 11 2026 23:29Z, applies the established postfix
non-null check ruling at docs/rulings.md:50. Both widened fs method views are
retired. The calls now use JSON.stringify(descriptor)! and JSON.stringify(legend)!.

## Literal conditions and pins

The pinned TypeScript source is 6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8.

Descriptor is a local const object literal. Its six spreads are local object
literals pinned within dumpTypes. Legend is namespace-level
`const legend: TraceRecord[] = []`. startTracing pushes records with
configFilePath, tracePath, and typesPath; stopTracing can clear the final
record's typesPath. Neither binding is reassigned and no object literal in
tracing.ts declares toJSON. Each adapter checks the literal shape, const binding,
reassignment, toJSON and computed members before accepting either whole-owner
state. LF/CRLF, idempotence and rejection mutants pass.

Adaptation 99 did require amendment: both of its dumpTypes owner pins contained
62's retired fs view. Both pins and their SHA-256 values now contain the checked
serialization spelling. Its two existing subtype-read changes are preserved.
The independent complete candidate apply succeeds, so this downstream pin is
actually exercised, rather than merely edited.

## Results completed before the lane

- Setup at the fetched main tip: exit 0. API npm ci: exit 0.
- Fresh candidate apply: exit 0; install, normal build, and harness: all exit 0.
- All 747 fresh source files exactly equal the measured candidate source files.
- All 79 independent adamic types invocations: exit 0.
- All ten normal emitted JavaScript files: byte-identical. javascript.json
  retains names, sizes, and SHA-256 values. No output normalization is used.
- Runtime exports: all 2,256 own string keys identical; zero symbol keys.
- Both small-project --generateTrace runs: exit 0. Both types.json files are
  378,816 bytes, SHA-256
  8efa301d785bc06e0cad58e012b1854e63bbd5db9c10a70a2941f02e1b1f8027.
  Both legend.json files are 179 bytes, SHA-256
  a463dc2ce94511041fe0a7898dd90695b5d218040db9b15899e8c9fbdf1f683c.
  The exported CLI uses its own sys object, a shared default-library location,
  and an exit callback that explicitly flushes dumpTracingLegend. Project mode
  otherwise does not flush that file. Both runs share absolute project, library,
  and trace paths. The written bytes are compared directly.

## Lowering

Both single-file full tracing censuses use LATENT_FULL=1,
LATENT_ONLY_FILE set to their exact adapted tracing.ts path, and
LATENT_ASSERT_NO_OUTPUT=1. Both have zero checker diagnostics and 23 units.
The repository's full census label says checker-rejected unconditionally;
the retained headers establish the actual diagnostic count of zero.

Before: 340:14 and 359:10 each have the method-parameter variance refusal in
refusal_scan and lowering. After: those refusals are absent. The calls now meet:

| Owner | Site | Kind | Exact reason |
| --- | --- | --- | --- |
| dumpTypes | tracing.ts:340:50 | NotYet | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |
| dumpLegend | tracing.ts:359:53 | NotYet | JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) |

Restoring 62's view restores Refused at 340:14. Restoring 63's view restores
Refused at 359:10. Both return in refusal_scan and lowering, with all 23 units
and zero checker diagnostics in each mutant.

The core.ts production entry first stop is tracing.ts:380:1, Refused,
`an ambient value declaration`, on `declare const require`. The small first-stop
driver calls unchanged production load.Load and lower.Lower and requires no IR
output. No backend is imported or invoked.

## Lane

The complete 398-shard run is in progress. Final ledger and child exits will be
added when it finishes.

## Provenance and discarded measurement attempts

Logs and child exits are retained. Initial setup began before main was fetched
and failed against the old submodule; setup at the fetched tip succeeds. An
initial baseline apply lost the parser while npm ci was run concurrently. It
failed before 99-e wrote anything and was resumed from its exact recorded
boundary using unchanged adapters from the baseline worktree. No concurrent
installation is used during the successful candidate pipeline.

The first project-mode CLI witness wrote types but no legend. A custom sys
identity then disabled tracing, so the final witness retains TypeScript's own sys
object. Its default-library location is explicitly shared because the first
otherwise successful before/after pair differed only in absolute decorator
library source paths. The final raw written files match without normalization.
Those attempts are preserved under initial-, identity-attempt-, and path-attempt-.

Two supplemental LATENT_FULL=0 statement-census invocations were terminated with
SIGTERM when inspection showed they expand statements rather than recording a
production entry first stop. They are not used as successful proof. The actual
production first-stop driver and its output are retained separately.

All saved evidence Go sources use .go.txt. The implementation tree remains clean.

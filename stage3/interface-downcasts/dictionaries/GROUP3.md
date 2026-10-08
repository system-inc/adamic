Built transitive dictionary array/generic reads and merged finite partial records and MapLike onto the existing shared record table.
Commits: 5029b110e5ce468ad90946745174e77394914dd2 plus this merge of records-lowering 456c981b1c108abae5f2c8d6ae665cfd6b92e2fe; integration base ba59427c.
Checks: focused touched-package gate, 39 compiler-source controls, 24 existing record oracle cases, process environment oracle, vet and candidate reproduction pass.
Mutants: skip-check, accept-wrong-shape and drop-transitive-check caught in C and JavaScript over both fixed and record producers; bad producer certificate and removed cycle check also caught.
Not covered: rich CompilerOptions unions, enumeration and uncertified writes; 28 candidate pairs / 222 candidate reads remain, exact reachability unmeasured.

Revised whole-family working estimate: October 18, 2026, 23:00 UTC. This replaces
October 17 after the actual record representation merge exposed storage tags,
producer certificates and ownership integration work. It is a working estimate
for per-pair contracts, not a promise that whole-program tsc compiles. Candidate
counts remain the accepted inventory until a checker-clean program yields the
exact table. No handoff permission or other-lane push is awaited.

## Shared storage

Every mutable MapLike/Record and finite Partial<Record<K,V>> producer uses the
records lane's counted Map table. The dictionary adapter reads that table. The
existing wrapper gains one non-reference numeric slot containing the producer's
physical element type. It neither copies entries nor introduces another record
representation. The reference-value flag alone cannot distinguish numbers from
booleans; this independent certificate allows raw scalar reads to classify the
actual source before testing the declared target contract.

Reference tables classify the existing heap header: strings, boxes, objects and
arrays keep the existing ownership representation. Record tag 14 preserves the
null/undefined slot tags 12 and 13 from the integrated checked-write work. Parent
field reads check the ordinary object heap kind; each element read checks its
logical type, and object/array descendants check at subsequent reads. Absent keys
produce undefined only when the selected contract admits it. Fixed-shape source
objects keep their identity and use the existing source probe.

Record calls carry a prepared read descriptor. When a program has view origins,
the backends conservatively use that descriptor for supported record reads,
including helpers lowered before the cast. Ordinary programs retain record.c's
original read/prototype observations. Unknown or unimplemented read contracts
and enumeration become lazy demanded refusals; a read descriptor cannot authorize
a source-slot write. Ordinary record writes keep their storage check. The
separate direct-view write refusal remains covered by its source fixture.

The native graph ownership pass has no record cleanup adapter. Unsafe writes
into counted records retain the existing cycle refusal; safe fresh record writes
still compile. The existing freshness element edge is used for dictionary reads.

## Source evidence and candidates

The earlier options, nested object and wrong-array controls still pass. Eight
new propagation controls exercise helper functions, generic index arguments,
callbacks and stored fields. Six paths controls cover a valid array, absent
optional container, wrong container, wrong element kind, wrong array element and
absent key. Two root controls exercise structural dictionary casts. A source
write control proves an uncertified fixed-object write stays NotYet.

Sixteen storage controls use compiler-produced records, not hand-written table
adapters: mutable MapLike targets, numbers, booleans, strings, mixed primitive
boxes, finite partial records, objects and arrays. Good values and admitted
absence match source Node in generated C and JavaScript. Wrong container/value
and transitive reads pin exit 70 and the complete named diagnostic in both
backends. Successful native runs include ASan, UBSan and the leak check; release
builds run too.

The MapLike<string[]> dynamic-key pair (type 10957) has 6 candidate reads and is
credited with this per-pair source evidence, including a genuine MapLike producer.
No rich CompilerOptions pair is credited from a narrower primitive union fixture.
The deduplicated inventory remains 29 / 228, with 1 / 6 credited and 28 / 222
remaining. The first pending pair is ParsedCommandLine.options, 111 reads.

## Commands and observed results

All output is retained in logs/group3-*.log; no test command was piped.

- `ADAMIC_GATE_UNCACHED=1 go test ./internal/ir ./internal/flow ./internal/fresh
  ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
  -run 'TestRecord|TestPartialRecord|TestCheckedViewDictionary|TestViewDictionary|
  TestLazyView|TestCheckedViewLazy|TestRegExpGroupCompoundRefusals|
  TestLibraryStringRefusals|TestObjectRefusalsExplainSoundness|TestEnumSlotViews|
  TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/require_node_perf_hooks'
  -count=1 -timeout 10m`: PASS. Lower 3.336s, native 45.927s, oracle 31.081s;
  flow 0.643s, IR/fresh/JavaScript compile (no matching tests).
- After the final regex ownership boundary: `go test ./internal/native
  ./internal/lower -run 'TestRegexProgramsKeepCheckedFieldReads|
  TestRegExpGroupCompoundRefusals|TestRecord|TestPartialRecord|TestViewDictionary'
  -count=1 -timeout 10m`: PASS, native 95.739s, lower 5.469s.
- Final source run: `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle
  -run '^TestCheckedViewDictionary|TestNodeProcessEnvironmentRuntime|
  TestNativeAgreesWithNode/internal/oracle/testdata/records_' -count=1 -v
  -timeout 10m`: PASS, 19.454s. Includes all 39 dictionary source controls,
  the source refusal, all dictionary mutants, 24 record cases and process.env.
- `go vet ./internal/ir ./internal/flow ./internal/fresh ./internal/lower
  ./internal/native ./internal/javascript ./internal/oracle`: PASS.
- `python3 stage3/interface-downcasts/dictionaries/rank-lazy-demand.py --check`:
  `28 candidate pairs / 222 candidate reads remain; exact runtime reachability unmeasured`.
- `git diff --check`: PASS.

Toolchain was already installed and reused: GOPROXY=https://proxy.golang.org|direct,
source /workspace/adamic-tools/env.sh, nproc=5 (quota 4 CPUs). The earlier setup
rerun printed Go 0.021s, Node 0.024s, Markdown 0.072s, clang 0.153s, submodules
3.141s, Go build 83.447s, cache warm 83.575s, done 83.605s; GROUP1 retains it.

The complete touched-package test run was attempted and FAILED. Its log is
preserved, not replaced by the focused PASS. Dictionary-caused regex/producer
refusal overlaps and the callable BindingPattern panic were fixed and their
specific gates now pass. Eight broader lower test groups and five native graph
groups were separately rerun at the pre-merge 5029b110 checkout and fail there
as well. The latter include GraphContainerBoundary's ASan use-after-free, graph
region accounting, closure calling convention and unreleased-anchor checks.
No whole-repository green gate is claimed. Baseline logs distinguish observation
from inference; these remaining failures need their owning lanes/integrator.

## Mutants

For each of skip-check, accept-wrong-shape and drop-transitive-check, two generated
backend variants run over fixed sources and two over shared-record producers.
Every mutant builds and runs to exit 0 with forbidden output (object or 42); the
corresponding unmutated control pins exit 70. The mutation is not caught by clang
or by a later incompatible pointer read. All twelve executions are logged.

A wrong producer certificate changes numeric table creation to advertise boolean
storage in the ordinary records_operations source, which has no view origins.
It builds, then storage_check stops at exit 70 with the exact storage message.
Only that check can catch this certificate error: the actual numeric payload is
unchanged and no dictionary element adapter executes.

Removing the counted-record cycle refusal causes both self-cycle source cases in
TestRecordRefusals to lower successfully, losing their expected refusal. The
mutant's two failures are logged; the original file was restored and the final
normal gate passes. The prior six component mutants and five guard tests also
remain green. Ordinary record semantic/ownership mutants run in the focused gate.

## Merge choices for the integrator

Nine files conflicted; each was resolved hunk by hunk:

- javascript.go keeps NodeBuffer calls and adds all record expression cases.
- expression.go keeps Buffer representation/proofs and phantom members alongside
  record representation, storage comparison and record expression dispatch.
- object.go keeps Buffer, directory and process handlers plus detached own calls.
- statements.go keeps process environment mutations, never evaluation and records.
- refusals.go keeps phantom/Node-process checks, record storage/prototype/delete
  checks and detached own handling. Unsupported signatures defer to actual producer
  lowering, preserving lazy cast admission. Method exemptions are combined.
- optional_widening_census_test.go preserves integrated admission columns and uses
  string(FileName()) for the unchanged checker shim pin; config paths use ToSlash.
- optional_widening_test.go retains integrated lazy optional admission rather than
  reinstating main's old unconditional optional-widening refusal.
- counts.md retains both sets of distinct fixture rows and the integrated optional
  checked-write rows, without inventing replacement counts.
- field_access_paths.a retains the integrated mutable optional source fixture.

Every additional shared hook is listed under this lane in checked-views-plan.md.
Only codex/views-dictionaries is pushed; no other lane, main or PR is modified.

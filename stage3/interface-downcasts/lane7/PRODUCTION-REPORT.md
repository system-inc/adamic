Built: production structural intersection dispatch with seven documented shared hooks; the root-only refusal now passes.
Commits: this group follows 058160aa; latest integration 566c1ec0 is fetched for merge before publication.
Checks: 16 source controls pass against Node, sanitized native, release native and JavaScript; cross-lane checked-view oracle passes uncached.
Mutants: disabling intersection dispatch, accepting a wrong field type and dropping a nested contract each fail the independent refusal test.
Uncovered: complete upstream pairs, compound union/array/callable intersections and deep recursive membership; no full-program tsc gate claimed.

Revised whole-family estimate: October 11, 2026, 23:00 UTC. This is a planning
estimate, conditional on resolving remaining family implementations, with no
shared-hook handoff dependency. Candidate inventory is explicitly used rather
than exact allocation reachability.

Production change:

- lower/expression.go, view_lazy.go, view_contracts.go and view_objects.go contain
  small named structuralViewIntersection / internStructuralViewIntersection hooks.
- IR ViewContract has an explicit Intersection flag distinct from ViewUnion.
- Native and JavaScript view_unions.go route that flag to lane-owned
  viewObjectIntersection methods. Checker-combined Fields represent every
  nonphantom constituent's field obligation and duplicate-field intersections.
- The emitters validate each projected snapshot with existing presence,
  initialization, runtime-type and literal checks. Missing optional values pass.
  Phantom-only arms have no runtime fields. Unsupported descendants keep lazy
  demand refusals. Recursive backedges retain subsequent projected-read guards;
  complete arbitrary-depth recursive membership is not certified by this group.

The first directly structural declared candidate is GeneratedIdentifier.emitNode
(4 reads). Its fixtures are still reduced contract fragments, so they are NOT
subtracted as a complete upstream pair. Additional source controls cover phantom
object brands and two constituents sharing a nested object field. Optional
numeric literal 1 accepts absence and refuses 2 with the pinned named message.

The former red root-only probe now passes with production dispatch. It reads
view.value without reading its invalid child and receives exit 70 naming
view.value.child.count, expected number, found boolean. No overlay or environment
flag is required for normal source tests. Historical overlay artifacts and prior
reports remain dated evidence and are superseded by this implementation.

Mutants change actual production contracts, avoiding redundant later checks:

- skip: clear the explicit intersection dispatch flag on the root-only wrong
  fixture. Both release backends print true and exit 0, failing the exit-70 pin.
- shape: clone the flags child contract and accept boolean instead of number.
  Cloning avoids changing the shared number descriptor used by valid id fields.
  Both release backends print true and exit 0; the independent pin fails.
- nested: remove id from AutoGenerateInfo's field obligations. Both release
  backends print true and exit 0; the independent nested refusal pin fails.

run-source-mutants.py runs each mutated source test. Each Go test exits nonzero
with THREE failed refusal assertions: sanitized native, release native and JS.
The mutant-control test separately proves both release artifacts complete with
exit 0 and stdout true, rather than failing compilation or a sanitizer.

Commands, all output captured directly to logs:

- GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh: exit 0,
  25.572s total, Node .019s, Go .020s, submodules .050s, markdown .062s,
  clang .148s, build 25.421s, cache warm 25.544s; nproc 5, quota 4 CPUs.
- source /workspace/adamic-tools/env.sh before every Go command.
- go test ./internal/lower ./internal/oracle -run
  '^(TestViewIntersectionContracts|TestCheckedViewIntersection.*)$' -count=1 -v:
  PASS, lower .203s / oracle 7.145s before six extra source cases.
- go test ./internal/oracle -run '^TestCheckedViewIntersectionSource$'
  -count=1 -v: PASS, 16 source cases, 9.777s.
- python3 stage3/interface-downcasts/lane7/run-source-mutants.py
  /tmp/intersections-production-mutants: driver PASS; three intentional failing
  test logs retained, each with both release and sanitized checks caught.
- go test ./internal/lower ./internal/ir ./internal/javascript ./internal/native
  -run 'TestView|TestLazyView|TestChecked|TestOptional' -count=1: PASS,
  lower 3.330s, IR no tests selected, JavaScript .690s, native 3.963s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView'
  -count=1: PASS, 38.570s.
- go vet ./internal/lower ./internal/ir ./internal/javascript ./internal/native
  ./internal/oracle: PASS with empty log.

Candidate counts remaining after this group: 45 pairs / 769 reads in the shared
intersection queue across owners. Complete upstream pairs newly certified: 0.
Production fixture families demonstrated: four (EmitNode fragment, Named &
Counted, object brand and duplicate-field intersection). Primitive brands are
not lane 7 completions. The historical 198 / 1,145 counts are not used as exact
remaining work. No cohere code copied, no PR, only own branch pushed.

October 8 production follow-up:

The production group was pushed at 593db75392fd5f1cdaf356325c5ef9bfad42c42e
(implementation 47c6c23a), after merging integration 566c1ec0 cleanly. Its
post-merge uncached checked-view oracle rerun passed in 34.299s.

The follow-up adds a named viewIntersectionReadFamily call in view_lazy.go.
Until their dispatch is implemented, union selection containing intersections,
compound supported children and recursive payloads refuse at demanded reads.
Unsupported descendants remain their own lazy obligations, and casts do not
refuse just because their descriptors are unsupported. Checker-type recursion
is inspected as well as the descriptor graph: optional recursive descriptors
can be copied before canonical Fields finish construction. The type walk
prevents those temporary empty copies from silently certifying a recursive
intersection. Complete recursive membership remains pending.

New Node controls show unread union-intersection and recursive casts print ok in
both backends. The read counterparts refuse compilation with field value and
unsupported union intersection / recursive intersection payload. Both bypass
mutants compile valid Go and fail those independent demanded-read assertions
with got <nil>. This is a deliberate refusal for unfinished families, not a
runtime implementation claim.

Candidate ownership accounting, using the published static queue:

- 15 explicit object-intersection candidates / 64 reads remain in lane 7's queue.
- 23 primitive-brand candidates / 692 reads belong outside this lane.
- 7 private builder/array alias overlaps / 13 reads remain unclassified; do not
  assume these are phantom or primitive merely from their display.
- Total remains 45 / 769. Newly certified complete upstream pairs remain 0.
  The EmitNode fixtures omit upstream internalFlags and AutoGenerateInfo.flags,
  among other declarations; the pinned original definition confirms that they
  are reduced fragments and must not be counted as a completed full pair.

The original type definitions were read through the existing submodule's git
objects at 050880ce59e30b356b686bd3144efe24f875ebc8 and cross-checked against the
same public commit. No source file was copied into this repository.

Leak checks are asserted on successful fixtures. An initial broader follow-up
run wrongly invoked the leak runner on intentional fatal refusals; it reported
exit 70. The harness was corrected to the existing successful-program convention
and rerun. Fatal refusals still run sanitized native, release native and JS with
verbatim output/exit pins.

Additional evidence:

- Compound demand control .376s PASS; bypass mutant .187s expected FAIL.
- Recursive demand bypass mutant .325s expected FAIL, got <nil>.
- Final scoped lower/oracle gate: lower .557s, oracle 11.543s PASS before the
  recursive payload was changed to a wrong value beyond the optional backedge.
- Final package filter: lower 4.448s, JavaScript .662s, native 3.913s PASS;
  IR has no tests selected by this filter. Touched-package vet PASS.
- Uncached checked-view oracle is rerun after final guard/leak changes; its final
  log is retained with the follow-up evidence. No full repository gate claimed.

Working whole-family date stays October 11, 2026, 23:00 UTC. No external hook
handoff is blocking further implementation. Compound selection and canonical
recursive runtime descriptors remain the next implementation work.

Final uncached checked-view oracle rerun: PASS, 34.669s. Final candidate object
queue still 15 pairs / 64 reads, plus 7 overlaps / 13 reads unresolved; no
complete upstream pair is subtracted by this follow-up.

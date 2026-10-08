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

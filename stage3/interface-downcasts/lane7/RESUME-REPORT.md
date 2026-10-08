Built: structural intersection descriptors, source fixtures, actual source mutants, and an incomplete shared-hook handoff.
Commits: integration ba59427ccc7afecae29a305c41e6e9c7867e5610 fast-forwarded into codex/views-intersections; this checkpoint follows it.
Checks: focused production and overlaid source tests pass; explicit root-only conjunction probe fails in both backends.
Mutants: skip checks, accept wrong field type, drop nested check each fail the independent exit-70 test in both backends.
Uncovered: production all-member dispatch, complete upstream pairs, compound intersections and exact reachability; zero production completions.

Revised working date: October 12, 2026, 23:00 UTC, conditional on shared
all-member dispatch landing promptly. The discovered dispatch gap makes this a
high-risk planning estimate. There is no evidence supporting an unconditional
whole-family delivery date.

The merge was a clean fast-forward: integration already contained this lane's
previous bd7042cc tip. No individual lane was merged. Refetch before publication
still reported ba59427c. No shared implementation file was edited.

The authoritative lazy report does NOT publish an allocation-exact reachability
table: it explicitly marks that measurement UNMEASURED because checker-rejected
tsc supplies no whole-program IR. Its newer candidate intersection queue has
45 pairs / 769 reads across owners. lazy-pair-progress.json ranks that queue,
retains every site, and leaves exact remaining pairs and reads null. Zero pairs
or reads are subtracted. Historical 198 / 1,145 conservative totals are not used
as exact remaining work. Primitive brands remain other owners' work.

The highest-count queue entries are primitive brands or compound aliases.
GeneratedIdentifier.emitNode is the first directly structural declared object
intersection (4 candidate reads). Four reduced-contract fixtures exercise its
first arm, second arm, nested AutoGenerateInfo, shared helper and optional absent
prefix. They are shape prototypes, not certification of the full upstream pair.
Named & Counted controls additionally exercise nested number/string failures.
Node runs the original .a sources. Adamic release C and JavaScript compile those
same sources with the experimental Go overlay. Named wrong-value messages are
pinned verbatim in checked_views_intersections_test.go.

The lane-owned builder preserves constituent IDs as Members and checker-combined
field types in Fields, reserves recursive IDs, and drops proven phantom-only
arms using phantomField from d90994da. It excludes primitive brands, array,
callable, tuple, indexed and nominal constituents. Unsupported descendants stay
lazy descriptors. Production source admission stays disabled.

IMPORTANT: shared-hooks.patch is explicitly INCOMPLETE. Do not apply its four
admission hooks as a finished feature. viewObjectUnion in both backends ignores
Members on ViewObject, so flattened metadata checks later field reads but does
not check every constituent when the intersection-valued field itself is read.
The root-only fixture reads view.value and compares its identity without reading
its invalid descendant. Node prints true; both overlaid backends incorrectly
exit 0 and print true. The explicit negative probe requires exit 70 and fails
both assertions. This demonstrates the remaining gap rather than hiding it.

Required integrator seams: lower/expression.go, lower/view_lazy.go,
lower/view_contracts.go and lower/view_objects.go are supplied as four reviewable
hunks. Additionally native/view_unions.go and javascript/view_unions.go must
recognize the conjunctive descriptor and route evaluated snapshots through the
lane-owned all-member components and a shared field/member matcher. Those shared
files belong to other lanes; the user's territory rule requires this handoff.
The prior component matcher is not connected to production source by this push.
The source tests are explicitly gated until that handoff lands; skipped tests
are never counted as completed pairs.

Commands (output captured directly to files, not piped):

- GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh: exit 0, total
  197.337s. Node .024s, Go .025s, markdown .068s, clang .186s, submodules 8.236s,
  Go build 197.188s, cache warm 197.301s. nproc=5; cgroup quota 4 CPUs.
- source /workspace/adamic-tools/env.sh before Go commands.
- go test ./internal/lower ./internal/oracle -run
  '^(TestViewIntersectionContracts|TestCheckedViewIntersection.*)$' -count=1:
  production focused gate passes; source/negative overlay probes skip explicitly.
- python3 lane7/make-integration-overlay.py /tmp/intersections-hooks:
  creates gofmt-formatted shared copies, overlay.json and incomplete patch;
  shared tree files remain untouched.
- ADAMIC_INTERSECTION_HOOKS=1 go test -overlay=/tmp/intersections-hooks/overlay.json
  ./internal/lower ./internal/oracle -run
  '^(TestViewIntersectionContracts|TestCheckedViewIntersection.*)$' -count=1 -v:
  lower .236s, oracle 2.820s pass (component controls plus eight source controls).
- python3 lane7/run-source-mutants.py /tmp/intersections-hooks/overlay.json
  /tmp/intersections-source-mutants: driver exits 0; each of three Go test runs
  exits nonzero because BOTH independent backend refusal assertions fail.
  Every mutant exits 0 with stdout, so a sanitizer crash is not the catcher.
- ADAMIC_INTERSECTION_CONJUNCTION_PROBE=1 go test with the overlay and
  -run '^TestCheckedViewIntersectionRootConjunctionProbe$' -count=1 -v:
  expected RED, .276s, both backends exit 0 instead of 70.
- Broader overlaid checked-view regression filter initially fails two optional
  error probes because pinned @types/node is absent. npm ci --prefix stage3/api
  installs the lockfile's three packages; regression is rerun afterward.

No full gate or whole-program tsc compilation is claimed. No cohere code copied.

Final regression rerun after dependency repair: lower .225s, oracle 31.706s,
PASS. Final production focused rerun: lower .686s, oracle 1.023s, PASS.
Remaining after this push: exact pairs/reads UNMEASURED; candidate queue still
45 / 769 across owners; completed production pairs/reads 0 / 0. The next required
step is shared all-member runtime dispatch, not another merge of lazy admission.

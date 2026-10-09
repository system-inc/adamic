Runtime review fixes for step 16 generic function values are complete.
Commit: this runtime review unit, on compiler/generic-values after 4bd9ab8f.
Checks: generic-value oracles, full lower, reader guard, JSON layout, counts and lane preflight passed.
Mutants: existing result/identity/alias mutants and four planted runtime mutants were caught.
Uncovered: other machines and workloads; byte measurements describe live closure storage, not RSS.

JSON now has a static adamic_object with heap {0, adamic_kind_object, 0}
and an empty shape. The other six built-in identities remain complete closure
structs. The new adamic.h and library_language.c lines use tabs.
TestGenericValueLibraryIdentityViews compares enumeration and typeof with Node,
and reads the closure identity fields under both release and ASan/UBSan.
Its first uncached run took 6.62 seconds.

bench/map_workload.a is absent. runtime-reference-set.a scales the object-keyed
Set shape in internal/oracle/testdata/set_foreach_objects.a to 1,024 keys and
30,720,000 membership queries. Its source on Node, JavaScript backend and both
release native builds printed 30720000. The baseline compiler carries runtime
50654a40; its cached runtime header and map source were checked against that
commit. Both builds use clang 20.1.8, -O2, -ffp-contract=off and
-fno-optimize-sibling-calls, without counters or sanitizers.

The machine is Linux x86_64, AMD EPYC 9V74; nproc reports 5, with cpu.max
400000 100000 (four CPUs). Seven pairs alternate base/head and head/base order.
No test suite ran during measurement. The first measurement was 0.239218 s
base and 0.294098 s head, best of seven: +22.94%. Static selection alone gave
+4.20%, so the address-only lookup also avoids repeated mode checks at hash
collisions. Final best of seven: 0.236179 s base and 0.201664 s head, -14.61%.
Final load averages: 0.28/0.29/0.52 before, 0.34/0.30/0.53 after.
Raw runs and intermediate measurements are in runtime-bench*.json.

The emitter selects address-only Map/Set constructors and array equality for
IR types that cannot hold a closure. Closure and boxed Union retain function
identity. Union is conservatively function-capable in the IR. Array union
comparisons still need their ordinary scalar kind discrimination. The lowerer
refuses widening a function to object/unknown before C emission, observed with
a Set<object> probe; this optimization does not turn such an input into an
unchecked object. Function adapters keep their ruled identity semantics.

runtime-closure-pressure.a holds 20,000 captured closures simultaneously.
The release probe wraps allocation and actual deallocation, recording live
closure allocation sizes and slab slot sizes; it does not change production
code. Both baseline and head printed Node's 199990000 and ended with zero live
closures. Header size grew from 40 to 48 bytes. Peak requested closure bytes
grew from 960,000 to 1,120,000: +160,000 bytes. Occupied 16-byte slab slots grew
from 960,000 to 1,280,000: +320,000 bytes. Chunk reservations, allocator metadata
and process RSS are not these measurements. Commands and archive paths are in
runtime-pressure-measure.log.gz; runtime-pressure-measure.py and the C probe
reproduce the measurement after emitting both fixture C files.

Mutants and catchers:

- Wrong equateValues result: TestGenericValueWrongResult,
  TestGenericValueIndexWrongResult and TestGenericValueUtilityWrongResult
  detect stdout disagreement with Node in native and JavaScript.
- Wrong declared-default result: TestGenericValueDeclaredDefaultWrongResult
  detects the changed constant result in both backends.
- Per-adapter declaration identity: TestGenericValueWrongIdentity detects
  stdout disagreement in both backends.
- Alias escaping into a polymorphic slot: TestGenericValueAliasEscapeMutant
  observes the required named refusal while Node executes the program.
- JSON shape count 1: TestGenericValueLibraryIdentityViews exits 1.
- Original JSON closure layout: the same test exits 1 on its missing shape.
- Function keys incorrectly assigned the address-only constructor:
  TestGenericValueComparer fails with native stdout disagreement.
- Function arrays incorrectly assigned address-only equality:
  TestGenericValueComparer fails because indexOf returns 1 instead of Node's 0.

The four new source mutants compile successfully. Each was restored, then the
JSON control and comparer passed again. Patches and failing logs are retained.

Commands and observations (all test output went to log files):

- export GOPROXY='https://proxy.golang.org|direct'; timeout 180 bash cloud/setup.sh;
  source /workspace/adamic-tools/env.sh. Setup passed: Go 0.019 s, Node 0.023 s,
  submodules 0.063 s, markdown dependencies 0.070 s, clang 0.155 s,
  go build 37.346 s; total 37.535 s. nproc: 5.
- go test ./internal/oracle -run '^TestGenericValue' -timeout 90s -v:
  all twelve tests passed, package 17.330 s.
- go test ./internal/lower -timeout 90s: full package passed, 38.170 s.
- go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s:
  passed, 10.426 s.
- go test ./internal/native -run '^TestGenericValueLibraryIdentityViews$'
  -timeout 90s -v: passed, first uncached leaf 6.62 s.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -timeout 210s
  -args -update-counts: passed, 85.488 s; counts.md unchanged.
- Lane preflight: git fetch -q origin main devtools/fast-gate cloud/merge-tree;
  git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -.
  Passed: gofmt/tools on 285 Go files, t.Parallel on 23 test packages,
  a-check on 100 .a files, vet on 23 packages, 9.9 s.
  The required lane check is run again after commit, before push.

This addresses runtime review of 4bd9ab8f toward roadmap step 16. No full gate
or unrelated package suite was run. The admission and census evidence remains
in the preceding unit; this unit changes runtime representation and selection.

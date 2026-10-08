Built: production admission and checked reads for one structural object plus scalar primitives, with optional undefined.
Commits: integrated ba59427ccc7afecae29a305c41e6e9c7867e5610; implementation tip is the commit containing this report.
Commands/results: selected production gate passed, source oracle 6.230s, lower 0.893s, oracle package 9.285s; vet and IR passed.
Mutants: skipped outer check, accepted wrong literal/scalar, dropped nested check; every variant caught in both release backends.
Uncovered: all 42 original candidate pairs / 181 candidate reads still need original-pair witnesses; exact reachability unmeasured.

Revised whole-family working date: **October 16, 2026, 23:00 UTC**. This is an
estimate including the remaining adapters and original-pair fixtures, not an
observed completion. Shared hook handoffs are no longer prerequisites: the lead
explicitly authorized this lane to implement minimal shared hooks.

The lazy inventory replaces the historical 73/267 demand with 42 candidate pairs
and 181 candidate reads. Its ranking and every candidate read site are preserved
in lazy-pair-progress.json. Exact production reachability requires a checker-clean
program and remains unmeasured. This report does not subtract reduced controls
from original-pair demand. Remaining: **42 candidate pairs / 181 candidate reads**.

The first reduced shapes model SourceFile.externalModuleIndicator (20 reads) and
Diagnostic.messageText plus three related messageText pairs (23 reads). Eight
source fixtures provide Node reference outputs, valid string/object/true values,
explicit undefined where admitted, wrong primitive rejection and transitive
nested rejection. Both compiled backends pin the full field, expected type,
found type and exit 70. Native runs include sanitized/leak-checked and release
builds; JavaScript runs execute under Node. These reduced Node/Chain interfaces
are not the full TypeScript declarations, and original-pair coverage remains zero.

Minimal shared hooks are in lower/object.go and lower/interface_cast.go (named
objectPrimitiveViewType guards), native/view_fields.go and
javascript/javascript.go (named viewObjectPrimitive dispatch). The lane-owned
runtime checks presence/readiness before payload inspection and selects scalar
members including literals. The sole object member retains its shared contract
and descendant obligations. Selecting it does not certify its entire structure.
Unsupported descendant reads continue to refuse lazily. Accessor admission is
unchanged. shared-hooks.patch records the four hook hunks; no overlay is used by
the final production gate. make-hook-overlay.py is a retained development aid.

Mutants run independently on lowered fixture IR, without changing compiler source:

- Allow false instead of true: both backends print false and exit 0; the outer
  named refusal pin catches it.
- Clear the outer view read: native prints absent and JS false, both exit 0;
  the refusal pin catches the skipped selection.
- Accept boolean for nested code/flags declared number: native prints 0 and JS
  false, both exit 0; each nested expected-number refusal pin catches it.
- Clear the nested code/flags view check: the same exit-0 outputs fail each
  transitive refusal pin independently.

Earlier development attempts to remove unchecked string reads crashed; those
were discarded and are not counted as semantic mutant evidence. The retained
numeric/boolean variants produce executable release code in both backends.

Validation command (environment sourced from /workspace/adamic-tools/env.sh):

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestCheckedViewObjectPrimitiveSource|TestCheckedViewLane4HelperReads|TestCheckedViewMixedSelection|TestLazyView|TestSharedArrayContractAdapter|TestDefaultTaggedInterface' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
go test ./internal/ir -count=1 -timeout 10m
git diff --check
```

The selected gate ran lower and oracle tests; native and JavaScript packages
reported no matching tests, and their emitters were exercised by the oracle.
The source oracle had no skips. LazyViewCensus and LazyViewAdaptedCensus skipped
because their opt-in census environment settings were absent. This is not a full
gate. IR passed in 30.751s before the four production hooks were applied; hooks do
not edit IR. Final production vet and diff check passed. Logs are in logs/.

Setup completed in 406.205s: Node .062s, Go .074s, clang .404s, markdown .165s,
submodules 12.938s, Go build 405.990s, cache 406.157s. nproc=5, CPU quota=4.
GOPROXY was https://proxy.golang.org|direct. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Not covered: array/Map/callable alternatives, intersections, multiple object
alternatives, indexed/element reads, full tsc declarations, inheritance/static
slot normalization, packed optional scalar storage, and original helper/generic/
callback paths for this family. Wrong or unsupported layouts fail closed;
no complete-family or exact-reachability result is claimed. Next work follows
the ranked ledger, beginning with intersection and string|number|PseudoBigInt
shapes, then array alternatives. Integrator receives this lane tip through the lead.

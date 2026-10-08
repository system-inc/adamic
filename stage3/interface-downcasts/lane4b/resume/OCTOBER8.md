Built: lazy array alternatives and scalar/object/array boxed field producers; added literal and comment source controls.
Commits: builds on dc5e5f26 and integration ba59427c; implementation tip is the commit containing this report.
Commands/results: selected production regression gate passed; oracle 42.283s; vet and diff check passed.
Mutants: outer removal, wrong-member and wrong-nested-shape acceptance, and transitive removal caught in both release backends.
Uncovered: 42 original candidate pairs / 181 candidate reads remain; reduced controls model 22 pairs / 103 reads.

Whole-family working target: **October 13, 2026, 23:00 UTC**, brought forward
from October 16. The adapter reused shared array machinery without a new element
runtime. This estimate still includes actual NodeArray contracts and original-pair
witnesses, intersections, dictionaries, tuple/element and callable cases.

Added next direct shape string|number|PseudoBigInt (10 candidate reads). Valid
string, number and object alternatives run; boolean outside the union refuses at
type.value with expected/found pinned. Wrong nested negative and base10Value reads
refuse at their names. This type has no absent alternative.

Added the comment array shape (16 candidate pairs / 50 reads). Valid string,
array and explicit undefined values run. Wrong boolean refuses at value.comment;
wrong nested flags refuses at first.flags. NodeArray<T> is currently a readonly
array alias in the reduced fixture; this is not a full TypeScript NodeArray
contract or a certificate for its inherited metadata. The rank-3 object union
with an intersection remains uncovered, as do all original read contexts.

Minimal new shared hook: lower/object.go invokes objectPrimitiveBoxedField to
admit real boxed union property initializers. It permits scalar unions and the
supported single structural object/array union only. Native already stores the
heap reference and physical tag in these slots; no native slot emitter changes
were needed. Positive controls now generate and store number/object and
string/array unions and test both runtime branches. The hook is listed in the
lane's plan. Scalar plus array membership reuses shared ViewArray descriptors;
element reads keep shared readiness and nested field obligations.

All 15 source cases were held to Node first, then native sanitized/leak-checked,
native release and JavaScript on Node, with exact stdout, stderr and exit pins.
Wrong-value programs deliberately differ from Node by their named exit-70
refusal. No source oracle skips were allowed.

Independent mutant observations, native then JavaScript:

- Accept boolean as a scalar union member: comment and literal print wrong and
  exit 0; each outer expected/found refusal pin catches the mutant.
- Remove the outer union check: comment prints absent / wrong and literal prints
  wrong / wrong, all exit 0; each refusal pin catches the mutant.
- Accept boolean for nested flags declared number: nested:0 / nested:false,
  exit 0; the first.flags pin catches it.
- Accept number for nested negative declared boolean: true:123 / 42:123,
  exit 0; the member.negative pin catches it.
- Remove the nested flags check: nested:0 / nested:false, exit 0.
- Remove the nested negative check: false:123 / 42:123, exit 0.

The last two mutations independently fail the respective nested refusal pins.
Existing first-group mutants were rerun as well. Development exposed an unrelated
narrowing guard when an outer mutant used an unboxed boolean field; that attempt
was not counted. A real mixed-union producer and the new boxed-field hook made
the independent removal mutant executable. Changing the nested boolean member
also changes its redundant allowed-literal metadata, so an unrelated literal
check cannot kill the wrong-shape acceptance mutant.

Validation (all output in logs):

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestCheckedViewObjectPrimitiveSource|TestCheckedViewLane4HelperReads|TestCheckedViewMixedSelection|TestLazyView|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestCheckedViewArrays' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
git diff --check
```

The native/JavaScript package selection contained no matching package-local tests;
the oracle exercised both emitters. Opt-in LazyViewCensus and AdaptedCensus skipped.
This is a selected gate, not the full repository gate. Setup/toolchain remain those
recorded in REPORT.md; resumed nproc=5. Fetched integration remained ba59427c.

Candidate bookkeeping in lazy-pair-progress.json keeps original completed counts
at zero. Four reduced shape controls model 22 pairs / 103 candidate reads; 20 pairs /
78 reads lack even a reduced control. **Original witnesses remaining: 42 pairs /
181 candidate reads.** Exact reachability remains unmeasured, as agreed with the
lead. These observations are not counts of production tsc reads proven safe.

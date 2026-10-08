Built optional array descriptors preserving their lazy element edge, after a clean merge of integration ba59427c.
Commits: this checkpoint follows ba59427ccc7afecae29a305c41e6e9c7867e5610; the published lane SHA is in the handoff.
Validation: seven Node-held probes cover present, absent, undefined, lazy, wrong-array, wrong-element and wrong-descendant-field reads.
Mutants: native/JavaScript array kind, native/JavaScript eager validation and JavaScript transitive field checks are caught during execution.
Limits: production runtime reachability is unmeasured; revised native-array family estimate is October 10 at 12:00 MDT.

## Scope and integration

The integration merge fast-forwarded cleanly. Its REPORT.md records the prior
conflict choices; none were reopened. Callables remain owned by lane 5, and no
callable files or certification hooks were changed. New tips go to the lead for
the integrator, rather than directly merging another lane.

The largest array candidate in lazy/census/read-demand-pairs.json.gz is
Symbol.declarations, Declaration[] | undefined, 364 reads. Its source property
is optional. Seven standalone probes hold that optional array mechanism to Node
and both backends, with a small representative Declaration payload. They do not
compile the unchanged complete TypeScript Symbol or Declaration interfaces.

On ba59427c, descriptor construction made this optional array an object-union
shape with no discriminant. Emitting the JavaScript field read panicked with
compiler bug: object union has no checked discriminant. The named hook
viewOptionalArrayContract in internal/lower/view_array_adapter.go keeps the
present array's kind and element edge, sets undefined admission, and preserves
the original declared name. Its three-line shared dispatch is in strictViewContract
in internal/lower/view_contracts.go. No executable storage representation changed
and no additional native array metadata was added in this checkpoint.

A field read never visits its elements. The lazy control contains a valid first
object and a later object whose name is a number; length and the first name still
print 2 and first. The selected element check and its later name check remain
independent. Missing optional fields and explicit undefined print missing.
The three negative controls stop with exit 70 in sanitized native, release native
and emitted JavaScript, with exact diagnostics pinned in the oracle:

- symbol(raw).declarations expected Declaration[] | undefined, found number.
- declarations[0] expected Declaration, found number.
- declaration.name expected string, found number.

The original Node controls print present, undefined and 9 for these unsafe casts;
optional-declarations-node-controls.json records those source outputs, separately
from the checked language's additional refusals.

## Measurement and remainder

The requested exact-reachability measurement in lazy/REPORT.md is explicitly
unmeasured in its adapted follow-up, lazy/ADAPTED-CENSUS.md: checker-rejected
entries have no executable IR or shared allocation-flow graph. The integrated
REPORT.md also says runtime reachability is unmeasured. All 2,936 descriptors
are admitted, while zero production entries compile. No fixture result here
changes those production observations. No checker diagnostic is bypassed.

lazy-array-priority.json freezes candidate source witnesses from that branch's
read-demand ledger, solely to choose work order. This is a static candidate
queue, not exact runtime reachability and not a production unlocked count.
Earlier backend tests cover other mechanisms, but are not silently mapped to
full TypeScript pair certificates. This new ranked fixture ledger starts with
one representative obligation, rather than claiming all prior mechanisms fully
certify every source pair.

| Array family | Candidate inventory | New ranked fixture obligations held | Candidate obligations not yet held by this ledger | Exact runtime remainder |
| --- | ---: | ---: | ---: | --- |
| Array contracts | 334 pairs / 3,189 reads | 1 / 364 | 333 / 2,825 | Unmeasured |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 | Unmeasured |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 | Unmeasured |
| Own array fields | 30 / 72 | 0 / 0 | 30 / 72 | Unmeasured |
| Tuples | 6 / 9 | 0 / 0 | 6 / 9 | Unmeasured |

Families overlap; do not sum this table. Next candidates include Block.statements
110, UnionType.types 107, Signature.typeParameters 90, SourceFile.statements 72
and Signature.parameters 71. Derived NodeArray representation, own fields,
source-certified reference writes, required undefined-valued array fields,
unwired consumers and tuple contracts remain incomplete. The revised date is an
estimate for this array family, not an estimate for whole-tsc compilation or
checker adaptations owned elsewhere.

## Validation and mutants

The first broad restored run found an invalid proposed mixed-element lazy probe:
the frontend cannot lower number | { name: string; } arrays. That witness and
its initial eager-scan results were discarded. The corrected lazy probe uses
representable object elements. The five final mutants require an executed wrong
value or panic/ASan refusal mismatch, and reject build failures, compiler panics,
JavaScript syntax failures and frontend NotYet stops as proof. They restore
production files in finally. Logs retain the initial gate failure as a failure.

- native-array-kind: removing the field array check causes an ASan failure when
  retaining the numeric payload as an array, caught by wrong-array.
- javascript-array-kind: accepting a non-array prints present, caught by wrong-array.
- native-eager-elements: inspecting every object's name at the field read reaches
  the malformed later name, caught by lazy.
- javascript-eager-elements: the same eager traversal is caught by lazy.
- javascript-transitive-field: accepting the later numeric name is caught by wrong-field.

run-optional-array-mutants.py and optional-array-logs retain each executed witness.
Restored uncached checked-view validation and touched-package vet are recorded in
those logs. The full repository gate and production census are not claimed.

Setup used GOPROXY=https://proxy.golang.org|direct and bash cloud/setup.sh, then
source /workspace/adamic-tools/env.sh. nproc=5, CPU quota=4. Timing: Go 0.036s,
Node 0.040s, submodules 0.105s, Markdown 0.118s, clang 0.238s; build 113.942s,
cache readiness 114.327s, done 114.376s.

Final restored observations: lower 12.618s, native 33.219s, JavaScript 8.105s,
IR 0.099s (no matching tests); the corrected uncached oracle rerun passes in
74.643s. go vet on lower/native/JavaScript/IR exits 0. Commands:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/ir ./internal/oracle -run 'Test.*View|TestLazyView|TestSharedArrayContractAdapter' -count=1 -timeout 15m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'Test.*View' -count=1 -timeout 15m
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/ir
```

The first command's oracle portion is the recorded invalid-witness failure, not
an overall green result. The second command is the complete restored corrected
checked-view oracle result. The four compiler package passes from the first
command remain valid: their source did not change during fixture correction.

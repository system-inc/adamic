Built: native and JavaScript mixed-union selector components with four-family Node controls and named wrong-value refusal pins.
Commits: starts from bc87b103; this selector checkpoint is recorded in branch history on codex/views-mixed-unions.
Commands: component oracle, existing checked-view regression, focused native/JavaScript guards, contract tests and vet pass.
Mutants: skip member testing, take the first untested member and omit object transitive matching caught independently in both backends, six semantic kills.
Uncovered: source admission/read dispatch, real native normalized-slot probing, complete reference adapters, brands and all 213 end-to-end pair completions.

Remaining: **213 pairs / 1,111 read sites**. Completed: **0 / 0**. The new
pair-progress.json freezes every accepted census pair, its priority/read count,
other-family obligations and completion evidence. A component does not count as
successful source lowering. The largest two pairs cover 448 reads and require
__String intersection/brand evidence. The first unbranded high-read shape is
EvaluatorResult.value: string|number|undefined, with 61 reads; its full exact
shape appears at 66 reads over four pairs. The selector also covers the general
object/primitive and untagged-object families, subject to complete adapters.

The native runtime API consumes a borrowed snapshot with logical kind and
normalized payload. It does not probe object fields, invent null/undefined tags,
copy readiness code or retain/release the payload. It tests kinds before using
payload members, enforces scalar literals, requires a nonzero reference contract
and pure matcher, and tries later alternatives when one fails. Neither a heap
object tag nor a missing matcher is accepted as structural conformance.
JavaScript implements the same contract. The matched member id/index must remain
on transitive reads after integration. Callback adapters must not invoke getters
or user bodies while testing alternatives.

The source fixtures are .a files under selection-fixtures/. Source Node runs
their actual TypeScript, including the casts. All valid member values match release C, ASan/UBSan C with leak checking, and
JavaScript component executions. Wrong source values remain visible on Node;
selectors stop with exit 70 and an exact field/union/found-category message.
Example:

    adamic: panic: field read failed: view.value matches no member of string | number | undefined; expected string | number | undefined, found boolean

The C oracle uses an independently supplied normalized snapshot and a small
complete two-contract adapter, rather than claiming the compiler lowered the
source fixture. The JavaScript oracle supplies the equivalent adapter and uses
the actual oracle panic runtime. This is deliberate component evidence; shared
field normalization, readiness, representation conversion and compiler emission
are not inferred from it. Reference example snapshots are borrowed fixture
records, not certified tsc allocations. The untagged test checks the second
member after the first structural alternative fails.

run-selection-mutants.py changes one owned implementation at a time and restores
it in finally. Native mutations build valid release C, with no clang failure;
semantic output or the pinned exit/message kills each. JavaScript mutations are
run independently with native restored. The three families of mutations are
exactly skip-member-check, take-first-member and drop-transitive-check. All six
logs are committed. Final component oracle passes in 2.206s; native guard tests pass in .190s and
JavaScript guards in .464s. Both implementations additionally refuse unknown kinds,
missing adapters, missing reference contract ids, null when only undefined is
allowed, and true when only false is allowed.

Validation (test output redirected to logs, never piped):

    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewMixedSelection$' -count=1 -v -timeout 10m
    ADAMIC_GATE_UNCACHED=1 python3 stage3/interface-downcasts/lane4/run-selection-mutants.py
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView' -count=1 -timeout 15m
    go test ./internal/native ./internal/javascript -run '^TestViewMixedUnionUnknownAndUnavailable$' -count=1 -timeout 10m
    go test ./internal/lower -run '^TestMixedUnionContract' -count=1 -timeout 10m
    go test ./internal/javascript -run '^TestView' -count=1 -timeout 10m
    go vet ./internal/native ./internal/javascript ./internal/oracle ./internal/lower
    git diff --check

The checked-view regression passes in 13.550s, contract tests in .344s and
existing JavaScript view tests in .487s. Final component/guard timings are in
logs/selection-oracle.log and selection-unavailable.log. No full gate or new
whole-tsc lowering result is claimed. Shared counts.md remains lane 1 territory;
these component fixtures do not add ordinary source-lowered count rows yet.

Lane 1 remains 9b85eada on the refreshed remote, with no lazy-admission push yet;
lane 2 0b141c26 is already merged. The prior shared cast.go/plan merge conflict
still blocks merging lane 1 as-is. No shared dispatch or protected file was
edited. The exact API/hook handoff is in docs/checked-views-plan.md. No PR is
opened; only the own branch is pushed.

Target for all 213 integrated pairs: **October 16, 2026 at 17:00 MDT**, conditional
on lazy admission, shared normalized-slot probing, and needed brand/schema proof
support landing by October 9. The dependency condition is material: runtime
string tags cannot establish __String phantom brands, and selector prototypes
cannot certify Map/array/callable bodies. This is a conditional planning date,
not an unconditional completion claim. If those prerequisites do not land, the
all-pairs date cannot be met soundly under the current territory/contract rules.

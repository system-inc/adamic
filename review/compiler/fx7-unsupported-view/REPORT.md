Built: all or nothing view admission at creation for fix-forward 5, item 142.  
Commits: refusal c8a7eb5d, fixtures 5483cc0a, current-main merge 9d1f1c3c; the evidence commit is the final delivery HEAD.  
Commands and outputs: full lower, reader guard, affected view oracle, counts refresh and lane checks; final results below.  
Mutant: revert.diff restores p37's wrong native output and fails all eleven new refusal expectations.  
Not covered: implementing the unsupported lowering families or running the whole repository gate.

The diagnostic's What is exactly `view type has an unsupported member: <name>`.
Where is the cast's creation site. The fix text names the member's declared type
and asks for a supported, checked type before creating the view. This is a
conservative refusal, including unread descendants and every union arm. No
partial IR program is returned.

The new lower fixtures are p37-shaped, integration's exact p21, and a
supported-only control. The control runs through `lowersAndAgreesWithNode` and
`lowersAndAgreesWithNodeNative`. A fourth test removes p37's read entirely;
creation still refuses. Existing class-data fixtures contain unread `any`
descendants and now refuse at creation. Existing array boundary fixtures also
refuse at creation rather than waiting for a demanded read. Each new refusal
leaf is a separate top-level parallel test.

Observed under the saved production revert:

| Witness | Source Node | Native | Catcher |
|---|---|---|---|
| p37 | `xx / 0`, exit 0 | `xx / undefined`, exit 0 | Native stdout comparison; creation refusal test also fails |
| exact p21 | `xx`, `3`, `12`, exit 0 | empty stdout, exit -1 | Native comparison; creation refusal test also fails |
| p37 with no read | admitted before the fix | admitted before the fix | Unread-member creation refusal assertion |
| two class-data fixtures | print `true` | lowering admits them before the fix | Both creation refusal assertions |
| five array-union fixtures and the V2 array arm | existing source witnesses | old read-site NotYet, or early empty-array NotYet | Six exact creation refusal assertions |

`revert.diff` is the reverse diff of the isolated refusal commit. It was applied,
run, and reversed back. `revert-probe.go.txt` is the temporary native probe's
source; the compilable temporary file was removed. `revert-mutant-final.log.txt`
records both packages failing for the intended checks. No clang warning was a
catcher. The earlier simplified p21-shaped probe agreed natively; replacing it
with integration's exact p21 exposed the native failure recorded above. Native's
exit -1 is the observed Go process exit value; no specific signal is claimed.

The lazy unsupported-descriptor path is not the only source of unsupported
contracts. `internal/lower/view_lazy.go:44` creates ViewUnknown descriptors and
returns them with nil error at line 52, including after swallowing a strict
contract error at lines 36-40. `view_lazy.go:21` delegates unions to another
builder. `internal/lower/view_unions_dispatch.go:26`, `:39`, and `:74` also mark
contracts Unsupported. These paths can leave unread unsupported descendants.
The common nonunion admission route was `interface_cast.go:311` (`viewSchema`),
which gathered fields without rejecting Unsupported. The other admission route,
`interface_cast.go:337` (`legacyView`), already refused callable and nonscalar
fields. All `view` callers now pass creation validation at
`internal/lower/interface_cast.go:39`, before either route. The inference from
that code is that a failed whole-object descriptor loses its fields and thus
its read checks; the p37 revert is the observed behavioral evidence.

Setup used `export GOPROXY='https://proxy.golang.org|direct'`,
`timeout 600 bash cloud/setup.sh`, and `/workspace/adamic-tools/env.sh`.
Its timing lines report Go 0.072s, Node 0.085s, clang 0.486s, markdown dependencies
4.941s, submodules 25.193s, build ready 460.782s, cache warm 460.968s, and done
461.044s. `nproc` is 5; cgroup cpu.max is 400000/100000. Setup succeeded. Full
lower first reported missing pinned `@types/node` 25.3.3 and reached its 90s
package timeout during cold builds. `timeout 90 npm ci --prefix stage3/api`
installed the repository-pinned dependencies, then full lower passed in 80.506s.

Validation commands, always with output sent to a file:

```text
source /workspace/adamic-tools/env.sh
go test ./internal/lower -run 'TestUnsupportedView|TestSupportedViewMembers' -v -count=1 -timeout 90s
go test ./internal/lower -v -count=1 -timeout 240s
go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s
go test ./internal/oracle -run 'TestCheckedViewUntagged|TestCheckedViewV2' -v -count=1 -timeout 180s
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 180s -args -update-counts
go vet ./internal/lower ./internal/oracle
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

The terminal invocations additionally carry `timeout` limits. The exact revert
selection is recorded in `COMMANDS.txt`. The first counts update exposed the
old partial class-data admission expectation. After migrating those expectations,
the counts update passed in 84.923s. The table adds the supported control's row
(4 allocations, 4 frees, 4 retains, 10 releases, peak 4, regions 0) and removes
the two now-refused class-data rows; all other existing rows are unchanged.

The affected oracle selection passed in 32.724s, including its existing mutants:
wrong-family selection, erased membership fields, disabled tuple identity,
disabled callable producer certification, and five old representation readers.
The null/undefined old readers were caught by UBSan; the three typed-array old
readers by the expected panic. The other four mutated checks were caught by
backend/output assertions. Full mutant results are in `oracle-views-final.log.txt`.

Final delivery results on merged current main cf735d9f:

| Check | Result |
|---|---|
| Full internal/lower | PASS, 121.415s |
| TestCallTargetReaders | PASS, 29.159s |
| Affected view oracle selection | PASS, 32.724s |
| Counts update after merge | PASS, 148.765s, no further counts diff |
| Separate go vet on lower and oracle | exit 0, no diagnostics |
| Integration lane | PASS, 15.7s; bounded lane vet skipped, separate vet passed |

New lower leaves on the final full run: unread member 0.28s, exact p21 callable
member 0.43s, p37 Map member 0.51s, supported control 3.55s. New oracle refusal
leaves: class good 0.47s, class wrong 0.51s, array good 0.16s, array wrong 0.15s,
array nested 0.17s, array empty 0.26s, array mixed 0.28s. The touched V2 array arm
leaf was 0.39s. All are under 60s. The lower control's first cold runtime build
was 38.38s; warmed fixture selection was 0.462s.

The first lane check passed in 0.4s for the isolated production commit. The
fixture lane passed in 15.7s: gofmt and declared tools on six Go files,
t.Parallel on two packages; its vet was skipped after the ten-second window.
A separate bounded go vet covers both changed packages. The repository's narrow
fetch refspec initially left origin/cloud/merge-tree unavailable; explicit remote
tracking refspecs restored that ref and the prescribed lane command then ran.

Automatic approval review rejected a proposed tail replacement in the array
oracle test as risking deletion of later tests. An exact replacement of the
single affected function succeeded and preserved all other file content.
Nothing remains blocked by approval review.

No unsupported family was implemented, and no native emitter or forbidden
orchestration file was edited. The full repository gate and whole native/oracle
packages were not run under the worker instructions. Review-lane pending
sidecars were retained. This unit lands compile-time refusal toward the rule
that every admitted view member has a supported checked lowering.

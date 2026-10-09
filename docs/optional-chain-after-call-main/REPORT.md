Built: optional receivers survive reassignment by a call, while required reads throw catchable TypeErrors, task #ht2nwj5.
Commits: source 611f6e09 rebuilt as 1a981c16 on main c7853d80; newer main 84743fda merged in 25f0ace4, then main 3a7e5216 merged in 6776d163; delivery SHA is reported after push.
Checks: six Node-held fixtures and one flow-edge test pass under 60 seconds each; counts regenerated; integration lane checks pass.
Mutants: optional receiver stop, JavaScript required-read panic, native required-read panic, missing flow edge and missing call propagation are all caught.
Scope: no V3 dependency, fixture weakening, skips, cohere copying, full package runs or full gate.

The original branch has one own commit, 611f6e09. Its optionalReceiver exemption is retained alongside main's acceptsUndefined guard, preserving comparisons, typeof, template observations, coalescing, conditional results, optional property/index reads and receiving slots that admit undefined. Required field reads emit a catchable TypeError; their exception behavior reaches lowering's call propagation, the flow graph and both emitters. Parentheses around optional receivers keep the same behavior. Native uses the error runtime's existing fixed name/message shape.

Cherry-pick conflicts:
- internal/lower/narrowed.go: keep acceptsUndefined and optionalReceiver, and retain both complete helpers. This preserves main's broader observation guards and the original optional-call receiver guard.
- internal/oracle/counts.md: start from main's complete table and regenerate it. No hand merge or obsolete source-branch values.
The merge of newer main has no conflicts.

All six original .a files are byte-for-byte identical to 611f6e09. The old gaps/computed.a witness already lowers on main and prints -1 on Node; its registration is now runnable. No fixture needs compiler/views-v3, so no t.Skip is present. Each runnable fixture is held to source Node, generated JavaScript, release native, native ASan/UBSan and leak checking. Both seed programs finish silently with exit 0; variants preserves absent values, required catches its TypeError, propagation catches the callee error and preserves the old assignment owner. No runtime C or owned compiler assembly file was edited.

Top-level test durations on a single Codex instance with CPU quota 4, nproc=5 (merged-tip, uncached):

| Top-level test | Seconds |
|---|---:|
| TestOptionalAfterCallComputed | 0.80 |
| TestOptionalAfterCallPropagation | 0.83 |
| TestOptionalAfterCallRequired | 0.85 |
| TestOptionalAfterCall79 | 0.85 |
| TestOptionalAfterCallVariants | 0.87 |
| TestOptionalAfterCall193 | 0.45 |
| TestOptionalAfterCallFlowEdges | 0.04 |

Commands and results:
- GOPROXY=https://proxy.golang.org|direct bash cloud/setup.sh (exit 0); source /workspace/adamic-tools/env.sh.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle ./internal/flow -run '^TestOptionalAfterCall' -count=1 -timeout 60s -v (exit 0). Each new top-level function calls t.Parallel first.
- python3 docs/optional-chain-after-call-main/run-mutants.py (exit 0; each of five altered Go tests fails its intended assertion or backend comparison, not compilation).
- GOMEMLIMIT=2GiB go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -parallel=4 -timeout 30m -args -update-counts (exit 0, 76.689s). This pre-existing table-wide test is unchanged; no new or touched test function exceeds 60s.
- git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 - (exit 0, on committed code from repository root).
  Output: lane checks 2.6 s: gofmt and tools on 8 Go files, t.Parallel on 2 test packages; vet 2 packages
This checkout's default refspec did not create the two remote-tracking tool branches, so an explicit fetch created them before the exact lane command ran.

Counts regeneration adds six rows and changes eleven existing rows, enumerated with before/after values in count-changes.json. Required-read failures now construct Error values rather than abort without an error object. Added exceptional control paths conservatively inhibit reuse: 09_tree allocations/frees change 19/19 to 33/33 and retains/releases 28/69 to 66/69. The remaining changed rows are narrowed_reads, narrowed_methods, narrowed_fields, reuse_narrowed, regexp_null_narrowed, element_access_presence, 047cb0d_n_element_plain, 047cb0d_n_element, 047cb0d_n_element_method and 047cb0d_n_arrayindex. Panicking fixtures count where they stop; every new fixture finishes, and allocations equal frees. These count changes were generated, not hand-adjusted.

The original optional-stop mutant is preserved by restoring a narrowing check only when optionalReceiver is true, leaving main's other acceptsUndefined exemptions active. Seed 79 then disagrees with Node. Independent JavaScript/native mutants replace a catchable failed required read with the old invariant panic. Removing lowering's throw propagation changes the propagation fixture's output. The first flow-edge mutant experiment survived the execution fixture, so a separate flow test selects the actual inserted required-read error in the fixture and requires CanThrow. That test fails when its edge is removed. This independent selection uses encoding/json rather than copying CanThrow's reflective walk.

Source fixture bytes are unchanged; native and JavaScript emit the same semantics. No compiler/runtime dependency beyond main was carried. Setup cumulative timing lines: Go 0.022s, Node 0.023s, submodules 0.065s, markdown dependencies 0.074s, clang 0.157s, Go build 40.350s, cache 40.523s, done 40.550s. nproc=5, cgroup CPU quota=4. Go 1.27.1, Node 24.19.0, clang 20.1.8. All commands sent test output to logs. evidence/ holds compressed logs and mutant command/results.

All three requested rows are defended in the complete 30-row bounded matrix.
Each has a distinct production mutant; all other 29 matrix rows pass, with no skips.
Production and tests restored; complete-package and repo-wide uniqueness remain unknown.

# Element-borrow defense

Starting origin/main: 7d113268b1903e4e47289f226e94ec2fb609e15b. Audit base b902a0ccc09e97940571388a1634450da3383559. Read CLAUDE.md, audit REPORT.md and rows-report.json (all fourteen row oracle/verdict records), target test file whole, element_borrow.go whole, relevant borrow.go/IR call-target definitions and neighboring loop, inheritance and chain test bodies. README.md, docs/0.1.md and docs/memory.md are unchanged from the earlier warm session. Fetched audit with the requested full refspec and retained its report, row reports, row list and menu. Fetch encountered HTTP 503 once, then succeeded.

CODE UNDER TEST: native planElementBorrows, borrowable, assignedLocals, changingFunctions, changes, unchanging and borrowElement, plus native C assembly and ownership emission reached by nbody. ORACLE: self-written IR membership and generated-C assertions in these three rows. No Node execution or outside-authority cross-check in the target rows. Neighbor tests may compare to Node. No oracle, test, fixture or harness was changed. The fixed aimed plan is saved in plan.json, written before any mutant outcome.

## Disk and clean baseline

First df /tmp /workspace: /tmp 8.8G total, 258M used, 8.6G free; /workspace 32G total, 17G used, 14G free. Prior-unit /tmp/adamic-gate removal first encountered intentionally unwritable filesystem-test directories. Restored owner permissions only under that named scratch tree, removed it, and recreated TMPDIR mode1777. Second df: /tmp 238M used, 8.6G free; /workspace unchanged. The requested 15G cannot fit on the /tmp mount. No repository or tools removed, no disk-failure baseline.

Warm env.sh worked; setup skipped, nproc=5. npm ci --prefix stage3/api ran before baseline. Native split/job overrides stayed unset. The whole package exceeded its binary budget at 90.048s with no earlier observed test failure. It was stopped by Go's timeout and narrowed before any mutation. Exact skips and incomplete whole run are saved, not credited as passing replay evidence.

Current go test -list reports 282 top-level functions versus audit 280; none vanished. Two added rows, TestRegExpNativeStepLimitBoundary, TestRuntimeKeyKeepsBoundaries, were included in every clean and mutant matrix. Three requested bodies are unchanged from the audit.

Source searches for C(), planElementBorrows() or unchanging() in test files selected 28 top-level rows, including every test in each matching file, not just the targets. Added both new rows: complete 30-row matrix in matrix-rows.json. This is a conservative source-call inventory, not complete transitive native reachability. 252 other top-level tests remain outside this replay and their mutant kills are unknown. The clean bounded baseline passed without skips: 6.914s binary, 9.095s command wall.

## Coverage and semantic differences

Each target ran alone with -coverpkg=./internal/native, -coverprofile=<Name.cover> and anchored -run '^Name$'. All passed. Coverage profiles and exclusive block lists are retained. Counts relative to the named subsumers: nbody has 673 exclusive positive blocks; call-target has 11; documentation claim has 0.

Nbody reaches borrowElement's fallback owner declaration at starting-main element_borrow.go:221, unlike its planner-only subsumer. Its unique assertion demands owner = NULL and no release of the indexed borrowed local for all five planned declarations. D1 changes the existing emitted declaration's constant/ownership option to retain the existing element. Its format string gains the pointer operand needed to express that ownership change; it does not add another emitted statement. This leaves answers correct but adds ownership work.

Call-target coverage reaches the unknown-closure return at line177, absent from the documentation claim. Its mutable callback can switch to an element-writing implementation; the doc fixture's callbacks are bounded and read-only. D2 changes the unknown target return from false to true, treating an unbounded callback as harmless. It violates the closureRead negative assertion without breaking the doc positives.

Documentation claim has no exclusive blocks. Its shared-line semantic difference is benign multi-target virtual dispatch: Reader and Counter implementations both only read array length. The call-target row's virtual alternatives include a writer, so already must refuse borrowing. D3 returns early for virtual calls before checking their actual targets. It conservatively rejects even the benign alternatives, preserves answers, and loses the throughVirtual borrow. This is the permitted return-early mutation, not an arbitrary injected operation.

## Observed catches

| ID | Origin-main file:line | Only failing matrix row | Failing line |
|---|---|---|---|
| D1 | internal/native/element_borrow.go:221 | TestNbodyIndexedElementsBorrow | element_borrow_test.go:48: adamic_local_36_other lost its borrowed element declaration |
| D2 | internal/native/element_borrow.go:177 | TestCallTargetsElementBorrowPlan | element_borrow_test.go:117: closureRead borrowed across a possible element write |
| D3 | internal/native/element_borrow.go:167 | TestDevirtualizeBorrowDocClaim | element_borrow_test.go:142: throughVirtual: length-only targets prevented borrowing |

Every Dn.diff is standalone against the starting main, no selector or harness dependency. All three passed go vet ./internal/native/ and git apply --check after restoration. The matrix is timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run <matrix.regex>, with ADAMIC_GATE_UNCACHED=1 and each ID's own ADAMIC_BUILD_CACHE_DIR=/tmp/defend-elements/cache/Dn. Full actual argv, timings, passed subcases and failures are in results.json; all other 29 top-level rows are listed separately per mutant in passed-rows.json. No narrow run timed out or aborted, and none skipped.

Only one aimed mutant was needed for each row; the three-attempt requirement applies to a not-defended result. No such verdict remains. No survivor, empty-answer probe, test edit, oracle edit or weakening was used. These are not executor twins. D1 and D3 additionally exercise answer-preserving cost regressions, rather than relying solely on different answers.

## Compiled cost witnesses

D1's clean and mutant nbody programs were actually compiled and run with --count. Both print -0.169075164 and -0.169086185. Clean retains/releases are 10/14; D1 15000041/15000045, an increase of 15000031 each. Allocations and frees stay 8/8 and peak stays7. This proves changed work with unchanged output, beyond a formatting-only generated-C difference. The changed C was accepted by the real native build tool.

D3's clean and mutant documentation fixture were also compiled and run with --count. Both print virtual1, virtual1, closure2. Retains/releases rise from 6/13 to 8/15; allocation/free/peak/region counts remain unchanged. The conservative fallback changes ownership work, not these answers. This witness compares clean and mutant executions, not a newly claimed Node oracle. Raw build/run logs, counts and exact commands are retained. D2 changes planner membership, already directly observed by its row; its dangerous callback program was not additionally executed.

The clean three-row run passed after the matrix. Sources were restored in finally blocks after both additional cost witnesses; final git diff --exit-code -- internal succeeds. No production source or test change is committed.

## Brief ambiguities, costs and owner findings

* The 15 GB floor cannot be reached on an 8.8 GB /tmp mount; workspace has14G free. Only identified old-unit scratch was removed. Permission recovery was needed because old oracle fixtures deliberately lock directories. No full-disk failure occurred.
* One audit fetch returned HTTP503 and required retry. Its historical base differs from required current origin/main. Two new tests were included; requested bodies did not move or change.
* Whole-package timeout required a bounded replay. All 30 selected rows complete; outside-set uniqueness is unknown. The source inventory includes whole caller files and newly added rows but does not prove exhaustive indirect native reachability. No other packages were tested.
* Go coverage is measured for native Go planner/emitter code, not C runtime execution. Zero exclusive lines for the doc row did not rule out the semantic difference demonstrated by D3.
* D1 changes an existing emitted declaration and its format operand; D3 adds only the guarded early return. They are ownership-option and return-early changes from the permitted menu, with no fixture-specific filename/name/selector condition.
* All target oracles are self. Internal documentation and benchmark comments are not external authority. The counted executions establish relative changed work, not a general performance threshold or a separate Node agreement claim.
* Nbody's count-five assertion does not name-check the exact set of locals. It then verifies emitted ownership for whatever five the plan selected. The observed D1 defense protects ownership work, not the strength of that identity check.
* The documentation row checks borrowing through virtual/closure targets. It does not assert that native dispatch itself was devirtualized. The call-target row asserts membership for bounded readers, unknown/writing callbacks and virtual writers; it does not execute the miscompiled callback. All three remain defended for the assertions they actually make.
* Existing requested defense branch already contains other rows' evidence. Preserve it with a normal merge, storing this run in borrow-plans/. D1/D2/D3 identifiers are local to this subdirectory; root artifacts retain their own identifiers. No force push or history rewrite.
* Setup was skipped. npm install wall time was not separately instrumented. Go compilation, native artifact builds and test execution are combined in command walls where noted, not falsely reported as isolated compiler-only time.

Timing: whole binary 90.048s; clean bounded command 9.095s. Mutant command walls: D1=13.961s (binary 6.621s), D2=14.359s (binary 7.108s), D3=14.876s (binary 7.499s). Vet total 1.083s. Restored clean command 2.175s. Additional counted build/run command timings are separately saved in nbody-witness-timings.json and doc-witness-timings.json. No full-package completion, repo-wide uniqueness or exhaustive C coverage is claimed.

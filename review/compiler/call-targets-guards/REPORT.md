Guard both surviving call-target mutants with focused IR tests and Node-backed fixture tests.
Keep product code unchanged and use only test files and review evidence.
Run every new leaf separately; each cold invocation passes below 60 seconds, setup included.
Catch audit M7 and M6 with exact assertions, plus an always-throws mutant for the false cases.
Reuse existing counted fixtures; no new fixtures or count rows, and no output-only M6 witness claimed.

Task #h98mfwt adds four bounded leaves under step 79's test grain. The branch starts at main f54b8bd3 and advances without conflicts to 5dced898, which changes only unrelated lint tests. The production call_targets.go bytes match main. No unlanded branch is a dependency.

CallMayThrow calls CallTargets and checks every reachable function's MayThrow. Its direct callers are lower.throwsOut, flow.CanThrow, native.evaluate and native.statement. The IR guard covers throwing and nonthrowing direct calls, a throwing later override, all-nonthrowing overrides, a static throw outside the reachable set, and structural accessor dispatch. Unrelated throwing functions do not spoil a nonthrowing call.

Lowering.call encodes a direct nested sibling as direct + 1 so zero means no direct target. ClosureTargets must subtract one before consulting function effects. Its callers include ClosureMayThrow, ClosureArgumentLayout, the allocation flow graph, field/element borrow planning, method dispatch and comparator analysis. The IR guard pins targets 0, 1 and 3, direct priority over conflicting literal and const evidence, the zero sentinel, and each target's throwing effect. Mutated adjacent indexes stay in range, so M6 is caught by the target assertion rather than a panic.

Both registered source fixtures are held to Node's stdout, stderr and exit code in emitted JavaScript, native release and sanitized native. Successful sanitized runs also pass the existing leak check. statements_small_throw.a prints the caught stored error and finally/rethrow lines. nested_mutual.a prints both parity results and exact recursive call counts. Their existing rows in counts.md remain unchanged. No count regeneration is needed because no fixture was added or changed.

Each cold row used go test <package> -run '^<test>$' -count=1 -json -timeout=90s. XDG_CACHE_HOME was empty per invocation, forcing native runtime libraries and oracle identity setup to be built; ADAMIC_GATE_UNCACHED=1 bypassed result caches. GOCACHE retained the prepared Go/toolchain cache. nproc is 5, cpu.max is 400000 100000, and GOMAXPROCS is 4. Only one measured test ran at a time. Setup succeeded in 36.270 seconds; setup.log has every timing line. Fetching audit evidence initially failed with No space left on device. Removing only the previous bridge unit's disposable product cache freed space and the retry succeeded.

| Test | Cold invocation seconds | Test seconds |
|---|---:|---:|
| TestCallMayThrowUsesReachableTargets | 2.112 | 0.00 |
| TestDirectClosureTargetsUseEncodedIndex | 2.095 | 0.00 |
| TestCallTargetThrowAgreesWithNode | 14.350 | 12.27 |
| TestDirectClosureCallAgreesWithNode | 23.023 | 21.09 |

The baseline and restored runs are in results.json; each new leaf passed before and after the overlays. Mutants use Go -overlay, so no product source was edited.

| Mutant | Catcher | Evidence |
|---|---|---|
| Audit M7: CallMayThrow immediately returns false | TestCallMayThrowUsesReachableTargets | direct throwing: false, want true; later override and accessor also fail |
| Audit M6: direct closure returns call.Direct instead of call.Direct - 1 | TestDirectClosureTargetsUseEncodedIndex | known target [1], want [0]; remaining direct indexes and effects also fail |
| CallMayThrow immediately returns true | TestCallMayThrowUsesReachableTargets | direct nonthrowing: true, want false; safe virtual sets also fail |
| Audit M7 on the stored-error source fixture | TestCallTargetThrowAgreesWithNode | native release and sanitized stdout are only finally; Node prints the caught-error lines too; both mutated binaries exit 0 |

The M7 oracle overlay additionally leaks; the stdout disagreement independently proves the runtime fault. JavaScript retains Node's result. Its restored oracle run passes. M6's exact IR guard is the primary proof; no output-only M6 oracle mutation was required or claimed. The existing sibling-recursion source is checked in both backends, but the unit does not claim a complete search for a source witness whose only M6 symptom is different output.

Commands are preserved in results.json and cold-results.json, with JSONL logs, exact overlay diffs and the two runners. No whole package or full gate was run. Lane checks run after committing; their output is recorded in lane-checks.txt. Only Linux amd64 was measured.

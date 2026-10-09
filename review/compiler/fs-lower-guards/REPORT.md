The four fs guards are now behavioral Node comparisons. See [BEHAVIOR.md](BEHAVIOR.md) for the follow-up evidence; the original IR evidence below is superseded for those rows.

Built five lowering guards and pinned method-value safety refusals for task #5x6wpqm.
Branch compiler/fs-lower-guards starts from main 83f3940e and includes current main before delivery.
Focused lowering tests and uncached Node comparisons in native split modes 0 and 1 passed.
Exact M20, M09, M11, M12 and M13 diffs failed their new guards; unrelated refusal also failed.
No full package runs or full gate; fs guards inspect lowering, without fs native execution.

The audit's TestLibraryMethodValues concern does not match the published test on main or the evidence commit: both already require successful lowering. It remains a success check. The broad err != nil check is TestLibraryMethodValueSafety; this change pins each case's refusal type and reason. An injected unrelated NotYet in the unsupported callback path now fails instead of passing. This is the conservative interpretation of the requested refusal tightening.

The parseInt guard inspects the callback return and requires the two IR arguments to read the element parameter as string and the index parameter as number. It kills M20 at lowering, before native compilation. A separate oracle test writes a temporary .a program, pins Node's stdout to 10,NaN,2, and compares the JavaScript backend, sanitized native, and release native. Native leak checks pass. Both ADAMIC_NATIVE_SPLIT=0 and 1 ran with ADAMIC_GATE_UNCACHED=1. No persistent fixture was added, so fixture counts are unchanged.

The fs guards require openSync string flags to lower to open, rmSync's explicit retryDelay:100 to lower to rm, existsSync to emit boolean exists, and statSync's default throw argument to be true. Production sources were restored after every mutant. Only _test.go and review/ files are committed.

Commands (from repository root after sourcing /workspace/adamic-tools/env.sh):
- GOPROXY='https://proxy.golang.org|direct' timeout 180 bash cloud/setup.sh (passed; setup.log records timing lines; nproc=5, cpu.max=400000 100000).
- Each test below: timeout 120 go test ./internal/lower -run '^NAME$' -count=1 -parallel=4 -timeout=90s -json, output redirected to NAME.json.
- Each audit mutant: git apply review/compiler/fs-lower-guards/Mxx.diff; timeout 120 go test ./internal/lower -run '^NAME$' -count=1 -timeout=90s -json > Mxx.json 2>&1; git apply -R the same diff.
- Unrelated refusal: apply unrelated-refusal.diff; timeout 120 go test ./internal/lower -run '^TestLibraryMethodValueSafety$/^unsupported_callback$' -count=1 -timeout=90s -json > unrelated-refusal.json 2>&1; restore.
- ADAMIC_NATIVE_SPLIT=0 ADAMIC_GATE_UNCACHED=1 GOMAXPROCS=4 timeout 120 go test ./internal/oracle -run '^TestParseIntMapIndexRadixAgreesWithNode$' -count=1 -parallel=4 -timeout=90s -json > oracle-split-0.json 2>&1.
- Same command with ADAMIC_NATIVE_SPLIT=1, output oracle-split-1.json.
- Lane checks after commit: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -. Output recorded in lane-checks.log.

Separate test processes, including process setup, all under 60 seconds (timings.json):
- TestParseIntMapUsesIndexRadix: 2.078s, exit 0.
- TestFSOpenStringFlagsLower: 2.274s, exit 0.
- TestFSRemoveDefaultRetryDelayLowers: 2.372s, exit 0.
- TestFSExistsOperation: 2.278s, exit 0.
- TestFSStatThrowsByDefault: 2.384s, exit 0.
- TestLibraryMethodValues: 2.521s, exit 0.
- TestLibraryMethodValueBoundaries: 2.371s, exit 0.
- TestLibraryMethodValueSafety: 2.480s, exit 0.
- TestParseIntMapIndexRadixAgreesWithNode split=0 uncached: 2.480s, exit 0.
- TestParseIntMapIndexRadixAgreesWithNode split=1 uncached: 2.576s, exit 0.

Mutant kills (exact diffs retained as .diff):
- M20: TestParseIntMapUsesIndexRadix, exit 1.
- M09: TestFSOpenStringFlagsLower, exit 1.
- M11: TestFSRemoveDefaultRetryDelayLowers, exit 1.
- M12: TestFSExistsOperation, exit 1.
- M13: TestFSStatThrowsByDefault, exit 1.
- unrelated-refusal: TestLibraryMethodValueSafety/unsupported_callback, exit 1; expected primitive adapter reason, got unrelated refusal.

Current main 0942c516 merged without conflicts. Post-merge focused lowering tests and both uncached oracle split runs passed (merged-lower.json and merged-oracle-{0,1}.json).
Lane output: lane checks 1.6 s: gofmt and tools on 3 Go files, t.Parallel on 2 test packages; vet 2 packages.

Main advanced again to 68db8ddd with meter test-only guards. Merged without conflicts and repeated the same focused lowering and uncached split-mode oracle checks; all passed. Final evidence: final-lower.json, final-oracle-0.json and final-oracle-1.json. Lane output: lane checks 0.9 s: gofmt and tools on 3 Go files, t.Parallel on 2 test packages; vet 2 packages.

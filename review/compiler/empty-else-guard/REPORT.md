Built: four switch acceptance rows for #qa0ybgb in internal/lower/switch_test.go through lowersAndAgreesWithNode.
Commits: base 3a1c8b5792fb66503290d2d494ce96302bc4386d; audit authority 0bf16ed151905f042fa305efba156a9bcb14f844; delivery SHA is reported with the push.
Commands and outputs: focused baseline passed, 0.207s; restored focus and TestCallTargetReaders results are recorded in their logs; lane checks run before push.
Mutant: authoritative u070 M02 fails all four new rows, JavaScript stdout A versus source Node AB; restored switch.go byte for byte.
Not covered: native execution or a general control-flow sweep; the JavaScript behavioral comparison catches this mutant without a native leg.

rows.md describes each source row. The conditional-break row embeds the exact survivor-switch.a from the audit branch, without editing the witness. The additional rows cover an explicit empty else, a bare empty case label entering an empty block and falling through, and an empty block before a following case's statement that must run. All four print AB on source Node and clean lowered JavaScript. No acceptance row passes on silence, and no IR assertion or golden snapshot was added.

M02.patch is the authority fetched from test-audit/internal-oracle-stage3_front. It changes only switchBodyLeaves' empty-body answer from false to true. run-mutant.py applies it, runs only the four new top-level leaves, requires each intended stdout-comparison failure, and restores the original file in a finally block. M02.log shows all four fail with `JavaScript backend stdout = "A\n", source Node = "AB\n"`, not a compiler error or build-warning failure. The M02 command exits 1; the verifying driver exits 0. No production change is committed. Because the JavaScript backend exposes the error, no local native helper was introduced and this unit does not depend on compiler/agree-native.

M11: a rulings oracle should pin refusal category and reason; pin the `0.1` prefix only if the ruling explicitly makes the diagnostic version part of the contract.

Setup command: `export GOPROXY='https://proxy.golang.org|direct'; timeout 300 bash cloud/setup.sh`, then source /workspace/adamic-tools/env.sh for every test shell. Setup passed. Timing lines: Go ready 0.017s, Node ready 0.021s, submodules ready 0.053s, Markdown dependencies skipped step-duration 0.007s and ready 0.070s, clang ready 0.141s, Go build ready 32.420s, test binaries deferred 32.553s, build cache warm 32.554s, done 32.581s. Versions: Go 1.27.1, Node 24.19.0, clang 20.1.8. nproc is 5 and cpu.max is 400000 100000, a four-CPU quota. setup.log preserves the complete output.

Exact commands, with all test output sent directly to logs:

```sh
timeout 120 go test ./internal/lower -run '^TestSwitch(ConditionalBreak|EmptyElse|EmptyCase|EmptyBlock)FallsThrough$' -count=1 -timeout 90s -v
timeout 180 python3 review/compiler/empty-else-guard/run-mutant.py
timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
```

The focused Go command ran for the baseline, M02, and restored checks. The driver gives its Go invocation an outer timeout 120 and -timeout 90s. Baseline leaf seconds: conditional break 0.18, explicit empty else 0.19, empty block 0.20, empty case 0.20. M02 leaf seconds: conditional break 0.17, explicit empty else 0.17, empty case 0.17, empty block 0.18. Each new leaf is under 60 seconds, setup included.

TestCallTargetReaders checks the new tests; they read none of Call.Function, CallClosure.Closure, or ArraySort.Comparator, so no allowlist entry is required. Lane checks use the prescribed integration script with pipefail and an outer timeout 180, after committing the tests. No whole package or full gate is run. No new internal/oracle fixture is added; the .a file under review is audit evidence, not an oracle fixture, so counts.md needs no new row.

This unit adds a behavioral guard for the missing fallthrough edge toward the compiler agreement work named by #qa0ybgb. The brief supplies no numbered roadmap step, so no number is inferred.

Final focused clean result: pass, 0.282s. Restored leaf seconds: conditional break 0.24, explicit empty else 0.22, empty case 0.24, empty block 0.27. TestCallTargetReaders passed in 12.786s. Logs: restored.log and call-target-readers.log.

# no-constant-condition

Based on `origin/area/stage1-lint` at `6bf7bcec73a04b85900c5a69a257b3cf34b35a40`.
Only this rule directory changes. The all-origin exact-name search returned no
existing port. The unmodified Go rule at cohere pin
`7945d102a6c18dd36adf9114a758ce646e8b2359` decides the output.

The port carries the boolean/value distinction through arrays, unary operators,
templates and logical identity; it preserves default, string and boolean loop
options. It imports the existing `numericLiteralSign` shared helper.
The typed oracle adapter restores the captured zero enum to its typed default:
Go capture serializes an empty `CheckLoops` as `""`, which its configuration
custom decoder refuses. Explicit values still use that upstream decoder.
The first adapter implementation failed the selected comparison for this default;
[evidence/adapter-default-refused.log](evidence/adapter-default-refused.log.gz) records
that failure and [evidence/selected.log](evidence/selected.log.gz) records the fix.

`testdata/shadowed-globals-generator.ts.txt` names intentional Go behavior:
shadowed `undefined` and `Boolean` still count as globals, and the generator's
`while (1) { yield 1; }` reports. The port matches these false positives.
`testdata/boolean-position.ts.txt` also compares omitted sparse-array nodes in
value position, where the Go switch does not treat them as constants despite
its nil-hole comment. It matches Go here too.

Validation commands, with `/workspace/adamic-tools/env.sh` sourced:

```
go run ./cmd/lint-registry
python3 stage1/cohere/lint/rules/no-constant-condition/testdata/check.py
python3 stage1/cohere/lint/rules/no-constant-condition/testdata/check.py mutant
python3 stage1/cohere/lint/rules/no-constant-condition/testdata/gate.py
```

Registry generation passed. Selected parity passed: **64 captured upstream cases**,
six owned witnesses, and inherited selected/all-rule rows, **530,535 identical
bytes** across Go cohere, source Node, emitted JavaScript and ASan/UBSan native.
See [selected.log](evidence/selected.log.gz); process wall time 169.834 seconds.
Captured input transport covers typed defaults and every upstream loop option.
Malformed configuration decoder assertions are upstream Go tests, not captured
finding cases; general configuration validation is not claimed by this port.

The mutant makes an array nonconstant in boolean position. It compiled and ran;
Node and emitted JavaScript disagreed with Go. Sanitized native equaled mutated
Node (**526,211 bytes**), so it also disagreed with Go. See
[mutant.log:33](evidence/mutant.log.gz). The mutant subtest passed in 61.52 seconds.

The first cold selected and native-mutant invocations each exceeded their original
20-minute test limit during sanitizer compilation. Their logs remain at
`/tmp/s13-constant-selected.log` and `/tmp/s13-constant-mutant.log`. The unchanged
assertions passed after increasing the rule-local runner limit to 60 minutes.
These timeouts are not semantic mutant catches or language refusals.

Setup passed: ordinary setup 2.985 seconds; WASI setup 346.453 seconds,
including Go cache warming. Exact timing lines are in
[setup.log](evidence/setup.log.gz) and [wasi-setup.log](evidence/wasi-setup.log.gz).
Go 1.27.1, clang 20.1.8, Node 24.19.0, WASI SDK 27; `nproc=5`, cgroup quota 4 CPUs.

Exact completed text logs are archived as deterministic `.log.gz` files; the
plain logs remain beside them in the workspace and under `/tmp`.

The full package result is recorded by the gate runner in
`evidence/whole-package-metrics.json` and `evidence/whole-package.jsonl`.
The runner verifies a clean TypeScript checkout at
`050880ce59e30b356b686bd3144efe24f875ebc8`, supplies WASI_SYSROOT,
enables ADAMIC_LINT_BENCH, and creates one fresh directory for both profile inputs.
A pre-existing `TestCheckerBridgeRefusalPending` skip awaits TSGoError support;
this unit cannot remove that blocker within its rule-directory territory.

The complete package invocation did **not** pass: it reached Go's 60-minute
limit with **80 completed test/subtest passes, 2 failures and 1 skip**; **70
checks remained unfinished**. The JSON metrics enumerate every unfinished check.
Process wall time was **3633.160 seconds**; `nproc=5`; load before
`1.14 2.66 4.89`, after `6.46 6.64 5.84`. All required inputs were supplied;
the one skip is the pre-existing `TestCheckerBridgeRefusalPending`, not a missing
input. Do not interpret completed event counts as full-suite success.

`TestChildCPUWaitGuard` failed its unchanged CPU assertion (1.244472 seconds
of child CPU against a 1-second budget). Its isolated retry passed with
522.026 milliseconds CPU; [cpu-wait-retry.log](evidence/cpu-wait-retry.log.gz)
records the observation. No shared guard code was changed.

`TestCompilerAndStage1Agree` initially refused the dirty new `.a` sources:
marking them for addition during review did not satisfy its committed-source
requirement. A local source checkpoint enabled the separate clean-corpus retry.
The full invocation's original failure and timeout remain recorded, not erased
by retries. The full run's owned mutant was still queued at timeout; the separate
owned semantic-mutant run above completed successfully on all three backends.

The clean `TestCompilerAndStage1Agree` retry passed: **814 inputs**, including
77 pinned TypeScript compiler files, 736 committed stage-1 sources and the explicit
generated registry. Go, Node, emitted JavaScript and sanitized native produced
**30,987,002 identical bytes**; process wall time **438.500 seconds**. See
[compiler-corpus-retry.log](evidence/compiler-corpus-retry.log.gz). The command was
`go test ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -v -timeout=60m`,
with the same corpus/WASI/profile inputs as the full run. The CPU guard retry used
`go test ./stage1/cohere/lint -run '^TestChildCPUWaitGuard$' -count=1 -v -timeout=60m`.

No required external helper or language feature blocked this port. The complete
package remains unproven because of its timeout and the named bridge skip;
selected owned parity, semantic mutant and the inherited compiler corpus passed.
The full runner must be invoked after committing new `.a` sources so the inherited
corpus's clean-source prerequisite can pass. Its source checkpoint was local and
was never pushed as a partial unit.

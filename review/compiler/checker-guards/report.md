Built: nine top-level parallel checker guards for task #7cawc2x, including all three audit survivors and six cheap direct calls.
Commits: d836f7c20bd72dca0747a9ce310e2a4ffa77c612 adds only tests; 56664764a58d17a27a41033c415781b30d1eac41 merges current main's test-only landing; the delivery commit records evidence.
Commands and outputs: focused control PASS (0.060s), nine mutant failures with nine restored passes, final focused coverage PASS (0.057s), lane checks PASS.
Mutants: M1, M4 and M5 fail their dedicated guards; six supplementary mutations fail the six direct-call guards; all are restored before delivery.
Not covered: the whole package, full gate, archive/C ABI, union or generic TypeParts, diagnostic localization arguments and other property declaration variants.

This holds the pinned checker bridge's framing and metadata contract toward #jy0v0am. Production code is unchanged. The branch starts on main e2492670 and merges main f91994f0, whose only changes are JavaScript tests and review evidence. Every delivery commit is test-only. No unlanded branch was merged.

The three .diff files are byte-for-byte copies from audit commit 53066332. audit-diff-sha256.json records their digests. The driver applies each diff with git apply, runs only its named guard, reverses the diff in a finally block, then requires that same guard to pass. It refuses compilation failures, panics and timeouts as mutant evidence. mutants.json and the paired compressed logs record every command and exit code.

| Audit mutant | Guard and observed failure |
| --- | --- |
| M1: U+FFFF counted as two UTF-16 units | TestGuardSymbolOriginFFFF: the real symbol-origin path fails the independent Go UTF-16 decoder with invalid frame length. Clean framing decodes exactly schema, question and the complete U+FFFF .a path without overrun. |
| M4: true encoded as 0 | TestGuardStrictNullChecks: raw-shape strict-null-checks metadata is 0 under the unchanged strict=true compiler options; the expected value is 1. |
| M5: list count increased by one | TestGuardSingleTypeRoot: raw-shape root count is 2 for the single number root; the expected count is 1. |

All six formerly unreached functions admit cheap direct calls; none remains deferred. Their concrete witnesses and proof mutations are:

| Function | Known answer and supplementary mutant |
| --- | --- |
| writeSymbolOrigin | One interface named Thing, with its actual source file and false declaration/default-library flags; changing its name to mutant is caught. |
| writePropertyInfo | A method with first parameter first: string, no initializer; corrupting StringKeyword annotation is caught. |
| diagnosticText | Two compiler diagnostic messages become first 世界; second; changing the separator is caught. |
| Query | The binary expression 1 + 2 has the pinned compiler's BinaryExpression kind and number type; changing the type to mutant is caught. |
| TypeParts | The exact number identifier has the pinned compiler's Number flags and a six-unit number name; changing the name to mutant is caught. |
| scopeTables | Two real binder symbols are alphabetically ordered alpha, zebra with fixed declaration byte spans; reversing their order is caught. |

Expected AST kinds and type flags come from the pinned TypeScript compiler shims; UTF-16 framing uses Go's Unicode encoder independently of fields.text. The sources are small temporary .a inputs. The strict options are the audit's strict/ES2022 configuration, with an explicit files list because config globs do not discover .a inputs. No registered oracle fixture or counts entry changes, so counts regeneration is not applicable.

These tests use in-process checker calls only. They do not invoke build commands, need archives or binaries, or introduce a product consumer. Therefore no TestProduct_* fetch is needed. Existing bridge product recipes are unchanged. No cohere code was copied.

Commands source /workspace/adamic-tools/env.sh and export GOPROXY='https://proxy.golang.org|direct'. Test output goes directly to files.

```sh
timeout 240s bash cloud/setup.sh > /tmp/checker-guards-setup.log 2>&1
timeout 150s go test ./bridge/tsgo/checker -run '^TestGuard' -count=1 -v -timeout 90s > review/compiler/checker-guards/control.log 2>&1
timeout 600s python3 review/compiler/checker-guards/run-mutants.py > review/compiler/checker-guards/mutant-driver.log 2>&1
timeout 150s go test ./bridge/tsgo/checker -run '^TestGuard' -count=1 -v -timeout 90s -coverprofile=review/compiler/checker-guards/guards-coverage.out > review/compiler/checker-guards/restored-final.log 2>&1
timeout 30s go tool cover -func=review/compiler/checker-guards/guards-coverage.out > review/compiler/checker-guards/functions.txt
timeout 60s git fetch -q origin main devtools/fast-gate cloud/merge-tree
timeout 150s bash -o pipefail -c 'git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'
```

Each mutant and restored command is go test ./bridge/tsgo/checker -run '^<named guard>$' -count=1 -v -timeout 90s, with a 150-second subprocess limit. git apply and reverse have 30-second limits. Each new leaf's measured seconds are in test-seconds.tsv; the maximum is 0.04 seconds, including temporary files and Open. The final nine-guard run reaches 54.5% of package statements. That is focused-run coverage, not a rerun of the audit's whole-package coverage. functions.txt confirms all six named functions are reached.

Setup passed: go 0.029s, node 0.030s, submodules 0.077s, markdown dependency step 0.008s and ready 0.086s, clang 0.191s, go build 38.059s, test binaries deferred 38.207s, cache warm 38.208s, total 38.246s. nproc=5, cpu.max=400000 100000. Go 1.27.1, Node 24.19.0 and clang 20.1.8. setup.log.gz retains all timing lines.

Two preliminary test attempts are retained: implicit config globs found no .a input, corrected by an explicit file list; an identifier Query expected a token, but the pinned position lookup skips tokens, so the final witness selects a binary expression. Neither changed production behavior. The final controls and mutation proofs use the corrected cases.

Lane output on the committed test change: lane checks 0.9 s: gofmt and tools on 1 Go files, t.Parallel on 1 test packages; vet 1 packages. git diff --check origin/main -- bridge/tsgo/checker/guards_test.go passes. Checking the added .diff evidence as ordinary text reports space-before-tab and trailing-space warnings from standard unified-diff context lines; the exact audit diffs are preserved unchanged. Final lane checks run again after the evidence commit and their output accompanies the pushed SHA.

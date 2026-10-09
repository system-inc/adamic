# Recovery landing on the current lint area

The branch merges `db2ecc00447f9ebe8adecb190f71ac222e5db860` without rebasing. Merge commit: `5e5b6cd6b3a5c086f696671fc7023cd2158a2c20`. The real lint contract change is `df3b9b4a3f41abf50fd2d48b8c5593b0c8b24dae`; the local recovery fast-path change is `595ea27d625a099214013ab6da30ef50e30d364b`.

The numeric node-kind minting sites, nodes.ts, JSX implementation and scanner implementation were not edited. There are no local edits to the forbidden compiler files. The earlier queued parity work is preserved locally on `codex/parser-queued-local`; the landing history contains ordinary revert commits, not a rewrite. JSX, parity-2's patches and the four later parser gaps remain the next unit.

## The lint contract

The harness introduced by `a46053af4d4a018c25bf34a5d4c0b7e035fdb8fc` compares recovered findings through the `recovery` mode. The four method-signature sources still carried the older `unsupported-recovery` marker, which requires refusal. Recovery parses them, so that marker was stale. The only lint edit changes that one marker to `recovery`.

The four original sources are now also raw, small parser fixtures:

- `testdata/recovery/interface-eof.ts.txt`
- `testdata/recovery/method-eof.ts.txt`
- `testdata/recovery/generic-method-open.ts.txt`
- `testdata/recovery/generic-method-close.ts.txt`

`TestRulesAgree` is run unmodified after this one-line commit. No Go test overlay replaces the harness. Its normal external-Go protocol adapter is unchanged. All four recovered cases enter the same byte comparison as every captured upstream case, across Go, source Node, emitted JavaScript on Node and sanitized native. The final unmodified TestRulesAgree passes in 129.85 seconds and reports 3,442 unique upstream cases and 13,650,414 identical output bytes. The selected lint package run, including both live mutants, passes in 296.220 seconds.

## Incomplete-input comparison

The area baseline is a detached worktree at exactly `db2ecc00447f9ebe8adecb190f71ac222e5db860`. That area does not contain the landing's incomplete-input test. Only the landing Go test files and external-Go protocol adapter were copied to it; its TypeScript parser source was unchanged. The shared cohere checkout supplies the same pinned external parser in both runs.

The baseline fails after four inputs. Case 2, `00002-ts.moduleSpecifiers.ts-cut-1.ts`, is a stranded `export`: Node and native refuse with exit 70. Case 3 also refuses. Case 4 misses Go's diagnostic 1109, `Expression expected.`. See `validation/unit1/area-incomplete.log`. Thus the incomplete-input gate was already failing without recovery; this is not a newly green baseline broken by the landing.

The landing compares 1,782 inputs before its unchanged 10-second subprocess deadline stops native on `01778-checker.ts-cut-189457.ts`. See `validation/unit1/incomplete.log`. Its process parses **and prints** the canonical tree; it is not a parse-only deadline.

A direct diagnostic run with a 60-second bound completes normally on that exact retained input. Go takes 0.400 seconds, Node 1.684 seconds and sanitized native 4.306 seconds. All produce exactly 8,519,744 bytes with SHA-256 `5bf4575f73b74e8d32838d032777215b6ef9cf0aa80ddb52bc36166dff0269b2`. This input does not loop. The failed corpus run overlapped instruction profiling; scheduling contention is an inference, not a demonstrated sole cause of the deadline. The deadline and corpus selection remain unchanged. This report does **not** claim all 22,497 scheduled malformed inputs pass.

The large retained input remains in the scratch artifact directory named by the log. It is not copied into testdata or pasted here. Reduced recovery witnesses stay in testdata.

## Clean-corpus instruction cost

All controls use the **same current merged compiler**, toolchain, manifest and native flags. The TypeScript source pin is `050880ce59e30b356b686bd3144efe24f875ebc8` (6.0.3). The base parser snapshot is the current area; before is the original landing `daf277b014a724b7ca1191b2bbb5efc1a10335ea`; after is the local fast-path change. Compiler-source changes in the newer area mean that the old 7.41G/8.92G totals must not be mixed with these freshly rebuilt controls.

The native release flags are `-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2`. There are no sanitizers in the instruction measurement.

Callgrind flags are `--tool=callgrind --callgrind-out-file=<artifact>`. Parser flags are `--manifest <77-file manifest> --whole --count`. This includes startup, scanning, parsing, allocation and counting, and excludes canonical tree printing. Separate tree runs compare all 44,766,682 bytes against typescript-go. Every control has 887,803 nodes; tree SHA-256 is `8ae015600498b915cc25abab82730299451ae990b50478980d5a3bc465801bfe`.

The fresh baseline is 5,762,180,628 instructions; the original landing is 7,216,199,207. The final optimized measurement is 6,343,677,798. Rebuilt measurements and profiles are in `validation/unit1/final-measurements.json` and its accompanying instruction log. The optimization removes 872,521,409 instructions, 60% of the added cost. The remaining 581,497,170 instructions are still 10.1% above the baseline; this is a reduction, not a claim of zero overhead.

The three largest **added self-cost** sinks before optimization were:

| Function | Area base | Original landing | Final after |
| --- | ---: | ---: | ---: |
| `adamic_array_index_of` | 60,858 | 259,899,741 | 11,888,472 |
| `release_last` | 182,772,684 | 348,520,310 | 220,208,271 |
| `adamic_string_equal` | 345,204,698 | 494,725,132 | 453,688,116 |

The complete function comparison is `validation/unit1/function-costs.json`. Inclusive recursive costs can count the same descent more than once, so the ranked growth table uses self costs. The largest total self costs after optimization are scanner code, scanner scan and string equality.

Recovery's anticipatory token classification allocated and searched string arrays on clean input. Its lexical-error hook copied an empty suffix of scanner errors for each token. Fixed token inventories now use switches, identifier classification takes the ordinary identifier fast path, and the scanner-error hook runs only when scanning adds an error. The remaining indexed error loop preserves each scanner diagnostic. A scanner-only identifier lookahead avoids copying unrelated parser state.

## Check commands

Each command sources `/workspace/adamic-tools/env.sh`. Test stdout and stderr go directly to the corresponding logs under `validation/unit1/`.

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/typescript/parser -skip '^(TestIncompleteCompilerAgrees|TestPerformance|TestWholePerformance)$' -count=1 -v -timeout=45m
go test ./stage1/typescript/parser -run 'TestMethodRecoveryAgrees|TestRecoveryMutants/(scanner-error-hook|reserved-token-switch)' -count=1 -v -timeout=20m
go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestLegacyMutants|TestCountGuardMutant)$' -count=1 -v -parallel=1 -timeout=60m
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_RECOVERY_ARTIFACTS=/workspace/scratch/parser-unit1-inputs go test ./stage1/typescript/parser -run '^TestIncompleteCompilerAgrees$' -count=1 -v -timeout=2h
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_RECOVERY_ARTIFACTS=/workspace/scratch/parser-unit1-area-inputs go -C /workspace/scratch/parser-unit1-area test ./stage1/typescript/parser -run '^TestIncompleteCompilerAgrees$' -count=1 -v -timeout=2h
go vet ./stage1/typescript/parser ./stage1/cohere/lint ./stage1/cohere/lint/registry
```

The parser package run passes in 893.423 seconds. Its whole-compiler check compares 44,766,682 bytes and its expression-only compiler check compares 28,836,875 bytes over all 77 files. The committed fixture-reader additions and two new mutants pass in the focused final run in 101.023 seconds. Cohere checks all ten parser files with the repository settings and isolated parser tsconfig, `--no-fix --no-cache`; it reports zero findings across 276 rules. Profiling replay is documented in `validation/unit1/README.md`.

## Checks and mutants

All test output is written directly to logs. Setup uses `GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh`, followed by `source /workspace/adamic-tools/env.sh`. Setup succeeds in 82.166 seconds; `nproc` is 5. Its complete timing lines are in `validation/unit1/setup.log`.

The instruction guard is proven by a real extra parse of each compiler file. Its printed trees and node counts remain identical, but its 12,164,378,266 instructions exceed the original landing's 7,216,199,207 budget. A separate accounting mutant increases the Callgrind summary by one; the reader rejects the mismatch with the sum of self costs.

The new `scanner-error-hook` mutant suppresses scanning diagnostics and is caught by the Go/Node/native diagnostic comparison on `testdata/recovery/unterminated-string.ts.txt`. The `reserved-token-switch` mutant treats `return` as an identifier and is caught by the same comparison on `testdata/recovery/reserved-binding.ts.txt`. Both finish normally; compiler errors do not count as a kill.

The six preserved recovery mutants are octal suggestion, expected-token diagnostic code, stranded-export diagnostic, generic child removal, EOF loop and speculative roots. The five byte mutants are caught by the external-Go comparison; the EOF loop is caught by the unchanged two-second process bound. Whole-tree, expression, node-count and JSX mutants are recorded individually in the parser-package log, including their Node/native catches. The lint mutation evidence is recorded in its final package log. `validation/unit1/MUTANTS.md` lists every one of the 34 mutants/checks and what caught it.

The full repository gate, throughput benchmarks and all 22,497 malformed inputs are not claimed green. Validation is scoped to the parser and lint packages, the 77 compiler files, live mutants, instruction guard and cohere/vet checks. The next unit's real JSX corpus, parity-2 patches and four parser gaps are deferred as requested.

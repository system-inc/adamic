Built: cf9f33c parser recovery landed on d65a8f9 with the unified lint harness and area JSX preserved.\
Commits: merge cd958cf109765b041f95a2041e99ee8eed85b647; diagnostic adaptation b7660e5a1cb686be00088a02fc205bfe3fa8e7d0.\
Commands and outputs: 77 whole files match Go; positive lint replay passes 1,987 captured cases; 8,923,007,866 instructions fits the 8,926,683,234 budget.\
Mutants: 33 caught, comprising 28 parser, two lint, one compiler byte, one extra parse and one instruction-accounting mutation.\
Not covered: full repository gate, successful 22,497-input malformed corpus gate, best-of-five speed benchmarks or every internal Go AST field.\

## Landing

Branch `codex/parser-recovery-land` starts at `origin/area/stage1-lint`, d65a8f931c98655936ae04c6899f38f14862b73e, and merges `origin/codex/parser-recovery`, cf9f33c3b02319e865bf09fc5a0cc914d8600d4b. The common base is 5d4c8012a0877094134e6c6bac367ff68f9313e8. The integration handoff is this report's final branch tip, printed in the response and verified on the remote. Only `codex/parser-recovery-land` is pushed for this landing.

All ten reported conflicts were reproduced. Every tracked file in `stage1/cohere/lint` is byte-for-byte the area version, including nonconflicting files that the recovery branch had changed. The parser's recovery test now builds the unified registry and every descriptor's real Go oracle adapter through an overlay. Parser and oracle conflict resolution retains both recovery and the area's JSX behavior, including `--jsx-kinds` and `--token-spans`.

Small token predicates moved into the existing grammar module to keep parser.ts at 1,999 lines under Cohere's limit. Lint replay exposed unimplemented scanner messages requiring literal arguments. Parser-side lexical handling now matches Go for legacy octal literals (including its signed-64-bit saturation), octal escapes and forbidden `\8` / `\9` escapes. Seven direct Go-backed probes cover those messages, negative literals, whitespace, saturation and a Unicode prefix. The new octal-message mutant must compile and terminate normally before its diagnostic mismatch counts.

The requested area ref contains textual `kind: string`, not the announced numeric `SyntaxKind` field. `nodes.ts`, the central `Parser.make` / `new ParseNode` minting body, the direct PrivateIdentifier constructor, `jsx.ts`, the scanner and all `internal/` files remain the area versions. Recovery changes nearby `make(...)` call sites and reads textual node kinds; the numeric-kind owner should preserve the added metadata at those sites during a later merge. No numeric field or generated kind table was edited. [The preservation check](validation/landing/preserved.log) exits zero.

The earlier parity work is separately recorded on `codex/parser-parity-2` at 0ddd99af161fb9759f2efb9e9aa557bf694128de. This landing uses the requested recovery tip rather than importing that branch's broader malformed-input fixes.

## Toolchain

`bash cloud/setup.sh` succeeded. Each subsequent command sourced `/workspace/adamic-tools/env.sh`. `nproc` returned 5; the cgroup quota is four cores. Go is 1.27.1, clang 20.1.8 and Node 24.19.0. [Raw setup output](validation/landing/setup.log):

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (26s)
setup: build cache warm (766s)
setup: done in 766s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The compiler corpus is Microsoft/TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db; its external typescript-go checkout is 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Go oracle entry points use build overlays; the external parser and AST are unchanged.

Valgrind was absent. System apt failed because uid 1000 cannot write its configuration and list directories; local indices had no candidate and the configured snapshot endpoint returned HTTP 403. Downloading Debian's valgrind_3.24.0-3_amd64.deb from deb.debian.org and extracting it under `/workspace/scratch/parser-land-valgrind` worked. `VALGRIND_LIB` points to its `usr/libexec/valgrind`; the binary is 3.24.0. [Smoke output](validation/landing/valgrind-workaround.log) records a successful `/bin/true` run.

## Checks and exact flags

All test output went directly to log files. `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3` was set for corpus commands. The [coverage audit](validation/landing/coverage.json) records all 25 ordinary parser test functions passing; the incomplete corpus failed and two optional speed benchmarks were skipped. The scoped lexical style cleanup occurred during the ordinary suite; focused recovery, the full positive lint replay and final native instruction/tree checks were restarted afterward.

| Command | Observed result | Evidence |
| --- | --- | --- |
| `go test ./stage1/typescript/parser -skip 'TestIncompleteCompilerAgrees\|TestMethodRecoveryAgrees\|TestRecoveryMutants\|TestRecoveredLintCasesAgree' -count=1 -v -timeout=90m` | PASS 828.479s; 77 whole files, 44,766,682 identical tree bytes; 77 expression files, 28,836,875 bytes; 348 JSX inputs, all 13 Go JSX kinds; 42 type-node kinds; 68 generated whole files; all 22 remaining parser mutants caught | [parser.log](validation/landing/parser.log) |
| `go test ./stage1/typescript/parser -run '^(TestMethodRecoveryAgrees\|TestRecoveryMutants\|TestRecoveredLintCasesAgree)$' -count=1 -v -timeout=30m` | PASS 262.545s; focused recovery probes, all six recovery mutants and eight real lint source/style combinations | [recovery.log](validation/landing/recovery.log) |
| `go test ./stage1/cohere/lint -run '^(TestRulesAgree\|TestCompilerAndStage1Agree\|TestOwnedWitnesses\|TestRegistrationMutant\|TestCountGuardMutant\|TestJsxLintTrees)$' -count=1 -v -timeout=60m` | FAIL 499.934s only on the preserved legacy refusal assertion; compiler/stage1, owned witnesses, JSX and both mutants PASS | [lint-oracle.log](validation/landing/lint-oracle.log) |
| `go test -overlay=/workspace/scratch/parser-land-counts/lint-recovery-overlay.json ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=45m` | PASS 115.116s; 1,987 captured cases; Go, Node, emitted JavaScript and sanitized native identical over 13,042,211 bytes | [lint-recovered.log](validation/landing/lint-recovered.log) |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=15m` | PASS 1.665s; native and Node each report one cache miss, zero hits; injected byte mismatch caught | [compiler-byte-mutant.log](validation/landing/compiler-byte-mutant.log) |
| `/workspace/scratch/cohere -directory /workspace/parser-recovery-land -tsconfig /workspace/scratch/parser-land-counts/parser-tsconfig.json -no-cache -no-fix stage1/typescript/parser/*.ts` | PASS; 276 rules, 10 checked, all 10 Adamic-ready; only the prelude excluded from lint | [cohere.log](validation/landing/cohere.log) |
| `go vet ./stage1/typescript/parser ./stage1/cohere/lint ./stage1/cohere/lint/registry` and `gofmt -d stage1/typescript/parser/recovery_test.go` | PASS, empty output | [vet.log](validation/landing/vet.log), [gofmt.log](validation/landing/gofmt.log) |

The preserved lint harness labels four recovered method-signature sources `unsupported-recovery` and asserts that the port refuses them. Recovery now accepts them, so that negative assertion fails. The positive replay changes only that label to `recovery` in a scratch test overlay, retains all original captured inputs and options, and compares findings, ranges and proposed repairs on all four runtimes. The [one-line overlay diff](validation/landing/lint-recovery-contract.diff) is reviewable. It does not make the unchanged lint test green. The harness also passed its 351 compiler/stage1 files (20,698,926 bytes), owned witnesses (44,914 bytes), and 54 captured JSX sources (45,527 tree bytes).

## Instruction budget

All branch snapshots were built with the same area compiler and runtime. The release command was `go run ./cmd/adamic build <snapshot>/stage1/typescript/parser/main.ts -o <binary>`. Native flags were:

```text
-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign
-ffp-contract=off -fno-optimize-sibling-calls -O2
```

No sanitizer, debug or allocation-count instrumentation was enabled for instruction accounting. The parser driver's `--count` counts AST nodes; it is distinct from the compiler's allocation instrumentation flag. The exact Callgrind command was:

```sh
VALGRIND_LIB=/workspace/scratch/parser-land-valgrind/usr/libexec/valgrind \
/workspace/scratch/parser-land-valgrind/usr/bin/valgrind --tool=callgrind \
  --callgrind-out-file=<label>.callgrind <binary> \
  --manifest /workspace/scratch/parser-land-counts/compiler.manifest --whole --count
```

The measured `Ir` includes startup, manifest reads, scanning, parsing, allocation and node-count traversal over all 77 files. Tree printing is measured separately for byte parity. Every run counts 887,803 nodes, matching the fresh external Go oracle.

| Snapshot | Instructions | Nodes |
| --- | ---: | ---: |
| Common base 5d4c801 | 7,403,103,936 | 887,803 |
| Area d65a8f9 | 7,412,969,065 | 887,803 |
| Recovery cf9f33c | 8,916,818,105 | 887,803 |
| Final landing b7660e5 | 8,923,007,866 | 887,803 |
| Actual extra-parse mutant | 17,295,608,569 | 887,803 |

Recovery adds 1,513,714,169 instructions over the common base. Landing adds 1,510,038,801 over the area. The allowed total is 8,926,683,234; landing is 3,675,368 below it. The count does not rise beyond what recovery adds.

The actual mutant creates a second `Parser` and calls `file()` before printing or counting the original parser. It finishes normally with the same 887,803 count and all 44,766,682 tree bytes identical to Go (SHA-256 8ae015600498b915cc25abab82730299451ae990b50478980d5a3bc465801bfe). Only the instruction budget catches it, at an excess of 8,368,925,335. The summary-only accounting mutant increments `summary: Ir` by one without changing any recorded costs; the shared profile reader rejects it with `callgrind self costs do not sum to summary`.

[Final measurements](validation/landing/final-measurements.json), [raw command output](validation/landing/instruction-final.log), [lossless profile hashes](validation/landing/profile-hashes.json), [accounting mutant](validation/landing/instruction-accounting-mutant.log), and [measurement replay](validation/landing/measure.py) retain the evidence. Six compressed profiles preserve the raw cost records, including the corrupted-accounting control. Self costs were checked against each real profile's summary.

## Every mutant

Ordinary mutants compile and finish normally; a compile error or unrelated crash does not count. The EOF-loop mutant alone is deliberately nonterminating and is caught on both runtimes by the shared two-second recovery subprocess bound.

| Mutants | Check that caught each |
| --- | --- |
| Expression precedence; optional access erased; parenthesized expression becomes arrow | Canonical expression bytes differ from Go on Node and native |
| For-of becomes for-in; type-only import phase lost; keyof becomes readonly | Whole-tree kind, semantic or operator bytes differ from Go on both runtimes |
| Expression counter; whole-tree counter | Tree bytes remain identical; count 0 differs from Go's 18 or 12 on both runtimes |
| JSX text payload; whitespace flag; namespace kind; self-closing kind; type-argument comma; attribute-list length; expression kind; child order; text start | Canonical JSX bytes differ from Go on both runtimes |
| Raw attribute value; JSX name payload | Scanner payload differences caught in whole-tree bytes on both runtimes |
| Dashed member accepted; private name accepted; escaped name accepted | Compiled mutants exit 0; Go-backed refusal controls catch acceptance on both runtimes |
| Octal suggestion; diagnostic code; stranded export; generic child; speculative roots | Diagnostic or recovered tree/root bytes differ from Go on both runtimes |
| EOF loop | Two-second deadline on Node and sanitized native |
| Lint subscription changes DebuggerStatement to EmptyStatement | Finding mismatch against real cohere on Node and native |
| Lint count plus one | Ordinary findings remain identical; count 3 differs from Go's 2 on both runtimes |
| Compiler one-byte output injection | Independent uncached Node/native byte comparison |
| Actual extra parse | Count and all tree bytes unchanged; instruction budget alone |
| Summary plus one | Self-cost sum invariant alone |

## Wider corpus limit

The final wider run used `unset ADAMIC_RECOVERY_START` and:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
ADAMIC_RECOVERY_ARTIFACTS=/workspace/scratch/parser-land-incomplete-final \
go test ./stage1/typescript/parser -run '^TestIncompleteCompilerAgrees$' \
  -count=1 -v -timeout=2h > /tmp/parser-land-incomplete-final.log 2>&1
```

Observed: all 77 files and 22,497 inputs were planned; comparison stopped after 1,818 inputs in 548.510s. Sanitized native case 1815, checker.ts cut token 239888, exceeded the existing ten-second per-process deadline. Slowest Go was 0.788s and Node 3.766s. The [complete log](validation/landing/incomplete.log) retains the failure. No continuation skip, diagnostic filtering or deadline change was applied.

The exact retained 2,132,263-byte input was then compared directly with a separate sixty-second investigative bound, using `--whole --recovery`. Observed: Go 0.315s, Node 1.325s, sanitized native 6.936s; all exit normally with empty stderr and 10,822,952 identical bytes, SHA-256 97e00f9c4e4191d8cc1b0f5abb6317b6725b4f4a3b940a2648ecb95265b5d154. Sanitized flags use the same common flags, replacing `-O2` with `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`. [Direct comparison](validation/landing/case1815.log) and compressed input/outputs are retained. Inference: the ten-second bound lacks concurrent margin for this input. The isolated result does not prove parity for the remaining malformed corpus, and the full corpus gate is still red.

No complete repository gate or best-of-five speed benchmark is claimed. Canonical tree and diagnostic fields are those in the existing protocol; related diagnostics, full JSDoc and every internal Go AST metadata field are outside that claim. Invalid JSX retains the area's explicit refusal behavior.

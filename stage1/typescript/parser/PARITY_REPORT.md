Built: class progress, heritage recovery and all 22,497 pinned comparison inputs matching typescript-go.\
Commits: parser 82fd6c83373b77ee1a34c241ec690004c0d0f6eb; deadline proof 76501d4da95605a078aa77fc746fac47b8da06a7.\
Commands and outputs: all 20 functional parser tests PASS at restart checkpoints; full corpus PASS 4286.34s; vet and cohere PASS.\
Mutants: all 52 caught (43 recovery, three expression, three whole-tree, two counts, one uncached compiler byte mutation).\
Not covered: full repository test gate, performance benchmarks, JSX, complete JSDoc, type checking, or every internal Go AST field.\

# Parser parity 2

Branch `codex/parser-parity-2` starts at fetched `origin/codex/parser-recovery` commit `cf9f33c3b02319e865bf09fc5a0cc914d8600d4b`, retaining its modifier-led arrow fix on top of the requested `b85afdd`. The previous recovery report remains historical evidence.

The port now recovers class members through the active class list instead of repeatedly parsing an invalid token. Empty and malformed heritage lists use the outer and inner list recovery rules, preserve missing commas, and represent valid interface heritage and implements names as type references.

Wider comparison reductions also corrected index parameters, object and class modifiers, accessors and constructors, assignment targets, arrow ambiguity and incomplete bodies, speculative expression type arguments, stranded catch/finally, block diagnostics, static-block await/yield context, property diagnostics, import and typeof qualifiers, malformed signature return separators, tuple list recovery, decorated expressions, and type references with a missing name before a dot. The external Go parser decides the tree and diagnostic output. Speculation restores scanner state, nodes, roots and diagnostics before the real parse.

The external corpus is Microsoft/TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. The unchanged comparison schedules 22,497 cut, remove and duplicate inputs across 77 compiler files. It compares complete diagnostic and canonical tree protocol bytes against typescript-go, on Node and sanitized native. This protocol does not claim every internal Go AST field.

## Toolchain

`bash cloud/setup.sh` succeeded: Go ready 1s; Clang ready 1s; Node ready 1s; submodules ready 1s; build cache warm 166s; done 166s. Go 1.27.1, Clang 20.1.8, Node 24.19.0. `nproc` returned 5 (cgroup CPU quota four cores). Each command sources `/workspace/adamic-tools/env.sh`.

## Final validation

Completed logs are linked below. Test commands write directly to log files. Continuations are debugging evidence and are not credited as the full gate. An environment restart interrupted earlier runs; superseded runs are not credited as final validation.

## Mutant evidence

Every recovery mutant below must compile. Ordinary mutants must terminate normally and differ from the Go oracle independently on Node and sanitized native. Only `eof-loop`, `class-recovery`, and `stranded-catch` are intentional nontermination checks, caught by the explicit two-second deadline on both runtimes.

| Mutant | Check that catches it |
| --- | --- |
| `decorated-expression` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `missing-type-qualifier` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `function-expression-name` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `corpus-eof-loop` | Thirty-second corpus deadline on Node and native |
| `diagnostic-code` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `stranded-export` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `generic-child` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `eof-loop` | Two-second deadline on Node and native |
| `speculative-roots` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `arrow-head-rejection` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `tuple-recovery` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `type-query-diagnostic` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `signature-return-token` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `import-qualifier` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `class-missing-name` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `property-semicolon` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `contextual-literal` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `static-await` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `static-yield` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `class-modifiers` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `static-block-modifier` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `block-equals-diagnostic` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `class-missing-brace` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `class-recovery` | Two-second deadline on Node and native |
| `empty-heritage` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `stranded-catch` | Two-second deadline on Node and native |
| `arrow-missing-block` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `generic-arrow-name` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `generic-arrow-block` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `generic-arrow-parameters` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `generic-arrow-return` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `generic-arrow-speculation` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `type-argument-speculation` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `type-member-modifiers` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `accessor-parameters` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `construct-recognition` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `accessor-follow` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `arrow-block` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `heritage-kind` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `assignment-left` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `object-modifiers` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `index-parameters` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| `heritage-comma` | Diagnostic or canonical tree bytes differ from Go on Node and native |
| Expression precedence | Binary tree shape differs from Go on Node and native |
| Optional access erased | Optional node flag differs from Go on Node and native |
| Parenthesized expression becomes arrow | Node kind differs from Go on Node and native |
| For-of becomes for-in | Statement kind differs from Go on Node and native |
| Type-only import phase lost | Import phase differs from Go on Node and native |
| Keyof becomes readonly | Type operator differs from Go on Node and native |
| Expression counter | Unchanged trees; count 0 differs from Go's 18 on Node and native |
| Whole-tree counter | Unchanged trees; count 0 differs from Go's 12 on Node and native |
| Compiler one-byte string mutation | Native string gains `!`; external Node comparison reports `stdout differs` |

The existing expression mutants change precedence, optional access and parenthesized-expression kind. The whole-tree mutants change for-of to for-in, erase a type-only import phase, and change keyof to readonly. Two counter mutants preserve tree bytes but produce zero instead of Go's 18 expression nodes or 12 whole-tree nodes. The filtered compiler oracle injects a one-byte output error. Final run results and exact commands are recorded below.

Development failures are not mutant evidence: initial class-loop input did not loop and was replaced with `class C { ) }`; several formatted-text anchors matched zero sites and were corrected. The arrow rejection anchor correction passed separately, while the combined run compiled with its older anchor failed. A temporary cohere cycle refusal and line-count/lint failures were corrected before final validation.

Case 19667 reduced to `f(a @x {b});`: observations showed Go retained a decorator under MissingDeclaration while the port produced two arguments, a missing identifier and `x`. Parsing the shared modifier list and emitting Go's zero-width expression diagnostic corrected the canonical tree. All 290 then-current focused cases passed; the deliberately wrong node kind was caught on both runtimes (109.206s combined).

Case 19913 reduced to `type T = .A | B;`: Go retained the missing left identifier and parsed the dotted name; the port stopped before the dot. Using the shared entity-name parser with a type diagnostic corrected recovery. All 295 focused cases and the deliberate missing-qualified-name mutant passed (170.316s combined).

Case 21534 reduced to `const x = { a: function function f() {} };`: the second reserved `function` is not an optional binding identifier. Reusing the binding-identifier predicate restores Go's diagnostic position and recovery tree. The deliberate reversal was caught independently on Node and native (40.17s).

The continuation from case 21534 passed all final 964 inputs in 193.767s. This completed exploration of every scheduled input, but does not replace the final unfiltered gate.

Parallel validation attempts exposed resource deadline contention. The large checker input timed out once at ten seconds, then matched Go and Node in a direct rerun: Go 0.559s, Node 2.658s, sanitized native 5.373s, 4,279,066 identical bytes. A previously passing object-method input timed out once at two seconds, then matched in 0.017s, 0.448s and 0.065s respectively. Neither failed aggregate run is credited. The final package gate runs alone; focused recovery deadlines remain two seconds.

The first isolated full attempt also exceeded ten seconds on checker cuts 1849-1851 with the harness's four workers. A direct case-1849 run terminated normally with 12,900,054 identical bytes: Go 0.820s, Node 2.214s, native 7.572s. Observation: the original deadline lacks margin for concurrent sanitized parse-and-print work at that size. The corpus deadline is now thirty seconds, shared through `incompleteDeadline`; selection, worker count, diagnostics and canonical byte comparisons remain unchanged. A deliberate EOF-loop mutant was executed with that exact shared deadline, and both Node and sanitized native were caught at thirty seconds (82.470s combined). The existing two-second recovery-loop mutants remain unchanged.

The final unfiltered corpus gate passed: 77 compiler files, 22,497 incomplete inputs, complete diagnostics and canonical tree bytes identical on Go, Node and sanitized native. `TestIncompleteCompilerAgrees` completed in 4286.34s. The slowest native parse-and-print took 22.577961594s (case 1919, checker remove token 130837), confirming why ten seconds was insufficient for the concurrent sanitized workload. The slowest Node run took 3.714768311s. No source changed during this gate.

A third environment restart interrupted the package run during `generic-arrow-block`, after the full unfiltered corpus test had passed and after the 299 focused cases had passed. Git verified the same committed source `76501d4da95605a078aa77fc746fac47b8da06a7`, with only this report untracked. The completed corpus result is retained. The aggregate regression and mutant suite is rerun with only `TestIncompleteCompilerAgrees` skipped; no completed corpus result is inferred from the partial regression run.

The first resumed regression command omitted `ADAMIC_TYPESCRIPT_SOURCE`, so the complete-compiler expression and whole-tree tests were conditionally skipped. That aggregate attempt is not credited. The corrected final regression command explicitly sets the pinned corpus path.

## Merge coordination

The user reports that `@system_cohere_lint_rules` is adding numeric `kind: SyntaxKind` metadata to `nodes.ts` and minting sites in `parser.ts`. This branch leaves `nodes.ts`, the central `Parser.make` / `new ParseNode` constructor (current parser lines 259-260), and the direct PrivateIdentifier constructor (line 356) unchanged. It changes parser imports and several `make(...)` call sites in primary expressions, arrow and parameter speculation, type and tuple recovery, type members, index parameters and object members. Those nearby call sites can overlap the other worker's patch; retain the numeric kind metadata while merging this branch's recovery control flow. No numeric kind fields or generated kind table were edited here. Parser source remained unchanged after the heads-up.

A fourth environment restart interrupted the pinned regression at the start of `TestWholeMutants`. Completed PASS results on the unchanged `76501d4` source are retained, including all 43 recovery mutants, 299 focused cases, eight lint combinations, all 42 type-node kinds, 68 generated whole files and obsolete import-attribute recovery. The remaining three tests (`TestWholeMutants`, `TestWholeCountCheckCatchesMutant`, `TestWholeCompilerAgrees`) run in a final filtered command with the pinned corpus environment. No single aggregate package PASS is claimed from either interrupted run; final functional coverage is the union of completed checks on the same immutable source.

The new recovery logic also reads the existing textual node kind in lookahead return-type validation and in `Parser.assignment` through `grammar.leftHandSideKind`. These consumers are listed for the numeric-kind integration alongside the overlapping mint call sites.

## Completed checks and exact commands

All parser source and test checks below ran on `76501d4da95605a078aa77fc746fac47b8da06a7`. Four environment restarts prevented one aggregate package completion; the coverage audit confirms that every one of the 20 functional test functions has a completed PASS result on that same source. Two performance tests were deliberately skipped. The interrupted package commands are not claimed as aggregate PASS results.

Each command first sources `/workspace/adamic-tools/env.sh`.

| Command | Completed output | Evidence |
| --- | --- | --- |
| `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/typescript/parser -count=1 -v -timeout 4h` | `TestIncompleteCompilerAgrees` PASS 4286.34s; 77 files, 22,497 inputs. Package interrupted later in recovery mutants. | [corpus.log](validation/parity-2/corpus.log) |
| `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/typescript/parser -skip '^TestIncompleteCompilerAgrees$' -count=1 -v -timeout 90m` | Every test through obsolete import attributes PASS; all 43 recovery mutants PASS 832.53s. Package interrupted at start of whole mutants. | [regression.log](validation/parity-2/regression.log) |
| `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/typescript/parser -run '^(TestWholeMutants\|TestWholeCountCheckCatchesMutant\|TestWholeCompilerAgrees)$' -count=1 -v -timeout 20m` | PASS 131.250s; remaining four parser mutants caught; 77 whole files and 44,766,682 identical tree bytes. | [whole.log](validation/parity-2/whole.log) |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 20m` | PASS 0.420s; native and Node cache misses prove fresh execution. | [compiler-oracle.log](validation/parity-2/compiler-oracle.log) |
| `go vet ./...` | PASS, empty output. | [go-vet.log](validation/parity-2/go-vet.log) |
| `/workspace/scratch/cohere --no-fix --no-cache stage1/typescript/parser/*.ts` | PASS: 276 rules, 14 checked, 13 of 13 Adamic-ready. | [cohere.log](validation/parity-2/cohere.log) |
| `gofmt -d stage1/typescript/parser/incomplete_test.go stage1/typescript/parser/recovery_test.go` and `git diff --check` | PASS, empty output. | [gofmt.log](validation/parity-2/gofmt.log), [diff-check.log](validation/parity-2/diff-check.log) |

The regular-expression alternatives in the whole-test command are literal pipe characters; they are command arguments, not a shell pipeline.

The final pinned regression also matched all 77 compiler expression files (28,836,875 bytes), all 1,676 generated expression cases (seed 720), all 299 focused recovery cases, eight real cohere source/style combinations, all 42 Go type-node kinds, 68 generated whole files, and obsolete import-attribute diagnostic 2880. The explicit [coverage audit](validation/parity-2/coverage.json) lists all 20 functional test names and all 43 recovery mutant names with no missing results. Combined with the three expression mutants, three whole-tree mutants and two counter mutants, that is 51 parser mutants, plus the independent compiler byte mutant.

The shared corpus deadline was separately proven by:

```bash
go test ./stage1/typescript/parser -run '^TestRecoveryMutants/corpus-eof-loop$' -count=1 -v -timeout 5m
```

Both Node and native were caught at the exact thirty-second deadline; PASS 82.470s. See [deadline-mutant.log](validation/parity-2/deadline-mutant.log). The original insufficient deadline and the direct large-input timing/equality proof are retained in [initial-deadline.log](validation/parity-2/initial-deadline.log) and [checker-measurement.log](validation/parity-2/checker-measurement.log).

The external Go parser checkout is `cohere/TypeScript/tsc` at `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`; cohere itself is `715ba94f3608a6500086b1076ce5cb7e51b836db`. The Go parser and AST were not modified: the oracle uses a virtual Go entry point supplied through a build overlay. The Microsoft compiler corpus checkout remains clean at the pinned v6.0.3 commit.

Setup timing lines and `nproc = 5` are preserved in [setup.log](validation/parity-2/setup.log):

```text
setup: go ready (1s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (166s)
setup: done in 166s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The raw logs preserve trailing tabs for empty canonical AST fields. The validation directory limits its Git whitespace attribute exception to `.log` data; parser source and report whitespace checks remain enabled.

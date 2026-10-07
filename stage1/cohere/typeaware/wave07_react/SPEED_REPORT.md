Built: numeric SyntaxKind listener declarations for the nine completed wave-07 ports; no new claims.
Commits: tested rule code cc198106ac646b62fd204cf88133680214c0e558, based on main e8ba3d5d81de4d3773c723914fccd4c76248b965.
Commands and outputs: see fresh speed-evidence command records and validation below.
Mutants: all nine numeric declaration mutants compile and exit normally; production Go byte comparison catches each.
Not covered: shared numeric-node driver and string-kind body migration, three React ports, four production bridge routes, full option matrices and emitted-JavaScript comparisons.

## Listener declarations

Each owned rule exports `listenerKinds: readonly number[]`. These numbers come from the pinned production Go rule's actual `Run` listener map, not a guessed TypeScript enum. The independent Go probe creates a live program and checker, runs each production rule's registration, sorts the keys and serializes them. A comment containing `exit` enables the blocking-stream rule's conditional registration.

| Rule | Numeric SyntaxKind | Kind |
| --- | --- | --- |
| no-undef-init | 261 | VariableDeclaration |
| @typescript-eslint/prefer-for-of | 249 | ForStatement |
| @typescript-eslint/consistent-indexed-object-style | 188, 201, 265 | TypeLiteral, MappedType, InterfaceDeclaration |
| nexus/correctness-no-process-exit-after-output | 307 | SourceFile |
| nexus/correctness-no-uncleared-race-timeout | 214 | CallExpression |
| nexus/correctness-require-blocking-standard-streams | 307 | SourceFile |
| prefer-rest-params | 79 | Identifier |
| prefer-promise-reject-errors | 214, 215 | CallExpression, NewExpression |
| prefer-regex-literals | 307 | SourceFile |

The native declaration probe matches 358 production Go bytes in ordinary and ASan/UBSan/LSan builds. Each of nine mutants changes its rule's first numeric key by one. Every mutant compiles, exits 0 with empty stderr and differs from the independent production oracle. No compiler failure or crash is counted as a declaration mutant kill.

## Shared API blocker

Current main's `stage1/typescript/parser/nodes.ts` exposes `ParseNode.kind` as a string, with no numeric SyntaxKind field. The shared driver does not consume these listener exports. The nine existing rule bodies retain their previous string-kind traversal and checker behavior. Consequently this change supplies the requested declarations but does not complete the instruction to act only on the handed node or remove per-rule refetches. Implementing that migration requires the shared numeric node API and driver; this unit's shared-file restriction prevents adding those here. No performance improvement is claimed.

The three React claims still require JSX parser integration. Main's parser rejects all three JSX controls with exit 70, while their production Go rules report. The isolated published JSX dependency parses them successfully, but it is not integrated into this branch. No React rule is represented as ported. Four dispatcher routes used by the continuation and regex validation remain private overlays. No shared parser, driver, generator, harness or protected compiler file was edited.

## Validation rerun

With `/workspace/adamic-tools/env.sh` sourced:

```sh
ADAMIC_WAVE07_LANDING=/workspace/wave-07-speed-landing python3 stage1/cohere/typeaware/wave07_react/landing_gate.py
python3 stage1/cohere/typeaware/wave07_react/validate_listeners.py
ADAMIC_WAVE07_LANDING=/workspace/wave-07-speed-landing python3 stage1/cohere/typeaware/wave07_react/landing_bench.py
```

All 16 landing steps exit 0: the original-trio agreement test; six continuation validators; two question validators; `go test ./bridge/tsgo/... -count=1`; decoder/refusal/flag guards; 15 filtered Node oracle fixtures plus its one-byte mutant; `go vet ./...`; formatting; branch JSX blocker and isolated published JSX dependency probes. Exact arguments, environments, elapsed times and outputs are archived. The full repository gate was not run. Source lint separately reports zero findings across 39 owned `.a` files with 276 rules.

| Controls | Findings | Canonical bytes |
| --- | ---: | ---: |
| Original trio | 173 | 47352 |
| Timeout | 11 | 6942 |
| Process exit | 16 | 9638 |
| Blocking streams | 16 | 11990 |
| Rest parameters | 9 | 3431 |
| Promise rejection | 22 | 8175 |
| Regex literals | 41 | 24341 |

Every row matches production Go under normal and ASan/UBSan/LSan execution. The repository corpus covers 287 files and the compiler corpus 77. Original-trio corpus outputs are 22753 bytes / 16 findings and 9282 bytes / 24 findings; each continuation rule emits zero findings with 18485 and 5318 canonical bytes respectively. Regex controls retain 30 suggestions; original controls retain 69 fix edits and seven suggestions. Findings, fixes and suggestions are compared as complete serialized bytes.

The rerun also kills all nine per-rule message/finding mutants by independent Go bytes, the native type-reference graph mutant, and five Go raw-fact mutants (type-reference graph, symbol value, flow edge, module path, regex character span). Released-handle registry retention mutants are caught by the required panic-70 checks. Bridge sanitizer/ownership/length guards, decoder wire mutants and the Node one-byte mutant pass again; their exact individual outputs are in the archived logs and the earlier landing report.

Toolchain setup: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 88s, total 88s. `nproc` is 5; cgroup quota is four CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. Production cohere pin is 715ba94f3608a6500086b1076ce5cb7e51b836db; TypeScript-Go is 8d550c837c90bd1805b047b7eeccc2baac2d5e7a; compiler corpus is TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8.

## Native time against Go

Three alternating complete-output runs per implementation and corpus, after builds and gates finish. Medians in seconds include startup, checker work and serialization. All 84 executions match their paired output. The declarations are not consumed by today's shared driver; native remains slower.

| Unit | Repository native / Go | Compiler native / Go |
| --- | --- | --- |
| original-trio | 0.336406 / 0.143251 | 2.131369 / 0.387616 |
| timer | 0.292942 / 0.139728 | 1.793276 / 0.317810 |
| process | 0.283373 / 0.147280 | 1.780066 / 0.304467 |
| streams | 0.359949 / 0.143629 | 1.759295 / 0.317813 |
| rest | 0.285571 / 0.136537 | 1.865883 / 0.332407 |
| promise | 0.257720 / 0.126504 | 1.732649 / 0.288957 |
| regex | 0.270590 / 0.138751 | 2.017758 / 0.312456 |

Proof is archived in [speed-evidence](speed-evidence), with commands, complete streams, fixtures, hashes, tested revision, dependency pins and timings. Previous landing evidence is preserved separately. Push destination is exclusively `codex/typeaware-wave-07`; no new claims or pull request.

Built: supplied-node CFG truthiness and labels helpers in separate .a files, plus the shared-table literal rewrite of the earlier Tailwind normalizer.
Commits: truthiness d6c77911d0cea8f28c4d3d52badc447ef065157d; labels 4b9465715c6c392fff27902ecc1580206f024b65; parked rules 7f99254378ee9fd8b50f4cfd9aa18cebe5291b2d, all on main c01907a7.
Checks: owned package PASS 105.859s before isolated regex rewrite; changed normalizer PASS 9.358s afterward; final owned vet clean. Two new helpers match 60,686 observations / 246,314 bytes across Go, source Node, emitted JS and sanitized native.
Mutants: six new predicate/label mutants and three revalidated regex controls compile/run and fail only actual-Go output comparison; the other 11 owned controls also remain green.
Uncovered: zero final rule blockers removed; counter bigint-return lowering and shared rule harness remain blocked; full CFG/rule integration and repository gate are not claimed. Other unclaimed helpers remain.

Current detailed reports are TRUTHINESS.md, LABELS.md and REGEX-MIGRATION.md. All four CFG consumers lose two additional dependencies, with residual lists retained in control_flow_readiness.json. Claims preceded code and both own branches remain rebased onto current main. No new helper claim is made after the two completed ports; work stops at the already measured exact-counter primitive boundary without changing shared files.

Previous landing evidence, retained as history:

Rebased: helper branch onto current origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 after the final fetch observed main advancing.
Commits: previous pushed helper 0b654cb13a383b717123bd63cadb2c00c1c04a23; rebased implementation b32f13a5c1ac98bba32d9c7bf2f3f01899b524be before this evidence commit.
Checks: complete owned helper package PASS 61.366s and vet clean on the rebased branch; all five delivered helpers retain four-backend actual-Go parity.
Mutants: six new CFG controls plus eight existing controls are caught; each new semantic mutant compiles and exits successfully on every backend before comparison fails.
Uncovered: exact nextBuildCount still refuses bigint return lowering; complete CFG and default rule integration remain blocked or uncovered; zero final rule blockers removed by the three new helpers alone.

Main's new changes add Stage 3 work, documentation and its oracle hook. The compiler, lint sources and pinned cohere are unchanged from f8013f0b. The helper rebase is clean and introduces no implementation change. Nevertheless the complete owned package was rerun uncached after rebasing, rather than treating the old run as a new result.

Fresh observations retain 10,983 bigint-normalization cases / 1,003,998 bytes, 30,267 numeric-predicate cases / 179,169 bytes and 26,732 allocator cases / 950,119 bytes. The three new .a helpers serve array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Their APIs, all four actual Go fixture families, every mutant, zero-call normalization boundary, allocator trace sizes and exact residual helper lists are in CONTROL-FLOW.md and control_flow_readiness.json. This is three prerequisites delivered to four consumers, not four complete rule ports. All four retain other helpers.

All previous Tailwind helper checks and supported-separator refusal controls also pass in the same invocation. The exact counter probe still matches Go on Node and still refuses at gaps/atomic-counter.a:4:10: stage 0 can't lower a function returning bigint yet. Its compiling number approximation differs at row 5 on all three backends. No production counter, native exact-counter parity or alternative-representation impossibility proof is claimed.

Commands with /workspace/adamic-tools/env.sh sourced:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=20m > /tmp/wave12-c019-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-c019-helper-vet.log 2>&1
```

Setup was rerun in this turn before main advanced: Go/clang/Node/submodules ready at 0s, cache warm 40s, total 40s, nproc 5, four-core quota, 17.6 GB. Toolchain and compiler files did not change in the new main delta. Scope: owned helper package only for this final refresh, previous inherited foundation/comment evidence remains in F801-LANDING.md. No complete gate or new throughput measurement. Only owned helper/claim files change, and only the helper branch is pushed. The parked sibling is separately rebased and rechecked before the final summary. No further helper is claimed beyond the three delivered CFG helpers and the previously blocked exact counter.

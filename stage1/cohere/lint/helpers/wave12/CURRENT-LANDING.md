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

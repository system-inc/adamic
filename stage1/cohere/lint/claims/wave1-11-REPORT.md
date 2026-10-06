Built: foundation merge and pushed claim; no new rule port, because directory registration cannot load .a modules.
Commits: c34b1a0 registration merge; e688af8 helper merge; 8940d16 claim pushed before any rule code.
Commands and outputs: setup 89s, nproc 5; registry PASS, witnesses PASS 5438 bytes, helpers PASS 22412 lines, external oracle PASS; .a probe exit 1.
Mutants: four inherited helper output mutants and external one-byte oracle control caught; no new rule mutant; .a loader rejection earns no semantic credit.
Not covered: new-rule findings/fixes, TypeScript compiler corpus, stage1 corpus, emitted JavaScript, new-rule sanitizers or throughput.

# Lint wave 1 slot 11: blocked handoff

## Assignment and existing ports

The helper REPORT.md refers to HELPERS.md for the ordered names. Its first section has 30 rules; the first three of its policy section are positions 31 to 33:

| Position | Rule | Disposition |
| --- | --- | --- |
| 31 | adamic/no-definite-assignment | Claimed, not implemented. Registration cannot accept an owned .a module. |
| 32 | base/boundary-no-global-container | Skipped: origin/codex/stage1-lint-batch4 already has base_boundary_no_global_container.ts, registered in batch4_registry.ts. |
| 33 | base/consistency-no-hand-built-declared-error | Skipped: the same branch has base_consistency_no_hand_built_declared_error.ts, registered in batch4_registry.ts. |

Batch 4's BATCH4.md names implementation 7bd94a2f92cc17d12bb69bbd78bfe2c4ac4ca63d and records 13 and 14 captured upstream cases respectively. Those are prior-worker observations, not reruns in this slot. No existing port was imported or rewritten.

## Concrete blocker

The requested foundation does not implement the requested .a contract:

- stage1/cohere/lint/registry/registry.go line 95 opens rule.ts unconditionally.
- Line 149 defaults the owned mutant to rule.ts; lines 151 to 152 require the mutant file to end in .ts.
- Line 187 generates a hardcoded import of rules/<slug>/rule.ts.
- stage1/cohere/lint/lint_test.go line 47 copies only .ts source modules into mutant builds. Its stage1 corpus collection also filters .ts files.

In an isolated copy of all five registered rules, `go run ./cmd/lint-registry -root <copy>` succeeds and prints all five names. Renaming only the debugger module to rule.a makes the same command exit 1 with `open .../rules/no-debugger/rule.ts: no such file or directory`. The baseline guards against interpreting an already-broken setup as evidence. See wave1-11-evidence/extension-probe.log and probe-registration.sh.

Observation: the foundation rejects an otherwise unchanged module when its source extension is .a. Inference: registering the remaining rule requires shared generator and harness changes, or a new .ts source/alias. The former is outside this unit's rule-directory territory and the latter contradicts the explicit extension requirement. Neither was done. There is no incomplete rule directory to poison registry discovery.

The required follow-up is shared .a support: discover rule.a, generate its actual import extension, allow .a mutant modules, copy .a dependencies, and include .a stage1 corpus files. Keep legacy .ts rules readable. This report is not a claim that that fix exists.

## Foundation reconciliation

Started from fetched origin/main d090af5. The checkout's fetch configuration tracked only main; fetching all branch refs made both foundations available. Historical recursive submodule fetch was interrupted after the top-level refs arrived; cloud/setup.sh subsequently updated the pinned submodules successfully.

Registration merged cleanly. Helpers did not: README.md, lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go conflicted. The resolution retains registration's architecture and imports helpers/, inventory/ and HELPERS.md. The helper branch's unrelated legacy dispatcher and volume additions were not integrated. Both foundation tips are ancestors of the merge commit, but this is a scoped reconciliation, not preservation of every legacy helper-branch feature.

The user-named docs/parallel-work.md does not exist in main or either foundation tip. CLAUDE.md and docs/lint-registration.md supply the available ownership/registration instructions.

## Toolchain

`bash cloud/setup.sh > /tmp/wave1-11-setup.log 2>&1` succeeded. Sourced `/workspace/adamic-tools/env.sh`, the installed path. Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc` returned 5; cgroup cpu.max is 400000 100000.

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (89s)
setup: done in 89s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

## Coverage limits

Findings per second for the assigned rules: native, Node and Go are all unmeasured in this slot. No new rule was implemented, so no parity or mutant success is claimed. The two skipped rules' prior measurements do not establish parity against this branch. Full repository gate and repository-wide vet were not run. No PR was opened and no compiler files were edited.

## Validation completed in this slot

Every test wrote directly to a log. The named tests actually ran.

| Command | Observed result | Evidence |
| --- | --- | --- |
| `go test ./stage1/cohere/lint/registry -count=1 -v` | PASS, 0.092s | wave1-11-evidence/registry.log |
| `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout 10m` | First attempt FAIL 107.603s: Go build succeeded, but dependency-download text violated the harness's empty-output condition. Retry PASS 11.842s, 5438 identical Go/Node/native bytes. | witnesses-first.log and witnesses-retry.log |
| `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout 20m` | PASS, 151.416s: 22412 baseline output lines match Go, Node and sanitized native; ten policy refusals match; four compiling semantic mutants caught; explicit gaps pass. | helpers.log |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 5m` | PASS, 60.581s; no cache hits | oracle.log |
| `git diff --check` | exit 0 | Empty output |

The inherited helper mutants ran on native and disagreed with the independent Go answer after compiling and executing successfully:

| Mutation | Comparison catch |
| --- | --- |
| JSON reader permits raw control characters | line 15157: valid instead of invalid |
| oneOf permits multiple matching branches | line 15166: valid instead of invalid |
| Policy renderer omits interpolation | line 22166: unresolved constructor placeholder instead of supplied sentinel |
| Strict options ignores unknown fields | line 15178: valid instead of invalid |

Registry tests also exercised structural rejection controls for duplicate names, unknown descriptor fields, missing hooks, invalid kinds, missing oracle/factory/class exports, absent listeners, unsafe names and duplicate adapters. These are validation failures, not semantic rule mutants. The external oracle's one-byte comparison control passed. None of these catches proves the unimplemented rule.

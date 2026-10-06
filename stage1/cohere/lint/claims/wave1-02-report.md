Built: published the slot 02 claim and reproducible foundation-blocker evidence; no new rule port.
Commits: registration merge 449f707; preimplementation claim 15f2ec9; this report is in the subsequent evidence commit.
Commands and outputs: setup PASS 81s, nproc 5, registry tests PASS 0.063s, filtered uncached oracle PASS 15.668s.
Mutants: 11 existing registry validation mutants rejected; no new semantic rule mutant was run or claimed.
Not covered: no-restricted-types implementation, findings/fix parity, emitted JavaScript parity, throughput, or full repository gate.

# Lint wave 1, slot 02 report

Branch: `codex/lint-wave1-02`, pushed without a pull request.

The requested three-rule assignment is obtained from the ordered 46-rule list
in HELPERS.md, which the helper REPORT.md links. REPORT.md itself contains no
ordered 46-rule list. docs/parallel-work.md is absent in origin/main and both
requested foundation refs. The available docs/lint-registration.md was read.

## Assignment and duplicate-port observations

| Position | Rule | Disposition |
|---|---|---|
| 4 | @typescript-eslint/no-explicit-any | Skip: origin/codex/stage1-lint-batch5 at dcef3eebfddcb297464bdce398d850775852587d, no_explicit_any.ts |
| 5 | @typescript-eslint/no-inferrable-types | Skip: origin/codex/stage1-lint-batch4-typescript at 63782c53678711717f9fd4f12762ce3d103b9a3d, no_inferrable_types.ts |
| 6 | @typescript-eslint/no-restricted-types | Claimed, blocked by incompatible foundations |

All origin heads were fetched. A scan of 191 distinct .a/.ts lint source blobs
across those refs found no no-restricted-types implementation. The two existing
ports were inspected; their historical validation was not rerun by this unit.
The original all-head fetch continued into a large historical TypeScript
submodule fetch; that recursive continuation was stopped after origin heads
were available. The required setup's pinned recursive submodule update succeeded.

## Foundations cannot satisfy this unit as fetched

Base main: d090af531216ddd3c25a0dede6b82d7c0a6edf76.
Registration: 48ecd9302bf3954a4ddbbd14c28ba09148c1c1a8.
Helpers: 5d13f5baaecaf11d4ea62de693426f69a1f41bba.

Registration merged successfully. Helpers then conflicted in these six files
under stage1/cohere/lint: README.md, lint.ts, lint_test.go, main.ts, settings.ts,
and testdata/oracle.go. The helper branch carries shared scanner changes that
collide with the registration migration. The merge was aborted. The pushed
branch contains the registration merge and the claim/evidence, not the helpers
merge. Neither foundation was silently discarded to force a green merge.

Reproduction without changing the worktree:

```
git merge-tree --write-tree --messages 449f707f8da0d7f53d78e27e8a3301215ac1516a origin/codex/lint-helpers
```

Exit 1, six conflicts. Raw output: wave1-02-evidence/merge.log.

The registration foundation also hardcodes rule.ts in Discover and in generated
imports, and accepts only .ts mutant module paths. Its descriptor has no source
extension field and rejects unknown fields. This is incompatible with the
explicit requirement that new Adamic files be .a. The inherited CLAUDE.md says
new rules may change only their own directory and must never change a dispatch,
oracle, corpus, or copied-file list. A rule-local file cannot correct these
shared generator and harness assumptions.

Two isolated scratch probes reproduced the blockers using an unchanged copy of
the existing debugger rule, with no production mutation:

- Rename rule.ts to rule.a and run the generator: exit 1,
  `open /tmp/lint-wave1-02-probe/rules/no-debugger/rule.ts: no such file or directory`.
- Keep the existing rule.ts copy and set mutant metadata's file to rule.a:
  exit 1, `mutant file must be an owned TS module`.

These are registration rejection probes, not semantic mutants caught by output
comparison. Their logs are extension.log and mutant-extension.log.

This part cannot be completed within the unit's current ownership and extension
requirements. The next prerequisite is a reconciled foundation with .a-aware
rule discovery, generated imports and test module copying. No compiler file,
shared dispatch or foundation source was edited by this unit. No unvalidated
partial rule is committed or described as ported.

## Toolchain and bounded validation

`bash cloud/setup.sh > /tmp/lint-wave1-02-setup.log 2>&1` succeeded:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (81s)
setup: done in 81s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0. nproc: 5.
Sourcing /opt/adamic-tools/env.sh failed with `No such file or directory`;
workaround: source /workspace/adamic-tools/env.sh, the configured tool directory.
Every validation command sourced that working path. Setup and test output were
written directly to log files and then read, never piped.

Commands:

```
go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-wave1-02-registry.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout 5m > /tmp/lint-wave1-02-oracle.log 2>&1
```

Registry: PASS, 0.063s. The existing validation mutants were rejected for each
of: duplicate public name, unknown field, missing named export, bad kind,
missing oracle export, missing listener, missing factory, missing class,
missing finish hook, unsafe public name, and duplicate oracle adapter.
These prove existing registry validation only; none substitutes for a required
compiling semantic no-restricted-types mutant.

Filtered oracle: PASS, 15.668s, all six input fixtures, six probe misses and zero
probe cache hits. Setup also compiled all packages and all test binaries without
running their tests. The full repository gate and lint parity suite were not run:
there is no rule implementation to validate. Raw logs are preserved next to this
report under wave1-02-evidence/.

## Throughput and uncovered work

Findings per second for native, Node and Go: unmeasured. No new rule was runnable,
so a zero rate or unrelated baseline would misrepresent this unit's performance.
No comparison of the assigned rule over TypeScript src/compiler, stage1 sources
or cohere rule tests was run. Findings, fixes, suggestions, configuration decode,
source/emitted JavaScript/sanitized native parity and a comparison-only mutant
all remain outstanding for no-restricted-types after the prerequisites are fixed.

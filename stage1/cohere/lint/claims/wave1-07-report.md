Built: a pushed three-rule reservation and reproducible foundation blocker evidence; no rule ports.
Commits: main d090af5, registration merge d077906, claim 60b5b47; report commit follows.
Commands and outputs: setup exit 0 in 78s, nproc 5; registry tests PASS in 0.083s; helpers merge exit 1.
Mutants: no rule mutants; scratch extension control exit 0, identical module renamed .a fails registration with exit 1.
Not covered: all three rule implementations, Go/Node/emitted-JS/native parity, rule mutants, throughput, full gate.

# Wave 1 slot 07 report

Branch: `codex/lint-wave1-07`. No pull request opened.

## Reservation

Positions 19, 20 and 21 are `no-underscore-dangle`, `no-unsafe-negation` and
`no-unsafe-optional-chaining`. The helpers REPORT.md does not itself list the
ordered names: it refers to HELPERS.md, whose measured handoff provides them.
The reservation was committed and pushed before any implementation.

Fetched all parent origin branches explicitly because this checkout's ordinary
fetch only tracks main. A scan of 222 refs found no matching stage1 lint source
or descriptors, excluding inventory/helper ledgers. Upstream Go files exist for
all three, serving as a name-search control. Frequency logs mention these rules
but do not implement them. No rule was skipped as already ported. This is a
bounded source-name scan, not a semantic audit of every branch.

## Observed foundation blockers

Started from `origin/main` at `d090af531216ddd3c25a0dede6b82d7c0a6edf76`.
Merged registration `48ecd9302bf3954a4ddbbd14c28ba09148c1c1a8` successfully.
Merging helpers `5d13f5baaecaf11d4ea62de693426f69a1f41bba` failed in six
shared files: lint README.md, lint.ts, lint_test.go, main.ts, settings.ts and
 testdata/oracle.go. The conflicted merge was aborted; the clean registration
merge remains. `git merge-tree --write-tree HEAD origin/codex/lint-helpers`
reproduced the six conflicts without changing the working tree, exit 1.
See `wave1-07-evidence/merge.log`.

The conflicts reconcile registration's directory dispatch with the helpers
branch's inherited monolithic rule ports. Choosing one shared side wholesale
would require deciding which existing functionality to retain. The unit's
territory does not include those dispatch/test/oracle files.

`docs/parallel-work.md` is absent on main and both requested foundations.
`git show <ref>:docs/parallel-work.md` fails for all three refs.

Registration also cannot register an Adamic `.a` rule within an owned rule
directory: registry.go reads `rule.ts` unconditionally, emits imports to
`rule.ts`, defaults mutants to that filename and requires a `.ts` suffix.
The lint test harness copies only `.ts` modules. These shared files must support
`.a` before an owned `.a` implementation can pass the requested harness.

A scratch copy of the existing no-debugger directory is the control:

```
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry -root /tmp/lint-wave1-07-registry-Ox9iGj
```

Before renaming the copied module: exit 0, stdout `no-debugger`.
After renaming only the copied rule.ts to rule.a: exit 1,
`open /tmp/lint-wave1-07-registry-Ox9iGj/rules/no-debugger/rule.ts: no such file or directory`.
Both runs wrote logs directly; they are saved under the evidence directory.
This is a negative registration probe, not a compiling semantic rule mutant.
No repository source was changed by this probe.

## Setup and validation

`bash cloud/setup.sh > /tmp/lint-wave1-07-setup.log 2>&1` exited 0.
Then sourced `/workspace/adamic-tools/env.sh`, the configured tool path.

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (78s)
setup: done in 78s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0; `nproc` printed 5.
The independent setup completed while the all-branches fetch remained in a
redundant recursive TypeScript submodule fetch. That redundant fetch was stopped
after the 222 parent refs were available; the successful setup log is independent.

`go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-wave1-07-registry-tests.log 2>&1`
exited 0: `PASS`, package time 0.083s. Existing descriptor rejection controls
and deterministic generation passed. No compiler or rule implementation changed.

## Coverage and required next work

No production Adamic files were written. No compiler files were edited.
No rule was ported, no rule mutant ran, and no byte-for-byte parity result or
native/Node/Go findings-per-second measurement is claimed. The full gate,
filtered external oracle and full lint runtime tests were not run.

To resume this unit, the shared foundation needs a reconciliation of the six
conflicts, `.a` discovery/import/mutant/copy support, and the missing parallel-work
document or its intended replacement. The pushed reservation remains available
for the three assigned rule directories. This report records blockers rather
than passing off unregistered code as a completed port.

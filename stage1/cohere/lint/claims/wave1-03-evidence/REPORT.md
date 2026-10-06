Built: pushed the three-rule claim and preserved registration; rule ports are blocked by incompatible foundations and .a registration.
Commits: registration merge 8b9b5849df7839cfcfe25cc4824a11671b15d01e; pre-code claim db16927123bdaa8c1da7a1024900d14039468cc6; this evidence commit follows.
Commands and outputs: setup exit 0 in 77s, nproc 5; registry tests PASS in 0.063s; filtered input oracle PASS in 9.946s; merge-tree exit 1; .a registry probe exit 1.
Mutants: no assigned-rule mutants were written or run; all 11 existing registry rejection probes passed and are listed below.
Not covered: the three ports, Go/Node/emitted-JavaScript/native parity, rule fixes, per-rule mutants, findings/s, and the full gate.

## Assigned rules

Positions 7 through 9 are @typescript-eslint/no-unused-expressions,
@typescript-eslint/unified-signatures and class-methods-use-this. The helper
REPORT.md refers to HELPERS.md for the ordered list; the list itself is in
HELPERS.md. All fetched origin branches were searched under stage1/cohere/lint.
Only old frequency evidence contained these names outside the inventory/helper
files. No implementation was found, so none was skipped as already ported.

## Observed blockers

Current origin/main was d090af5. Its checkout fetched only main by default;
fetching all branch refs exposed the foundations. Recursive fetching then
stalled in cohere/TypeScript; it was stopped after the origin refs were fetched.
The required setup subsequently initialized the pinned submodules successfully.

Registration 48ecd93 merged successfully. Helpers 5d13f5b did not: README.md,
lint.ts, lint_test.go, main.ts, settings.ts and testdata/oracle.go conflicted.
The helper branch carries the earlier monolithic rule engine in addition to
its standalone helpers and inventory. The failed merge was aborted. A later
merge-tree probe reproduced all six conflicts without changing the worktree.
The raw output is in merge.log.

The requested docs/parallel-work.md does not exist on main or either foundation.
The registration contract available here is docs/lint-registration.md.

Registration requires rule.ts and generates imports ending in rule.ts. Its
mutant validator accepts only .ts modules, and its copied-port harness discovers
only .ts modules. An isolated copy of no-debugger with its source named rule.a
fails generation: "open .../rules/no-debugger/rule.ts: no such file or directory".
The probe exits 1; extension.log contains its exact output. No production rule
was mutated. No new Adamic .ts source was written.

Inference: providing registered .a rule ports requires shared registry/harness
changes; integrating the helpers requires shared-engine conflict resolution.
Both exceed this unit's rule-directory territory. Silently selecting one whole
engine or adding .ts rule files would violate the requested contract.

## Validation

All test output was written directly to logs and then read.

- bash cloud/setup.sh > /tmp/lint-wave1-03-setup.log 2>&1: exit 0.
- source /workspace/adamic-tools/env.sh: used for all Go probes and tests.
- go test ./stage1/cohere/lint/registry -count=1 -v: PASS, 0.063s.
- go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout 5m: PASS, 9.946s, all six fixtures, six probe misses and zero cache hits.
- git merge-tree --write-tree HEAD origin/codex/lint-helpers: exit 1, six conflicts.
- go run ./cmd/lint-registry -root <temporary .a rule directory>: exit 1.

Setup timings: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s;
build cache warm 77s; done 77s. nproc: 5. Go 1.27.1, clang 20.1.8,
Node 24.19.0. Setup also compiled repository packages and test binaries without
running tests. That is not a complete gate pass.

Existing registry rejection probes: duplicate public name, unknown descriptor
field, missing named hook, bad AST kind, missing oracle export, no listener,
missing factory, missing class, missing finish hook, unsafe public name and
duplicate oracle adapter. Each was caught by registry validation. These are
existing metadata checks, not compiling semantic mutants for the assigned rules.

No native/Node/Go findings-per-second figures are reported: no new rules were
implemented, and measuring the existing five would not certify this unit.
No pull request was opened. The claim was pushed before any implementation;
this branch contains no implementation of the three claimed rules.

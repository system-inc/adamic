Claimed react/no-unsafe, react/self-closing-comp and sort-vars; no rule implementation delivered because the shared registration contract rejects .a modules.
Commits: ad1bc3e registration merge, 179d3bf helpers merge, 7915fd3 claim pushed before implementation; this report commit is named in the final response.
Commands: setup passed in 75s; nproc=5; registry tests PASS in 0.088s; filtered uncached byte oracle PASS in 26.233s; blocker probes failed as expected.
Mutants: no assigned-rule semantic mutant run; registry tests rejected descriptor mutants; .a compatibility probes were rejected before execution and earn no semantic-mutant credit.
Not covered: three rule ports, findings/fixes parity on any required corpus, emitted JS, sanitized native, throughput and full repository gate.

# Observations

Started from origin/main d090af5 on codex/lint-wave1-10. Initial fetch succeeded,
but remote.origin.fetch selected only main. Explicitly fetching all branch refs
made the named foundations available. The registration foundation at 48ecd930
merged cleanly. The helper foundation at 5d13f5ba did not: six shared driver
files conflicted (README.md, lint.ts, lint_test.go, main.ts, settings.ts and
 testdata/oracle.go). Preserved the registration side of these conflicts and
retained the standalone helpers and inventory. No compiler file was edited.

The helper REPORT.md does not contain the ordered 46-name list; it points to
HELPERS.md. Positions 28, 29 and 30 there are react/no-unsafe,
react/self-closing-comp and sort-vars. Searched fetched origin branch source
blobs under stage1/cohere/lint for all three names. Only helper readiness,
descriptor/coverage and inventory metadata matched. No assigned rule was
skipped as ported. The claim was pushed before any rule implementation.

CLAUDE.md and docs/lint-registration.md require rule-owned files and forbid
shared dispatch/oracle/corpus list changes. docs/parallel-work.md was absent
from origin/main and the registration foundation, and remains absent here.
Followed the actual registration contract for diagnosis.

# Reproduced blockers

1. Registry .a module rejection. Copied the five existing rule directories to
   /tmp/lint-wave1-10-probe/rules. Ran `go run ./cmd/lint-registry -root
   /tmp/lint-wave1-10-probe`: exit 0, listing all five names. Renamed the copied
   no-debugger/rule.ts to rule.a, leaving its contents unchanged. The same
   command exited 1: `open .../rules/no-debugger/rule.ts: no such file or directory`.
   registry.go reads rule.ts at line 95 and emits rule.ts imports at line 187.
   The descriptor has no source-path field, and unknown fields are rejected.
2. Registry .a mutant rejection. Restored the original copied rule module,
   copied its messages module to messages.a and set the scratch mutant's file
   to messages.a. The command exited 1: `mutant file must be an owned TS module`.
   registry.go line 151 checks the .ts suffix. The harness also omits .a from
   its copied port modules at lint_test.go line 47. A rule-owned .a file cannot
   satisfy the registration and mutant contracts without shared changes.
3. Sort-vars fix-range rejection. Built the unchanged testdata/oracle.go through
   a Go overlay with the evidence selection.go choosing the real upstream rules.
   `go build -overlay=/tmp/lint-wave1-10-probe/overlay.json -o
   /tmp/lint-wave1-10-probe/oracle <two virtual Go files>` exited 0.
   Ran that binary with a manifest selecting sort-vars on `var b, a;`. It printed
   the real finding at column 8 then exited 2 with `panic: unexpected fix shape`.
   The upstream finding spans the out-of-order declarator; its single fix spans
   the whole block. oracle.go line 61 rejects unequal ranges. RuleContext.report
   similarly couples repair and diagnostic ranges. Relabeling the finding range
   to match the repair would violate the requested Go parity.
4. JSX corpus rejection. The same unchanged Go oracle, selecting the real
   react/self-closing-comp on `const x = <Foo></Foo>;`, exited 2 with
   `panic: invalid corpus ...: [Type expected.]`. oracle.go line 89 always chooses
   ScriptKindTS. Filename or witness content cannot select TSX there. The Adamic
   parser source search also found no JSX handling; the directly observed blocker
   reported here is the Go harness rejection, not a claim about all parser behavior.

The probes operate only on scratch files. These failures establish compatibility
limits; none is credited as a rule semantic mutant. No new Adamic .ts source was
authored. No partial or placeholder rule descriptor was installed.

# Validation and timing

All test output was redirected to files and read after execution.

- `bash cloud/setup.sh > /tmp/lint-wave1-10-setup.log 2>&1`: exit 0.
  Go 1.27.1, clang 20.1.8, Node 24.19.0. Timing: go ready 1s;
  clang ready 1s; node ready 1s; submodules ready 1s; build cache warm 75s;
  done in 75s on 5 processors; cgroup cpu.max 400000 100000; 17.6 GB.
- Sourced /workspace/adamic-tools/env.sh for probes and tests. `nproc`: 5.
- `go test ./stage1/cohere/lint/registry -count=1 -v >
  /tmp/lint-wave1-10-registry-tests.log 2>&1`: PASS, 0.088s. Includes existing
  descriptor-rejection mutants and deterministic regeneration.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  '^TestTheOracleCatchesOneByte$' -count=1 -timeout 10m >
  /tmp/lint-wave1-10-filtered-oracle.log 2>&1`: PASS, 26.233s.
- `git diff --check`: exit 0, empty output before this report was added.

Native, Node and Go findings/second are unavailable: no assigned rule completed
registration or a parity run, so an inherited five-rule timing would not measure
this unit. The full lint package, helper package, repository-wide vet and full
repository gate were not run. The merge resolution is not claimed validated by
these narrow tests beyond the registration package.

# Required foundation work

Inference from the reproduced failures: support rule.a and .a mutant/copy paths
in the shared generator and harness; carry fix ranges separately from diagnostic
ranges through Finding, formatting and repair application; select TSX and provide
the JSX AST adapter for JSX rule cases. These files are outside this unit's
three-rule territory. The claim remains reserved, with implementation blocked.
Evidence logs and the Go selection adapter are adjacent in wave1-10-evidence/.

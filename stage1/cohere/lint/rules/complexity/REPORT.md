# Complexity

Ports the pinned Go cohere rule without changing its implementation or shared
registration files. The SourceFile listener maintains independent counters for
functions, class static blocks and class field initializers. Computed field keys
remain in the enclosing scope. Static property names use the existing shared
`propertyName` helper.

Oracle behavior retained even where surprising:

- A raw `maximum: 0` silences the rule, while `max: 0` reports empty functions.
  Captured typed `Maximum: 0` means the decoded reporting threshold of zero.
- Functions stored in properties borrow the property's name and head range,
  even when a function expression has its own name.
- The Go head-range finder scans raw characters, including comments. The
  `raw-head.ts.txt` witness retains its opening parenthesis inside a comment.
- Classic counts case clauses; modified counts switches. Program-level decisions
  never produce a finding.

Eight witnesses cover decisions, nested scopes, names/modifiers, both switch
variants, zero thresholds, logical assignments and raw head ranges. Their
options are adjacent JSON files. Findings have no fixes or suggestions, matching
Go.

The `complexity-zero-entry-paths` mutant changes every new frame's initial count
from one to zero. The generated harness must compile it and catch its differing
messages against Go on Node and emitted JavaScript.

Validation logs are under `/workspace/scratch/s13-complexity-*`. Setup uses
`cloud/setup.sh` and its WASI SDK option. The full package uses the clean
TypeScript pin `050880ce59e30b356b686bd3144efe24f875ebc8`, enabled benchmarks and
a single fresh directory for both profile variables. No shared helper or compiler
files are modified.

## Observed validation

- `go run ./cmd/lint-registry` passed. The focused scratch-overlay test matched
  181 unique captured upstream cases, eight configured witnesses and inherited
  generated rows across Go, Node, emitted JavaScript and sanitized native:
  560,240 identical output bytes (`/workspace/scratch/s13-complexity-focused.log`).
- The entry-path mutant compiled, ran and disagreed with Go on Node and emitted
  JavaScript (focused log, lines 11 and 23), and sanitized native
  (`/workspace/scratch/s13-complexity-native-mutant-2.log`, line 6).
  The native probe uses original witnesses because the copied mutant tree does
  not retain their option sidecars.
- The full lint package passed with 152 passing test events, zero failing test
  events and one skip: `TestCheckerBridgeRefusalPending`. Package time was
  1,337.397 seconds; command wall time was 1,339.250 seconds. `nproc` was 5
  (cgroup CPU quota: 4). Load averages before: 0.53, 2.28, 2.39; after: 3.60,
  4.89, 4.20. Complete events: `/workspace/scratch/s13-complexity-full.jsonl`;
  counts and machine metadata: `/workspace/scratch/s13-complexity-full-summary.json`.
- The full command was `go test ./stage1/cohere/lint -count=1 -json -timeout=30m`
  with `ADAMIC_GATE_UNCACHED=1`, `ADAMIC_TEST_WASI=1`, `ADAMIC_LINT_BENCH=1`,
  `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3`, and both
  `ADAMIC_LINT_PROFILE_DIR` and `ADAMIC_LINT_PROFILE_SNAPSHOTS` set to the fresh
  `/workspace/scratch/s13-complexity-profile.OiLUGf`. WASI SDK 27 is installed at
  `/workspace/adamic-tools/wasi-sdk`.
- Initial setup timing lines: Node ready 0.032s; Go ready 0.034s; clang ready
  0.280s; markdown dependencies ready 0.888s; submodules ready 3.403s; Go build
  ready 210.440s; build cache warm 210.605s; done 210.629s. Logs:
  `/workspace/scratch/s13-complexity-setup.log` and
  `/workspace/scratch/s13-complexity-wasi-setup.log`.

No rule helper or language gap blocked this port. A zero-skip whole-package
result is unavailable on this baseline: `checker_pending_test.go:51` skips
because `internal/load/prelude.d.ts` lacks `TSGoError`, awaiting the existing
`codex/tsgo-errors-as-values` unit. That dependency is outside this rule's
territory. General malformed-input parser recovery remains the inherited
harness boundary described in `stage1/cohere/lint/GAPS.md`.

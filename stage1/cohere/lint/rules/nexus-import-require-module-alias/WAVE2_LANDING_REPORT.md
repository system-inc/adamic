# Wave 2 landing verification

This report supersedes the historical area and parking reports in these rule
directories. The landing branch starts at
95968dd93ad0876245f181af63931c3134e14b3d, which contains the recovered parser,
current shared finding model and helper packages. No shared harness, parser,
registry generator or compiler files are changed.

Ten winning, claimed rule directories are restored here. The previously parked
constructor rule now includes all 92 upstream cases and a firing recovery
witness at
`../typescript-eslint-no-unnecessary-parameter-property-assignment/testdata/modified-object-binding.ts.txt`.
Its messages are preserved verbatim in its own messages.a. Both historical
parser and fix-budget reproducer files are retained. The batch-only shouting
rule was already integrated into the base and is not duplicated here.

## Owned correction

The first full package run caught a stale no-lonely-if whitespace implementation
on its existing BOM boundary witness. Current Go cohere uses ECMAScript
whitespace. The port now uses String.trim(), including BOM and excluding NEL.
Its active mutant substitutes a non-whitespace character for BOM before trim,
withholding a fix while still exiting successfully. The old extra-space mutant
metadata is retained as verification/prior-mutant.json in that rule directory.

## Inputs and commands

Setup: `bash cloud/setup.sh`, then
`source /workspace/adamic-tools/env.sh`. Setup completed in 202.375 seconds;
Go 0.052, Node 0.047, clang 0.243, Markdown 0.905, submodules 3.550,
Go build 202.012, warm 202.347 seconds. nproc reports 5 (CPU quota 4).

ADAMIC_TYPESCRIPT_SOURCE is /tmp/adamic-typescript-6.0.3, official tag v6.0.3
at 050880ce59e30b356b686bd3144efe24f875ebc8.
ADAMIC_LINT_BENCH=1. Both ADAMIC_LINT_PROFILE_DIR and
ADAMIC_LINT_PROFILE_SNAPSHOTS are /tmp/wave14-unpark-profile-final.

The full package command is
`go test -json -count=1 -timeout 85m -parallel 2 ./stage1/cohere/lint`.
The registry command is `go run ./cmd/lint-registry`.
Every added Go file is checked with gofmt -l.

## Additional native evidence

The shared mutant gate tests syntax mutants on Node and emitted JavaScript and
uses a native canary. To verify every owned mutant on native as well, the
rule-local verification/native_mutants.py tool copies the source graph, applies
one active descriptor mutant at a time and compiles with the repository's
load, lower, JavaScript and sanitized BuildSplitTSGo pipeline. It never modifies
the working source graph or the shared harness. The Go oracle is independently
built by the shared harness from real current cohere rule implementations.

Each comparison includes every captured upstream case, every owned raw witness,
all stage1 Adamic source files and TypeScript 6.0.3 src/compiler. Malformed inputs
use the recovered parser mode; no case is filtered out. A mutation only counts
as caught when the runtime exits zero with empty stderr and its bytes differ
from Go. After the ten mutations, the original graph is rebuilt and every
manifest must match Go on Node, emitted JavaScript and sanitized native.

The command and tool paths are recorded in evidence/wave2-native-mutants.log.
The completed full-package and native results are recorded below.

## Completed independent results

All ten source mutants were caught on all three runtimes, with zero exits and
empty stderr. Restored sources match Go for every manifest on all three runtimes.
The proof digests are in evidence/wave2-byte-proof.json; the complete sequence
is in evidence/wave2-native-mutants.log. All 11 added Go files produce no
gofmt -l output, registry generation passes, and go vet ./stage1/cohere/lint
passes. No owned rule requires an unavailable helper or checker fact.

## Upstream coverage

| Rule | Unique upstream cases |
| --- | ---: |
| nexus/import-require-module-alias | 28 |
| @typescript-eslint/no-confusing-non-null-assertion | 28 |
| @typescript-eslint/no-duplicate-enum-values | 57 |
| @typescript-eslint/no-dynamic-delete | 42 |
| @typescript-eslint/no-extra-non-null-assertion | 19 |
| @typescript-eslint/no-misused-new | 45 |
| @typescript-eslint/no-unnecessary-parameter-property-assignment | 92 |
| no-lone-blocks | 77 |
| no-lonely-if | 32 |
| no-loss-of-precision | 151 |

These counts include every real upstream test captured by each descriptor
prefix; none of the owned cases are excluded. Witnesses and the natural source
corpora are additional to these upstream counts.

## Throughput observations

The full package benchmarks report best-of-five throughput on this host,
with all ten owned rules registered alongside the integrated rules.
These are whole-harness observations, not isolated rule timings.

- best of 5 JSX Go: 0.059848s, 5263.33 findings/s (346 files, 315 findings)
- best of 5 JSX native: 0.121942s, 2583.20 findings/s (346 files, 315 findings)
- best of 5 JSX Node: 0.486667s, 647.26 findings/s (346 files, 315 findings)
- best of 5 Go: 2.005697s, 14153.68 findings/s (77 files, 28388 findings)
- best of 5 native: 11.622459s, 2442.51 findings/s (77 files, 28388 findings)
- best of 5 Node: 6.813578s, 4166.39 findings/s (77 files, 28388 findings)

## Full package result

PASS in 1684.243 seconds. The final run records 156 passes, zero failures and
one skip including subtests; top-level results are 41 passes, zero failures
and one skip. TestOwnedWitnesses, TestMutants, TestRulesAgree,
TestCompilerAndStage1Agree, TestThroughput, TestProfileArtifacts and
TestProfileSnapshotsAgree all pass.

The sole skip is the pre-existing TestCheckerBridgeRefusalPending at
checker_pending_test.go:51. It awaits codex/tsgo-errors-as-values: tsgoInspect
must return TSGoError from the C error buffer. No input-controlled test skips,
and no owned rule or upstream case is omitted. This unit does not change or
bypass that shared pending test. The complete JSON event log is in
evidence/wave2-package.log and the counts are in evidence/wave2-package-summary.json.

Scope: the full lint package, independent owned byte comparisons and package
vet were run. The repository-wide gate was not run. No rules remain parked in
this owned landing. Rules reassigned to other workers by DEDUP_LEDGER.md are
not restored, and the already integrated shouting rule is not duplicated.

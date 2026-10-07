Built: rebased wave 13 onto current main c01907a7; eight rules complete and one partially blocked, with no new claims.
Commits: rebased code 811d4eb6; this report and validation/landing-c019 record the renewed landing checks.
Commands and outputs: original agreement gate PASS 206.978s; next controls 129/129, more controls 430/433 in normal and ASan runs; both frozen corpora equal.
Mutants: original three rule mutants, two export-question mutants and released-registry mutant caught again; all 18 listener metadata mutants pass their rejection checks.
Not covered: three shared parser recovery inputs, numeric-node driver migration, full repository gate and emitted-JavaScript rule comparison.

Main added Stage 3 sources, documentation and a dormant internal/oracle test hook.
The stage 0 compiler, bridge, shared parser and existing wave 13 sources are unchanged
from f8013f0b. The rebase completed without conflicts. Existing freshly built
f801 compiler, C archives and native suites were therefore reused. Toolchain setup
was not repeated; its last reported total is 217s and nproc remains 5.
No shared files were edited or reverted, including allocator leak checks.

## Renewed checks

TestWave13AgreementAndMutants ran against the production Go oracle with normal and
sanitized native outputs: 18,957 control bytes, repository 19,144 bytes with one
finding, compiler 7,087 bytes with four findings. Released handles refuse with
exit 70. The registry mutant exits 0 and is caught by the required-refusal check.

Original rule mutants unassigned, caught and exports compile and exit 0 with empty
stderr; byte comparison catches offsets 1345, 4136 and 14959. The export-chain-flags
and export-module-lookup mutants likewise require byte comparison, at offsets 14677
and 17114. Six continuation rule mutants and other bridge/foundation mutants were
not rerun because their compiler and sources are unchanged; their current-main
f801 evidence remains in LANDING_F801_REPORT.md.

Both continuation suites reran every upstream control under normal and ASan builds.
Next has 129 equal controls. More has exactly three expected parser failures among
433 controls, with exit 70 and no findings before rule execution:
53d1c0a9ffc2fefa and aff6ced2ca61fe89 require recovery from JSX-like syntax in .ts;
defcb4c4ce6a2921 requires accepting a yield label and break yield.
No rule workaround, omitted fixture or shared parser edit was introduced.
The React IR/SSA/capture parking exception does not apply to this core rule.

Both continuation suites reran the frozen compiler (77 roots, 4933 bytes) and
repository (287 roots, 18485 bytes) populations, with zero findings and exact Go
agreement. Timing in these logs is concurrent verification, not a new benchmark.
The previous isolated native/Go compiler ratios of 3.25 to 7.63 remain historical
measurements over unchanged artifacts. No speed improvement is claimed.

The whole owned listener metadata package passes, including nine source-declaration
and nine JSON mutations. TestTheOracleCatchesOneByte passes. The new Stage 3 hook
skips as designed with no fixture environment. Scoped go vet passes. The initial
filtered metadata command matched no metadata tests; the subsequent unfiltered
owned-package run provides that coverage.

## Regex and driver status

The fetched codex/lint-regex table contains 107 sites at the same cohere pin and
zero rows for these nine rule implementations. Inspection of their production
Go sources and native implementations found no rule regex or hand-rolled regex
replacement to migrate. The Go regexp in listener_kinds_test.go parses test
metadata only. No new regex code was required. The fetched table branch SHA is
recorded in summary.json.

Numeric listener declarations and rule.json kinds already exist for all nine
rules. Shared ParseNode still exposes string kinds and existing drivers pass
whole-file contexts rather than supplied-node numeric callbacks. Completing that
migration requires the shared driver/parser work; no per-node conversion fallback
was added. No batch-8 Diagnostic landing SHA has been supplied.

This branch is pushed to its own name only. It remains partially blocked rather
than fully oracle-green. No additional claims were taken under the landing cap.

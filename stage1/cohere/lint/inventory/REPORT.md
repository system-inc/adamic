# Inventory unit report

Built a Go inventory generator for all 491 registered rules in ten families.
Captured 50,501 harness invocations for 485 rules; six have no captured invocation.
Measured Go cohere on 77 compiler files and 325 tracked repository source files.
Found 60 existing implementations across scanner, batch 2 and batch 3, with documented limits.
Prepared helper rankings, three dispatch waves and three filled worker prompts.

Branch `codex/lint-inventory` is cut from requested scanner tip
`0090256e607c3f2de7d5b67cef680ec010f95c1d`, rather than main, following the
unit-specific instruction. Only `stage1/cohere/lint/inventory/` is changed.
No parser, compiler, runtime, other lint files or submodule content changed.

## Artifacts and observations

- `inventory.json` and `inventory.md`: registry-complete names, family, defining
  and same-stem companion source files/lines, test files and executed caller
  locations, fixer/suggestion/option flags, decoder presence, direct/transitive
  resolved external calls, checker questions, exact branch tips and statuses,
  raw Go frequencies and explicit unknowns. Source convention totals 209,859
  lines; shared files can serve multiple rules. There are 2,161 ranked function
  dependencies, with package rankings in PIPELINE.md.
- `PIPELINE.md`: land common AST/range/report/repair/option contracts, then shared
  judgments, before fan-out. Among 431 rules without an observed port on the
  available tips: 30 syntax candidates without a shared judgment blocker, 198
  syntax rules waiting on named helpers, and 203 type/binding rules waiting on
  named bridge questions. These are conservative queues conditional on a proven
  AST adapter, not proof every Go AST method already exists in Adamic.
- Highest-use shared judgments include policy rendering (108 rules), property
  names (50), JSX element parts (45), comments (40), import bindings (31) and
  binding writes (25). Entire helper packages and functions are ranked with
  rule membership, not just a top-ten list. Existing batch 2 and 3 contexts and
  repair records should be reconciled by one integrator rather than rebuilt
  independently by every worker.
- `PROMPTS.md`: a reusable worker prompt and concrete examples for no-debugger
  (conditional automatic fix), no-warning-comments (comments/options), and
  no-unsafe-call (type questions and eight messages). The first two already
  exist and are verification/lift examples, not duplicate port assignments.

The measured registry reports 85 rules with automatic fix paths, 50 with
suggestion paths and 213 with option decoders. Type information is conservatively
required by 204 rules' reachable helpers; 197 declare NeedsTypeChecker.

Compiler frequencies are available for 487 rules; four require options and
remain unknown. Repository frequencies are available for 291 rules; missing
`.a` checker answers and required settings invalidate the remaining counts.
Both corpora have zero parse exclusions. The repository corpus includes
intentional negative fixtures, and frequencies bypass configuration/suppressions
with nil/default options. They are potential rule-API findings, not the current
lint gate's enabled-rule findings. No-debugger's compiler/repository counts are
1/0; no-unsafe-call's compiler count is 90 and repository count is unknown.

## Tip evidence

- Scanner: `0090256e607c3f2de7d5b67cef680ec010f95c1d`, 30 observed rules.
  VOLUME.md bounds method-signature-style's malformed-source recovery.
- Batch 2: `c4373c05259c7235cf1202d5cd4138c16d498caa`, 20 additional rules.
  BATCH2.md, not its inherited five-rule REPORT.md, is the current report.
  One malformed no-async-promise-executor case is explicitly refused.
- Batch 3: `fa9781c2c0911c52312002ba97ffd4b560d178ae`, ten additional rules.
  This branch appeared during the run and was fetched before final generation.
  BATCH3.md bounds two identifier rules on nine missing JSX fixture shapes.
- Batch 4 was absent from the remote at the final check. Its work is unknown,
  not presumed unported. Refresh that ref before assigning overlapping work.

Stage status requires an executable selector and names its source and report.
Message-table entries alone are rejected. It records a port with its documented
scope, not an independent certification of every Go/native semantic branch.

## Setup

`git fetch origin` succeeded; requested non-main refs were fetched explicitly
because the remote's default refspec fetched only main. `bash cloud/setup.sh`
succeeded. Each shell sourced `/workspace/adamic-tools/env.sh`.

```text
go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
v24.19.0
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (38s)
setup: done in 38s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

Cohere pin: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
TypeScript v6.0.3 pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Compiler checkout is scratch-only at `/tmp/lint-inventory-typescript`.

## Commands and results

All test output was written to logs, never piped. Run from the repository root
with the setup environment sourced, except the explicitly named cohere cwd.

```sh
go run ./stage1/cohere/lint/inventory -compiler /tmp/lint-inventory-typescript > /tmp/lint-inventory-full.log 2>&1
# Initial full run captured tests, then refused an incomplete .a program.
go run ./stage1/cohere/lint/inventory -compiler /tmp/lint-inventory-typescript -capture-tests=false > /tmp/lint-inventory-full-v3.log 2>&1
# Completed measurement, restoring the captured ledger.
go run ./stage1/cohere/lint/inventory -capture-tests=false -measure=false -reuse /workspace/adamic/stage1/cohere/lint/inventory/inventory.json > /tmp/lint-inventory-refresh-batch3.log 2>&1
python3 stage1/cohere/lint/inventory/testdata/analyze.py > /tmp/lint-inventory-analysis.log 2>&1
go test -count=1 -v ./stage1/cohere/lint/inventory > /tmp/lint-inventory-tests-final.log 2>&1
python3 stage1/cohere/lint/inventory/testdata/mutants.py stage1/cohere/lint/inventory/validation > /tmp/lint-inventory-mutants-final.log 2>&1
go vet ./... > /tmp/lint-inventory-vet-all.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v > /tmp/lint-inventory-oracle.log 2>&1
# cwd cohere:
go test ./internal/lint/registry -count=1 -v > /tmp/lint-inventory-registry.log 2>&1
```

The capture command inside the generator is
`go test -overlay=<capture-overlay> -count=1 -timeout=20m ./internal/lint/rules/...`
from cohere. It finished in 2m5.556570556s with exit 1: every family except
Tailwind passed. Tailwind and its collapse package failed because repository
fixtures point at unavailable local design systems, and their nonempty-corpus
controls correctly failed. The unmodified failure output is in cohere-tests.log.
Eight captured Tailwind rule counts are partial; five Tailwind rules and
nexus/localization-no-untranslated-value have no captured harness invocation.
Their test files are still listed. Direct Context/corpus tests are not relabeled
as zero cases.

Final generator package tests pass; repository-wide vet exits zero with empty
output; filtered external oracle passes in 5.183s; cohere registry's full gate
passes in 24.109s, including its ESLint core corpus comparisons. Raw registry
output is compressed losslessly in validation/registry.log.gz. Logs for all
other commands and mutations are retained under validation/ or beside the data.

## Mutants

Each baseline runs under the identical exact test filter before its mutant.
Every counted mutant compiles and fails the named test, not Go's build check.

| Mutant | Check that catches it |
|---|---|
| Treat FAIL/unrun-family text as an ok summary | TestFailedFamilyCannotBeMarkedPassed |
| Replace the unsupported .a input with supported .ts | TestAdamicCorpusRequiresExplicitBridge |
| Treat any quoted rule name as an implementation | TestBranchEvidenceRequiresExecutableSelector |
| Drop traversal into called sibling functions | TestTransitiveSiblingAndMethodDependencies |
| Hide reachable Context.TypeChecker field use | TestTransitiveSiblingAndMethodDependencies |
| Inflate distinct-rule helper denominator by one | TestHelperRankingCountsRulesOnce |
| Print an unmeasured frequency as zero | TestUnknownFrequencyIsNotZero |
| Count each actual Go finding twice | TestRegisteredCorpusControl |

The real no-debugger positive control must fire exactly once, and a clean
identifier control must add no finding. The resolved-call fixture also proves
an unreachable private helper is excluded. The .a boundary test builds the
otherwise identical .ts control first, then asserts the pinned cohere program
refuses .a even with sourceExtensions at the correct top-level config location.
Its error says one named file, zero program files; syntax-only measurement
parses that input separately rather than accepting an incomplete program.

A first message-name mutation left an unused regexp import and did not compile;
it is discarded, not counted. The final mutation keeps the import used and
produces a wrong answer. Early runs also exposed a nil-program access in the
positive control and synthetic harness rule names absent from the registry;
both were corrected before the completed measurement.

## Not covered

No lint rule is newly ported by this unit. No complete repository go test gate
or existing 30-rule native parity/mutation suite was rerun: this unit changes
only the inventory tool, and runs its package plus the filtered external oracle,
cohere's rule capture and registry suite, and repository-wide vet.

Static resolution cannot certify all reflective/interface callbacks. Option
schemas are identified by decoder APIs, not reconstructed into a complete
schema language. Source counts use the documented filename convention; shared
sibling code is named separately. Test counts are harness invocations, including
repeated configurations and controls, not unique fixture strings or assertions.
No positive control or subject-population counter was added for each of all
491 rules. Runtime frequency counts do not prove every branch was exercised.

Ready-wave membership is an inference conditional on common adapters and source
review. No shared helper's Go/Adamic equivalence is certified here, and missing
checker, parser recovery, JSX, settings and external theme data remain explicit
prerequisites. Dynamic branch tips can change after the recorded snapshot.

Built: merged the current lint area and moved six existing wave ports to its single RuleContext.checker; no new claims.
Commits: exact merge, tested source and rule hashes are in checker-landing-evidence/metadata.json; final branch SHA is reported once.
Checks: four rules fully match captured cases and witnesses; two partially match; the enabled-input lint run is 27 pass / 7 fail / 1 skip, with its corpus snapshot failure corrected by a passing rerun.
Mutants: all 81 registered rule mutants caught; all six owned mutants caught again on Node, emitted JavaScript and native; C bridge ASan and stale-handle mutants caught.
Stopped: shared parser recovery, capture recovery and JSX census blockers; six additional ports await checker/helper dependencies and three hook claims remain parked.

## Merge and ownership

Merged origin/area/stage1-lint c4bdc23fa into the existing branch, not rebased.
Both sides' tests remain. Active unified descriptors use only RuleContext.checker,
with typed and node true, named kind subscriptions and the handed ParseNode.
Each port's helpers and unchanged-Go adapter live in its own descriptor directory.
Configured adapters decode their upstream options types. No shared harness,
registration generator, checker routing or protected compiler source was changed.
The historical standalone drivers remain as historical evidence, not another
checker attached to the unified harness. No claim beyond existing wave ownership
was added.

Eight wave-owned Go bridge files were adapted to the incoming SourceFile and
RootedFilePath APIs. Their filtered facts tests pass in 0.400 seconds. The entire
checker fact package passes in 0.403 seconds. Full C bridge tests pass in 173.278
seconds, including ASan's string-length mutation and the retained-released-handle
mutation. Their complete logs are archived. Registry, gofmt, vet and diff checks
pass. Setup took 220.504 seconds; nproc 5, CPU quota four cores, memory 17.6 GB.

## Independent rule comparison

The shared TestRulesAgree stops at its first refusal. To keep the other ports
moving, a separate runner preserves that run's complete upstream capture,
unchanged Go oracle, sanitized native artifact and emitted JavaScript. Every
owned unique source/file/options row is compared independently. Native records
only the area's checker answers; source Node and emitted JavaScript replay those
answers through the area's Checker and its source/project guards. There are no
worker-authored answers or checker substitutes. All complete outputs, stderr,
projects, manifests and replay records are archived in owned-cases.tar.gz; original
upstream captures and emitted artifact are compressed beside it. Artifact and
source hashes and the exact runner are retained.

| Rule | Identical captured rows / total | Status |
| --- | ---: | --- |
| no-undef-init | 43 / 43 | Green, including its owned witness |
| @typescript-eslint/prefer-for-of | 97 / 97 | Green, including its owned witness |
| prefer-rest-params | 23 / 23 | Green, including its owned witness |
| react/jsx-no-undef | 45 / 45 | Green, including its owned witness |
| @typescript-eslint/consistent-indexed-object-style | 119 / 121 | Two explicit parser refusals; both configured witnesses green |
| prefer-promise-reject-errors | 123 / 124 | One captured malformed-TSX row rejected by the shared Go oracle; owned witness green |

All 453 rows are present: 450 agree and the three failures are retained, not
removed or renamed. Exact sources, options and failure paths are listed in
owned-receipt.json and CHECKER_BLOCKERS.md. Every real upstream test is captured
by the descriptor's prefix. These are current captures, unlike the historical
standalone reports that filtered parser-recovery controls.

TestOwnedWitnesses passes over every descriptor across Go, source Node, emitted
JavaScript and sanitized native. TestMutants passes all 81 registered mutation
subtests. The exact six owned message mutants also pass a separate fresh rerun
following whitespace cleanup; no mutant survives on any port runtime. All are
valid programs that exit cleanly; independent Go bytes detect their changed
findings, rather than a compiler or sanitizer failure. C bridge mutations have
their separately named sanitizer and lifetime detectors.

## Full enabled-input package and corrected corpus attempt

checker-landing-evidence/run.py takes the checkout and a fresh output directory.
It runs the entire lint package with ADAMIC_LINT_BENCH=1, the pinned TypeScript
source tree, and fresh profile output and profile-snapshot inputs. Nothing was
filtered from that run. events.jsonl.gz and stderr.log.gz contain complete output;
receipt.json includes commands, every input, all load samples and counts.

Full run: 27 pass / 7 fail / 1 skip at top level; 112 pass / 7 fail / 1 skip
including subtests. Wall time 1980.200 seconds, nproc 5. Starting load was
3.953 / 5.954 / 5.226, ending load 1.859 / 2.962 / 3.947, and maximum sampled
one-minute load 8.397. All 81 rule mutants and all owned witnesses passed.

Six failures remain shared blockers: TestRulesAgree, TestNodeTableIsLinkOnly,
TestShardsAgree and TestProfileSnapshotsAgree stop on the empty index-signature
parser refusal; TestJsxLintTrees and TestJsxLintReleaseAndThroughput fail the fixed
expected JSX-case map, which lacks the new rules. CHECKER_BLOCKERS.md names the
exact upstream consumers and reproducers. Neither assertion was relaxed.

The seventh failure, TestCompilerAndStage1Agree, was caused by removing trailing
whitespace from helpers while that corpus test was reading those helpers as
inputs. Its complete failed snapshot comparison remains in the full event log.
A fresh stable-tree TestCompilerAndStage1Agree passed over all 960 compiler and
repository files in 650.221 seconds; stable-corpus.log.gz records the rerun.
metadata.json proves that trailing whitespace is the only rule-source difference
from the compiled source commit. The six owned mutant rerun passed after cleanup.
The supplemental pass is kept separate; the original full-run counts are not
rewritten. Corpus rows without a program explicitly print the harness's typed
rule declines; typed upstream and witness certification above supplies projects.
This does not claim full typed compiler-corpus coverage for every owned rule.

TestCheckerBridgeRefusalPending is the only Go-test skip, awaiting TSGoError in
internal/load/prelude.d.ts and the bridge error-as-value transport. Its exact
reason is in the log. No missing-input check was bypassed, and the complete
repository-wide gate was not run.

## Native time versus Go

Per-case process/program/lint elapsed sums under the concurrent package gates,
not an isolated benchmark. Each successful row's outputs are identical.

| Rule | Native seconds | Go seconds |
| --- | ---: | ---: |
| no-undef-init | 3.344232 | 2.638133 |
| prefer-for-of | 7.105589 | 5.569041 |
| consistent-indexed-object-style | 8.959839 | 7.154580 |
| prefer-rest-params | 1.711808 | 1.351764 |
| prefer-promise-reject-errors | 9.635242 | 7.721882 |
| jsx-no-undef | 3.349752 | 2.635921 |

Partial rows include attempted-runtime time; the two blocked rules' timing is not
a certification of the missing comparisons. Full shared throughput rounds and
findings are in the full event log. Native remains slower on these measurements.

## Deferred existing ownership

CHECKER_BLOCKERS.md lists all exact cohere symbols and file:line dependencies for
six other existing ports: no-process-exit-after-output, no-uncleared-race-timeout,
require-blocking-standard-streams, prefer-regex-literals, jsx-fragments and
no-adjacent-inline-elements. Four old private question implementations still
lack production routes, and two pragma consumers lack an external declaration
view. They are not represented as green unified rules. Three hook rules remain
parked for high-level IR, SSA and capture analysis. Their prior claims and checks
were preserved. No new rule was claimed in this turn.

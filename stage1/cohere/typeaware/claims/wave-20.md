# Type-aware wave 20

Branch: `codex/typeaware-wave-20`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Claims, using descending compiler plus repository counts and lexical ties from
VOLUME_REPORT.md's `validation-volume/*-all.counts`, excluding the 26 ports:

| Remaining position | Rule | Combined findings |
| --- | --- | ---: |
| 58 | @typescript-eslint/no-floating-promises | 0 |
| 59 | @typescript-eslint/no-implied-eval | 0 |
| 60 | @typescript-eslint/no-meaningless-void-operator | 0 |

Checked 268 fetched origin refs for claims and native source ports before claiming.
No conflicting claim or native port found. The checker-dependent count inventory
contains 197 rules; method-signature-style is one of the 26 ports but is not in
that checker-dependent count inventory, leaving 172 ranked candidates.

## Second batch

The first batch is complete and pushed in `41b6c3e1`. After fetching all origin
heads on October 7, the first three remaining by-volume candidates, excluding
ports on origin/main and origin/codex/tsgo-c-library and claims on all 325 fetched
origin refs, are:

| Rule | Combined findings |
| --- | ---: |
| nexus/correctness-no-process-exit-after-output | 0 |
| nexus/correctness-no-uncleared-race-timeout | 0 |
| nexus/correctness-require-blocking-standard-streams | 0 |

The scan inspected 33 distinct Markdown claim blobs, finding 96 claimed inventory
rules, and 25 checker-dependent existing ports. It leaves 76 unclaimed entries.
This update is pushed before second-batch implementation.

## Third batch

The second batch is complete and pushed in `deb33e49`. After fetching all
origin heads on October 7, the first three remaining candidates in the combined
by-volume ranking, excluding ports on origin/main and origin/codex/tsgo-c-library
and claim documents on all 356 origin refs, are:

| Rule | Combined findings |
| --- | ---: |
| prefer-promise-reject-errors | 0 |
| prefer-regex-literals | 0 |
| prefer-rest-params | 0 |

The scan inspected 33 distinct Markdown claim blobs, finding 132 claimed
checker-dependent inventory rules and 25 existing checker-dependent ports.
It leaves 40 unclaimed entries before these three claims. Compressed test
evidence beneath claims/ was excluded from claim documents. This update is
pushed before third-batch implementation.

## Fourth batch

The previous nine rules are complete and pushed through `1ef7ed97`.
Fetched all 389 origin refs on October 7 and inspected 33 distinct
Markdown claim documents. The first three remaining combined-volume candidates
are:

| Rule | Combined findings |
| --- | ---: |
| react-hooks/set-state-in-effect | 0 |
| react-hooks/set-state-in-render | 0 |
| react-hooks/static-components | 0 |

This claim is pushed before implementation. Native JSX parsing and React
control-flow substrate availability are being checked; shared files will not
be modified to work around a missing dependency.

Fourth-batch status: all three claims are blocked by missing shared native JSX
parsing, demonstrated on a valid TSX control (Go exit 0, native parser panic 70).
See each rule directory's `BLOCKED.md`. No fourth-batch port is claimed complete.

## Fourth batch parked

Per Ahra's latest instruction, all three fourth-batch claims are PARKED and
counted as finished for the landing-first cap, without claiming complete ports:
react-hooks/set-state-in-effect, react-hooks/set-state-in-render, and
react-hooks/static-components. Blockers: native source-to-HIR lowering, SSA,
capture translation, memoization and control analysis. Shared analysis modules
are being ported on #dnv6f2c; JSX is landing on area/stage1-lint. Existing partial
metadata and blocker evidence are already pushed. Reconcile overlapping claims
with waves 06 and 29 during integration; this worker adds no duplicate kernels.

## Fifth batch

All existing work is based on current main f8013f0b and oracle-green, pushed at
d837657ab. All-heads selection inspected 529 origin refs and 33 distinct Markdown
claim blobs. Exact rule-name boundaries prevent confusing namespaced rules with
core names. Inventory: 197; 154 claimed; 25 baseline checker-dependent ports;
18 unclaimed before this batch, ranked by combined findings and lexical ties.
The first three eligible rules, each with combined volume zero, are:

| Rule | Status |
| --- | --- |
| react/jsx-fragments | Native analysis tested; shared registration blocked |
| react/jsx-no-constructed-context-values | Native analysis tested; shared registration blocked |
| react/jsx-no-undef | Native analysis tested; shared registration blocked |

These rules use AST and checker facts, not the parked native React HIR/SSA
pipeline. Claim update is pushed before any implementation. Shared JSX and
numeric node handoff availability will constrain their source entry points.

Fifth-batch status: BLOCKED, not parked and not ported. Shared native JSX
productions and numeric handed-node interfaces are not integrated on current
main or area/stage1-lint. Go controls and numeric declaration checks pass;
valid JSX panics before native dispatch. See each new rule's BLOCKED.md.
The three metadata declarations are not a source-analysis completion claim.

The named ab70f38d4 harness is now incorporated with current main c01907a7.
Fresh native parsing closes the JSX failure; handed-node and shared finding
support are present. Numeric ParseNode kinds remain unavailable. Fifth-batch
source analyses are still unfinished; fourth-batch HIR claims remain parked.
See react-jsx-fragments/HARNESS_REBASE_REPORT.md. No additional claims.

Fifth-batch progress: react/jsx-no-undef now has a tested native source analysis
in its own directory (661963afd): 53 controls/29 findings, both corpora, sanitizer,
semantic mutant and released query PASS. Shared registration/checker context and
Node fact transport remain unfinished. The other two JSX analyses are not yet
implemented; the batch is not complete and no further rules are claimed.

Landing update: rebased onto origin/main b8fb957aa; preserved ab70f38d4.
The three current JSX rule.json declarations use named ast.Kind values.
Numeric-kind blockers in earlier reports are superseded by the user correction.
Shared checker-context registration and the other two JSX analyses remain unfinished.
No additional claims. See react-jsx-fragments/LANDING_B8FB957AA_REPORT.md.

Fifth-batch progress: react/jsx-fragments now has a tested native analysis
(7f646aa90): 57 controls/34 findings, both corpora, sanitizer, semantic mutant
and released query PASS. Current main b8fb957aa and named harness 41eb6eab2
are ancestors. Shared checker-context registration remains pending for the
two JSX source analyses; jsx-no-constructed-context-values remains unfinished.
No further rules were claimed. See react-jsx-fragments/ANALYSIS_REPORT.md.

Latest-main landing: main advanced to 39638d9e2 during the push verification.
Own commits were rebased onto a base containing that main and harness 41eb6eab2.
Fragment source is now 4e4d58063; validated landing source is 5dc1ba68f.
No new rules claimed. See react-jsx-fragments/LANDING_39638D9E2_REPORT.md.

Integrated-lint landing: rebased onto origin/area/stage1-lint d65a8f931,
including main 39638d9e2 and the requested 50a5f105 integration. Validated
source 26f1d659f: all eleven implemented analyses, all mutants and released
checks green again. Shared typed registration and constructed-context-values
source remain unfinished; no new claims. See
react-jsx-fragments/LANDING_AREA_D65A8F931_REPORT.md.

Fifth-batch source completion: constructed-context-values now has the full
construction and memo/escape native analysis, validated at 47e48bb178ebd0020cba80346d8cafef9259c6cb.
171 controls/97 findings, both corpora, sanitizer, five semantic mutants and
three released queries pass. All three JSX source analyses are now tested.
Shared checker-context registration remains blocked; these AST/checker rules
are not parked HIR/SSA rules. See the owned constructed-context ANALYSIS_REPORT.md.

Final continuation audit after the fifth-batch source push: 626 origin refs,
33 distinct Markdown claim blobs, 197 ranked rules, 172 claimed and 25 baseline
checker-dependent ports. No unclaimed rules remain; no sixth batch was claimed.
Skipped the three apparent colon-delimited entries already reserved by wave 01:
nexus/correctness-no-implicit-return, @typescript-eslint/no-deprecated and
no-else-return. Full selection evidence is in the owned constructed-context
validation/remaining-selection.json.gz. Shared typed registration remains blocked.

Current-main landing: clean rebase onto origin/area/stage1-lint b84a9d931,
including main c7991b900. Validated source 3dcf3c918db0695bd00551d7a130024245877b16.
All twelve source analyses green again: 1,132 controls/870 findings, both corpora,
sanitisers, all mutants and released handles. Shared typed registration remains
blocked; older HIR claims remain parked. The full gate and its seventeen mandatory
external-input checks were not run or claimed green. Fresh 637-ref audit finds
no unclaimed rule; no new claims. See the owned constructed-context
LANDING_C7991B900_REPORT.md and validation/landing-c7991b900 evidence.

Legacy-registry integration landing: clean rebase onto area b46914832,
including main c7991b900. Validated source 0a8eed884329ef0965de8f52699df17f2c140a86.
All twelve private analyses green again: 1,132 controls/870 findings, both
corpora, sanitizers, every mutant and released handle. Shared typed registration
remains blocked; older HIR claims remain parked. No new claims: the fresh
650-ref audit finds no unclaimed rule. Full gate and its seventeen mandatory
input checks were not run or claimed green. See the owned constructed-context
LANDING_B46914832_REPORT.md and validation/landing-b46914832 evidence.

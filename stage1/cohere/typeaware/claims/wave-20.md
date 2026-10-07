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
| react/jsx-fragments | Claimed |
| react/jsx-no-constructed-context-values | Claimed |
| react/jsx-no-undef | Claimed |

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

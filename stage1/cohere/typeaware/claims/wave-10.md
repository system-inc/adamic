# Type-aware wave 10 claim

Branch: `codex/typeaware-wave-10`.
Base: `origin/codex/tsgo-c-library`, `0d540f413625f016f20fea39761c7b184f335de6`.

Positions use the combined compiler and repository counts linked by
VOLUME_REPORT.md, sorted by descending count with lexical ties, excluding the
26 rules already ported on the base branch.

| Remaining position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 28 | no-unmodified-loop-condition | 4 | 0 | 4 |
| 29 | @typescript-eslint/no-redundant-type-constituents | 3 | 0 | 3 |
| 30 | @typescript-eslint/prefer-includes | 2 | 1 | 3 |

All origin branches were fetched and their distinct stage1 trees searched before
this claim. No implementation or claim for these three rules was found. Existing
inventory entries, count registrations, configuration and skip-types evidence do
not constitute ports. No rule was skipped.

## Continuation claim

The original three ports are completed and pushed in 8cfa5d02. After fetching all
origin heads, the next three unclaimed checker-dependent rules in the combined
VOLUME_REPORT.md full count tables are:

- nexus/correctness-no-process-exit-after-output (combined volume 0)
- nexus/correctness-no-uncleared-race-timeout (combined volume 0)
- nexus/correctness-require-blocking-standard-streams (combined volume 0)

Selection inspected production Adamic source on origin/codex/tsgo-c-library and
origin/main, and every distinct Markdown claim blob under
stage1/cohere/typeaware/claims on all origin branches. The ranking has 197 checker
rules; 25 base ports occur in it, 96 remaining names are mentioned in claim files,
and 76 candidates remain. These are the first three. No implementation precedes
this claim commit and push. Shared harness and registration generator remain
outside this continuation's edits.

Continuation status: all three continuation rules are now fully ported and
native-validated after rebasing onto origin/main e8ba3d5d. The original three
rules also pass their native oracle again. Process-output and blocking-streams
match Go bytes over 92 controls and compiler77/repository287, normal and under
sanitizers, with one clean-exit native mutant each. Timeout and released-handle
checks pass. The earlier internal/fresh blocker is closed on this main.
See ../wave_10_next/LANDING_REPORT.md and evidence/landing for observations,
commands, timings and limits. No shared harness, generator or protected compiler
file was edited. These six claims remain reserved through integration.

## Second continuation claim, after landing validation

All six earlier claims are ported, native-green on origin/main e8ba3d5d and
pushed on this branch in 95d3ced8 before this selection. The next three names
in VOLUME_REPORT.md's combined by-volume ranking that are neither ported on
main/base nor claimed under the typeaware claims directory on any origin head:

- react/forbid-elements (combined volume 0)
- react/forbid-prop-types (combined volume 0)
- react/iframe-missing-sandbox (combined volume 0)

The post-landing fetch scanned 465 origin refs, 33 distinct claim blobs and all
197 checker-ranked rules. The base/main implement 25 ranked rules; 151 names
are mentioned in existing claim files, leaving 21 available. Names stored without
the @typescript-eslint prefix by the original native suite were normalized before
selection. Checking these three names across 129 distinct origin stage1 trees
found only config/sets.ts inventory references, not implementations. None was
skipped. This claim is committed and pushed before implementation work.

Second continuation status: PARKED, reserved but unported. A minimal valid TSX input
produces one production Go iframe finding, while the unchanged native stage-1
parser exits 70 expecting GreaterThanToken at the self-closing slash. Shared
JSX parsing blocks these rules before listener dispatch. The current parking instruction counts these claims as finished for the
landing cap while preserving their incomplete implementation status. Blockers:
shared native JSX parsing and incomplete rule adapters.
JSX is landing on area/stage1-lint; analysis modules are tracked on #dnv6f2c.
No shared parser is edited.
See ../wave_10_react/README.md for the input, exact error, Go verdict and limits.
The continuation below is authorized by the React parking instruction.

## Third continuation claim after React parking

Landing-ready source and evidence are pushed in 6d4e8cfc, rebased on current
origin/main f8013f0b. All six completed ports pass fresh native Go-byte oracles,
sanitizers, released-handle guards and mutants. The three React claims above
are explicitly parked and count as finished for the work-in-progress cap.

The next eligible checker-dependent rules in combined by-volume order are:

- require-await (combined volume 0)
- structure/react-hook-no-any-type (combined volume 0)
- symbol-description (combined volume 0)

All 528 origin refs were fetched. Scanning 33 distinct claim blobs across the
197 ranked names leaves 18 unclaimed rules. The base/main ranked ports remain
unchanged. The thirteen react/ candidates are skipped under the React parking
instruction; the atomic-updates candidate imports the control_flow_graph module and
is skipped because it needs that analysis. require-await uses AST/checker
facts; structure/react-hook-no-any-type listens to CallExpression and uses
checker/React binding facts without JSX or high-level analysis; symbol-description
uses call syntax and declaration origins. No production port was found for
these names on main/base. This claim is pushed before implementation.

New modules will be .a, in owned rule directories, with rule.json kinds and
supplied-node listeners. Byte agreement, native mutants, sanitizer/release checks
and native/Go timing remain required; a claim is not a completion claim.

Third continuation status: IN PROGRESS, not complete or parked. Canonical kind-name
rule.json metadata and isolated symbol/hook verdict modules are written in wave_10_leaf.
Their supplied-fact Go comparisons and sanitizers pass. require-await's body
filters pass context-free Go controls; candidate reporting explicitly refuses.
The handed-node contract is specified by harness ab70f38d4; live bridge/corpus
adapters are not implemented or validated. Numeric kind minting is not planned. Contextual require-await reporting remains unported. See
../wave_10_leaf/README.md for exact observations and limits. No further claim.

Latest landing base: c01907a7. Six completed-rule oracle gates and the isolated
new-module gate are green after the clean rebase; source tip 39310513 precedes
this evidence commit. React claims remain PARKED. The three non-analysis claims
remain IN PROGRESS, with no additional reservation. See wave_10_leaf's current
landing section for commands, comparisons, mutants, timings and limits.


Continuation after the canonical-kind and regex corrections: symbol-description
and structure/react-hook-no-any-type now have live handed-node native adapters.
Each passes full Go finding/fix/suggestion bytes on its positive controls and
compiler77/repository287, normal and sanitized, with live clean-exit mutants and
released-handle refusals. require-await has named-kind body filters and a separately
validated 32-case native reporting module, including suggestions. Its full
contextual/generic/heritage contract path remains unimplemented; the bridge does
not yet expose resolved-signature target/type-parameter facts for declared-demand
replay. Shared-registry/context integration also remains pending. No further
reservation is made. Exact final gates and timing observations are in
../wave_10_leaf/README.md and evidence/live-continuation.


Current landing base is b8fb957a, after main advanced during final validation.
Rebase is clean; all six completed-rule oracle gates and all new isolated/live/
reporting gates pass again with the current compiler. A focused uncached RegExp
Node/native/emitted-JavaScript oracle passes too. The two new live native rules
have full corpus-byte validation; require-await remains IN PROGRESS for its
unimplemented contextual/generic/heritage decision and bridge facts. No further
claim is taken. See the current landing section in wave_10_leaf/README.md.


Latest continuation status: all three non-analysis rules are fully ported and native-tested on origin/main 39638d9e2. require-await's contextual/generic/heritage decision and dedicated raw bridge question are implemented. All three match full Go findings, fixes and suggestions on positive controls and compiler77/repository287, normal and sanitized; require-await also matches 87 upstream source controls. Native mutants and released handles pass. All six earlier ports are re-green on this base; the three earlier React claims remain PARKED with their named blockers. Shared registry/context wiring remains with the harness worker. No additional reservation precedes the completion push. See ../wave_10_leaf/README.md latest completion section.

# Opaque solution-builder timer slot

Base main is `de5c28cc11d8a12a76cefb13e2bb1c78a2bb4585`, fetched again before
this follow-up. It has not moved since the stopped five-member measurement.
This change is on top of `f160842e52675fe42ad442cfe80d71b93351b44f` on
`typescript/adapt40-any5`. Upstream remains TypeScript 6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`; the adapter and proofs use the
pinned stock 6.0.3 API from stage3/api, on Node v24.19.0.

`SolutionBuilderState.timerToBuildInvalidatedProject`, tsbuildPublic.ts:424,
now uses unknown. The three writes are undefined at initialization, the host
setTimeout result, and undefined at reset. Reads are two truthiness tests and
forwarding to the host's any-input clearTimeout. No reading expression changes.
A NodeJS.Timeout-only contract is not proved for this host-polymorphic slot.
The other four measured members remain any, for the reasons in ANY5.md.

## Rule and guards

The new `opaque-timer-slot` class is in class-rules.json and timer-slot.cjs,
called through classes.cjs by adaptation 40. Before its sole possible write,
stock TypeScript resolves the member and every matching compiler reference,
checks the contextual initialization owner, the six exact use shapes and their
enclosing functions, the registration label, and the host registration return
and cancellation input types. It rejects optional, modified or newly public
member shapes and accepts only any or the already-adapted unknown annotation.
Both comments-retained and comments-removed transpilation must be identical.

The new class rerun writes zero files. Real scratch mutants with a string
annotation, an extra read and a changed registration argument each fail before
writing. A scratch any-to-unknown application writes once, and its next pass
writes zero files. See [the guard proof](evidence/timer-slot/mutants.json)
and command logs.

Rerunning the entire old adaptation 40 after all later adapters reaches its
pre-existing brand line-number guard on both control and treatment:
`missing reviewed brands: src/compiler/types.ts`. That exploratory rerun makes
no source change; the timer class is checked directly. The full pipeline passes
with each adapter at its normal position. An exploratory comparison also
incorrectly expected zero diagnostics under the stricter 47 helper flags; the
correct paired check finds the same 140 diagnostics, with no additions or
changed identities. Both normal upstream compiler-and-test builds pass.

## Scope, pins and API

[The scope and emission proof](evidence/timer-slot/scope-emission.json.gz)
compares all 746 source files. The only source difference is
`timerToBuildInvalidatedProject: any` to unknown. All ten emitted JavaScript
artifacts and all 715 emitted declaration files are byte-identical. Source maps
and build metadata are outside that byte-identity claim.

SolutionBuilderState is private. Neither it nor the timer member occurs in
built public typescript.d.ts, typescript.internal.d.ts or the API reference.
The public API diff is empty. No new public line is sanctioned, so an API
undo-retype mutant is not required for this private field.

The active 90, 91 and 92 pin manifests have no tsbuildPublic.ts entry. Their
own adapters each rerun successfully with zero files written. No later source
hash or offset pin changes.

Regenerating the API sanction changes only class-rules.json's proof hash.
Normalized declarations, generated diff bytes, owners and counts are identical.
The resulting harness runtime hash requires rebinding the same 398 shard
commands. The plan comparison proves that task assignments, expected counts,
identity digests and all other unit data are unchanged. The generated Go shard
file is refreshed and generate_shards.py --check passes.

## Sharded oracle

| Tree | Accepted units | Passing | Expected failing | Pending | Longest command |
| --- | ---: | ---: | ---: | ---: | ---: |
| Control | 398 | 106,366 | 1 | 0 | 11.553 s |
| Retyped | 398 | 106,366 | 1 | 0 | 11.440 s |

All 106,367 test identities, statuses and multiplicities match. The sole
failure is the existing Public APIs acknowledgement failure. Both API baseline
diffs are byte-identical, SHA-256
`6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`.
Control wall time is 717.488 seconds;
retyped wall time is 716.535 seconds. All memory-limit and OOM counters stay zero.
See [the oracle comparison](evidence/timer-slot/oracle-comparison.json).

Both runs execute every command from the final stage3/lane/shards.json, without extra
filters or command overrides, using the adapted-tree product path. Four units
run concurrently within each phase, each with one upstream worker and a
1,400 MiB V8 old-space limit. Control and treatment phases are sequential, after
all build, adaptation and guard checks finish. Completed private views are
removed only after preserving reports, observations and the byte unit's API
snapshots. The runner's counts, identities, exits, baseline paths, API guard
and 30-second limit are unchanged.

The first control launch hit a preparation error before any test ran: the
build-time node_modules symlink conflicted with the product runner's own link.
Only those two task-created links were removed. That rejected launch is
retained separately and contributes no accepted test observations. Both
accepted phases use fresh result directories and execute all commands again.

## Reproduction

Build fresh control and treatment trees with stage3/apply.sh and their locked
upstream dependencies, then run npm run build on each. NODE_PATH must name
stage3/api/node_modules for the retained TypeScript proof scripts. Run the
scope/emission proof and guard mutants before removing build-time node_modules
links for the normal adapted-tree product-mode shard runner. Every command and
its outcome is recorded in the evidence directory.

The retained coordinator runs the actual manifest commands and preserves every
unit's reports, observations and logs. Its comparison checks all pass/fail
identities with multiplicity, not just totals, and checks identical byte-unit
baseline diff and public API snapshots. Final scope and artifact hashes are
checked again after both oracle phases.

[Post-oracle checks](evidence/timer-slot/post-oracle-check.json) verify source
bytes directly against the apply tree objects, including generated untracked
files, and recheck every emitted artifact hash.
[commands.json](evidence/timer-slot/commands.json) records the preparation and
verification outcomes. Both complete raw oracle outputs are retained as
compressed archives beside the comparison.

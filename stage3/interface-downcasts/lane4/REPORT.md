Built: a verified mixed-union census and transactional union member-graph hook; runtime admission is unfinished.
Commits: bc83664c (territory/hooks), e9b5dce8 (census), dc5a6b8b (metadata/probes), ba8464fa (hook/report), on codex/views-mixed-unions.
Validation: focused contract tests, checked-view oracle and vet pass; the full lower package fails an independently reproduced inherited test.
Mutants: omit-member, accept-unknown and keep-failed-graph were caught by metadata assertions; the three requested runtime mutants remain unrun.
Uncovered: backend union selection/conversion, runtime exit-70 pins, branded intersections, dictionaries and current-main landing.

## Census and push checkpoints

Stock TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8 reports
zero project diagnostics. All 2,936 pinned spans/texts matched. The independent
census cross-checks every site's three mixed-union family bits against lane 1's
ledger. 2,868 sites have mixed-union target-contract dependencies. Traversal,
including its stops at arrays/tuples and callables, matches the original census.
These are overlapping declared-contract dependencies, not runtime read counts or
whole-file lowering successes. See CENSUS.md for the longer table, and
census-summary.json for expanded members and exact witnesses.

| Top exact checker shape | Sites |
| --- | ---: |
| `false \| string[] \| undefined` | 2,863 |
| `false \| VersionPaths \| undefined` | 2,863 |
| `string \| false \| undefined` | 2,863 |
| `__String` | 2,862 |
| `__String \| undefined` | 2,862 |

| Push checkpoint | Remaining only lane 4 | Lane 4 plus others |
| --- | ---: | ---: |
| bc83664c, territory | 0 | 2,868 |
| e9b5dce8, census | 0 | 2,868 |
| dc5a6b8b, metadata and source probes | 0 | 2,868 |
| ba8464fa, hook/report | 0 | 2,868 |
| Final evidence packaging checkpoint | 0 | 2,868 |

The remaining-family arrays are preserved in census-sites.json. Merging array
helpers does not prove that every recursively encountered array contract is now
supported. No family was removed and no new complete source site is claimed.
There are 68 sites without a lane 4 dependency.

## What is implemented

`internal/lower/view_unions_mixed.go` exposes
`viewMixedUnionContractHook = internMixedUnionViewContract`, using the existing
viewContractHook signature and recursive builder. It reserves the union id before
recursion, retains each distinct member contract, rejects zero/unknown child
contracts, and removes all newly reserved ids and memo entries if a member fails.
A failed build cannot later be reused as a certificate. Recursive object members
link back to the same union id. Phantom primitive brands retain their refusal.

Focused tests cover string/object, number/object, object/readonly-array,
boolean/object, two untagged objects, recursive object members, unknown child
contracts and failed-build retries. This is descriptor construction, not compiler
admission. Shared dispatch does not call the new hook yet.

Six `.a` probes cover valid members and malformed reads of the first three ranked
shapes. Source Node prints the values, including malformed 42 and bad:42. Both
backend commands still refuse at compile time: mixed payload fields say NotYet;
VersionPaths is refused for its index signature. Exact outputs are recorded in
probe-observations.json. No emitted native or JavaScript execution, exit 70, or
required runtime-mutant result is inferred from those compiler refusals.

## Concrete owner handoffs

The full requests and territory reservation are in docs/checked-views-plan.md,
under Lane 4. The first hook is now declared and implemented in the lane-owned
file; lane 1 need only route eligible union construction to it. Admission must
wait for all required read paths.

Lane 1 owns the remaining changes: route mixed/untagged union field registration;
allow complete union contracts through the current representation/alias gates;
route union casts through view(); and wire both backend read dispatchers to
selection, conversion and retained member contracts. The native prerequisite is a
non-panicking readiness-aware slot probe returning presence, initialization,
logical kind and normalized payload, with lane 1's null/undefined distinction.
It must respect the existing owner/static/accessor behavior. Current
adamic_object_view panics on the first failed member, so it cannot try untagged
alternatives. A heap-object tag alone cannot prove structural membership.
Lane 4 cannot safely implement this by copying object.c's readiness machinery.

The exact top five require more than those hooks. Upstream __String contains
`string & { __escapedIdentifier: void }` and
`void & { __escapedIdentifier: void }` alongside InternalSymbolName literals.
The original census classifier calls intersections primitive category "other";
it does not establish that runtime string checks certify those brands.
VersionPaths.paths is MapLike<string[]> and requires dictionary contracts.
Conservative assumption: retain intersection/brand and dictionary refusals until
their contracts or allocation certificates are supplied. Neither is silently
replaced with an unbranded string or an empty object contract.

The full lower package fails TestSharedArrayContractAdapter for
`readonly (number | string)[]`: lane 1's adapter test expects metadata, while the
merged lane 2 builder refuses mixed/optional boolean elements. The same filtered
test fails on isolated baseline bc83664c, before either lane 4 Go file exists.
Lane 2/lane 1 should reconcile descriptor construction with that admission gate;
no lane 2 file was edited here.

## Validation and actual mutants

Setup used GOPROXY='https://proxy.golang.org|direct' before bash cloud/setup.sh.
Timing lines: Node 0.087s, Go 0.113s, submodules 0.120s, clang 0.737s, markdown
1.388s, build ready 68.078s, cache warm 68.206s, done 68.321s. nproc is 5;
cgroup CPU quota is 4 cores. Go 1.27.1, clang 20.1.8, Node 24.19.0.
The printed /workspace/adamic-tools/env.sh was sourced in every build/test shell.
Setup succeeded; /opt/adamic-tools/env.sh was absent, so the printed path was used.

All test invocations wrote to logs, without piping their runs:

```text
go test ./internal/lower -run TestMixedUnionContract -count=1 -timeout 10m
  PASS, final logs/contracts.log
python3 stage3/interface-downcasts/lane4/run-contract-mutants.py
  exit 0, all three metadata mutants caught, logs/contract-mutants.log
go test ./internal/lower -count=1 -timeout 30m
  FAIL, 29.306s, inherited TestSharedArrayContractAdapter, logs/lower.log
go -C /tmp/views-mixed-baseline test ./internal/lower -run TestSharedArrayContractAdapter -count=1 -timeout 10m
  FAIL, 0.161s, same failure, logs/baseline-lower.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestCheckedView -count=1 -timeout 30m
  PASS, 34.134s, logs/oracle.log
python3 stage3/interface-downcasts/lane4/observe-probes.py
  exit 0, six Node controls and twelve compiler-refusal observations, logs/probes.log
go vet ./...
  exit 0, no diagnostics, logs/vet.log
gofmt -l internal/lower/view_unions_mixed.go internal/lower/view_unions_mixed_test.go
  no output, logs/format.log
git diff --check
  no output, logs/diff-check.log
```

| Actual metadata mutant | Intended catcher |
| --- | --- |
| Omit member ids from the union | TestMixedUnionContractGraph: union lost a member contract |
| Accept ViewUnknown child ids | TestMixedUnionContractUnknownMemberFails: unavailable member certified |
| Keep the failed provisional graph | TestMixedUnionContractFailureDoesNotCertifyRetry: failed graph remains memoized as certified |

Each mutant exited 1 with its intended assertion, and the runner restored the
source independently. No build failure, clang warning or sanitizer kill counts.
These are not substitutes for skip-runtime-member-check, take-first-member or
drop-transitive-object-check. Those three requested mutants cannot yet run on
admitted lane 4 programs and remain outstanding. The full repository gate and
counts regeneration were not run; no new runtime oracle fixture row was added.

## Integration and delivery date

Lane 2 0b141c26 was merged as 4d41b66f before implementation. Final remote checks
still show lane 1 b4cfd1aa and lane 2 0b141c26, with no newer pushes to merge.
Every push went only to codex/views-mixed-unions. No PR was opened.

Landing was attempted against current origin/main
48c05d091f0a43c31cbe051b1d6578d99eeedf19. Conflicts are in
`docs/escape-hatches.md`, `internal/lower/cast.go`,
`internal/lower/expression.go` and `internal/lower/refusals.go`. They are outside
lane 4 territory. The merge was aborted; logs/main-merge.log and
main-merge-conflicts.patch.gz preserve the exact conflicts for their owners. The patch is compressed to keep
its significant diff-prefix whitespace out of source whitespace checks.
This checkpoint is pushed, but not landed or fully green.

There is no defensible completion date for the exact five ranked shapes until
shared dispatch, dictionary support and a sound brand-certificate policy exist.
The provisional October 9 at 22:00 UTC estimate applies only to reifiable union
selection and reduced contracts, conditional on shared hooks/nullish support by
October 8 at 22:00 UTC. It is not a promise for VersionPaths's dictionary or the
two __String shapes. Those dependencies must be resolved before a complete
step-3 date can be committed to. No unattended future work is scheduled.

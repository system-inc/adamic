Built lazy target descriptors and shared-flow refusals at unsupported reads.
Pushed implementation: de8b2ec7 and f7100da7; base merges: 592f1f71 and c01ae313.
Focused lower/oracle tests pass; exact census matches 2,936 spans, production compile count is 0.
Direct, generic, callback, stored-field, missing-name and demand-deletion mutants are caught.
Full integration gate and a checker-clean whole-tsc allocation-flow family census remain uncovered.

Lazy admission records unsupported contracts as metadata. It does not change executable IR
representations or remove backend checks. The existing shared allocation graph joins helpers,
returns, callbacks and projected stores. Its Unknown frontier retains checks/refusals. The
read pass runs before readiness erasure. Wider helper interface ids retain unsupported field
families through a conservative name fallback. Descendant objects and arrays retain view
provenance; unsupported contracts cannot be write certificates.

All supported field reads retain the existing conservative program-wide checks, except checks
that the shared shape pass proves redundant. Wired array reads, map/visit/reduce callbacks,
pop, join and for-of demand their element contracts. Unwired consumers still fail closed.
Syntax reads that cannot be represented in IR retain conservative preflight refusals at the
read, naming the field and family. This is deliberately conservative and does not claim
perfect separation of ordinary values from views for every unsupported syntax family.

`unread.a` and `unread-untagged.a` print `okok` on Node, native sanitized, native release and
JavaScript. Their unread callable field does not exist in the source allocation. `unread-mixed.a`
prints `true` despite its unread mixed primitive member. The optional Error probe remains a limitation: Node prints `missing`, while the native
read exits 70, naming `code`, expected `string | undefined` and found unsupported
representation. It is not counted among the backend-matching positive fixtures.

The four callable source mutants all refuse at line 9, the `opaque` read, naming `callable`.
Deleting `if demanded`'s guard makes helper, wider-interface, Unknown and every wired array
consumer control fail, along with all four source-mutant controls. The mutation was restored.
The missing-name source mutant prints undefined on Node; native sanitized, release and JS
exit 70 at `value.name`: `expected string, found missing`.

The first lane merges preceded the user's integrator-only coordination change. No later lane
branch was merged. The integration branch was not advertised by origin when checked. This
branch is pushed for the integrator; it is not described as landed or fully green.

## Census observations

Pinned upstream: TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8, accessed
through the existing TypeScript submodule's git objects and a detached scratch worktree.
Stock checker diagnostics: zero after upstream diagnostic-table generation. The original
4,101-site ledger and 1,758-site tagged refinement select exactly 2,936 unchanged cast spans.
Native parser byte endpoints are converted to the ledger's UTF-16 endpoints and texts checked.

| Measurement | Tagged | Untagged | Total |
|---|---:|---:|---:|
| Lazy descriptor admission | 1,758 | 1,178 | 2,936 |
| Unchanged whole-project production compilation | 0 | 0 | 0 |

Descriptor admission is not cast lowering. Production loading stops before lowering, including
TS2412 under exactOptionalPropertyTypes and TS1484 under verbatimModuleSyntax. No checker
error was bypassed. Therefore this run cannot certify per-cast production read-flow dispositions.
The exact diagnostic witness and measured counts are in census/admission-summary.json.

The rerun of lane 2's stock read-demand inventory finds 67,118 explicit source reads, 25,848
static candidate reads and 2,825 distinct candidate (type, field) pairs. These are overlapping
family obligations, not successful-lowering counts or closed-world allocation-flow evidence.
Partial scalar/object/optional/array/callable/union support is not assumed complete. Remaining
unsupported families include intersections, mixed and object/primitive unions, dictionaries,
any, nominal class fields, tuples, unknown and generic contracts. Opaque callable signatures,
nullable representations and unwired array consumers also remain conditional blockers.

| Read-contract family | Distinct pairs | Candidate reads |
|---|---:|---:|
| object or interface contracts | 1101 | 11569 |
| scalar contracts | 982 | 8839 |
| nullish members | 865 | 6276 |
| optional properties | 754 | 5728 |
| array contracts | 334 | 3189 |
| object union | 228 | 2468 |
| array element or consumer reads | 251 | 1602 |
| callable contracts | 308 | 1503 |
| array intrinsic contracts | 179 | 794 |
| intersection field contract | 45 | 769 |
| mixed primitive union | 23 | 538 |
| object plus primitive union | 42 | 181 |
| dictionary contracts | 18 | 174 |
| any field contract | 55 | 104 |
| array property reads | 30 | 72 |
| dictionary reads | 14 | 66 |
| nominal class fields | 1 | 42 |
| tuple contracts | 6 | 9 |
| unknown field contract | 1 | 2 |
| generic field contract | 1 | 1 |

## Reproduction and validation limits

All test output was written to logs, never piped. Completed evidence is under census/.

```
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
go test ./internal/lower ./internal/oracle \
  -run 'TestCheckedView|TestView|TestLazyView|TestSharedArrayContractAdapter|TestNodeFSFileQualifiedErrorType|TestNodeLibraryQualifiedTypeAssertion' \
  -count=1 -timeout 10m > /tmp/lazy-final-focus.log 2>&1
NODE_PATH=stage3/fixtures/assertions/api/node_modules node \
  stage3/interface-downcasts/blocking-families.cjs /tmp/lazy-typescript \
  stage3/fixtures/assertions/ledger.json stage3/interface-downcasts/other-locations.tsv \
  /tmp/lazy-census > /tmp/lazy-blocking-census.log 2>&1
NODE_PATH=stage3/fixtures/assertions/api/node_modules node \
  stage3/interface-downcasts/lane2/view-read-census.cjs /tmp/lazy-typescript \
  /tmp/lazy-census/blocking-families-sites.json /tmp/lazy-census \
  > /tmp/lazy-read-census.log 2>&1
LAZY_VIEW_CENSUS_ROOT=/tmp/lazy-typescript \
LAZY_VIEW_CENSUS_SEEDS=/tmp/lazy-census/blocking-families-sites.json \
LAZY_VIEW_CENSUS_OUTPUT=/tmp/lazy-census/admission-summary.json \
go test ./internal/lower -run '^TestLazyViewCensus$' -v -count=1 \
  > /tmp/lazy-admission-census-complete.log 2>&1
```

Setup finished in 1,059.576s on 5 processors: Go 0.292s, Node 0.302s, clang 0.966s,
Markdown 1.193s, submodules 693.572s; Go build/cache readiness 1,059.536s elapsed.
Pinned stage3 Node declarations were installed separately for package-wide tests.

The complete lower package run still fails pre-lazy merged-baseline assertions for predicate
markers, overload diagnostics, phantom arrays and nested functions. A Go source overlay at
c01ae313 reproduces those failures. Two eager optional-Error compile-time assertions were intentionally updated. The
optional Error source probe records the native representation refusal at the read. No claim of a fully green repository gate is made.
An automatic approval review rejected an earlier proposed executable-type rewrite and empty-tag
bypass; that rewrite never ran. The implemented alternative preserves executable types and
backend cast checks. No rejected action remains pending.

The final focused run passed: lower 9.503s, checked-view oracle 33.735s. The complete
JavaScript package passed. The complete native package finished in 388.451s and failed
five graph-region tests: TestGraphRegionsMillion, TestGraphContainerBoundary,
TestGraphLazyRegions, TestGraphClosureEnvironment and TestGraphRegionsRuntime.
The failures include an AddressSanitizer map-set use-after-free and leaks. All five
were reproduced with the pre-lazy c01ae313 source overlay (native baseline 1.116s).
The runtime was not changed to suppress these failures. Logs are saved in census/.
Integration remains blocked on these baseline failures and production checker policy;
no final integration SHA was available from origin at the last branch check.

## Optional Error follow-up (October 7)

`new Error('plain') as NodeJS.ErrnoException` now reads absent `code` as
undefined, matching Node's `missing`. The native Error producer owns only
name and message, and certifies their string tags; it no longer invents an
initialized optional code slot with an unknown representation. The own-message
probe prints `plain:missing` in Node, sanitized native, release native and JavaScript.
This covers absent optional strings and present string producer evidence, not
arbitrary nullable unions. Null remains invalid for `code?: string`.

The number and null payload mutants both stop at the code read with exit 70,
naming code, string and the found number/null in all three compiled runs.
A temporary implementation mutant treating null as undefined was caught by
`TestCheckedViewLazyOptionalCodeMutants/null`: native ran on and printed missing.
The runtime was restored. Focused uncached lazy oracles passed in 4.098s.
The broader class-inheritance exception oracle still fails on an unrelated
boolean-array view; restoring the original Error runtime reproduces it.
Logs: `census/nullish-pass.log`, `census/null-collapse-mutant.log`,
`census/error-baseline.log`.

The adapted-tree follow-up supersedes the stock-tree production census: see
[ADAPTED-CENSUS.md](ADAPTED-CENSUS.md) for exact cast counts and all remaining
checker diagnostic rows. The designated integration tip is now merged.

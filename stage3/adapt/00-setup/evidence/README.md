# Checker stops, 2026-10-19

Measured main `2b22b31ccc8b218ac564fba8f8641adb63160e6f` with stock
TypeScript **6.0.3**, against the completed `stage3/apply.sh` tree. No source
repair is made, including no Set/Map iterator type arguments. The evidence is
[checker-stops-2026-10-19.json](checker-stops-2026-10-19.json).

The required pair uses Adamic's ES2024 library view, including the four
`regexpLibraryFS` declaration rewrites reproduced in scratch. The checker and
CLI JavaScript are stock. A second pair uses untouched stock declarations.
Both pairs produce exactly the same diagnostic lists and full message chains.
`types: []` excludes automatic Node types; none occur in the recorded graph.
The prelude variant removes **only** the two MapIterator/SetIterator `next`
declarations, replacing their characters with spaces to preserve positions.
The original prelude and repository library files remain untouched.

| Run | Diagnostics | CLI exit |
|---|---:|---:|
| Stock libraries, no prelude | 234 | 2 |
| Stock libraries, filtered prelude | 235 | 2 |
| Adamic library view, no prelude | 234 | 2 |
| Adamic library view, filtered prelude | 235 | 2 |
| Stock libraries, no prelude, noUncheckedIndexedAccess false | 135 | 2 |
| Stock libraries, no prelude, exactOptionalPropertyTypes false | 166 | 2 |
| Stock libraries, no prelude, useUnknownInCatchVariables false | 230 | 2 |
| Adamic library view, original prelude mutant | 319 | 2 |

The **46 sys.ts TS2591 diagnostics are included in 234**, not additional to
it. The filtered prelude removes missing console and adds two diagnostics:
`commandLineParser.ts:2957:13 TS2322` for JSON.stringify's optional result and
`core.ts:1637:11 TS2740` for the custom Set's seven missing ES2025 methods.
Neither required run reports builder.ts:1246.

| Class | No prelude | Filtered prelude | Judgment and existing adaptation |
|---|---:|---:|---|
| Unchecked indexed reads, narrowing and overload fallout | 99 | 99 | #k881crd; review 30/31/32/33 obligations individually |
| Optional storage, callback and structural value contracts | 61 | 61 | Truthful declaration candidates, 20 plus owning partition; 49 for private lifecycle storage |
| Cascades after rejected result/watcher assignments | 6 | 6 | Same optional owner candidates, 20/32; remeasure after owner repair |
| Unknown caught host values | 5 | 5 | #k881crd and 47; a guard changes JavaScript |
| sys.ts missing Node names, TS2591 | 46 | 46 | Node-typing gap, 65/setup |
| Other Node/V8 declarations and contextual typing fallout | 15 | 15 | Node-typing gap, 65/setup |
| Optional source-map-support declaration | 1 | 1 | Host dependency gap, 65/setup |
| Console host declaration | 1 | 0 | Supplied by existing prelude |
| Prelude JSON.stringify result | 0 | 1 | Compiler library contract, #k881crd; 32 owns caller |
| Custom Set versus prelude extensions | 0 | 1 | #k881crd; existing 30 whole-sites decline |
| **Total** | **234** | **235** | **No newly proven tsc type lie** |

These are judgments for dispatching waves, not proof that a blanket `!`, cast,
or optional-owner widening is truthful. Each unique diagnostic has an exact
file, line, UTF-16 column, code, complete emitted message chain, related
information, source line, indexed-read partition, class and run membership.
Each class records its reasoning and adaptation candidates. No new Node
counterexample was established, so no upstream ledger entry is claimed.

The six cascades are moduleNameResolver.ts:642,644,647,648 and
resolutionCache.ts:1282,1297. They disappear with exact optional properties off;
the rejected result/watcher construction is the earlier obligation. One
unchecked-read diagnostic also disappears with exact optional properties off,
so option deltas overlap: 99 and 68 cannot simply be added. Disabling unknown
catch variables removes all five catch diagnostics but exposes a new optional
nodeSystem object diagnostic at sys.ts:1524:15, giving a net reduction of four.
Both the new control diagnostic and every original-prelude-only diagnostic
are included and classified in the JSON, outside the primary count columns.

There are **79 roots**: the pin's 77 compiler TypeScript files, the generated
diagnostic map, and adaptation 47's hostErrors.ts. Each required graph contains
those roots plus 69 ES2024 declaration files, with the filtered prelude adding
one file. The JSON retains each exact root and the SHA-256 of every loaded
source/library file. All 79 root hashes match across all eight completed runs
and the final adapted tree. Compiler options match `internal/load/load.go`,
including strict mode, noUncheckedIndexedAccess, exactOptionalPropertyTypes,
verbatimModuleSyntax, force module detection, bundler resolution, noEmit and
ES2024. Strict mode includes strictBindCallApply and unknown catch variables.
No upstream tsconfig or implicit Node/DOM libraries are substituted.

Reproduction uses the pinned upstream checkout and API install described by
`stage3/source.json` and `stage3/api/package-lock.json`:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/onediag-typescript-source
timeout 180 bash stage3/apply.sh /tmp/onediag-adapted > /tmp/onediag-apply.log 2>&1
timeout 360 node stage3/adapt/00-setup/evidence/checker-census.cjs "$PWD" /tmp/onediag-adapted /tmp/onediag-census > /tmp/onediag-census.log 2>&1
timeout 150 node stage3/adapt/00-setup/evidence/checker-census.cjs "$PWD" /tmp/onediag-adapted /tmp/onediag-census --option-controls > /tmp/onediag-option-controls.log 2>&1
timeout 30 node stage3/adapt/00-setup/evidence/verify-checker-census.cjs > /tmp/onediag-verify.log 2>&1
```

Use a new adapted-tree directory for apply. Every run also stores its exact
stock `tsc.js --project ... --pretty false` command and exit in the JSON;
project files are generated with an explicit file list. Each CLI invocation
has its own 60-second limit and writes a complete log. The stock API then
serializes the diagnostics, requiring its formatted output to match the CLI
log byte for byte. The committed CLI logs let the verifier check table
completeness without running the compiler again. Classifications are reviewed
judgments saved in the table, not generated from error codes by the driver.

The restored-next mutant adds **84 diagnostics**, removes none, and reproduces:

```text
src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

Applying the filtered-census exclusion check to that real mutant run exits 1.
Four independent evidence-input mutants also exit 1: omit a diagnostic (caught
against CLI bytes), change its class (class counts), omit a loaded root (source
coverage), and insert Node types (Node exclusion). Exact commands, exits and
catchers are retained in `checker-validation.json` with their logs. These
metadata mutants test evidence validation; only restored-next changes checker
input. No repair erasure proof or baseline oracle is claimed for this
measurement-only change. The meter and whole Go gate were not run.

An early census overlapped apply and failed CLI/API comparison. It was discarded;
all eight retained runs started after apply completed. Setup and stage3 apply
test outcomes are recorded in `checker-validation.json`.

Setup first reached its 600-second limit after the submodule fetch consumed
483.940 seconds. The bounded retry completed successfully: Go 0.023s, Node
0.025s, submodules 0.067s, markdown dependencies 0.075s, clang 0.155s, shared
cache 0.823s, Go build 168.641s, cache warm 168.719s, done 168.750s.
`nproc` is 5; the cgroup quota is four CPUs. Node is v24.19.0, Go is 1.27.1,
and clang is 20.1.8. The environment file is `/workspace/adamic-tools/env.sh`.
All 23 stage3 apply tests passed in 16.142 seconds with a 90-second hard limit.

Main advanced during final verification to
`141fec7702d4aed21429724b2afd7efcd980f277`. It was merged without conflict.
An exact Git diff proves all census inputs, loader declarations, pinned API
inputs and apply tests are byte-identical to the measured base. The full
23-test apply suite passed again after the merge in 14.482s, and the evidence
verifier passed. The table preserves the original measurement base and records
this separate main-tip verification; the census was not rerun or relabeled.

## Optional-value contracts, wave 1

The [62-row split](optional-contracts-2026-10-19.json) preserves exact census IDs,
positions and codes. It follows the innermost incompatible slot in each message,
with every declaration owner of that slot and its AST form. Anonymous type
literals are identified as such. Public membership is checked against the actual
`built/local/typescript.d.ts` from a fresh, locked upstream build, rather than
assuming that an exported source declaration is public. Owner positions refer
to the completed apply tree; the adapter ledger retains its earlier positions.

The [fresh adaptation 20 report](optional-contracts-20-before-2026-10-19.json)
retains its complete `declined`, `initially_declined` and `declined_owners` lists.
Each split row includes its matching entries, or explicitly says it is outside
the seed code set or absent at adaptation 20 time. The adapter's broad reason
is preserved separately from the reviewed explanation of the resolver gap.
Counts are diagnostic rows, so the TS2375 and TS2379 at resolutionCache:1332
count twice; multiple declaration owners do not multiply a row.

| Reason | Count | Visibility | Form |
|---|---:|---|---|
| nullable-return | 1 | internal | property |
| nullable-argument | 5 | internal | property |
| nullable-assignment | 5 | internal | property |
| generic-owner | 4 | internal | property |
| indexed-read | 1 | internal | property |
| generic-receiver | 1 | internal | property |
| required-view | 1 | internal | property |
| nested-expression | 1 | internal | property |
| nested-generic-result | 1 | internal | property |
| late-exposure | 2 | internal | property |
| unsupported-code | 2 | public | property |
| unsupported-code | 9 | internal | property |
| optional-method | 22 | public | method |
| optional-method | 7 | internal | method |

There are 29 method rows (22 public, seven internal). Their syntax and variance
remain unchanged. The property groups contain 33 rows. The smallest selected
resolver group is the one nullable contextual return row,
`D4012d5810140`, checker.ts:6035:9 TS2375, targeting internal
`SymbolVisibilityResult.aliasesToMakeVisible`. Other resolver gaps are deferred
to keep this wave bounded. The generic and required-view groups retain their
existing rule boundaries.

Method slots for the language lead, three examples from each visibility group:

- Public: 22 rows. `src/compiler/program.ts:472:11` (CompilerHost.createHash); `src/compiler/program.ts:4956:9` (ModuleResolutionHost.directoryExists); `src/compiler/program.ts:4957:9` (ModuleResolutionHost.getDirectories).
- Internal: 7 rows. `src/compiler/checker.ts:54261:5` (ModuleSpecifierResolutionHost.readFile); `src/compiler/program.ts:1876:11` (ModuleSpecifierResolutionHost.directoryExists); `src/compiler/program.ts:2599:9` (EmitHost.createHash).

Validate the split and actual owner/API locations with:

```sh
timeout 60 node stage3/adapt/00-setup/evidence/verify-optional-contracts.cjs stage3/adapt/00-setup/evidence/optional-contracts-2026-10-19.json /tmp/optional-before > /tmp/optional-split-check.log 2>&1
```

This classification is committed before the resolver repair. Repair measurements,
mutants and toolchain results follow below.


## Optional-contract wave 1 repair proofs

The [proof report](optional-contracts-wave1/proof.json) and
[exact census deltas](optional-contracts-wave1/optional-census-delta.json)
record the final outcome. Six complete CLI/API-backed runs each remove exactly
`D4012d5810140` (checker.ts:6035:9 TS2375), add none, and preserve every remaining
full diagnostic and related-information chain. The required no-prelude run is
234 -> 233; the filtered-prelude run is 235 -> 234. The full resolver-revert
apply and census restore the original counts, exact diagnostic sets and every
original source hash in all six runs. This leaves 61 of the original 62 rows.

The smallest selected group is the internal nullable-return row. Resolving the
single object arm exposes three internal property owners at adaptation 20,
with five `resolved_findings` rows. Two unions are already supplied by later
adaptations. The final source delta is **one line in types.ts**, adding undefined
to `SymbolVisibilityResult.aliasesToMakeVisible`, preserving its question mark.
Stock `transpileModule` emits exactly the same 139,659 JavaScript bytes for this
file. `check-wave1.cjs` checks the entire source file set, changed-file bytes,
all three new ledger owners, and their exclusion from the actual built API.
The emitted `typescript.d.ts` files are byte-identical before and after.
The real stage3/lane `api_check` reports 222 composed and 222 sanctioned
declarations, with 194 added and 193 removed diff lines; the API stays unchanged.

The [retained directory](optional-contracts-wave1/) contains full raw census
JSON, every CLI stream, full adaptation reports, actual emitted API snapshots,
and full apply logs compressed with deterministic gzip. Other proof logs are
plain text. `before.json.gz`, `after.json.gz` and `mutant.json.gz` are fresh
observations, not relabeled copies of the initial evidence table.
`verify-optional-repair.py` checks baseline identity against that original table,
CLI log SHA-256, the exact one-ID delta and absence of new diagnostic chains,
source-file coverage, unchanged options, and full revert identity.

Every mutant run in this unit:

| Mutant | Observed catcher |
|---|---|
| Omit one split row | Split verifier exits 1 on row count. |
| Flip a split owner's public flag | Split verifier exits 1 on group/API classification. |
| Change an owner line | Stock AST owner lookup exits 1. |
| Revert the resolver to split commit e1c0cb0d | Focused witness exits 1 (zero owners instead of one); full fresh apply and census restore D4012d5810140 in all six runs. Using that real result as the repaired observation fails the delta verifier. |
| Append a runtime expression to the witness | Transpile byte comparator detects inequality. |
| Append a runtime expression to the changed compiler file | The same equality assertion throws; the proof runner requires that failure. |
| Omit the errorModuleName resolved finding | Owner-to-finding ledger check exits 1. |
| Convert public ModuleResolutionHost.directoryExists to a property | The actual lane check rejects the unsanctioned API declaration. |

The census driver also runs its existing original-prelude iterator-next mutant:
319 before, 318 repaired, 319 reverted. Its complete diagnostic IDs are retained;
this is a separate pre-existing control, not the resolver-revert proof.

Commands run from the repository root, with each output retained in a log:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh > /tmp/optional-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/optional-source
timeout 240 bash stage3/apply.sh /tmp/optional-before > /tmp/optional-apply-before.log 2>&1
timeout 180 npm ci --prefix /tmp/optional-before --no-audit --no-fund > /tmp/optional-upstream-install.log 2>&1
timeout 300 npm run build --prefix /tmp/optional-before > /tmp/optional-build-before.log 2>&1
# Move locked dependencies out of the census graph after building.
# /tmp/optional-deps/node_modules preserves Node's package-directory layout.
timeout 360 node stage3/adapt/00-setup/evidence/checker-census.cjs "$PWD" /tmp/optional-before /tmp/optional-census-baseline > /tmp/optional-census-baseline.log 2>&1
# Split committed as e1c0cb0d before changing the resolver.
timeout 240 bash stage3/apply.sh /tmp/optional-after > /tmp/optional-apply-after.log 2>&1
# Build with the identical locked dependencies, then remove their temporary link.
timeout 300 npm run build --prefix /tmp/optional-after > /tmp/optional-build-after-final.log 2>&1
timeout 360 node stage3/adapt/00-setup/evidence/checker-census.cjs "$PWD" /tmp/optional-after /tmp/optional-census-after > /tmp/optional-census-after.log 2>&1
PYTHONPATH="$PWD/stage3" PYTHONDONTWRITEBYTECODE=1 timeout 90 python3 -m unittest stage3/test_apply.py > /tmp/optional-apply-tests-final.log 2>&1
timeout 90 node stage3/adapt/20-optional-declarations/check-nullable-return.cjs > /tmp/optional-return-check.log 2>&1
timeout 60 node stage3/adapt/20-optional-declarations/check-wave1.cjs /tmp/optional-before /tmp/optional-after /tmp/optional-20-before.json /tmp/optional-20-after.json > /tmp/optional-wave1-check.json 2> /tmp/optional-wave1-check.log
PYTHONDONTWRITEBYTECODE=1 timeout 60 python3 stage3/adapt/00-setup/evidence/check-optional-api.py /tmp/optional-before /tmp/optional-after > /tmp/optional-api-check-final.log 2>&1
# Temporarily restore e1c0cb0d's adapt.cjs with a shell EXIT trap to restore the repair.
timeout 240 bash stage3/apply.sh /tmp/optional-mutant > /tmp/optional-apply-mutant.log 2>&1
timeout 360 node stage3/adapt/00-setup/evidence/checker-census.cjs "$PWD" /tmp/optional-mutant /tmp/optional-census-mutant > /tmp/optional-census-mutant.log 2>&1
# Revert witness uses a scratch copy of e1c0cb0d's adapter, expected exit 1.
timeout 90 node stage3/adapt/20-optional-declarations/check-nullable-return.cjs /tmp/optional-revert-adapt.cjs > /tmp/optional-return-revert-mutant.log 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout 60 python3 stage3/adapt/00-setup/evidence/verify-optional-repair.py > /tmp/optional-census-delta.json 2> /tmp/optional-census-delta.log
```

The apply suite passes **23 tests in 13.990s**, and the final file-set rerun
passes all **23 in 13.213s** (`optional-apply-tests-complete.log`). An initial invocation
without `PYTHONPATH=stage3` failed importing `caches`; only its generated bytecode
was removed, and the corrected complete suite passed. Two early census attempts
were discarded: the first overlapped apply, and the second overlapped dependency
installation, causing a CLI/API difference at source-map-support. Only the
completed, frozen dependency-view runs above are retained as observations.
A first after-build attempt used a dependency directory whose name was not
`node_modules`; Node could not resolve picocolors. Moving it to
`/tmp/optional-deps/node_modules` fixed resolution; the complete build passed.
These failed logs are retained separately and are not claimed as proofs.

Setup succeeded on the first bounded invocation; `nproc` is **5**, with
`cpu.max=400000 100000`. Its timing lines are:

```text
setup: node ready (0.063s)
setup: go ready (0.084s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.540s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=0.853s
setup: markdown dependencies ready (1.034s)
setup: submodules ready (246.157s)
setup: shared cache ready (254.358s)
setup: go build ready (491.898s)
setup: test binaries deferred (use --warm-tests) (492.019s)
setup: build cache warm (492.020s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (492.055s)
setup: source /workspace/adamic-tools/env.sh
```

Node is v24.19.0, Go 1.27.1, clang 20.1.8. The public method groups remain
for the language lead, with counts and examples above. The full native gate,
full upstream oracle and fixes for the other groups were not run or claimed.
No PR is opened; this branch must land with its census evidence base.


## Optional-contract wave 2

The wave-1 must is fixed and pushed as `8b7b0554`: one comment line only in
`nullable-return.a`. The real Adamic compiler reports TS2375 in place, the header
matches it, and `check-nullable-return.cjs` passes without stripping the comment.
Wave 2 is cut from that exact pushed commit and must land with wave 1.

The [complete disposition ledger](optional-contracts-wave2/groups.json) covers
all 30 remaining internal property rows. **24 are repaired and six are declined.**
The main census is **233 -> 203**, with **30 exact IDs removed and zero new IDs**:
24 scoped rows plus six original optional-assignment cascades. All remaining
full messages and related-information chains are unchanged. Both public-property
rows and all 29 method rows are preserved, including the control-only method row.

| Original group | Rows | Repaired IDs | Declined IDs |
|---|---:|---|---|
| nullable-argument | 5 | `D9cecacb705c3`, `Dedba949185e9`, `Db5cc84e69483`, `Dbf236c4b28d8`, `Dd8176fe26c6e` | None |
| nullable-assignment | 5 | `Db18a6d195733`, `D4c7cd2fa4786`, `D398ee22e023b`, `Dfef6d22c3ab0` | `D5af6153ea92f` |
| generic-owner | 4 | `Dba090b507f07`, `Dd08d8be5f37d`, `D5fa96141b629`, `Dba1da66fe167` | None |
| unsupported-code | 9 | `Dcc8edd0800fa`, `D06736c072d16`, `Dbb70df85cbc3`, `D899ab2dda4eb`, `Dc00d790b49ab`, `D85a162474867`, `Daf767ecf9b24`, `D255c75fe6d0a` | `Dcb3573ed8c8f` |
| late-exposure | 2 | `Df257019e13f7`, `D551a5cf2dd60` | None |
| indexed-read | 1 | None | `D84985126581f` |
| generic-receiver | 1 | None | `Da43ecabe1c93` |
| required-view | 1 | None | `Dba945f1b0d91` |
| nested-expression | 1 | None | `Da5c0ee2a5f85` |
| nested-generic-result | 1 | `D1c9885b9e665` | None |

The nullable contexts are resolver lookup gaps. TS2322/TS2345 structural data
chains can supply the same compatible stored-undefined evidence as the original
seed codes. The TS2420 and one TS2345 disappear through shared owners. The
generic-owner rows are arrays/callables, whose declarations stay objects for
every type argument; their instantiated slot types prove the non-undefined
source values fit. Bare T and other unresolved generic value contracts remain
excluded. The concrete nested generic result is resolved through its required
data field, without widening a generic parameter.

The builder constructions already store undefined before adaptation 20; earlier
member errors mask the optional-property errors. A mechanical contextual-object
scan supplies those owner facts before the masking errors are removed later.
This refines the wave-1 late-exposure explanation: the writes were present all
along, but 20 did not receive diagnostics for those particular slots.

Three scoped rows are declined because a partial repair would expose a public
property error, violating both this wave's public boundary and no-new-chain
requirement. The resolver records the blockers explicitly:

- CompilerOptions.build is blocked by public CompilerOptions.project.
- ResolvedModuleFull.originalPath is blocked by public packageId and inherited
  ResolvedModule.isExternalLibraryImport.
- ResolvedTypeReferenceDirective.originalPath is blocked by public packageId.

The indexed read has no density/presence proof. The generic receiver can be a
specialization narrowing files. The required watcher view contains a required
owner as well as an optional owner. These remain their original exact IDs.

The six incidental cascade IDs removed are:
`D12935f679493`, `D5977dd79a645`, `D664bf39bc400`, `D31a062fa64e8`, `Ddb064de445c4`, `D9cce204205f3`.

The corrected adapter selects **401 owners**, up from 376. All **25** new owners
are internal and present in its `resolved_findings` ledger. The completed apply
tree changes eight files. Every changed file has byte-identical stock
`transpileModule` output; an actual appended runtime expression is rejected for
each file by the same comparator. Both Node witnesses retain their own-key
fingerprints. Actual `typescript.d.ts` bytes are identical, and the real lane
API check accepts all **222** sanctions (194 added and 193 removed diff lines).
A public method-to-property input mutant is rejected by that check.

The full primary driver and both option controls run with frozen dependencies:

| Run | Wave 1 | Wave 2 |
|---|---:|---:|
| Stock, no prelude | 233 | 203 |
| Stock, filtered prelude | 234 | 204 |
| Adamic library view, no prelude | 233 | 203 |
| Adamic library view, filtered prelude | 234 | 204 |
| Stock, no unchecked indexed access | 134 | 104 |
| Original-prelude iterator-next mutant | 318 | 288 |
| Stock, exact optional properties off | 166 | 166 |
| Stock, known catch variables | 229 | 199 |

The baseline six observations are the retained complete wave-1 tree, with two
fresh option controls. Its source hashes remain unchanged. Each final run
removes the same 30 IDs except the no-exact control, which has no delta.
Compressed raw observations retain the complete source/library hashes and CLI
logs; `verify-optional-wave2.py` checks identities, log hashes, unchanged options,
related information, public/method preservation and every group disposition.

A provisional full apply/census removed the internal slots but exposed three
new public-property chains. It is retained as discarded exploration, not a
passing result. The final blocking rule and negative witness preserve those
old diagnostic IDs. The new fixture initially measured TS2322; after adding
the public-blocker control its final body measures TS2375. Its header was
remeasured and corrected, and the final in-place compiler result agrees.

Setup succeeded in **15.365s**, `nproc=5`, cgroup quota four CPUs:

```text
setup: go ready (0.022s)
setup: node ready (0.023s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.007s
setup: markdown dependencies ready (0.069s)
setup: submodules ready (0.071s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.174s)
setup: shared cache ready (0.920s)
setup: go build ready (15.259s)
setup: test binaries deferred (use --warm-tests) (15.336s)
setup: build cache warm (15.338s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (15.365s)
setup: source /workspace/adamic-tools/env.sh
```

Commands, final apply-test results and per-group mutant observations follow in
the retained proof report. Full native and full upstream-oracle gates are not
claimed for this resolver-only wave.


Wave-2 final validation and mutants:

The final apply suite passes **23 tests in 15.657s**. The complete evidence
verifier passes all eight ordinary/control observations and all eight per-group
or safety reverts. The resolver is restored to SHA-256
`bda0907a98ea721eeaabfb7e6882b152aa80e957d38a295671c25ef8ea76bf97`.
The retained [proof report](optional-contracts-wave2/proof.json) includes every
returned ID, exact reverted resolver file and full CLI/API-backed observations.

| Revert | Main census | Observed result |
|---|---:|---|
| Nullable argument path | 212 | Caught; four scoped argument IDs, four dependent generic IDs and the linked TS2420 return. |
| Nullable assignment path | 209 | Caught; two scoped IDs and four cascades return; shared nonRecursive owner remains fixed through the argument path. |
| Generic object-owner comparison | 207 | Caught; all four generic-owner IDs return. |
| TS2322/TS2345 path | 209 | Caught; six scoped IDs return; shared-owner discharges remain. |
| Nested initializer path | 203 | Unchanged for the declined scoped row. Both witnesses also remain covered through the shared construction scan; no kill or scoped repair is claimed. |
| Nested generic result path | 204 | Caught; D1c9885b9e665 returns. |
| Public-blocker exclusion | 203 | Caught; same count but three new public-property chains, rejected by the final verifier's exact no-new-chain comparison. |
| Masked contextual construction scan | 205 | Caught; both scoped builder IDs return. |

The initial mutant harness expected the declined nested-initializer witness to
fail. It survived through the shared scan. That exploratory survival is retained,
and the final record explicitly requires its six-run census to remain unchanged.
All repaired-group reverts fail a focused witness and restore scoped IDs in every
one of the six census profiles. The public-blocker safety revert fails the blocked
peer witness and the no-new-chain check. An omitted owner attribution also fails
the resolved-findings check; actual runtime-expression mutations are caught in
every changed file, and the actual API method conversion is rejected.

Full apply checkouts were monitored and completed scratch trees were removed
when disk usage rose. Every report, CLI stream, source/library hash manifest and
actual reverted resolver input remains retained. No incomplete run is relabeled
as a passing observation.

Representative exact commands (all test output goes to logs):

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/optional-source
timeout 90 go build -o /tmp/optional-adamic ./cmd/adamic > /tmp/optional-header-build.log 2>&1
timeout 30 /tmp/optional-adamic c stage3/adapt/20-optional-declarations/testdata/nullable-return.a > /tmp/optional-header-after.stdout 2> /tmp/optional-header-after.log
timeout 90 node stage3/adapt/20-optional-declarations/check-nullable-return.cjs > /tmp/optional-header-witness.log 2>&1
# The header-only commit is pushed before cutting wave 2 from origin's exact ref.
timeout 300 bash stage3/apply.sh /tmp/optional-wave2-final > /tmp/optional-wave2-final-apply.log 2>&1
# Identical locked dependencies are linked for build, then unlinked before census.
timeout 300 npm run build --prefix /tmp/optional-wave2-final > /tmp/optional-wave2-final-build.log 2>&1
timeout 360 node stage3/adapt/00-setup/evidence/checker-census.cjs "$PWD" /tmp/optional-wave2-final /tmp/optional-wave2-final-census > /tmp/optional-wave2-final-census.log 2>&1
timeout 180 node stage3/adapt/00-setup/evidence/checker-census.cjs "$PWD" /tmp/optional-wave2-final /tmp/optional-wave2-final-census --option-controls > /tmp/optional-wave2-final-controls.log 2>&1
timeout 90 node stage3/adapt/20-optional-declarations/check-wave2.cjs /tmp/optional-after /tmp/optional-wave2-final /tmp/optional-20-after.json /tmp/optional-wave2-final-report.json > /tmp/optional-wave2-final-bytes.json 2> /tmp/optional-wave2-final-bytes.log
PYTHONDONTWRITEBYTECODE=1 timeout 60 python3 stage3/adapt/00-setup/evidence/check-optional-api.py /tmp/optional-after /tmp/optional-wave2-final > /tmp/optional-wave2-final-api.log 2>&1
PYTHONPATH="$PWD/stage3" PYTHONDONTWRITEBYTECODE=1 timeout 90 python3 -m unittest stage3/test_apply.py > /tmp/optional-wave2-final-apply-tests.log 2>&1
timeout 120 node stage3/adapt/20-optional-declarations/check-internal.cjs > /tmp/optional-wave2-restored-internal.log 2>&1
timeout 90 node stage3/adapt/20-optional-declarations/check-nullable-return.cjs > /tmp/optional-wave2-restored-return.log 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout 3600 python3 stage3/adapt/20-optional-declarations/run-wave2-mutants.py /tmp/optional-wave2-mutants /tmp/optional-wave2-final-census/raw.json > /tmp/optional-wave2-mutants.log 2>&1
# Bounded resume after recording the declined-group survival; completed groups are skipped.
PYTHONDONTWRITEBYTECODE=1 timeout 2400 python3 stage3/adapt/20-optional-declarations/run-wave2-mutants.py /tmp/optional-wave2-mutants /tmp/optional-wave2-final-census/raw.json > /tmp/optional-wave2-mutants-resume.log 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout 60 python3 stage3/adapt/00-setup/evidence/verify-optional-wave2.py > /tmp/optional-wave2-verify.log 2>&1
```

Each mutant apply has a 300-second limit and each unchanged census driver a
360-second limit. The main harness is monitored during those bounded subprocesses.
Final header checks invoke the real compiler in place on both `.a` witnesses:
both exit 1 with first code TS2375, matching their first-line headers. No header
is stripped by the resolver witness scripts.

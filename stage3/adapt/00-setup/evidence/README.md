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
mutants and toolchain results will be added separately.

# Required indexed reads in the program partition

## Current coverage after wave 4

This wave reviews all 58 indexed expressions in moduleNameResolver.ts (17),
moduleSpecifiers.ts (12), and sys.ts (29). It adds one required path-mapping
array read and two bounded dense Buffer byte reads. The cumulative ledger has
222 entries: 68 assertions and 154 declines across eleven reviewed files;
eight files change. No U-zero site is selected. All files are planned before
any write, using stock 6.0.3 AST occurrence addresses and unchanged CRLF.

| File | TS2345 before/after | TS18048 before/after | TS2532 before/after | TS2322 before/after | TS2538 before/after | Five-code before/after | Full checker incremental before/after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| moduleNameResolver.ts | 3 / 3 | 0 / 0 | 0 / 0 | 4 / 4 | 0 / 0 | 7 / 7 | 38 / 38 |
| moduleSpecifiers.ts | 1 / 1 | 0 / 0 | 1 / 0 | 1 / 1 | 0 / 0 | 3 / 2 | 12 / 11 |
| sys.ts | 1 / 1 | 0 / 0 | 0 / 0 | 1 / 1 | 0 / 0 | 2 / 2 | 61 / 61 |

The new files' five-code census is 12 to 11; cumulative census is 132 to 35
and incremental census is 36 to 35. Before and after retain all 78 compiler
roots and unchanged Adamic checker options. Every remaining target-code and
other-code diagnostic is recorded with its full chain and reviewed reason in
`declined-findings.json` and `other-findings.json`. `whole-files.json` gives
original 00+10+30 and cumulative full-checker totals. Its sys.ts original total
of 62 includes a finding already removed by a previous wave; this wave's
incremental sys.ts total remains 61.

**New full-checker zero files: none.** The already-zero reviewed files remain
symbolWalker.ts and performance.ts. The new files retain exact optional
contracts, return-path and enum findings; sys.ts also retains missing Node
host types and unknown catches. typesVersions values are deliberately checked
and diagnosed by the caller, and an invalid preferred ending already fails in
Debug.assertNever. Neither is replaced by an assertion.

Stock emitted JavaScript is byte-identical for all eleven files. The second
actual adapter CLI run makes zero edits. A `paths[key]!` to `?? 0` mutant is
caught by the independent stock JavaScript comparison and the site contract,
each exiting 1; `wave4-mutant-proof.json` records both. Final default oracle: **106,367 passing, zero failing, zero pending**,
with an empty baseline diff. Install 3.686s, build 7.822s, tests 334.644s,
total 346.217s. The tree contains only 00, 10, 30 and 32, excluding adaptation 20.
Prior evidence is archived under wave1/, wave2/ and wave3/.

Actual logs are `/tmp/adapt32-wave4-*.log`; the default oracle's full phase
logs are in `/tmp/adapt32-wave4-oracle/`. Run the cumulative reproduction
commands below; the new mutant command is:

```sh
node "$unit/wave-mutant.cjs" /tmp/adapt32-before /tmp/adapt32-after /tmp/new-wave4-mutant moduleSpecifiers.ts 'paths[key]' > /tmp/new-wave4-mutant.log 2>&1
```

## Historical coverage after wave 3

Wave 3 reviews **all 60 indexed expressions in program.ts and all 67 in
commandLineParser.ts** and adds 35 and 11 required-read assertions respectively.
The cumulative adapter reviews eight files, changes six, and has 164 ledger
entries: **65 assertions and 99 declines**, with no U-zero adaptations.
The existing `options[setting] ?? defaultValue` is recorded as an unchanged
handling shape; the adapter validates its right operand and does not introduce
a default there. All eight files are planned before any is written.

| Wave 3 file | TS2345 before/after | TS18048 before/after | TS2532 before/after | TS2322 before/after | TS2538 before/after | Five-code total before/after | Full checker before/after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| program.ts | 32 / 6 | 13 / 0 | 3 / 0 | 3 / 0 | 1 / 0 | 52 / 6 | 90 / 38 |
| commandLineParser.ts | 10 / 4 | 5 / 0 | 3 / 0 | 4 / 4 | 1 / 0 | 23 / 8 | 40 / 25 |

Cumulative five-code census: **120 to 24**. Wave 3 incremental census across
the eight reviewed files: **85 to 24**. All remaining findings have their full
diagnostic chain and reason in `declined-findings.json` (24 target-code
findings) or `other-findings.json` (78 findings outside those codes).
The before and after census use all 78 roots and unchanged Adamic options.

**New whole files at zero in this wave: none.** Among reviewed partition
files, `symbolWalker.ts` and `performance.ts` remain at zero for the full
checker, both already zero before these waves. Program and command-line parser
retain optional-property, enum, return-path, catch/iterator and validation
findings. Assertions cannot legitimately repair those contracts under the
required-indexed-read scope. `whole-files.json` records every full-checker code
count and zero-file result for the eight reviewed files. This does not claim
the 40-of-78 target is met.

Stock 6.0.3 emitted JavaScript is byte-identical for all eight files. CRLF
counts are unchanged and the second actual CLI invocation makes zero edits.
`wave-mutant.cjs` substitutes `(commonPathComponents[i] ?? 0)` in program.ts
and `(args[i] ?? 0)` in commandLineParser.ts in separate scratch trees. Both
independent stock emit comparisons fail with exit 1, and both site checks fail
with exit 1. `wave3-program-mutant-proof.json` and
`wave3-command-mutant-proof.json` record the results. The mutants are absent
from the final tree.

Wave 3 default oracle: **106,367 passing, zero failing, zero pending**;
install, build and tests each exited 0, with an empty (0-byte) baseline diff.
Install took 4.981s, build 19.321s, tests 449.445s, total 473.805s.
The first attempted run was stopped after the census exposed a second
cache-setter argument error hidden behind its first argument error. Orphaned
workers from that attempt were terminated; its result is not proof. The final
run builds and tests the corrected paired resolution read ledgered above.
The final tree contains 00, 10, 30 and this adapter, without adaptation 20.
Prior-wave results are preserved under `wave1/` and `wave2/` and in commits
`85be312` and `fa8928a`. Historical tables below describe those earlier waves;
current root-level census, proof and finding files describe wave 3.

To reproduce wave 3, prepare the pinned 00+10+30 before tree and an after
copy, then run the adapter, census, verify and oracle commands in the
Reproduction section. Additional mutant commands are:

```sh
node "$unit/wave-mutant.cjs" /tmp/adapt32-before /tmp/adapt32-after /tmp/new-program-mutant program.ts 'commonPathComponents[i]' > /tmp/new-program-mutant.log 2>&1
node "$unit/wave-mutant.cjs" /tmp/adapt32-before /tmp/adapt32-after /tmp/new-command-mutant commandLineParser.ts 'args[i]' > /tmp/new-command-mutant.log 2>&1
```

Actual final-wave logs are `/tmp/adapt32-wave3-*-final.log`, with full oracle
phase logs under `/tmp/adapt32-wave3-oracle-final/`. Final raw census is under
`/tmp/adapt32-wave3-census-final/`. The scope remains this adaptation directory;
`debug.ts` and `path.ts` are excluded and no source checkout is committed.

## Historical waves 1 and 2

This is a **partial partition 32**, on TypeScript 6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. It adds **19 required-read `!`
checks** in `binder.ts`, `semver.ts`, `programDiagnostics.ts`, and `tracing.ts`. It also
reviews `symbolWalker.ts` and `performance.ts` without changing them. The
ledger has one entry for each of the 37 parsed indexed expressions in these
six files: 19 asserted and eighteen declined. No U-zero site was selected.

**debug.ts and path.ts belong to adaptation 30 and are excluded**, following
the user's updated ownership ruling. No checker, emitter, factory, parser,
scanner, core, utilities, transformer, or other partition file is edited by
this adapter. All new repository files are under this directory. TypeScript
source remains in scratch checkouts and is never committed.

## Method

The adapter follows adaptation 30: stock TypeScript 6.0.3 parses current text,
addresses expressions by occurrence rather than line offsets, validates total
occurrence counts, and plans every selected file before writing any. Each
required read gets `!` at its original evaluation point. CRLF bytes, repeated
reads, callback boundaries, and existing optional handling are preserved. There
are no new optional chains, skips, hoists, defaults, or uninitialized assertions.
Classes and one-line invariants are recorded in `sites.json` and below.

`!` is the user's loud required-value check natively and erased syntax on Node.
The ledger states obligations of populated compiler structures, not a proof
that arbitrary arrays are dense. Invalid states can fail loudly in Adamic.
Stock emitted JavaScript is byte-identical for every selected file.

## Prerequisites and toolchain

Base main: `e011f8f60899586d6373a5ccb07335ad82cfbf3c`.
Adaptation 10: `a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e`.
Adaptation 30: `07f637cccd600dbafb09cae3d5a57541fffe9f5e`.
Both prerequisite branches were merged before this unit was written.
Adaptation 20 is excluded. The scratch before tree contains 00, 10, and 30;
the after tree adds this adapter only.

`bash cloud/setup.sh` passed: Go ready 1s, clang ready 1s, Node ready 1s,
submodules ready 1s, build cache warm 298s, total 298s. `nproc` is 5 and
`cpu.max` is `400000 100000`. Commands source
`/workspace/adamic-tools/env.sh`. Node 24.19.0, Go 1.27.1, clang 20.1.8.

## Census for waves 1 and 2

The unchanged Adamic loader options check all 78 prepared compiler roots.
The exporter uses the same Go overlay as adaptation 30 and never invokes the
meter's implicit optional-declaration adaptation. These tables retain exactly
the reviewed six files and requested codes; full diagnostic chains are saved
in `census-before.json` and `census-after.json`.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| semver.ts before | 12 | 3 | 0 | 0 | 0 | 15 |
| semver.ts after | 9 | 1 | 0 | 0 | 0 | 10 |
| programDiagnostics.ts before | 4 | 0 | 0 | 2 | 0 | 6 |
| programDiagnostics.ts after | 0 | 0 | 0 | 0 | 0 | 0 |
| tracing.ts before | 2 | 15 | 2 | 1 | 0 | 20 |
| tracing.ts after | 0 | 0 | 0 | 0 | 0 | 0 |
| symbolWalker.ts before | 0 | 0 | 0 | 0 | 0 | 0 |
| symbolWalker.ts after | 0 | 0 | 0 | 0 | 0 | 0 |
| performance.ts before | 0 | 0 | 0 | 0 | 0 | 0 |
| performance.ts after | 0 | 0 | 0 | 0 | 0 | 0 |
| binder.ts before | 1 | 1 | 1 | 1 | 0 | 4 |
| binder.ts after | 0 | 0 | 0 | 0 | 0 | 0 |
| **Total before** | 19 | 19 | 3 | 4 | 0 | **45** |
| **Total after** | 9 | 1 | 0 | 0 | 0 | **10** |

Every remaining requested-code finding is listed here and in
`declined-findings.json`:

| Site | Code | Reason |
| --- | --- | --- |
| semver.ts:149:25 | TS2345 | major is a destructured regexp capture, not an element-access expression; its mandatory capture invariant needs a separate adaptation. |
| semver.ts:287:48 | TS18048 | simple is a split/for-of binding, not an indexed read, under the checker regex/string model. |
| semver.ts:288:48 | TS2345 | rangeRegExp capture 1 is optional and parseComparator handles undefined; its string parameter declaration does not represent that contract. |
| semver.ts:302:20 | TS2345 | major is derived from a destructured regexp capture, not an indexed expression at this use. |
| semver.ts:302:42 | TS2345 | major is derived from a destructured regexp capture, not an indexed expression at this use. |
| semver.ts:303:20 | TS2345 | major is derived from a destructured regexp capture, not an indexed expression at this use. |
| semver.ts:304:20 | TS2345 | major is derived from a destructured regexp capture, not an indexed expression at this use. |
| semver.ts:319:21 | TS2345 | leftResult.major inherits the destructured-capture type; this is a property read, not an indexed read. |
| semver.ts:323:21 | TS2345 | rightResult.major inherits the destructured-capture type; this is a property read, not an indexed read. |
| semver.ts:339:21 | TS2345 | rightResult.major inherits the destructured-capture type; this is a property read, not an indexed read. |

Five-code findings fall from **45 to 10**. This does not mean all six files
pass the complete checker. `symbolWalker.ts` and `performance.ts` have no
checker findings both before and after. `programDiagnostics.ts` still has
TS2412, and `tracing.ts` still has ten findings across TS1294, TS18046, TS2379,
TS2412, and TS2591. These are outside this required-indexed-read wave.
`other-findings.json` lists every remaining diagnostic outside the five codes
in the six reviewed files, with its reason for exclusion.

## Proof for waves 1 and 2

`verify.cjs` independently emits each file with stock 6.0.3 and requires
byte-identical JavaScript. It also checks newline and CRLF counts, every site
contract, and idempotence. All six files pass. A second actual adapter CLI
run reports zero assertions and zero U-zero edits for each file.

The mutant changes `left[i]!` in semver's prerelease comparison to
`(left[i] ?? 0)`. The independent emitted-JavaScript comparison fails with
`unexpected emitted JavaScript change in semver.ts`, exit 1. The adapter's
site check independently fails with `required read defaulted: semver.ts:171:32`,
exit 1. `mutant.cjs` reproduces both failures in a new scratch tree and
`mutant-proof.json` records them. The mutant is absent from the final tree.
No oracle or census mutant result is claimed.

The support wave default oracle passed: 106367 passing, 0 failing, 0 pending,
empty baseline diff. Its full evidence is under `wave1/`. The binder wave
also passed the default oracle: **106367 passing, 0 failing, 0 pending**,
all phase exits 0, and an empty (0-byte) `baseline.diff`. Install took 4.725s,
build 10.853s, tests 521.665s, total 537.303s. The tree contains 00, 10, 30,
and both waves of 32; adaptation 20 is absent. `oracle-report.json` records
the exact commands and result.
An earlier scratch oracle containing path.ts checks was stopped after the
ownership correction; it exited 1 and is not used as passing evidence.

## Binder wave

Wave 2 adds four binder assertions after the support wave was pushed. Its
incremental census is 14 to 10 across the six reviewed files; binder alone is
4 to 0. Cumulative baseline counts are 45 to 10. All ten remaining requested-code
findings are the semver findings listed above. Binder also retains eighteen
findings outside these codes: TS1294, TS2412, TS7029, and TS7030.

`binder-mutant.cjs` replaces `antecedents[0]!` with `(antecedents[0] ?? 0)`.
Both the stock emitted-JavaScript comparison and the site check fail with exit 1.
The original semver mutant is retained. Each wave has its own default oracle
run; the support wave evidence is preserved under `wave1/`.

## Reproduction

Prepare a pinned scratch tree with adaptations 00, 10, and 30 only, then copy
it to a separate after tree. From the repository:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
unit=stage3/adapt/32-indexed-reads-program
bash "$unit/census.sh" /tmp/adapt32-before /tmp/new-cumulative-before > /tmp/new-cumulative-before.log 2>&1
node "$unit/adapt.cjs" /tmp/adapt32-after > /tmp/new-cumulative-adapt.log 2>&1
bash "$unit/census.sh" /tmp/adapt32-after /tmp/new-cumulative-after > /tmp/new-cumulative-after.log 2>&1
node "$unit/verify.cjs" /tmp/adapt32-before /tmp/adapt32-after > /tmp/new-cumulative-verify.log 2>&1
node "$unit/adapt.cjs" /tmp/adapt32-after > /tmp/new-cumulative-idempotence.log 2>&1
node "$unit/mutant.cjs" /tmp/adapt32-before /tmp/adapt32-after /tmp/new-cumulative-mutant > /tmp/new-cumulative-mutant.log 2>&1
node "$unit/binder-mutant.cjs" /tmp/adapt32-before /tmp/adapt32-after /tmp/new-binder-mutant > /tmp/new-binder-mutant.log 2>&1
stage3/oracle/run.sh /tmp/adapt32-after /tmp/new-cumulative-oracle > /tmp/new-cumulative-oracle.log 2>&1
```

`CENSUS_TYPESCRIPT` may select the absolute stock typescript.js path instead
of `NODE_PATH`. Actual logs use `/tmp/adapt32-wave1-*` and `/tmp/adapt32-wave2-*`; complete
oracle phase logs are under `/tmp/adapt32-wave1-oracle/` and
`/tmp/adapt32-wave2-oracle/`. Support commit: `85be312`.

## Remaining partition work

Module resolution and specifiers, sys,
builders, watchers, resolutionCache, tsbuild files, expressionToTypeNode,
executeCommandLine, and other owned compiler files are not covered by the first three waves.
Their source-local invariants have not yet been reviewed and no required-read
assertions are inferred for them. These waves do not establish the critical
path target of 40 complete compiler files past the checker. The whole Adamic
repository gate, native tsc build, browser integration, and ESLint-rule tests
are not run for this source-adaptation-only wave.

## Site ledger

| Site | Read | Class | Action | Invariant or decline reason |
| --- | --- | --- | --- | --- |
| programDiagnostics.ts:192:30 | `file.libReferenceDirectives[reason.index]` | U-position | assert | The lib-reference reason index comes from this source file directive list. |
| programDiagnostics.ts:253:94 | `fileIncludeReasons[0]` | U-endpoint | assert | Processing an extra reason pushes its diagnostic before this cache merge reads the first element. |
| programDiagnostics.ts:256:123 | `fileIncludeReasons[0]` | U-endpoint | assert | Processing an extra reason pushes its diagnostic before this cache merge reads the first element. |
| programDiagnostics.ts:365:60 | `rootNames[reason.index]` | U-position | assert | The root-file inclusion reason records an index into this program root-name list. |
| programDiagnostics.ts:382:66 | `resolvedProjectReferences?.[reason.index]` | U-position | decline | Debug.checkDefined already checks the optional indexed result. |
| programDiagnostics.ts:397:25 | `referencesSyntax.elements[index]` | U-position | assert | The syntax array length guard covers the project-reference index recorded during reference traversal. |
| programDiagnostics.ts:410:122 | `options.lib[reason.index]` | U-position | assert | A LibFile reason with an index refers to the populated explicit compiler-options lib list. |
| semver.ts:171:32 | `left[i]` | U-parallel | assert | The loop uses the minimum length of populated prerelease identifier arrays. |
| semver.ts:172:33 | `right[i]` | U-parallel | assert | The loop uses the minimum length of populated prerelease identifier arrays. |
| semver.ts:283:30 | `match[1]` | U-regex | assert | A successful hyphenRegExp match requires the nonempty left partial-version capture. |
| semver.ts:283:40 | `match[2]` | U-regex | assert | A successful hyphenRegExp match requires the nonempty right partial-version capture. |
| semver.ts:288:48 | `match[1]` | U-regex | decline | rangeRegExp capture 1 is optional; parseComparator handles undefined in its switch. |
| semver.ts:288:58 | `match[2]` | U-regex | assert | A successful rangeRegExp match requires the nonempty partial-version capture 2. |
| symbolWalker.ts:76:17 | `visitedTypes[type.id]` | U-table | decline | The visited-type table is deliberately sparse; absence means not yet visited. |
| symbolWalker.ts:79:13 | `visitedTypes[type.id]` | U-table | decline | This is a pure store marking the type visited, not an indexed read. |
| symbolWalker.ts:190:17 | `visitedSymbols[symbolId]` | U-table | decline | The visited-symbol table is deliberately sparse; absence means not yet visited. |
| symbolWalker.ts:193:13 | `visitedSymbols[symbolId]` | U-table | decline | This is a pure store marking the symbol visited, not an indexed read. |
| tracing.ts:122:13 | `legend[legend.length - 1]` | U-endpoint | assert | startTracing pushes the current trace record before stopTracing updates it. |
| tracing.ts:175:66 | `eventStack[index]` | U-position | assert | pop and popAll pass an in-range index into stack entries populated by push. |
| tracing.ts:220:27 | `legend[legend.length - 1]` | U-endpoint | assert | startTracing pushes the current trace record before dumpTypes reads its path. |
| tracing.ts:230:26 | `types[i]` | U-loop | assert | dumpTypes traverses the populated catalog built by recordType at its bounded index. |
| tracing.ts:332:47 | `symbol?.declarations?.[0]` | U-endpoint | decline | getLocation accepts undefined for a missing symbol or first declaration. |

| binder.ts:1428:20 | `antecedents[0]` | U-endpoint | assert | addAntecedent builds a populated flow list and length === 1 selects its sole flow node. |
| binder.ts:1746:21 | `clauses[i]` | U-loop | assert | The outer loop and guarded increment keep i inside the parser-built populated switch-clause list. |
| binder.ts:1750:22 | `clauses[i]` | U-loop | decline | bind explicitly accepts undefined and handles absence nearby; preserve this optional argument read. |
| binder.ts:1757:28 | `clauses[i]` | U-loop | assert | The outer loop and guarded increment keep i inside the parser-built populated switch-clause list. |
| binder.ts:1933:17 | `state.inStrictModeStack[state.stackIndex]` | U-position | decline | This is a pure store saving the mode, not an indexed read. |
| binder.ts:1934:17 | `state.parentStack[state.stackIndex]` | U-position | decline | This is a pure store saving the parent, not an indexed read. |
| binder.ts:2005:39 | `state.inStrictModeStack[state.stackIndex]` | U-position | decline | Stack slot zero is deliberately undefined; the following test handles missing saved modes. |
| binder.ts:2006:33 | `state.parentStack[state.stackIndex]` | U-position | decline | Stack slot zero is deliberately undefined; the following test handles missing saved parents. |
| binder.ts:3205:54 | `node.arguments[0]` | U-endpoint | decline | BindableObjectDefinePropertyCall already types arguments as a populated three-element tuple. |
| binder.ts:3372:64 | `node.arguments[0]` | U-endpoint | decline | The existing PropertyAccessExpression cast covers this tuple read; no required indexed-absence finding remains. |
| binder.ts:3399:61 | `node.arguments[0]` | U-endpoint | decline | BindableObjectDefinePropertyCall already types arguments as a populated three-element tuple. |
| binder.ts:3401:77 | `node.arguments[0]` | U-endpoint | decline | BindableObjectDefinePropertyCall already types arguments as a populated three-element tuple. |
| binder.ts:3490:22 | `declaration.arguments[2]` | U-endpoint | decline | The isBindableObjectDefinePropertyCall guard supplies a three-element tuple with a required descriptor at index 2. |
| binder.ts:3501:22 | `declaration.arguments[2]` | U-endpoint | decline | The isBindableObjectDefinePropertyCall guard supplies a three-element tuple with a required descriptor at index 2. |
| binder.ts:3636:63 | `symbolExport.declarations[0]` | U-endpoint | assert | A conflicting bound prototype export symbol has a populated declaration list whose first declaration receives the diagnostic. |

## Wave 3 site ledger

| Site | Read | Class | Action | Invariant or decline reason |
| --- | --- | --- | --- | --- |
| program.ts:357:38 | `commonPathComponents[i]` | U-parallel | assert | The minimum-length loop traverses both populated normalized path-component arrays. |
| program.ts:357:88 | `sourcePathComponents[i]` | U-parallel | assert | The minimum-length loop traverses both populated normalized path-component arrays. |
| program.ts:996:18 | `node.elements[0]` | U-endpoint | assert | The import-attribute length-one guard selects the sole parser-populated attribute. |
| program.ts:1152:16 | `components[1]` | U-endpoint | decline | The component is used only in string concatenation, which already accepts absence; no required-value indexed finding exists. |
| program.ts:1154:12 | `components[i]` | U-loop | decline | The while condition tests component presence before comparison and concatenation; absence terminates traversal. |
| program.ts:1154:29 | `components[i]` | U-loop | decline | The while condition tests component presence before comparison and concatenation; absence terminates traversal. |
| program.ts:1155:41 | `components[i]` | U-loop | decline | The while condition tests component presence before comparison and concatenation; absence terminates traversal. |
| program.ts:1208:29 | `file.referencedFiles[index]` | U-position | assert | The ReferencedFile reason records an index into the corresponding populated source-file reference list. |
| program.ts:1211:29 | `file.typeReferenceDirectives[index]` | U-position | assert | The ReferencedFile reason records an index into the corresponding populated source-file reference list. |
| program.ts:1212:93 | `file.typeReferenceDirectives[index]` | U-position | assert | The ReferencedFile reason records an index into the corresponding populated source-file reference list. |
| program.ts:1215:29 | `file.libReferenceDirectives[index]` | U-position | assert | The ReferencedFile reason records an index into the corresponding populated source-file reference list. |
| program.ts:1282:46 | `program.getResolvedProjectReferences()[index]` | U-parallel | decline | resolvedProjectReferenceUptoDate accepts undefined for unresolved references and handles it nearby. |
| program.ts:1311:25 | `oldResolvedRef.commandLine.projectReferences[index]` | U-parallel | assert | Resolved child references and commandLine.projectReferences share their recorded index; the original reference list is populated. |
| program.ts:1525:24 | `options[option.name]` | U-table | decline | typeof deliberately inspects an optional compiler option value. |
| program.ts:1796:55 | `automaticTypeDirectiveNames[i]` | U-loop | assert | The bounded loop indexes the populated automatic type-directive name list. |
| program.ts:1796:107 | `resolutions[i]` | U-parallel | assert | The resolution worker returns one record per automatic type-directive name, including failed resolutions; the cache setter requires that record. |
| program.ts:1798:21 | `automaticTypeDirectiveNames[i]` | U-loop | assert | The bounded loop indexes the populated automatic type-directive name list. |
| program.ts:1800:21 | `resolutions[i]` | U-parallel | assert | The resolution worker returns one resolution record per automatic type-directive name, including unsuccessful resolutions. |
| program.ts:1803:40 | `automaticTypeDirectiveNames[i]` | U-loop | assert | The bounded loop indexes the populated automatic type-directive name list. |
| program.ts:1804:36 | `resolutions[i]` | U-parallel | decline | The existing optional chain explicitly handles a missing resolution record or resolved directive. |
| program.ts:1861:38 | `parent?.commandLine.projectReferences[index]` | U-parallel | decline | The optional parent reference is tested by the existing root-reference fallback. |
| program.ts:1861:87 | `oldProgram.getProjectReferences()[index]` | U-parallel | assert | When the optional parent reference is absent, the callback index addresses the populated old root-reference list. |
| program.ts:2261:27 | `entries[i]` | U-loop | assert | The bounded loop traverses the populated resolution-entry input list. |
| program.ts:2284:21 | `(result ??= new Array(entries.length))[i]` | U-position | decline | The indexed target is a pure store assembling the resolution list, not an indexed read. |
| program.ts:2302:21 | `(result ??= new Array(entries.length))[i]` | U-position | decline | The indexed target is a pure store assembling the resolution list, not an indexed read. |
| program.ts:2315:52 | `result[unknownEntryIndices[index]]` | U-position | decline | The indexed target is a pure store assembling the resolution list, not an indexed read. |
| program.ts:2315:59 | `unknownEntryIndices[index]` | U-parallel | assert | Unknown entries and their original indices are pushed together and the resolution worker returns one record per unknown entry. |
| program.ts:2324:32 | `(parent ? parent.commandLine.projectReferences : projectReferences)[index]` | U-parallel | assert | forEachProjectReference supplies the index shared by the populated original project-reference list and its resolution list. |
| program.ts:2981:46 | `lineStarts[line]` | U-position | decline | String.slice accepts an absent offset; its existing end-of-file handling remains unchanged. |
| program.ts:2981:64 | `lineStarts[line + 1]` | U-position | decline | String.slice accepts an absent offset; its existing end-of-file handling remains unchanged. |
| program.ts:3116:70 | `parent.modifiers[decoratorIndex]` | U-position | assert | A successful guarded findIndex supplies the index of an existing parser-populated modifier. |
| program.ts:3124:78 | `parent.modifiers[decoratorIndex]` | U-position | assert | A successful guarded findIndex supplies the index of an existing parser-populated modifier. |
| program.ts:3130:69 | `parent.modifiers[trailingDecoratorIndex]` | U-position | assert | A successful guarded findIndex supplies the index of an existing parser-populated modifier. |
| program.ts:3131:69 | `parent.modifiers[decoratorIndex]` | U-position | assert | A successful guarded findIndex supplies the index of an existing parser-populated modifier. |
| program.ts:3479:58 | `supportedExtensions[0]` | U-endpoint | decline | forEach accepts an absent extension group and handles it without dereferencing the result. |
| program.ts:3526:30 | `(FileIncludeKind as any)[reason.kind]` | U-table | decline | The existing any cast is a tracing enum-name lookup, not a required indexed-absence finding. |
| program.ts:3780:25 | `file.typeReferenceDirectives[index]` | U-loop | assert | The bounded type-directive loop traverses the populated source-file directive list. |
| program.ts:3781:52 | `resolutions[index]` | U-parallel | assert | The type-resolution worker returns one resolution record per type directive, even for failed resolutions. |
| program.ts:3923:36 | `resolutions[index]` | U-parallel | assert | The resolution worker returns one record per module name and the adjacent assertion checks equal list lengths. |
| program.ts:3924:36 | `moduleNames[index]` | U-loop | assert | The bounded module-resolution loop traverses the populated module-name list. |
| program.ts:3925:66 | `moduleNames[index]` | U-loop | assert | The bounded module-resolution loop traverses the populated module-name list. |
| program.ts:3926:57 | `resolutions[index]` | U-parallel | assert | The resolution worker returns one record per module name and the adjacent assertion checks equal list lengths. |
| program.ts:3927:81 | `resolutions[index]` | U-parallel | assert | The resolution worker returns one record per module name and the adjacent assertion checks equal list lengths. |
| program.ts:3957:36 | `file.imports[index]` | U-parallel | decline | isInJSFile explicitly accepts undefined; the adjacent required flags read carries the populated-import obligation. |
| program.ts:3957:62 | `file.imports[index]` | U-parallel | assert | The earlier index < file.imports.length condition guards this read in the populated imports portion of moduleNames. |
| program.ts:4148:29 | `options.paths[key]` | U-table | decline | isArray validates the optional or malformed path-option value and diagnoses invalid values in the else branch. |
| program.ts:4149:33 | `options.paths[key]` | U-table | assert | The preceding isArray check establishes that the paths entry exists and is an array. |
| program.ts:4154:39 | `options.paths[key][i]` | U-loop | decline | The substitution value is deliberately validated with typeof and diagnosed when absent or malformed. |
| program.ts:4154:39 | `options.paths[key]` | U-table | assert | The preceding isArray check establishes that the paths entry exists and is an array. |
| program.ts:4374:13 | `ModuleKind[moduleKind]` | U-table | decline | The existing enum lookup condition or fallback handles absence; the optional diagnostic-name value is accepted. |
| program.ts:4378:36 | `ModuleKind[moduleKind]` | U-table | decline | The existing enum lookup condition or fallback handles absence; the optional diagnostic-name value is accepted. |
| program.ts:4379:42 | `ModuleResolutionKind[moduleKindName as any]` | U-table | decline | The existing enum lookup condition or fallback handles absence; the optional diagnostic-name value is accepted. |
| program.ts:4383:13 | `ModuleResolutionKind[moduleResolution]` | U-table | decline | The existing enum lookup condition or fallback handles absence; the optional diagnostic-name value is accepted. |
| program.ts:4387:42 | `ModuleResolutionKind[moduleResolution]` | U-table | decline | The existing enum lookup condition or fallback handles absence; the optional diagnostic-name value is accepted. |
| program.ts:4557:54 | `ModuleKind[options.module]` | U-table | decline | createDeprecatedDiagnostic explicitly accepts an absent diagnostic value. |
| program.ts:4593:29 | `(parent ? parent.commandLine.projectReferences : projectReferences)[index]` | U-parallel | assert | forEachProjectReference supplies the index shared by the populated original project-reference list and its resolution list. |
| program.ts:4624:121 | `initializer.elements[valueIndex]` | U-position | assert | The syntax-array length guard covers the recorded nonnegative option/reference index in a populated parser array. |
| program.ts:4673:123 | `referencesSyntax.elements[index]` | U-position | assert | The syntax-array length guard covers the recorded nonnegative option/reference index in a populated parser array. |
| program.ts:5099:53 | `option[d.skippedOn]` | U-table | decline | The optional compiler flag is deliberately tested for truthiness. |
| program.ts:5191:40 | `imports[index]` | U-position | assert | The resolution-index contract and imports.length guard select a populated import string literal. |
| commandLineParser.ts:274:55 | `entry[0]` | U-endpoint | decline | libEntries is already typed as fixed two-element tuples, so entry[0] is a required string in the checker. |
| commandLineParser.ts:1979:23 | `args[i]` | U-loop | assert | The while-loop bound covers the populated command-line or response-file argument list before i is incremented. |
| commandLineParser.ts:2048:26 | `args[i]` | U-position | decline | The next option argument may be absent; nearby diagnostics, comparisons, defaults or converters deliberately handle that absence. |
| commandLineParser.ts:2050:13 | `options[opt.name]` | U-table | decline | This indexed option target is a pure store, not an indexed read. |
| commandLineParser.ts:2055:17 | `options[opt.name]` | U-table | decline | This indexed option target is a pure store, not an indexed read. |
| commandLineParser.ts:2070:14 | `args[i]` | U-position | decline | The next option argument may be absent; nearby diagnostics, comparisons, defaults or converters deliberately handle that absence. |
| commandLineParser.ts:2074:13 | `args[i]` | U-position | decline | The next option argument may be absent; nearby diagnostics, comparisons, defaults or converters deliberately handle that absence. |
| commandLineParser.ts:2077:21 | `options[opt.name]` | U-table | decline | This indexed option target is a pure store, not an indexed read. |
| commandLineParser.ts:2077:79 | `args[i]` | U-position | decline | The next option argument may be absent; nearby diagnostics, comparisons, defaults or converters deliberately handle that absence. |
| commandLineParser.ts:2082:38 | `args[i]` | U-position | decline | The next option argument may be absent; nearby diagnostics, comparisons, defaults or converters deliberately handle that absence. |
| commandLineParser.ts:2083:21 | `options[opt.name]` | U-table | decline | This indexed option target is a pure store, not an indexed read. |
| commandLineParser.ts:2090:21 | `options[opt.name]` | U-table | decline | This indexed option target is a pure store, not an indexed read. |
| commandLineParser.ts:2090:70 | `args[i]` | U-position | decline | The next option argument may be absent; nearby diagnostics, comparisons, defaults or converters deliberately handle that absence. |
| commandLineParser.ts:2094:61 | `args[i]` | U-position | decline | The next option argument may be absent; nearby diagnostics, comparisons, defaults or converters deliberately handle that absence. |
| commandLineParser.ts:2095:21 | `options[opt.name]` | U-table | decline | This indexed option target is a pure store, not an indexed read. |
| commandLineParser.ts:2105:21 | `options[opt.name]` | U-table | decline | This indexed option target is a pure store, not an indexed read. |
| commandLineParser.ts:2105:101 | `args[i]` | U-position | decline | The next option argument may be absent; nearby diagnostics, comparisons, defaults or converters deliberately handle that absence. |
| commandLineParser.ts:2111:13 | `options[opt.name]` | U-table | decline | This indexed option target is a pure store, not an indexed read. |
| commandLineParser.ts:2281:57 | `jsonSourceFile.parseDiagnostics[0]` | U-endpoint | decline | The conditional selects the optional first parse diagnostic; the returned error field already accepts undefined. |
| commandLineParser.ts:2442:52 | `sourceFile.statements[0]` | U-endpoint | decline | The existing optional chain handles an empty JSON statement list. |
| commandLineParser.ts:2468:38 | `sourceFile.statements[0]` | U-endpoint | decline | The existing optional chain handles an empty JSON statement list. |
| commandLineParser.ts:2515:21 | `result[keyText]` | U-table | decline | This indexed target is a pure store, not an indexed read. |
| commandLineParser.ts:2690:29 | `computedOptions[option]` | U-table | assert | for-in enumerates populated entries in the compiler-created computedOptions table. |
| commandLineParser.ts:2691:34 | `computedOptions[option]` | U-table | assert | for-in enumerates populated entries in the compiler-created computedOptions table. |
| commandLineParser.ts:2693:17 | `impliedCompilerOptions[option]` | U-table | decline | This indexed target is a pure store, not an indexed read. |
| commandLineParser.ts:2693:50 | `computedOptions[option]` | U-table | assert | for-in enumerates populated entries in the compiler-created computedOptions table. |
| commandLineParser.ts:2707:25 | `computedOptions[option]` | U-table | decline | The existing optional chain handles a dependency name absent from computedOptions. |
| commandLineParser.ts:2721:9 | `specs[0]` | U-endpoint | decline | The equality test already accepts an absent first specification and preserves the existing branch result. |
| commandLineParser.ts:2795:27 | `options[name]` | U-table | decline | The existing CompilerOptionsValue cast includes undefined; the optional option value is handled by its converter. |
| commandLineParser.ts:2901:24 | `allSetOptions[0]` | U-endpoint | assert | The nonempty allSetOptions guard selects its first remaining key before emitOption removes that key. |
| commandLineParser.ts:2901:42 | `options[allSetOptions[0]]` | U-table | decline | emitOption accepts an optional option value and applies its existing default after selecting the populated key. |
| commandLineParser.ts:2901:50 | `allSetOptions[0]` | U-endpoint | assert | The nonempty allSetOptions guard selects its first remaining key before emitOption removes that key. |
| commandLineParser.ts:2933:24 | `options[setting]` | U-table | decline | The existing nullish default handles absence and must remain at this evaluation point. |
| commandLineParser.ts:2943:24 | `optionDeclarations.filter(c => c.name === settingName)[0]` | U-endpoint | decline | The following explicit Debug.fail handles a missing filtered option. |
| commandLineParser.ts:2977:13 | `result[name]` | U-table | decline | This indexed target is a pure store, not an indexed read. |
| commandLineParser.ts:2979:17 | `options[name]` | U-table | decline | The existing CompilerOptionsValue cast includes undefined; the optional option value is handled by its converter. |
| commandLineParser.ts:3229:58 | `raw[prop]` | U-table | decline | The raw JSON value is validated with null/array checks and its existing cast; malformed or absent values are handled nearby. |
| commandLineParser.ts:3230:25 | `raw[prop]` | U-table | decline | The raw JSON value is validated with null/array checks and its existing cast; malformed or absent values are handled nearby. |
| commandLineParser.ts:3231:32 | `raw[prop]` | U-table | decline | The raw JSON value is validated with null/array checks and its existing cast; malformed or absent values are handled nearby. |
| commandLineParser.ts:3268:13 | `options[option.name]` | U-table | decline | The explicit undefined test guards this optional option value and downstream type checks handle malformed values. |
| commandLineParser.ts:3269:27 | `options[option.name]` | U-table | decline | The explicit undefined test guards this optional option value and downstream type checks handle malformed values. |
| commandLineParser.ts:3295:9 | `(result ??= assign({}, options))[option.name]` | U-position | decline | This indexed target is a pure store, not an indexed read. |
| commandLineParser.ts:3313:9 | `(result ??= list.slice())[index]` | U-position | decline | This indexed target is a pure store, not an indexed read. |
| commandLineParser.ts:3322:22 | `mapLike[key]` | U-table | decline | The array check and substitution helper explicitly accept absent map entries. |
| commandLineParser.ts:3323:77 | `mapLike[key]` | U-table | decline | The array check and substitution helper explicitly accept absent map entries. |
| commandLineParser.ts:3325:9 | `(result ??= assign({}, mapLike))[key]` | U-position | decline | This indexed target is a pure store, not an indexed read. |
| commandLineParser.ts:3464:21 | `ownConfig.raw[propertyName]` | U-table | decline | The inherited raw option is tested for presence and map accepts undefined; the result target is a pure store. |
| commandLineParser.ts:3465:21 | `extendsRaw[propertyName]` | U-table | decline | The inherited raw option is tested for presence and map accepts undefined; the result target is a pure store. |
| commandLineParser.ts:3466:21 | `result[propertyName]` | U-table | decline | The inherited raw option is tested for presence and map accepts undefined; the result target is a pure store. |
| commandLineParser.ts:3466:48 | `extendsRaw[propertyName]` | U-table | decline | The inherited raw option is tested for presence and map accepts undefined; the result target is a pure store. |
| commandLineParser.ts:3542:30 | `(value as unknown[])[index]` | U-loop | decline | The unknown extends entry is deliberately checked with isString and invalid entries follow the diagnostic path. |
| commandLineParser.ts:3551:25 | `(valueExpression as ArrayLiteralExpression | undefined)?.elements[index]` | U-position | decline | The optional syntax node is accepted by diagnostic conversion, which falls back to a compiler diagnostic. |
| commandLineParser.ts:3557:114 | `(valueExpression as ArrayLiteralExpression | undefined)?.elements[index]` | U-position | decline | The optional syntax node is accepted by diagnostic conversion, which falls back to a compiler diagnostic. |
| commandLineParser.ts:3592:69 | `rootCompilerOptions[0]` | U-endpoint | assert | rootCompilerOptions exists only after append adds a root option property, so its first property name is populated. |
| commandLineParser.ts:3592:203 | `rootCompilerOptions[0]` | U-endpoint | assert | rootCompilerOptions exists only after append adds a root option property, so its first property name is populated. |
| commandLineParser.ts:3613:17 | `currentOption[option.name]` | U-table | decline | The option target is a pure store; the raw JSON argument value is accepted by the converter, including absence. |
| commandLineParser.ts:3785:13 | `(defaultOptions || (defaultOptions = {}))[opt.name]` | U-table | decline | The option target is a pure store; the raw JSON argument value is accepted by the converter, including absence. |
| commandLineParser.ts:3785:90 | `jsonOptions[id]` | U-table | decline | The option target is a pure store; the raw JSON argument value is accepted by the converter, including absence. |
| commandLineParser.ts:3887:120 | `valueExpression?.elements[index]` | U-position | decline | The optional syntax node is accepted by diagnostic conversion, which falls back to a compiler diagnostic. |
| commandLineParser.ts:4142:68 | `wildcardDirectories[existingPath]` | U-table | decline | existingFlags is explicitly tested for undefined before comparison and replacement. |
| commandLineParser.ts:4144:21 | `wildcardDirectories[existingPath == undefined ? existingPath : path]` | U-table | decline | This is a pure store or delete of a table entry, not an indexed read. |
| commandLineParser.ts:4159:32 | `wildcardDirectories[path]` | U-table | decline | This is a pure store or delete of a table entry, not an indexed read. |
| commandLineParser.ts:4185:33 | `match[0]` | U-regex | assert | A successful wildcardDirectoryPattern.exec always populates match[0] with the entire match. |
| commandLineParser.ts:4186:19 | `match[0]` | U-regex | assert | A successful wildcardDirectoryPattern.exec always populates match[0] with the entire match. |
| commandLineParser.ts:4248:21 | `extensionGroup[i]` | U-loop | assert | The bounded descending loop traverses the populated built-in extension group selected for this file. |
| commandLineParser.ts:4269:17 | `out[key]` | U-table | decline | The output target is a pure store and the input converter explicitly returns undefined for an absent option. |
| commandLineParser.ts:4269:59 | `opts[key]` | U-table | decline | The output target is a pure store and the input converter explicitly returns undefined for an absent option. |

## Wave 4 site ledger

| File:line:column | Class | Decision | Invariant / decline reason |
| --- | --- | --- | --- |
| moduleNameResolver.ts:368:19 | U-table | decline | readPackageJsonField deliberately validates absent or malformed JSON fields with typeof before returning. |
| moduleNameResolver.ts:474:43 | U-table | decline | readPackageJsonTypesVersionPaths validates the returned paths with typeof and diagnoses absent or malformed version entries. |
| moduleNameResolver.ts:954:59 | U-table | decline | The any-cast value is serialized by a helper that accepts undefined; this is not a required indexed value. |
| moduleNameResolver.ts:1255:80 | U-position | decline | The equality comparison accepts an absent character and uses it to reject a directory boundary. |
| moduleNameResolver.ts:1430:90 | U-table | decline | The enum-name diagnostic argument is optional and trace already accepts an absent name. |
| moduleNameResolver.ts:1435:94 | U-table | decline | The enum-name diagnostic argument is optional and trace already accepts an absent name. |
| moduleNameResolver.ts:2295:46 | U-table | decline | The target loader accepts unknown, including undefined, and explicitly validates or rejects the target nearby. |
| moduleNameResolver.ts:2351:54 | U-table | decline | The target loader accepts unknown, including undefined, and explicitly validates or rejects the target nearby. |
| moduleNameResolver.ts:2566:9 | U-endpoint | decline | The equality test accepts absence for an empty module name and retains the unscoped-package path. |
| moduleNameResolver.ts:2592:37 | U-parallel | decline | A shorter candidate path legitimately has no component here; the comparison rejects the self-name match. |
| moduleNameResolver.ts:2630:26 | U-table | decline | The resulting mainExport is explicitly tested for truthiness before loading. |
| moduleNameResolver.ts:2712:24 | U-table | decline | The target loader accepts unknown, including undefined, and explicitly validates or rejects the target nearby. |
| moduleNameResolver.ts:2718:28 | U-table | decline | The target loader accepts unknown, including undefined, and explicitly validates or rejects the target nearby. |
| moduleNameResolver.ts:2724:28 | U-table | decline | The target loader accepts unknown, including undefined, and explicitly validates or rejects the target nearby. |
| moduleNameResolver.ts:2729:28 | U-table | decline | The target loader accepts unknown, including undefined, and explicitly validates or rejects the target nearby. |
| moduleNameResolver.ts:2820:43 | U-table | decline | The target loader accepts unknown, including undefined, and explicitly validates or rejects the target nearby. |
| moduleNameResolver.ts:3177:34 | U-table | decline | forEach accepts an absent substitution list; malformed path mappings preserve their existing resolution behavior. |
| moduleSpecifiers.ts:146:17 | U-position | decline | The character comparison tests escape syntax and already handles an absent character without dereferencing it. |
| moduleSpecifiers.ts:332:12 | U-endpoint | decline | The optional tuple result can be empty; the result[1] guard and optional kind contract handle absence. |
| moduleSpecifiers.ts:332:33 | U-endpoint | decline | The optional tuple result can be empty; the result[1] guard and optional kind contract handle absence. |
| moduleSpecifiers.ts:332:62 | U-endpoint | decline | The optional tuple result can be empty; the result[1] guard and optional kind contract handle absence. |
| moduleSpecifiers.ts:801:22 | U-table | decline | Optional dependency fields are tested for presence and object type before enumeration. |
| moduleSpecifiers.ts:920:34 | U-endpoint | decline | The optional first ambient-module candidate is tested for presence before dereference. |
| moduleSpecifiers.ts:928:35 | U-table | assert | for-in traverses the populated compiler path-mapping entries, whose substitution arrays are required by this iterator. |
| moduleSpecifiers.ts:1111:35 | U-table | decline | The export/import target helper accepts unknown and explicitly checks its type, preserving absent or malformed targets. |
| moduleSpecifiers.ts:1134:122 | U-table | decline | The export/import target helper accepts unknown and explicitly checks its type, preserving absent or malformed targets. |
| moduleSpecifiers.ts:1165:121 | U-table | decline | The export/import target helper accepts unknown and explicitly checks its type, preserving absent or malformed targets. |
| moduleSpecifiers.ts:1387:13 | U-endpoint | decline | The existing switch default calls Debug.assertNever for an absent or invalid preferred ending; that failure handling remains unchanged. |
| moduleSpecifiers.ts:1411:38 | U-endpoint | decline | The existing switch default calls Debug.assertNever for an absent or invalid preferred ending; that failure handling remains unchanged. |
| sys.ts:155:17 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:171:13 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:171:29 | U-table | decline | The partial custom level may be absent; the existing || falls back to the default level. |
| sys.ts:171:53 | U-table | decline | The level key is keyof Levels and addresses an already required declared property; no indexed-absence finding exists. |
| sys.ts:195:29 | U-position | decline | Polling queues intentionally contain holes, and the nearby presence check handles them. |
| sys.ts:200:13 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:209:13 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:215:13 | U-position | decline | Polling queues intentionally contain holes, and the nearby presence check handles them. |
| sys.ts:218:17 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:219:17 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:290:84 | U-table | decline | The mapped polling-interval table declares every PollingInterval key as required; this read already has a required number type. |
| sys.ts:328:21 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:332:53 | U-table | decline | The mapped polling-interval table declares every PollingInterval key as required; this read already has a required number type. |
| sys.ts:338:17 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:343:17 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:484:73 | U-table | decline | The mapped polling-interval table declares every PollingInterval key as required; this read already has a required number type. |
| sys.ts:712:81 | U-position | decline | The equality test accepts an absent boundary character to reject an unrelated directory. |
| sys.ts:1566:24 | U-table | decline | The environment entry is optional and the existing || supplies the empty string. |
| sys.ts:1796:29 | U-typed | decline | The BOM length guard covers these bytes; equality checks already accept absence and no required-value diagnostic exists. |
| sys.ts:1796:51 | U-typed | decline | The BOM length guard covers these bytes; equality checks already accept absence and no required-value diagnostic exists. |
| sys.ts:1801:34 | U-typed | assert | The byte-pair loop rounds Buffer length down to an even number and advances by two, so both bytes are within the dense Buffer. |
| sys.ts:1802:21 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:1802:33 | U-typed | assert | The byte-pair loop rounds Buffer length down to an even number and advances by two, so both bytes are within the dense Buffer. |
| sys.ts:1803:21 | U-position | decline | The indexed target is a pure store, not an indexed read. |
| sys.ts:1807:29 | U-typed | decline | The BOM length guard covers these bytes; equality checks already accept absence and no required-value diagnostic exists. |
| sys.ts:1807:51 | U-typed | decline | The BOM length guard covers these bytes; equality checks already accept absence and no required-value diagnostic exists. |
| sys.ts:1811:29 | U-typed | decline | The BOM length guard covers these bytes; equality checks already accept absence and no required-value diagnostic exists. |
| sys.ts:1811:51 | U-typed | decline | The BOM length guard covers these bytes; equality checks already accept absence and no required-value diagnostic exists. |
| sys.ts:1811:73 | U-typed | decline | The BOM length guard covers these bytes; equality checks already accept absence and no required-value diagnostic exists. |

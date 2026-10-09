# Required indexed reads in program support, wave 1

This is a **partial partition 32**, on TypeScript 6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. It adds **15 required-read `!`
checks** in `semver.ts`, `programDiagnostics.ts`, and `tracing.ts`. It also
reviews `symbolWalker.ts` and `performance.ts` without changing them. The
ledger has one entry for each of the 22 parsed indexed expressions in these
five files: 15 asserted and seven declined. No U-zero site was selected.

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

## Census

The unchanged Adamic loader options check all 78 prepared compiler roots.
The exporter uses the same Go overlay as adaptation 30 and never invokes the
meter's implicit optional-declaration adaptation. These tables retain exactly
the reviewed five files and requested codes; full diagnostic chains are saved
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
| **Total before** | 18 | 18 | 2 | 3 | 0 | **41** |
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

Five-code findings fall from **41 to 10**. This does not mean all five files
pass the complete checker. `symbolWalker.ts` and `performance.ts` have no
checker findings both before and after. `programDiagnostics.ts` still has
TS2412, and `tracing.ts` still has ten findings across TS1294, TS18046, TS2379,
TS2412, and TS2591. These are outside this required-indexed-read wave.

## Proof

`verify.cjs` independently emits each file with stock 6.0.3 and requires
byte-identical JavaScript. It also checks newline and CRLF counts, every site
contract, and idempotence. All five files pass. A second actual adapter CLI
run reports zero assertions and zero U-zero edits for each file.

The mutant changes `left[i]!` in semver's prerelease comparison to
`(left[i] ?? 0)`. The independent emitted-JavaScript comparison fails with
`unexpected emitted JavaScript change in semver.ts`, exit 1. The adapter's
site check independently fails with `required read defaulted: semver.ts:171:32`,
exit 1. `mutant.cjs` reproduces both failures in a new scratch tree and
`mutant-proof.json` records them. The mutant is absent from the final tree.
No oracle or census mutant result is claimed.

The default stage 3 oracle passed: **106367 passing, 0 failing, 0 pending**,
all phase exits 0, and an empty (0-byte) `baseline.diff`. Install took 14.715s,
build 12.878s, tests 698.527s, total 726.255s. The tree contains 00, 10, 30,
and this support wave; adaptation 20 is absent. `oracle-report.json` records
the exact commands and result.
An earlier scratch oracle containing path.ts checks was stopped after the
ownership correction; it exited 1 and is not used as passing evidence.

## Reproduction

Prepare a pinned scratch tree with adaptations 00, 10, and 30 only, then copy
it to a separate after tree. From the repository:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
unit=stage3/adapt/32-indexed-reads-program
bash "$unit/census.sh" /tmp/adapt32-before /tmp/new-wave1-before > /tmp/new-wave1-before.log 2>&1
node "$unit/adapt.cjs" /tmp/adapt32-after > /tmp/new-wave1-adapt.log 2>&1
bash "$unit/census.sh" /tmp/adapt32-after /tmp/new-wave1-after > /tmp/new-wave1-after.log 2>&1
node "$unit/verify.cjs" /tmp/adapt32-before /tmp/adapt32-after > /tmp/new-wave1-verify.log 2>&1
node "$unit/adapt.cjs" /tmp/adapt32-after > /tmp/new-wave1-idempotence.log 2>&1
node "$unit/mutant.cjs" /tmp/adapt32-before /tmp/adapt32-after /tmp/new-wave1-mutant > /tmp/new-wave1-mutant.log 2>&1
stage3/oracle/run.sh /tmp/adapt32-after /tmp/new-wave1-oracle > /tmp/new-wave1-oracle.log 2>&1
```

`CENSUS_TYPESCRIPT` may select the absolute stock typescript.js path instead
of `NODE_PATH`. Actual logs use `/tmp/adapt32-wave1-*`; complete oracle phase
logs are under `/tmp/adapt32-wave1-oracle/`.

## Remaining partition work

Binder, program, commandLineParser, module resolution and specifiers, sys,
builders, watchers, resolutionCache, tsbuild files, expressionToTypeNode,
executeCommandLine, and other owned compiler files are not covered by wave 1.
Their source-local invariants have not yet been reviewed and no required-read
assertions are inferred for them. This wave does not establish the critical
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

# Required indexed reads, core, parser/scanner/factory, debug and path partitions

The adapter covers the three core/utilities files plus parser.ts, scanner.ts,
all ten factory files, debug.ts, and path.ts. The cumulative ledger has 227
required-read assertions, four U-zero operands, and 89 declines across 17 files;
per-wave evidence follows. Scanner slice coverage is complete; the parser slice
audit is held until namespace-member slicing lands.

The first wave changed `src/compiler/core.ts`, `utilities.ts`, and
`utilitiesPublic.ts` in TypeScript 6.0.3, pinned at
`050880ce59e30b356b686bd3144efe24f875ebc8`. It adds **144 `!` checks** and
**four numeric U-zero expressions**, and records **43 declined sites** below.
Every original read, repeated read, callback boundary, increment, and compound
assignment stays at its original evaluation point. Text insertion preserves
upstream CRLF bytes. No source printer, optional chain, skip, or hoisted read is
used. The only defaults are the four reviewed numeric bitwise encoder operands.

The user's ruling makes `!` a loud required-value check in Adamic and erased
syntax on Node. These are checked invariant obligations, not proofs that length
bounds alone establish density or that arbitrary generic payloads exclude
undefined. Caller arrays/records must have defined values at the listed required
positions; callbacks must preserve any slots read after them. Invalid states
can fail loudly in Adamic even when Node's erased source would continue.

The classes are those in `docs/tsc-strictness.md`: U-loop, U-parallel, U-endpoint,
U-position, U-table, U-grid, U-regex, U-zero, and U-typed. Selection was reviewed
against parsed reads and the actual five-code census. `sites.json` records each
expression and occurrence; line/column numbers document the pinned source and
are never edit addresses. The adapter parses current text with stock 6.0.3,
validates occurrence counts, plans every scoped file before writing, and fails
on unexpected site shapes. Later partitions must revise their own reviewed
ledger rather than blanket-asserting every indexed expression.

## Rules learned from the first wave

- Keep all five requested diagnostic codes in the census, including unrelated
  overload errors. A lower count is not a claim that the file now passes the
  checker; give each residual finding its actual cause.
- Trace an indexed value through its local variable before asserting it. The
  decoder's `nextCode: number` annotation looked required, but exhaustion was
  deliberately handled by the following bitwise continuation test. An explicit
  number annotation alone does not establish a presence obligation.
- Treat emitted JavaScript as a separate proof. Required `!` additions must
  erase byte for byte; allow only the individually reviewed U-zero operands in
  the comparison. A populated-input oracle can miss an illegal `?? 0` default.
- Record a repeated read separately even after a nearby absence guard. The
  addRange decline intentionally leaves a diagnostic because eliminating it
  would obscure its legitimate absence handling or change an observable read.
- Pin expression occurrence counts, not historical source offsets. Type-import
  adaptation changes columns; parse the current text, validate all files before
  writing, and reject site drift rather than guessing a replacement location.

## Wave 3 ownership notice

Program partition: adaptation 30 now owns `src/compiler/debug.ts` and
`src/compiler/path.ts`; remove these two files from partition 32. Ownership
supplied by the user: 31 owns checker.ts; 33 owns emitter.ts, sourcemap.ts,
transformer.ts, visitorPublic.ts, and transformers/; 32 owns every remaining
top-level src/compiler/*.ts outside 30 and 31/33. The scanner closure is covered.
The parser audit is on hold until namespace-member slicing lands.

## Tree and scope

Adamic base: `ef3d907ecdc4c771b016f7d9c52372def057a340`.
Pipeline: `origin/codex/stage3-base` at `8728405135d329efc12c837a7a6c293234abbe1c`.
Adaptation 10: **`a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e`**.
**Adaptation 20 was excluded**, as instructed. Both direct application and a
fresh `apply.sh` discovery run produced byte-identical adapted versions of the
three files. Only this unit's directory is committed; upstream and pipeline
scratch copies remain outside the repository.

Toolchain setup passed: Go 0s, clang 0s, Node 0s, submodules 1s, build cache 127s,
total 127s. `nproc` was 5, CPU quota 4. Environment:
`source /workspace/adamic-tools/env.sh`; Node 24.19.0, Go 1.27.1, clang 20.1.8.

## First-wave census

The real Adamic loader/checker, with unchanged fixed strictness options, checked
78 prepared compiler roots. The exporter uses a Go overlay and **no implicit
optional-declaration adaptation**. The filter retains exactly the three owned
paths and all findings bearing the five requested codes, even when their root
cause is not undefined. Checked-in JSON preserves each full diagnostic chain.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| core.ts before | 66 | 0 | 3 | 10 | 2 | 81 |
| core.ts after | 2 | 0 | 0 | 0 | 0 | 2 |
| utilities.ts before | 28 | 6 | 28 | 6 | 0 | 68 |
| utilities.ts after | 1 | 1 | 0 | 2 | 0 | 4 |
| utilitiesPublic.ts before | 6 | 7 | 1 | 2 | 0 | 16 |
| utilitiesPublic.ts after | 0 | 0 | 0 | 0 | 0 | 0 |
| **Total before** | **100** | **13** | **32** | **18** | **2** | **165** |
| **Total after** | **3** | **1** | **0** | **2** | **0** | **6** |

Remaining findings:

- `core.ts:992`, TS2345: addRange explicitly handles undefined nearby; its
  repeated read stays unchanged despite the checker's lost narrowing.
- `core.ts:1721`, TS2345: MultiMap.add's candidate is not an indexed read.
- `utilities.ts:1330`, TS18048: a split/map callback parameter, not an indexed
  read in the source, under the checker's regex/string model.
- `utilities.ts:7721` and `7725`, TS2322: decoder exhaustion is intentional;
  the following bitwise continuation test handles the undefined probe.
- `utilities.ts:11201`, TS2345: String.replace.call overload mismatch, unrelated
  to indexed absence.

These three files still do not pass the full Adamic checker. This partition
removes required indexed-read findings; it does not implement the compiler's
other language features or repair legitimate optional/non-indexed contracts.

## First-wave proof and mutants

`verify.cjs` independently emits JavaScript with stock TypeScript. It requires
byte-identical output after allowing exactly the four reviewed U-zero operands,
checks every site contract, and verifies idempotence. A real second CLI run also
reported zero assertions and zero U-zero edits in all three files.
`zero-probe.cjs` extracts the original encoder/decoder functions and runs 12
cases against Node Buffer: empty input, all tail lengths, zero bytes, and UTF-8.
It covers the declined decoder exhaustion reads as well.

The required mutant replaces the newly inserted `arr[0]!` at
`utilities.ts:10364:24` with `(arr[0] ?? 0)`. It is caught by:

- Adamic census: **6 -> 7**, adding TS2345 because zero is not an arbitrary T.
- Stock emitted-JavaScript comparison: unexpected bytes in utilities.ts, exit 1.
- Site contract: `required read defaulted: utilities.ts:10364:24`, exit 1.

The base64-only probe passes this required-read mutant, as expected from its
scope. A second mutant changes one U-zero tail default from 0 to 64; that probe
fails on input `a`, **`YU==` versus Node Buffer's `YQ==`**, exit 1. Neither mutant
is retained in the adapted tree. The default oracle was not rerun on mutants;
no claim is made that ordinary populated-input suite cases observe a `?? 0`
replacement. `proof.json` and `census-mutant.json` retain the exact observations.

The **default stage 3 oracle passed: 106,367 passing, 0 failing, 0 pending**.
Install, build, and tests all exited 0; runners were `all`, tests unfiltered,
workers 4, and the upstream run was `--light=false`. `baseline.diff` is **empty
(0 bytes)**; no baseline was accepted or modified. Install took 2.273s, build
23.068s, tests 331.654s, total 357.034s. `oracle-report.json` preserves the exact
commands and counts. Lint remains disabled by the default oracle; browser and
ESLint-rule integrations and a native Adamic tsc build were not run.

## Reproduction

From the Adamic repository, with stock TypeScript 6.0.3 on `NODE_PATH` or supplied
through `CENSUS_TYPESCRIPT`:

```sh
source /workspace/adamic-tools/env.sh
unit=stage3/adapt/30-indexed-reads
bash "$unit/census.sh" /tmp/tsc-before /tmp/indexed-before > /tmp/indexed-before.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-after > /tmp/indexed-adapt.log 2>&1
bash "$unit/census.sh" /tmp/tsc-after /tmp/indexed-after > /tmp/indexed-after.log 2>&1
node "$unit/verify.cjs" /tmp/tsc-before /tmp/tsc-after > /tmp/indexed-verify.log 2>&1
node "$unit/zero-probe.cjs" /tmp/tsc-before /tmp/tsc-after > /tmp/indexed-zero.log 2>&1
node "$unit/adapt.cjs" /tmp/tsc-after > /tmp/indexed-idempotence.log 2>&1
stage3/oracle/run.sh /tmp/tsc-after /tmp/indexed-oracle > /tmp/indexed-oracle.log 2>&1
```

Prepare `/tmp/tsc-before` with setup and adaptation 10 only, then copy it to
`/tmp/tsc-after` before adapting. In a scratch worktree of the pipeline at
`a3ef0dc`, copy this directory into `stage3/adapt/30-indexed-reads` after building
the before tree, then run `apply.sh` again at a fresh output path. This commit
does not carry pipeline files on main. The observed scratch paths are
`/tmp/stage3-indexed-10-harness`, `/tmp/stage3-indexed-pristine` (three-file snapshot),
`/tmp/stage3-indexed-10-tree`, and `/tmp/stage3-indexed-pipeline-tree`. Census logs
are `/tmp/stage3-indexed-census-{before,after}.log`; oracle phase logs are in
`/tmp/stage3-indexed-10-oracle/`. All command output was redirected to logs.

## Second wave: parser, scanner, and factory

Second-wave toolchain setup passed: Go 0s, clang 1s, Node 1s, submodules 1s,
build cache 24s, total 24s. `nproc` was 5, CPU quota 4. Versions match the first
wave. The first-wave findings and source changes remain intact.

The adapter now also validates `parser.ts`, `scanner.ts`, and all ten
`factory/*.ts` files. This wave adds **46 required-read assertions**, **36
new declined sites**, and **no U-zero defaults**. Combined with the first wave,
the growing ledger contains **190 assertions, four U-zero operands, and 79
declines** across 15 files. Six factory files have no required indexed reads;
they remain explicitly included in the census and emitted-JavaScript check.
No file outside the two assigned partitions is adapted.

The bounded Unicode search needs its populated start/end pair at each even
position. The factory trampoline's current node is installed before entering
each frame; onEnter populates that frame's user state before left/operator/right
and exit, while root previous-state absence stays optional. Debug equality
operands and the loop condition remain separate unchanged reads. The compiler's
checker/emitter callbacks return populated WorkArea states; the generic API's
required-state obligation is recorded rather than inferred from array bounds.

New census, on exactly these twelve files with the same five codes:

| File | Before | After |
| --- | ---: | ---: |
| parser.ts | 23 | 6 |
| scanner.ts | 9 | 2 |
| factory/baseNodeFactory.ts | 0 | 0 |
| factory/emitHelpers.ts | 5 | 0 |
| factory/emitNode.ts | 3 | 0 |
| factory/nodeChildren.ts | 0 | 0 |
| factory/nodeConverters.ts | 0 | 0 |
| factory/nodeFactory.ts | 4 | 0 |
| factory/nodeTests.ts | 0 | 0 |
| factory/parenthesizerRules.ts | 0 | 0 |
| factory/utilities.ts | 11 | 0 |
| factory/utilitiesPublic.ts | 0 | 0 |
| **Total** | **55** | **8** |

`wave2-census-before.json` and `wave2-census-after.json` preserve the per-code
counts and complete chains. The remaining eight findings are:

- parser.ts:10458, TS2322: SyntaxCursor promises Node, but the implementation
  deliberately returns undefined on misses and handles misses locally.
- parser.ts:10656, TS2322: the optional AMD dependency name is forwarded in
  an object literal; this is an optional-declaration issue outside this unit.
- parser.ts:10735 and 10741, TS2322, and 10737, TS18048: the existing OR over
  alternative quote captures has an optional result. Each individual capture
  legitimately can be absent; a required result annotation would concern the
  OR expression, not an indexed read. An empty single-quoted capture also
  falls through to the absent other branch, so presence must not be assumed.
- parser.ts:10794, TS2322: optional pragma arguments can be absent and the
  original dictionary write forwards their value.
- scanner.ts:491, TS2532 and TS2322: next-line lookahead can be absent on the
  final line. The comparison then falls through to the text-length/return
  branch; asserting either repeated lookahead would change this behavior.

Stock TypeScript emitted JavaScript is byte-identical for every new file.
Both in-memory idempotence and a real second CLI run pass, with zero edits in
all 15 scoped files. A fresh apply.sh discovery run uses setup, adaptation 10
at a3ef0dc, and the extended adaptation 30; adaptation 20 remains excluded.

The new mutant replaces `map[0]!` at scanner.ts:360:16 with `(map[0] ?? 0)`.
**The census does not catch it:** all eight diagnostics remain identical,
because the replacement is still a number. Stock emitted-JavaScript comparison
fails on scanner.ts and the site-contract check reports `required read
defaulted: scanner.ts:360:16`, both exit 1. The default oracle was not rerun
on this mutant; no passing populated-input suite is claimed to detect it.
`wave2-census-mutant.json` records the observed coverage limit.

The second-wave **default oracle passed: 106,367 passing, 0 failing, 0
pending**. Install, build, and unfiltered `all` suites each exited 0, with
four workers and `--light=false`. `wave2-baseline.diff` is **empty (0 bytes)**;
no baseline was accepted or changed. Install took 2.377s, build 22.467s, tests
336.481s, total 361.366s. `wave2-oracle-report.json` and `wave2-proof.json`
retain the commands, counts, hashes, idempotence, and mutant observations.
Lint, browser/ESLint integrations, and native Adamic tsc compilation were not
run. The eight new residual findings are preserved deliberately.

To reproduce this wave's census, append `--wave2` to census.sh's two existing
arguments, for example:

```sh
bash stage3/adapt/30-indexed-reads/census.sh /tmp/tsc-before /tmp/wave2-before --wave2 > /tmp/wave2-before.log 2>&1
bash stage3/adapt/30-indexed-reads/census.sh /tmp/tsc-after /tmp/wave2-after --wave2 > /tmp/wave2-after.log 2>&1
```

The before tree has adaptation 10 and the first indexed-read wave; the new
files are untouched there. The current adapted tree is
`/tmp/stage3-indexed-pipeline-tree`; the fresh discovery tree is
`/tmp/stage3-indexed-wave2-discovery`. The 15-file original-source snapshot for
emitted-JavaScript comparison is `/tmp/stage3-indexed-pristine`. Raw second-wave
census, verification, idempotence, and mutant logs use the
`/tmp/stage3-indexed-wave2-*` prefix. Oracle phase logs are in
`/tmp/stage3-indexed-wave2-oracle/`.

## Site ledger

Each row identifies one read in the pinned source, its class, action, and the
one-line invariant or decline reason. Repeated expressions on the same line
have separate columns and separate rows. Pure stores and existing casts are
listed when encountered in the possibly-undefined inventory, but not adapted.

### First wave

| Site | Read | Class | Action | Invariant or decline reason |
| --- | --- | --- | --- | --- |
| core.ts:36:37 | `array[i]` | U-loop | assert | The forEach loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:53:37 | `array[i]` | U-loop | assert | The forEachRight loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:73:33 | `array[i]` | U-loop | assert | The firstDefined loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:110:30 | `arrayA[i]` | U-parallel | assert | zipWith asserts equal lengths; both caller arrays are populated at the shared index. |
| core.ts:110:41 | `arrayB[i]` | U-parallel | assert | zipWith asserts equal lengths; both caller arrays are populated at the shared index. |
| core.ts:128:21 | `input[i]` | U-loop | assert | The intersperse loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:148:27 | `array[i]` | U-loop | assert | The every loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:169:23 | `array[i]` | U-loop | assert | The find loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:185:23 | `array[i]` | U-loop | assert | The findLast loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:201:23 | `array[i]` | U-loop | assert | The findIndex loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:212:23 | `array[i]` | U-loop | assert | The findLastIndex loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:223:34 | `array[i]` | U-loop | assert | The contains loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:246:23 | `array[i]` | U-loop | assert | The countWhere loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:281:29 | `array[i]` | U-loop | assert | The filter loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:286:30 | `array[i]` | U-loop | assert | The filter loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:302:15 | `array[i]` | U-loop | assert | The filterMutate loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:303:31 | `array[i]` | U-loop | assert | The filterMutate loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:325:27 | `array[i]` | U-loop | assert | The map loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:353:26 | `array[i]` | U-loop | assert | The sameMap loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:359:35 | `array[i]` | U-loop | assert | The sameMap loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:378:19 | `array[i]` | U-position | decline | flatten deliberately ignores falsy or absent entries. |
| core.ts:403:29 | `array[i]` | U-loop | assert | The flatMap loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:422:29 | `array[i]` | U-loop | assert | The flatMapToMutable loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:462:26 | `array[i]` | U-loop | assert | The sameFlatMap loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:484:30 | `array[i]` | U-loop | assert | The mapAllOrFail loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:498:34 | `array[i]` | U-loop | assert | The mapDefined loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:566:31 | `array[pos]` | U-loop | assert | The spanMap loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:622:31 | `array[i]` | U-loop | assert | The some loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:642:18 | `arr[i]` | U-loop | assert | The getRangesWhere loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:693:16 | `array[indices[0]]` | U-position | assert | indices is a nonempty permutation of positions in the populated input; its first and traversed entries remain valid. |
| core.ts:693:22 | `indices[0]` | U-endpoint | assert | indices is a nonempty permutation of positions in the populated input; its first and traversed entries remain valid. |
| core.ts:694:37 | `indices[0]` | U-endpoint | assert | indices is a nonempty permutation of positions in the populated input; its first and traversed entries remain valid. |
| core.ts:696:23 | `indices[i]` | U-position | assert | indices is a nonempty permutation of positions in the populated input; its first and traversed entries remain valid. |
| core.ts:697:22 | `array[index]` | U-position | assert | indices is a nonempty permutation of positions in the populated input; its first and traversed entries remain valid. |
| core.ts:706:34 | `array[i]` | U-position | assert | indices is a nonempty permutation of positions in the populated input; its first and traversed entries remain valid. |
| core.ts:712:30 | `array[i]` | U-loop | assert | The deduplicateEquality loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:738:16 | `array[0]` | U-endpoint | assert | The length < 2 return precedes the first read; later reads traverse the populated sorted input. |
| core.ts:741:22 | `array[i]` | U-loop | assert | The length < 2 return precedes the first read; later reads traverse the populated sorted input. |
| core.ts:784:53 | `array[idx - 1]` | U-position | assert | Binary search returns an insertion position; adjacent reads have explicit endpoint guards. |
| core.ts:787:64 | `array[idx]` | U-position | assert | Binary search returns an insertion position; adjacent reads have explicit endpoint guards. |
| core.ts:824:31 | `array1[i]` | U-parallel | assert | Equal lengths precede traversal; both caller arrays supply values at the shared index. |
| core.ts:824:42 | `array2[i]` | U-parallel | assert | Equal lengths precede traversal; both caller arrays supply values at the shared index. |
| core.ts:850:23 | `array[i]` | U-position | decline | compact deliberately drops undefined and other falsy entries. |
| core.ts:878:53 | `arrayB[offsetB]` | U-loop | assert | Nested bounds cover the populated sorted inputs; predecessor reads additionally require offset > 0. |
| core.ts:878:70 | `arrayB[offsetB - 1]` | U-loop | assert | Nested bounds cover the populated sorted inputs; predecessor reads additionally require offset > 0. |
| core.ts:886:57 | `arrayA[offsetA]` | U-loop | assert | Nested bounds cover the populated sorted inputs; predecessor reads additionally require offset > 0. |
| core.ts:886:74 | `arrayA[offsetA - 1]` | U-loop | assert | Nested bounds cover the populated sorted inputs; predecessor reads additionally require offset > 0. |
| core.ts:889:30 | `arrayB[offsetB]` | U-loop | assert | Nested bounds cover the populated sorted inputs; predecessor reads additionally require offset > 0. |
| core.ts:889:47 | `arrayA[offsetA]` | U-loop | assert | Nested bounds cover the populated sorted inputs; predecessor reads additionally require offset > 0. |
| core.ts:894:33 | `arrayB[offsetB]` | U-loop | assert | Nested bounds cover the populated sorted inputs; predecessor reads additionally require offset > 0. |
| core.ts:991:13 | `from[i]` | U-position | decline | addRange explicitly tests for undefined; keep this absence test. |
| core.ts:992:21 | `from[i]` | U-position | decline | addRange explicitly handles absence nearby; preserve the repeated read after its test even though the checker cannot narrow it. |
| core.ts:1030:37 | `array[x]` | U-position | assert | The sort compares positions generated from the same populated array. |
| core.ts:1030:47 | `array[y]` | U-position | assert | The sort compares positions generated from the same populated array. |
| core.ts:1045:15 | `array[i]` | U-loop | assert | The arrayReverseIterator loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:1052:13 | `array1[pos]` | U-position | decline | rangeEquals compares possibly absent entries directly without requiring a present value. |
| core.ts:1052:29 | `array2[pos]` | U-position | decline | rangeEquals compares possibly absent entries directly without requiring a present value. |
| core.ts:1072:24 | `array[offset]` | U-position | decline | elementAt returns undefined for an unavailable position. |
| core.ts:1084:68 | `array[0]` | U-endpoint | decline | firstOrUndefined exposes an optional result. |
| core.ts:1100:12 | `array[0]` | U-endpoint | assert | Debug.assert(array.length !== 0) and the caller populated-array contract require the first value. |
| core.ts:1117:68 | `array[array.length - 1]` | U-endpoint | decline | lastOrUndefined exposes an optional result. |
| core.ts:1123:12 | `array[array.length - 1]` | U-endpoint | assert | Debug.assert(array.length !== 0) and the caller populated-array contract require the last value. |
| core.ts:1133:11 | `array[0]` | U-endpoint | decline | singleOrUndefined exposes an optional result. |
| core.ts:1162:11 | `array[0]` | U-endpoint | decline | singleOrMany exposes an optional result. |
| core.ts:1211:36 | `array[middle]` | U-position | assert | Binary search maintains low <= middle <= high within the populated input. |
| core.ts:1240:26 | `array[pos]` | U-position | assert | The initial position and every reduction position lie in the populated caller input. |
| core.ts:1247:36 | `array[pos]` | U-loop | assert | The initial position and every reduction position lie in the populated caller input. |
| core.ts:1279:44 | `map[key]` | U-table | decline | getProperty returns undefined for a missing own key. |
| core.ts:1316:25 | `(collection as MapLike<T>)[key]` | U-table | assert | Own-key enumeration selects a value in the caller record; required payloads must be defined. |
| core.ts:1376:35 | `left[key]` | U-table | assert | Own-property guards select corresponding entries in both caller records; comparer payloads must be defined. |
| core.ts:1376:46 | `right[key]` | U-table | assert | Own-property guards select corresponding entries in both caller records; comparer payloads must be defined. |
| core.ts:1412:23 | `array[i]` | U-loop | assert | The arrayToMap loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:1427:23 | `array[i]` | U-loop | assert | The arrayToNumericMap loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:1441:23 | `values[i]` | U-loop | assert | The arrayToMultiMap loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:1468:27 | `values[i]` | U-loop | assert | The groupBy loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:1586:24 | `elements[headIndex]` | U-position | decline | dequeue already casts its read; casts are outside this required-read partition. |
| core.ts:1587:9 | `elements[headIndex]` | U-position | decline | This is a pure write clearing the queue slot, not a read. |
| core.ts:1677:32 | `candidates[i]` | U-position | assert | Small-set candidates has exactly two populated entries; i and 1 - i select those entries. |
| core.ts:1682:48 | `candidates[1 - i]` | U-parallel | assert | Small-set candidates has exactly two populated entries; i and 1 - i select those entries. |
| core.ts:2004:38 | `arr[i]` | U-loop | assert | The maxBy loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2220:42 | `s1[i - 1]` | U-loop | assert | The levenshteinWithMax loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2220:70 | `s2[j - 1]` | U-loop | assert | The levenshteinWithMax loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2327:13 | `array[i]` | U-position | decline | Equality comparison can legitimately compare an undefined item. |
| core.ts:2343:20 | `array[i + 1]` | U-loop | assert | The orderedRemoveItemAt loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2350:20 | `array[array.length - 1]` | U-loop | assert | The unorderedRemoveItemAt loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2366:23 | `array[i]` | U-loop | assert | The unorderedRemoveFirstItemWhere loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2418:19 | `values[i]` | U-loop | assert | The findBestPatternMatch loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2500:25 | `newItems[newIndex]` | U-loop | assert | The enumerateInsertsAndDeletes loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2501:25 | `oldItems[oldIndex]` | U-loop | assert | The enumerateInsertsAndDeletes loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2520:18 | `newItems[newIndex++]` | U-loop | assert | The enumerateInsertsAndDeletes loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2524:17 | `oldItems[oldIndex++]` | U-loop | assert | The enumerateInsertsAndDeletes loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2538:27 | `arrays[index]` | U-position | assert | The worker index is below the outer array length; recursive calls advance only while another populated factor exists. |
| core.ts:2564:41 | `array[index]` | U-loop | assert | The takeWhile loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| core.ts:2580:41 | `array[index]` | U-loop | assert | The skipWhile loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| utilities.ts:939:31 | `newResolutions[i]` | U-position | decline | An absent resolution is explicitly compared and handled in this loop. |
| utilities.ts:940:23 | `names[i]` | U-parallel | assert | names and newResolutions share positions; the name is required even when its resolution is absent. |
| utilities.ts:1017:12 | `getLineStarts(sourceFile)[line]` | U-position | assert | The caller line selects a populated entry in the computed line-start map. |
| utilities.ts:1041:23 | `lineStarts[lineIndex]` | U-position | assert | lineIndex is a valid line; the next-line read is guarded against the final line. |
| utilities.ts:1043:19 | `lineStarts[lineIndex + 1]` | U-position | assert | lineIndex is a valid line; the next-line read is guarded against the final line. |
| utilities.ts:1121:34 | `to[statementIndex]` | U-loop | assert | The insertStatementsAfterPrologue loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| utilities.ts:1134:34 | `to[statementIndex]` | U-loop | assert | The insertStatementAfterPrologue loop traverses populated caller input at its bounded index; callbacks must preserve any subsequently required slot. |
| utilities.ts:1251:34 | `node.jsDoc[0]` | U-endpoint | assert | hasJSDocNodes establishes a nonempty parser-built JSDoc array. |
| utilities.ts:2355:55 | `info.declaration.parameters[0]` | U-endpoint | assert | An index-signature declaration has its grammar-required populated first parameter. |
| utilities.ts:2577:79 | `(node as CaseOrDefaultClause).statements[0]` | U-endpoint | assert | The case/default branch checks statements.length > 0 before the first statement. |
| utilities.ts:3108:28 | `tsConfigSourceFile.statements[0]` | U-endpoint | assert | The nonzero statement count guards the parser-built first expression statement. |
| utilities.ts:3785:77 | `(getLeftmostAccessExpression(node.initializer) as CallExpression).arguments[0]` | U-endpoint | decline | An existing string-literal cast already covers this indexed read; no new assertion is required. |
| utilities.ts:3829:10 | `node.typeArguments[0]` | U-endpoint | assert | The JSDoc index-signature shape checks typeArguments.length === 2. |
| utilities.ts:3829:69 | `node.typeArguments[0]` | U-endpoint | assert | The JSDoc index-signature shape checks typeArguments.length === 2. |
| utilities.ts:3856:17 | `args[0]` | U-endpoint | assert | args.length === 1 guards the parser-built argument used by the required node predicate. |
| utilities.ts:4125:38 | `expr.arguments[1]` | U-endpoint | assert | The call-shape guard requires exactly three populated arguments. |
| utilities.ts:4126:40 | `expr.arguments[0]` | U-endpoint | assert | The call-shape guard requires exactly three populated arguments. |
| utilities.ts:4353:20 | `findAncestor(node.initializer, (node): node is RequireOrImportCall => isRequireCall(node, /*requireStringLiteralLikeArgument*/ true))?.arguments[0]` | U-endpoint | decline | The existing optional ancestor/result path handles absence. |
| utilities.ts:4418:20 | `node.arguments[0]` | U-endpoint | decline | The API returns an optional external module name. |
| utilities.ts:5210:64 | `heritageClause.types[0]` | U-endpoint | decline | The API exposes an optional first heritage element. |
| utilities.ts:6086:20 | `diagnostics[result]` | U-position | decline | lookup returns an optional diagnostic. |
| utilities.ts:6088:68 | `diagnostics[~result - 1]` | U-position | assert | The binary-search insertion position has an explicit nonnegative predecessor guard. |
| utilities.ts:6089:20 | `diagnostics[~result - 1]` | U-position | decline | lookup returns an optional predecessor diagnostic. |
| utilities.ts:6286:25 | `indentStrings[1]` | U-endpoint | assert | The cache starts with levels 0 and 1 and appends every level through the requested nonnegative integer. |
| utilities.ts:6288:28 | `indentStrings[current - 1]` | U-position | assert | The cache starts with levels 0 and 1 and appends every level through the requested nonnegative integer. |
| utilities.ts:6290:12 | `indentStrings[level]` | U-position | assert | The cache starts with levels 0 and 1 and appends every level through the requested nonnegative integer. |
| utilities.ts:6294:12 | `indentStrings[1]` | U-endpoint | assert | indentStrings is initialized with the four-space string at position 1. |
| utilities.ts:6766:84 | `accessor.parameters[0]` | U-endpoint | assert | The nonempty parser-built parameter list guards the predicate on its first parameter. |
| utilities.ts:6767:16 | `accessor.parameters[hasThis ? 1 : 0]` | U-position | decline | The accessor API returns an optional parameter. |
| utilities.ts:6785:31 | `signature.parameters[0]` | U-endpoint | assert | The nonempty parser-built parameter list guards the first-parameter predicate. |
| utilities.ts:6945:62 | `leadingComments[0]` | U-endpoint | assert | The nonempty leading-comment list guards each repeated first-comment read. |
| utilities.ts:6946:104 | `leadingComments[0]` | U-endpoint | assert | The nonempty leading-comment list guards each repeated first-comment read. |
| utilities.ts:7090:19 | `lineMap[currentLine + 1]` | U-position | assert | Computed line numbers select populated starts; the final-line case supplies text.length + 1. |
| utilities.ts:7095:68 | `lineMap[firstCommentLineAndCharacter.line]` | U-position | assert | Computed line numbers select populated starts; the final-line case supplies text.length + 1. |
| utilities.ts:7618:78 | `symbol.declarations[0]` | U-endpoint | assert | length(symbol.declarations) > 0 guards the first declaration. |
| utilities.ts:7684:17 | `charCodes[i]` | U-loop | assert | i < charCodes.length selects the populated first byte of each input block. |
| utilities.ts:7685:18 | `charCodes[i]` | U-loop | assert | i < charCodes.length selects the populated first byte of each input block. |
| utilities.ts:7685:52 | `charCodes[i + 1]` | U-zero | zero | A missing numeric tail byte contributes zero through bitwise coercion before the output-padding branch. |
| utilities.ts:7686:18 | `charCodes[i + 1]` | U-zero | zero | A missing numeric tail byte contributes zero through bitwise coercion before the output-padding branch. |
| utilities.ts:7686:56 | `charCodes[i + 2]` | U-zero | zero | A missing numeric tail byte contributes zero through bitwise coercion before the output-padding branch. |
| utilities.ts:7687:17 | `charCodes[i + 2]` | U-zero | zero | A missing numeric tail byte contributes zero through bitwise coercion before the output-padding branch. |
| utilities.ts:7712:26 | `codes[i]` | U-loop | assert | i < codes.length guards the current byte; exhaustion probes are separately declined. |
| utilities.ts:7721:36 | `codes[i]` | U-position | decline | The decoder probes past the final byte; undefined & mask terminates the continuation loop. |
| utilities.ts:7725:28 | `codes[i]` | U-position | decline | The decoder probes past the final byte; undefined & mask terminates the continuation loop. |
| utilities.ts:7762:42 | `input[i]` | U-position | assert | Compiler callers supply complete four-character base64 blocks, including padding; each digit position is required. |
| utilities.ts:7763:42 | `input[i + 1]` | U-position | assert | Compiler callers supply complete four-character base64 blocks, including padding; each digit position is required. |
| utilities.ts:7764:42 | `input[i + 2]` | U-position | assert | Compiler callers supply complete four-character base64 blocks, including padding; each digit position is required. |
| utilities.ts:7765:42 | `input[i + 3]` | U-position | assert | Compiler callers supply complete four-character base64 blocks, including padding; each digit position is required. |
| utilities.ts:8286:51 | `symbol.declarations[0]` | U-endpoint | decline | The explicit truthiness test handles an absent declaration before the repeated read. |
| utilities.ts:8302:31 | `children[i]` | U-position | decline | nodeIsPresent explicitly accepts undefined. |
| utilities.ts:8303:29 | `children[i]` | U-position | decline | The destination and return type allow undefined after nodeIsPresent. |
| utilities.ts:8584:90 | `args[+index]` | U-position | decline | Debug.checkDefined already performs the required runtime check. |
| utilities.ts:8605:43 | `localizedDiagnosticMessages[message.key]` | U-table | decline | A missing localized key explicitly falls back to message.message. |
| utilities.ts:8770:21 | `lastChain.next[0]` | U-endpoint | assert | The single-chain representation gives every present next list a populated first link. |
| utilities.ts:8806:25 | `d2.relatedInformation[index]` | U-parallel | assert | Equal related-information lengths precede traversal of the populated diagnostic arrays. |
| utilities.ts:8889:39 | `c1[i]` | U-parallel | assert | Equal chain lengths precede recursive size comparison of populated elements. |
| utilities.ts:8889:51 | `c2[i]` | U-parallel | assert | Equal chain lengths precede recursive size comparison of populated elements. |
| utilities.ts:8905:43 | `c1[i]` | U-parallel | assert | The preceding size/shape comparison establishes corresponding populated chain elements. |
| utilities.ts:8905:62 | `c2[i]` | U-parallel | assert | The preceding size/shape comparison establishes corresponding populated chain elements. |
| utilities.ts:8909:13 | `c1[i]` | U-parallel | assert | The preceding size/shape comparison establishes corresponding populated chain elements. |
| utilities.ts:8912:42 | `c1[i]` | U-parallel | assert | The preceding size/shape comparison establishes corresponding populated chain elements. |
| utilities.ts:8912:55 | `c2[i]` | U-parallel | assert | The preceding size/shape comparison establishes corresponding populated chain elements. |
| utilities.ts:9370:12 | `compilerOptions[flag]` | U-table | decline | The strict-option fallback explicitly handles undefined. |
| utilities.ts:9414:69 | `jsxImportSourcePragmas[jsxImportSourcePragmas.length - 1]` | U-endpoint | decline | The following optional accesses explicitly handle a missing pragma. |
| utilities.ts:9416:59 | `jsxRuntimePragmas[jsxRuntimePragmas.length - 1]` | U-endpoint | decline | The following optional accesses explicitly handle a missing pragma. |
| utilities.ts:9564:48 | `aParts[aParts.length - 2]` | U-endpoint | decline | isNodeModulesOrScopedPackageDirectory explicitly accepts and checks undefined. |
| utilities.ts:9565:48 | `bParts[bParts.length - 2]` | U-endpoint | decline | isNodeModulesOrScopedPackageDirectory explicitly accepts and checks undefined. |
| utilities.ts:9566:30 | `aParts[aParts.length - 1]` | U-endpoint | assert | Both path-component lists have length >= 2 before their last components are canonicalized. |
| utilities.ts:9566:82 | `bParts[bParts.length - 1]` | U-endpoint | assert | Both path-component lists have length >= 2 before their last components are canonicalized. |
| utilities.ts:9711:54 | `components[0]` | U-endpoint | assert | getNormalizedPathComponents always returns a populated root component, even for relative paths. |
| utilities.ts:9859:17 | `results[0]` | U-grid | assert | No includes allocates bucket 0; otherwise results mirrors includeFileRegexes and findIndex selects its bucket. |
| utilities.ts:9864:21 | `results[includeIndex]` | U-grid | assert | No includes allocates bucket 0; otherwise results mirrors includeFileRegexes and findIndex selects its bucket. |
| utilities.ts:10364:24 | `arr[0]` | U-endpoint | assert | Debug.assert(arr.length !== 0) guards the first value; the loop traverses remaining populated values. |
| utilities.ts:10367:32 | `arr[i]` | U-loop | assert | Debug.assert(arr.length !== 0) guards the first value; the loop traverses remaining populated values. |
| utilities.ts:10502:9 | `segments[segment]` | U-typed | assert | Allocation rounds bitsNeeded up to 16-bit segments; bit offsets and nonzero spill stay inside that capacity. |
| utilities.ts:10504:23 | `segments[segment + 1]` | U-typed | assert | Allocation rounds bitsNeeded up to 16-bit segments; bit offsets and nonzero spill stay inside that capacity. |
| utilities.ts:10514:46 | `segments[segment]` | U-typed | assert | Allocation rounds bitsNeeded up to 16-bit segments; bit offsets and nonzero spill stay inside that capacity. |
| utilities.ts:10631:19 | `array[0]` | U-endpoint | assert | array.length < 2 returns before the first read; remaining reads traverse the populated caller array. |
| utilities.ts:10633:24 | `array[i]` | U-loop | assert | array.length < 2 returns before the first read; remaining reads traverse the populated caller array. |
| utilities.ts:12183:22 | `node.arguments[0]` | U-endpoint | decline | An existing conditional-type cast already covers this read. |
| utilities.ts:12185:128 | `node.arguments[0]` | U-endpoint | assert | node.arguments.length >= 1 guards the parser-built first import argument. |
| utilities.ts:12186:22 | `node.arguments[0]` | U-endpoint | decline | An existing conditional-type cast already covers this read. |
| utilitiesPublic.ts:431:20 | `spans[i]` | U-loop | assert | i and j traverse the populated sorted spans, with j < spans.length checked at every read. |
| utilitiesPublic.ts:433:73 | `spans[j]` | U-loop | assert | i and j traverse the populated sorted spans, with j < spans.length checked at every read. |
| utilitiesPublic.ts:434:48 | `spans[j]` | U-loop | assert | i and j traverse the populated sorted spans, with j < spans.length checked at every read. |
| utilitiesPublic.ts:435:65 | `spans[j]` | U-loop | assert | i and j traverse the populated sorted spans, with j < spans.length checked at every read. |
| utilitiesPublic.ts:493:16 | `changes[0]` | U-endpoint | assert | Empty input returns first; the single and multiple branches require changes[0], then traverse populated changes. |
| utilitiesPublic.ts:498:21 | `changes[0]` | U-endpoint | assert | Empty input returns first; the single and multiple branches require changes[0], then traverse populated changes. |
| utilitiesPublic.ts:505:28 | `changes[i]` | U-loop | assert | Empty input returns first; the single and multiple branches require changes[0], then traverse populated changes. |
| utilitiesPublic.ts:708:22 | `matchResult[1]` | U-regex | assert | A successful anchored locale regexp makes capture 1 mandatory; capture 2 remains optional. |
| utilitiesPublic.ts:709:23 | `matchResult[2]` | U-regex | decline | Capture 2 is optional and the locale-loading branch handles its absence. |
| utilitiesPublic.ts:894:45 | `hostNode.declarationList.declarations[0]` | U-endpoint | decline | The explicit declaration truthiness guard handles absence before its repeated read. |
| utilitiesPublic.ts:1047:25 | `paramTags[i]` | U-parallel | assert | The parameter position is nonnegative and the corresponding filtered tag count is checked before the read. |
| utilitiesPublic.ts:1263:45 | `comments[0]` | U-endpoint | decline | The equality assertion compares possibly absent comments and needs no present value. |
| utilitiesPublic.ts:1263:61 | `comments[1]` | U-endpoint | decline | The equality assertion compares possibly absent comments and needs no present value. |
| utilitiesPublic.ts:1353:53 | `node.parent.typeParameters[0]` | U-endpoint | decline | An equality comparison with node legitimately permits an absent element. |
| utilitiesPublic.ts:2653:48 | `(parseTreeNode.parent as SignatureDeclaration).parameters[paramIdx - 1]` | U-position | decline | The previous-parameter branch explicitly handles an absent predecessor. |

### Second wave

| Site | Read | Class | Action | Invariant or decline reason |
| --- | --- | --- | --- | --- |
| scanner.ts:360:16 | `map[0]` | U-endpoint | assert | All four Unicode range tables are nonempty populated lists of start/end pairs. |
| scanner.ts:373:13 | `map[mid]` | U-position | assert | The range-map binary search rounds mid down to an even position; lo + 1 < hi keeps both pair cells in bounds. |
| scanner.ts:373:41 | `map[mid + 1]` | U-position | assert | The range-map binary search rounds mid down to an even position; lo + 1 < hi keeps both pair cells in bounds. |
| scanner.ts:377:20 | `map[mid]` | U-position | assert | The range-map binary search selects a populated start cell at its even mid position. |
| scanner.ts:415:12 | `tokenStrings[t]` | U-table | decline | tokenToString explicitly returns undefined for non-textual token kinds; the reverse map is intentionally sparse. |
| scanner.ts:427:12 | `regExpFlagCharCodes[f]` | U-table | decline | The reverse regexp-flag map exposes undefined for flags without a character code. |
| scanner.ts:486:17 | `lineStarts[line]` | U-position | assert | The computed nonempty line map and line clamp/failure check leave a valid populated line entry. |
| scanner.ts:491:22 | `lineStarts[line + 1]` | U-position | decline | The next start is legitimately absent on the last line; the comparison is false and the debugText/return branch handles that case. |
| scanner.ts:491:45 | `lineStarts[line + 1]` | U-position | decline | The next start is legitimately absent on the last line; the comparison is false and the debugText/return branch handles that case. |
| scanner.ts:494:28 | `lineStarts[line + 1]` | U-position | assert | The explicit line < lineStarts.length - 1 guard establishes the populated next-line entry. |
| scanner.ts:512:31 | `lineStarts[lineNumber]` | U-position | assert | computeLineOfPosition returns the valid populated predecessor entry in the nonempty computed line map. |
| scanner.ts:1334:29 | `text[identifierStart]` | U-position | decline | The literal comparison accepts an absent character without dereferencing or requiring a value. |
| scanner.ts:1858:22 | `text[pos]` | U-position | assert | The preceding digit test on charCodeUnchecked(pos) succeeds only at a present character in the source string. |
| scanner.ts:2325:38 | `text[pos + 1]` | U-position | decline | The shebang lookahead comparison accepts absent text at EOF. |
| scanner.ts:3644:17 | `match[1]` | U-regex | decline | The switch and final undefined return handle an unrecognized or absent directive capture; no required-value use occurs. |
| factory/nodeFactory.ts:6897:31 | `source[statementOffset]` | U-loop | assert | statementOffset is nonnegative and below the populated caller statement-list length before its predicate. |
| factory/nodeFactory.ts:6927:31 | `source[statementOffset]` | U-loop | assert | The supplied prologue offset is defined, nonnegative, and below the populated caller statement-list length. |
| factory/nodeFactory.ts:6966:41 | `array[i]` | U-loop | assert | start is a nonnegative prologue partition offset and i < array.length bounds each populated input element. |
| factory/nodeFactory.ts:7047:42 | `statements[i]` | U-loop | decline | The prologue read already has an explicit PrologueDirective cast; no additional required-read annotation is needed. |
| factory/nodeFactory.ts:7051:43 | `declarations[i]` | U-loop | decline | The prologue read already has an explicit PrologueDirective cast; no additional required-read annotation is needed. |
| factory/nodeFactory.ts:7539:9 | `destRanges[key]` | U-table | decline | The source-map range payload explicitly permits undefined, and the other indexed expression is its unchanged pure destination store. |
| factory/nodeFactory.ts:7539:27 | `sourceRanges[key]` | U-table | decline | The source-map range payload explicitly permits undefined, and the other indexed expression is its unchanged pure destination store. |
| factory/emitNode.ts:148:12 | `node.emitNode?.tokenSourceMapRanges?.[token]` | U-table | decline | The existing optional-chain lookup and optional return type deliberately handle a missing token range. |
| factory/emitNode.ts:157:5 | `tokenSourceMapRanges[token]` | U-table | decline | This is a pure store of an explicitly optional source-map range, not a required read. |
| factory/emitNode.ts:297:24 | `sourceEmitHelpers[i]` | U-loop | assert | The bounded loop traverses the populated helper list; compaction writes only earlier indices. |
| factory/emitHelpers.ts:466:78 | `elements[i]` | U-loop | assert | The bounded loop visits populated binding elements before the final rest element. |
| factory/emitHelpers.ts:470:34 | `computedTempVariables[computedTempVariableOffset]` | U-parallel | assert | computedTempVariables is asserted defined and contains one populated entry for each computed binding property in traversal order. |
| factory/emitHelpers.ts:720:23 | `input[i]` | U-parallel | assert | A real tagged-template call supplies args.length + 1 populated string segments, so the shared args index selects a segment. |
| factory/emitHelpers.ts:721:34 | `args[i]` | U-loop | assert | The bounded loop selects a populated unique-name argument supplied by the tagged-template call. |
| factory/emitHelpers.ts:723:19 | `input[input.length - 1]` | U-endpoint | assert | The tagged-template string array has one final populated segment beyond all substitutions. |
| factory/utilities.ts:265:32 | `children[0]` | U-endpoint | assert | The children.length > 0 and length <= 1 branch selects the sole populated transformed JSX child. |
| factory/utilities.ts:292:32 | `children[0]` | U-endpoint | assert | The children.length > 0 and length <= 1 branch selects the sole populated transformed JSX child. |
| factory/utilities.ts:1281:48 | `userStateStack[stackIndex - 1]` | U-position | decline | onEnter accepts an optional previous state; the root frame explicitly supplies undefined. |
| factory/utilities.ts:1282:27 | `stateStack[stackIndex]` | U-position | decline | Debug.assertEqual already checks the slot against the known enter function; its operand API permits absence. |
| factory/utilities.ts:1283:54 | `nodeStack[stackIndex]` | U-position | assert | The initial frame and pushStack install the node before enter runs; the active frame index selects that populated node. |
| factory/utilities.ts:1295:27 | `stateStack[stackIndex]` | U-position | decline | Debug.assertEqual already checks the slot against the known left function; its operand API permits absence. |
| factory/utilities.ts:1298:41 | `nodeStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before left; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1298:69 | `userStateStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before left; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1298:97 | `nodeStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before left; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1313:27 | `stateStack[stackIndex]` | U-position | decline | Debug.assertEqual already checks the slot against the known operator function; its operand API permits absence. |
| factory/utilities.ts:1316:28 | `nodeStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before operator; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1316:65 | `userStateStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before operator; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1316:93 | `nodeStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before operator; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1327:27 | `stateStack[stackIndex]` | U-position | decline | Debug.assertEqual already checks the slot against the known right function; its operand API permits absence. |
| factory/utilities.ts:1330:42 | `nodeStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before right; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1330:71 | `userStateStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before right; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1330:99 | `nodeStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before right; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1345:27 | `stateStack[stackIndex]` | U-position | decline | Debug.assertEqual already checks the slot against the known exit function; its operand API permits absence. |
| factory/utilities.ts:1347:39 | `nodeStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before exit; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1347:62 | `userStateStack[stackIndex]` | U-parallel | assert | enter installed the active frame user state before exit; node and user-state stacks share the same live frame index. |
| factory/utilities.ts:1351:30 | `stateStack[stackIndex]` | U-position | decline | The state-function equality comparison does not require a present operand; leave the original side-selection read. |
| factory/utilities.ts:1352:64 | `userStateStack[stackIndex]` | U-position | assert | After popping a child, the parent index selects its already-entered populated user state for foldState. |
| factory/utilities.ts:1366:27 | `stateStack[stackIndex]` | U-position | decline | Debug.assertEqual already checks the slot against the known done function; its operand API permits absence. |
| factory/utilities.ts:1403:30 | `nodeStack[stackIndex]` | U-position | decline | The circularity assertion compares values and does not require the indexed operand to be present. |
| factory/utilities.ts:1481:16 | `stateStack[stackIndex]` | U-position | decline | The done-function comparison does not require a present operand; the separate subsequent callee read is asserted at its own point. |
| factory/utilities.ts:1482:26 | `stateStack[stackIndex]` | U-position | assert | Initial allocation and pushStack install every active state function; each transition preserves a valid live frame index. |
| parser.ts:1290:28 | `current[i]` | U-loop | assert | The parser-built NodeArray is populated; the reverse loop checks i >= 0 before enqueueing its child. |
| parser.ts:1616:35 | `result.statements[0]` | U-endpoint | decline | The existing optional chain handles a missing first JSON expression statement. |
| parser.ts:1876:35 | `sourceFile.statements[pos]` | U-position | assert | pos starts at zero or is a successful statement search result; the active await range has a populated statement at pos. |
| parser.ts:1877:35 | `sourceFile.statements[start]` | U-position | assert | start is a successful findNextStatementWithAwait result, not its -1 sentinel. |
| parser.ts:1904:51 | `sourceFile.statements[pos]` | U-position | assert | pos >= 0 excludes the failed-search sentinel; successful search selects a populated boundary statement. |
| parser.ts:1925:35 | `sourceFile.statements[pos]` | U-position | assert | The final pos >= 0 branch selects the populated boundary statement returned by the preceding search. |
| parser.ts:1945:51 | `statements[i]` | U-loop | assert | The bounded loop traverses the parser-built populated statement list. |
| parser.ts:1954:52 | `statements[i]` | U-loop | assert | The bounded loop traverses the parser-built populated statement list. |
| parser.ts:6056:53 | `children[children.length - 1]` | U-endpoint | decline | An empty JSX children list legitimately has no last child; the following optional kind test handles it. |
| parser.ts:9025:44 | `comments[0]` | U-endpoint | decline | The newline equality comparisons tolerate absence and require no dereference; the nonempty guard is retained. |
| parser.ts:9025:68 | `comments[0]` | U-endpoint | decline | The newline equality comparisons tolerate absence and require no dereference; the nonempty guard is retained. |
| parser.ts:9032:37 | `comments[comments.length - 1]` | U-endpoint | assert | comments.length is nonzero and pushComment built a populated string list before its last comment is trimmed. |
| parser.ts:9036:47 | `comments[comments.length - 1]` | U-endpoint | assert | The last comment remains populated across the local trimEnd call and before any pop or replacement. |
| parser.ts:10454:23 | `currentArray[currentArrayIndex]` | U-position | decline | The syntax cursor represents misses with undefined and handles them through its current-node tests and rescan. |
| parser.ts:10469:35 | `currentArray[currentArrayIndex]` | U-position | decline | The syntax cursor represents misses with undefined and handles them through its current-node tests and rescan. |
| parser.ts:10523:39 | `array[i]` | U-loop | decline | visitArray explicitly checks child before using it, preserving absence handling in the syntax cursor. |
| parser.ts:10636:25 | `arg.arguments["no-default-lib"]` | U-table | decline | The pragma optional-key comparison accepts a missing no-default-lib argument. |
| parser.ts:10717:22 | `tripleSlash[1]` | U-regex | assert | A successful tripleSlashXMLCommentStartRegEx match requires its non-whitespace capture 1. |
| parser.ts:10731:35 | `matchResult[2]` | U-regex | decline | Captures 2 and 3 are alternative quote branches; either capture can be absent and the existing OR chooses the branch. |
| parser.ts:10731:53 | `matchResult[3]` | U-regex | decline | Captures 2 and 3 are alternative quote branches; either capture can be absent and the existing OR chooses the branch. |
| parser.ts:10733:74 | `matchResult[1]` | U-regex | assert | getNamedArgRegEx always captures the whitespace/name/equal prefix in group 1; the quote branches concern other groups. |
| parser.ts:10769:18 | `match[1]` | U-regex | assert | Both caller regexps require a nonempty pragma-name capture 1 on successful matches. |
| parser.ts:10774:18 | `match[2]` | U-regex | decline | Capture 2 is optional pragma text and getNamedPragmaArguments explicitly handles its absence. |
| parser.ts:10787:26 | `pragma.args[i]` | U-loop | assert | The bounded loop selects a populated argument-definition entry in the pragma metadata. |
| parser.ts:10788:14 | `args[i]` | U-parallel | decline | The missing-argument test explicitly accepts absence for an optional pragma definition. |
| parser.ts:10794:33 | `args[i]` | U-parallel | decline | The nearby argument test permits missing optional arguments and the original indexed dictionary write forwards that value. |

## Third wave: debug, path, and slice coverage

Program partition: adaptation 30 owns **debug.ts and path.ts**, moved from 32
by the user's instruction. No other partition's files were changed. The new
47 ledger entries comprise 37 assertions and nine declines in these two files,
plus one newly discovered decline in utilities.ts's unknown-returning option
query. The original 273 ledger entries remain identical.

Graph connector rows and cells are allocated and zero-filled, so their reads
are U-grid required values, including the reads in compound assignments.
Grid rows are allocated but their graph-node cells are deliberately sparse;
assert the row at its existing evaluation point and decline the optional cell.
Path component arrays carry a populated root string, including the empty
relative root; component skipping and root truthiness tests stay unchanged.

Further audit rule: inspect top-level union TypeFlags, not the printed type
text. Undefined can be hidden by an alias, or appear inside a callback type
without making the callback optional. This found one omitted optional lookup,
utilities.ts:9402, which now has an explicit decline.

### Census and proof

The unchanged Adamic loader checked 78 compiler roots on the tree with 00, 10
and 30 only; 20 remains excluded. `census.sh <tree> <out> --wave3` scopes these
two files, retaining every requested code and full diagnostic chain.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| debug.ts before | 4 | 12 | 18 | 3 | 0 | 37 |
| debug.ts after | 0 | 0 | 0 | 1 | 0 | 1 |
| path.ts before | 5 | 1 | 0 | 1 | 0 | 7 |
| path.ts after | 0 | 0 | 0 | 0 | 0 | 0 |
| **Total before** | **9** | **13** | **18** | **4** | **0** | **44** |
| **Total after** | **0** | **0** | **0** | **1** | **0** | **1** |

The remaining TS2322 at debug.ts:171 is the deliberate write of undefined to
the assertion cache, not a required indexed read. It needs an honest optional
value declaration and is outside this adaptation.

Default `stage3/oracle/run.sh` passed **106,367 tests, 0 failing, 0 pending**,
with all runners, no test filter, four workers, and light=false. The baseline
diff is **0 bytes**. Install/build/tests exited 0 and took 4.889/10.429/231.925s;
total 247.304s. `wave3-oracle-report.json` records the exact commands. No baseline
was accepted or changed. Native tsc, lint, and browser integrations were not run.

`verify.cjs` passed stock emitted-JavaScript byte comparison, site contracts,
and idempotence for all 17 files. The actual second adapter CLI run made zero
edits, and fresh `apply.sh` discovery produced identical hashes in all 17
files. `wave3-proof.json` retains these hashes and counts. Toolchain setup:
Go 0s, clang 0s, Node 0s, submodules 0s, build cache 30s, total 30s; nproc 5,
CPU quota 4. Test and verification output went to `/tmp/stage3-indexed-wave3-*`
logs. Full before/after/mutant diagnostic chains are checked in as wave3 JSON.

The new required-read mutant changes debug.ts:1059:48 from
`columns[node.level]!` to `(columns[node.level] ?? 0)`. The census **misses it**:
still one finding, with identical diagnostics. Stock emitted JavaScript changes
and the site contract reports `required read defaulted: debug.ts:1059:48`; both
exit 1. The default oracle was not rerun on this mutant. A separate scanner
mutant changes `map[0]!` to `(map[0] ?? 0)` and regenerates its slice manifest;
the new slice audit rejects it with `required assertion missing:
src/compiler/scanner.ts:360`, exit 1. Neither mutant is retained.

### Slice audit and stopping point

Fetched scanner-proof at `21cad99e86bef769d29d86710ee71a9d5600726a`. Its
`stage3/slice/slice.cjs` emitted both requested entries immediately, so no
20-minute retry was needed to obtain the slices. The createScanner entry plus
ScriptTarget/SyntaxKind reaches eight files and 91 declarations. The
createSourceFile entry reaches 78 files and 5,121 declarations through namespace
references; this is the observed output of the published tool, not a claim that
the parser has a compact dependency closure.

`slice-audit.cjs <full-tree> <slice.json> <report.json>` checks each source span's
SHA-256, identifies reached indexed expressions at their original positions,
and joins them to the expression/occurrence ledger. It distinguishes pure stores
and top-level undefined unions. A present type alone does not prove runtime
density; an unowned optional expression is a review candidate, not automatically
a required read. The whole-tree site contract independently validates declines
and U-zero shapes. Both complete audit reports are checked in.

Scanner: **29 indexed expressions, zero unresolved candidates**. All are in
core.ts, scanner.ts, or utilities.ts, which 30 covers. Its remaining files
(commandLineParser.ts's reached target declaration, corePublic.ts's interface,
the generated diagnostics, and types.ts) have no reached indexed expressions.
Debug's reached assertions have none either.

At the third-wave checkpoint, the parser audit found **1,715 indexed expressions,
zero unreviewed candidates within 30**, and **988 possibly-undefined candidates
outside 30**. Ownership had not yet been supplied, so external coverage was
unverified. The user subsequently supplied the ownership map above and asked
to hold the parser audit until namespace-member slicing lands. Its historical
audit artifacts remain unchanged; no parser audit was rerun in this follow-up.

### Third-wave site ledger

| Site | Read | Class | Action | Invariant or decline reason |
| --- | --- | --- | --- | --- |
| debug.ts:168:36 | `assertionCache[key]` | U-table | decline | Cleared cache slots are explicitly undefined; the following cachedFunc !== undefined guard handles them. |
| debug.ts:378:28 | `match[1]` | U-regex | assert | A successful literal function-name regexp requires its nonempty capture 1. |
| debug.ts:392:42 | `members[0]` | U-endpoint | assert | getEnumMembers pushes populated number/name tuples; the nonzero length branch selects its populated first tuple. |
| debug.ts:392:64 | `members[0]` | U-endpoint | assert | getEnumMembers pushes populated number/name tuples; the nonzero length branch selects its populated first tuple. |
| debug.ts:996:29 | `links[id]` | U-table | decline | The graph cache legitimately misses; the following !graphNode branch allocates and records its node. |
| debug.ts:1059:48 | `columns[node.level]` | U-position | assert | fill initializes every width to zero; computed node levels are nonnegative and below the computed graph height. |
| debug.ts:1071:35 | `children[i]` | U-loop | assert | getChildren pushes populated graph nodes and the loop bounds their dense list. |
| debug.ts:1110:36 | `switchStatement.caseBlock.clauses[i]` | U-loop | assert | The flow switch range comes from populated parser clauses and bounds the active clause interval. |
| debug.ts:1137:17 | `grid[node.level][node.lane]` | U-grid | decline | This is an unchanged pure cell store; only its required allocated row read receives an assertion. |
| debug.ts:1137:17 | `grid[node.level]` | U-grid | assert | The column-width map allocates every grid row; node.level selects a row within graph height. |
| debug.ts:1140:35 | `children[i]` | U-loop | assert | getChildren pushes populated graph nodes and the loop bounds their dense list. |
| debug.ts:1145:21 | `connectors[node.level][child.lane]` | U-grid | assert | Every connector row and cell is filled with zero; node level and child lane are within the computed graph dimensions. |
| debug.ts:1145:21 | `connectors[node.level]` | U-grid | assert | Every connector row and cell is filled with zero; node level and child lane are within the computed graph dimensions. |
| debug.ts:1148:21 | `connectors[node.level][node.lane]` | U-grid | assert | Every connector row and cell is filled with zero; the node level and lane are within graph dimensions. |
| debug.ts:1148:21 | `connectors[node.level]` | U-grid | assert | Every connector row and cell is filled with zero; the node level and lane are within graph dimensions. |
| debug.ts:1152:36 | `parents[i]` | U-loop | assert | getParents pushes populated graph nodes and the loop bounds their dense list. |
| debug.ts:1156:21 | `connectors[node.level - 1][parent.lane]` | U-grid | assert | A node with parents has positive level; its preceding connector row and each parent lane are allocated and zero-filled. |
| debug.ts:1156:21 | `connectors[node.level - 1]` | U-grid | assert | A node with parents has positive level; its preceding connector row and each parent lane are allocated and zero-filled. |
| debug.ts:1163:47 | `connectors[column - 1][lane]` | U-grid | assert | column > 0 selects an allocated preceding connector row; the bounded lane selects its zero-filled cell. |
| debug.ts:1163:47 | `connectors[column - 1]` | U-grid | assert | column > 0 selects an allocated preceding connector row; the bounded lane selects its zero-filled cell. |
| debug.ts:1164:46 | `connectors[column][lane - 1]` | U-grid | assert | The bounded column selects an allocated connector row; lane > 0 selects its zero-filled predecessor cell. |
| debug.ts:1164:46 | `connectors[column]` | U-grid | assert | The bounded column selects an allocated connector row; lane > 0 selects its zero-filled predecessor cell. |
| debug.ts:1165:37 | `connectors[column][lane]` | U-grid | assert | Both nested loop bounds select an allocated connector row and zero-filled cell. |
| debug.ts:1165:37 | `connectors[column]` | U-grid | assert | Both nested loop bounds select an allocated connector row and zero-filled cell. |
| debug.ts:1169:25 | `connectors[column]` | U-grid | assert | The nested loop bounds select an allocated connector row for the unchanged cell store. |
| debug.ts:1176:39 | `connectors[column][lane]` | U-grid | assert | The bounded column and lane select an allocated connector row and populated numeric cell. |
| debug.ts:1176:39 | `connectors[column]` | U-grid | assert | The bounded column and lane select an allocated connector row and populated numeric cell. |
| debug.ts:1178:34 | `grid[column][lane]` | U-grid | decline | Grid cells are intentionally sparse; the following node guard or absence test handles missing cells. |
| debug.ts:1178:34 | `grid[column]` | U-grid | assert | The bounded column (or guarded next column) selects an allocated grid row, while its cell remains legitimately optional. |
| debug.ts:1181:58 | `columnWidths[column]` | U-parallel | assert | The bounded grid column selects the corresponding initialized column-width entry. |
| debug.ts:1188:58 | `columnWidths[column]` | U-parallel | assert | The bounded grid column selects the corresponding initialized column-width entry. |
| debug.ts:1192:98 | `grid[column + 1][lane]` | U-grid | decline | Grid cells are intentionally sparse; the following node guard or absence test handles missing cells. |
| debug.ts:1192:98 | `grid[column + 1]` | U-grid | assert | The bounded column (or guarded next column) selects an allocated grid row, while its cell remains legitimately optional. |
| debug.ts:1199:17 | `lanes[lane]` | U-position | assert | lanes is filled with empty strings and writeLane receives a lane within the render loop bounds. |
| path.ts:515:18 | `pathComponents[0]` | U-endpoint | decline | The original truthiness test handles an empty or absent root, and the repeated guarded read already narrows; keep both reads. |
| path.ts:540:22 | `components[0]` | U-endpoint | assert | some excludes an empty list; parsed path components always begin with a populated root string, possibly empty. |
| path.ts:542:27 | `components[i]` | U-loop | decline | The next !component guard explicitly skips empty or absent components. |
| path.ts:547:21 | `reduced[reduced.length - 1]` | U-endpoint | decline | The original equality comparison tolerates absence without requiring a value. |
| path.ts:552:22 | `reduced[0]` | U-endpoint | decline | The root truthiness test intentionally distinguishes empty relative roots; no dereference requires presence. |
| path.ts:912:42 | `aComponents[i]` | U-parallel | assert | Both normalized path component lists are populated and i is below their shared minimum length. |
| path.ts:912:58 | `bComponents[i]` | U-parallel | assert | Both normalized path component lists are populated and i is below their shared minimum length. |
| path.ts:986:31 | `parentComponents[i]` | U-parallel | assert | Parsed component lists are populated and the child-length guard covers every bounded parent index. |
| path.ts:986:52 | `childComponents[i]` | U-parallel | assert | Parsed component lists are populated and the child-length guard covers every bounded parent index. |
| path.ts:1016:52 | `fromComponents[start]` | U-parallel | assert | Both populated parsed component lists bound start before canonicalizing the shared position. |
| path.ts:1017:50 | `toComponents[start]` | U-parallel | assert | Both populated parsed component lists bound start before canonicalizing the shared position. |
| path.ts:1076:28 | `pathComponents[0]` | U-endpoint | assert | getPathComponentsRelativeTo returns a populated root-first list in either branch, including an empty root for relative paths. |
| utilities.ts:9402:9 | `options[option.name]` | U-table | decline | The unknown-returning option query deliberately forwards an absent option; its consumers decide defaults or compare absence. |

## Scanner slice coverage completed with supplied ownership

This follow-up audits the same eight-file, 91-declaration createScanner slice,
including ScriptTarget and SyntaxKind. It checks **29 indexed expressions**:
**11 existing assertions**, **three documented declines**, **six pure stores**,
and **nine uses whose types already support their original operation**.
There are **zero unresolved sites** and no adapter or site-ledger changes.
Presence assertions remain construction/caller obligations; the census is
limited to the five requested codes.

| Slice file | Owner | Indexed expressions |
| --- | --- | ---: |
| commandLineParser.ts | 32 | 0 |
| core.ts | 30 | 14 |
| corePublic.ts | 32 | 0 |
| debug.ts | 30 | 0 |
| diagnosticInformationMap.generated.ts | 32 | 0 |
| scanner.ts | 30 | 11 |
| types.ts | 32 | 0 |
| utilities.ts | 30 | 4 |

`scanner-coverage-owner-sites.json` contains explicit lists for 31, 32, 33, and
unassigned files. **All four lists are empty**: the outside files have no
indexed expressions in their reached slice declarations. The audit now assigns
file ownership using the supplied map and lists every outside indexed expression,
including pure stores and already accepted reads, for its owner. It never edits
an outside file or treats ownership alone as proof of a required read.

The actual gathered slice was checked by the unchanged Adamic loader: **eight
roots, all five requested codes zero**. A scratch before control removes exactly
the slice's 11 adaptation-30 indexed assertions; its census is **10 findings**
(one TS2345 and nine TS2532). Afterward it is **0**. Core contributes 3 -> 0,
scanner 4 -> 0, utilities 3 -> 0; the other five files stay zero. Full diagnostic
chains are in `scanner-coverage-census-{before,after}.json`. Stock TypeScript
emits identical JavaScript before/after for every gathered file.

The rerun mutant replaces scanner.ts:360's `map[0]!` with `(map[0] ?? 0)`, using
a matching regenerated manifest. The audit rejects it with `required assertion
missing: src/compiler/scanner.ts:360`, exit 1. The gathered slice's five-code
census **misses the mutant**, still zero; its JSON is retained separately.

The full 17-file site contract, emitted-JavaScript comparison, and idempotence
checks passed again. All 17 adapted source hashes equal `wave3-proof.json`.
Since the adapted tree is unchanged, the prior default oracle result applies:
106,367 passing and a zero-byte baseline diff. The full oracle was not rerun
for this audit-tool/report-only change. The parser audit remains on hold.

Commands, each redirected to a scratch log:

```sh
CENSUS_TYPESCRIPT=/path/to/typescript node stage3/adapt/30-indexed-reads/slice-audit.cjs <full-tree> <scanner-slice>/slice.json <audit.json>
bash stage3/adapt/30-indexed-reads/census.sh <scanner-slice> <census-out> --scanner-slice
CENSUS_TYPESCRIPT=/path/to/typescript node stage3/adapt/30-indexed-reads/adapt.cjs --check <full-tree>
CENSUS_TYPESCRIPT=/path/to/typescript node stage3/adapt/30-indexed-reads/verify.cjs <before-tree> <full-tree>
```

Evidence: `scanner-coverage-audit.json`, `scanner-coverage-proof.json`, census
JSON, and the owner-site lists. Delivery stays on codex/stage3-indexed-reads;
main is reserved for @system_adamic_integration. No force-push or rebase of a
pushed branch is allowed; bring new main work in with a merge of origin/main.

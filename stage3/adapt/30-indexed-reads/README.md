# Required indexed reads, first partition

This adapter changes only `src/compiler/core.ts`, `utilities.ts`, and
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
validates occurrence counts, plans all three files before writing, and fails
on unexpected site shapes. Later partitions must revise their own reviewed
ledger rather than blanket-asserting every indexed expression.

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

## Census

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

## Proof and mutants

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

## Site ledger

Each row identifies one read in the pinned source, its class, action, and the
one-line invariant or decline reason. Repeated expressions on the same line
have separate columns and separate rows. Pure stores and existing casts are
listed when encountered in the possibly-undefined inventory, but not adapted.

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

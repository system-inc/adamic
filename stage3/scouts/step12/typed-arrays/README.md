# Step 12: scanner Uint16Array

Twelve witnesses, six native matches, six stops. No silent miscompile was observed.
All twelve source mutants fail the original Node stdout expectation; six also
fail it natively. Every compiled baseline and mutant matches its own Node source.
Finished baseline programs pass ASan/UBSan with LeakSanitizer and balanced counts.
Exact stdout, stderr, exit codes and mutation replacements are in `results.json`.

## Inventory

Stock TypeScript **6.0.3 compiler API**, not cohere's rule, inventories the
scanner's retained declaration closure. The selection is the 89 code declarations
in `stage3/slice/evidence/member-scanner-manifest.json`. A fresh gathering reproduced
89 code declarations in eight files, with 78 modules in its evaluation graph.
`inventory.cjs` asks the checker for typed-array receiver types, including unions,
and visits constructors, indexes and properties. It conservatively visits entire
selected namespaces. Inventory is over source before slice adaptation 52, so the
locations below refer to upstream source, rather than gathered line numbers.

The eight code files are `commandLineParser.ts`, `core.ts`, `corePublic.ts`,
`debug.ts`, `diagnosticInformationMap.generated.ts`, `scanner.ts`, `types.ts`, and
`utilities.ts`, all under `src/compiler/`. Every typed-array operation is in
`parsePseudoBigInt`, which the scanner calls to normalize bigint token values.
There are **six syntactic sites: one constructor, three reads, three writes,
and one length access**. A compound store counts as both a read and a write.
There are **zero subarray calls** and no other typed-array kinds in this closure.
Names of typed-array constructors in unrelated library-feature tables are not
runtime typed-array uses and are outside the retained declarations.

| Upstream location | Expression | Operations | Witnesses |
|---|---|---|---|
| utilities.ts:10491:22 | `new Uint16Array((bitsNeeded >>> 4) + (bitsNeeded & 15 ? 1 : 0))` | Constructor | 01, 11 |
| utilities.ts:10502:9 | `segments[segment] \|= shiftedDigit` | Read, write | 02, 07, 12 |
| utilities.ts:10504:23 | `segments[segment + 1] \|= residual` | Read, write | 03, 07, 12 |
| utilities.ts:10508:31 | `segments.length` | Length | 04 |
| utilities.ts:10514:46 | `mod10 << 16 \| segments[segment]` | Read | 05, 08 |
| utilities.ts:10516:13 | `segments[segment] = segmentValue` | Write | 06 |

`evidence/inventory.json` preserves the machine-readable inventory. The user's
`utilities.ts:65:22` stop is the gathered constructor: the fresh slice also places
it at 65:22. That is upstream 10491:22.

## Compiler combination and reproduction

Main is `45487a809f89885a3fc651cd590e7dabf31362dc`; the runtime branch tip is
`04f18a2a78ab8f7576e2c92deefa5a924099dad1`. TypeScript is tag `v6.0.3`, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. The scratch compiler merges the
**whole runtime branch**, including its earlier typed-array support and runtime
ancestors, into that main. It is not just a cherry-pick of the Uint16 extension.
The scratch staged tree is `df5fbf019596ac43fd1211b5116e25d152400b78`.
Neither scratch checkout is committed or pushed.

Four merge conflicts required resolution. `cast.go` keeps main's current cast
proof. `expression.go` keeps main's enum dispatch and also invokes the incoming
typed-array dispatch; its other typed-array representation checks merge normally.
The conflicting shared oracle registry and counts table keep main's versions:
this scout does not run that registry or claim integrated package validation.
`prepare.sh` reproduces these resolutions and checks the conflict set. It reuses
setup's initialized submodules via a scratch symlink; `-buildvcs=false` avoids
VCS stamping failures on that mount. No production compiler file changed here.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step12-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/scouts/step12/typed-arrays/prepare.sh NEW_SCRATCH > /tmp/step12-prepare.log 2>&1
python3 stage3/scouts/step12/typed-arrays/run.py \
  --compiler NEW_SCRATCH/adamic --output NEW_RESULTS --check > /tmp/step12-run.log 2>&1
```

The prepare script needs both pinned commits already fetched. It deliberately
refuses an existing scratch path. `run.py` executes the source with
`node --disable-warning=ExperimentalWarning oracle/node.mjs FILE`, compiles and
runs every baseline and source mutant, then runs compiled baselines with
`--count` and `--sanitize`. Sanitized runs set `ASAN_OPTIONS=detect_leaks=1` and
`UBSAN_OPTIONS=halt_on_error=1`. `--check` compares current observations with the
recorded baseline diagnostics, outputs, counts and sanitizer results. Every
subprocess writes stdout/stderr directly to files. Local counts are refreshed in
`counts.md`; these scouts are not registered shared oracle fixtures.

To rebuild the inventory, use stock `typescript@6.0.3` and a pinned source tree
with its generated diagnostics. The measured source had adaptations 00, 10 and
50; these do not change the typed-array statements or their line numbers.

```sh
SLICE_TYPESCRIPT=/path/to/typescript/lib/typescript.js \
  node stage3/scouts/step12/typed-arrays/inventory.cjs SOURCE_TREE \
  stage3/slice/evidence/member-scanner-manifest.json > inventory.json 2> inventory.log
```

The real slice was gathered with the existing `stage3/slice/run.sh`, entries
`scanner.ts:createScanner`, `types.ts:ScriptTarget`, and `types.ts:SyntaxKind`.
Its existing scanner adaptation profile applied 52–57, 59, 81–82 and 85.
The byte/import audit passed: **168 source spans and 78 ordered import lists**.
No new production source adaptation was made.

## Witness results and exact stops

The numeric and optional numeric driver values use template strings because
Adamic's `console.log` declaration takes one string. The statement reductions
retain the typed-array operations. Witness 09 retains the complete gathered
`parsePseudoBigInt` body, including existing adaptation 52's checked reads.
Its dependencies are reduced to the numeric CharacterCodes properties and a
module exporting a throwing `never` helper; it does not reproduce Debug's
unrelated stack-trace machinery or namespace. This dependency remapping also
applies to 07 and 08. The helper does not change successful Node results.

| Witness | Result | Meaning |
|---|---|---|
| 01_constructor | Refused | Original partial-word capacity formula uses numeric truthiness |
| 02_or_store | Checker | Original compound element read is possibly undefined |
| 03_residual_store | Checker | Original next-word compound read is possibly undefined |
| 04_length | Matches | Highest index is length minus one; writes preserve fixed length |
| 05_read | Matches | Original bitwise combination yields 655359 |
| 06_write | Matches | Division quotient stores 65535; supplemental 65536 store wraps to 0 |
| 07_checked_or_store | NotYet | Existing checked-read form hits the never-call lowering stop |
| 08_checked_read | NotYet | Existing checked-read form hits the same never-call stop |
| 09_parse_pseudo_bigint | Refused | Complete gathered conversion stops at comma update |
| 10_subarray_control | Matches | Supplemental runtime view aliases parent in both directions |
| 11_constructor_boolean_control | Matches | Explicit boolean condition isolates allocation and zero fill |
| 12_dense_or_control | Matches | Dense read control isolates Uint16 truncation and octal carry |

These are the complete diagnostic messages after the location prefix; exact
prefixes and stderr bytes are retained in `results.json`:

- 01, line 5:60: `Adamic 0.1 refuses a number as a condition; compare it explicitly, like name.length > 0 or count !== 0`
- 02, line 6:1: `error TS2532: Object is possibly 'undefined'.`
- 03, line 8:15: `error TS2532: Object is possibly 'undefined'.`
- 07, line 9:43: `stage 0 can't lower a void call used as a value yet`
- 08, line 9:56: `stage 0 can't lower a void call used as a value yet`
- 09, line 41:67: `Adamic 0.1 refuses the comma operator; write each expression as its own statement`

All stopped builds exit 1. Main alone stops on the type-correct constructor
control with `stage 0 can't lower new an Identifier yet` at 11, line 5:22
(`evidence/main-witness.txt`). Adding the runtime branch clears that refusal.
The exact main-only diagnostic differs from the user's supplied scanner stop;
we have not substituted the supplied text for this observation.

Controls 10–12 are explicitly additional measurements. There is no tsc subarray
span for 10. Control 11 changes the numeric condition to `(bitsNeeded & 15) !== 0`.
In 07 the residual condition is explicit too, to isolate the checked read.
Control 12 uses `?? 0` only for its known dense, zero-initialized elements. It is
**not** a proposal to replace production failure checks with zero.

The whole fresh slice also has an earlier integration stop on this main:
`corePublic.ts:9:5: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added`
(`evidence/scanner-build.txt`). Previous scouts' discovery adaptations are not
included in this fresh rebuild. This measurement does not claim a native scanner
run or that Uint16Array is the only remaining scanner blocker.

## Mutants actually run

These are source mutants, one independently generated per witness. Each runs on
Node and is compared with the original source's Node stdout. Each compiled mutant
also matches its own Node source and differs from the original expectation.
They establish sensitivity of these outputs; they are not runtime/compiler
mutations and do not prove uncompiled native paths.

| Witness | Mutant | Observed catch |
|---|---|---|
| 01 | Drop partial-word rounding | Node: 4 bits gets length 0 instead of 1 |
| 02 | OR zero instead of shifted digit | Node: 0 instead of 61440 |
| 03 | Put carry in current word | Node: `3,0` instead of `0,3` |
| 04 | Omit length minus one | Node and native: first index 3 instead of 2 |
| 05 | Shift remainder by 15 | Node and native: 327679 instead of 655359 |
| 06 | Increment quotient before store | Node and native: wraps to 0 instead of 65535 |
| 07 | OR zero instead of shifted digit | Node: first word 0 instead of 32768 |
| 08 | Shift remainder by 15 | Node: 327679 instead of 655359 |
| 09 | Drop residual carry | Node: `0o777777n` gives 65535 instead of 262143 |
| 10 | Write 65 rather than 66 through view | Node and native: parent/view expose 65 |
| 11 | Drop partial-word rounding | Node and native: length 0 instead of 1 for 4 bits |
| 12 | OR zero instead of shifted digit | Node and native: first word 0 instead of 32768 |

Witness 09 also runs large hex and octal literals, binary zero, mixed prefix
case and decimal leading zeroes. Its octal carry mutant changes the 24-digit
octal case from 4722366482869645213695 to 4667026250644221394943. That hardest
whole-function case is Node-held, with native compilation explicitly blocked.

## Findings for runtime task #rkvx6ny

Observed on main plus 04f18a2a: Uint16 numeric-length construction and zero fill,
element length, scalar element loads/stores, modulo-65536 stores, shared subarray
views, and checked dense loads using `?? 0` all lower and agree with Node.
The octal spill control produces `32768,3`, proving this reduction observes both
low-word truncation and next-word carry. Counts balance and sanitizer checks pass.

04f18a2a's implementation and its own report additionally cover number-array
construction, fill, same-kind overlapping set, byteLength, iteration, and view
lifetimes. Those broader facilities were not rerun by this scout and are not
scanner closure sites. See that branch's `docs/typed-arrays.md` and
`docs/typed-arrays-lowering.md` for its reported proofs and contract.

The remaining observed issues are checker proof for compound reads, numeric
conditions, comma updates, and lowering a throwing `never` call as a value.
The latter affects the real slice's dense-read repair and deserves attention
from compiler lowering; storage support alone does not clear that path.
Neither production code nor the runtime branch was repaired by this scout.

Not covered: a complete scanner/native token run, parser/checker closures,
other element kinds, ArrayBuffer/DataView, set/fill/iteration outside the scanner,
out-of-bounds write policy, allocation exhaustion, or parallel sharing. No whole
package tests or full gate were run. The source mutant for each blocked witness
is proven on Node only; no sanitizer claim is made for its unreachable native path.

Setup succeeded after a stalled recursive fetch was replaced by targeted fetches.
Go 1.27.1, clang 20.1.8, Node 24.19.0; `nproc` is 5 and CPU quota is 4.
Timing lines: Node ready 0.097s, Go ready 0.102s, clang ready 0.506s,
markdown dependencies ready 2.089s, submodules ready 6.588s, go build ready
209.924s, test binaries deferred 210.462s, build cache warm 210.500s,
done 210.660s. Complete setup and focused verification logs are under `evidence/`.
The reproduced compiler passes the same twelve-witness `run.py --check`;
`evidence/reproduced-verification.txt` records the successful run. The prepare
script reproduced the staged tree hash above. Static syntax checks also passed.

TypeScript excerpts are Apache-2.0; see `NOTICE` and the repository's `LICENSE-APACHE`.

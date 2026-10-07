# Optional widening, waves 1 and 2

This adaptation declares three lazy type caches on `Type`, corrects the parallel
cache declarations on `UnionOrIntersectionType` to optional, declares
`errorModuleName` on `SymbolVisibilityResult`, and adds the two configuration metadata
fields on `JsonSourceFile`. It also makes the derived accessibility result allow
its existing explicit undefined initializer. It changes only type declarations.
`adapt.cjs` parses the current tree with stock TypeScript 6.0.3, resolves the
owning interfaces with the checker, copies the exact reviewed target contracts,
and rejects incompatible existing declarations. It does not rewrite relation
sites, add runtime initialization, or weaken checker options.

## Owner proofs and ranking

| Owner family | Additions or correction | Inventory sites cleared |
|---|---|---:|
| `Type` and parallel `UnionOrIntersectionType` cache declarations | `resolvedBaseConstraint?: Type`, `resolvedIndexType?: IndexType`, `resolvedStringIndexType?: IndexType` | 14 |
| `JsonSourceFile` | `extendedSourceFiles?: string[]`, internal `configFileSpecs?: ConfigFileSpecs` | 8 |
| `SymbolVisibilityResult` | `errorModuleName?: string \| undefined` | 2 |

The cache properties are genuinely absent immediately after `createType` creates
a type. `createIntersectionType` and `getUnionTypeFromSortedList` populate flags,
constituents and alias data, without initializing these caches. The compiler and
services type constructors do not initialize them either. `getResolvedBaseConstraint`
lazily assigns the result of `getImmediateBaseConstraint`; it is a `Type`,
including the compiler's sentinel types. No cache writer stores present undefined.
`getIndexTypeForGenericType` separately fills the ordinary and
string-only caches with `createIndexType`, an `IndexType`. The corresponding
`InstantiableType` declarations already describe these as optional. A general
`Type` view can hold an instantiable, union or intersection object with those
same caches present. No cache setter writes a boolean, string or unrelated object.
The stock-checker writer/declaration audit is saved per property in the evidence.

The original required cache declarations on `UnionOrIntersectionType` do not
match construction or lazy getters. Simply adding optional base slots caused
TS2320 at `ResolvedType`, `IterableOrIteratorType` and `PromiseOrAwaitableType`,
which join object and union views. Correcting those three parallel declarations
to the same optional slots removes the conflict. This is an owner correction,
not a cast at a use. All six cache declarations remain `@internal`.

`hasVisibleDeclarations` constructs visibility results without a module error
name. `isAnySymbolAccessible` and `isSymbolAccessibleWorker` also construct
accessibility results, which extend the same visibility shape, with an
`errorModuleName` returned by `symbolToString`, or explicitly undefined. These
objects can be held through the base visibility interface. The existing derived
property is optional and string-valued; its declaration now explicitly permits
undefined to match the conditional initializer. The audit finds the compatible
initializers and declaration; no setter supplies a different type. Both interfaces are
`@internal`.

`Parser.parseJsonText` creates a normal JSON source file, attaches parse metadata,
and returns it as `JsonSourceFile` without either configuration property. The
configuration pipeline consumes those same objects through `TsConfigSourceFile`.
`parseJsonConfigFileContentWorker` writes `getConfigFileSpecs()` (a concrete
`ConfigFileSpecs`) to `sourceFile.configFileSpecs`. `parseConfig` writes
`arrayFrom(result.extendedSourceFiles.keys())` (a `string[]`) when extends metadata
exists. The audit finds exactly these two writers and their existing derived
contracts. Neither writes undefined or a different type. Both fields can be absent
on ordinary JSON files and present on configuration JSON files. The source copies
the exact target types. Only `extendedSourceFiles` is public; `configFileSpecs`
retains `@internal`.

Wave 1 originally included explicit `| undefined` on the caches. The stricter
comparison found that this disagreed with the exact optional targets. Wave 2
corrects those declarations to the exact target contracts, and corrects the
accessibility result's explicit-undefined contract. This is recorded rather than
rewriting the pushed wave-1 history. Final strict diagnostics are 8,712 before,
8,711 after, with zero introduced diagnostic code/message pairs. The program
still has existing errors under Adamic's options; this is not a whole-program
strict-checker pass.

## Measurements and review boundary

The complete production-rule inventory on a scratch compiler with refusal
`f1c9173eeeb073743c7233ad11f0d12c47363503` merged is **415 before, 391 after**:
24 sites cleared (wave 1: 16, wave 2: eight), no new sites. The stock owner catalog
resolves all inventory expressions back to source declarations. `evidence/wave2/census.json` ranks all
source/property groups, and `remaining-sites.json` and `remaining-sites.md`
retain every remaining site, source, target, expression and decline reason.

The latent census is a different meter: it examines eligible bodies of a
checker-rejected program and reports only its encountered lowering reasons.
It reports **73 before, 67 after** on the final sanctioned control. Wave 1 alone
reported 73 to 73 there, and 80 to 80 with all non-slice prerequisites. The final
67 findings are in `evidence/wave2/latent-remaining-sites.json`, with exact reasons.
They are labeled `refusal_scan` by the latent tool. Full raw evidence is recorded
separately. The 415 inventory is not claimed to be 415 eligible latent-lowering findings or 415
runtime failures. No whole-program native compilation is claimed.

The adaptation changes one source file and adds exactly one published API
property: `JsonSourceFile.extendedSourceFiles?: string[]`. `check-api.cjs` removes
that exact AST property from the emitted snapshot and requires byte equality
with the prerequisite snapshot. All other selected declarations are internal.
The oracle control mechanically reconstructs exactly the previously sanctioned 189 optional lines, 28 adaptation-40 lines
and one readonly line, and checks all 60,930 other reference baselines.

The final default sanctioned-control oracle passes **106,367 tests, zero failures,
zero pending**, with zero baseline differences. Install, build and tests exit 0;
test time 293.922s, wall time 303.894s, four workers, all runners. All ten emitted
JavaScript files are byte identical, all 710 source hashes in the sanctioned
control survive the second application, and 75 changes only
`src/compiler/types.ts`. The final API proof permits exactly the single JSON
property addition on top of the prerequisite lines and checks 60,930 other
reference baselines byte-for-byte.

## Existing integration limitations

The unmodified `stage3/apply.sh` on `origin/area/stage3` stops at adaptation 52:
`Error: adaptation 52 requires a declaration slice`. The temporary scanner
adaptations explicitly prohibit full-tree input. No infrastructure or neighboring
adaptation was changed. Scratch validation excludes those slice-only directories.

A full-tree control containing every other current adaptation reproduced 415
sites, but its API snapshot additionally contains changes from 31, 32 and 33.
The mechanical sanctioned-line checker rejects those existing differences.
Its default oracle produced 106,366 passing tests, one failing API snapshot test,
and only `api/typescript.d.ts` as a baseline difference. Therefore the sanctioned
oracle control uses prerequisites 00, 10, 20, 30, 40, 45, 46 and 70 only. This
subset is explicit in the provenance. It does not silently accept extra API lines.

## Reproduction

Use a scratch worktree from this branch and **merge**, never rebase, the refusal
branch there. Initialize its cohere dependency or link the already initialized
checkout. Do not merge the refusal into the adaptation publication branch.
Tool setup printed Go ready 0s, clang ready 1s, Node ready 1s, submodules ready
1s, build cache warm 165s, done 165s. `nproc` was 5, CPU quota four processors.
Source `/workspace/adamic-tools/env.sh` in every build/test shell.

On a tree produced by the sanctioned predecessors:

```sh
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/adapt/75-optional-widening/adapt.cjs TREE > adapt.json 2> adapt.log
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/adapt/75-optional-widening/audit.cjs TREE > audit.json 2> audit.log
npm run build --prefix TREE > build.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/adapt/75-optional-widening/verify.cjs BEFORE TREE > verify.json 2> verify.log
OPTIONAL_WIDENING_CONFIG=TREE/src/compiler/tsconfig.json OPTIONAL_WIDENING_OUTPUT=sites.jsonl go test ./internal/lower -run TestOptionalWideningCensus -v -count=1 -timeout 30m > census.log 2>&1
bash stage3/oracle/run.sh TREE NEW_ORACLE_OUTPUT > oracle.log 2>&1
```

Use `prepare.py NEW_TREE` to reproduce the explicit sanctioned prerequisite
subset without changing the repository scoreboard. Build pristine, BEFORE and
adapted trees with upstream's build. Before the oracle, run
`check-api.cjs PRISTINE BEFORE TREE OPTIONAL20_LEDGER READONLY70_LEDGER --accept-api`.
It invokes the existing adaptation-30 proof for the prerequisite API lines and
then mechanically checks the single addition. Only the mechanically proved
snapshot is accepted in the disposable tree. No repository baseline is edited.

`mutant.cjs TREE wrong-type` changes the new base constraint cache to boolean.
The stock-checker audit fails its declaration compatibility assertion. The
`drop-owner` mutant removes all three `Type` additions; the complete inventory
rises from 391 to 405. The `wrong-json-type` mutant changes the public
JSON member to `number[]` and fails the same audit. The `drop-json-owner` mutant
removes both JSON declarations and raises the inventory from 391 to 399. Mutants
run on an isolated tree. `verify.cjs` checks all emitted JavaScript bytes and source hashes, reruns 75, and requires zero edits.

`proof-mutants.cjs BEFORE TREE` creates an isolated copy, changes an emitted
JavaScript byte, and verifies that the byte comparison fails. It then supplies
an owner declaration requiring another edit, and verifies that the zero-edit
idempotence assertion fails. The API checker rejects a wrong public property
type and an extra unreviewed API line. Their recorded results accompany the
owner mutants. `check-strict.py BEFORE_AUDIT AFTER_AUDIT` requires the five
strict options and zero introduced code/message pairs. Its negative control
injects the real first-wave cache-contract diagnostics into the baseline report
and fails; unrelated prerequisite diagnostics are excluded. `measure-latent.py`
counts the rule's exact reason rather than optional-call syntax or file paths.

Two wave-2 default oracle attempts stopped after worker exits, with two observed
cgroup OOM kills. Disposable trials occupied almost 8 GB of memory-backed `/tmp`.
Removing five already-recorded trial checkouts freed about 4 GB; the final attempt
uses the unchanged four-worker default. Failed attempt reports are retained.

## Declines for @system_adamic

`never` sources have no owning object declaration to enlarge. The general `Type`
to `TypeParameter` cases need subtype proof: adding `constraint` alone exposes
more optional fields, and `target` has incompatible meanings across type kinds.
Parsed package JSON can contain a non-string version, which its reader handles
with a runtime type check; claiming `version?: string` is not truthful.
Compiler-only `SourceFile` lacks a method required by services augmentation;
a direct optional member conflicts with that declaration, and a heritage change
needs a separate public API proof. Other remaining owners are explicitly
unproven, not asserted impossible. They are declined pending construction and
writer review. The complete site ledger names every one.

A broader `TextRange.source` candidate was tested in scratch and declined: it
builds, but adding that optional target field propagates to every TextRange view
and expands the inventory from 399 to 859 sites. Bare nodes/ranges are fallback
views of a distinct SourceMapRange shape; no universal declaration is published.

The first-wave estimate, sent after pushing `bbb99fad`, was 24–40 original sites
by October 7 23:49 MDT. These two waves deliver 24. Remaining owners need evidence;
no guarantee is made that the deferred sites have declaration-only fixes.

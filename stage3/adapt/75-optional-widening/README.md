# Optional widening, wave 1

This adaptation declares three lazy type caches on `Type`, corrects the parallel
cache declarations on `UnionOrIntersectionType` to optional, and declares
`errorModuleName` on `SymbolVisibilityResult`. It changes only type declarations.
`adapt.cjs` parses the current tree with stock TypeScript 6.0.3, resolves the
owning interfaces with the checker, copies the exact reviewed target contracts,
and rejects incompatible existing declarations. It does not rewrite relation
sites, add runtime initialization, or weaken checker options.

## Owner proofs and ranking

| Owner family | Additions or correction | Inventory sites cleared |
|---|---|---:|
| `Type` and parallel `UnionOrIntersectionType` cache declarations | `resolvedBaseConstraint?: Type \| undefined`, `resolvedIndexType?: IndexType \| undefined`, `resolvedStringIndexType?: IndexType \| undefined` | 14 |
| `SymbolVisibilityResult` | `errorModuleName?: string \| undefined` | 2 |

The cache properties are genuinely absent immediately after `createType` creates
a type. `createIntersectionType` and `getUnionTypeFromSortedList` populate flags,
constituents and alias data, without initializing these caches. The compiler and
services type constructors do not initialize them either. `getResolvedBaseConstraint`
lazily assigns the result of `getImmediateBaseConstraint`; it is a `Type` or
undefined. `getIndexTypeForGenericType` separately fills the ordinary and
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
property is optional and string-valued. The audit finds its compatible initializer
and declaration; no setter supplies a different type. Both interfaces are
`@internal`.

## Measurements and review boundary

The complete production-rule inventory on a scratch compiler with refusal
`f1c9173eeeb073743c7233ad11f0d12c47363503` merged is **415 before, 399 after**:
16 sites cleared, no new sites. The stock owner catalog resolves all inventory
expressions back to source declarations. `evidence/wave1/census.json` ranks all
source/property groups, and `remaining-sites.json` and `remaining-sites.md`
retain every remaining site, source, target, expression and decline reason.

The latent census is a different meter: it examines eligible bodies of a
checker-rejected program and reports only its encountered lowering reasons.
It reports **73 before, 73 after** on the sanctioned control, and **80 before, 80 after** with all non-slice prerequisites. Its full raw evidence is recorded separately. The 415
inventory is not claimed to be 415 eligible latent-lowering findings or 415
runtime failures. No whole-program native compilation is claimed.

The adaptation changes one source file and adds no published API property:
all selected declarations are internal. The published snapshot must be byte
identical before and after 75. The oracle control mechanically reconstructs
exactly the previously sanctioned 189 optional lines, 28 adaptation-40 lines
and one readonly line, and checks all 60,930 other reference baselines.

The default sanctioned-control oracle passes **106,367 tests, zero failures, zero pending**, with zero baseline differences. Install, build and tests exit 0; test time 451.162s, wall time 462.214s. All ten emitted JavaScript files are byte identical, all 711 source hashes survive the second application, and 75 changes only `src/compiler/types.ts`.

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

Before the oracle, run the existing adaptation-30 `check-api-baselines.cjs` with
`TSC_ADAPT_READONLY_REPORT` set to 70's owner report and 20's declaration ledger,
a built pristine pinned control, and `--accept-api`. Only the mechanically proved
snapshot is accepted in the disposable tree. No repository baseline is edited.

`mutant.cjs TREE wrong-type` changes the new base constraint cache to boolean.
The stock-checker audit fails its declaration compatibility assertion. The
`drop-owner` mutant removes all three `Type` additions; the complete inventory
rises from 399 to 413. Mutants run on an isolated tree. `verify.cjs` checks all
emitted JavaScript bytes and source hashes, reruns 75, and requires zero edits.

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

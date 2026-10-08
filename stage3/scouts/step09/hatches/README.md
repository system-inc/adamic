# Step 09: double casts and refused hatches

Scout for #kbcj5nr; checked-cast owner #b5w3ycg. Base main is
`45487a809f89885a3fc651cd590e7dabf31362dc`; upstream TypeScript 6.0.3 is
`050880ce59e30b356b686bd3144efe24f875ebc8`. No adaptation or compiler implementation was changed.

## Scope and observations

The fresh `stage3/apply.sh` tree has **79 TypeScript files and 82 regular files** in
`src/compiler`, including three JSON/config files. The historical 81-file closure
predates adaptation 47's `hostErrors.ts`. I measured the actual main tree rather
than excluding that file to obtain the requested historical count. File names and
SHA-256 hashes are recorded in `ledger.json` and the compressed API inventory.
Coordinates throughout this scout are **main's adapted tree**, not upstream line
numbers before adaptations. Generated compiler code is included; library declaration
files are used for typing but not counted as corpus sites. Comments and emitted-JS
strings are excluded: they are not operations performed by the compiler itself.

| Inventory | Count | What the count establishes |
| --- | ---: | --- |
| `as unknown as` | 5 | AST bridge chains |
| `as any as` | 13 | AST bridge chains |
| Explicit `any` tokens | 119 | Additional context, not 119 double casts |
| Broad `Function` type references | 0 | Adapted corpus type references |
| Broad `Function`-typed value expressions | 1 | `debug.ts:376:26`, `Function.prototype` |
| Property writes including compound assignments and increments | 3,011 | Complete syntactic inventory for these store forms |
| Possible additions selected by static analysis | 49 | Candidates, not an exact runtime-expando census |
| `Object.defineProperty` / `Object.defineProperties` | 2 / 7 | Nine calls, 24 descriptor names |
| `Object.assign` into existing object | 1 | `utilities.ts:8581:5`, allocator update |
| Predicate/assertion annotations | 651 | Includes callbacks and overload signatures |
| Declarations with a body | 580 | 3 proven, 577 refused by main proof routine |
| Declarations without a body | 71 | All refused by main proof routine |

There are 16 `asserts` annotations: two on declarations with bodies and 14 on
bodyless declarations. Of all 651 annotations, main proves only
`core.ts:1769:42` (`isString`), `core.ts:1773:39` (`isNumber`), and
`watchPublic.ts:739:77` (`isFileMissingOnHost`). All other 648 have exact refusal
messages in the ledger. **Refused means main did not prove the annotation**, not
that its body is necessarily dishonest. Typical real tag guards exceed the current
proof syntax. The empty `Debug.type` assertion overload is a specific unsound
promise demonstrated independently by fixture 03.

## Method and exact evidence

`inventory.cjs` uses the stock TypeScript **6.0.3 compiler API**, a program/checker,
AST node kinds, receiver types, declaration origins, and direct literal allocations.
It inventories the entire adapted corpus. This is not a text search. It records
both `defineProperty` and `defineProperties`, resolves global `Object` symbols,
and records all selected stores even when a field is declared in an interface.

TypeScript accepts annotated predicates on trust, so its checker alone cannot answer
whether Adamic proves their bodies. `make-overlay.py`, `predicate_hook.go.txt`, and
`predicate-probe/main.go` expose **unchanged main `provePredicate` and `castProof`**
in a scratch Go overlay. The overlay preserves the 320 baseline whole-corpus checker
diagnostics so measurements can continue. It disables normal Load entry points and
native lowering; the probe asserts that IR output is disabled. Production compiler
files are neither changed nor committed. This is a latent proof measurement on a
diagnostic-bearing corpus, **not a successful native build**.

`direct-casts.cjs` makes virtual edits removing only each of the 18 bridge casts.
The virtual edits yield 336 retained checker diagnostics. All 18 resulting direct casts are still **Refused** by main's exact cast proof;
there are zero checked-tag or checked-class successes. Diagnostics, API bodies,
locations, types, direct edit texts, file hashes and counts are recorded in:

- `LEDGER.md`: every bridge, descriptor, Function value, selected property candidate,
  and predicate annotation by file:line:column.
- `ledger.json`: each selected site's coverage rule, source-edit proposal or stated
  lack of an established equivalent edit, and exact proof diagnostics.
- `evidence/inventory.json.gz`: raw API inventory, including **all 3,011 writes**
  and all 119 `any` tokens.
- `evidence/predicates.json.gz`, `evidence/direct-proofs.json.gz`: raw main proof
  observations, including full checker diagnostics.
- `fixture-results.json`: exact Node stdout/stderr/exit and main refusal messages.

## Coverage and truthful edits

All 18 double casts remain open. Existing main checked casts handle discriminated
object unions and nominal class downcasts; they do not certify an `unknown`/`any`
bridge, phantom array brands, arbitrary generic results, function variance, dynamically
installed debug methods, or cross-enum conversions. Their concrete direct replacements
were measured rather than assumed to work.

Adaptation 40 changes sorted-array brand declarations but leaves their required
non-runtime brand and these casts. A truthful static redesign removes the phantom
required field, uses arrays/readonly arrays, and preserves sortedness through audited
producers. Adaptation 41 removes broad Function annotations and the old Function.name
hatch but leaves intrinsic `Function.prototype` access. `func.toString()` is not an
equivalent replacement because an override may run. Adaptations 47 and 76 add host/key
checks; their presence does not prove every predicate or dynamic store. Optional-field
adaptations do not establish runtime own-property presence.

The ledger assigns each cast a concrete domain proposal: actual AST unions or producer
callbacks instead of arbitrary T, inline JSDoc/source-file checks, caller-supplied generic
comparators, external typed debug formatting, split builtin/custom builder contracts,
and explicit PollingWatchKind-to-WatchFileKind mapping. These are **proposals**, not
completed source adaptations or proven equivalent consumer rewrites. Public generic
contracts cannot be repaired by another uncheckable assertion.

For each refused predicate, the ledger retains its body and main's reason. Inlining
real checks at consumers or removing an unsupported refinement promise is a direction
for an adaptation; it is not a universal mechanical fix. Bodyless callback/overload
contracts require implementation evidence. For empty `Debug.type`, returning void is
truthful; consumers must supply real checks because arbitrary erased T cannot be
validated by that empty function.

Descriptor sites remain open. The configFile descriptor is hidden and non-writable;
plain assignment changes observable behavior (fixture 02). Debug descriptors provide
getters and debugger hooks. The factory's id/symbol descriptors forward reads and
writes to a redirect target; replacing them with copies breaks aliasing. External
metadata/typed accessor rewrites need consumer and reflection audits. None is claimed
covered by checked casts.

`Object.assign(objectAllocator, alloc)` updates an already existing declared allocator
shape. It is **not confirmed expando**, and native acceptance was not measured. Replacing
it with fixed-slot assignments requires checking enumerable extra keys, getters, and
order across the actual producers; types alone do not establish equivalence.

## Deliberate incomplete part

I did **not** establish every property added after creation, nor a truthful equivalent
source edit for every open site. The API partitions stores as 2,766 declared-slot writes,
196 indexed dictionary/array writes, 18 any-receiver candidates, 17 unresolved computed
keys, 13 absent-in-literal candidates, and one missing-slot candidate. Declared properties
can still be added late at runtime; conversely an absent-in-literal field may have been
installed by an earlier dynamic store. Constructors, factories, aliases, prototype state,
conditional initialization, external objects and arbitrary keys need flow-sensitive
allocation/ownership analysis or runtime traces. The 49-row shortlist is explicitly
**not** exhaustive runtime classification. Every syntactic store is preserved for that
follow-up. Destructuring writes, reflective mutations through aliases, and unseen external
implementations are outside this store detector. No full adapted-tree Node suite or
native closure build was run, and no complete source adaptation was made.

## Three witnesses and mutants

The `.a` files keep the relevant adapted tsc body statements and a small driver.
Node's existing `oracle/node.mjs` is the external oracle, using Node 24.19.0; the runner
was not modified. Main is built separately without the measurement overlay.

| Witness | Original Node stdout | Main | Mutant and caught result |
| --- | --- | --- | --- |
| Sorted phantom brand | `1\n3\n0\n` | Refused unchecked cast | Seed `[0]`: `2\n0\n0,1\n` |
| configFile descriptor | `target\nfalse\nfalse\n` | Refused defineProperty | enumerable true: `target,configFile\ntrue\nfalse\n` |
| Empty generic assertion | `string\nnot-a-number\n` | Refused unproved bodyless overload | Real number check throws: exit 70, stderr `adamic: panic: Error: proof missing\n` |

Each source mutant fails equality with the original Node stdout/stderr/exit. None of the
three originals emits a native binary; there is no observed silent native miscompile and
no claim that native mutants were executed. `audit.py` additionally catches three
ledger mutations: a dropped cast, accepting empty Debug.type, and inventing a checked
direct cast. Positive controls are the three predicates main does prove.

## Reproduce the measurement

Run from the repository root on the pinned base. Use a fresh external scratch directory;
TREE is the output of main's `stage3/apply.sh`. Install stock TypeScript 6.0.3 outside the
repository. All test output goes to files. Commands used here (paths may be changed):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step09-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc > /tmp/step09-nproc.txt
STAGE3_CACHE=/workspace/scratch/step09-cache bash stage3/apply.sh /workspace/scratch/step09-adapted > /tmp/step09-apply.log 2>&1
go build -o /workspace/scratch/step09-adamic ./cmd/adamic > /tmp/step09-compiler.log 2>&1
HATCH_TYPESCRIPT=/workspace/scratch/step09-cache/api/node_modules/typescript/lib/typescript.js node stage3/scouts/step09/hatches/inventory.cjs /workspace/scratch/step09-adapted > /tmp/step09-inventory.json 2> /tmp/step09-inventory.stderr
python3 stage3/scouts/step09/hatches/make-overlay.py /workspace/adamic /workspace/scratch/step09-overlay > /tmp/step09-overlay-path.txt
go build -tags hatch_predicate_measurement -buildvcs=false -overlay=/workspace/scratch/step09-overlay/overlay.json -o /workspace/scratch/step09-predicate-probe ./stage3/scouts/step09/hatches/predicate-probe > /tmp/step09-probe-build.log 2>&1
/workspace/scratch/step09-predicate-probe /workspace/scratch/step09-adapted > /tmp/step09-predicates.json 2> /tmp/step09-predicates.stderr
node stage3/scouts/step09/hatches/direct-casts.cjs /workspace/scratch/step09-adapted /tmp/step09-inventory.json > /tmp/step09-direct-overlay.json
/workspace/scratch/step09-predicate-probe /workspace/scratch/step09-adapted /tmp/step09-direct-overlay.json > /tmp/step09-direct-proofs.json 2> /tmp/step09-direct-proofs.stderr
python3 stage3/scouts/step09/hatches/audit.py /tmp/step09-inventory.json /tmp/step09-predicates.json /tmp/step09-direct-proofs.json /workspace/scratch/step09-adapted > /tmp/step09-audit.log 2>&1
python3 stage3/scouts/step09/hatches/run-fixtures.py --compiler /workspace/scratch/step09-adamic --output /tmp/step09-fixtures-check --check > /tmp/step09-fixtures-check.log 2>&1
```

The API source was installed with `npm install --prefix ... typescript@6.0.3`; the bare
TypeScript cache was seeded from an existing pinned clone before apply. Setup succeeded:
Go 0.019s, Node 0.021s, submodules 0.062s, markdown 0.064s, clang 0.169s, go build 9.165s,
tests deferred 9.292s, cache warm 9.294s, done **9.319s**. `nproc` = **5**, CPU quota 4.
Raw setup/audit/fixture outputs and compressed `apply.txt.gz` are under `evidence/`. Only these scout measurements
and fixture checks were run; no whole-package or full gate confirmation was run.

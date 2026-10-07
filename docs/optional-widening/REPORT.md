Built an independent optional-property widening refusal with position fixtures, Node oracles, and a stage 3 inventory.
Code commit: c9c3e7ced58f23cec6ad22f6205f125d303cec3a; based on current origin/main 39638d9.
Validation: lower package and complete uncached oracle pass; 2,000 developer probes pass; 415 stage 3 sites found.
Mutants: all seven fail their intended assertions, including structural-as-class and skipped argument acceptance.
Limits: full repository gate stopped at about four minutes; the ruled class exception still admits a demonstrated inheritance hole.

## Rule and cohere handoff

`internal/lower/optional_widening.go` owns the rule. `refusals.go` calls it at the same AST relation pass as mutable invariance and nominal ancestry, after the existing mutable-view diagnostic. No optional-property logic was put in `invariance.go`, and no cohere source was copied. The implementation uses the checker shims referenced through the cohere submodule.

For an otherwise checker-accepted value relation `S -> T`, refuse when a property `p` is optional in `T`, `GetPropertyOfType(S, p)` is absent, and `S` is structural. The exact predicate is based on the optional symbol flag, not whether a property's value type includes `undefined`. Both writable and readonly optional properties are covered. A source that declares `p` is checked normally by TypeScript; its incompatible property type remains a loader checker error.

A nominal class instance, including a generic class instance, is exempt for its own missing property, as ruled. Its declared fields are still walked. A fresh object literal is exempt at its outer shape; its field expressions, shorthand properties, and spreads are separate relation sites. A fresh array or tuple literal similarly checks its elements at their own sites. Conditional branches are checked individually. A fresh rebuild means enumerating known fields, for example `{ x: view.x }`; `{ ...view }` copies hidden fields and does not establish absence.

The type walk follows shared properties at every depth, source union members, compatible target union members, type-parameter constraints, instantiated structural generic properties, array/tuple elements, and single nongeneric call signatures. Function result relations run forward and parameter relations run in the opposite direction. It retains a visited type-pair set for recursive types. Only checker-assignable pairs are widening relations; unrelated pairs and checked downcasts stay with their existing rules. The prelude's `WeakBrand` is a compiler ownership marker, so it is unwrapped rather than treated as a hidden runtime field.

For a leading object spread, compare the spread source with the contextual literal target and exclude explicitly overwritten target fields. Because an overwrite can repair an otherwise incompatible field, the spread root does not require whole-object assignability before excluding those fields. Recursion into retained fields still requires assignability. The `spread_other_missing.a` fixture proves that overwriting `y` does not conceal a missing optional `z`.

Covered relation positions are variable initializer, assignment, call argument, return, property write, object spread, array element, tuple element, explicit object field, shorthand object field, expression-bodied arrow return, parameter default, class-field initializer, and upcast. Nested structural fields and instantiated generic structural fields are walked, including readonly fields.

The pinned original diagnostic is:

```text
initializer.a:3:42: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source { x: number; }, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

The test pins the entire message, with its `testdata/optional_widening/` path prefix. Node runs the original source and prints `string`; Adamic now refuses its initializer before lowering.

## Exact refused program shapes

All position fixtures begin with:

```typescript
const original = { x: 1, y: 'wrong' };
const view: { x: number } = original;
```

The following suffixes are complete refused examples for cohere's lint. The `.a` fixtures are in [internal/lower/testdata/optional_widening](../../internal/lower/testdata/optional_widening).

| Fixture | Suffix |
|---|---|
| initializer | `const wider: { x: number; y?: number } = view; console.log(typeof wider.y);` |
| assignment | `let wider: { x: number; y?: number } = { x: 0 }; wider = view;` |
| argument | `function take(value: { x: number; y?: number }): void { console.log(typeof value.y); } take(view);` |
| return | `function take(): { x: number; y?: number } { return view; } take();` |
| property_write | `const holder: { value: { x: number; y?: number } } = { value: { x: 0 } }; holder.value = view;` |
| spread | `const wider: { x: number; y?: number } = { ...view };` |
| array_element | `const wider: { x: number; y?: number }[] = [view];` |
| tuple_element | `const wider: [{ x: number; y?: number }] = [view];` |
| field | `const wider: { value: { x: number; y?: number } } = { value: view };` |
| shorthand | `const wider: { view: { x: number; y?: number } } = { view };` |
| nested | `const holder: { readonly value: { x: number } } = { value: view }; const wider: { readonly value: { x: number; y?: number } } = holder;` |
| generic | `interface Slot<T> { readonly value: T; } const holder: Slot<{ x: number }> = { value: view }; const wider: Slot<{ x: number; y?: number }> = holder;` |
| arrow | `const take = (): { x: number; y?: number } => view;` |
| cast | `const wider = view as { x: number; y?: number };` |
| parameter_default | `function take(value: { x: number; y?: number } = view): void {} take();` |
| class_field | `class Holder { value: { x: number; y?: number } = view; } new Holder();` |
| readonly | `const wider: { readonly x: number; readonly y?: number } = view;` |

The overwrite probe is:

```typescript
const original = { x: 1, y: 'wrong', z: 'wrong' };
const view: { x: number; y: string } = original;
const wider: { x: number; y?: number; z?: number } = { ...view, y: 3 };
```

It refuses `z`. The incompatible declared-property control is:

```typescript
const source = { x: 1, y: 'wrong' };
const wider: { x: number; y?: number } = source;
```

TypeScript rejects that control as before, rather than the new refusal pass manufacturing a diagnostic.

## Allowed programs and compatibility repairs

Each allowed fixture is registered in `internal/oracle/optional_widening_test.go`, without editing `oracle_test.go`, and runs against source Node, generated JavaScript, release native C, sanitized native C, and the leak check.

| Fixture | Shape | Node stdout |
|---|---|---|
| optional_widening_class.a | `class Point { x = 1; }`, instance assigned to `{ x: number; y?: number }` | `undefined`, then `1` |
| optional_widening_fresh.a | `{ x: 1 }` assigned directly; hidden source rebuilt as `{ x: view.x }` | `undefined`, then `undefined` |
| optional_widening_declared.a | source declares `y?: number`; compatible required `y: number` viewed through readonly optional fields | `2`, then `3` |

A required mutable `y: number` seen as mutable `y?: number` still trips the existing mutable-invariance rule. That separate policy was not changed; the required-property compatibility oracle therefore uses a readonly target. The optional declared-source oracle uses the same optional shape on both sides.

An explicit spread overwrite `{ ...view, y: 3 }` is allowed by the refusal pass and tested directly. Native lowering still cannot add a field absent from the spread source's declared shape, so that positive is not claimed as a compiled Node/native oracle.

Two existing oracle fixtures needed repair. `e4eec87_u02_optional_absent.a` now declares the raw source as `Reading`, so its optional `value` is declared without adding a runtime field. The Boolean iterator in `user_iterators.a` is now a nominal `BooleanIterator` class, preserving its concrete method origin while satisfying the ruled class exception. Merely adding a structural return annotation caused the existing iterator-origin NotYet, so the final repair uses a class. Both repaired fixtures agree with Node. The existing iterator mapper-index test now uses a nominal source and still asserts its original index-representation NotYet.

The complete counts table was regenerated. Only the iterator repair and the three new fixtures changed rows:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| user_iterators.a before | 669 | 669 | 466 | 915 | 67 | 0 |
| user_iterators.a after | 663 | 663 | 458 | 907 | 64 | 0 |
| optional_widening_class.a | 3 | 3 | 2 | 7 | 3 | 0 |
| optional_widening_fresh.a | 3 | 3 | 2 | 8 | 3 | 0 |
| optional_widening_declared.a | 6 | 6 | 2 | 12 | 4 | 0 |

The iterator's class replaces an index capture cell and two function closures with ordinary class methods. The counts are observations of that fixture repair, not an optimization claim for the new rule.

## Mutants actually run

Each mutant changes the production Go rule, runs its named tests, and restores the original source in `finally`. The scripts are [mutants.py](mutants.py) and [extra-mutants.py](extra-mutants.py). All seven test invocations exit 1. No mutant was killed by a compiler error or clang warning.

| Mutant | What caught it |
|---|---|
| Treat every structural source as an exempt class | initializer and argument fixtures compile, lowering returns nil; refusal assertions fail |
| Skip call-argument sites | argument fixture compiles, lowering returns nil; refusal assertion fails |
| Skip shared-property recursion | nested and generic fixtures compile, lowering returns nil; refusal assertions fail |
| Remove the class exemption | allowed class fixture is refused; allowed-source assertion fails |
| Remove fresh-literal exemption | allowed fresh fixture is refused; allowed-source assertion fails |
| Skip object spread sites | spread fixture compiles, lowering returns nil; refusal assertion fails |
| Require whole-object assignability before ignoring spread overwrites | overwrite-with-other-missing-property fixture compiles, lowering returns nil; `z` refusal assertion fails |

Individual logs are under [evidence](evidence). The two required mutants were repeated after the assignability guard was finalized. The additional mutants also ran on the final production rule, and the restored compiler passed all lowering tests afterward.

## Developer refusal probes

Fetched `origin/devtools/refusal-probes` at exactly `5e3c7b73be88d2be60e0f78def962ad72da14d84`. Its command and package were linked to this compiler through a scratch Go overlay. The only tool adjustment registers the new `refuseOptionalWidening` helper with the existing `optional-widening` catalog entry; otherwise its source audit rejects any new helper before probing. No catalog scenes or judging logic were changed.

```sh
go run -overlay=/tmp/optional-widening-tools/overlay.json ./cmd/adamic-refusals \
  -only optional-widening -seed 1 -count 2000 \
  -out /tmp/optional-widening-probe-findings-final \
  > /tmp/optional-widening-probes-final.jsonl 2> /tmp/optional-widening-probes-final.log
```

Exit 0; summary `Programs: 2000`, `Counts: {"refused-as-expected": 2000}`. No finding programs were emitted. The developer tool checks each repaired neighbor through load, lower, and C generation first, then requires this category's specific refusal. [probes.jsonl](evidence/probes.jsonl) preserves its complete boundary list and summary. This is generated scene coverage, not 2,000 distinct relation-position shapes or a runtime oracle for every probe.

The overlay can be recreated with [probe_overlay.py](probe_overlay.py), after fetching the developer branch, and then the command above rerun.

## Stage 3 census: every site

**415 distinct AST locations, in 42 matching files, among 79 compiler implementation files scanned.** Each location is listed with its first missing optional property, exact source type and exact target type in [stage3-sites.md](stage3-sites.md). [stage3-sites.jsonl](stage3-sites.jsonl) also includes the complete expression text. Counts by file are in [stage3-file-counts.json](stage3-file-counts.json).

This is TypeScript's stage 3 **compiler project**, `src/compiler/tsconfig.json`, on the adapted tree produced by current `stage3/apply.sh`, from upstream v6.0.3 commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Declaration files and files outside that compiler directory are not counted. Hashes of the compiler source, config, and dependency lock are in [stage3-hashes.json](stage3-hashes.json). Services, server, harness and test projects were not separately inventoried.

```sh
stage3/apply.sh /tmp/optional-widening-tsc > /tmp/optional-widening-apply.log 2>&1
npm ci --prefix /tmp/optional-widening-tsc --ignore-scripts --no-audit --no-fund \
  > /tmp/optional-widening-tsc-install.log 2>&1
OPTIONAL_WIDENING_CONFIG=/tmp/optional-widening-tsc/src/compiler/tsconfig.json \
OPTIONAL_WIDENING_OUTPUT=/tmp/optional-widening-tsc-sites.jsonl \
  go test ./internal/lower -run TestOptionalWideningCensus -v -count=1 -timeout 30m \
  > /tmp/optional-widening-census-final.log 2>&1
```

Exit 0, `79 files, 415 optional-widening relation sites`, 12.809s in the final recorded run. The first inventory was deliberately discarded because it paired unrelated function parameter types and checked downcasts; the final production walk requires an accepted assignment relation. A subsequent spread-overwrite fix added five genuine root-spread matches, bringing 410 to 415.

The census invokes the production `optionalAtSite` walker directly, so it sees every matching site even when another Adamic refusal would stop lowering first. It uses upstream's project options and bound checker types; it does not run native lowering or certify a checker-rejected program. No successful whole-project stage 0 build is claimed, and 415 is a count of this rule's matching AST sites, not a count of functions that demonstrably misbehave at runtime. Only the first failing property per AST site is recorded, just as the refusal diagnostic does. Nested conditional nodes and their branches can be distinct matching locations.

## Commands, timing and limits

Every test invocation wrote output to a log; none was piped.

```sh
bash cloud/setup.sh > /tmp/optional-widening-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./internal/lower -count=1 > /tmp/optional-widening-lower-restored.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m \
  > /tmp/optional-widening-oracle-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts \
  > /tmp/optional-widening-counts-final-clean.log 2>&1
gofmt -l cmd internal > /tmp/optional-widening-format.log
go vet ./... > /tmp/optional-widening-vet-restored.log 2>&1
git diff --check
```

Observed: setup exit 0; Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 120s, done 120s. `nproc` is 5; cgroup quota is 4 CPUs (`400000 100000`), reported memory 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. `/opt/adamic-tools/env.sh` was absent; the configured tools path is `/workspace/adamic-tools/env.sh`, which was sourced in build/test shells.

All restored lowering tests pass (8.670s); the complete oracle package passes uncached (122.898s); the focused allowed-and-repaired fixture oracle passes uncached (25.211s). Format and vet logs are empty, exit 0; `git diff --check` is clean. The final counts update passes, with only the rows listed above moving.

The full command `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/optional-widening-gate.log 2>&1` was stopped after about four minutes under CLAUDE.md's worker-gate exception, exit 143. Its bridge/corpus work was still running. [partial-gate.log](evidence/partial-gate.log) is partial; the full repository gate is not claimed green. The complete touched lower and oracle packages were then validated separately.

`git fetch origin && git merge origin/main` reported `Already up to date`. Only `codex/optional-widening-refusal` is committed and pushed; no PR was opened and no main or area branch was changed.

Not separately proven by fixtures: array-spread relation propagation, destructuring relations, overload signature views, generic function-value views, every union/intersection combination, or all non-compiler stage 3 projects. The implementation shares the existing relation-site recognition, whose extra cases are not an exhaustive new soundness proof.

## Remaining class-exemption hole observed

The ruling's class exception assumes a class's fields are truly absent. Current main now supports inheritance, so that premise can fail through a base view:

```typescript
class Base { x = 1; }
class Derived extends Base { y = 'wrong'; }
const view: Base = new Derived();
const wider: { x: number; y?: number } = view;
console.log(typeof wider.y);
```

Observed on the restored final compiler: Node prints `string`, exit 0. `adamic c` and `adamic build` accept the program. Native execution prints `adamic: panic: compiler bug: a numeric field holds a reference`, exit 70. The acceptance also existed before this rule; this unit follows the expressly ruled nominal-class exception and does not close it. This is a remaining type lie and Node/native mismatch, not a passing oracle or a claim that class views are proven safe.

A fixed-field guarantee for base-class views, or a revised exception limited to sources whose dynamic field absence is proven, needs a separate ruling. The simplest class fixture requested here has no subclasses and does pass Node. The inherited-class counterexample is recorded as evidence rather than registered as a positive oracle fixture that would falsely claim agreement.

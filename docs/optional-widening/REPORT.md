Refined the class exemption across whole-program transitive descendants and added six developer widening lies with pinned refusals and repaired oracles.
Commits: b9715978334b534815a01fdea9eedbb3ae2136bb; prerequisite repair 9887f3772dc8d9aa4bd1a4898ae21110f6408010 (including f1c9173); current main b6b1538.
Validation on repaired base: lower passes (42.590s), stage3/fixtures passes (29.361s), complete uncached oracle passes (150.221s); counts, vet, format and whitespace checks pass. Setup completed in 178s; nproc 5, cgroup quota 4 CPUs.
Mutants: seven independent mutants fail their intended assertions, including direct subclasses only, caught by class_transitive compiling.
Limits: Function programs skipped for codex/refusal-pass-rulings; compound generic matching is conservative; full repository gate and stage 3 census not rerun.

## October 7 refinement

The refinement branch merges the verified prerequisite repair `9887f3772dc8d9aa4bd1a4898ae21110f6408010`, retaining the cast-specific proven-relations message and all non-cast message pins. The rule follows @system_adamic's October 7 ruling. When a nominal source lacks an optional target property, scan every class declaration and class expression in the loaded program's roots and imported modules, then follow nominal ancestry transitively. A descendant whose property type is not assignable to the target property type refuses the widening. The message identifies the descendant, property, and declaration location. A compatible descendant stays allowed even when the source view hides its property.

This exemption relies on Adamic's closed world: the whole program is compiled together. If separate compilation ever arrives, the exemption must be revisited. The same assumption is written beside the production rule.

Generic ancestry is matched against the source's invariant type arguments. Direct parameters and nested reference arguments are substituted before checking the descendant's property. The positive generic fixture has both `Derived<T> extends Base<T>` with `y: T` and an unrelated `Base<string>` descendant with `y: string`: a `Base<number>` view remains allowed and Node prints `number` then `2`. An unresolved compound generic argument is conservatively retained as a possible descendant. This can refuse a compatible compound instantiation whose proof would require more inference; no exhaustive generic inference is claimed.

The new class fixtures are `class_derived`, `class_transitive`, `class_expression`, `class_generic`, and `class_generic_compound` in `internal/lower/testdata/optional_widening`. Each pins its entire diagnostic in a `.refused` file, normalizing only the repository's absolute path. Source Node prints `string` for each. The two accepted subclass fixtures are registered in the oracle and print `number` then `2` in source Node and both backends.

The original counterexample is now refused with:

```text
testdata/optional_widening/class_derived.a:4:42: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from class source Base, but subclass Derived declares y at testdata/optional_widening/class_derived.a:2:30 with an incompatible type; make the subclass property assignable to the target, or build a fresh object with known fields (adamic/no-optional-widening)
```

`TestOptionalWideningWholeProgram` also covers a descendant in an imported module while the value is constructed as the base class. An additional case exercises multiple checker roots directly; native `Lower` accepts one entry file, so that case is a refusal-pass check, not a claim that multiple-entry native lowering exists.

## Developer observed programs

Fetched `origin/devtools/refusal-lies` at `4fb0b5ed9ce8f0137be47cbe3f7c9b611b54e51e`. Copied the six `widening_*.a` programs and their `.expected` Node observations unchanged into the lower fixtures. Every program is refused, and each complete message is pinned separately. `TestOptionalWideningObservedNode` checks the independent Node observations, including exit 0 and empty stderr.

| Original fixture | Observed Node stdout, lines separated by commas |
|---|---|
| widening_boolean_as_number | false, boolean, 2 |
| widening_boolean_as_number_array | boolean, false, 2 |
| widening_boolean_as_number_ops | false, true, 10, true, false |
| widening_boolean_as_number_parameter | 2, 101, 3 |
| widening_boolean_as_number_return | boolean, false, 2 |
| widening_false_as_number | false, boolean, 1 |

Their accepted neighbors retain the original producer and numeric consumer, replacing the widening relation with a fresh object containing only the known `x` field. The array neighbor maps each element to that fresh shape, with an explicit result type; the parameter neighbor rebuilds at the argument; the return neighbor rebuilds at the return. Each repaired source is independently run on Node and compared against generated JavaScript, release native, sanitized native, and the leak check. These are repairs, not claims that the original lying outputs can be soundly retained.

Skipped `function_type_length.a` and its observation. All Function-typed programs belong to `codex/refusal-pass-rulings`, as requested.

Eight new oracle count rows were generated. No existing row moved. The two subclass programs each record allocations 3, frees 3, retains 2, releases 7, peak 3, regions 0. The repaired developer programs' complete counts are in `internal/oracle/counts.md`.

## Refinement mutants and verification

[class-mutants.py](class-mutants.py) mutates the real rule independently and restores it in `finally`. Every test exits 1 through its intended assertion; none is killed by a build error, panic, or clang warning.

| Mutant | What catches it |
|---|---|
| Direct subclasses only | class_transitive compiles; refusal assertion sees nil |
| Skip class expressions | class_expression compiles; refusal assertion sees nil |
| Skip imported modules | WholeProgram/true loses the Derived refusal |
| Skip extra checker roots | WholeProgram/false loses the Derived refusal |
| Skip generic substitution | generic_subclass is wrongly refused as Derived<T> |
| Skip unresolved compound arguments | class_generic_compound compiles; refusal assertion sees nil |
| Exempt structural sources | All six developer widening programs compile; refusal assertions see nil |

Logs are under [evidence](evidence), with the `class-mutant-` prefix. The first whole-oracle attempt found a path error in the new observation test; paths now use the harness's absolute-path contract. The broad lower attempt also found the checker's `Types()` accessor requires a union/intersection guard, now fixed. Final restored results are recorded below; earlier failed attempts are not counted as green validation.

```sh
bash cloud/setup.sh > /tmp/optional-widening-2-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 docs/optional-widening/class-mutants.py > /tmp/optional-widening-2-mutants.log 2>&1
go test ./internal/lower -count=1 > /tmp/optional-widening-2-lower-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/optional-widening-2-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/optional-widening-2-counts.log 2>&1
gofmt -l cmd internal > /tmp/optional-widening-2-format.log
go vet ./... > /tmp/optional-widening-2-vet.log 2>&1
git diff --check
```

Observed final results: lower exit 0, `ok github.com/system-inc/adamic/internal/lower 32.125s`; complete uncached oracle exit 0, `ok github.com/system-inc/adamic/internal/oracle 154.607s`; counts update exit 0, `ok github.com/system-inc/adamic/internal/oracle 33.210s`. Vet and format logs are empty, and the final whitespace check is clean. The focused uncached optional-widening oracle also passed (2.251s), before the final conservative compound-generic guard; the complete final oracle above includes that guard and all new observations.

The focused command was:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/optional_widening' -count=1 -timeout 30m > /tmp/optional-widening-2-oracle-focused.log 2>&1
```

[class-lower.log](evidence/class-lower.log), [class-oracle.log](evidence/class-oracle.log), [class-counts.log](evidence/class-counts.log), [class-setup.log](evidence/class-setup.log), and [class-mutants.log](evidence/class-mutants.log) preserve the final outputs.

Setup timings: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 178s, done 178s. `nproc` is 5 and cgroup `cpu.max` is `400000 100000`, four CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. The environment file printed by setup is `/workspace/adamic-tools/env.sh`, sourced in test shells.

Latest `git fetch origin && git merge origin/main` reports `Already up to date`. Only `codex/optional-widening-2` is pushed; no PR is opened. No forbidden central compiler or oracle file was edited, and no cohere code was copied.

Not covered: complete generic inference for compound inherited arguments, every generic/class-expression combination, or separate compilation. The standalone stage 3 census has no loaded Adamic program, so its advisory inventory still excludes the descendant refinement. The prior census and developer random probes below were not repeated. Full repository corpus tests were not run in this refinement; the touched lower and oracle packages and vet are the reported gate.

## Original refusal unit report, historical

The following records the prerequisite unit's observations before this refinement. Its class hole is closed for the fixtures above; its unconditional exemption description is historical.

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

## Class-exemption hole observed before the October 7 refinement

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


## October 7 main compatibility repair

`codex/optional-widening-refusal-2` starts at main `b6b1538` and merges the refusal at `f1c9173eeeb073743c7233ad11f0d12c47363503`, without the descendant refinement. Main's proven-relations check correctly refuses casts first. Both checks reject the same optional-field lie; the cast now pins that complete proven-relation diagnostic. Every non-cast refusal fixture pins its complete optional-widening message independently in a `.refused` file. Only absolute repository paths are normalized.

The two TypeScript-derived object fixtures were never recorded as Compiles. `01_reference_spreads.a` was Refused for `adamic/single-spread`, and is now Refused for missing optional `preserve` at `7:111`. `30_host_optional_method.a` was NotYet for an optional call at `6:12`, and is now Refused for missing optional `getCompilerHost` at `8:73`. Only their `stage0` outcome and diagnostic fields changed. All bytes outside the `stage0` values, including the Node observations and TypeScript provenance, are identical. Source programs and the TypeScript NOTICE are unchanged.

Linux verification:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./stage3/fixtures ./internal/oracle -count=1 -timeout 30m > /tmp/optional-widening-refusal-2-gate.log 2>&1
go vet ./internal/lower ./stage3/fixtures ./internal/oracle > /tmp/optional-widening-refusal-2-vet.log 2>&1
gofmt -l internal/lower/optional_widening_test.go > /tmp/optional-widening-refusal-2-format.log
git diff --check
```

Exit 0: lower 40.585s, stage3/fixtures 28.874s, complete uncached oracle 155.118s. Vet and format logs are empty; whitespace check is clean. Setup: Go, clang, Node and submodules ready 0s, cache warm and done 48s; `nproc` 5, cgroup quota 4 CPUs. Tool environment: `/workspace/adamic-tools/env.sh`.

Mutants actually run and restored: append `MUTANT` to the production non-cast diagnostic, caught by the assignment exact-message pin; use a stale cast diagnostic record, caught by the cast exact-message pin. Both tests exit 1 through the intended diagnostic assertions. Scratch copies with both original stage0 records also exit 1 through precisely two stage0 comparisons while both Node comparisons pass. No compiler errors or clang warnings kill these checks.

The repair gate and mutant logs are preserved under `evidence/refusal-2-*.log`. The inherited-class hole remains unchanged on this prerequisite branch and belongs to the refinement branch. No full repository gate, census, or Function worker work is claimed by this repair.


## Refinement verified on the repaired prerequisite

`codex/optional-widening-2` merges `codex/optional-widening-refusal-2` at `9887f3772dc8d9aa4bd1a4898ae21110f6408010`. The test conflict was resolved by retaining the repair's cast-specific proven-relations expectation and using exact complete-message pins for all non-cast fixtures, including every descendant and developer fixture. The production refinement is unchanged from `b9715978334b534815a01fdea9eedbb3ae2136bb`.

The required direct-subclasses-only mutant was repeated on this merged base. The two-level fixture compiled; its intended refusal assertion failed with `got <nil>`, exit 1. Production was restored before the final gate. The seven earlier refinement mutants remain documented above; this merged-base repeat is additional evidence.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./stage3/fixtures ./internal/oracle -count=1 -timeout 30m > /tmp/optional-widening-2-repaired-gate.log 2>&1
go vet ./internal/lower ./stage3/fixtures ./internal/oracle > /tmp/optional-widening-2-repaired-vet.log 2>&1
gofmt -l internal/lower/optional_widening_test.go > /tmp/optional-widening-2-repaired-format.log
git diff --check
```

Exit 0: lower 42.590s, stage3/fixtures 29.361s, whole uncached oracle 150.221s. Vet and format logs are empty; whitespace check is clean. `evidence/class-repaired-gate.log` and `evidence/class-repaired-mutant-direct.log` preserve the merged-base results. The two repaired TypeScript-derived status records retain identical Node observations on this branch too. Function programs remain skipped for the other worker, and compound generic matching remains conservative.

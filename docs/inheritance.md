# Class inheritance

Adamic lowers nominal class hierarchies to native objects with no garbage collector. An object has one allocation and one descriptor for its most-derived class. Inherited fields and method slots retain their base layout as a prefix, so a base-typed reference needs no conversion. Ordinary method calls select the implementation from the object's method table; `super.method(...)` calls the base implementation directly.

## Construction and ownership

Base field initializers run first, then the base constructor body, then derived field initializers and the derived constructor body. Constructor parameter defaults run in the corresponding constructor, after base fields for a root constructor. Derived fields run only after `super(...)` returns. A constructor cannot use `this` before that point.

One destructor chain releases derived fields before base fields, each once. Heap teardown remains iterative. Arena and immortal children do not acquire reference counts merely because they are stored in class fields. Private fields are qualified by their declaring class, including when base and derived classes use the same spelling.

`instanceof` follows nominal ancestry. Abstract classes cannot be constructed; concrete subclasses must implement abstract methods. Class identity does not survive an object spread, including when Perceus reuses the source allocation.

## Soundness

Overrides accept the base method's parameter type or a wider type and return its result type or a subtype. Parameters are not bivariant. Mutable inherited fields retain their declared type, and readonly fields cannot narrow mutable contents unsafely. Unsupported changes to native representations receive `NotYet` rather than generating an incompatible function-pointer call.

A class reference requires nominal ancestry. A structurally similar object or unrelated class cannot supply the method table a virtual call expects. This requirement also applies inside containers, function signatures, unions and `Weak`. Use interfaces for structural data views.

The cycle finder includes inherited fields and resolves checker private symbols to their native slots. Ownership-cycle diagnostics retain the original field spelling and propose a `Weak` repair.

## Verification

The oracle compares each fixture with its source on Node, generated JavaScript, native release and native under ASan/UBSan. Terminating fixtures must also pass LeakSanitizer. `internal/oracle/counts.md` records measured allocation, free, retain, release, peak and region counts.

The fixtures cover three hierarchy levels, abstract methods, overriding through base-typed arrays and parameters, direct super calls, observable initializer and constructor order, ancestor tests, private fields, a live Weak parent, and constructor and virtual-method exceptions. Mutants were caught for static dispatch, early derived initialization, skipped base destruction, unsound overrides, missing inherited cycle fields, premature this, nominal-view violations, omitted virtual throws, private-field aliasing and retained spread identity. The base-destructor mutant leaked computed strings, rather than relying on immortal literals.

Override signature regressions include adding a number default, removing a number default, and removing defaults from number and string methods. Lowering refuses each with `NotYet` naming `factor` and its repair. Removing the parameter comparison accepts all three and fails the refusal tests. Result and parameter-count comparison mutants also fail refusal assertions. Identical defaults, optional parameters and parameter widening with the same representation agree with source Node through both backends and pass sanitizers and the leak check.

The initial implementation passed `go vet ./...` and `go test -count=1 -timeout 30m ./...`. Cohere formatting passed on equivalent `.ts` copies; a full Cohere lint run had findings and is not claimed green.

## Regions and reuse

Inheritance does not disable region allocation or Perceus reuse. A region escape summary joins every virtual implementation: if any stores an argument, the call cannot receive that argument in a temporary region. Class allocators may allocate in a region when their constructed object does not escape initialization. Virtual return values currently keep the heap calling convention.

Virtual methods share a parameter ownership convention. If one implementation consumes a parameter for reuse, all implementations receive an owned count and release it. The global-move analysis follows every virtual target, preventing a move when an override can observe the global.

The continuation compared all 173 existing counted rows with pre-inheritance main (`fe3b9f2`). No allocation, retain, release or peak count increased, and no region count fell. Some frees became region teardown instead. The hierarchy memory fixture records 48 allocations, 42 heap frees, 22 retains, 60 releases, peak 11 and six values in regions. It exercises both reused spreads and class allocation in regions.

Mutants that ignored escaping overrides triggered ASan on an object stored beyond its temporary region. A mutant that omitted the shared consumption convention leaked under LeakSanitizer. Following only the static target for a global move produced a null access caught by UBSan and Node exit parity. Restoring either blanket optimizer guard failed the plan test. The optimization step used native and lowering package tests, filtered oracles and the complete counts update. The final full gate below verifies the combined implementation.

Outside the original lowering/native territory, this continuation updates `internal/oracle/class_inheritance_test.go` to register the memory fixture, `internal/oracle/testdata/class_inheritance_memory.a` to exercise dynamic ownership, and `internal/oracle/counts.md` to record measurements. `internal/native/emit.go` has small integration changes for joined region arguments and constructor allocation, and moves the final field-store temporary rather than retaining and releasing it. `docs/inheritance.md` is the reader-facing report.

## Generic hierarchies

`Box<T> extends Base` and `Pair<T> extends Box<T>` are monomorphized by concrete type argument identity. This avoids sharing a native layout between arrays and objects, or between arguments whose base projections differ. Base layouts and inherited method signatures use the actual substituted base view, including reversed type arguments and concrete subclasses of generic classes.

Every monomorphization retains a distinct descriptor and typed method table, while all descriptors from the same source declaration share one erased identity for `instanceof`. An identity-only descriptor supports testing an as-yet unconstructed generic class. Growing polymorphic recursion is refused with a repair, bounded by the existing depth limit of 32.

The generic oracle covers number, string, boolean, object and array fields, base-typed dispatch, overrides and super methods, a concrete `Pair<string>` subclass, reversed arguments, a Weak parent, and erased identity before construction. Generic factories retain the checker’s concrete type mapping, including projected base arguments. Tests also reject generic parameter narrowing, structural lookalikes in nominal arguments and constraints, and a cycle through an inherited generic field.

Collapsing the class cache to representation labels failed the separate-layout assertion for object fields. Comparing descriptors instead of erased identities changed Node/native ancestor results with both exiting successfully. Omitting base substitution failed the generic layout test. Native and lowering package tests, filtered oracles, vet and the complete counts update passed; all 173 pre-branch rows still have no regression.

Additional integration files for this step are `internal/ir/ir.go` (erased identity metadata), `internal/javascript/javascript.go` (matching erased ancestry tests), and `internal/native/runtime/adamic.h` (the descriptor identity). The oracle registration, counts table and new `internal/oracle/testdata/class_inheritance_generic.a` hold the generic fixture. Fresh structural literals are checked through their constituent class values rather than treated as pre-existing mutable views; spreads retain the aggregate check.

## Conditional construction

Derived constructors track whether `super` has returned on every path. Branches, nested branches, effect-only ternaries, switches, loops, throws and caught constructor failures preserve JavaScript's binding and initializer order. A successful call binds `this` before derived field initialization. A failed base constructor leaves the binding uninitialized and permits a retry.

Returning normally without a binding throws JavaScript's `ReferenceError`. A second successful call runs the base constructor on a separate object before throwing the duplicate-binding `ReferenceError`; it neither overwrites the first object's fields nor reruns derived initializers. Cleanup releases both failed constructions and temporary second objects.

The conditional fixture makes order observable with side effects and computed strings. It covers each branch, missing and duplicate calls, a retry after a throwing base, abrupt branches, switches, zero/one/two loop iterations and generic conditional constructors. Refusal tests cover `this` and `super.method` captures before definite initialization. Captures, catch/finally paths and loop backedges are conservative; accepted programs never read an uninitialized binding.

Mutants that bound `this` before the base returned, ran derived fields first, omitted the missing-super check, or reused the first object on a second call disagreed with Node. Disabling the conditional-this check failed its refusal test. Removing nominal argument or constraint checks admitted structural lookalikes and failed refusal tests. Removing generic factories' concrete mapping failed the projected-layout test. All mutants compiled; compiler warnings are not the evidence.

## Continuation mutant runs

Each run exited 1 at the named oracle or assertion, then restored the implementation. These are observations from actual runs, not predictions about possible failures.

| Mutant | What caught it |
| --- | --- |
| `static_escape` | Region-plan assertion and ASan use-after-free after an override stored an arena argument. |
| `split_consumption` | Shared-ownership assertion and LeakSanitizer. |
| `disable_regions` | Hierarchy region-plan assertion. |
| `disable_reuse` | Hierarchy spread-reuse assertion. |
| `global_targets` | UBSan null access and Node exit parity after a static-only global-move decision. |
| `collapsed_layouts` | Missing object layout in the monomorphization assertion. |
| `erased_identity` | Node/native instanceof stdout differed while both exited 0. |
| `unmapped_bases` | The positive generic layout test failed to lower a substituted base. |
| `ready_before_return` | A retry after base failure became an uncaught duplicate-binding error instead of Node's successful construction. |
| `early_fields` | Initializer-order stdout differed, with all backends exiting 0. |
| `missing_super_return` | Missing ReferenceError changed stdout. |
| `second_super` | The first object's value was overwritten and derived initialization repeated; stdout differed. |
| `conditional_this` | The pre-super capture refusal test got no error. |
| `nominal_arguments` | The generic class-view refusal test got no error. |
| `nominal_constraints` | The generic constraint refusal test got no error. |
| `factory_mapper` | The projected generic factory layout test could no longer lower the concrete field type. |

## Current limits

Super calls used as values or inside deferred functions report `NotYet`; use statements in each branch. Generic class arguments must have known native representations. Explicit replacement constructor returns, computed base expressions, static members, declare/abstract fields, overloaded overrides and changed native override representations or parameter counts report `NotYet`. Override signatures are compared in their lowered native forms, including defaulted and optional parameters and void results. A changed parameter representation names the parameter and asks for the base method's parameter form; no adaptation thunk is generated. Identical defaults and parameter widening with the same representation remain supported. Calls through a union of class types require `instanceof` narrowing first. Structural interface method dispatch remains a separate stage-0 gap.

Base constructors cannot publish this or call methods while derived fields are uninitialized. Field initializers may read earlier fields and inherited fields after super; future-field reads and arbitrary this escapes are refused. Lexically early captures of this are conservatively refused.

Virtual return allocation currently uses the heap calling convention. Region escape summaries must consider every dynamic target, and reuse must give every implementation the same parameter ownership convention. Weak expiration is tested natively; source Node erases the Weak annotation, so it is not an expiration oracle.

## Integration files

Files outside the original new lowering/native files and `internal/lower/class.go`, across the initial unit and continuation:

| File | Reason |
| --- | --- |
| `internal/flow/build.go` | Include every virtual target's exceptional control flow. |
| `internal/lower/cycles.go` | Preserve private field spelling in cycle diagnostics. |
| `internal/lower/exceptions.go` | Propagate virtual exceptions through all possible targets. |
| `internal/lower/fresh.go` | Match private symbols to qualified write slots. |
| `internal/lower/lower_test.go` | Remove the obsolete ordinary-instanceof gap expectation. |
| `internal/lower/object.go` | Qualify private fields by declaring class. |
| `internal/lower/refusals.go` | Apply override and nominal-view checks before lowering. |
| `internal/native/borrow.go` | Visit ancestry operands while classifying the test as pure. |
| `internal/native/runtime/heap.c` | Use the class destructor chain with iterative teardown. |
| `internal/native/runtime/object.c` | Give plain and copied objects no class identity. |
| `internal/native/runtime/region.c` | Release region children through the shared class helper. |
| `internal/oracle/testdata/class_identity.a` | Test spread identity loss and instanceof side effects. |
| `internal/oracle/testdata/class_inheritance.a` | Test dispatch, ancestry, abstracts, Weak and private fields. |
| `internal/oracle/testdata/class_inheritance_exceptions.a` | Test virtual throws and failed-construction cleanup. |
| `internal/oracle/testdata/class_inheritance_order.a` | Make constructor and field order observable. |
| `internal/oracle/testdata/class_layouts.a` | Use interfaces for structural literals instead of nominal fakes. |
| `internal/native/region.go` | Join virtual escapes and recognize class allocators. |
| `internal/native/reuse.go` | Normalize virtual parameter ownership and follow all targets for moves. |
| `internal/native/emit.go` | Small hooks for region calls, field ownership and named errors. |
| `internal/native/runtime/adamic.h` | Add erased class identity to descriptors. |
| `internal/ir/ir.go` | Carry erased identity and constructor ReferenceError names. |
| `internal/javascript/javascript.go` | Match erased identity and named errors in the second backend. |
| `internal/fresh/fresh.go` | Analyze virtual calls, ancestry operands and error-name operands. |
| `internal/lower/lower.go` | Small conditional-super hook and generic factory cache field. |
| `internal/lower/expression.go` | Lower class instanceof and resolve concrete generic representations. |
| `internal/oracle/class_inheritance_test.go` | Register the new differential fixtures. |
| `internal/oracle/testdata/class_inheritance_memory.a` | Exercise regions, reuse and dynamic ownership. |
| `internal/oracle/testdata/class_inheritance_generic.a` | Exercise generic hierarchies, identities and factories. |
| `internal/oracle/testdata/class_inheritance_conditional.a` | Exercise branching construction and exceptional binding. |
| `internal/oracle/counts.md` | Record every fixture's measured memory costs. |
| `docs/inheritance.md` | Make the implementation and verification report available in the repository. |

The generic factory mapper currently reads the checker's private signature mapper through a checked reflection bridge, alongside the existing checker bridge in `instantiate.go`. A shim accessor would remove this dependency. Unsupported mappings are diagnosed instead of silently choosing a layout.

## Reproducing the continuation

Source `/workspace/adamic-tools/env.sh` for this cloud setup. It installed Go 1.27.1, clang 20.1.8 and Node 24.19.0; `nproc` reported 5 with a four-core quota. Setup timings were Go 0s, clang 1s, Node 1s, submodules 1s, warm cache 16s, total 16s. Tests write directly to logs.

```sh
gofmt -l cmd internal
go vet ./...
go test -count=1 ./internal/lower ./internal/oracle \
    -run 'TestInheritance|TestNativeAgreesWithNode/internal/oracle/testdata/(class_|generic_functions)' \
    > /tmp/inheritance-focused.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle \
    -run TestCountsAreRecorded -args -update-counts \
    > /tmp/inheritance-counts.log 2>&1
go test -count=1 -timeout 30m ./... > /tmp/inheritance-gate.log 2>&1
```

The final counts update passed in 99.847s. Comparing all 173 pre-branch rows printed `regressions: []`: allocations, frees, retains, releases and peak never increased; region counts never decreased. New fixture measurements:

| Fixture | Allocations | Heap frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Hierarchy memory | 48 | 42 | 22 | 60 | 11 | 6 |
| Generic hierarchy and factories | 89 | 89 | 104 | 167 | 39 | 0 |
| Conditional constructors | 164 | 164 | 148 | 251 | 35 | 0 |

The generic row increased from its earlier version because the fixture gained six constructed objects, generic factories and additional output. It did not exist on pre-branch main; existing input rows remain unchanged or improve.

Cohere formatted the three new fixtures through equivalent `.ts` copies with `--format-only --no-cache` (exit 0, three files formatted). Full Cohere lint, performance comparisons and cross-platform ABI behavior are not claimed. `internal/native/native.go` and `internal/oracle/oracle_test.go` were not edited. No pull request was opened.

The final `gofmt -l cmd internal` and `go vet ./...` produced no output (exit 0). The focused factory oracle passed (lower 1.015s, oracle 17.260s), and the expanded lowering refusal/layout tests passed in 0.523s. The full final gate exited 0, including the recorded-count check and all sanitizer oracles:

```text
?   	github.com/system-inc/adamic/bench	[no test files]
?   	github.com/system-inc/adamic/cmd/adamic	[no test files]
?   	github.com/system-inc/adamic/cmd/adamic-fuzz	[no test files]
ok  	github.com/system-inc/adamic/internal/flow	153.203s
ok  	github.com/system-inc/adamic/internal/fresh	33.114s
ok  	github.com/system-inc/adamic/internal/fuzz	11.992s
?   	github.com/system-inc/adamic/internal/ir	[no test files]
?   	github.com/system-inc/adamic/internal/javascript	[no test files]
ok  	github.com/system-inc/adamic/internal/load	1.280s
ok  	github.com/system-inc/adamic/internal/lower	8.155s
ok  	github.com/system-inc/adamic/internal/native	267.845s
ok  	github.com/system-inc/adamic/internal/oracle	690.283s
ok  	github.com/system-inc/adamic/stage1/cohere/formatfiles	118.354s
ok  	github.com/system-inc/adamic/stage1/cohere/gitignore	185.082s
ok  	github.com/system-inc/adamic/stage1/cohere/graphql	168.140s
ok  	github.com/system-inc/adamic/stage1/cohere/mediaquery	99.830s
ok  	github.com/system-inc/adamic/stage1/cohere/suppression	117.854s
ok  	github.com/system-inc/adamic/stage1/cohere/values	101.486s
```

The initial unit is `b3b50ee`; the continuation pushed optimizer support as `0395573` and generic hierarchy support as `281ee9d`, each after its green checks. The following commit completes conditional construction and the final generic factory and nominal-argument checks.

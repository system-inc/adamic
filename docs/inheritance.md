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

The initial implementation passed `go vet ./...` and `go test -count=1 -timeout 30m ./...`. Cohere formatting passed on equivalent `.ts` copies; a full Cohere lint run had findings and is not claimed green.

## Regions and reuse

Inheritance does not disable region allocation or Perceus reuse. A region escape summary joins every virtual implementation: if any stores an argument, the call cannot receive that argument in a temporary region. Class allocators may allocate in a region when their constructed object does not escape initialization. Virtual return values currently keep the heap calling convention.

Virtual methods share a parameter ownership convention. If one implementation consumes a parameter for reuse, all implementations receive an owned count and release it. The global-move analysis follows every virtual target, preventing a move when an override can observe the global.

The continuation compared all 173 existing counted rows with pre-inheritance main (`fe3b9f2`). No allocation, retain, release or peak count increased, and no region count fell. Some frees became region teardown instead. The hierarchy memory fixture records 48 allocations, 42 heap frees, 22 retains, 60 releases, peak 11 and six values in regions. It exercises both reused spreads and class allocation in regions.

Mutants that ignored escaping overrides triggered ASan on an object stored beyond its temporary region. A mutant that omitted the shared consumption convention leaked under LeakSanitizer. Following only the static target for a global move produced a null access caught by UBSan and Node exit parity. Restoring either blanket optimizer guard failed the plan test. Native and lowering package tests, the filtered oracle and the complete counts update verify this step; the full suite is not rerun for this optimization step.

Outside the original lowering/native territory, this continuation updates `internal/oracle/class_inheritance_test.go` to register the memory fixture, `internal/oracle/testdata/class_inheritance_memory.a` to exercise dynamic ownership, and `internal/oracle/counts.md` to record measurements. `internal/native/emit.go` has small integration changes for joined region arguments and constructor allocation, and moves the final field-store temporary rather than retaining and releasing it. `docs/inheritance.md` is the reader-facing report.

## Current limits

Generic inheritance and conditional or repeated super are pending extensions. Explicit replacement constructor returns, computed base expressions, static members, declare/abstract fields, overloaded overrides and changed native override representations or parameter counts report `NotYet`. Calls through a union of class types require `instanceof` narrowing first. Structural interface method dispatch remains a separate stage-0 gap.

Base constructors cannot publish this or call methods while derived fields are uninitialized. Field initializers may read earlier fields and inherited fields after super; future-field reads and arbitrary this escapes are refused. Lexically early captures of this are conservatively refused.

Virtual return allocation currently uses the heap calling convention. Region escape summaries must consider every dynamic target, and reuse must give every implementation the same parameter ownership convention. Weak expiration is tested natively; source Node erases the Weak annotation, so it is not an expiration oracle.

# Class features

Task zek5q21, on codex/class-features, builds on inheritance tip 7b30af6. Instance accessors, private members and readonly fields were added in 7860a06. The next step adds static fields, methods, accessors and blocks, including private statics.

## Implemented

Instance getters and setters use per-class descriptors, searched from the dynamic class toward its base. Overrides replace an entire descriptor, as JavaScript does. Super property reads and writes call the base accessor with the current receiver. Property reads and writes become IR calls, including through structural views, so borrowing and exception analyses see their targets. Generated setter fallback writes preserve the original write-site metadata.

Literal accessors are closures with an explicit receiver, stored privately in their object. They support this, lexical captures, enumeration and spread. Object.keys enumerates public own properties, with numeric index names first in ascending order. Spread reads enumerable accessors in that order and produces data properties. It drops private storage and accessor closures, preventing plain-object reuse of those layouts. Its receiver remains owned across callbacks, and freshness conservatively accounts for callback escape.

Private fields and methods are qualified by their declaring class. Base and derived classes can use the same private spelling without sharing storage or dispatch. Enumeration and spread exclude private fields. TypeScript rejects private access outside the declaring scope. The fixture includes an inherited private Weak parent reference.

Readonly enforcement remains TypeScript's checker rule, not runtime freezing. Writes outside the declaring constructor are rejected, including ordinary methods and derived constructors. Repeated assignments in the declaring constructor remain valid. Objects held by readonly fields can still have mutable contents.

Override checks reject narrower setter parameters, unsafe getter results and overrides masking an inherited descriptor half, with a repair in the diagnostic. A narrowed getter reread is rejected when it can return a different value: read it once into a local and narrow that local. The cycle finder follows literal accessor capture cells and excludes computed results from physical fields.

Cohere's CSS values parser now uses private identifiers for its methods. Its private-method gap is marked closed and exercised against Node, sanitizers and the leak check.

## Static members

Constructor objects have their own class descriptors and method tables. Instance generic type arguments do not duplicate static storage. Static methods and accessors are installed before initialization; field initializers and blocks execute in source order, even if the class is never constructed. `this` is the current constructor object. Virtual calls through `typeof Base` keep the derived receiver; `super` selects the base implementation with that receiver. Super data reads search the base constructor, while writes create an own property on the current receiver.

Inherited data remains live until an own write shadows it. Presence slots distinguish inherited data from shadowing data, including numeric zero and undefined. Own keys follow the order in which properties are created, and inherited fields, methods, accessors and internal storage are excluded. Each derived constructor owns its parent constructor; the normal derived-to-base destructor chain releases its own values and that parent once.

A class body's inner name refers to its constructor during initialization. The outer declaration binding is published after initialization finishes. Constructor lookup precedes evaluating new arguments and checks declaration readiness. Outside helpers cannot read that binding during its temporal dead zone. TypeScript rejects rebinding class declarations. Use a separate variable for constructor values. Override checks apply to both instance and static members. The cycle finder follows static fields and the strong constructor-parent edge, including constructor values held through construct-signature interfaces. It excludes the checker's synthetic prototype property.

Initializers cannot treat a future field as its nonnullable declared type. Direct reads, static methods, accessors and module helpers are examined; unknown indirect calls and callbacks are diagnosed until initialization finishes. This is conservative: JavaScript permits reading undefined where the source annotation claims a string or number, while Adamic requires that annotation to stay true.

Private static fields, methods and accessors use an exact declaring-class brand. A derived constructor cannot access its base's private members through an inherited method's this. Brand checks are IR calls with exception edges; private writes evaluate their RHS before checking the brand. Qualified storage stays out of enumeration, and a derived private Weak reference to its parent constructor does not count toward a cycle.

## Current limits

- Computed and quoted static names, implicit constructor properties such as prototype/name/length, ambient static declarations and construction through constructor values report NotYet.
- Structural method signatures in programs with statics are conservatively diagnosed; use typeof the declaring class for static method views.
- Spreads in programs using static constructor objects are conservatively diagnosed until spread can materialize their changing own shapes.
- Uninitialized nonnullable static fields are diagnosed. Give them an initializer; indirect initialization calls need a declaration the safety check can examine.
- Optional accessor reads, accessor literals containing spreads, computed names and ordinary methods inside accessor literals report NotYet.
- Accessors sharing a name need compatible native representations; this conservatively excludes some generic instantiations and overrides.
- Reading setter-only descriptors and writing getter-only descriptors report NotYet. Conservative dispatch can also reject unrelated same-name members.
- Spreading literal getters that can throw or fail through a runtime library operation is diagnosed until callbacks have explicit exception edges. Setter-only literal spread is diagnosed too.
- Getters without an adamic_value representation and setters requiring two native words are diagnosed. Literal property names beginning #accessor: are reserved for closure storage and diagnosed.

## Verification

Setup succeeded with Go 1.27.1, clang 20.1.8 and Node 24.19.0. Timing lines: go 0s, clang 0s, node 0s, submodules 0s, cache 13s, total 13s. nproc printed 5. The environment file is /workspace/adamic-tools/env.sh.

The fixtures class_features_private.a and class_features_accessors.a compare source Node, generated JavaScript, native release and ASan/UBSan, and require a clean leak check. They cover overrides and super, side effects, captures, literal this, numeric enumeration, spread, exceptions, receiver replacement during compound assignment, private method separation, private enumeration, readonly constructor assignments and a private Weak back-reference.

The initial full gate passed the oracle suite in 854.197s and native tests in 290.866s. It found missing generated write-site metadata and the newly closed cohere gap. Both were repaired and their tests rerun. Focused ownership verification passed: lower 1.409s, oracle 6.693s. Final package verification passed: lower 12.286s, native 148.021s, fresh 39.550s and cohere values 126.530s. The latest focused oracle passed in 6.328s. Vet, formatting and git diff checks were clean. A combined invocation mistakenly passed the oracle counts flag to other packages; those binaries rejected the flag. The corrected package command above passed. Its oracle portion separately passed fixture verification and the complete counts update in 208.722s.

Reproduce with output in logs:

    source /workspace/adamic-tools/env.sh
    bash cloud/setup.sh > /tmp/adamic-class-features-setup.log 2>&1
    go vet ./... > /tmp/adamic-class-features-vet.log 2>&1
    gofmt -l cmd internal > /tmp/adamic-class-features-format.log 2>&1
    go test -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/fresh ./stage1/cohere/values > /tmp/adamic-class-features-core-packages.log 2>&1
    go test -count=1 ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/class_features' > /tmp/adamic-class-features-ownership.log 2>&1
    go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/adamic-class-features-counts-final.log 2>&1

Compare counts against the 180 rows at the inheritance tip: allocations, frees, retains, releases and peak must never increase; region counts must never decrease. The final comparisons printed: Compared 180 pre-branch rows; regressions: [] and Compared 173 pre-branch rows; regressions: []. The two new rows are below; columns are allocations, frees, retains, releases, peak and values in regions.

| Fixture | Allocations | Frees | Retains | Releases | Peak | In regions |
|---|---:|---:|---:|---:|---:|---:|
| class_features_private.a | 47 | 47 | 37 | 65 | 15 | 0 |
| class_features_accessors.a | 67 | 67 | 53 | 110 | 21 | 0 |

## Mutants

Each mutation changed real lowering or runtime code, was run, then restored. These failures did not depend on compiler warnings.

| Mutant | What caught it |
|---|---|
| Getter replaced by raw slot read | Node/native stdout: 2.5 instead of 6, getter side effects missing. A reduced fixture kept both native executions at exit 0; release and sanitized output also differed for the misread closure slot. |
| Setter skipped | Node/native stdout: 2 instead of 6, setter trace and setter exception missing. |
| Private fields in public shape | Node/native stdout: qualified private names appeared in enumeration and spread. |
| Derived private method redirected to base | Both backends: privateone/privateone instead of privateone/child1. |
| Readonly diagnostic TS2540 discarded | TestClassFeaturesReadonlyChecker accepted an illegal write and failed. The loader mutation was temporary. |
| Setter contravariance check skipped | TestClassFeaturesAccessorRefusals accepted a narrower setter and failed. |
| Capture-cell edges omitted | TestClassFeaturesAccessorCaptureCycle accepted a reference-count cycle and failed. |
| Narrowed getter check skipped | TestClassFeaturesNarrowedAccessor accepted a changing getter reread and failed. |
| Static diagnostic skipped | TestClassFeaturesStaticDeclarationsAreDiagnosed accepted dropped initialization and failed. This proves the diagnostic, not static support. |
| Setter write-site set to zero | TestEveryWriteIsRecordedAndKnown found unrecorded writes and failed. |
| Numeric keys in insertion order | Node/native stdout: 10,2,label instead of 2,10,label. |

## Files outside the original territory

The original territory is new files under internal/lower and internal/native, plus internal/lower/class.go. Additional files:

| File | Why |
|---|---|
| internal/ir/ir.go | Carry visibility, descriptors, accessor calls and enumeration. |
| internal/javascript/javascript.go | Match dispatch, hidden storage and literal enumeration. |
| internal/fresh/fresh.go | Visit new expressions and account for spread callback escape. |
| internal/lower/lower.go | Small preparation/finalization hooks and static diagnostic. |
| internal/lower/object.go | Literal/property hooks, narrowing, enumeration and spread. |
| internal/lower/refusals.go | Remove the blanket accessor refusal. |
| internal/lower/lower_test.go | Remove its obsolete getter refusal expectation. |
| internal/lower/cycles.go | Follow captures and distinguish physical fields. |
| internal/native/emit.go | Small shape/call/exception hooks; own spread receivers across callbacks. |
| internal/native/reuse.go | Prevent reuse when spread drops private storage or closures. |
| internal/native/runtime/adamic.h | Declare descriptors and helpers. |
| internal/native/runtime/object.c | Copy public fields and materialize accessor results in order. |
| internal/oracle/class_inheritance_test.go | Register differential fixtures. |
| internal/oracle/testdata/class_features_private.a | Hold private/readonly behavior and ownership to Node. |
| internal/oracle/testdata/class_features_accessors.a | Hold accessor behavior and ownership to Node. |
| internal/oracle/counts.md | Record measured memory costs. |
| stage1/cohere/values/gaps_test.go | Mark the private-method gap closed and exercise it. |
| stage1/cohere/values/GAPS.md | Document closure and removed workaround. |
| stage1/cohere/values/parser.ts | Replace private-method workaround and its calls with private identifiers. |
| docs/class-features.md | Publish implementation, verification and remaining scope. |

## Static verification

The static fixture covers three levels, static fields and blocks interleaved through observable side effects, static field and block this, static super methods/accessors/data, dispatch through typeof Base, live inherited data, shadow writes, enumeration, unused classes, empty ancestors, shared generic statics, class inner names, inherited reference reads whose RHS replaces the base field, and caught exceptions with allocated strings. The private static fixture covers exact brands, inherited private failures, RHS-before-brand order, private enumeration and a Weak parent constructor. It runs against source Node, generated JavaScript, native release and ASan/UBSan with leak checks.

Setup: Go 0s, clang 0s, Node 0s, submodules 1s, build cache 31s, total 31s; nproc 5. The first package run passed lower in 7.355s, native in 61.331s and fresh in 18.686s. The initial full counts run passed in 133.599s and compared all 182 existing rows with regressions: []. The full gate later passed every other package and failed its count check because the static fixture grew while the table still recorded an earlier version: measured 105/105/66/157/28/0, recorded 101/101/61/148/28/0. The oracle package took 916.826s and native 357.800s. Final verification refreshes the full table and runs the touched packages plus filtered static oracles, following the cloud gate fallback for a slow machine.

Static mutants are restored after each run. Wrong field values, static method dispatch, getter-as-data, skipped setter, skipped blocks and wrong inherited reads changed oracle stdout. Reordering a block and the following field changed abicdef to aibcdef with normal exit 0 on both backends. Skipping the base destructor leaked 133 bytes in two allocations under ASan. Omitting future-read checks and treating constructor types as code accepted unsafe programs in the lowering tests.

Additional integration files in this step:

| File | Why |
|---|---|
| internal/ir/ir.go | Static descriptors, presence slots and own-definition writes. |
| internal/javascript/javascript.go | Constructor objects, prototypes and own field definitions. |
| internal/lower/lower.go | Register constructor storage and execute static declarations. |
| internal/lower/expression.go | Resolve the inner class binding separately from the outer declaration. |
| internal/lower/object.go | Diagnose implicit constructor properties. |
| internal/lower/class_inheritance.go | Share override checks with static members and require constructor ancestry. |
| internal/lower/class_accessors.go | Preserve own definitions, static super data and conservative spread safety. |
| internal/lower/class_features.go | Remove the obsolete blanket static diagnostic. |
| internal/lower/class_features_test.go | Exercise static effects and soundness refusals. |
| internal/lower/cycles.go | Follow constructor fields and strong parent edges. |
| internal/native/class_inheritance.go | Emit static descriptor metadata. |
| internal/native/emit.go | Write own static slots and emit constructor typeof. |
| internal/native/runtime/adamic.h | Declare static lookup, presence metadata and typeof helpers. |
| internal/native/runtime/class_features.c | Enumerate current own static keys. |
| internal/oracle/class_inheritance_test.go | Register the static oracle. |
| internal/oracle/testdata/class_features_static.a | Hold static semantics and ownership to Node. |
| internal/oracle/testdata/class_features_static_private.a | Hold private brands, write order and Weak ownership to Node. |
| internal/oracle/counts.md | Record measured costs. |
| docs/class-features.md | Publish the static implementation and its limits. |

New implementation files: internal/lower/class_static.go, internal/lower/class_static_private.go and internal/native/runtime/class_static.c.

The additional constructor-binding and private mutants bypass outside-helper temporal dead zone checks, skip static override soundness, replace exact private brands with ancestry, expose private static fields, check private write brands before the RHS, read the outer class binding inside its body, skip the constructor-parent edge and ignore constructors held behind construct-signature interfaces. Each is held by its dedicated lowering test or a Node oracle. An initial inner-binding mutation failed at compilation because it left an unused Go local; that attempt does not count. The corrected mutation preserves the local and must fail through the oracle.

| Static mutant | What caught it |
|---|---|
| Replace field initializers with wrong strings | Both oracle backends changed values and initialization trace, with exit 0. |
| Dispatch static methods by the declared type | The typeof Base parameter printed the base label instead of the leaf label. |
| Read a static getter as data | Native printed 2/3/3 instead of 2/4/8 and lost getter effects. |
| Skip a static setter | Native values and setter trace differed from Node. |
| Skip static blocks | Oracle output lost block effects and the unused-class block. |
| Ignore own shadow fields when reading inherited data | Native shadow values differed from Node. |
| Skip the base destructor | ASan reported 133 bytes leaked in two allocations. |
| Omit future-field initialization checks | TestClassFeaturesStaticSoundness accepted unsafe direct/helper/alias reads. |
| Treat constructor values as fieldless code | TestClassFeaturesStaticSoundness accepted public and private constructor cycles. |
| Swap a block and its following field | Both backends printed aibcdef instead of abicdef, with exit 0. |
| Ignore outside-helper temporal dead zones | TestClassFeaturesStaticSoundness accepted a helper reading the unfinished class declaration. |
| Skip static override checks | TestClassFeaturesStaticSoundness accepted a narrower method parameter and mutable field. |
| Use ancestry for private static brands | The private oracle hit a null string read under UBSan instead of catching the required TypeError. |
| Check private write brands before the RHS | Both backends printed s3/abs instead of s5/abwss, with exit 0. |
| Enumerate private static fields | Native Object.keys exposed #value and #parent storage names. |
| Resolve inner class names to the outer binding | Both backends panicked on StaticName's temporal dead zone while Node succeeded. |
| Omit the constructor-parent edge | TestClassFeaturesStaticParentCycle accepted a cycle hidden by different constructor signatures. |
| Ignore constructors behind interfaces | TestClassFeaturesStaticInterfaceCycle accepted the same strong cycle through a construct-signature view. |

Final verification commands, with each test's output redirected to its log:

    source /workspace/adamic-tools/env.sh
    go test -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/fresh > /tmp/adamic-static-packages-final.log 2>&1
    go test -count=1 -timeout 30m ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/class_features' > /tmp/adamic-static-oracles-final.log 2>&1
    go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts > /tmp/adamic-static-counts-green.log 2>&1
    go vet ./... > /tmp/adamic-static-vet-final.log 2>&1
    gofmt -l cmd internal > /tmp/adamic-static-format-final.log 2>&1

Final results: class-feature Node/release/ASan/UBSan/leak oracles passed in 19.235s; counts refresh passed in 135.501s. Native passed in 91.593s and fresh in 21.733s. Lowering initially found that the computed-base diagnostic no longer contained its expected wording. Restoring that wording gave a green full lower package in 2.549s. Vet, formatting and git diff checks were clean. The counts comparisons printed Compared 182 pre-branch rows; regressions: [] and Compared 173 pre-branch rows; regressions: []. The table now contains 184 rows.

| New fixture | Allocations | Frees | Retains | Releases | Peak | In regions |
|---|---:|---:|---:|---:|---:|---:|
| class_features_static.a | 110 | 110 | 84 | 191 | 28 | 0 |
| class_features_static_private.a | 42 | 42 | 71 | 112 | 12 | 0 |

The diagnostic-only repair was verified with:

    go test -count=1 -timeout 30m ./internal/lower > /tmp/adamic-static-lower-green.log 2>&1

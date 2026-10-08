# Step 22: OrdinaryToPrimitive

Runtime unit `runtime-step22-toprimitive`, roadmap step 22 (#nzbprw9), October 8, 2026. Branch
`runtime/step22-to-primitive`, based on area/runtime
`9bc8201f1c97c4c18a03d263a40ece933ffa752e`.

## Decision and algorithm

system_adamic's October 8 points 5 and 6 require OrdinaryToPrimitive for
object template substitutions and addition, and catchable JavaScript errors.
Symbol.toPrimitive remains NotYet. Adamic's soundness panics remain terminal.

[ECMA-262, OrdinaryToPrimitive](https://tc39.es/ecma262/#sec-ordinarytoprimitive),
verified against [spec source 6db9135](https://github.com/tc39/ecma262/blob/6db9135b3d77fd38a81a5a1f1bb6ed77fcca9066/spec.html):

1. String hint: method names are `toString`, then `valueOf`.
2. Number hint: method names are `valueOf`, then `toString`.
3. For each name, perform Get, propagating its exception. If callable, call
   with the original receiver as `this` and no arguments, propagating its
   exception. Return its result immediately if that result is not an Object.
4. If neither returns a primitive, throw TypeError.

OrdinaryToPrimitive itself has no default hint. ToPrimitive maps the default
hint used by `+` to number for ordinary objects; the C API does that mapping.
Templates use string. A non-callable member is skipped, including one that
shadows a callable prototype member. Lookup of the second member happens only
if needed and after the first lookup/call has completed. Functions and arrays
returned by either method are Objects, too.

Observed on Node v24.19.0, with both methods returning objects, both absent,
and both returning functions: `error.name` is `TypeError`, and `error.message`
is exactly `Cannot convert object to primitive value`. The ECMA algorithm
specifies the error type, not this engine-specific wording.

## Compiler C contract

Declarations are in `internal/native/runtime/adamic.h`; implementation is in
`internal/native/runtime/exceptions.c`:

```c
adamic_primitive adamic_ordinary_to_primitive(
    adamic_heap *receiver,
    enum adamic_primitive_hint hint,
    adamic_primitive_get get);
```

The existing `adamic_method` and closure ABIs return untagged `adamic_value`.
Shapes identify reference slots but do not distinguish a string result from
an object result, nor carry a method's result type. Inspecting arbitrary
untagged number bits as a heap pointer is unsafe. The compiler therefore
supplies a typed adapter; this unit does not change either existing ABI.

- `receiver` is borrowed and must remain strongly held until the helper
  returns, even if a getter or method changes its original storage location.
  It may be a class instance, ordinary object, array, or other supported
  ECMAScript Object representation.
- `hint` is `adamic_hint_string` for templates and `adamic_hint_default` for
  addition. Explicit number hint is also supported. An invalid hint is a
  terminal compiler-bug panic.
- `get(receiver, name)` performs precisely that property's Get and IsCallable,
  including supported inherited members, overriding fields, and getters. It
  must not prefetch either property. Missing/non-callable returns `call = NULL`.
  It must not use `adamic_object_callee` blindly: that routine requires an
  existing callable and panics on a missing method.
- The returned method contains a call thunk and `owner`, one owned reference
  keeping the retrieved callable or adapter context alive. NULL is allowed
  for a static method thunk. The helper releases owner after the call, after
  a skipped non-callable, and after an exceptional lookup. The call thunk
  receives the original receiver and owner and invokes with zero arguments.
- A normal call returns a tagged `adamic_primitive`: undefined, null, boolean,
  number, string, or object. Null and undefined have different tags despite
  sharing NULL storage elsewhere. String/object results transfer one owned
  reference to the helper. Numbers/booleans are unboxed. Every non-primitive
  result, including arrays, functions and boxed JavaScript objects, uses
  `adamic_primitive_object`; Adamic's internal boxes for primitive union members
  must be unboxed/tagged according to their language value.
- The helper transfers an accepted primitive result to its caller and releases
  every rejected object result before trying the next name. The caller owns
  an accepted string and releases it after use.
- Get/Call failures set the existing thread-local `adamic_thrown`. An exceptional
  call returns the zero undefined result with no owned value. An exceptional
  Get may return an owner, which the helper releases. The helper stops at once
  and preserves that exact pending Error. Entry with an already pending Error
  performs no conversion and returns zero.
- Exhaustion creates the same Error object layout as `adamic_error_new`, with
  name `TypeError` and the observed Node message, and transfers its one count
  to `adamic_thrown`. It does not call panic. The caller must test the pending
  word before using the dummy result and route through ordinary scope/temporary/
  region cleanup and the nearest handler. Catch takes and clears the pending
  word; uncaught reporting uses existing `adamic_uncaught`.

The lowering must mark this operation and reachable callers MayThrow. It must
preserve operand evaluation order: evaluate both addition operands first, then
convert the left and then right, stopping on an exception. Conversion does not
implement addition's subsequent primitive string/numeric choice, ToString, or
ToNumeric. Templates apply ToString to the returned primitive.

Symbol.toPrimitive lookup/call and Symbol/BigInt return tags are not implemented.
Lowering must issue NotYet for a conversion needing those, or an unsupported
builtin/exotic adapter (including Date's default-hint override); it must not
silently bypass their semantics. No emission file or lowering file was edited.
At this base, `internal/lower/expression.go:842` and `:850` still reject object
and object-union templates. End-to-end object interpolation is therefore NotYet.

## TypeScript compiler research

Corpus: TypeScript 6.0.3,
[050880ce](https://github.com/microsoft/TypeScript/tree/050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler),
the same 77 compiler files used in batch 8. The original checkout's compiler
project options were used, including strict mode and strictBindCallApply=false;
upstream's diagnostic-map generator was run before checking. TypeScript 6.0.3
and Node declarations, including source-map-support, produced zero diagnostics.

The AST/checker audit visited all 77 files recursively. It counted **338
untagged template expressions, 837 binary additions, and 333 += expressions**:
1,508 expressions and 2,868 coercion operands. Tagged templates do not perform
these conversions and were excluded. String-branded intersections (`Path`,
`__String`) remain primitives and were excluded from the object inventory.

There are **three concrete object/array operand sites**, all templates:

| Site | Operand type | Shape and conversion |
|---|---|---|
| src/compiler/debug.ts:383:135 | string[] or undefined | Declaration kind names; Array.prototype.toString/join when present |
| src/compiler/semver.ts:450:37 | Version | Comparator operand; Version.toString |
| src/compiler/commandLineParser.ts:2958:119 | primitive or primitive[] | PresetValue in the failed compiler-option lookup diagnostic; array branch uses join |

No statically concrete object `+` or `+=` operand was found. **18 additional
operands** are generic, any or unknown candidates; they are not observed runtime
object invocations. The complete file:line:column table follows. Numeric-array
and string-producing any sites are retained to make the audit's boundary visible.

| Site | Operation | Type / finding |
|---|---|---|
| core.ts:1786:59 | template | TIn or undefined, failed cast diagnostic; may receive an object |
| core.ts:2221:20 | + left | any; edit-distance previous[] is filled with numbers |
| core.ts:2222:20 | + left | any; same numeric matrix |
| core.ts:2225:39 | + left | any; previous[] numeric matrix |
| core.ts:2225:67 | + left | any; current[] numeric matrix |
| debug.ts:226:30 | template | T, assertEqual's left value |
| debug.ts:226:39 | template | T, assertEqual's right value |
| tracing.ts:66:81 | template | any, caught error's message or error itself |
| sys.ts:806:117 | template | any, timerToUpdateChildWatches; reset to undefined at :774, host/callback effects not proven |
| moduleNameResolver.ts:946:21 | + right | unknown; guard excludes typeof object, but functions are also ECMAScript Objects |
| checker.ts:10953:71 | template | any, __debugFlags or flags; intended flags, not proven by the cast |
| sourcemap.ts:315:25 | += right | any, String.fromCharCode.apply result; builtin returns a primitive string |
| watchUtilities.ts:815:120 | template | any, args[0] from watcher callback |
| watchUtilities.ts:815:131 | template | any, args[1] or empty string from watcher callback |
| watchUtilities.ts:830:38 | template | T, watch flags |
| watchUtilities.ts:830:74 | template | string or X, outer watch detail |
| watchUtilities.ts:830:187 | template | X, first nested watch detail |
| watchUtilities.ts:830:202 | template | Y constrained only by presence, second nested watch detail |

Paths in that table are relative to `src/compiler/`. It counts coercion operands,
not executions. No dynamic tsc run or whole-program value-flow proof was made;
untyped host values and generic instantiations can carry additional behavior.
In particular, absence of an explicit Symbol.toPrimitive declaration in these
files is not proof that an external host object has no such method.

Only compiler classes **Version** (`semver.ts:133`) and **VersionRange**
(`semver.ts:232`) declare toString. No compiler class declares valueOf.
The source-map generator's returned object also defines a toString closure at
`sourcemap.ts:83`; its observed compiler uses call it explicitly. Builtin Array,
Object, Error and host prototypes are outside this class declaration count.

[sites.json](../cloud/reports/step22-to-primitive/sites.json) preserves every
candidate operand, exact expression/type, all 77 source SHA256 values, method
inventory and diagnostics. [audit.mjs](../cloud/reports/step22-to-primitive/audit.mjs)
regenerates it; it does not change the compiler source. The generated diagnostic
map is a dependency, not one of the 77 handwritten audited files.

```sh
# In a scratch checkout of the pinned upstream TypeScript commit:
node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json
npm install --prefix /tmp/step22-research typescript@6.0.3 @types/node@26.6.4 @types/source-map-support@0.5.10
# From Adamic:
node cloud/reports/step22-to-primitive/audit.mjs /path/to/TypeScript \
  /tmp/step22-research/node_modules/typescript/lib/typescript.js \
  /tmp/step22-research/node_modules/@types /tmp/step22-sites.json
```

## Evidence and remaining work

`internal/native/testdata/to-primitive/node.a` runs directly on Node as the
external truth. `runtime.c` calls the new contract directly through typed
Get/Call adapters. These are C-contract witnesses, not accepted source programs
or a claim that compiler-generated adapters exist. The Version fixture is a
reduced three-numeric-field shape with no prerelease/build suffix; the declaration
array mirrors debug's string array. Other cases exercise adversarial versions
of those shapes.

Twenty cases cover string/default/number hints, non-callable and absent members,
fallback after object results, getter and method exceptions, lookup after both
method and getter mutation, all six supported return tags, and functions returned
as Objects. Each normal case matches Node stdout and exits zero, under ASan,
UBSan and the shared leak check, with allocation/free counts balanced. The cases
run with sanitizer malloc and the developer-tools slab lane separately.

Mutants are compiled against isolated runtime copies by
`TestOrdinaryToPrimitiveMutants`. Wrong order must disagree with Node on **every
one of the twenty cases**, while still finishing leak-clean. Removing TypeError
must disagree on objects, missing methods and function results, also leak-clean.
Omitting rejected-object release must be caught by LeakSanitizer. Two additional
C protocol controls prove that an invalid hint is terminal and a pending Error
prevents even Get; removing either guard is caught by its own control. Compiler
errors
are not accepted as mutant catches. No mutant changes the checkout.

The first harness run manually called heap_end as well as its registered exit
hook, producing a slab-lane ASan failure at the second shutdown. The harness was
corrected to leave shutdown to the runtime exit hook; this was not a conversion
helper defect. The restored expanded focused run passes in 37.096s. The final
twenty-case plus protocol-guard run, including all five runtime mutants, passes
in 13.093s.

Setup completed successfully: Go ready 0.091s, Node ready 0.104s, clang ready
0.503s, markdown ready 2.100s, submodules ready 4.626s, build ready 180.044s,
cache warm 180.221s, total 180.249s. nproc=5; cgroup cpu.max=400000 100000.
Every toolchain shell sources `/workspace/adamic-tools/env.sh`.

Exact sanitizer compile/link flags (clang 20.1.8):

```text
-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable
-Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter
-Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -pthread
-DADAMIC_COUNT -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all
```

The slab variant additionally uses `-DADAMIC_SLABS`. The final link adds the include directory, binary output, main.c, runtime
archive/link inputs and `-lm`.
These are correctness tests, with no parser instruction-count or speed claim.

Final completed checks, with every command's output written to a log:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/native -count=1 -parallel 4 -timeout 30m > /workspace/scratch/step22-to-primitive/native-final.log 2>&1
# PASS, 545.191s. Production helper and the twenty conversion witnesses unchanged.
go test ./internal/native -run '^TestOrdinaryToPrimitive' -count=1 -v -timeout 15m > /workspace/scratch/step22-to-primitive/protocol-final.log 2>&1
# PASS, 13.093s. Includes the two protocol tests added after the full package started.
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/.*(exceptions|throw|error|class_as_interface)|TestCountsAreRecorded' -count=1 -parallel 4 -timeout 30m > /workspace/scratch/step22-to-primitive/oracle.log 2>&1
# PASS, 106.818s. Includes the complete existing counts table; no rows changed.
go vet ./... > /workspace/scratch/step22-to-primitive/vet.log 2>&1
# PASS, empty log.
go vet ./internal/native > /workspace/scratch/step22-to-primitive/vet-native-final.log 2>&1
# PASS, empty log, after the protocol tests were added.
gofmt -l cmd internal > /workspace/scratch/step22-to-primitive/gofmt-final.log 2>&1
git diff --check > /workspace/scratch/step22-to-primitive/diff-final.log 2>&1
# Both PASS, empty logs.
```

The initial full native run was interrupted by an environment restart and has
no result to claim; the completed rerun above replaces it. The final audit also
regenerated byte-identical evidence with zero diagnostics. Its dependencies are
TypeScript 6.0.3, @types/node 26.6.4, @types/source-map-support 0.5.10 and
undici-types 8.9.0. Node's actual oracle runtime is v24.19.0.

The full repository gate, macOS execution, WASI execution, tsc dynamic conversion
frequencies, compiler lowering and Symbol.toPrimitive were not covered. No
oracle fixture registry or counts row was added: these new cases use the C seam
explicitly allowed by the unit while lowering is unavailable.

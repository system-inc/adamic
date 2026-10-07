# Accessor branch coverage

These observations record d785e87 and coverage commit d6b9675. They are
historical. See [merge.md](merge.md) for reconciliation with main and the
current fixture and counts checks. The inferred mixed array is now refused.

20 programs added to the oracle, 4 source/native differences kept here, and 57
rejected probes kept under unsupported/. Cut coverage/accessors from d785e87.
Compiler code is unchanged. See coverage.md for the case-by-case inventory and
observations.json for exact standalone commands, stdout, stderr and exit codes.

Three differences are the documented narrowing checks, not silent miscompiles.
The inferred mixed array is an accepted program that dispatches through a nominal
method table on a plain object. Its source prints 1 then 2, but native prints only
1 and panics with exit 70. JavaScript backend also prints only 1 and panics.

For repeated_number.a and repeated_boolean.a, accessors.go:189 creates ir.Unwrap
on the second read. For repeated_string.a, accessors.go:194 calls defined, which
wraps the second read in ir.Defined (narrowed.go:66). The branch documents those
checks. They still differ from source Node, so they are kept out of testdata.

The inferred mixed array's likely cause is freshOrWidened's unconditional literal
exemption at internal/lower/invariance.go:497, combined with the missing contextual
type in classViewRefusal. The new accessor-field view check never proves the plain
object has the class layout; accessors.go:171 nevertheless emits virtual dispatch.
This is an inference from the code and the observed failure, not a compiler fix.

## Mutation

Changed one line, internal/lower/accessors.go:231:

```go
value := ir.Expression(ir.NumberConstant{Value: 1})
```

to Value: 2. The new operators program compiled on both backends and failed the
oracle only on stdout: Node ended with 10, native and JavaScript ended with 12.
The exact original bytes were restored in a finally block. mutant.log contains
the failure. The restored program passed in the final uncached gate.

## Toolchain

Source /workspace/adamic-tools/env.sh, the configured ADAMIC_TOOLS directory.
Successful rerun of bash cloud/setup.sh:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (65s)
setup: done in 65s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc: 5. Go 1.27.1, Node v24.19.0, clang 20.1.8.
The first setup attempt overlapped the checkout switch and failed its warm-build
vet phase; it was rerun on the stable branch. The first mutation attempt omitted
the environment source and invoked an unrelated system Go; it restored the code
and was rerun with the configured toolchain. Neither is the reported proof.

## Commands

Run from the repository root, after sourcing the environment:

```bash
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/accessors_coverage_' -count=1 -timeout 10m
python3 notes/accessors/run_observations.py
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
```

The standalone runner executes, for every saved .a program:

```bash
node --disable-warning=ExperimentalWarning oracle/node.mjs <file>
go run ./cmd/adamic build <file> -o /tmp/adamic-accessors/<stem>
/tmp/adamic-accessors/<stem>
go run ./cmd/adamic js <file>
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/adamic-accessors/<stem>.mjs
```

Native and JavaScript runs happen only after a successful build. All 20 fixtures
and all 4 differences built. Rejected probes have no native stdout because no
binary was produced. observations.json records the expanded argv for every run.
Exploratory oracle rounds fixed fixture console signatures and separated unsupported
cases; all final source is represented in observations.json.

The differences were temporarily registered in internal/oracle/accessors_probe_test.go
with lowers=true and checked=false and run with:

```bash
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/notes/accessors/differences/' -count=1 -timeout 10m
```

All four failed as expected, recorded in differences.log. The temporary registration
was removed. Its reported leak checks exit 70 because the mismatching native program
panics, not because LeakSanitizer reported leaked allocations.

With the one-line mutation in place:

```bash
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/accessors_coverage_operators.a' -count=1 -timeout 10m
```

## Rejected probes

The complete Node outputs and build stderr are in observations.json. First build
diagnostic for each probe follows. These are unsupported cases, not output differences.

| Probe | Build diagnostic |
|---|---|
| [abstract.a](unsupported/abstract.a) | adamic: /workspace/adamic/notes/accessors/unsupported/abstract.a:1:20: stage 0 can't lower abstract or ambient accessors (adamic/abstract-accessor) yet |
| [accessor_replaces_field.a](unsupported/accessor_replaces_field.a) | notes/accessors/unsupported/accessor_replaces_field.a:1:57: error TS2611: 'total' is defined as a property in class 'A', but is overridden here in 'B' as an accessor. |
| [added_half.a](unsupported/added_half.a) | adamic: /workspace/adamic/notes/accessors/unsupported/added_half.a:1:67: stage 0 can't lower an override changing the getter/setter descriptor halves (adamic/accessor-descriptor-override) yet |
| [ambient.a](unsupported/ambient.a) | adamic: /workspace/adamic/notes/accessors/unsupported/ambient.a:1:19: stage 0 can't lower abstract or ambient accessors (adamic/abstract-accessor) yet |
| [base_constructor_escape.a](unsupported/base_constructor_escape.a) | adamic: /workspace/adamic/notes/accessors/unsupported/base_constructor_escape.a:1:69: Adamic 0.1 refuses this escaping a base constructor before derived fields are initialized; use this only to read or write initialized base fields; call methods and publish the object after construction |
| [captured_this_result.a](unsupported/captured_this_result.a) | adamic: /workspace/adamic/notes/accessors/unsupported/captured_this_result.a:1:7: Adamic 0.1 refuses 'this', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function this() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| [compound_value.a](unsupported/compound_value.a) | adamic: /workspace/adamic/notes/accessors/unsupported/compound_value.a:1:109: stage 0 can't lower an accessor assignment or update used as a value (adamic/accessor-update-value) yet |
| [computed.a](unsupported/computed.a) | adamic: /workspace/adamic/notes/accessors/unsupported/computed.a:1:11: stage 0 can't lower computed or private accessor names (adamic/accessor-name) yet |
| [constructor_escape.a](unsupported/constructor_escape.a) | adamic: /workspace/adamic/notes/accessors/unsupported/constructor_escape.a:1:89: Adamic 0.1 refuses this escaping a constructor before every field is set (stored, passed, or a method called on it, which could read a field that holds undefined while its type says otherwise); assign every field first, then use this |
| [destructure.a](unsupported/destructure.a) | adamic: /workspace/adamic/notes/accessors/unsupported/destructure.a:1:49: stage 0 can't lower spread, destructuring or indexed access of an accessor-bearing value (adamic/accessor-property-operation) yet |
| [discarded.a](unsupported/discarded.a) | adamic: /workspace/adamic/notes/accessors/unsupported/discarded.a:6:1: stage 0 can't lower a PropertyAccessExpression as a statement yet |
| [erased_accessor.a](unsupported/erased_accessor.a) | adamic: /workspace/adamic/notes/accessors/unsupported/erased_accessor.a:1:84: Adamic 0.1 refuses a view erasing accessor x (adamic/accessor-view-erasure); keep a class view declaring the accessor; erasing it could expose it later as an optional field |
| [erased_base.a](unsupported/erased_base.a) | adamic: /workspace/adamic/notes/accessors/unsupported/erased_base.a:1:98: Adamic 0.1 refuses a view erasing accessor total (adamic/accessor-view-erasure); keep a class view declaring the accessor; erasing it could expose it later as an optional field |
| [field_replaces_accessor.a](unsupported/field_replaces_accessor.a) | notes/accessors/unsupported/field_replaces_accessor.a:1:76: error TS2610: 'total' is defined as an accessor in class 'A', but is overridden here in 'B' as an instance property. |
| [function_view.a](unsupported/function_view.a) | adamic: /workspace/adamic/notes/accessors/unsupported/function_view.a:1:133: Adamic 0.1 refuses an accessor and a field sharing the view of x (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |
| [generic.a](unsupported/generic.a) | adamic: /workspace/adamic/notes/accessors/unsupported/generic.a:1:14: stage 0 can't lower accessors on generic classes (adamic/generic-accessor) yet |
| [generic_function_receiver.a](unsupported/generic_function_receiver.a) | adamic: /workspace/adamic/notes/accessors/unsupported/generic_function_receiver.a:1:97: stage 0 can't lower accessor dispatch through a type parameter (adamic/generic-accessor-receiver) yet |
| [generic_setter.a](unsupported/generic_setter.a) | adamic: /workspace/adamic/notes/accessors/unsupported/generic_setter.a:1:14: stage 0 can't lower accessors on generic classes (adamic/generic-accessor) yet |
| [getter_narrowing.a](unsupported/getter_narrowing.a) | adamic: /workspace/adamic/notes/accessors/unsupported/getter_narrowing.a:1:133: stage 0 can't lower a narrowed accessor result with a different native representation (adamic/accessor-narrowing) yet |
| [getter_only_write.a](unsupported/getter_only_write.a) | notes/accessors/unsupported/getter_only_write.a:1:76: error TS2540: Cannot assign to 'total' because it is a read-only property. |
| [indexed.a](unsupported/indexed.a) | adamic: /workspace/adamic/notes/accessors/unsupported/indexed.a:1:74: stage 0 can't lower spread, destructuring or indexed access of an accessor-bearing value (adamic/accessor-property-operation) yet |
| [initializer_escape.a](unsupported/initializer_escape.a) | adamic: /workspace/adamic/notes/accessors/unsupported/initializer_escape.a:1:50: Adamic 0.1 refuses this in a field initializer before the fields it reads are initialized; declare the field it reads earlier, or initialize it in the constructor after super and all required fields are set |
| [interface_argument.a](unsupported/interface_argument.a) | adamic: /workspace/adamic/notes/accessors/unsupported/interface_argument.a:1:168: Adamic 0.1 refuses an accessor and a field sharing the view of total (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |
| [interface_cast.a](unsupported/interface_cast.a) | adamic: /workspace/adamic/notes/accessors/unsupported/interface_cast.a:1:103: Adamic 0.1 refuses an accessor and a field sharing the view of total (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |
| [interface_descriptor.a](unsupported/interface_descriptor.a) | adamic: /workspace/adamic/notes/accessors/unsupported/interface_descriptor.a:1:18: stage 0 can't lower an accessor declaration outside a class (adamic/accessor-declaration) yet |
| [interface_field.a](unsupported/interface_field.a) | adamic: /workspace/adamic/notes/accessors/unsupported/interface_field.a:1:126: Adamic 0.1 refuses an accessor and a field sharing the view of total (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |
| [interface_getter_parameter.a](unsupported/interface_getter_parameter.a) | adamic: /workspace/adamic/notes/accessors/unsupported/interface_getter_parameter.a:1:115: Adamic 0.1 refuses an accessor through a structural property view (adamic/accessor-field-view); keep the nominal class type; an interface does not prove the virtual slot's layout |
| [interface_inherits_accessor.a](unsupported/interface_inherits_accessor.a) | adamic: /workspace/adamic/notes/accessors/unsupported/interface_inherits_accessor.a:1:91: Adamic 0.1 refuses an accessor and a field sharing the view of x (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |
| [literal.a](unsupported/literal.a) | adamic: /workspace/adamic/notes/accessors/unsupported/literal.a:1:17: stage 0 can't lower object-literal accessors (adamic/object-accessor) yet |
| [method_replaces_accessor.a](unsupported/method_replaces_accessor.a) | notes/accessors/unsupported/method_replaces_accessor.a:1:88: error TS2426: Class 'A' defines instance member accessor 'total', but extended class 'B' defines it as instance member function. |
| [mutable_getter_result_override.a](unsupported/mutable_getter_result_override.a) | adamic: /workspace/adamic/notes/accessors/unsupported/mutable_getter_result_override.a:1:173: Adamic 0.1 refuses an unsound accessor override (adamic/accessor-override); keep the inherited accessor descriptor and its read and write types |
| [narrowed_to_undefined.a](unsupported/narrowed_to_undefined.a) | adamic: /workspace/adamic/notes/accessors/unsupported/narrowed_to_undefined.a:1:162: stage 0 can't lower a value of type undefined yet |
| [nested_writable_interface.a](unsupported/nested_writable_interface.a) | adamic: /workspace/adamic/notes/accessors/unsupported/nested_writable_interface.a:1:104: Adamic 0.1 refuses an accessor and a field sharing the view of x (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |
| [never_result.a](unsupported/never_result.a) | adamic: /workspace/adamic/notes/accessors/unsupported/never_result.a:1:74: stage 0 can't lower a value of type never yet |
| [null_result.a](unsupported/null_result.a) | adamic: /workspace/adamic/notes/accessors/unsupported/null_result.a:1:15: stage 0 can't lower a function returning null yet |
| [object_setter.a](unsupported/object_setter.a) | adamic: /workspace/adamic/notes/accessors/unsupported/object_setter.a:1:17: stage 0 can't lower object-literal accessors (adamic/object-accessor) yet |
| [optional.a](unsupported/optional.a) | adamic: /workspace/adamic/notes/accessors/unsupported/optional.a:1:104: stage 0 can't lower optional accessor access (adamic/optional-accessor) yet |
| [optional_boolean_field.a](unsupported/optional_boolean_field.a) | adamic: /workspace/adamic/notes/accessors/unsupported/optional_boolean_field.a:1:11: stage 0 can't lower a field of type boolean \| undefined yet |
| [override_representation.a](unsupported/override_representation.a) | adamic: /workspace/adamic/notes/accessors/unsupported/override_representation.a:1:72: stage 0 can't lower an accessor override changing native representation (adamic/accessor-override-representation) yet |
| [parameter_destructure.a](unsupported/parameter_destructure.a) | adamic: /workspace/adamic/notes/accessors/unsupported/parameter_destructure.a:1:57: stage 0 can't lower spread, destructuring or indexed access of an accessor-bearing value (adamic/accessor-property-operation) yet |
| [partial_descriptor.a](unsupported/partial_descriptor.a) | adamic: /workspace/adamic/notes/accessors/unsupported/partial_descriptor.a:1:87: stage 0 can't lower an override changing the getter/setter descriptor halves (adamic/accessor-descriptor-override) yet |
| [private_name.a](unsupported/private_name.a) | adamic: /workspace/adamic/notes/accessors/unsupported/private_name.a:1:11: stage 0 can't lower computed or private accessor names (adamic/accessor-name) yet |
| [readonly_interface.a](unsupported/readonly_interface.a) | adamic: /workspace/adamic/notes/accessors/unsupported/readonly_interface.a:1:100: Adamic 0.1 refuses an accessor and a field sharing the view of x (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |
| [removed_getter.a](unsupported/removed_getter.a) | adamic: /workspace/adamic/notes/accessors/unsupported/removed_getter.a:1:95: stage 0 can't lower an override changing the getter/setter descriptor halves (adamic/accessor-descriptor-override) yet |
| [setter_only_read.a](unsupported/setter_only_read.a) | adamic: /workspace/adamic/notes/accessors/unsupported/setter_only_read.a:1:49: Adamic 0.1 refuses a setter-only property read (adamic/setter-only-read); declare the missing accessor, or use an explicit method |
| [setter_parameter_narrowing.a](unsupported/setter_parameter_narrowing.a) | adamic: /workspace/adamic/notes/accessors/unsupported/setter_parameter_narrowing.a:1:87: Adamic 0.1 refuses an unsound accessor override (adamic/accessor-override); keep the inherited accessor descriptor and its read and write types |
| [setter_value.a](unsupported/setter_value.a) | adamic: /workspace/adamic/notes/accessors/unsupported/setter_value.a:1:101: stage 0 can't lower an accessor assignment or update used as a value (adamic/accessor-update-value) yet |
| [spread.a](unsupported/spread.a) | adamic: /workspace/adamic/notes/accessors/unsupported/spread.a:1:74: stage 0 can't lower spread, destructuring or indexed access of an accessor-bearing value (adamic/accessor-property-operation) yet |
| [static.a](unsupported/static.a) | adamic: /workspace/adamic/notes/accessors/unsupported/static.a:1:11: stage 0 can't lower static accessors (adamic/static-accessor) yet |
| [static_setter.a](unsupported/static_setter.a) | adamic: /workspace/adamic/notes/accessors/unsupported/static_setter.a:1:11: stage 0 can't lower static accessors (adamic/static-accessor) yet |
| [super.a](unsupported/super.a) | adamic: /workspace/adamic/notes/accessors/unsupported/super.a:1:97: stage 0 can't lower super accessor access (adamic/super-accessor) yet |
| [undefined_result.a](unsupported/undefined_result.a) | adamic: /workspace/adamic/notes/accessors/unsupported/undefined_result.a:1:15: stage 0 can't lower a function returning undefined yet |
| [union.a](unsupported/union.a) | adamic: /workspace/adamic/notes/accessors/unsupported/union.a:1:138: stage 0 can't lower accessor dispatch through a union (adamic/accessor-union) yet |
| [unrelated_getter_interface.a](unsupported/unrelated_getter_interface.a) | adamic: /workspace/adamic/notes/accessors/unsupported/unrelated_getter_interface.a:1:163: Adamic 0.1 refuses an accessor and a field sharing the view of x (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |
| [update_value.a](unsupported/update_value.a) | adamic: /workspace/adamic/notes/accessors/unsupported/update_value.a:1:100: stage 0 can't lower an accessor assignment or update used as a value (adamic/accessor-update-value) yet |
| [void_result.a](unsupported/void_result.a) | adamic: /workspace/adamic/notes/accessors/unsupported/void_result.a:1:82: stage 0 can't lower a value of type void yet |
| [writable_interface.a](unsupported/writable_interface.a) | adamic: /workspace/adamic/notes/accessors/unsupported/writable_interface.a:1:91: Adamic 0.1 refuses an accessor and a field sharing the view of x (adamic/accessor-field-view); keep the nominal class type; structural property views do not carry accessor dispatch |

Final validation: `gofmt -l cmd internal` printed nothing, `go vet ./...` passed,
and `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...` passed every
package. The full output is in [gate.log](gate.log). Compiler source is restored
exactly to d785e87 after the one-line mutation.

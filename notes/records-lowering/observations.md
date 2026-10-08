# Final program observations

All 13 admitted programs agree on stdout, stderr and exit code across Node, native and the JavaScript backend. No unintended difference was observed.

The following observations were collected after merging main at `e8ba3d5` into the coverage branch. Native builds use the exact `go run ./cmd/adamic build <file> -o <out>` command; JavaScript executions use the oracle module loader.

## Admitted programs

| Program | Node/native/JavaScript exit | Agreement |
| --- | ---: | --- |
| [records_coverage_calls.a](../../internal/oracle/testdata/records_coverage_calls.a) | 0 | Full output and exit agreement |
| [records_coverage_coalesce_scalars.a](../../internal/oracle/testdata/records_coverage_coalesce_scalars.a) | 0 | Full output and exit agreement |
| [records_coverage_evaluation.a](../../internal/oracle/testdata/records_coverage_evaluation.a) | 0 | Full output and exit agreement |
| [records_coverage_fresh.a](../../internal/oracle/testdata/records_coverage_fresh.a) | 0 | Full output and exit agreement |
| [records_coverage_growth.a](../../internal/oracle/testdata/records_coverage_growth.a) | 0 | Full output and exit agreement |
| [records_coverage_inherited_signature.a](../../internal/oracle/testdata/records_coverage_inherited_signature.a) | 0 | Full output and exit agreement |
| [records_coverage_iteration_edges.a](../../internal/oracle/testdata/records_coverage_iteration_edges.a) | 0 | Full output and exit agreement |
| [records_coverage_json_options.a](../../internal/oracle/testdata/records_coverage_json_options.a) | 0 | Full output and exit agreement |
| [records_coverage_key_forms.a](../../internal/oracle/testdata/records_coverage_key_forms.a) | 0 | Full output and exit agreement |
| [records_coverage_nested_arrays.a](../../internal/oracle/testdata/records_coverage_nested_arrays.a) | 0 | Full output and exit agreement |
| [records_coverage_order_edges.a](../../internal/oracle/testdata/records_coverage_order_edges.a) | 0 | Full output and exit agreement |
| [records_coverage_prototype_own.a](../../internal/oracle/testdata/records_coverage_prototype_own.a) | 0 | Full output and exit agreement |
| [records_coverage_values.a](../../internal/oracle/testdata/records_coverage_values.a) | 0 | Full output and exit agreement |

## Notes programs

The eight dynamic prototype-name probes below are expected stops, not differences. Each native stop exactly matches the JavaScript backend, with empty stdout and exit 70. The literal probes are compile-time refusals. Other probes are unsupported forms or safety refusals; no native executable is produced.

The runtime missing-member guard is in `internal/native/runtime/record.c:84`; literal reads are refused in `internal/lower/records.go:425`, and literal membership keys in `internal/lower/records.go:72`. Numeric literal property-name lowering calls `recordKey` at `internal/lower/records.go:239`; its checker type does not satisfy ordinary numeric expression lowering.

### coalesce_optional_number.a

```typescript
const r: Record<string, number | undefined> = {};
console.log(String(r['x'] ??= 1));
```

Node exit 0, stdout:

```text
1
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/coalesce_optional_number.a:2:20: stage 0 can't lower record ??= with a tagged optional value yet
exit status 1
```

### cycle.a

```typescript
interface R {[key: string]: R}
const r: R = {};
r['self'] = r;
console.log(Object.keys(r).join(','));
```

Node exit 0, stdout:

```text
self
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/cycle.a:1:11: Adamic 0.1 refuses R, a record whose values can reach back to its holder: a cycle reference counting cannot free, and the write at notes/records-lowering/cycle.a:3:1 may close one; use weak links or write only values proven unable to reach the record (adamic/cycle-capable)
exit status 1
```

### dynamic_in___proto__.a

```typescript
const r: Record<string, number> = {};
function probe(key: string): void {console.log(String(key in r));}
probe('__proto__');
```

Node exit 0, stdout:

```text
true
```

Native and JavaScript exit 70, stdout '', stderr:

```text
adamic: panic: record member '__proto__' is missing; records hold own keys only
```

### dynamic_in_constructor.a

```typescript
const r: Record<string, number> = {};
function probe(key: string): void {console.log(String(key in r));}
probe('constructor');
```

Node exit 0, stdout:

```text
true
```

Native and JavaScript exit 70, stdout '', stderr:

```text
adamic: panic: record member 'constructor' is missing; records hold own keys only
```

### dynamic_in_hasOwnProperty.a

```typescript
const r: Record<string, number> = {};
function probe(key: string): void {console.log(String(key in r));}
probe('hasOwnProperty');
```

Node exit 0, stdout:

```text
true
```

Native and JavaScript exit 70, stdout '', stderr:

```text
adamic: panic: record member 'hasOwnProperty' is missing; records hold own keys only
```

### dynamic_in_toString.a

```typescript
const r: Record<string, number> = {};
function probe(key: string): void {console.log(String(key in r));}
probe('toString');
```

Node exit 0, stdout:

```text
true
```

Native and JavaScript exit 70, stdout '', stderr:

```text
adamic: panic: record member 'toString' is missing; records hold own keys only
```

### dynamic_read___proto__.a

```typescript
const r: Record<string, number> = {};
function probe(key: string): void {console.log(String(r[key]));}
probe('__proto__');
```

Node exit 0, stdout:

```text
[object Object]
```

Native and JavaScript exit 70, stdout '', stderr:

```text
adamic: panic: record member '__proto__' is missing; records hold own keys only
```

### dynamic_read_constructor.a

```typescript
const r: Record<string, number> = {};
function probe(key: string): void {console.log(String(r[key]));}
probe('constructor');
```

Node exit 0, stdout:

```text
function Object() { [native code] }
```

Native and JavaScript exit 70, stdout '', stderr:

```text
adamic: panic: record member 'constructor' is missing; records hold own keys only
```

### dynamic_read_hasOwnProperty.a

```typescript
const r: Record<string, number> = {};
function probe(key: string): void {console.log(String(r[key]));}
probe('hasOwnProperty');
```

Node exit 0, stdout:

```text
function hasOwnProperty() { [native code] }
```

Native and JavaScript exit 70, stdout '', stderr:

```text
adamic: panic: record member 'hasOwnProperty' is missing; records hold own keys only
```

### dynamic_read_toString.a

```typescript
const r: Record<string, number> = {};
function probe(key: string): void {console.log(String(r[key]));}
probe('toString');
```

Node exit 0, stdout:

```text
function toString() { [native code] }
```

Native and JavaScript exit 70, stdout '', stderr:

```text
adamic: panic: record member 'toString' is missing; records hold own keys only
```

### fixed_view.a

```typescript
const r: Record<string, number> = {x: 1};
const fixed: {x?: number} = r;
console.log(String(fixed.x));
```

Node exit 0, stdout:

```text
1
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/fixed_view.a:2:7: stage 0 can't lower a Record<string, number> seen as { x?: number; } (fixed objects and records have different storage; copy explicitly) yet
exit status 1
```

### invariance.a

```typescript
interface Animal {name: string}
interface Dog extends Animal {bark: string}
const dogs: Record<string, Dog> = {};
const animals: Record<string, Animal> = dogs;
animals['x'] = {name: 'cat'};
console.log(Object.keys(dogs).join(','));
```

Node exit 0, stdout:

```text
x
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/invariance.a:4:41: Adamic 0.1 refuses a value of type Record<string, Dog> seen as Record<string, Animal>, which can write Animal where Dog is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
exit status 1
```

### json_function_values.a

```typescript
const r: Record<string, () => string> = {x: () => 'item'};
console.log(String(JSON.stringify(r)));
```

Node exit 0, stdout:

```text
{}
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/json_function_values.a:2:35: stage 0 can't lower JSON.stringify record values that may provide a callable toJSON yet
exit status 1
```

### json_object_values.a

```typescript
const r: Record<string, {name: string}> = {x: {name: 'item'}};
console.log(String(JSON.stringify(r)));
```

Node exit 0, stdout:

```text
{"x":{"name":"item"}}
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/json_object_values.a:2:35: stage 0 can't lower JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata) yet
exit status 1
```

### literal_in___proto__.a

```typescript
const r: Record<string, number> = {};
console.log(String('__proto__' in r));
```

Node exit 0, stdout:

```text
true
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/literal_in___proto__.a:2:20: Adamic 0.1 refuses record key __proto__ names an Object.prototype member; use an own-property test and a supported own-key operation; inherited values are outside the record's value type
exit status 1
```

### literal_in_constructor.a

```typescript
const r: Record<string, number> = {};
console.log(String('constructor' in r));
```

Node exit 0, stdout:

```text
true
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/literal_in_constructor.a:2:20: Adamic 0.1 refuses record key constructor names an Object.prototype member; use an own-property test and a supported own-key operation; inherited values are outside the record's value type
exit status 1
```

### literal_in_hasOwnProperty.a

```typescript
const r: Record<string, number> = {};
console.log(String('hasOwnProperty' in r));
```

Node exit 0, stdout:

```text
true
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/literal_in_hasOwnProperty.a:2:20: Adamic 0.1 refuses record key hasOwnProperty names an Object.prototype member; use an own-property test and a supported own-key operation; inherited values are outside the record's value type
exit status 1
```

### literal_in_toString.a

```typescript
const r: Record<string, number> = {};
console.log(String('toString' in r));
```

Node exit 0, stdout:

```text
true
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/literal_in_toString.a:2:20: Adamic 0.1 refuses record key toString names an Object.prototype member; use an own-property test and a supported own-key operation; inherited values are outside the record's value type
exit status 1
```

### literal_read___proto__.a

```typescript
const r: Record<string, number> = {};
console.log(String(r['__proto__']));
```

Node exit 0, stdout:

```text
[object Object]
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/literal_read___proto__.a:2:20: Adamic 0.1 refuses record key __proto__ names an Object.prototype member; use own keys
exit status 1
```

### literal_read_constructor.a

```typescript
const r: Record<string, number> = {};
console.log(String(r['constructor']));
```

Node exit 0, stdout:

```text
function Object() { [native code] }
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/literal_read_constructor.a:2:20: Adamic 0.1 refuses record key constructor names an Object.prototype member; use own keys
exit status 1
```

### literal_read_hasOwnProperty.a

```typescript
const r: Record<string, number> = {};
console.log(String(r['hasOwnProperty']));
```

Node exit 0, stdout:

```text
function hasOwnProperty() { [native code] }
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/literal_read_hasOwnProperty.a:2:20: Adamic 0.1 refuses record key hasOwnProperty names an Object.prototype member; use own keys
exit status 1
```

### literal_read_toString.a

```typescript
const r: Record<string, number> = {};
console.log(String(r['toString']));
```

Node exit 0, stdout:

```text
function toString() { [native code] }
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/literal_read_toString.a:2:20: Adamic 0.1 refuses record key toString names an Object.prototype member; use own keys
exit status 1
```

### mixed_signature.a

```typescript
interface R {named: number; [key: string]: number}
const r: R = {named: 1};
console.log(String(r['named']));
```

Node exit 0, stdout:

```text
1
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/mixed_signature.a:1:29: stage 0 can't lower an index signature beside named members, or a non-mutable unrestricted string signature (dictionary storage cannot preserve named-property contracts) yet
exit status 1
```

### multiple_spread.a

```typescript
const r: Record<string, number> = {x: 1};
const copy: Record<string, number> = {...r, ...r};
console.log(Object.keys(copy).join(','));
```

Node exit 0, stdout:

```text
x
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/multiple_spread.a:2:45: Adamic 0.1 refuses a spread after the first field; spread once, first (adamic/single-spread)
exit status 1
```

### number_signature.a

```typescript
const r: {[key: number]: number} = {1: 1};
console.log(String(r[1]));
```

Node exit 0, stdout:

```text
1
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/number_signature.a:1:11: stage 0 can't lower an index signature beside named members, or a non-mutable unrestricted string signature (dictionary storage cannot preserve named-property contracts) yet
exit status 1
```

### numeric_literal_key.a

```typescript
const r: Record<string, number> = {10: 10};
console.log(String(r[10]));
```

Node exit 0, stdout:

```text
10
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: lower: notes/records-lowering/numeric_literal_key.a:1:36: the checker gave a numeric literal a type that isn't a number literal
exit status 1
```

### optional_boolean.a

```typescript
const r: Record<string, boolean | undefined> = {x: undefined};
console.log(String(r['x']));
```

Node exit 0, stdout:

```text
undefined
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/optional_boolean.a:1:48: stage 0 can't lower record values with an unsupported slot representation yet
exit status 1
```

### optional_index.a

```typescript
const r: Record<string, number> = {};
console.log(String(r?.['x']));
```

Node exit 0, stdout:

```text
undefined
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/optional_index.a:2:20: stage 0 can't lower optional record indexing yet
exit status 1
```

### or_assign.a

```typescript
const r: Record<string, number> = {zero: 0};
console.log(String(r['zero'] ||= 5));
```

Node exit 0, stdout:

```text
5
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/or_assign.a:2:30: Adamic 0.1 refuses ||=; write the if
exit status 1
```

### prototype_initializer.a

```typescript
const r: Record<string, number> = {__proto__: 1};
console.log(Object.keys(r).join(','));
```

Node exit 0, stdout:

```text
(empty)
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/prototype_initializer.a:1:36: stage 0 can't lower a prototype-setting record literal yet
exit status 1
```

### readonly_signature.a

```typescript
const r: Readonly<Record<string, number>> = {x: 1};
console.log(String(r['x']));
```

Node exit 0, stdout:

```text
1
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/readonly_signature.a:1:7: stage 0 can't lower a value of type Readonly<Record<string, number>> yet
exit status 1
```

### spread_fixed_in.a

```typescript
const fixed = {x: 1};
const r: Record<string, number> = {...fixed};
console.log(String(r['x']));
```

Node exit 0, stdout:

```text
1
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/spread_fixed_in.a:2:39: stage 0 can't lower a record without a pure mutable string index signature yet
exit status 1
```

### spread_fixed_out.a

```typescript
const r: Record<string, number> = {x: 1};
const fixed = {...r, named: 2};
console.log(String(fixed.named));
console.log(Object.keys(fixed).join(','));
```

Node exit 0, stdout:

```text
2
x,named
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/spread_fixed_out.a:2:16: stage 0 can't lower spreading a value yet
exit status 1
```

### spread_widening.a

```typescript
interface Narrow {name: 'dog'}
interface Wide {name: string}
const dogs: Record<string, Narrow> = {x: {name: 'dog'}};
const animals: Record<string, Wide> = {...dogs};
console.log(Object.keys(animals).join(','));
```

Node exit 0, stdout:

```text
x
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/spread_widening.a:4:40: Adamic 0.1 refuses record spread shares values with an invariant-mutable slot that the target can widen; keep the value's mutable fields invariant or copy the values too (adamic/invariant-mutable)
exit status 1
```

### union_values.a

```typescript
const mixed: Record<string, number | string> = {n: 1, s: 'text'};
console.log(String(mixed['n'])); console.log(String(mixed['s'])); console.log(String(mixed['missing']));
```

Node exit 0, stdout:

```text
1
text
undefined
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/union_values.a:1:48: stage 0 can't lower record values with an unsupported slot representation yet
exit status 1
```

### weak_values.a

```typescript
import type {Weak} from 'adamic';
interface Item {name: string}
const r: Record<string, Weak<Item>> = {};
console.log(Object.keys(r).join(','));
```

Node exit 0, stdout:

```text
(empty)
```

Native build exit 1; JavaScript emission also stops. Diagnostic:

```text
adamic: notes/records-lowering/weak_values.a:3:39: stage 0 can't lower record values with an unsupported slot representation yet
exit status 1
```

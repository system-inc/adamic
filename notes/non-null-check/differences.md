# Non-null output differences

22 programs differ from source Node; nine new oracle programs agree.

All source Node runs exit 0 with empty stderr. All native failure runs exit 70.

Twenty differences follow the branch's documented invariant panic contract.
The two absent-field cases also disagree between native and generated JavaScript.

## Programs and exact observations

### missing_array_outside.a

[missing_array_outside.a](missing_array_outside.a)

```ts
const values: number[] = [];
let calls = 0;
function index(): number { calls += 1; console.log(`index ${calls}`); return 0; }
console.log('before');
console.log(`value ${values[index()]!}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nindex 1\nvalue undefined\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nindex 1\n"; stderr "adamic: panic: non-null assertion failed: values[index()]! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nindex 1\n"; stderr "adamic: panic: non-null assertion failed: values[index()]! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_array_reference_outside.a

[missing_array_reference_outside.a](missing_array_reference_outside.a)

```ts
let calls = 0;
function result(): number[] | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nresult 1\ntrue\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_array_reference_try.a

[missing_array_reference_try.a](missing_array_reference_try.a)

```ts
let calls = 0;
function result(): number[] | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
try {
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nresult 1\ntrue\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_array_try.a

[missing_array_try.a](missing_array_try.a)

```ts
const values: number[] = [];
let calls = 0;
function index(): number { calls += 1; console.log(`index ${calls}`); return 0; }
try {
console.log('before');
console.log(`value ${values[index()]!}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nindex 1\nvalue undefined\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nindex 1\n"; stderr "adamic: panic: non-null assertion failed: values[index()]! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nindex 1\n"; stderr "adamic: panic: non-null assertion failed: values[index()]! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_chain_outside.a

[missing_chain_outside.a](missing_chain_outside.a)

```ts
function box(): { readonly value: number } | undefined { console.log('box 1'); return undefined; }
console.log('before');
console.log(`value ${box()?.value!}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nbox 1\nvalue undefined\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nbox 1\n"; stderr "adamic: panic: non-null assertion failed: box()?.value! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nbox 1\n"; stderr "adamic: panic: non-null assertion failed: box()?.value! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_chain_try.a

[missing_chain_try.a](missing_chain_try.a)

```ts
function box(): { readonly value: number } | undefined { console.log('box 1'); return undefined; }
try {
console.log('before');
console.log(`value ${box()?.value!}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nbox 1\nvalue undefined\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nbox 1\n"; stderr "adamic: panic: non-null assertion failed: box()?.value! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nbox 1\n"; stderr "adamic: panic: non-null assertion failed: box()?.value! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_closure_reference_outside.a

[missing_closure_reference_outside.a](missing_closure_reference_outside.a)

```ts
let calls = 0;
function result(): (() => number) | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nresult 1\ntrue\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_closure_reference_try.a

[missing_closure_reference_try.a](missing_closure_reference_try.a)

```ts
let calls = 0;
function result(): (() => number) | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
try {
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nresult 1\ntrue\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_field_outside.a

[missing_field_outside.a](missing_field_outside.a)

```ts
const box: { value?: string } = {};
console.log('before');
console.log(`value ${box.value!}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nvalue undefined\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\n"; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n".

Generated JavaScript: exit 70; stdout "before\n"; stderr "adamic: panic: non-null assertion failed: box.value! is null or undefined\n".

Likely cause: internal/native/emit_expressions.go:108 emits a required field lookup
through fieldSlot instead of an absence-aware optional field read. The lookup reaches
internal/native/runtime/object.c:49 when the literal has no value slot. The check
constructed at internal/lower/non_null.go:52 never receives an undefined value.

### missing_field_try.a

[missing_field_try.a](missing_field_try.a)

```ts
const box: { value?: string } = {};
try {
console.log('before');
console.log(`value ${box.value!}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nvalue undefined\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\n"; stderr "adamic: panic: compiler bug: a field the checker proved is there is missing\n".

Generated JavaScript: exit 70; stdout "before\n"; stderr "adamic: panic: non-null assertion failed: box.value! is null or undefined\n".

Likely cause: internal/native/emit_expressions.go:108 emits a required field lookup
through fieldSlot instead of an absence-aware optional field read. The lookup reaches
internal/native/runtime/object.c:49 when the literal has no value slot. The check
constructed at internal/lower/non_null.go:52 never receives an undefined value.

### missing_map_outside.a

[missing_map_outside.a](missing_map_outside.a)

```ts
const values = new Map<string, boolean>();
let calls = 0;
function key(): string { calls += 1; console.log(`key ${calls}`); return 'missing'; }
console.log('before');
console.log(`value ${values.get(key())!}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nkey 1\nvalue undefined\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nkey 1\n"; stderr "adamic: panic: non-null assertion failed: values.get(key())! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nkey 1\n"; stderr "adamic: panic: non-null assertion failed: values.get(key())! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_map_reference_outside.a

[missing_map_reference_outside.a](missing_map_reference_outside.a)

```ts
let calls = 0;
function result(): Map<string, number> | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nresult 1\ntrue\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_map_reference_try.a

[missing_map_reference_try.a](missing_map_reference_try.a)

```ts
let calls = 0;
function result(): Map<string, number> | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
try {
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nresult 1\ntrue\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_map_try.a

[missing_map_try.a](missing_map_try.a)

```ts
const values = new Map<string, boolean>();
let calls = 0;
function key(): string { calls += 1; console.log(`key ${calls}`); return 'missing'; }
try {
console.log('before');
console.log(`value ${values.get(key())!}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nkey 1\nvalue undefined\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nkey 1\n"; stderr "adamic: panic: non-null assertion failed: values.get(key())! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nkey 1\n"; stderr "adamic: panic: non-null assertion failed: values.get(key())! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_narrowed_outside.a

[missing_narrowed_outside.a](missing_narrowed_outside.a)

```ts
function run(): void {
let value: number | undefined = 0;
const clear = (): void => { value = undefined; };
if (value !== undefined) { clear();
console.log('before');
console.log(`value ${value!}`);
console.log('after');
}
}
run();
```

Source Node: exit 0; stdout "before\nvalue undefined\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\n"; stderr "adamic: panic: non-null assertion failed: value! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\n"; stderr "adamic: panic: non-null assertion failed: value! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_narrowed_try.a

[missing_narrowed_try.a](missing_narrowed_try.a)

```ts
function run(): void {
let value: number | undefined = 0;
const clear = (): void => { value = undefined; };
if (value !== undefined) { clear();
try {
console.log('before');
console.log(`value ${value!}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
}
}
run();
```

Source Node: exit 0; stdout "before\nvalue undefined\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\n"; stderr "adamic: panic: non-null assertion failed: value! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\n"; stderr "adamic: panic: non-null assertion failed: value! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_null_outside.a

[missing_null_outside.a](missing_null_outside.a)

```ts
let calls = 0;
function result(): RegExpMatchArray | null { calls += 1; console.log(`match ${calls}`); return 'a'.match(/z/); }
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nmatch 1\nfalse\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nmatch 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nmatch 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_null_try.a

[missing_null_try.a](missing_null_try.a)

```ts
let calls = 0;
function result(): RegExpMatchArray | null { calls += 1; console.log(`match ${calls}`); return 'a'.match(/z/); }
try {
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nmatch 1\nfalse\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nmatch 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nmatch 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_result_outside.a

[missing_result_outside.a](missing_result_outside.a)

```ts
let calls = 0;
function result(): string | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
console.log('before');
console.log(`value ${result()!}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nresult 1\nvalue undefined\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_result_try.a

[missing_result_try.a](missing_result_try.a)

```ts
let calls = 0;
function result(): string | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
try {
console.log('before');
console.log(`value ${result()!}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nresult 1\nvalue undefined\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. This is the branch's intentional loud
invariant check. Native and generated JavaScript agree.

### missing_object_reference_outside.a

[missing_object_reference_outside.a](missing_object_reference_outside.a)

```ts
let calls = 0;
function result(): { readonly name: string } | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
```

Source Node: exit 0; stdout "before\nresult 1\ntrue\nafter\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. Native and generated JavaScript agree.

### missing_object_reference_try.a

[missing_object_reference_try.a](missing_object_reference_try.a)

```ts
let calls = 0;
function result(): { readonly name: string } | undefined { calls += 1; console.log(`result ${calls}`); return undefined; }
try {
console.log('before');
const value = result()!;
console.log(`${value === undefined}`);
console.log('after');
} catch { console.log('caught'); } finally { console.log('finally'); }
```

Source Node: exit 0; stdout "before\nresult 1\ntrue\nafter\nfinally\n"; stderr "".

Native release build: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Generated JavaScript: exit 70; stdout "before\nresult 1\n"; stderr "adamic: panic: non-null assertion failed: result()! is null or undefined\n".

Responsible line: internal/lower/non_null.go:52 constructs Coalesce with Panic;
source TypeScript erases the assertion. Native and generated JavaScript agree.

## Compile-time limitations

### blocked_boolean_field.a

```ts
function run(box: { readonly value?: boolean }): void { console.log(`${box.value!}`); }
run({ value: false });
```

Source Node: exit 0; stdout "false\n"; stderr "".

Build: exit 1.

```text
adamic: /workspace/adamic/notes/non-null-check/blocked_boolean_field.a:1:72: stage 0 can't lower a field of type boolean | undefined yet
exit status 1
```

### blocked_brackets.a

```ts
let calls = 0;
function box(): { readonly value?: number } { calls += 1; return { value: 0 }; }
console.log(`${box().value!} ${calls}`);
function read(value: { readonly n?: number; readonly s?: string }): void {
    console.log(`${value['n']!} '${value['s']!}'`);
    if (value.n !== undefined) { console.log(`${(value['n'])!}`); }
}
read({ n: 0, s: '' });
```

Source Node: exit 0; stdout "0 1\n0 ''\n0\n"; stderr "".

Build: exit 1.

```text
adamic: /workspace/adamic/notes/non-null-check/blocked_brackets.a:5:20: stage 0 can't lower an ElementAccessExpression yet
exit status 1
```

### blocked_generic.a

```ts
function unwrap<T>(value: T | undefined): T { return value!; }
console.log(`${unwrap<number>(0)}`);
```

Source Node: exit 0; stdout "0\n"; stderr "".

Build: exit 1.

```text
adamic: /workspace/adamic/notes/non-null-check/blocked_generic.a:1:54: stage 0 can't lower a value of type NonNullable<T> yet
exit status 1
```

### blocked_null_undefined.a

```ts
function run(value: RegExpMatchArray | null | undefined): RegExpMatchArray { return value!; }
console.log(run('a'.match(/a/))[0]!);
```

Source Node: exit 0; stdout "a\n"; stderr "".

Build: exit 1.

```text
adamic: /workspace/adamic/notes/non-null-check/blocked_null_undefined.a:1:14: stage 0 can't lower a value of type RegExpMatchArray | null | undefined yet
exit status 1
```

### blocked_optional_call.a

```ts
function run(value: (() => number) | undefined): number { return value?.()!; }
console.log(`${run(() => 0)}`);
```

Source Node: exit 0; stdout "0\n"; stderr "".

Build: exit 1.

```text
adamic: /workspace/adamic/notes/non-null-check/blocked_optional_call.a:1:66: stage 0 can't lower a call through ?. (an optional call) yet
exit status 1
```

### blocked_scalar_null.a

```ts
function run(value: number | null): number { return value!; }
console.log(`${run(0)}`);
```

Source Node: exit 0; stdout "0\n"; stderr "".

Build: exit 1.

```text
adamic: /workspace/adamic/notes/non-null-check/blocked_scalar_null.a:1:14: stage 0 can't lower a value of type number | null yet
exit status 1
```

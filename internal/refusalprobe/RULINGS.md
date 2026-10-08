# Refusal rulings

Rulings from @system_adamic, received October 8, 2026. Refused means a permanent
language refusal with a message saying what to write instead; NotYet means intended support.

Observations below are from main `45487a809f89885a3fc651cd590e7dabf31362dc`
through the probe loader, lowerer and C generator. Every neighbor compiled.
The complete programs retain the surroundings of `generate(1, 0, entry)`.
No skipped subtests are used.

## any

Ruling: Explicit any is permanently Refused in `.a`. In tsc's `.ts`, roadmap step 09 must rewrite every site or use a checked unknown.

Matches: no. Compiler work remains; keep this bad program and neighbor ready to become a refusal pair when main implements the ruling.

Main observed: NotYet: a value of type any

Bad program for the ruled refusal (`main.a`):

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function run<T>(item: T): T {
let value: any = 1;

return item;
}
run(1);
```

Neighbor (`main.a`), accepted:

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function run<T>(item: T): T {
let value: number = 1;

return item;
}
run(1);
```

## expando

Ruling: A function expando is permanently Refused. Write an object with a call method or a namespace.

Matches: no. Compiler work remains; keep this bad program and neighbor ready to become a refusal pair when main implements the ruling.

Main observed: NotYet: assigning a field of a value

Bad program for the ruled refusal (`main.a`):

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function value() {} value.extra = 1;
```

Neighbor (`main.a`), accepted:

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
const value = {call: (): void => {}, extra: 1};
```

## eval

Ruling: eval is permanently Refused. Write the computation directly.

Matches: no. Compiler work remains; keep this bad program and neighbor ready to become a refusal pair when main implements the ruling.

Main observed: NotYet: reading eval

Bad program for the ruled refusal (`main.a`):

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function run<T>(item: T): T {
eval('1');

return item;
}
run(1);
```

Neighbor (`main.a`), accepted:

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function run<T>(item: T): T {
const value = 1;

return item;
}
run(1);
```

## function-type

Ruling: The Function type is Refused as an annotation in `.a`, including an unused parameter. Write a call signature. In `.ts`, holding a Function value is accepted; calling through it is refused.

Matches: no. Compiler work remains; keep this bad program and neighbor ready to become a refusal pair when main implements the ruling.

Main observed: Accepted; lowering and C generation completed.

Bad program for the ruled refusal (`main.a`):

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function take(value: Function): void {}
```

Neighbor (`main.a`), accepted:

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function take(value: () => void): void {}
```

## new-function

Ruling: new Function is permanently Refused. Write a function or closure.

Matches: no. Compiler work remains; keep this bad program and neighbor ready to become a refusal pair when main implements the ruling.

Main observed: NotYet: new an Identifier

Bad program for the ruled refusal (`main.a`):

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function run<T>(item: T): T {
const value = new Function('return 1');

return item;
}
run(1);
```

Neighbor (`main.a`), accepted:

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function run<T>(item: T): T {
const value = (): number => 1;

return item;
}
run(1);
```

## record

Ruling: An empty Record<string, number> is accepted and stays accepted.

Matches: yes. Both the empty Record and its Map neighbor are active accepted probes.

Main observed: Accepted; lowering and C generation completed.

Accepted program (`main.a`):

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function run<T>(item: T): T {
const value: Record<string, number> = {};

return item;
}
run(1);
```

Neighbor (`main.a`), accepted:

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
function run<T>(item: T): T {
const value = new Map<string, number>();

return item;
}
run(1);
```

## merging

Ruling: Class/interface merging that claims a field the class never initializes is Refused in `.a` at the declaration, even unused. In `.ts`, a read of that field is a checked read.

Matches: no. Compiler work remains; keep this bad program and neighbor ready to become a refusal pair when main implements the ruling.

Main observed: Accepted; lowering and C generation completed.

Bad program for the ruled refusal (`main.a`):

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
class Box { n = 1; } interface Box { extra: number; } const box = new Box();
```

Neighbor (`main.a`), accepted:

```typescript
import { panic } from 'adamic';
const padding = 891;
console.log(`${padding}`);
class Box { n = 1; } const box = new Box();
```

## Step 04 catalog observations

Observed on merged `devtools/port-refusalprobe`
`69385a98a6685431cf10852e499ff876578c0ccf`:

- `nodeLibraryRefusal`: `import {userInfo} from 'node:os'; userInfo();`
  returns NotYet `node:os.userInfo`. Its neighbor `console.log('user');` compiles.
- `typedArrayUnsupported`: `const value = new Int8Array(4);` returns NotYet
  `typed array element type Int8Array`. Replacing Int8Array with Uint8Array compiles.
- `predicateArguments`: the `predicate-argument` catalog program returns Refused
  `an unproven predicate argument for parameter callback`. Its proven typeof neighbor compiles.
- `namespaceRefusal`: the `namespace` catalog program returns Refused
  `this in a namespace function; a qualified call and a detached call have different receivers`.
  Its explicit-state neighbor compiles. Type-only namespaces are now accepted.

These are reachable contracts, not new language rulings. The two NotYet entries
stay outside the permanent-refusal generator; tests demand their exact diagnostic
and compile their neighbors. No skips are added.

The complete package suite still fails independently of these four contracts:
the audit also lacks `argumentsRefusal`, `nodeBufferUnsupportedUse`, and
`libraryMethod`. Existing negative probes `definite-local`, `definite-field`,
`void`, `in`, `comma`, `and-assign`, `or-assign`, `arguments`, `truthiness`, and
`parameter-properties` now compile. The strict-diagnostic regression also uses
the accepted definite-local program. Are these accepted constructs intended
support or missing refusals? This follow-up records the observation without
changing their language contracts or weakening the failing checks.

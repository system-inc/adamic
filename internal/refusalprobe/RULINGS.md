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

The initial follow-up found three more unmapped helpers and ten accepted legacy
negative probes. The parent clarified that main's acceptances are intended except
for the six permanent refusals above. None of those ten probes is in that list.
`definite-local`, `definite-field`, `void`, `in`, `comma`, `and-assign`, `or-assign`,
`arguments`, `truthiness`, and `parameter-properties` are now accepted pairs.
Their original programs and neighbors are both compiled under varied surroundings;
this includes definite assignment before use inside a generic function. The six
compiler-work boundaries above remain unchanged, and the empty Record remains accepted.

The remaining helper contracts are now covered:

- `argumentsRefusal`: `function count(): void { console.log(`${arguments[0]}`); } count();`
  returns Refused `arguments other than a read of arguments.length`. Its neighbor
  reads `arguments.length` and compiles.
- `nodeBufferUnsupportedUse`: `import {Buffer} from 'node:buffer'; console.log(`${Buffer.from('abc').byteOffset}`);`
  returns NotYet `Buffer.byteOffset outside the census value reads`. Reading
  `length` instead compiles.
- `libraryMethod` is a classifier, not an error-returning helper. Its shorthand
  object-field refusal branch is reachable: `const method = Number.parseInt; const object = {method};`
  returns NotYet `a library method value outside a const alias (an object field erases its receiver and callable ABI); use an arrow`.
  The neighbor `const method = (text: string): number => Number.parseInt(text); const object = {method};` compiles.

No new regressions were identified among the ten reclassified inputs.

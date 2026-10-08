# Questions awaiting a language ruling

These seven questions have been sent to @system_adamic. They remain unresolved.
Observed against main `45487a809f89885a3fc651cd590e7dabf31362dc` through the
probe loader, lowerer and C generator. Every neighbor compiled. These are
observations, not decisions about which constructs the language should refuse.

The complete programs below reproduce `generate(1, 0, entry)`, including its
surroundings. The entries remain non-executable catalog boundaries; they do
not create test skips.

## any

Question: should explicit any get Refused rather than main's observed NotYet for a value of type any?

Main observed: NotYet: a value of type any

Probe (`main.a`):

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

Question: should a function expando get Refused rather than main's observed NotYet assigning a field of a value?

Main observed: NotYet: assigning a field of a value

Probe (`main.a`):

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
function value() {}
```

## eval

Question: should eval get Refused rather than main's observed NotYet reading eval?

Main observed: NotYet: reading eval

Probe (`main.a`):

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

Question: should an unused Function-typed parameter be refused? Main accepts this program.

Main observed: Accepted; lowering and C generation completed.

Probe (`main.a`):

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

Question: should new Function get Refused rather than main's observed NotYet new an Identifier?

Main observed: NotYet: new an Identifier

Probe (`main.a`):

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

Question: should an empty Record<string, number> be refused? Main accepts this program.

Main observed: Accepted; lowering and C generation completed.

Probe (`main.a`):

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

Question: should class/interface merging with an unused claimed field be refused? Main accepts this program.

Main observed: Accepted; lowering and C generation completed.

Probe (`main.a`):

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

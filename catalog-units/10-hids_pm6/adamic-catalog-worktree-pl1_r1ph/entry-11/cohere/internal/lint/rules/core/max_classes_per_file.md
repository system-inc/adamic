# `max-classes-per-file`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **18** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## Where cohere differs from ESLint

A class declared inside a function, a method's body included, does not count. ESLint counts every class in the file. Kirk ruled it on #mbbg6js: one class per file names the design unit a reader greps for, and a class declared inside a test's `describe` or `it` is a fixture, not a unit. A class in a namespace or a static block still counts. Class expressions are left out through upstream's own `ignoreExpressions`, which ahra and the ESLint configuration both set. ESLint runs the same ruling through `nexus/max-classes-per-file` (Nexus `MaxClassesPerFileRule.ts`), the core rule with a nested declaration left out of its count, and the engine parity check pairs it with this rule, so the two engines agree (#cn8sthd). Its fixture shares this rule's rows verbatim.

## What it checks

Enforce a maximum number of classes per file

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

18 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/files/MagicBytes.ts:10`**

```
import { bytesToAscii } from '@nexus/source/encoding/ByteEncoding';
```

> File has too many classes (3). Maximum allowed is 1

**`libraries/structure/libraries/nexus/source/money/MonetaryDecimal.ts:34`**

```
export interface DecimalStringConvertibleInterface {
```

> File has too many classes (2). Maximum allowed is 1

**`libraries/structure/libraries/nexus/source/structured-text/json/JsonValueTransformers.ts:2`**

```
import type { JsonType } from '@nexus/source/structured-text/json/types/JsonTypes';
```

> File has too many classes (3). Maximum allowed is 1

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:2`**

```
import type { ConstructorType, NestedConstructorType } from '@nexus/source/types/ClassTypes';
```

> File has too many classes (7). Maximum allowed is 1

**`libraries/structure/libraries/nexus/source/types/PrimitiveType.test.ts:2`**

```
import type { ConstructorType } from '@nexus/source/types/ClassTypes';
```

> File has too many classes (2). Maximum allowed is 1

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


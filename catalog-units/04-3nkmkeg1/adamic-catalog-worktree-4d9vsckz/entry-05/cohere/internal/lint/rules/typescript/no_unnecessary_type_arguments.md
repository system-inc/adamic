# `@typescript-eslint/no-unnecessary-type-arguments`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **79** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | yes |

## What it checks

Disallow type arguments that are equal to the default

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

79 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/collections/Dictionary.test.ts:582`**

```
const result = dictionaryFrom<unknown>(input);
```

> This is the default value for this type parameter, so it can be omitted

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:173`**

```
const constructorArray: NestedConstructorType<object> = [SimpleClass, ComplexClass, NoParamClass];
```

> This is the default value for this type parameter, so it can be omitted

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:183`**

```
const deeplyNested: NestedConstructorType<object> = [SimpleClass, [ComplexClass, [NoParamClass]]];
```

> This is the default value for this type parameter, so it can be omitted

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:188`**

```
expect(typeof (deeplyNested[1] as NestedConstructorType<object>[])[0]).toBe('function');
```

> This is the default value for this type parameter, so it can be omitted

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:189`**

```
expect(Array.isArray((deeplyNested[1] as NestedConstructorType<object>[])[1])).toBe(true);
```

> This is the default value for this type parameter, so it can be omitted

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.


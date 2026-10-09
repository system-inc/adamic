# `@typescript-eslint/method-signature-style`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **39** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Enforce using a particular method signature syntax

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

39 in the tree. Showing the first few.

**`libraries/structure/StructureSettings.ts:145`**

```
TypedDocumentString: new (value: string) => { toString(): string };
```

> Shorthand method signature is forbidden. Use a function property instead

**`libraries/structure/libraries/nexus/source/coordination/DeferredValue.ts:15`**

```
get(): T;
```

> Shorthand method signature is forbidden. Use a function property instead

**`libraries/structure/libraries/nexus/source/coordination/DeferredValue.ts:21`**

```
get(): Promise<T>;
```

> Shorthand method signature is forbidden. Use a function property instead

**`libraries/structure/libraries/nexus/source/coordination/types/ExecutableInterface.ts:11`**

```
execute(...args: T): void | Promise<void>;
```

> Shorthand method signature is forbidden. Use a function property instead

**`libraries/structure/libraries/nexus/source/files/FileTypeValidator.ts:19`**

```
validate(mimeType: string, data: ArrayBuffer): Promise<FileValidationResultInterface>;
```

> Shorthand method signature is forbidden. Use a function property instead

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.


# Contextual signature for the canonical file name

At `src/compiler/path.ts:1051:126` in the preceding adapted tree, the
`identity` fallback in `getRelativePathFromDirectory` needs a concrete
contextual signature. The local `GetCanonicalFileName` annotation supplies
the string-to-string substitution under `docs/rulings.md:129`. Adamic ruled
yes on the annotation on Nov 22, 2026, and Compiler accepted it under the
split on Nov 23, 2026 (J-7sj4k10).

Before writing, the adapter checks that `path.ts` imports
`GetCanonicalFileName` under its own name and that `core.ts` declares it as
the string-to-string type quoted below.

## Before

```ts
const getCanonicalFileName = typeof getCanonicalFileNameOrIgnoreCase === "function" ? getCanonicalFileNameOrIgnoreCase : identity;
```

## After

```ts
const getCanonicalFileName: GetCanonicalFileName = typeof getCanonicalFileNameOrIgnoreCase === "function" ? getCanonicalFileNameOrIgnoreCase : identity;
```

## Existing type in `src/compiler/core.ts`

```ts
export type GetCanonicalFileName = (fileName: string) => string;
```

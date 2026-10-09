# `@typescript-eslint/require-await`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **436** when audited; not re-measured since, see below |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow async functions which do not return promises and have no `await` expression

## Why this recommendation

All 6 samples are `export async function generateMetadata(): Promise<Metadata>`, a Next.js framework signature contract present in 65 route files that must stay async whether or not it awaits.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Yes** to **No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

**Stale, kept as the audit's record.** The 436 below were the `generateMetadata` world, and that world
is gone: on 2026-10-01, 43 route files write `generateMetadata` without `async`, and the core
`require-await` rule cohere enables reports none of the 22 that keep it. The core rule's current numbers
(cohere 3, ESLint 89, and why 86 of ESLint's are contract sites cohere is right to leave) are in
`../core/require_await.md`. This typed variant is not ported, so it has no cohere count, and its
ESLint count was not re-measured.

436 in the tree when audited. Showing the first few.

**`app/(os-layout)/ahra/page.tsx:12`**

```
export async function generateMetadata(): Promise<Metadata> {
```

> Async function 'generateMetadata' has no 'await' expression

**`app/(os-layout)/art/page.tsx:8`**

```
export async function generateMetadata(): Promise<Metadata> {
```

> Async function 'generateMetadata' has no 'await' expression

**`app/(os-layout)/contacts/[id]/page.tsx:9`**

```
export async function generateMetadata(): Promise<Metadata> {
```

> Async function 'generateMetadata' has no 'await' expression

**`app/(os-layout)/contacts/page.tsx:11`**

```
export async function generateMetadata(): Promise<Metadata> {
```

> Async function 'generateMetadata' has no 'await' expression

**`app/(os-layout)/data/layout.tsx:14`**

```
export async function generateMetadata(): Promise<Metadata> {
```

> Async function 'generateMetadata' has no 'await' expression

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


# `@typescript-eslint/no-unsafe-call`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **212** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow calling a value with type `any`

## Why this recommendation

All 6 sampled sites are Buffer.from(...) poisoned by @cloudflare/workers-types declaring `Buffer: any` (index.d.ts:348) alongside @types/node in tsconfig `types`, so this is one type-resolution config fix rather than 212 real defects.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Yes** to **Maybe**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

212 in the tree. Showing the first few.

**`app/api/finance/transaction/[transactionId]/attachments/route.ts:42`**

```
const originalBytes = Buffer.from(await file.arrayBuffer());
```

> Unsafe call of an `any` typed value

**`app/api/spotify/callback/route.ts:36`**

```
const basic = Buffer.from(`${creds.clientId}:${creds.clientSecret}`).toString('base64');
```

> Unsafe call of an `any` typed value

**`app/api/spotify/callback/route.ts:36`**

```
const basic = Buffer.from(`${creds.clientId}:${creds.clientSecret}`).toString('base64');
```

> Unsafe call of an `any` typed value

**`app/api/tasks/[id]/attachments/route.ts:30`**

```
const uploadedBytes = Buffer.from(await file.arrayBuffer());
```

> Unsafe call of an `any` typed value

**`libraries/structure/libraries/nexus/source/encoding/StreamTransforms.test.ts:32`**

```
const expected = Buffer.from(testData).toString('base64');
```

> Unsafe call of an `any` typed value

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


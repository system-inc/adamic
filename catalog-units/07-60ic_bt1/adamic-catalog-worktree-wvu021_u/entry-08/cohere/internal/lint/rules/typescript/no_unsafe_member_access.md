# `@typescript-eslint/no-unsafe-member-access`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **395** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow member access on a value with type `any`

## Why this recommendation

5 of 6 sampled sites are `.from`/`.toString` on the `Buffer: any` global injected by @cloudflare/workers-types, so the count measures a broken ambient type, not unsafe code.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **Yes** to **Maybe**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

395 in the tree. Showing the first few.

**`app/api/finance/transaction/[transactionId]/attachments/route.ts:42`**

```
const originalBytes = Buffer.from(await file.arrayBuffer());
```

> Unsafe member access .from on an `any` value

**`app/api/spotify/callback/route.ts:36`**

```
const basic = Buffer.from(`${creds.clientId}:${creds.clientSecret}`).toString('base64');
```

> Unsafe member access .from on an `any` value

**`app/api/spotify/callback/route.ts:36`**

```
const basic = Buffer.from(`${creds.clientId}:${creds.clientSecret}`).toString('base64');
```

> Unsafe member access .clientId on an `any` value

**`app/api/spotify/callback/route.ts:36`**

```
const basic = Buffer.from(`${creds.clientId}:${creds.clientSecret}`).toString('base64');
```

> Unsafe member access .clientSecret on an `any` value

**`app/api/spotify/callback/route.ts:36`**

```
const basic = Buffer.from(`${creds.clientId}:${creds.clientSecret}`).toString('base64');
```

> Unsafe member access .toString on an `any` value

### What fixing looks like

Not auto-fixable, and site-by-site is the wrong place to start. Measured on ahra, 2026-09-08:
**188 of 507 sites are `Buffer`**, and they are not unsafe code at all. `@cloudflare/workers-types`
ships `declare const Buffer: any` (index.d.ts:486), and the base TypeScript configuration lists it
after `node` in `types`, so the `any` wins over the real declaration from `@types/node`. A two-line
file calling `Buffer.from(...).toString('hex')` produces nine findings.

Removing `@cloudflare/workers-types` from the `types` array was measured on the same tree:

    no-unsafe-member-access    507 -> 264
    no-unsafe-assignment       435 -> 331
    no-unsafe-call             257 ->  42
    no-unsafe-argument         141 ->  86
    total                    1,340 -> 723
    type diagnostics            15 ->  20

617 findings for one config line. The five new type errors are real bugs the `any` was hiding
(`Uint8Array` passed where `Buffer` is required, in ThingsApi and the see/look route), which is the
rule doing its job one level up.

The general lesson, and the reason this section exists: **on any of the four unsafe-* rules, count
by the type that is `any` before touching a single call site.** A broken ambient declaration
concentrates hundreds of findings behind one cause, and sweeping them individually is not just slow
but wrong, because it treats a configuration defect as a code defect. Sample the sites, group them
by the value whose type is `any`, and fix the declaration if one dominates.


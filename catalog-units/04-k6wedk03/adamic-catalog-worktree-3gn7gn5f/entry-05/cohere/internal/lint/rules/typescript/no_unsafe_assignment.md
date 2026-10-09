# `@typescript-eslint/no-unsafe-assignment`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **387** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow assigning a value with type `any` to variables and properties

## Why this recommendation

Catches a defect rather than a preference. The cleanup is real but bounded.

## Violations

387 in the tree. Showing the first few.

**`ProjectSettings.tsx:56`**

```
icon: AhraOsIcon,
```

> Unsafe assignment of an `any` value

**`app/api/finance/transaction/[transactionId]/attachments/route.ts:42`**

```
const originalBytes = Buffer.from(await file.arrayBuffer());
```

> Unsafe assignment of an `any` value

**`app/api/finance/transaction/[transactionId]/attachments/route.ts:50`**

```
bytes: originalBytes,
```

> Unsafe assignment of an `any` value

**`app/api/spotify/callback/route.ts:35`**

```
const creds = JSON.parse(secretsApi.getValueOrThrow('SpotifyCredentials'));
```

> Unsafe assignment of an `any` value

**`app/api/spotify/callback/route.ts:36`**

```
const basic = Buffer.from(`${creds.clientId}:${creds.clientSecret}`).toString('base64');
```

> Unsafe assignment of an `any` value

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


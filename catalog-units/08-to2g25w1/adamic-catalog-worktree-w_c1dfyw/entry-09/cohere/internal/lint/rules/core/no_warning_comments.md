# `no-warning-comments`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **10** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow specified warning terms in comments

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

10 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/protocols/base/errors/BaseError.ts:237`**

```
// TODO: Remove the 403 fallback once base PR #859 ships (409 + ALREADY_AUTHENTICATED error code)
```

> Unexpected 'todo' comment: 'TODO: Remove the 403 fallback once base...'

**`libraries/structure/libraries/nexus/source/protocols/http/Cookies.ts:76`**

```
// TODO decide if we want this log or anythign else here
```

> Unexpected 'todo' comment: 'TODO decide if we want this log or...'

**`libraries/structure/source/api/graphql/forms/utilities/GraphQlFieldMetadataExtraction.tsx:190`**

```
// TODO: Remove this - hard coding this fix for now
```

> Unexpected 'todo' comment: 'TODO: Remove this - hard coding this fix...'

**`libraries/structure/source/components/notices/NotSignedInNotice.tsx:12`**

```
// TODO: Clean this up. Just a quick solution for now.
```

> Unexpected 'todo' comment: 'TODO: Clean this up. Just a quick...'

**`libraries/structure/source/modules/engagement/layouts/EngagementProvider.tsx:104`**

```
// TODO: Get the load time of the Next.js route
```

> Unexpected 'todo' comment: 'TODO: Get the load time of the Next.js...'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


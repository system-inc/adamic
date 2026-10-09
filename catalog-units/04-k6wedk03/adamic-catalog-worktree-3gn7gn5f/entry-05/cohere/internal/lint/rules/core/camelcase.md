# `camelcase`

| | |
|---|---|
| **Recommendation** | **Strong No** |
| Violations in ahra | **1208** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce camelcase naming convention

## Why this recommendation

Beyond overlapping the in-house naming rules, the violations are snake_case GraphQL and REST wire keys such as `before_direct` in useMessagesFeedRequest.ts, which must match the upstream API spelling per the Module-Scope Constants convention on values that cross a boundary.

> Reviewed against real violation sites and the house conventions in `CLAUDE.md`,
> which moved this from **No** to **Strong No**. The first pass classified rules by
> correctness-versus-style crossed with violation count, and that is blind to
> whether a rule fights a convention we chose on purpose.

## Violations

1208 in the tree. Showing the first few.

**`app/(os-layout)/_hooks/useMessagesFeedRequest.ts:194`**

```
before_direct: String(before.direct),
```

> Identifier 'before_direct' is not in camel case

**`app/(os-layout)/_hooks/useMessagesFeedRequest.ts:195`**

```
before_group: String(before.group),
```

> Identifier 'before_group' is not in camel case

**`app/(os-layout)/_hooks/useMessagesFeedRequest.ts:196`**

```
before_broadcast: String(before.broadcast),
```

> Identifier 'before_broadcast' is not in camel case

**`app/(os-layout)/os/wisdom/_components/WisdomRulingRow.tsx:64`**

```
out_of_band: {
```

> Identifier 'out_of_band' is not in camel case

**`app/(os-layout)/os/wisdom/_components/WisdomRulingRow.tsx:95`**

```
with_feedback: 'with feedback',
```

> Identifier 'with_feedback' is not in camel case

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


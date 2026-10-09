# `operator-assignment`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **50** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require or disallow assignment operator shorthand where possible

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

50 in the tree. Showing the first few.

**`app/(os-layout)/contacts/_components/ContactsListView.tsx:56`**

```
organizationCount = organizationCount + 1;
```

> Assignment (=) can be replaced with operator assignment (+=)

**`app/(os-layout)/contacts/_components/ContactsListView.tsx:59`**

```
personCount = personCount + 1;
```

> Assignment (=) can be replaced with operator assignment (+=)

**`app/(os-layout)/os/sensation/_components/SensationPerceptionRibbon.tsx:273`**

```
runLength = runLength + 1;
```

> Assignment (=) can be replaced with operator assignment (+=)

**`app/api/finance/connections/route.ts:82`**

```
unlinkedAccountCount = unlinkedAccountCount + 1;
```

> Assignment (=) can be replaced with operator assignment (+=)

**`app/api/finance/entity/[entityKey]/inventory/route.ts:88`**

```
lockedCount = lockedCount + 1;
```

> Assignment (=) can be replaced with operator assignment (+=)

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.


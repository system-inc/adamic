# `react/jsx-no-leaked-render`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **428** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow problematic leaked values from being rendered

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

428 in the tree. Showing the first few.

**`app/(os-layout)/_components/kingdom/OsKingdomImageCard.tsx:72`**

```
{properties.version.isCurrent && (
```

> Potential leaked value that might cause unintentionally rendered values or rendering crashes

**`app/(os-layout)/_components/kingdom/OsKingdomMemberAvatar.tsx:134`**

```
{properties.houseSigilUrl && (
```

> Potential leaked value that might cause unintentionally rendered values or rendering crashes

**`app/(os-layout)/_components/kingdom/OsKingdomMemberNode.tsx:188`**

```
{member.currentActivity && <p className="line-clamp-2 text-xs/snug content--2">{member.currentActivity}</p>}
```

> Potential leaked value that might cause unintentionally rendered values or rendering crashes

**`app/(os-layout)/_components/kingdom/activity/OsKingdomActivityToolUse.tsx:41`**

```
{canInspect && (
```

> Potential leaked value that might cause unintentionally rendered values or rendering crashes

**`app/(os-layout)/_components/profile/OsProfileFeed.tsx:54`**

```
{properties.feed.truncated && (
```

> Potential leaked value that might cause unintentionally rendered values or rendering crashes

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.


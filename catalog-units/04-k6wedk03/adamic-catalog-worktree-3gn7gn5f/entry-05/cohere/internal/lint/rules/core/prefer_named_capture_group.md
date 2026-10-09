# `prefer-named-capture-group`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **563** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Enforce using named capture group in regular expression

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

563 in the tree. Showing the first few.

**`app/(os-layout)/contacts/[id]/ContactHumanizeToken.tsx:15`**

```
.replace(/([a-z0-9])([A-Z])/g, '$1 $2')
```

> Capture group '([a-z0-9])' should be converted to a named or non-capturing group

**`app/(os-layout)/contacts/[id]/ContactHumanizeToken.tsx:15`**

```
.replace(/([a-z0-9])([A-Z])/g, '$1 $2')
```

> Capture group '([A-Z])' should be converted to a named or non-capturing group

**`app/(os-layout)/contacts/_components/ContactsListView.tsx:27`**

```
const match = birthDate.match(/(\d{4})?-?(\d{2})-(\d{2})/);
```

> Capture group '(\d{4})' should be converted to a named or non-capturing group

**`app/(os-layout)/contacts/_components/ContactsListView.tsx:27`**

```
const match = birthDate.match(/(\d{4})?-?(\d{2})-(\d{2})/);
```

> Capture group '(\d{2})' should be converted to a named or non-capturing group

**`app/(os-layout)/contacts/_components/ContactsListView.tsx:27`**

```
const match = birthDate.match(/(\d{4})?-?(\d{2})-(\d{2})/);
```

> Capture group '(\d{2})' should be converted to a named or non-capturing group

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


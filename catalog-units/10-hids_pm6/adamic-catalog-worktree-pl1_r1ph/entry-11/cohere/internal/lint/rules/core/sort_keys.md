# `sort-keys`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **27906** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Require object keys to be sorted

## Why this recommendation

Overlaps a rule we already enforce in-house, so it would report the same defect under a second name.

## Violations

27906 in the tree. Showing the first few.

**`LintConfiguration.ts:99`**

```
{ directory: 'libraries/structure/libraries/nexus', alias: '@nexus' },
```

> Expected object keys to be in ascending order. 'alias' should be before 'directory'

**`LintConfiguration.ts:100`**

```
{ directory: 'libraries/structure', alias: '@structure' },
```

> Expected object keys to be in ascending order. 'alias' should be before 'directory'

**`LintConfiguration.ts:101`**

```
{ directory: '.', alias: '@project' },
```

> Expected object keys to be in ascending order. 'alias' should be before 'directory'

**`ProjectSettings.tsx:13`**

```
identifier: 'ahra',
```

> Expected object keys to be in ascending order. 'identifier' should be before 'version'

**`ProjectSettings.tsx:15`**

```
ownerDisplayName: 'Kirk Ouimet',
```

> Expected object keys to be in ascending order. 'ownerDisplayName' should be before 'title'

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


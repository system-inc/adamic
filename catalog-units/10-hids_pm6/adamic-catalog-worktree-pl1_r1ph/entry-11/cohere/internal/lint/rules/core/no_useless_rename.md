# `no-useless-rename`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **3** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Disallow renaming import, export, and destructured assignments to the same name

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

3 in the tree. Showing the first few.

**`libraries/structure/source/layouts/RootLayout.tsx:25`**

```
import { getSupportedLocaleCode as getSupportedLocaleCode } from '@structure/source/localization/TranslationsServerSide';
```

> Import getSupportedLocaleCode unnecessarily renamed

**`modules/apple/contacts/ContactsApi.ts:170`**

```
return allContacts.map(({ _pk: _pk, _dbPath: _dbPath, ...contact }) => contact);
```

> Destructuring assignment _pk unnecessarily renamed

**`modules/apple/contacts/ContactsApi.ts:170`**

```
return allContacts.map(({ _pk: _pk, _dbPath: _dbPath, ...contact }) => contact);
```

> Destructuring assignment _dbPath unnecessarily renamed

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.


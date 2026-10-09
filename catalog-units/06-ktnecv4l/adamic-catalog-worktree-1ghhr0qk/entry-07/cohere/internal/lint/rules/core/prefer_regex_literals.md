# `prefer-regex-literals`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **3** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow use of the `RegExp` constructor in favor of regular expression literals

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

3 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/text/String.ts:526`**

```
const unicodeFormatCharacters = new RegExp('[\\u202A-\\u202E\\u2066-\\u2069\\u200B-\\u200F\\uFEFF]', 'g');
```

> Use a regular expression literal instead of the 'RegExp' constructor

**`libraries/structure/libraries/nexus/source/types/Constructor.test.ts:349`**

```
expect(isConstructor(new RegExp('test'))).toBe(false);
```

> Use a regular expression literal instead of the 'RegExp' constructor

**`libraries/structure/libraries/nexus/source/types/EnumLike.test.ts:271`**

```
expect(isEnumLike(new RegExp('test'))).toBe(false);
```

> Use a regular expression literal instead of the 'RegExp' constructor

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


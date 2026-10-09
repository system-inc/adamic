# `@typescript-eslint/no-unnecessary-type-conversion`

| | |
|---|---|
| **Recommendation** | **No** |
| Violations in ahra | **73** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow conversion idioms when they do not change the type or value of the expression

## Why this recommendation

Stylistic with a cleanup cost that is not obviously worth the convention it buys.

## Violations

73 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/protocols/email/Email.ts:21`**

```
let latestEmailContent = threadHtml.toString();
```

> Calling a string's .toString() method does not change the type or value of the string

**`libraries/structure/source/components/forms/fields/text-area/FieldInputTextArea.tsx:37`**

```
return fieldContext.state.value == null ? '' : String(fieldContext.state.value);
```

> Passing a string to String() does not change the type or value of the string

**`libraries/structure/source/components/forms/fields/text/FieldInputText.tsx:43`**

```
return fieldContext.state.value == null ? '' : String(fieldContext.state.value);
```

> Passing a string to String() does not change the type or value of the string

**`libraries/structure/source/components/maps/data/scripts/GenerateMapData.ts:372`**

```
return code === '-99' ? '' : String(code);
```

> Passing a string to String() does not change the type or value of the string

**`libraries/structure/source/components/navigation/pagination/PaginationControls.tsx:213`**

```
href={constructHrefWithExistingUrlSearchParameters(Number(properties.page) - 1)}
```

> Passing a number to Number() does not change the type or value of the number

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


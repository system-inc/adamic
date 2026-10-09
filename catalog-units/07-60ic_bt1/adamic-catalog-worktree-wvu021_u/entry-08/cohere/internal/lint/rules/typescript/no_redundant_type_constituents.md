# `@typescript-eslint/no-redundant-type-constituents`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **21** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Disallow members of unions and intersections that do nothing or override type information

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

21 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/validation/schema/Schema.ts:152`**

```
message?: string | RenderableValidationMessageType;
```

> 'unknown' overrides all other types in this union type

**`libraries/structure/libraries/nexus/source/validation/schema/Schema.ts:188`**

```
errorMessage?: TranslationTemplateType<TSlots> | RenderableValidationMessageType;
```

> 'unknown' overrides all other types in this union type

**`libraries/structure/libraries/nexus/source/validation/schema/Schema.ts:191`**

```
successMessage?: TranslationTemplateType<TSlots> | RenderableValidationMessageType;
```

> 'unknown' overrides all other types in this union type

**`libraries/structure/libraries/nexus/source/validation/schema/Schema.ts:205`**

```
TranslationTemplateType<TSlots> | ValidatorMessageOverridesInterface<TSlots> | RenderableValidationMessageType;
```

> 'unknown' overrides all other types in this union type

**`libraries/structure/source/api/graphql/forms/utilities/GraphQlFieldMetadataExtraction.tsx:126`**

```
graphQlFieldArray: GraphQlFieldInterface[] | unknown,
```

> 'unknown' overrides all other types in this union type

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


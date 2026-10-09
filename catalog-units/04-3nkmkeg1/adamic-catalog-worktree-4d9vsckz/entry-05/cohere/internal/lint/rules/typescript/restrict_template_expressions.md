# `@typescript-eslint/restrict-template-expressions`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **34** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Enforce template literal expressions to be of `string` type

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

34 in the tree. Showing the first few.

**`libraries/structure/libraries/nexus/source/security/cryptography/JsonWebToken.ts:125`**

```
throw new Error(`signJsonWebToken: unsupported algorithm '${options.algorithm}'. Supported: ES256, RS256.`);
```

> Invalid type "never" of template literal expression

**`libraries/structure/libraries/nexus/source/security/secrets/Secret.test.ts:35`**

```
expect(`value=${secret}`).toBe('value=[redacted]');
```

> Invalid type "Secret<string>" of template literal expression

**`libraries/structure/libraries/nexus/source/structured-text/json/JsonValueTransformers.ts:42`**

```
throw new Error(`Invalid date value: ${jsonValue}`);
```

> Invalid type "unknown" of template literal expression

**`libraries/structure/source/components/forms/survey/SurveySchemas.ts:179`**

```
dateSchema.minimum(new Date(component.minimumDate), `Date must be after ${component.minimumDate}`);
```

> Invalid type "string | Date" of template literal expression

**`libraries/structure/source/components/forms/survey/SurveySchemas.ts:182`**

```
dateSchema.maximum(new Date(component.maximumDate), `Date must be before ${component.maximumDate}`);
```

> Invalid type "string | Date" of template literal expression

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


# `@typescript-eslint/dot-notation`

| | |
|---|---|
| **Recommendation** | **Yes, with `allowProtectedClassPropertyAccess`, in place of the core `dot-notation`.** Bracket access is TypeScript's sanctioned way past `protected`, so a test reaching a protected hook keeps its brackets, and the dot form the core rule would propose is a type error (TS2445) |
| Measured precision | exact against typescript-eslint 8.71.0: all 63 of upstream's rows and 40 edge rows replayed through the installed rule, findings, spans and fix output identical |
| Plugin | `@typescript-eslint` |
| Auto-fixable | yes, the core rule's repair |
| Needs type information | yes |

## What it checks

What the core `dot-notation` checks: a bracket access with a literal key that is a valid identifier
reads better as a dot access, `a['b']` as `a.b`, and under `allowKeywords: false` a keyword after a
dot must be bracketed. The judgment and the repairs are the core rule's, shared through
`ecmascript/dotnotation`.

What it adds is one question, asked before a computed access is reported, and only when one of the
three typed options is on:

- `allowPrivateClassPropertyAccess` and `allowProtectedClassPropertyAccess`: the accessed property's
  first declaration carries `private` or `protected` as its first modifier. Decorators are not
  modifiers, so `@tracked private value` counts; `static private value` does not, because its first
  modifier is `static`.
- `allowIndexSignaturePropertyAccess`: no property is found, and the object's type has an index
  signature whose key is string-like (`string`, a template literal, `Lowercase<string>`). A number or
  symbol key does not count, and `any` has no index signature. A tsconfig that sets
  `noPropertyAccessFromIndexSignature` turns this on whatever the option says, because there the dot
  form is a type error.

The property is found as upstream finds it: the checker's symbol for the key, or failing that the
object's non-nullable type's property of the same name for a string key. So `x[('p')]` and
`` x[`p`] `` are judged like `x['p']`, while `x[null]` and `x[true]` never find a property.

## Options

| Option | Default |
|---|---|
| `allowKeywords` | `true` |
| `allowPattern` | `""` |
| `allowIndexSignaturePropertyAccess` | `false` |
| `allowPrivateClassPropertyAccess` | `false` |
| `allowProtectedClassPropertyAccess` | `false` |

An unknown key is refused, as upstream's schema does.

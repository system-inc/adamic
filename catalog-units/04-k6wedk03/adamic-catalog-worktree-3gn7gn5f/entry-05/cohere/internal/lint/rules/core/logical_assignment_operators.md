# `logical-assignment-operators`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **8** |
| Plugin | `eslint core` |
| Auto-fixable | yes |
| Needs type information | no |

## What it checks

Require or disallow logical assignment operator shorthand

## Where cohere differs from ESLint

Under `always`, cohere is silent inside any function React Compiler compiles: components and hooks, and the callbacks nested in them, decided by the same `IsInsideComponentOrHook` the react rules use. The compiler refuses all three shorthands (`Handle ||= operators in AssignmentExpression`, babel-plugin-react-compiler 1.0.0), so the rewrite this rule suggests would cost the function its compilation. ESLint keeps reporting there, and those are false positives we do not reproduce. Three ahra sites showed it: `UsersRolesPage.tsx:121`, `RestEndpointNodeContent.tsx:335` and `WebSocketViaSharedWorkerProviderInternal.tsx:309`.

The option is `reactCompiler` in the second element (`["always", {"reactCompiler": false}]`), and it defaults to on, because every tree we lint runs the compiler. A project that does not can turn it off to get those findings back. `never` is not gated: it expands the shorthand, which is the form the compiler accepts.

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

8 in the tree. Showing the first few.

**`libraries/structure/source/components/forms/fields/FieldMessage.tsx:104`**

```
shouldShowSuccesses = shouldShowSuccesses && !hasErrors && validationResults.successes.length > 0;
```

> Assignment (=) can be replaced with operator assignment (&&=)

**`libraries/structure/source/components/markdown/utilities/MarkdownUtilities.ts:58`**

```
node.data = node.data || {};
```

> Assignment (=) can be replaced with operator assignment (||=)

**`libraries/structure/source/components/markdown/utilities/MarkdownUtilities.ts:59`**

```
node.data.hProperties = node.data.hProperties || {};
```

> Assignment (=) can be replaced with operator assignment (||=)

**`libraries/structure/source/components/markdown/utilities/MarkdownUtilities.ts:189`**

```
current.data = current.data || {};
```

> Assignment (=) can be replaced with operator assignment (||=)

**`libraries/structure/source/components/markdown/utilities/MarkdownUtilities.ts:209`**

```
node.data = node.data || {};
```

> Assignment (=) can be replaced with operator assignment (||=)

### What fixing looks like

Auto-fixable. `eslint --fix` rewrites these, so the cleanup is a command plus a review of the diff.


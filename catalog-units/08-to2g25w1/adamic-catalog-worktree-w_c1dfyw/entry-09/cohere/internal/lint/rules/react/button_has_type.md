# `react/button-has-type`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **13** |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow usage of `button` elements without an explicit `type` attribute

## Where cohere is deliberately quieter than ESLint

Upstream is type-blind: any `type` that is not a literal or a ternary of literals is `complexType`.
cohere asks the checker first, and stays silent when the expression's type is a union whose every
constituent is a string literal the configuration permits. `complexType` exists because a computed
value could evaluate to something that submits a form; when the checker proves it cannot, the
finding is false, and reporting it forces either a suppression that lies or a ternary that only
placates the rule.

The real site is `libraries/structure/source/components/buttons/Button.tsx:222`,
`<button type={type}>` with `type = 'button'` destructured from a prop typed
`'button' | 'submit' | 'reset'`. ESLint reports it; cohere does not (13 findings before, 12 after).

The exemption is exactly as wide as the proof. It still reports when `undefined` is a constituent
(an optional prop with no default, where HTML's `submit` applies), on `string`, `any`, an error type,
a type parameter, a union with one invalid member, and a valid union containing a type the config
switches off.

Fixtures: `TestButtonHasTypeTrustsAProvenType` in `button_has_type_test.go` (five silent rows, the
first being the `Button.tsx:222` shape, and seven reporting controls). One upstream-measured row moved:
`` `bu${''}tton` `` folds to the literal `"button"` and is silent here, so the "interpolating template"
row in `TestButtonHasTypeReadsStaticValuesUpstreamsWay` now interpolates a `string`. ESLint trusts the
same proof through Nexus's `ReactButtonHasTypeRule`, which wraps this rule under its own name and runs
these rows verbatim (Nexus fc3a1cc), so the engines agree on the site (#cn8sthd).

## Why this recommendation

Stylistic. Small enough to adopt if we want the convention, not urgent if we do not.

## Violations

13 in the tree. Showing the first few.

**`libraries/structure/source/components/buttons/Button.tsx:218`**

```
<button ref={ref as React.Ref<HTMLButtonElement>} type={type} disabled={isDisabled} {...commonProperties}>
```

> The button type attribute must be specified by a static string or a trivial ternary expression

**`libraries/structure/source/components/containers/AccordionItem.tsx:54`**

```
<button
```

> Missing an explicit type attribute for button

**`libraries/structure/source/components/forms/fields/markup/lexical/LexicalFloatingLinkEditor.tsx:93`**

```
<button onClick={handleUpdateLink}>{translations.LexicalFloatingLinkEditor.update}</button>
```

> Missing an explicit type attribute for button

**`libraries/structure/source/components/forms/fields/markup/lexical/LexicalFloatingLinkEditor.tsx:94`**

```
<button onClick={handleRemoveLink}>{translations.LexicalFloatingLinkEditor.remove}</button>
```

> Missing an explicit type attribute for button

**`libraries/structure/source/components/markdown/ContentNavigator.tsx:127`**

```
<button
```

> Missing an explicit type attribute for button

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.


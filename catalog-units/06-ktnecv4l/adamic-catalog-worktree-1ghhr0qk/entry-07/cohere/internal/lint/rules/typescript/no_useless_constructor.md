# `@typescript-eslint/no-useless-constructor`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **1** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow unnecessary constructors

## Status: declined as a port, because it would register a second NAME rather than a second check

There is no `.go` beside this file, and that is the finding rather than an omission.

This rule is `getESLintCoreRule('no-useless-constructor')` plus two filters: an accessibility test
and a parameter test. `core/no_useless_constructor.go` already carries both, and its doc comment says
so with a measurement taken when it was ported.

**That claim was re-measured independently for this batch rather than inherited.** Upstream's
extension and upstream's core were driven over identical inputs on the installed 8.67.0 / 10.8.1
builds, across the extension's own 41-case corpus:

    total 41   identical 41   divergent 0   config problems 0

Zero divergence, so the extension narrows nothing the core does not already narrow here. Our core
rule reproduces that corpus exactly, and the whole core package is green.

The remaining work a port would represent is the NAME, and nothing asks for it: the config enables
`no-useless-constructor` at `CohereSettings.json:520` and never mentions the namespaced spelling. A
wrapper would also need this rule's 427-line body lifted out of the rule package, because a rule
package may not import another rule package, for no behavioural change.

## Re-verified 2026-09-06, in a later batch

This rule was dispatched again as part of a four-rule type-aware batch, and the decline was checked
rather than inherited, since the document itself says the zero is a fact about two installed versions
and not a property of the rules.

Nothing had moved. The core rule `no-useless-constructor` is still registered, still enabled at
`CohereSettings.json:520`, and its whole fixture set is green. The namespaced spelling still appears
nowhere in the config. `cohere --rules` lists the core name and not the `@typescript-eslint/` one,
which is the state this document describes.

So the decline stands and no code was written. The other three rules in that batch --
`no-inferrable-types`, `prefer-regexp-exec`, `consistent-type-definitions` -- were ported.

## If this is revisited

Re-run the differential before assuming the zero still holds; it is a fact about two installed
versions, not a property of the rules. The harness used is an ESLint `Linter` driven with the
typescript-eslint parser and a project tsconfig, refusing any message with no `messageId` so a
configuration complaint cannot be counted as a finding.

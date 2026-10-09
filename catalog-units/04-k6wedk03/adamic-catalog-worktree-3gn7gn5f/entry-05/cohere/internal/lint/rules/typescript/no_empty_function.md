# `@typescript-eslint/no-empty-function`

| | |
|---|---|
| **Recommendation** | **Maybe** |
| Violations in ahra | **45** |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow empty functions

## Status: declined as a port, because it would register a second NAME rather than a second check

There is no `.go` beside this file, and that is the finding rather than an omission.

This rule is `getESLintCoreRule('no-empty-function')` with four extra `allow` values layered over it.
`core/no_empty_function.go` already carries all four, and its doc comment says so with a measurement
taken when it was ported.

**That claim was re-measured independently for this batch rather than inherited.** Upstream's
extension and upstream's core were driven over identical inputs on the installed 8.67.0 / 10.8.1
builds, across the extension's own 16-case corpus:

    total 16   identical 14   divergent 0   config problems 2

The two rows that are not identical are not behavioural. They are the cases spelling an allow value
in kebab case, `private-constructors` and `protected-constructors`, which eslint core REFUSES at
config load rather than answering differently:

    Value "private-constructors" should be equal to one of the allowed values.

So the only real difference between the two rules is the spelling of two option values, which is what
the core rule's own doc comment already records. Every case where both configurations load agrees.

The remaining work a port would represent is the NAME, and nothing asks for it: the config enables
`no-empty-function` at `CohereSettings.json:516` and never mentions the namespaced spelling.

## If this is revisited

The kebab spellings are the one thing a consumer might want. If a project needs to write
`"private-constructors"` in `CohereSettings.json`, that is a change to `DecodeNoEmptyFunctionOptions`
accepting both spellings, not a new rule.

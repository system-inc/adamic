# `nexus/correctness-no-discarded-pure-result`

| | |
|---|---|
| **Recommendation** | **Yes, at `error` now.** 0 findings: a guard |
| Findings | **ahra 0** (measured 2026-10-03; the coverage report lists it under "ran and found nothing" over all 3,795 files) |
| Measured precision | nothing to measure on the tree; the fixtures carry it |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to tie the method to its declaration in the default library and to read each argument's type; without a checker the rule registers nothing |

## What it checks

An expression statement that is a call of a side-effect-free method of the default library's `String`,
`Array` or `ReadonlyArray` (`name.trim();`, `path.replace(a, b);`, `list.concat(more);`,
`list.slice(1);`), decided by every declaration of the method's symbol, with no argument that is a
function by type. The doc comment lists the method sets and what was left out on purpose (`normalize`,
`repeat`, `localeCompare`, `toLocaleString`, which can throw and so could be a validation; every
callback array method, which is a style question).

## Where it came from

A guard: zero in ahra when built. Found by the cross-language and JavaScript-catalog passes of the
new-rules sweep (`#tevhg3f`, after staticcheck SA4017, go vet `unusedresult`, sonarjs `no-ignored-return`
and unicorn `no-unused-builtin-method-return`), built in `#j03vwm6`. The core `no-unused-expressions`
does not catch it, because a call is always allowed there.

## Reconciling with the research count

Research counted **0**, cohere counts **0**. A line-oriented text probe over every `.ts`/`.tsx` in ahra
for `receiver.method(...);` with a pure method name returned 160-odd lines, every one read as either the
last line of a multi-line expression (`].join('\n');` closing a `return [...]`, a `&&` chain) or a
receiver that is not a string or array (`router.replace(...)`, `tasksSelection.replace(...)`,
`stopWatch.split(...)`). None is a statement that drops a pure result.

## Verification

- The slip fires and its fix is silent: `name.trim(); path.replace(...); list.concat(more);` reports
  three, the same three with results kept reports none. 8 more firing shapes (optional call,
  parentheses, readonly array, a `string | string[]` union, a chain), 12 silent shapes.
- Mutation check, all killed: function arguments ignored (a callback, a named replacer, a replacer of
  type `any`), only the `any` part of the callable test removed (a replacer of type `any`), any interface
  accepted (a mutating method, a method that can throw), any file accepted (a local interface named
  `Array`).

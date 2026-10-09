# `nexus/correctness-no-global-listener-target-assertion`

| | |
|---|---|
| **Recommendation** | **Yes, at `error`.** Zero findings today; it guards a shape each new component can write fresh |
| Findings | **ahra 0** (measured 2026-10-03; research count 0, and 1 in `Code.tsx` before its fix) |
| Measured precision | no findings to read |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve `document`, `window`, `addEventListener`, `Event.target` and the element types to lib.dom; without a checker the rule registers nothing |

## What it checks

An `as T` or `<T>` around `event.target`, where `event` is the first parameter of a handler passed to
`document.addEventListener` or `window.addEventListener` and `T` is a lib.dom element type narrower than
`Element`, `HTMLElement`, `SVGElement` or `MathMLElement` (`HTMLTextAreaElement`, `HTMLInputElement`,
`SVGPathElement`). A listener that high up hears every element's events, so the target is whatever was
clicked or focused, and a member read through the assertion is `undefined` on any other element. The handler
is an inline function or arrow, or a function declaration or `const` function in the same file. Everything
is matched by declaration, so a local `document` or a project's own element interface is not read.

## Where it came from

Structure's `source/components/code/Code.tsx`, found uncommitted by the own-history pass of the new-rules
sweep (`#tevhg3f`, item 2), built in `#j03vwm6`. A `keydown` listener on `document` read `event.target as
HTMLTextAreaElement` to insert spaces on Tab: Tab on any other control was swallowed, and with no caret on
that control `code.substring(undefined)` duplicated the editor's contents. The fix is a React `onKeyDown` on
the textarea reading `event.currentTarget`.

## Existing rules checked

`@typescript-eslint/no-unsafe-type-assertion` reports this assertion too, along with every other narrowing
assertion: the sweep measured 2,747 findings on ahra with it on, so it is off and cannot be adopted at zero.
No unicorn or sonarjs port exists in cohere.

## What it leaves alone

- `as HTMLElement`, `as Element`, `as Node` and the other general types: three in ahra (`TasksCenter.tsx`,
  `TaskDetailFocusOverlay.tsx`, `SourcesAccordion.tsx`), each followed by `.closest()`, all benign.
- An assertion the checker already agrees with: when the narrowed type of `event.target` (after an
  `instanceof` guard) is assignable to `T`.
- Listeners on anything else (`textarea`, `document.body`, `ref.current`): the target is constrained to that
  subtree.
- An assertion to a project's own interface (a custom element): missed, never false.

**One judgment for review.** A guard the checker cannot see, such as `tagName === 'TEXTAREA'` or
`event.target === textareaReference.current` before the assertion, still reports. The code is correct at
runtime there; the assertion is still unchecked by the types, and `instanceof HTMLTextAreaElement` says the
same thing and narrows. If that is judged a false positive, the narrowing exemption is where to widen it.
There are no such sites in ahra today.

## Verification

- Fixtures from the real site both ways: the `Code.tsx` Tab handler on `document` (one finding) and the fixed
  shape on the textarea (silent). 6 more firing shapes (window, arrow, the `<T>` spelling, `!`,
  `as unknown as`, a `const` handler, a nested callback, a union of specific types, an SVG type); 18 silent
  shapes (`HTMLElement`, `Element`, `Node`, `SVGElement`, an `instanceof` guard, a textarea's own listener,
  `ref.current`, `document.body`, a parameter named `document`, `currentTarget`, a nested handler's own
  `event`, a `let` handler, `as Window` on a message, a function that is not a listener, a parameter typed
  `{ target: unknown }`, a handler that is not a function), a custom element interface, a local `window`, and
  a project declaration merged into `Document.addEventListener`.
- Mutation check, all 11 killed: the global receiver, `addEventListener` from the default library, the event
  parameter's symbol, `Event.target` from the default library, the four general types, the `Element` base
  walk, element types from the default library, the narrowing exemption, identifier handlers, and the `const`
  requirement.
- The built binary reports the planted shape on a scratch project with ahra's compiler options.

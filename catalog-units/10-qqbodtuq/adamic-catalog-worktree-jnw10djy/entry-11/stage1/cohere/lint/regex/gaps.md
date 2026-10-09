# Regex blockers for @system_adamic_library

## Dynamic RegExp source is refused

Reproducer: `testdata/dynamic_gap.a`, unchanged on source Node:

```ts
import { programArguments } from 'adamic';
const pattern = programArguments()[0] ?? 'TODO';
console.log(`${new RegExp(pattern, 'u').test('TODO')}`);
```

Node prints `true` with argument `TODO`. Native lowering reports:

```
stage 0 can't lower RegExp with a nonconstant pattern yet
```

`TestDynamicPatternGap` proves both observations. `options.a` preserves the requested
constructor contract and is deliberately isolated from the supported literal module.
The CSS printer's constructors use constant imported pattern data, so its success
does not establish support for runtime rule options. No matcher fallback was added.

This prevents id-length's `exceptionPatterns`, no-inline-comments' `ignorePattern`,
and no-warning-comments' generated configurable terms/decorations from following
the required RegExp path on native. Their full finding migrations were not made.

## Raw option dialect is not Go regexp

Run `TestOptionDialectGap`. Every listed Go pattern is accepted by Go regexp.

| Pattern | Input | Go | `new RegExp(pattern, 'u')` on Node |
| --- | --- | --- | --- |
| `\s` | U+00A0 | false | true |
| `a$` | `a` followed by LF | false | false |
| `(?i)todo` | `TODO` | true | SyntaxError |
| `\p{Greek}` | `α` | true | SyntaxError |
| `a\z` | `a` | true | SyntaxError |
| `(?P<word>a)` | `a` | true | SyntaxError |

The last four have port-time JS spellings (`/todo/iu`, `/\p{Script=Greek}/u`,
`/a(?![\s\S])/u`, `/(?<word>a)/u`). Raw arbitrary options cannot be translated
once at port time: the source does not exist until configuration is read.
A constructor implementation that disagrees with Node is not an acceptable fix.
The fleet needs a decision on restricting options to a proven common dialect,
or changing the requirement for runtime translation. This unit does neither.

## Harness integration, #zmh9v36

Current main has no directory registration foundation. Its lint tests copy a fixed
`portFiles` list, which excludes `regex/patterns.a`. Adding a table import to the
existing `comments.ts` would break those tests. No shared harness files were edited;
the existing no-warning-comments matcher remains intact. The table exports the
fixed `warningSelfDirective` and `inlineCommentDirective` literals for integration.

The parked `codex/lint-wave1-06` branch was not changed by this unit.

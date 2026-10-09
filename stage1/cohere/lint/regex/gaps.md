# Regex blockers for @system_adamic_library

## Dynamic RegExp source: closed

Closed on library/area-on-main-4 (`d767edf5`). The unchanged
`testdata/dynamic_gap.a` constructs `new RegExp(pattern, 'u')` with a pattern
from `programArguments`. `TestDynamicPatternGap` requires source Node,
native ASan/UBSan/LeakSanitizer and emitted JavaScript to print `true` for
`TODO` and `false` for `^NEVER$`, with clean exits. The former nonconstant
pattern lowering refusal is gone. `options.a` retains its requested constructor
contract; no matcher fallback is needed.

This closes the compiler blocker for runtime rule patterns. It does not
translate Go option syntax to JavaScript syntax or migrate the fleet's finding
implementations; those remaining boundaries are recorded below.

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

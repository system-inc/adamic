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
the required RegExp path on native. Their constructor migrations and finding comparisons now exist; see MIGRATIONS.md.

## JavaScript option oracle migration, #7mztrdd

Ruling from @system_cohere_lint: option patterns are JavaScript by contract.
The earlier RE2-versus-JavaScript option comparison used the wrong oracle.
Cohere's option-consuming rules will migrate to its JavaScript-semantics
`esregexp` package under #7mztrdd. There is no requirement to translate arbitrary
Go-only syntax in user options. `new RegExp(pattern, 'u')` remains the native contract.

`TestOptionPatternsAgreeWithESRegexp` now compares unchanged user patterns against
`cohere/internal/lint/ecmascript/regexp.Compile(source, "u")` and independent Node.
Nine controls agree, covering Unicode whitespace, anchors, named captures,
lookahead, lookbehind, backreferences, line terminators and invalid syntax.
The Go adapter is added through an owned virtual-file overlay so no cohere source
or shared lint oracle is edited. The constant-pattern Go regexp table is unchanged.

### Pinned esregexp property escape gap

`TestOptionPropertyGap` retains this separately, excluded from the nine agreement
controls. With source `\p{Script=Greek}`, flags `u`, and input `α`, Node prints
`true`. Pinned cohere `esregexp.Compile` returns an error, reported by the observer
as `SyntaxError`. The pair is recorded in `testdata/option_property_gap.json`.
Reproduce with the owned test; no translation, fallback or engine change was added.
When this gap closes the gap test fails and tells the worker to put the case back
into the option agreement corpus. This is a property of the pinned esregexp package,
not proof that the JavaScript option contract needs changing.

The runtime compiler is approved as Codex 01a114e3 on codex/regex-runtime-compiler.
Both emitted JavaScript and sanitized native remain pending because the shared
lowerer rejects nonconstant patterns before backend selection. Tests activate both
legs automatically when lowering succeeds. #7mztrdd remains the Go consumer
integration dependency; captures use an explicit esregexp overlay. The property
mismatch remains named and no fallback was added.

## Harness integration, #zmh9v36

Current main has no directory registration foundation. Its lint tests copy a fixed
`portFiles` list, which excludes `regex/patterns.a`. The migrated `comments.ts` now imports the owned regex modules, so that shared
copy list needs integration. No shared harness files were edited; the production
no-warning-comments hand matcher has been replaced. The table exports the
fixed `warningSelfDirective` and `inlineCommentDirective` literals for integration.

The parked `codex/lint-wave1-06` branch was not changed by this unit.

Built: native no-label-var plus partial native regex flag/front-end and character-sequence logic; all new Adamic source uses .a.
Commits: claim 0a4e2f177d20799dd045b6ed2cf99ba9f18e548f was pushed before code; this report accompanies the implementation commit.
Commands and outputs: label verify.py, invalid_regexp/verify_frontend.py, verify_helpers.py and Go bridge package tests pass within the scope below.
Mutants: label membership inversion, value-to-variable scope mask, u/v priority removal, modifier lower-bound shift and retained released handle are each caught.
Not covered: complete regex ports, shared registration/profile integration, one undefined-label parser fixture, emitted-JavaScript comparisons and the repository-wide gate.

## Selection and ownership

The original three wave-09 ports remain complete and pushed at 03ea146b with
validation in d0e9ba7c. The previous Nexus continuation was withdrawn after an
earlier concurrent claim; its archived experiment is unchanged.

Fetched all origin heads before this selection: 335 remote refs, 33 distinct
Markdown claim blobs and 114 ranked rule names in claims. Excluded actual native
ports on main ef3d907e and bridge 5afbdb83 and all claim mentions. The first three
remaining combined-volume entries with lexical ties were no-invalid-regexp,
no-label-var and no-misleading-character-class, each zero in both corpora.
Claim 0a4e2f17 was pushed, then a second all-head fetch found no other reservation.
The final refresh has 342 refs and finds a later wave-14 reservation, 350d4776,
01:22:58 UTC, after this claim at 01:20:27 UTC. Evidence preserves both claims.

All new implementation and validation files are in this owned directory.
No shared generator, harness, parser, compiler emitter, lowerer or oracle file was
edited. The new checker question is exercised through a scratch Go overlay,
never by changing the checkout's dispatch. Integration instructions are in
label_var/registration.txt. Its Go implementation returns raw GetSymbolsInScope
names and flags; the native rule decides the lint result.

## Complete available label-rule comparison

Run after sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/typeaware/wave09_core/label_var/verify.py > /workspace/wave-09-core-label-test.log 2>&1
go test -overlay /workspace/wave-09-core/checker-overlay.json ./bridge/tsgo/checker ./bridge/tsgo/archive > /workspace/wave-09-core-bridge-tests.log 2>&1
```

The isolated native driver parses actual sources and asks the raw checker question.
An independent loader/walker invokes unmodified production Go NoLabelVar.Run.
It compares every serialized rule name, message ID/text, byte range, fix and
suggestion, rather than counts. This rule has no fixes or suggestions, which are
explicitly serialized as zero. The scopes include hoisting, globals, functions,
classes, type-only bindings, enums, namespaces, catches, destructuring, imports,
Unicode and CRLF.

| Inputs | Files | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Supported controls | 34 | 18 | 7700 |
| TypeScript compiler, pinned 050880ce | 77 | 0 | 7859 |
| Frozen repository corpus already used by wave 09 | 287 | 0 | 18485 |

All three sets match under normal and address/undefined/leak sanitizer builds.
One Go-valid upstream control is refused by the shared native parser:
`undefined: for(;;) { break undefined; }` produces panic 70, expected semicolon
at offset 9. The verifier records the source and error in parser-rejections.json
and excludes it explicitly. Therefore this is not complete upstream-fixture
coverage. Both required corpora are compared in full, with no corpus exclusions.

Three whole-process runs load/check/parse/lint/serialize on each side, comparing
full output on every repetition. Median seconds:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 1.448450 | 0.330366 | 4.38 |
| Repository | 0.211157 | 0.154313 | 1.37 |

The native implementation is slower here; these are process measurements, not
checker-subtracted lint-only timings. Setup was already completed in this same
workspace: cloud/setup.sh reported 88 seconds, nproc 5; its earlier wave report
preserves the setup evidence. This continuation reuses that installed environment.

Released-handle inspection refuses with exactly panic 70 and
`adamic: panic: invalid or released checker handle`, under normal and sanitizer
builds. A scratch release-table mutant retains the handle and exits 0 instead.
The supplemental sanitizer probe is built from /workspace/wave-09-core/released.a
with --tsgo /workspace/wave-09-core/checker-asan.a --sanitize. Bridge checker tests
pass; archive has no tests. No repository-wide gate was run.

## Partial no-invalid-regexp

no_invalid_regexp.a contains constructor selection, global-symbol declaration
checks, string-literal argument rules, option-provided flag sets, duplicate/unknown
and u/v precedence, canonical compiler flags, pattern error formatting and the
unknown-flags dual-compile decision. The compiler and rune formatter are explicit
constructor dependencies, so no guessed compiler result can silently suppress a
finding. A missing compiler raises a named panic.

```
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_frontend.py > /workspace/wave-09-core-frontend-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/verify_helpers.py > /workspace/wave-09-core-helper-final-test.log 2>&1
```

Sixteen actual flag-only source controls match the complete production Go rule
on 11 findings and 2040 serialized bytes, normal and sanitizer. They exercise
parenthesized constructors, new/call forms, global/type-only/local bindings,
unknown patterns, Unicode and CRLF. The test compiler callback panics if reached:
these controls deliberately do not test pattern compilation. The quote callback
covers only the tested printable flag characters, not the complete Go %q domain.

Sixty standalone flag cases, including extra flags and astral flag characters,
match the production private invalidFlagsMessage helper on 2843 bytes, normal
and sanitizer. The oracle shim calls that helper; it does not replicate the
native judgment. Quote spellings are explicit test inputs. Missing native
esregexp.Compile semantics and exact error strings, full rune quoting and option
schema decoding remain blockers. Pattern-message formatting and compile-flag
helpers compile but have no dedicated runtime comparison yet. No full compiler
or repository comparison is claimed for this partial rule.

## Partial no-misleading-character-class

no_misleading_character_class.a contains all six production sequence judgments
and messages: combining marks, emoji modifiers, regional indicators, joined glyphs,
surrogate pairs with code-point escapes and pairs without Unicode flags. Inputs
explicitly carry character values, source spans and escape provenance. The output
includes the production suggestion-eligibility mark, not a constructed suggestion.
The conservative pattern-meaning guard used before suggesting u is also ported.

The helper verifier compares 3282 synthetic sequences, spanning range boundaries,
escaped and code-point-escape variants, and separated/chained joiners, plus 16
meaning-change patterns. The Go shim calls unchanged misleadingSequenceFindings
and patternMeaningChangesUnderUnicodeFlag. Findings compare exact ranges, order,
IDs, descriptions and eligibility. These are helper frames, not full rule findings,
fixes or suggestions. Normal and sanitizer builds match. Helper process timings
are saved but must not be read as native/Go full-rule performance.

Missing native regexsyntax class parsing, constant-expression/reference tracking
for RegExp aliases and strings, cooked-to-raw source maps and complete Unicode-flag
suggestion edits prevent the full rule. No shared adapter or registration file was
changed. No full-corpus or emitted-JavaScript parity is claimed for this rule.

## Mutants and evidence

Every mutant builds successfully and its normal invocation exits 0 with empty
stderr. Only comparison to production Go catches the semantic mutants:

- Reverse label scope membership: controls differ.
- Scope checker mask Value -> Variable: function/class bindings disappear; controls differ.
- Remove u/v mutual-exclusion priority: flag helper bytes differ.
- Shift emoji-modifier lower bound from 1f3fb to 1f3fc: sequence helper bytes differ.
- Retain a released checker handle: it succeeds where the required panic check expects 70.

validation/mutants.json records hashes and first differing byte for the semantic
mutants. Compressed outputs preserve the normal/sanitizer/oracle comparisons.
Validation logs, frozen manifests, measurements, claim audit and parser rejection
are saved under validation. Scratch overlays are evidence only; Go copies use
.go.txt to avoid accidental package discovery. Further claims are not taken while
these two regex ports remain incomplete.

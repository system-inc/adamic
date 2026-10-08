# ecmascript/directives package claim

Branch: lint-helpers/ecmascript-directives. Base: origin/area/stage1-lint at 334509eea.
Triage d7ab0bc4, rank 24: zero retained ports, ten missing helpers; build fresh.
All 1,523 origin refs audited for helper and consuming-rule paths; all claims
on 43 origin lint-helpers/ and codex/lint-helpers* branches inspected. No
shared directives implementation or reservation exists. Earlier eligible
packages are reserved or delivered; tonight exclusions honored. Dotnotation
is skipped because CompileAllowPattern compiles runtime regexes. Generic
strict-options is already delivered; its missing contracts are rule-local.
This package performs no runtime regex compilation.

Helpers: ParseDisable, ParseEnable, Recognize, endsWord, isLineComment,
parseRuleNames, splitDirective, splitReason, splitScope, stripCommentMarkers.
Consumer unblocked: @eslint-community/eslint-comments/require-description.
Forecast: one rule alone / 133 cumulative with preceding complete packages;
conditional helper readiness, not a findings-parity claim. Suppression also
consumes the grammar. Existing rule attempts on codex/lint-wave1-02, -05 and
-06 are candidates for adaptation, not a shared package port.

Claim before code; fetch again and yield to an earlier competing reservation.
One helper per file; capture every actual use in consuming upstream tests and
compare unchanged Go against source Node, emitted JavaScript and sanitized
native. One compiling semantic mutant per helper, caught on Node and native.
Prove the consumer with its own rule mutant and run complete helper/lint input
sets before one finished-unit push. Format and check every changed Go file.

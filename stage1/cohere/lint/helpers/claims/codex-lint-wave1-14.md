# Helpers from lint wave slot 14

Branch: codex/lint-helpers-from-codex-lint-wave1-14. Base: origin/codex/lint-helpers.
Previous rule work pushed through 9d2c673b on codex/lint-wave1-14.
Fetched every origin head and inspected all five unique helper claim files
across 389 origin refs. The highest unclaimed concrete symbol count is six.
The seven-consumer strict-option label is rule-local work covered by the existing
option prerequisites, not a new named helper. Comments remain reserved by the
original helper bundle.

Claim: github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.leadingInteger
File: slot14/leading_integer.a. Six remaining consumers, zero final blockers alone:
- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Preserve four-byte whitespace set, optional sign, ASCII leading digits,
absence versus zero and Go 64-bit wrapping arithmetic. Represent the integer
exactly rather than as an unsafe JavaScript number. Compare real Go on all six
consumer fixture sources plus integer, Unicode and overflow controls, source
Node, emitted JavaScript and sanitized native, with a compiling mutant.
This claim is pushed before implementation.

## Race correction and replacement

Withdraw leadingInteger: slot 05 claim 652db0c7 at 02:30:32 UTC precedes
our 2b2947ff at 02:30:56 UTC. No duplicate implementation is delivered.
Bounded comparisons passed but live consumer coverage is blocked by absent
Kirk-local Tailwind installation and theme fixtures. Preserve that evidence
without counting its prerequisite removals.

Refetched all 404 origin refs and inspected all 16 distinct helper claim blobs.
Replacement claim: github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.nodesFromStaticDeclarations
File: slot14/nodes_from_static_declarations.a. Six consumers, the same six
Tailwind rules above, zero final blockers alone. This ties the highest unclaimed
concrete-symbol count. Preserve sequence, every property/value/presence/important
field, empty input, fresh output list and fresh declaration allocation. Real Go
framework declarations and independent controls decide behavior. Push before code.

## Second race and final outcome

Withdraw nodesFromStaticDeclarations: slot 09 claim 4d352341 at 02:35:21 UTC
precedes our 71e76328 at 02:36:10 UTC. Final refresh inspected 416 origin refs
and 17 distinct helper claim blobs. Both duplicates are excluded from executable
delivery and counted as zero prerequisite removals. Sources are archived as
.a.txt witnesses solely to reproduce their external comparisons. No active
helper reservation remains. No other helper is claimed.

Stop on the observed fixture blocker: all six original consumer suites recorded
zero live calls. Canonical/unknown guards fail because the pinned Go fixtures
require a missing Tailwind installation and Kirk-local theme.css under
/Users/kirkouimet/Projects/ahra/app/_theme/styles. Four suites exit zero through
skips, which is not a live consumer parity pass. Reproduction, independent
Go/helper evidence, mutants and exact limits are in ../slot14/REPORT.md.

## First retained helper claim, after parking

Both owned branches are rebased onto current main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, re-green and pushed: rule branch ad3fcb3c403721167fe2b41f95e9e3589ed8ced0, helper evidence branch ab302a57652659628f78640cd4cae2e8d9826a97. The rule branch is parked with shared context/registration/Diagnostic blockers named in its owned PARKED.md. No shared harness is edited.

Claim: github.com/system-inc/cohere/internal/lint/ecmascript/regexp.decodeFixedHex
File: slot14/regexp_decode_fixed_hex.a.

Refreshed all 580 origin refs and inspected 20 distinct helper claim contents, plus the delivered comment reservation in HELPERS.md. This concrete symbol ties the highest remaining unclaimed fan-out at four. Preserve Go byte offsets, fixed-digit ParseUint acceptance and 32-bit bounds, decoded rune/width/default fields, Unicode errors and non-Unicode identity fallback. Byte-backed inputs preserve UTF-8 behavior without confusing byte and UTF-16 offsets. Compare actual private Go over every consuming rule suite and boundary controls on source Node, emitted JavaScript and sanitized native, with a compiling semantic mutant. Push this claim before writing code.

Consumers, four dependency entries and zero complete rule blocker sets removed alone:
- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

## Second retained helper

Both owned branches are now rebased onto origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad, re-green and pushed: parked rules 48b945b3702152030c2ad6c3c3a278750b305397 and retained helper 804b2979bac2156fa3cdb81ae124cd114648a73c.

Claim: github.com/system-inc/cohere/internal/lint/ecmascript/regexp.decodeUnicodeEscape
File: slot14/regexp_decode_unicode_escape.a.
Refetched all 610 origin heads and inspected all 20 distinct recursive helper claim contents plus HELPERS.md reservations. This symbol ties the highest unclaimed concrete fan-out at four and uses the already delivered fixed-hex helper. Push before code. Preserve fixed Unicode escape delegation, Unicode-only brace syntax, first closing brace, exact byte width, uint32 parsing and MaxRune bounds, exact error bytes and decoded defaults. Compare the actual private Go helper across all four consumer suites, captured targeted paths and bounded controls, source Node, emitted JavaScript and sanitized native with a compiling semantic mutant.

Consumers, four additional dependency entries, zero complete blocker sets removed alone:
- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

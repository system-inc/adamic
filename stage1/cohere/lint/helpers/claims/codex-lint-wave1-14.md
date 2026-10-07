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

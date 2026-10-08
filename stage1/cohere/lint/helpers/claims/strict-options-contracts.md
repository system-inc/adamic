# Strict option decoding and schema validation contracts claim

Branch: lint-helpers/strict-options-contracts, base origin/area/stage1-lint.
Triage d7ab0bc4 rank 13: earlier eligible packages reserved, including rules/core at 351d296c (earlier than withdrawn local f3bad7d2).
Existing generic strict-options primitives are landed; this claim covers only missing per-rule decoding/default/schema contracts, not duplicate generic primitives.
Forecast: two rules alone, 107 cumulative with earlier complete packages.

Consuming rules:
- boundaries/dependencies
- id-length
- nexus/import-require-path-alias
- no-constant-condition
- no-extra-boolean-cast
- no-restricted-exports
- object-shorthand

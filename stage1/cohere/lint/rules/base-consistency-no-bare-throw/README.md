# Bare throws

Claimed in 1f3be00e before implementation. The rule declines capture/declaration/
vocabulary/test paths, normalizes separators, checks only a bare built-in new
expression thrown directly, leaves AggregateError and qualified constructors
alone, and reports the new-expression range with exact policy wording. No fixes.

The private capture adapter preserves original filename suffixes and path
segments, and its deduplication key includes filename. All rule fixtures with
valid parseable source are compared. Three intentionally malformed Go recovery
probes cannot pass the fix engine and are explicitly excluded. The rule adapter
returns the real Go rule; it does not normalize its findings or fixes.

prepare-validation.py only writes scratch copies and a Go overlay. Shared
registration, parser, runtime, test harness and generated lists remain unedited.
REPORT.md records the complete three-rule unit and older reservation recovery.

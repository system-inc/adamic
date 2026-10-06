# Lint wave 1 slot 15

Owner: codex/lint-wave1-15.

The ordered handoff is in HELPERS.md, referenced by helpers/REPORT.md:

43. nexus/import-require-node-namespace
44. structure/network-no-invalidate-cache-literal-key
45. structure/network-no-string-literal-query

All fetched origin branch stage1 trees were checked for implementation filenames;
none of these three rules was found. No rule is skipped.

Claim pushed before implementation. New Adamic files use .a.

Foundation merges were not both clean: helpers conflicted with registration in
six shared driver files. The directory registration versions were retained;
standalone helpers and inventory were merged. docs/parallel-work.md is absent
from main and both foundation branches; docs/lint-registration.md supplies the
available rule-directory contract.

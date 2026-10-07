# typescript-no-this-alias

Matches declaration and compound-assignment aliases, asymmetric target/initializer unwrapping, TypeScript filename gates, allow lists and decoded ReportDestructuring options. All 40 upstream cases matched. Manifest options are already decoded Go NoThisAliasOptions, as required by directory registration; arbitrary configuration-decoder acceptance is not certified.

Own implementation and helpers use `.a`, with a directory descriptor, unchanged upstream Go adapter, raw witnesses and a semantic mutant.

Default registration still insists on `rule.ts`; the shared profile test also fails to compile against the registration foundation. The two non-null rules additionally use `suggestion-edits:<upstream message id>` with a byte-exact ordered edit list because the inherited driver accepts only one edit over the finding range. The scratch Go driver independently serializes actual upstream edits into that protocol. No shared source was edited. These are comparison-validated candidates pending infrastructure integration.

See [the evidence report](../../claims/wave1-05-second-evidence/REPORT.md) for reproduction, the full source corpus, mutants, throughput and limits.

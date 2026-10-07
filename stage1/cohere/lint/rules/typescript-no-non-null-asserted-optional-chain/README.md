# typescript-no-non-null-asserted-optional-chain

Distinguishes assertions over whole optional chains from mid-chain assertions. Preserves parenthesis handling, optional-chain roots, suggestion message id and removal byte range. All 24 upstream cases matched.

Own implementation and helpers use `.a`, with a directory descriptor, unchanged upstream Go adapter, raw witnesses and a semantic mutant.

Default registration still insists on `rule.ts`; the shared profile test also fails to compile against the registration foundation. The two non-null rules additionally use `suggestion-edits:<upstream message id>` with a byte-exact ordered edit list because the inherited driver accepts only one edit over the finding range. The scratch Go driver independently serializes actual upstream edits into that protocol. No shared source was edited. These are comparison-validated candidates pending infrastructure integration.

See [the evidence report](../../claims/wave1-05-second-evidence/REPORT.md) for reproduction, the full source corpus, mutants, throughput and limits.

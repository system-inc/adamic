# typescript-no-this-alias

Matches declaration and compound-assignment aliases, asymmetric target/initializer unwrapping, TypeScript filename gates, allow lists and decoded ReportDestructuring options. All 40 upstream cases matched. Manifest options are already decoded Go NoThisAliasOptions, as required by directory registration; arbitrary configuration-decoder acceptance is not certified.

Own implementation and helpers now use `.ts` under Ahra's explicit fallback, with a directory descriptor, unchanged upstream Go adapter, raw witnesses and a semantic mutant.

Directory registration now discovers these `.ts` modules; the shared profile test still fails to compile against the registration foundation. The two non-null rules additionally use `suggestion-edits:<upstream message id>` with a byte-exact ordered edit list because the inherited driver accepts only one edit over the finding range. The scratch Go driver independently serializes actual upstream edits into that protocol. No shared source was edited. These are comparison-validated candidates pending infrastructure integration.

See [the evidence report](../../claims/wave1-05-second-evidence/REPORT.md) for reproduction, the full source corpus, mutants, throughput and limits.

## Ahra correction, October 7

Sources now use `.ts` under Ahra's explicit fallback; integration will rename
Adamic sources to `.a`. This changes owned filenames/imports and mutant targets,
not rule behavior. Historical evidence used the `.a` version before this correction.
No further rules were claimed and no shared source was edited.

The shared harness branch `codex/lint-harness-dot-a` was not published on origin
when fetched. Normal lint-package compilation remains blocked by
`profile_test.go:32:23: cannot range over portFiles` because registration made it
a function. The Go serializer also rejects suggestions outside the diagnostic
range or with multiple edits, which the two non-null rules require. The earlier
scratch comparison preserved every suggestion id, range and replacement, but
those protocol changes still need the infrastructure owner's integration.
Work stops at these shared blockers, as requested. No complete default-harness
certification is claimed after the filename fallback.

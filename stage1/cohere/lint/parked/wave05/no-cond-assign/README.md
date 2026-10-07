# no-cond-assign

Implements upstream except-parens/always behavior, compound assignment tokens and conditional ancestry. The owned assignment-token mutant is caught by all three comparison runtimes.

Candidate implementation now uses `.ts` under Ahra's fallback; directory discovery succeeds after the rename. Shared infrastructure was not edited. Four-way comparison uses the reproducible scratch overlay in [wave1-05 evidence](../../claims/wave1-05-evidence/REPORT.md). These candidates are not certified for the full requested corpus until the remaining harness and parser boundaries are resolved.

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

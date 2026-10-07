# Type-aware wave 27

Branch: codex/typeaware-wave-27.
Base: origin/codex/tsgo-c-library, 0d540f413625f016f20fea39761c7b184f335de6.

Claimed rules, positions 79 through 81 after excluding existing ports:

- base/correctness-require-orm-column-nullable-parity
- base/correctness-require-serializable-nullable-parity
- base/correctness-require-verify-array-parity

Ranking reconstructed from VOLUME_REPORT.md's linked compiler-all.counts and
repository-all.counts, descending combined count with lexical ties. All three
have zero compiler and repository findings. The 26 existing ports exclude 25
entries from this 197-rule checker ranking; method-signature-style is outside
that checker-only table. No matching port file or claim was found in stage1
on any fetched origin branch before this claim. No rules were skipped.

## Second batch

The initial three ports, their independent byte oracle, normally exiting mutants,
corpus/sanitizer checks and evidence were pushed in
5da2444b15f307146e3353fe840c66f54bb62340 before selecting this batch.

After fetching all origin heads without recursion, the first available entries
in the same 197-rule combined by-volume ranking are:

- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams

All have combined count zero. Checked ports on origin/codex/tsgo-c-library and
origin/main and Markdown claim records on all 325 fetched origin refs, including
unqualified names for these candidates. No matches were found for these three.
This claim update is pushed before implementing the second batch.

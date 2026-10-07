# sourcemap.ts: 2 -> 1

Read all 823 source lines before editing. One required result is asserted at its
original call. required-values.json contains the exact AST address and invariant.
The remaining TS2345 (line 216) is listed verbatim in after.json and declined in
the existing indexed ledger. Getter-counterexample.json records actual Node
behavior proving the guard does not establish the payload's presence: three
property reads, undefined payload, serialized content [null]. No getter reads
are cached, skipped, moved or defaulted.

Stock JavaScript bytes and adapter idempotence pass on all 30 owned files.
The new call-result ! -> ?? 0 mutant is caught by JavaScript and site-contract
checks. Default oracle: 106367 passing, zero failing/pending,
empty baseline diff, 210.677 seconds. The API projection is exactly
20's 189 plus 40's 28 approved lines; every other reference file is identical.
No native checker diagnostic is filtered. The pinned latent source-only meter
counts one remaining diagnostic, so this file is not counted as closed.

Reproduce using the same commands as visitorPublic's proof, substituting the
before snapshot /tmp/emit33-close-source-before-sourcemap, current source tree
/tmp/emit33-close-meter-source, oracle tree /tmp/emit33-close-tree, and census
command latent-file-census.sh /tmp/emit33-close-meter TREE OUTPUT. The mutant uses
MUTANT_FILE=sourcemap.ts and MUTANT_EXPRESSION=JSON.stringify(toJSON()).

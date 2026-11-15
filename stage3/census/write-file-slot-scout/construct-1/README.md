# Construction adaptation acceptance

Final acceptance index: [ACCEPTANCE.json](ACCEPTANCE.json).
All 79 native outcomes: [production-roots.md](production-roots.md).
Owner routing: [OWNERSHIP.md](OWNERSHIP.md).
Complete census pairing: [pairs.json](pairs.json).
Final check exits: [final-checks-exits.json](final-checks-exits.json),
[actual-stage-proofs.json](actual-stage-proofs.json),
[actual-small-proofs.json](actual-small-proofs.json), [built-proofs.json](built-proofs.json).
The two 93 legacy failures are measured on main itself and documented rather than
claimed as successful mutation tests. Earlier incorrectly parameterized witness
attempts in the raw checks index are superseded by the final index above.

The frozen implementation is 51104a74ba9e981fe93404e21e9c66298e47425f; its merge
with main is 0adcd82e25a88f32c50f625c21a0f127853537c1. The 21 KB frozen-proof.bundle
preserves both commits, requiring only the parent and main named in ACCEPTANCE.json.
Use `git bundle verify frozen-proof.bundle`, then `git fetch frozen-proof.bundle HEAD`
in a clone that contains those parents. The published feature completes proof
text/citations by a documentation-only amendment. Code, pin and ledger hashes
are listed in code-pin-ledger-identity.json; all 747 final fresh source hashes
match the once-tested lane in final-source-identity.json.

Census is full, single-file sys.ts, LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1,
CommonJSBundle, with complete relation trailers and zero diagnostics. Unique
Refused/NotYet findings are paired by kind, phase, reason and position-stripped
message. The raw Boundary disappearance is recorded separately.

Public control/candidate bundles are byte-identical; the stock upstream API diff
is inherited unchanged. Internal declarations lose the same @internal helper.
All actual source, single-file JavaScript and built-bundle diffs are under diffs/.
The real-body witness has 29 equal Node events and 10 equal harness events. Runtime
keys lose exactly the helper, writer name changes, and length/flags remain stable.

Whole lane was run once without filters or expectation changes: PASS, 106,366
passing, 0 pending, only the sanctioned API row. Its raw oracle exit is 1 and lane
exit is 0, as recorded. All logs and verdicts are preserved under lane/.
MANIFEST.json records full uncompressed hashes; large logs may use .gz.
Recipes retain the measured absolute scratch paths for unambiguous provenance.
No adaptation is submitted.

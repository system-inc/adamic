# Ledger note for task #b5w3ycg

The proposed unconditional builder chain check was measured before adaptation.
On main 6761c7374 with the complete stage 3 oracle, the identity-return path in
convertOrRepopulateDiagnosticMessageChain ran 6 times, with 0 chains lacking
messageText, category, or code. The wider no-repopulate-info path ran 7 times,
with 0 non-full chains. All runners, eight workers, no test filter: 106,366
passes and the one sanctioned Public APIs failure; complete pass/fail identities
match main. The check is therefore allowed by Adamic's condition 3. This is an
observed invariant over that suite, not a claim of exhaustive public input
validation. Keep the guard unconditional; Debug.assert is ungated even after
cli/debug-assertion-level at b1eb58de.

This note is kept within the new adaptation directory to honor the task's
single-directory commit boundary; stage3/upstream/LEDGER.md is not edited.

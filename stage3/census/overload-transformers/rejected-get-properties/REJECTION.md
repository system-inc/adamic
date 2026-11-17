# Rejected measurement candidate, not an installed adaptation

This directory preserves the exact getProperties draft measured for #ev14ch9. Its proposed installed path was stage3/adapt/99-correctness-overload-get-properties. It is deliberately outside stage3/adapt and has no landing branch or ledger line.

The full owner census removes the overload refusal but adds three replays of the unchecked cast refusal at transformers/utilities.ts:590:12, plus their Boundary records. That fails the mandatory 0-new gate. See ../REPORT.md and ../delta-latent-utilities.json. The ordinary adapter contracts, stock check, Adamic types, restoring mutant and public declaration comparison do not override that failure.

Reproduce the adapter controls from the repository root with NODE_PATH=stage3/api/node_modules node stage3/census/overload-transformers/adapter-controls.cjs. The recorded measurement commands predate relocation and name the proposed installed path; use this directory's identical adapt.cjs and mutant.cjs when replaying those commands.

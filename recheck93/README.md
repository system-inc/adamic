# Recheck: 93 single-step scope, all three trees

Stopped per the user's explicit instruction because chain-plus-main fails.

| Repository tree | 93 adapter exit | 93 check.cjs exit |
| --- | ---: | ---: |
| main fe5a1a54f045e9bce4e5aa32ca74885a317fbebd | 0 | 1 |
| chain c2f3c21f72 merged with that main (0945d4355) | 0 | 1 |
| candidate 63f46490 merged with that main (512f2a8ef, identical candidate tree; message-only amendment afterward) | 0 | 1 |

All fail with `TestBuiltinSourceScope: edit outside the three imports and removed declarations`.

## Invocation search and exact inputs

- stage3/apply.py:320-324 records refs/stage3/93-builtin-imports-before and -after immediately around the invocation of 93's adapt.cjs. Apply itself does not invoke check.cjs.
- stage3/adapt/93-builtin-imports/proofs/prepare.py:42 runs `node check.cjs CONTROL AFTER CHECKS_JSON`. Its BUILTIN_PREFIX_SNAPSHOTS=1 path accepts exact prefix snapshots from official apply.
- proofs/archive.py:24 reads the same 93 before/after refs. proofs/oracles.py:60 also invokes the exported source/transpile checks on its CONTROL and AFTER pair.
- Search of all *_test.go and test_*.py found no invocation naming check.cjs.

Used the actual full-apply input immediately before 93, tree 19c16edb2132cbe647f72aae6dee57fa6302176d, captured in the previous run. Verified every pipeline input before 93 (including apply.py, source.json and API locks/digest) is byte-identical across the three repository trees; see prefix-validation.json. For each tree independently, copied that common input to BEFORE and AFTER, ran that tree's unmodified 93 adapter once on AFTER, then ran that tree's unmodified check.cjs BEFORE AFTER. No later adapter ran on these snapshots. Candidate reproduction also matches the previously recorded immediate-before/immediate-after result.

## Cause

93's adapt.cjs itself adds the concrete setTimeout/clearTimeout forwarding arrows (the 995spev amendment, lines 103-110 in main/chain). Those are present already on main and chain. The candidate's 93 adapter additionally widens fsWatchFileWorker's polling interval. Its check.cjs undo() only restores the three local builtin declarations and removes the imports; it does not undo the timer forwarding changes. The failing main and chain cases therefore establish a pre-existing scope-oracle mismatch, without needing the candidate interval edit or any later pipeline transformations.

## Disposition

No adaptation branch pushed. Candidate remains 63f46490764862fef5af3ce6793557d4000994bc, parent c2f3c21f72b04e4c3acf7731913f0fc8ab89c2ca. Remaining original proof work was not resumed after this mandatory stop. This follow-up changes only evidence. No oracle, adaptation or pin was edited. Raw source snapshots and full assertion logs stay local; published follow-up contains commands, exit codes, log hashes and short failure excerpts only.

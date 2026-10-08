# Current validated parser slice

`current.tar.gz` contains the mirrored slice, copied-span manifest, entry driver
and kind names. Extract it into a new directory. `current.json` records the area
commit, TypeScript pin, counts and artifact hash. The existing 81-file dump
reference and 10,406-case manifest remain the acceptance targets.

Re-cut with the scanner-owned `stage3/slice/run.sh` from area b61e7064, after
`stage3/apply.sh` and the existing temporary-65 adapter. The seven entry symbols
are recorded in current.json and slice.json. `verify.cjs` passed before driver
staging: 2,080 exact source spans and 79 ordered import lists. Namespace wrappers
and imports are the slicer's only rewrites. Existing temporaries 60–65 belong to
the input adapted source; no scratch discovery stub is in this artifact.

All six temporaries remain. Removing 60–64 individually restores checker
errors; 65's original process-as-any probe still refuses and the area's source
still has that cast and its older local process declaration. No temporary 80.
The fetched area does not have adaptation 48; memoize retains its original body.
This artifact follows the exact current area tip rather than inventing it.

Full-tree and slice Node dumps are byte-identical to reference.json. All 10,406
case hashes equal cases-reference.json, including 1,996 parse diagnostic rows.
Native compilation stops at the predicate callback signature in core.ts:616:97.
The compiler scratch also has a detected function-value count miscompile and
is unsafe for native acceptance until its closure convention is reconciled.
Full evidence and reproduction logs are under ../evidence/front25/.
Use ../recut.sh to reproduce the source gathering.

Source declarations are from Microsoft TypeScript 6.0.3, copyright Microsoft
Corporation, Apache License 2.0. See LICENSE.txt copied from the pinned checkout.

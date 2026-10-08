# Current validated parser slice

`current.tar.gz` contains the mirrored slice, copied-span manifest, entry driver
and kind names. Extract it into a new directory. `current.json` records the area
commit, TypeScript pin, counts and artifact hash. The existing 81-file dump
reference and 10,406-case manifest remain the acceptance targets.

Re-cut with the scanner-owned stage3/slice/run.sh from main efe9f404 plus
temporary 65. Every adaptation in that checkout is applied by apply.sh.
The seven entry symbols are recorded in current.json and slice.json.
verify.cjs passes: 2,083 exact source spans and 79 ordered import lists.
Namespace wrappers and imports are the slicer's only rewrites. Temporaries
60–65 belong to the input adapted source; no discovery stub is in this artifact.
65 is now the grouped Node builtin type workaround; the old process-only
implementation is retired. No temporary 80 is reapplied; memoize 48 is present.

Full-tree and slice Node dumps are byte-identical to reference.json:
36,429,231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Node-end and JSDoc mutants are caught. The existing 10,406-case reference remains
the native acceptance target; this unit does not rerun that separate corpus.
Native compilation on main stops at debug.ts:113:19, Error.captureStackTrace
typing, before C in both split modes. The clean generic-function-value merge
has the same full-slice stop; its isolated conditional probe matches Node and
its one-byte output mutant is caught. No parser native acceptance is claimed.

Full evidence is under ../evidence/front31/. Use ../recut.sh to reproduce.

Source declarations are from Microsoft TypeScript 6.0.3, copyright Microsoft
Corporation, Apache License 2.0. See LICENSE.txt copied from the pinned checkout.

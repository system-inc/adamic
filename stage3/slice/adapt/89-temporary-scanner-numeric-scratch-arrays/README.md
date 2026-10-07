Temporary: comes out when sparse-array element inference and bounded numeric read proofs land

Plan published before implementation. In the scanner slice's reached
levenshteinWithMax only, change its two `new Array(s2.length + 1)` constructors
to `new Array<number>(s2.length + 1)`. Keep the length constructor and its real
holes. If strict indexed access requires it, add erased non-null markers to
reads of these arrays whose initialization is proved by the original loops.
Every write is numeric; previous is fully initialized before its first read,
and current writes column zero, the lower band, the active band and the upper
band before the arrays swap. The active band's current[j - 1] is already set.
No runtime statements, array lengths, or write order change.

This answers core.ts:113:9, `stage 0 can't lower an array of any yet`, after
library-array-holes 0140eed closes length-form construction. The erased typing
must emit byte-identical JavaScript, the slice's expanded Node dump must match
the full reference, and the same typing in a scratch full tree must pass the
stage 3 baseline with only the already sanctioned API differences. A changed
numeric write must fail the JavaScript identity check. No compiler edits land
on this branch. The pinned corpus and reference tree are retained as the same
inputs used for the established expanded dump.

Implemented: two constructors and six numeric reads. Emitted JavaScript is
byte-identical (1,872 bytes); changing a numeric write fails that identity
check. Numeric-storage native control prints `3`, matching Node; its one-byte
mutant is caught. Expanded slice Node output matches 1,369,432 tokens and
466 errors. Full baseline: 106,366 passing, zero pending, one API mismatch
exactly reconstructed by the pinned source tree's sanctioned owner proof;
all 60,930 other reference baselines remain identical. Current proof tools
expect newer full-tree adaptations, so this pinned reference tree uses its
matching 0b484f8f proof tools. No API reference was accepted or edited.

The next native blocker in both split modes is core.ts:128:43, a checked string
read coalescing to Debug.fail's never result. Evidence: scanner latest-native-retry.json.

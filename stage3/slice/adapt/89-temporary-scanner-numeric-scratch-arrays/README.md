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

# Ownership audit

Parent: `a77ca4a82271f18c142d8843ea76604b30d3833d`. The initial audit preceded edits;
`ownership-grep.txt` reproduces its whole-directory search on the immutable parent.
The original pin files are preserved under `ownership-parent/`.

| Directory | Pin or scope | Amendment |
|---|---|---|
| 93-builtin-imports | whole getNodeSystem, proofs/targets.json | construct the writer; regenerate after pin using adapt.cjs --pin |
| 99-wasm-crypto-import | whole getNodeSystem, proofs/targets.json, before/after text and hashes | compose construction with parent 1533; regenerate both pins using adapt.cjs --pin |
| 99-correctness-write-file-view | whole patchWriteFileEnsuringDirectory, owners.json | remove owner; retain the landed capture as moved-helper and restoring-mutant specification; regenerate using adapt.cjs --pin |
| 99-correctness-write-file-construct | previously unpinned initializer tail statement and harness import/module-local helper | new directory; generate owners.json using adapt.cjs --pin; carries ledger line |
| 99-correctness-timer-closures | whole TestServerHost encloses createServerHost | not touched: class bytes and call remain identical |

No actual owner pins the sys initializer tail statement or the entire harness file.
Historical evidence inventories are not live owner pins. The other sys owners
(exit/TTY and nested watching functions) retain their bytes. The construction and
helper deletion therefore amend their original owners rather than shadowing pins.

All generators report sourceWrites:0. No pin JSON was hand edited. Fresh full
application verifies the composition against the exact actual predecessor owners.

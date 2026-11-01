# 95h: accepted four-slot contract

Current shipping head: **1be494854**, branch `typescript/axolotl-checker-h`,
based on 95a. 95c remains unchanged at `7b1e14c56`.

The complete current report and canonical proof set are in
[accepted-four-slot/](accepted-four-slot/README.md). The accepted implementation
is [candidate/adapt.cjs](accepted-four-slot/candidate/adapt.cjs); it widens only
four @internal ModuleSpecifierResolutionHost slots and leaves 97b's four alone.

With accepted 97b 27162cab on both sides:

| Profile | Whole-program before / after | Removed / added |
| --- | ---: | ---: |
| tsc | 125 / 124 | 1 / 0 |
| services/server | 3341 / 3339 | 2 / 0 |
| all 711 sources | 8488 / 8486 | 2 / 0 |

utilities.ts own-file is **153 to 152**. Every reader's before/after attribution,
including zero rows, is in the current per-file CSV and census JSON files.
The oracle is **106366 passing, 1 standing Public APIs failure, 0 pending**,
352.424 seconds, with no timeout, exception or new public API byte.
The built and reference declaration hashes match main exactly.

The independent verifier covers 746 source files, checks exactly four declaration
edits, preserves 10681 CRLFs, and proves complete types.ts JavaScript identity
with comments retained (139659 bytes) and removed (121398 bytes). Both adapters
rerun with zero edits. The readFile undo restores exact baseline census sets and
is rejected by the verifier and adapter check.

Fresh full apply passes with H before 97b. All requested pairwise forward/reverse
compositions and the full B/C/D/E/H/97b stack have identical source maps. The full
stack's tsc census is 110 whole-program and 2 checker own-file diagnostics.

The remaining top-level proof files and README-blocked-two-slot.md preserve the
**superseded** two-slot stop report at Git commit `a6ce2a25d`. Do not use those
older top-level census files as the four-slot acceptance measurements. The
initial-three-slot-investigation directory is still older, before the 97b seam.
All current shipping-directory files are limited to adapt.cjs, verify.cjs,
sites.json and README.md; no evidence feeds that directory's apply input key.

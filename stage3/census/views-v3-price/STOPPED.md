# Declaration census checkpoint at the box reserve

Stopped at the recorded conservative box reserve, rather than starting another
ten-minute pair without enough allowance before 17:04 UTC. This box was already
77 minutes old when this continuation began. No measurement process remains
running. The paired closure count is still incomplete.

The frozen shared plan covers 81 files, 4,564 top-level declarations and 10,721
full-mode units. Both Go inventories agree with the independent TypeScript
inventory. All planned units occur once, in contiguous source intervals with
unchanged nesting and ownership. Lost, duplicate and moved-unit controls fail;
contiguous refinement preserves the complete inventory.

Two new declaration chunks completed sequential main/head runs, runtime unit
and timer audits, pairing, commits and ordinary pushes:

- d009: builder.ts, 55 units, 298.393 seconds.
- d036: executeCommandLine.ts, 69 units, 427.478 seconds.

Together with the four posted seed files, 161 units are paired. Each side has
160 attempted units, 94 finding-free attempted units and zero panicked units.
Retained IR totals for those observed units are main (0 ViewElement, 0 checks)
and head (48 ViewElement, 28 checks). Finding-free head units contain (15, 12).
These are observed-unit totals, not the missing full closure totals.

There were no interrupted new declaration chunks. The runner records and packs
interruptions, then substitutes disjoint contiguous child intervals without
rerunning the parent; a single indivisible declaration is explicitly blocked.
The earlier unmodified full-file baseline timeout remains in budget-stop.tar.gz.

Resume with a fresh explicit box budget and the same INDEX.json and compiler
pins. Completed seeds and chunks must not run again. The next pending piece
recorded by this run is d023; execution order may prioritize remaining work but
must not move a unit's frozen assignment. Full per-file and top-ten totals are
unavailable until every active leaf is complete.

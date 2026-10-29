# Void-container and generic-void findings, main 51981922 (for #p7t8tmh)

Extracted from the after run's `latent.jsonl` (main `51981922` + `55c1c38f`, the meter's first-error latent step over
tsc's 79 files; sha256 `4a942058...`, 27,453,371 bytes, [../README.md](../README.md)). That file was still on the box,
so nothing was rerun. The whole `latent.jsonl` is not committed.

- [void-container-rows.jsonl.gz](void-container-rows.jsonl.gz): every finding whose reason starts "a read of a void
  container": 9,610 rows, all Refused, in 56 files, at 9,504 distinct sites.
- [generic-void-rows.jsonl.gz](generic-void-rows.jsonl.gz): every finding whose reason contains "generic instantiated
  at void": 107 rows, all Refused, in 24 files, at 76 distinct sites.
- [summary.txt](summary.txt): counts by reason and by file, the 30 most frequent sites (with their source line), and for
  generic-void also the counts by the generic body's read site its reason names.

Each row has `file` (the record's file), `unit`, `where`, `kind`, `reason`, `text`, `phase`, `line`, `column`, `start`
and `end`; paths are relative to the adapted `src/compiler`.

**`start` and `end` are not in the findings.** `LatentFinding`'s `site_start` and `site_end` are empty on every one of
these findings, so `start` is computed: the character offset in the adapted source of `where`'s 1-based line and
column, in the file `where` names (4 generic-void rows have a `where` in another file than their record). Every row's
start was checked to begin the text of its line at its column. `end` is null throughout; no end is recorded.

Site counts are flat for void containers: no site carries more than 2 findings (a site with two roles, such as "as a
value" and "as an operand of &&"), so the 30 most frequent are ties at 2. For generic-void, `find`'s `T = JSDocTag` at
`core.ts:170:23` leads, with 21 findings at one site and 22 by its read site.

Made with [void_rows.py.txt](void_rows.py.txt): `python3 -I void_rows.py LATENT_JSONL ADAPTED/src/compiler/ OUTPUT_DIR`.

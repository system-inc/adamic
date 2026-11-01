# Evidence for 98-axolotl-b (3535ab4958e5)

Ran on a Linux cloud box: node v22.23.4, stock typescript 6.0.3 from
stage3/api, upstream v6.0.3 at 050880ce59e3, census built from main
f3fc89cc26bf (`go build ./stage3/census/tool`).

- apply-before.log, apply-after.log: `stage3/apply.sh` on main and with
  98-axolotl-b; apply-final.log reran it with the committed README (same src).
  patch-set-after.md carries the row `98-axolotl-b | 1 | 5 | 5`.
- census-*.log and census-*.jsonl.gz: the census over each tree's
  src/compiler. counts.txt and mutant-counts.txt: count.py, which calls
  stage3/meter/report.py's report() on them. es2015-diagnostics-*.txt: the
  es2015.ts findings in each.
- es2015-before-after.diff, src-trees-differ.txt: the only change.
- transpile.txt: transpile.cjs, sha256 of es2015.ts's JavaScript before and
  after, comments kept and removed: identical.
- idempotence.txt: a second adapt.cjs run reports 0 assertions; the hash is
  unchanged.
- probes.txt: partial, doubled, drift and moved-parent probes, each refused
  with nothing written.
- mutant.diff, mutant-counts.txt: the 4437 `!` dropped, es2015.ts 0 to 1
  (TS2345 at 4437:29), no other file moving.
- sites-*.txt (sites-probe.cjs): each site's occurrence, nearest function-like
  node and parent kind before and after. producer-probe.txt
  (producer-probe.cjs): stock 6.0.3 transpiling empty and all-omitted for-of
  patterns and spread calls to ES5 without throwing.

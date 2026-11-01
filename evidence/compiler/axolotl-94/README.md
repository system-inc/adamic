# Evidence for stage3/adapt/94-axolotl-property-read

Adaptation commit d42771c3 on compiler/axolotl-94, base main 4d63c0a6. Node
24.22.2, stock typescript 6.0.3 from stage3/api, upstream v6.0.3 at 050880ce.

- apply-main-control.log.gz: apply.sh on main (the control tree). The census
  binary built at main's tip is byte-identical to the one that measured it.
- apply.log.gz, apply-at-commit.log.gz: apply.sh with 94, before and at the
  commit; both exit 0 and give the same src/ tree.
- idempotence.log: a second adapter run, zero files. hostErrors.sha256 is the
  adapted file's hash, checked again after the mutant run.
- witnesses.{json,log}: 21 of 21 agree with a plain read.
- mutant.{json,log}: the null and undefined check removed; null and undefined
  go red, 19 of 21, exit 1.
- own-file-{before,after}.{txt,json}: the own-file table attributed as
  stage3/meter/report.py does, 145 in 21 of 79 files on both, identical sets.
- census-main.jsonl.gz, census-with-94.jsonl.gz: the raw census.
  census-outcomes.txt: the one lowering outcome that moves, hostErrors.ts
  accepted to Refused, "throwing a TypeError" at hostErrors.ts:14:15.
- stock-npm-ci.log, stock-build.log: stock npm ci and npm run build of the
  adapted tree, both exit 0.
- test_apply.log: python3 stage3/test_apply.py, 23 tests OK.

98-axolotl-c was not cut. 98c-require-probe.txt records why row 2 stopped:
assigning sys.ts's `require` to a number reports type `Require`, @types/node
25.3.3's global NodeJS.Require from internal/load/node_types, `(id: string): any`,
not adaptation 65's file-local declarations, which exist only in
performanceCore.ts and tracing.ts.

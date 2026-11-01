# Evidence for stage3/adapt/98-axolotl-c

Adaptation commit 3fb29032 on compiler/axolotl-98c, on 94's 8a7cd267. Node
24.22.2, stock typescript 6.0.3 from stage3/api, upstream v6.0.3 at 050880ce.

- apply-94.log.gz: apply.sh at 94's tip, the tree before 98. apply.log.gz: with 98-c, exit 0.
- idempotence.log, adapted.sha256: a second run writes zero files; sys.ts's hash holds.
- refuses-without-94.log: the adapter over main's tree without 94, exit 1, nothing written.
- own-file-before.* (with 94) and own-file-after.* (with 98-c): 145 to 142, sys.ts 4 to 1
  (1617:13 kept), every other file identical. census-with-94.jsonl.gz and
  census-with-98c.jsonl.gz are the raw census.
- js/: stock transpile diffs of sys.ts, comments kept and removed: only the hostErrors.js
  import and row 3's read change.
- witnesses.log: row 3, 14 of 14 agree. mutant.log: 94's null check removed, 12 of 14, exit 1.
- stock-npm-ci.log and stock-build.log: both exit 0, build-tests included.
- row2-require-type-probe.txt: sys.ts's require is type Require, @types/node's global.

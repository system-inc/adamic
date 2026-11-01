# Evidence for stage3/adapt/98-axolotl-c (work in progress)

Adaptation commit 0193ea39 on compiler/axolotl-98c-wip, on 94's 8a7cd267.
Node 24.22.2, stock typescript 6.0.3 from stage3/api, upstream v6.0.3.

- apply-94.log.gz: apply.sh at 94's tip, the tree before 98. apply.log.gz: with 98, exit 0.
- idempotence.log, adapted.sha256: a second run writes zero files; hashes hold.
- own-file-before.* (with 94) and own-file-after.* (with 98-c): 145 to 143; sys.ts 4 to 2,
  losing 1281:43, 1617:13, 1708:56, keeping 1598:69, gaining 1524:15 (createSHA256Hash, unmasked).
  census-with-94.jsonl.gz and census-with-98c.jsonl.gz are the raw census.
- js/: stock transpile diffs of sys.ts and types.ts, comments kept and removed. types.ts is
  empty; sys.ts differs only at the hostErrors.js import and row 3.
- witnesses.log: row 3, 14 of 14 agree. mutant.log: 94's null check removed, 12 of 14, exit 1.
- stock-npm-ci.log exits 0; stock-build.log fails in build-tests with 24 errors from row 4.
- row2-require-type-probe.txt: sys.ts's require is type Require (@types/node), not 65's.
  row2-probe.log: a scratch copy with row 2 applied as ruled; sys.ts then reads only 1524:15.

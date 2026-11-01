# Evidence for stage3/adapt/98-axolotl-e

Adaptation commit 93decc68 on compiler/axolotl-98e, on 98-c's 3fb29032. Node 24.22.2,
stock typescript 6.0.3 from stage3/api, upstream v6.0.3 at 050880ce.

- apply.log.gz: apply.sh with 98-e, exit 0. idempotence.log: a second run writes zero files.
- own-file-before.* (98-c) and own-file-after.* (98-e): 142 to 141, sys.ts 1 to 0, every
  other file identical. census-with-98e.jsonl.gz is the raw census.
- js/: stock transpile diffs of the five files, comments kept and removed.
- witnesses.log: 16 of 16. mutant-helper.log: 94's null check removed, 12 of 16, exit 1.
- stock-npm-ci.log, stock-build.log: exit 0, build-tests included.
- mutant-server-build.log: the server's widening undone, build-tests fails with 24 errors.
  restore.log, adapted.sha256, stock-build-restored.log: restored through the adapter, the
  hashes match and the build is green.
- mutant-row5.log: row 5 undone, sys.ts:1524:15 TS2375 back. mutant-row5-restored.log:
  restored, no sys.ts diagnostic.
- proofs/: the four built d.ts diffs, the normalizer's changes and api-lines.md.

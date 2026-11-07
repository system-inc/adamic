# takeWhile: stopped on a sparse-result counterexample

Task b5w3ycg, wave row 11. Base origin/main:
`f0271c6a7b2e92900eb7046d8246cfb89f52cfb6`.
Requested site: adapted core.ts:2595:16, census F42,
Refused "a cast the runtime can't check" at internal/lower/cast_proof.go:188.
Rule class: proposed JavaScript-changing established-shape adaptation.

No adapter is built or installed. No compiler, oracle, public declaration,
or ledger entry is changed. The task's required sparse Node witness disproves
unqualified result equivalence. This checkpoint preserves that finding for a
ruling before implementation. The lane count was not run; no claim is made
that tsc supplies this input or that a lane counterexample was found.

## Ownership and prior work

Read 99-correctness-axolotl-s/README.md first on the fetched origin/main.
It owns skipWhile, filterMutate, clear and createQueue and leaves takeWhile's
cast intact. Also read 99k, 99z and 99w's PROOF.md.
The pre-write search was:

```sh
git grep -n takeWhile origin/main -- 'stage3/adapt/*/adapt.cjs' 'stage3/adapt/*/*owners*.json' 'stage3/adapt/*/*sites*.json'
```

Only 30-indexed-reads/sites.json matches: lines 1110 and 1116 pin the indexed
read site, inserting a non-null assertion. Its adapter does not pin the whole
takeWhile owner. No prior whole-owner takeWhile adaptation was found.

## Reproduction and result

```sh
node stage3/adapt/99-correctness-core-takewhile/witness.cjs > /tmp/takewhile-witness.json 2> /tmp/takewhile-witness.log
```

The script executes the task's original runtime body and proposed runtime body,
with TypeScript annotations, cast and indexed-read assertion erased. It does
not claim to extract owners from full-applied trees. It compares results using
Node's isDeepStrictEqual and additionally checks `0 in result` independently.
Predicate calls are recorded, including undefined values with explicit tags.
Exit 0 means the script reproduced the expected counterexample, not that the
adaptation passed equivalence.

| Witness | Same result | Same predicate sequence |
| --- | --- | --- |
| undefined array | yes | yes |
| empty array | yes | yes |
| all accepted | yes | yes |
| none accepted | yes | yes |
| rejection in the middle | yes | yes |
| sparse hole rejected | yes | yes |
| sparse hole accepted | NO | yes |

For `new Array(1)` and `value => value === undefined`, both predicates receive
undefined exactly once. Both results have length 1. The original slice result
has no property 0; the proposed push result has own property 0 with value
undefined. Thus `0 in originalResult` is false and `0 in proposedResult` is true.
This is observable even without a getter, mutation, Proxy or array subclass.

99z documents a separate ruling allowing an off-count difference for elementAt;
this checkpoint does not assume that ruling authorizes takeWhile's sparse
result-shape difference. A ruling can explicitly allow this difference outside
the lane, or require a representation preserving holes. Until then, the
requested sparse equality proof cannot pass for the prescribed push loop.

## Proof status at stop

- Node witnesses: 7 classes, 1 result counterexample, 0 predicate-sequence differences.
- Whole-lane count and caller census: not run; no fabricated count table.
- Setup and dependency installs: not run; Node witness requires no dependencies.
- Both full applies: not run; no adapter built.
- Paired replay: not run; no adapter built.
- Adamic types over 79 roots: not run.
- Emitted JavaScript hunks: not run; only documentation and witness files changed.
- typescript.d.ts and tsserverlibrary.d.ts comparison: not run.
- Full core.ts site census before/after: not run.
- Original-owner mutant: not run; original remains installed.
- LEDGER.md: unchanged because no adaptation is admitted or built.
- All commands started for this checkpoint have exited; no background process remains.

Evidence is witness.cjs, witness.json and witness.log. The empty stderr log is
retained. All files are confined to this directory. There are no Go files.

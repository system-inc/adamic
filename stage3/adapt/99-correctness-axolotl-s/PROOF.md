# 99s initial proofs (historical)

The queue amendment has separate current evidence in [QUEUE-BOX-PROOF.md](QUEUE-BOX-PROOF.md).

Base: b729c7e07010fe6db34c6afc1d11c8ca63a86ae3. The new branch has one commit
on that base. Frozen main for the ordinary compiler census and fresh reference
oracle is 82b0c076395636800764ed1af7ada569681fc2ae. The compiler and parser
provenance is in compiler-provenance.json; no production compiler is patched.
Proofs ran before the final commit, so the accepted execution records the base
commit plus the working adaptation. Exact owner pins, source comparison and
post-oracle snapshot checks identify the measured source.

| Check | Result |
| --- | --- |
| Baseline full apply and fresh compiler build | Exit 0 |
| Finished full stage3/apply.sh --build-product | Exit 0; apply-result.json and apply-after.log.gz |
| No-op second adapter apply on full-applied tree | Zero files, edits and writes; source bytes preserved |
| Owner contract | 29 refused fixtures before any write; LF/CRLF and foreign edit preserved; replay preserves bytes and mtime |
| Adamic types, all 79 sorted compiler roots | Exit 0 before and after; no diagnostics |
| AST skipWhile caller audit | One call at transformers/ts.ts:1089; filter supplies its own narrowing; no caller check added |
| Source scope | All 746 files compared; only the four pinned core.ts owners differ |
| skipWhile JavaScript | Identical with comments retained and removed |
| Public declarations | typescript.d.ts and tsserverlibrary.d.ts byte-identical in both lib and freshly built output; skipWhile absent |
| Splice Node witnesses | Zero, positive truncation and own-length cases recorded; FIFO, reference release, undefined payload and sparse initialization preserved |
| Direct filterMutate mutant | Concrete number call: splice lowers; restored length is NotYet at 9:5, assigning a field of a value |
| Direct clear mutant | Same copied body with concrete number[] storage: splice lowers; restored length is NotYet at 2:5, assigning a field of a value |
| Dequeue draft | Cast-free and type-checks; sparse-input behavior differs; left unbuilt and open |
| generate_shards --check | Exit 0 |
| Lane Go fixtures, unchanged 99q command | Exit 0 in 128.544 seconds; no isolated retries needed |
| Frozen-main full lane | PASS; 106,366 passing, one sanctioned failure, zero pending |
| Finished-tree full lane | PASS; same counts and sanctioned public API failure |
| Title and outcome identities | Exact equality including duplicates: 106,367 occurrences, 106,285 distinct identities, 19,676 tasks |
| Ordinary per-file census | All 79 roots; 73 Refused, 4 NotYet, 2 accepted; no checker rejection |

The identity comparison retains runner, task file, full title path, outcome and
multiplicity. Both canonical multisets have SHA-256
`cfd153ef9477522fbe501e463bf926ff0f42864cf33f1d7a4ba55344c80c08b4`.
main-tasks.jsonl.gz and accepted-tasks.jsonl.gz preserve the full observed IPC
results. The canonical multisets are also retained separately. Counts alone
are not used as an identity proof.

The accepted lane reuses the actual full-applied tree, runs a zero-edit replay,
then the unchanged stage3/oracle/run.sh and unchanged stage3/lane/check.py guard.
finish-lane.py records this provenance explicitly in execution.json. It does not
restore or substitute oracle results. Both oracles use all runners, no filter,
eight workers and the ordinary 40-second upstream timeout. Install and build
exit 0, upstream tests exit 1 only for the sanctioned API baseline, and both
lane guards pass. The complete reports, baseline differences and compressed logs
are under oracle-main/ and oracle-accepted/.

The first lane Go attempt from the stopped work hit known 55-second Python
wrapper deadlines under load. initial-lane-go.log.gz retains it, and the
terminated retry is also retained. The final unchanged suite passes with the
full CPU quota. No test, selection or deadline was weakened. The command matches
99q, excluding separately executed shard/product executions:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
```

run-lane-go.py supports the authorized isolated retries if the known deadline
recurs; none were necessary in the final run. The pending Loom split is
chg_9as1ymj5. Initial stopped and partial evidence uses the initial- prefix;
PREVIOUS-STOP.md describes that earlier state, not the finished result.

Direct probes use this tree's ordinary adamic compiler, not the measurement
overlay. They copy the adapted bodies and differ only at the truncation
statement. clear's fixture specializes its parameter to number[] to isolate
concrete array storage; it does not claim the real unknown[] owner has lowered.
The whole-stack site census reports its actual erased-storage NotYet wall.
Current main's NotYet classification at class.go:471 is recorded beside the
Refused ruling in docs/rulings.md:285 and :439, owed enforcement by #53r37a3.
No diagnostic is relabeled. See direct-mutant-result.json and CENSUS.md.

Reproduction of the new adapter and local proofs:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source stage3/apply.sh --build-product /workspace/s-repro
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-s/adapt.cjs /workspace/s-repro
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-s/contract.cjs /workspace/s-before
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-s/witness.cjs
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-s/direct-mutant.cjs /workspace/scratch/s-bin/adamic-base
NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-axolotl-s/compare.cjs /workspace/s-before /workspace/s-repro
python3 stage3/lane/generate_shards.py --check
```

The ordinary census command is `/workspace/scratch/s-bin/tool
/workspace/s-probe/src/compiler /workspace/scratch/s2-census.jsonl`, with stdout
and stderr redirected to s2-census.log. CENSUS.md records full site-census
commands and coverage limitations. No partial-root fallback, backend change,
refusal suppression, extra caller check, or cast is added.

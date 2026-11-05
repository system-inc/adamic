# Count before building sites 4, 5 and 6

Base: b729c7e07010fe6db34c6afc1d11c8ca63a86ae3. TypeScript v6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8. Node v24.19.0, Linux x64.
No conditional caller edit was built until count-result.json passed.

| Site | Owning function | Calls | Undefined values |
| --- | --- | ---: | ---: |
| 14118, control | resolveTypeReferenceMembers, source.typeParameters | 439,874 | 111,773 |
| 16980 | getTypeArguments, final return value | 2,384,805 | 0 |
| 41574 | instantiateTypeWithSingleGenericCallSignature, concatenate result | 101 | 0 |
| 41628 | getUniqueTypeParameters, concatenate argument | 20 | 0 |

Each counter also received one planted undefined at startup. All four planted
events occur exactly once in counts.jsonl.gz; their summaries each show one call
and one undefined value. The same record function and undefined branch handle
planted and real calls. Planted calls use separate summary keys and are excluded
from the table. An exclusive-create marker prevents child processes from planting
again. The real absence control is well above zero and retains active test titles.
The conditional sites have no undefined test titles because no such event occurred.

The full base apply exited 0. count-instrument.py copied the complete output tree
with symlinks preserved and inserted exactly four calls into the copy's checker.ts.
Each call evaluates its original expression once and returns the original value.
The uncounted output tree was not instrumented. Both runs used this checkout's
unmodified stage3/oracle/run.sh, as stage3/lane/run.sh does: all runners, eight
workers, no filter, default timeout. The reference ran first, the counted copy
second. Both unchanged lane checkers passed, including the sanctioned API diff.

| Run | Passing | Failing | Pending | Tests seconds |
| --- | ---: | ---: | ---: | ---: |
| Uncounted base output | 106,366 | 1 | 0 | 358.954 |
| Counted scratch copy | 106,366 | 1 | 0 | 358.009 |

The sole failure in each is `unittests:: Public APIs for typescript.d.ts should be
acknowledged when they change`; baseline_diffs is exactly `["api/typescript.d.ts"]`.
Exact title/outcome multisets match with zero additions or removals. Raw observation
includes the worker event and host replay, giving 212,732 pass and two fail events
in each stream. Every multiplicity is even; normalized occurrences are 106,366
passing and one failing. Duplicate titles are retained as multiplicities.

The normalized identity multiset SHA-256 is
`61b98459713bf84d52f00222ecbafc36adfe60279ef9cbcd665c4778d1e83c63`.
Raw counter events and both identity streams are gzipped beside this file.
count-result.json retains all control test titles and per-site totals.

The preload wraps Mocha runnable execution for active titles and pass/fail/pending
events for identities. Its VM shim exposes the same recorder to evaluated compiler
code. Summaries flush per process on exit; undefined events append synchronously.
Recording errors are not swallowed. summarize-counts.py enforces expected totals,
all four planted controls, nonzero execution at every site, the real control,
exact identity equality and agreement of undefined events with summaries.

The initial optional compiler cache warmup competed for memory and was terminated
before any OOM kill; both full oracles completed unchanged. Setup and compiler
builds are separate from these test observations. No runner, worker count,
timeout, test expectation or baseline was altered.

## Exact oracle invocations

The scripts use /workspace/scratch/99r as their scratch root. Reproduction needs
a fresh scratch root with no counted copy, oracle outputs, event files or planted
marker. BEFORE is a full apply of b729c7e0 at that root's before/ directory.
The scripts can be copied there before the base checkout is selected.

```sh
python3 stage3/adapt/99-correctness-axolotl-r/count-instrument.py
NODE_OPTIONS=--require=/workspace/adamic/stage3/adapt/99-correctness-axolotl-r/count-preload.cjs CENSUS_IDENTITIES=/workspace/scratch/99r/reference-identities.jsonl stage3/oracle/run.sh /workspace/scratch/99r/before /workspace/scratch/99r/reference-oracle > /workspace/scratch/99r/reference-oracle.log 2>&1
NODE_OPTIONS=--require=/workspace/adamic/stage3/adapt/99-correctness-axolotl-r/count-preload.cjs CENSUS_IDENTITIES=/workspace/scratch/99r/counted-identities.jsonl SORTED_COUNTS=/workspace/scratch/99r/counts.jsonl stage3/oracle/run.sh /workspace/scratch/99r/counted /workspace/scratch/99r/counted-oracle > /workspace/scratch/99r/counted-oracle.log 2>&1
python3 stage3/adapt/99-correctness-axolotl-r/summarize-counts.py
```

Each oracle's expected exit is 1 for the sanctioned failure; invoke both even
though the first exits 1. The verifier rejects any other outcome multiset.
The actual recorded preload was a scratch copy of count-preload.cjs; only its
location differs from the commands above.

# 99r on 99s

Base: origin/typescript/core-runtime-checks-2 at
7d40082a09c23c9f0f7526819d0af85cbda870b1, the single 99s commit on b729c7e0.
Accepted 99r commit 04ee86cd974ebaf1a9af32232868868be38fbc9d was cherry-picked
cleanly onto it. Only the 99r directory is added by this commit.

| Check on the combined tree | Result |
| --- | --- |
| Full stage3/apply.sh from pinned TypeScript v6.0.3 | Exit 0 |
| Second apply of both 99r and 99s adapters | Zero files and edits each; all 746 source files preserve bytes and mtimes |
| adamic types, binary built from this chained checkout | All 79 recursive compiler roots; exit 0, zero diagnostics |
| Complete core.ts transpile, stock TypeScript 6.0.3 | Byte-identical before/after 99r, with comments kept and removed |
| generate_shards.py --check, configured tool PATH | Exit 0 |

The transpile comparison reverses only 99r's five exact core owner pins in
memory. Both sides retain 99s's exact after-text for skipWhile, filterMutate,
clear and createQueue, so its intentional JavaScript changes remain present.
The retained-comment output is 65,100 bytes, SHA-256
75d076a9396da0279e46de0c0a593f5d06c190ddba3572b37b8e23fc63540523.
The removed-comment output is 45,987 bytes, SHA-256
b81d42f936ead2d454b0ed0ca380bd800f6be6c0ce27a2bc176eddc70cd725f7.

Oracle and census were not rerun. Their earlier reports remain historical
measurements of the original 99r stack. Logs, replay snapshots and comparison
results for this chain are in /workspace/scratch/99r-chain/.

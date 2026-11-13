# Timer closures: watchPublic and tsbuildPublic census pairing

Main: `0aba99914be812ef7a0676c5ddd93024cefc81b5`.
Adaptation: `f37253b347c4f7e7bc223ed05fe04edb8cef0cd4`.
The entire adaptation tree is identical to `108e46471`: tree
`4c729608a4684ad14f0c11720123c24c83348fb2`.
The cached applied trees were verified against all 747 recorded source hashes
before measuring. Both sides use the same measurement-only binary built from
0aba99914; its exact SHA-256 is in provenance.json and its build metadata is in
binary-build-info.txt. Compiler and census source inputs are unchanged between
0aba99914 and f37253b3. No source adaptation was modified or pushed here.

## Counts

Counts include every reported Refused, NotYet and Boundary record, preserving
multiplicity across units.

| File | Side | Refused | NotYet | Boundary | Total | Units |
|---|---|---:|---:|---:|---:|---:|
| watchPublic.ts | main | 32 | 3 | 2 | 37 | 66 |
| watchPublic.ts | adapted | 32 | 3 | 2 | 37 | 66 |
| tsbuildPublic.ts | main | 61 | 6 | 8 | 75 | 113 |
| tsbuildPublic.ts | adapted | 61 | 6 | 8 | 75 | 113 |

## Every gone and new finding

| File | Retained | Gone: reason, where, unit | New: reason, where, unit |
|---|---:|---|---|
| watchPublic.ts | 37 | None | None |
| tsbuildPublic.ts | 75 | None | None |

There is no new finding to attribute to the IIFE, destructured setTimeout,
Function.call, new optional callback parameter, or an older site now reached.
All five classification categories therefore have zero entries.

This is not a claim that the new timer shapes lower successfully. In both trees,
the createWatchProgram implementation at watchPublic.ts:419:1 and
scheduleBuildInvalidatedProject at tsbuildPublic.ts:2059:1 have status
`deferred_generic_instantiation`. Their changed timer syntax is not established
as supported by this file-scoped census.

The pre-existing tsbuildPublic timer presence guards still produce these records
on both sides:

| Kind | Where | Reason | Unit |
|---|---|---|---|
| Refused | tsbuildPublic.ts:2061:10 | a method read as a value (setTimeout would lose its object, and this with it) | scheduleBuildInvalidatedProject, tsbuildPublic.ts:2059:1 |
| Refused | tsbuildPublic.ts:2061:39 | a method read as a value (clearTimeout would lose its object, and this with it) | scheduleBuildInvalidatedProject, tsbuildPublic.ts:2059:1 |

classification.json retains the exact generic-owner coverage records from both
runs. comparison.json contains every retained before/after pair as well as the
empty gone/new arrays, with full reasons, where positions and units.

## Commands and completeness

All four children exited 0. No run timed out, no checker diagnostics were recorded,
and both relation-overflow endpoints were empty for every run. The IR/no-output
guard was enabled. These are lowering measurements, not native executions.

The exact commands and environments are in the four status JSON files. Each is:

```
LATENT_FULL=1 LATENT_ONLY_FILE=<tree>/src/compiler/<file> \
LATENT_ASSERT_NO_OUTPUT=1 GOMAXPROCS=1 GOMEMLIMIT=3GiB \
  /workspace/scratch/timer-latent-tip <tree>/src/compiler <output.jsonl>
```

| Run | Child exit | Seconds |
|---|---:|---:|
| main watchPublic.ts | 0 | 21.230 |
| adapted watchPublic.ts | 0 | 21.537 |
| main tsbuildPublic.ts | 0 | 47.738 |
| adapted tsbuildPublic.ts | 0 | 47.437 |

measure.py runs the four jobs concurrently with a ten-minute limit per child;
the measurement piece completed in under a minute. compare.cjs exited 0 and uses
the same named unit/finding scope, kind, phase and reason pairing as the sys.ts
proof, preserving multiplicity. audit.py exited 0 and additionally verifies that
all 112 retained pairs' unit and finding positions map to exactly the same
unchanged source lines and columns. Thus unchanged totals are not hiding moved
findings with the same reason. Raw JSONL, logs, status records and the four source
snapshots are retained alongside the report.

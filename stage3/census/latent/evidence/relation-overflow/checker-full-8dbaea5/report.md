# checker.ts full census: task 8dbaea5

- Code pushed before the run: abb65ae8df966975f02402f7a7c4ca064394422b.
- Main adaptations: 6f68dd0fc5ee78c0b1230482d2c2089f02b1856b.
- TypeScript source: 050880ce59e30b356b686bd3144efe24f875ebc8 (6.0.3).
- Entry: /workspace/scratch/8dbaea5/adapted/src/tsc/tsc.ts.
- Selection: LATENT_FULL=1, LATENT_ONLY_FILE=/workspace/scratch/8dbaea5/adapted/src/compiler/checker.ts.
- Per-unit relation queries: disabled. Only load and end endpoints are queried.
- Run complete: False; supervised child exit: -15; wrapper exit: 241.
- Stop reason: stopped early to collect and push evidence before the environment's approximately two-hour turn limit; the 100-minute timeout had not fired.
- RelationOverflow after load: "".
- RelationOverflow at run end: null.
- Answer source: after-load query timing row; end unavailable.
- Completed timed units: 1718; last completed: /workspace/scratch/8dbaea5/adapted/src/compiler/checker.ts:35819:5.
- Serialized file records: 0; serialized units: 0.
- Wall time: 4717.147 seconds; peak RSS: 10900.8 MiB.

An empty string means that query detected no relation overflow. A null end
answer means the completion trailer is absent and the end answer is unavailable.
File observations are buffered until lowering returns. A partial header-only
JSONL is not a zero-findings result.
Timed units count completed lowering attempts, not skipped/deferred catalogue
entries. A last timing location is the last completed unit, not a claim about
the exact node executing when a cutoff occurred.

The runtime is the meter entry overlay on this branch's latent overlay. Its
resolved source reach keeps the tsc entry first, while the file selection lowers
only checker.ts. The adapted tree was generated with archived main adaptation
scripts. measure.sh is the actual wrapper. The run has a 6000-second GNU timeout
and a 12288-MiB resident bound. No native output is claimed.
Job 1 validation and both red reader mutants are in ../meter-readers/.

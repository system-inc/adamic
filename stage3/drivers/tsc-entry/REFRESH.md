Built a fresh adapted tsc entry tree after merging main efe9f404 into the existing branch.
Merge commit: 364cd19a; both native split builds use compiler source identical to main.
First stop: builder.ts:1246:69, TS2345; the minimal Node witness exits 0.
Latent status: new by exact message text in main's stage3/census/latent/REPORT.md.
Later stops are pending; no native binary or lowering observation is claimed.

The exact stopping diagnostic is:

```text
src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

Both `ADAMIC_NATIVE_SPLIT=0` and `1 ADAMIC_NATIVE_JOBS=5` exit 1 at checking,
with identical full diagnostic streams. The first minimal witness is
[01-indexed-path.a](probes/01-indexed-path.a); Node 24.19.0 prints `undefined`
and exits 0. Its native attempt gives the same TS2345 text and exits 1.

Every reached file outside `src/compiler`, called out separately:

- `src/tsc/tsc.ts`
- `src/tsc/_namespaces/ts.ts`

The closure is 81 files and matches stock TypeScript's independent program
loader. `executeCommandLine.ts` remains inside `src/compiler`.
The actual adapted tsc entry runs on Node and prints `Version 6.0.3`.

The requested main latent report records main b8fb957a, not current efe9f404.
Matching uses only its main-unmerged all-exact-reasons table, unescapes Markdown
pipes, and removes diagnostic/kind prefixes. This checker message is absent,
so it is marked new relative to that table; it is not a new lowering reason.
The report hash and original census provenance are recorded in the evidence.

Setup succeeded: Go ready 0.109s, Node ready 0.114s, markdown ready 0.255s,
submodules ready 0.270s, clang ready 0.657s, Go build ready 60.453s,
cache warm 60.587s, done 60.755s. `nproc=5`, cgroup quota 4 CPUs.
The environment file is `/workspace/adamic-tools/env.sh`.

Fresh evidence is under `evidence/main-efe9f404/`; the preceding unit's evidence
remains intact. All production compiler and adaptation sources match main.
The full repository gate and the conditional native tiny harness have not run.

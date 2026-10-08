Built the adapted tsc entry tree, its 81-file graph, and attempted both native split modes.
Compiler base c6761c24; first-stop evidence is committed on codex/stage3-tsc-entry-build.
Both builds exit 1 at builder.ts:1246:69, TS2345; Node entry prints Version 6.0.3.
Dropped-entry-edge mutant is caught by the independent TypeScript graph check.
Native execution and later stopping-body experiments are pending in this first evidence push.

## First stop

Both `ADAMIC_NATIVE_SPLIT=0` and `1 ADAMIC_NATIVE_JOBS=5` run main's unchanged
compiler on `/tmp/tsc-entry-adapted/src/tsc/tsc.ts`. Both exit 1; their complete
stderr streams are identical, 640 lines each. The first returned diagnostic is:

```text
src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

This file is inside `src/compiler`. It is a checker stop, before lowering,
not a NotYet or Refused finding. `probes/01-indexed-path.a` runs successfully
on Node, printing `undefined`, and main refuses its native build with the
same TS2345 message and elaboration. Giving the indexed value a defined
empty-string fallback removes this checker stop: `adamic types` exits 0.
The control native build instead reaches NotYet for an array of Path.
This control proves checker sensitivity, not successful native compilation.

## Module graph

The complete source closure is 81 files, independently confirmed with stock
TypeScript's program loader. Every reached file outside `src/compiler` is:

- `src/tsc/tsc.ts`
- `src/tsc/_namespaces/ts.ts`

`executeCommandLine.ts` is inside `src/compiler` in the pinned 6.0.3 tree.
The graph includes type-only edges. Host loads are external modules, recorded
in `evidence/closure.json`, not additional TypeScript source files.

## Setup and provenance

`bash cloud/setup.sh` succeeded with the requested GOPROXY. Timing lines:
Node ready 0.106s, Go ready 0.273s, clang ready 1.077s, markdown dependencies
ready 1.873s, submodules ready 212.269s, Go build ready 511.191s, cache warm
511.301s, done 511.381s. `nproc` is 5; cgroup quota is 4 CPUs.
The environment is `/workspace/adamic-tools/env.sh`, Node v24.19.0,
Go go1.27.1, clang 20.1.8. Complete setup output is in `evidence/setup.log`.
An early compiler build raced submodule setup and failed on the missing
`cohere/TypeScript/tsc/go.mod`; retry after setup's submodule step succeeded.

`bash stage3/apply.sh /tmp/tsc-entry-adapted` succeeded. The generated table
is at the tree root `patch-set.md`, not `stage3/patch-set.md`; a copy is retained.
It records 21 adaptations, 78 changed files, 5,128 lines added, 5,102 removed.
Source SHA256s and compiler/upstream commits are in `evidence/provenance.json`.
No compiler or adaptation source was edited.

## Node and mutants

The scanner-style stock-TypeScript loader runs the actual tsc entry point.
The initial ESM-only attempt failed because `sys` was undefined. Providing
upstream's CommonJS host globals makes `--version` print `Version 6.0.3`,
exit 0. Both attempts' streams are retained; no source adaptation was applied.

A scratch graph mutant omits traversal of the entry's import. Graph generation
still exits 0 with one file; `check-graph.cjs` exits 1 against TypeScript's
81-file program loader. The unmutated graph passes. This proves the graph
comparison can detect a missing dependency.

## Meter and limits

The newest main meter is `20261008T000754Z.hFDR7Q`, compiler f4efdd23.
The first TS2345 text does not occur in its src/compiler lowering reasons.
It is a checker finding, so the number of newly observed lowering messages
is zero so far. Ordinary lowering was not attempted and no latent overlay
was used for this entry build. Native C size, initialization, performance,
ownership and diagnostics agreement remain unmeasured.

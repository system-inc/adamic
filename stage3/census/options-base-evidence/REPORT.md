# OptionsBase scout for #6c4swvt piece 3

- Adamic main was `32d927c43c08d2df0cd7f891ffa82942c8fa85ca`.
- Library's record-to-fixed head was `bc3285c066a8914ffa34ddd047cafd0793d0c4bb`. The ancestry check exited 1, with merge base `7ed9da77f6125271b18de835f5391abf056dc274`.
- Both worktrees used cohere `5360753e19e4417675a90e92a1823bea66b13918` and its TypeScript checker `d92d9bfee114c80be2c375d72edae966176e3a4f`. Both recursive submodule updates exited 0.
- The meter's shallow `v6.0.3` TypeScript clone and pin check both exited 0. The pin was `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Readonly index probe

Stock `tsc --ignoreConfig --strict --noEmit readonly-index-probe.ts` exited 0. A prior invocation without `--ignoreConfig` exited 1 with TS5112 because its working directory contained a tsconfig. Stock `checker.getIndexInfosOfType(checker.getTypeAtLocation(parameter))` gives one string index with `isReadonly: true` for `readOption` and one writable index for `probe`'s source. Adamic's full latent probe exited 0 with zero checker diagnostics and reports these stops.

- At the call, `a Narrow seen as Readonly<Wide> (fixed objects and records have different storage; copy explicitly)`.
- At the index access, `a value of type Readonly<Wide>, a type with an index signature that isn't a string record`.

The census exit records stream completion. The two findings are lowering stops. Adamic's `recordElement` excludes a readonly string index from ordinary record storage on both pins.

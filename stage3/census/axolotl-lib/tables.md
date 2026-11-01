Measured on checker-rejected programs. Latent mode: first-error (`LATENT_FULL=0`).

| Measurement | Main | ES2020 | ES2020, capped Node references |
| --- | ---: | ---: | ---: |
| tsc entry diagnostics | 144 | 144 | 179 |
| Compiler aggregate diagnostics | 144 | 144 | 179 |
| Own-file diagnostic sum | 145 | 145 | 144 |
| External diagnostic union | 0 | 0 | 36 |
| Unlocated diagnostic union | 0 | 0 | 0 |
| Latent distinct NotYet | 652 | 652 | 652 |
| Latent distinct Refused | 2115 | 2115 | 2115 |
| Latent distinct SkippedDependency | 3 | 3 | 3 |
| Latent distinct error | 0 | 0 | 0 |
| Latent distinct panic | 0 | 0 | 0 |
| Units attempted | 4017 | 4017 | 4017 |
| Units deferred_generic_instantiation | 498 | 498 | 498 |
| Units skipped_checker_body | 28 | 28 | 28 |

Own-file sum follows the meter: union all census diagnostics by primary compiler-file location, then sum each file's unique diagnostics. It is not a sum of imported repeats. Latent sites use the meter's distinct `(kind, where, reason, text)` key.

Comparison: main to es2020.

Diagnostics, entry: 0 disappear; 0 appear.

Diagnostics, compiler_whole: 0 disappear; 0 appear.

Diagnostics, own_file: 0 disappear; 0 appear.

Diagnostics, external: 0 disappear; 0 appear.

Diagnostics, unlocated: 0 disappear; 0 appear.

Every changed latent reason row:

| Kind | Reason | Main | es2020 | Example removed | Example added |
| --- | --- | ---: | ---: | --- | --- |
| No changed rows | | | | | |

Comparison: main to es2020-capped-node.

Diagnostics, entry: 1 disappear; 36 appear.

```text
Disappears: src/compiler/sys.ts:1708:56: error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'string | ArrayBufferView<ArrayBufferLike>'.
  Type 'undefined' is not assignable to type 'string | ArrayBufferView<ArrayBufferLike>'.
```

```text
Appears: bundled:///node/node_modules/@types/node/child_process.d.ts:318:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/dgram.d.ts:558:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/events.d.ts:581:91: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:2155:41: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:2167:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:330:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:336:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:657:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:496:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:982:41: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:994:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/globals.typedarray.d.ts:15:15: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/globals.typedarray.d.ts:74:38: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/inspector.d.ts:49:66: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/net.d.ts:617:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/perf_hooks.d.ts:461:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:427:35: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:494:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:583:34: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:688:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/readline.d.ts:223:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/readline.d.ts:62:46: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/sqlite.d.ts:241:35: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/sqlite.d.ts:516:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/stream.d.ts:1004:21: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/stream.d.ts:672:21: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/test.d.ts:2187:21: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:144:25: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:34:53: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:64:25: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:86:51: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/util/types.d.ts:204:57: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:421:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:437:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:453:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/worker_threads.d.ts:394:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

Diagnostics, compiler_whole: 1 disappear; 36 appear.

```text
Disappears: src/compiler/sys.ts:1708:56: error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'string | ArrayBufferView<ArrayBufferLike>'.
  Type 'undefined' is not assignable to type 'string | ArrayBufferView<ArrayBufferLike>'.
```

```text
Appears: bundled:///node/node_modules/@types/node/child_process.d.ts:318:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/dgram.d.ts:558:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/events.d.ts:581:91: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:2155:41: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:2167:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:330:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:336:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:657:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:496:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:982:41: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:994:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/globals.typedarray.d.ts:15:15: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/globals.typedarray.d.ts:74:38: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/inspector.d.ts:49:66: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/net.d.ts:617:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/perf_hooks.d.ts:461:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:427:35: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:494:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:583:34: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:688:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/readline.d.ts:223:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/readline.d.ts:62:46: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/sqlite.d.ts:241:35: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/sqlite.d.ts:516:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/stream.d.ts:1004:21: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/stream.d.ts:672:21: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/test.d.ts:2187:21: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:144:25: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:34:53: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:64:25: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:86:51: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/util/types.d.ts:204:57: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:421:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:437:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:453:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/worker_threads.d.ts:394:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

Diagnostics, own_file: 1 disappear; 0 appear.

```text
Disappears: src/compiler/sys.ts:1708:56: error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'string | ArrayBufferView<ArrayBufferLike>'.
  Type 'undefined' is not assignable to type 'string | ArrayBufferView<ArrayBufferLike>'.
```

Diagnostics, external: 0 disappear; 36 appear.

```text
Appears: bundled:///node/node_modules/@types/node/child_process.d.ts:318:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/dgram.d.ts:558:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/events.d.ts:581:91: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:2155:41: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:2167:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:330:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:336:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs.d.ts:657:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:496:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:982:41: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/fs/promises.d.ts:994:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/globals.typedarray.d.ts:15:15: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/globals.typedarray.d.ts:74:38: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/inspector.d.ts:49:66: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/net.d.ts:617:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/perf_hooks.d.ts:461:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:427:35: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:494:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:583:34: error TS2304: Cannot find name 'AsyncDisposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/quic.d.ts:688:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/readline.d.ts:223:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/readline.d.ts:62:46: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/sqlite.d.ts:241:35: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/sqlite.d.ts:516:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/stream.d.ts:1004:21: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/stream.d.ts:672:21: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/test.d.ts:2187:21: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:144:25: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:34:53: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:64:25: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/timers.d.ts:86:51: error TS2304: Cannot find name 'Disposable'.
```

```text
Appears: bundled:///node/node_modules/@types/node/util/types.d.ts:204:57: error TS2304: Cannot find name 'Float16Array'.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:421:17: error TS2550: Property 'dispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:437:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/v8.d.ts:453:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

```text
Appears: bundled:///node/node_modules/@types/node/worker_threads.d.ts:394:17: error TS2550: Property 'asyncDispose' does not exist on type 'SymbolConstructor'. Do you need to change your target library? Try changing the 'lib' compiler option to 'esnext' or later.
```

Diagnostics, unlocated: 0 disappear; 0 appear.

Every changed latent reason row:

| Kind | Reason | Main | es2020-capped-node | Example removed | Example added |
| --- | --- | ---: | ---: | --- | --- |
| NotYet | a function returning Timeout | 0 | 1 |  | src/compiler/sys.ts:51:18 |
| NotYet | a function without a body | 2 | 0 | src/compiler/sys.ts:51:1 |  |
| NotYet | a value of type Timeout &#124; undefined | 0 | 1 |  | src/compiler/sys.ts:52:31 |

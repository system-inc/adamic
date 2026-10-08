Node observation counts, TypeScript 6.0.3. Refreshed by this unit.

| Fixture | AST nodes | Symbols | Bind diagnostics |
|---|---:|---:|---:|
| aliases.a | 23 | 5 | 0 |
| members.a | 45 | 12 | 0 |
| scopes.a | 40 | 8 | 0 |
| checker-use.a | 31 | 10 | 0 |
| checker-values.a | 24 | 10 | 0 |
| emitter-text.a | 22 | 5 | 0 |

These are Node binder observations. No fixture was registered in internal/oracle;
no native allocation/frees counts are available and its counts.md is unchanged.

Checker/emitter fixture observations, options fixed by fixtures/components.json:

| Input | Pre-emit diagnostics | Emitted outputs |
|---|---:|---|
| checker-use.a | 2 (2322, 2345) | checker-use.js |
| checker-values.a | 0 | checker-values.js |
| emitter-text.a | 0 | emitter-text.js, .js.map, .d.ts, .d.ts.map |

The checker message chain retains 2322 -> 2200 -> 2322. Both fixture projects
have zero global diagnostics and emitSkipped=false. Synthetic related-location
and Unicode-separator probes are in test-components.cjs, not additional files.

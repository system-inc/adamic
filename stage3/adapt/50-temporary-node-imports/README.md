Temporary: comes out when Adamic accepts require of a Node builtin with a literal specifier.

This unit declines the static rewrite. It does not claim to have removed the
runtime requires. Run `node adapt.cjs <tree>` through the stage 3 pipeline with
stock TypeScript 6.0.3 on NODE_PATH. The adapter parses all TypeScript under `src`,
checks the upstream lock's @types/node 25.3.3 pin, and prints each actual call,
its proposed `node:` specifier, and the reason it remains. Comments, strings
containing example programs, and existing static imports are not require calls.
It changes no source bytes and is idempotent.

All seven builtin calls in the pinned compiler are delayed or conditional:

| Source | Builtin | Why retained |
| --- | --- | --- |
| sys.ts:1470 | fs | getNodeSystem body; initialized only on the Node-like branch |
| sys.ts:1471 | path | same |
| sys.ts:1472 | os | same |
| sys.ts:1476 | crypto | same, plus catch for reduced Node installations without crypto |
| sys.ts:1650 | inspector | enableCPUProfiler body, after the active-session early return |
| tracing.ts:63 | fs | startTracing body, only while fs is undefined, with a load-error catch |
| performanceCore.ts:35 | perf_hooks | tryGetPerformance body, inside Node-like detection and a catch |

The seven calls include fs twice. `source-map-support` at sys.ts:1597
is an optional third-party package, not a builtin. It remains inside its catch.
The plugin require at sys.ts:1619 has a nonliteral specifier and remains dynamic.
An unconditional static import would move failures outside these guards and
make inspector/tracing load before the first request. Merely keeping an
assignment conditional after hoisting its import does not retain the load.
Namespace imports can also change the identity and mutability of CommonJS
module objects. No such change is justified by a passing diagnostic suite.

The host fixtures use static `node:` imports as their explicit input surface;
that does not authorize changing tsc's lazy initialization. A future rewrite
needs a policy that preserves runtime load timing, optional availability and
CommonJS module-object behavior. This unit makes no compiler or library changes.

Validation: audit.cjs checks the seven-site census, retained optional/dynamic
calls, unchanged source hashes and identical repeated reports. It runs real
source/specifier, source-comment and declaration-lock mutants. The full stage 3
oracle passes 106,367 assertions with zero baseline diffs on the unchanged tree.
A separate source diagnostic-value mutant builds successfully and then fails
one targeted upstream baseline assertion. oracle-report.json,
oracle-mutant-report.json and oracle-mutant-baseline.json preserve those observations.
See ../../fixtures/host/validation.md for commands, timings and all mutant catches.
The full uncached Adamic gate was not run.

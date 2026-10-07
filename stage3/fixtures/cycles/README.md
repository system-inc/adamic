# Module cycles

Ten small source-derived module programs mirror TypeScript 6.0.3's barrel shape:
`main.a` imports `_namespaces/ts.a`; that barrel re-exports modules which import
it back. Imports are narrowed and renamed to `.a`; interfaces are reduced to the
fields needed by the excerpts. Extracted function bodies and initializer
statements are unchanged. Drivers print results for compiler filenames and
extensions. All fixtures use erasable syntax and the current `oracle/node.mjs`;
no enum/namespace runner change was used.

The census at `origin/codex/tsc-census` (`429c1177`) reports a 76-file source
strongly connected component and 77 ExportDeclaration sites in three files.
`module_cycles.json`, `module_edges.json`, and `cycle-probes.jsonl.gz` were read.
The source census includes imports whose specifiers are used only as types.
Stock TypeScript emit erases those imports: the verified emitted graph has a
73-file runtime component. `types.ts`, `factory/nodeTests.ts`, and
`builderStatePublic.ts` leave the component. All 80 module bodies, including the
upstream-generated diagnostics and the tsc entry/barrel, still participate in
the entry's evaluation order. This measures stock source, not the in-flight
adapted tree or its policy for retaining type-import effects.

The October 6 ruling accepts cycles with ECMAScript semantics. A read of an
uninitialized binding at load time is the hazard. Function declarations are
initialized during module instantiation, before their bodies execute; fixture
10 must remain safe. Current main still refuses every cycle here, including
all nine successful Node cases.

| Fixture | Real source form | Node stdout / exit |
| --- | --- | --- |
| 01_call_time | `path.fileExtensionIs` calls `core.endsWith` through the barrel | four filename results; 0 |
| 02_reverse_barrel | same functions, dependency order reversed | same four results; 0 |
| 03_extensions | `fileExtensionIsOneOf` loops and calls across the cycle | four extension-list results; 0 |
| 04_directory_callback | `startsWithDirectory` calls the canonicalizer and `startsWith` | three directory-prefix results; 0 |
| 05_safe_load_read | `utilities.emptyFileSystemEntries` copies `core.emptyArray` | `0 0 true`; 0 |
| 06_import_order_mutant | identical to 05 except main's two imports are reversed | empty; 70 through oracle, 1 on raw Node |
| 07_namespace_barrel | star barrel plus a namespace barrel, as with `ts.performance` | `true false`; 0 |
| 08_named_reexports | named re-export list instead of star exports | same four filename results; 0 |
| 09_function_value_load | real `performance.nullTimer` stores `core.noop` twice | `true true`; 0 |
| 10_hoisted_before_body | same initializer runs before core's body; noop is already initialized | `true true`; 0 |

`status.json` uses relative paths to each directory's `main.a`, and stores the
exact Node oracle stdout, stderr and exit. Stage0 records the exact diagnostic
from `go run ./cmd/adamic build`, without Go's trailing `exit status 1` line.
Raw build output is retained in `logs/`. None compiled, so there was no native
output to compare and no exercised opportunity to detect a silent miscompile.
No production compiler files or shared fixtures test were edited.

## Load-time ledger

[LEDGER.md](LEDGER.md) lists 58 original top-level statements: 57 exercised on
Linux from `src/tsc/tsc.ts` with `--version`, plus one static direct conditional
read. There are 914 original imported-value read occurrences at 658 distinct
sites, including reads inside functions invoked by a top-level statement.
The separate generated diagnostic initializer contributes 2,130 occurrences.
Every observed provider module completed before its binding was read.
No adaptation candidate is established on this path.

`ledger.cjs` uses the stock TypeScript 6.0.3 compiler API to resolve imported
aliases and declaration locations, trace their actual reads, retain each
originating top-level statement, and verify Node's module-body order against
an independent ESM DFS. Const-enum references are left intact for stock constant
folding. The instrumented program and an uninstrumented emit have identical
module dependencies and print exactly `Version 6.0.3\n`. Upstream's Node host
expects CommonJS `require`, `__filename`, and `__dirname`; a bootstrap supplies
these while the source modules themselves retain ESM evaluation semantics.
`order.json` is the verified graph and order; `ledger.json` preserves binding,
declaration, read locations and chronological order evidence; `trace.json.gz`
preserves all occurrences, including generated input.

This is a bounded ledger, not a claim of exhaustive cross-platform analysis.
I did not certify unexecuted branches inside load-time-called functions,
Windows/macOS, development hooks or optional host packages. Cohere's
`nexus/correctness-no-import-cycle-load-time-read` rule was not run. The API
inventories direct top-level reads even in branches not taken; it does not
prove every possible transitive call in every host configuration. No source
adaptation is proposed on incomplete evidence.

## Mutants and validation

Fixture 06 reverses only these two imports from fixture 05:

```typescript
import { emptyFileSystemEntries } from './_namespaces/ts.a';
import { emptyArray } from './core.a';
```

In 05, barrel -> core -> utilities initializes `emptyArray` before utilities.
In 06, main -> core -> barrel -> utilities reaches the read while core is
still waiting for the barrel. Raw Node throws
`ReferenceError: Cannot access 'emptyArray' before initialization`, with empty
stdout and exit 1. Its complete stack is recorded in
`logs/06_import_order_mutant.raw-node.json`. The specified oracle runner
normalizes the error to
`adamic: panic: ReferenceError: Cannot access 'emptyArray' before initialization\n`,
exit 70. The stdout/stderr/exit comparison catches the import-order mutant.
Stage0 refuses both; this mutant does not prove stage0's eventual hazard rule.

A separate tooling mutant reverses the expected module body order. The
independent DFS versus observed Node order assertion catches it. The artifact
audit also removes a ledger row and moves a provider-end event after its read;
the raw-trace reconciliation and chronological evidence checks catch them.

Reproduce from Adamic's root after toolchain setup, with upstream v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8` and stock typescript@6.0.3 installed in
a scratch API directory:

```sh
source /workspace/adamic-tools/env.sh
node /tmp/cycles-typescript/scripts/processDiagnosticMessages.mjs /tmp/cycles-typescript/src/compiler/diagnosticMessages.json > /tmp/cycles-generate.log 2>&1
NODE_PATH=/tmp/cycles-api/node_modules node stage3/fixtures/cycles/ledger.cjs /tmp/cycles-typescript /tmp/cycles-trace-new > /tmp/cycles-ledger.log 2>&1
python3 stage3/fixtures/cycles/summarize.py /tmp/cycles-trace-new > /tmp/cycles-summary.log 2>&1
python3 stage3/fixtures/cycles/record.py > /tmp/cycles-record.log 2>&1
python3 stage3/fixtures/cycles/audit.py > /tmp/cycles-audit.log 2>&1
```

Setup: Go go1.27.1; clang 20.1.8; Node v24.19.0. Timing lines: Go ready 0s;
clang, Node and submodules ready 1s; build cache warm 113s; total 113s.
`nproc` = 5; `cpu.max` = `400000 100000` (four-CPU quota). Setup succeeded.
All fixture runs, tracing, artifact audit and the filtered existing oracle are
logged under `logs/`. The oracle command was
`go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/functions\.a$' -count=1 -timeout 30m -v`;
it selected both `functions.a` and `generic_functions.a`. Both passed in 10.217s. The full gate and TypeScript's own adapted-tree suite
were not run for this fixture-only unit.

The hardest real initializer, `parser.ts:441`/`factory/nodeFactory.ts:7407`,
creates the node factory and reaches memoization and parenthesizer helpers
while loading. Its imported reads are traced in the ledger. I did not reduce
that whole initializer into a fixture: its node/type/allocator dependencies
would turn this into a large factory-feature fixture. These small fixtures
cover its cycle scheduling requirement, not its complete factory behavior.
Enums, actual namespace declarations, native lifetime behavior, and adapted
TypeScript integration remain outside this bucket.

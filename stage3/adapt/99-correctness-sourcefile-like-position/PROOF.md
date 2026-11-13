# Proof

Measured before constructing this adapter, on main 751879544ea6e627e67c91d6ab3c4ba595be6356:

| Check | Observation |
| --- | --- |
| Ownership | No whole-interface pin; 71-writable-views edits lineMap only |
| JavaScript | 604 source files, 1,208 emits with comments kept/removed, byte-identical |
| Full scanner census | 138 -> 137 findings; only scanner.ts:469:12 removed; no new or changed findings |
| Full types.ts census | 0 -> 0 findings |
| SourceFile-to-SourceFileLike replay | utilities.ts:7959:37: no finding on either tree |
| Upstream declarations | npm ci and unmodified Hereby dts succeed separately on main and candidate |
| typescript.d.ts | Byte-identical, SHA-256 f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516 |
| tsserverlibrary.d.ts | Byte-identical, SHA-256 cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd |
| Lane API guard | Pass, 273 sanctioned declarations; sanctioned-api.json unchanged |
| Adamic types | Complete sorted 79-root compiler input, exit 0, zero diagnostics |

The initial flattened declaration probe was invalid on the baseline and was not used as proof. The replacement runs TypeScript's own project declaration build and scripts/dtsBundler.mjs, preserving per-project options.

The restoring mutant reinstates exactly the method signature in the before pin. Its required red observation is scanner.ts:469:12's unbound-method refusal. Owner drift/no-write controls, idempotence, the mutant observation, generated-shard checks and the complete unfiltered stage3/lane/run.sh acceptance are recorded on the [evidence branch](https://github.com/system-inc/adamic/tree/typescript/sourcefile-like-position-evidence/stage3/census/sourcefile-like-position). The adaptation is committed only after that full lane passes. No checker, oracle, API sanction, worker count or expected test count is weakened.

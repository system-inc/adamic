# Five timeout probes

All five previously timed-out inputs are slow successful C compiles on both recorded revisions. None was observed hanging or refusing. Each probe ran sequentially with a hard 45 second process-group cap. These investigate the original run; main is its recorded cf735d9f revision, not the latest origin/main.

| Path | Main cf735d9f seconds | Fixes ced32bf9 seconds | Result |
| --- | ---: | ---: | --- |
| stage1/cohere/css/gaps/printer_boundaries.ts | 15.901 | 15.940 | both accepted, exit 0 |
| stage1/cohere/estree/gaps/deepBinary.ts | 24.942 | 24.577 | both accepted, exit 0 |
| stage1/cohere/estree/gaps/emptyGenerics.ts | 25.086 | 24.478 | both accepted, exit 0 |
| stage1/cohere/estree/gaps/portParserRecovery.ts | 25.243 | 24.658 | both accepted, exit 0 |
| stage1/cohere/estree/gaps/typeMemberInitializer.ts | 24.971 | 24.617 | both accepted, exit 0 |

A separate SIGQUIT diagnostic at five seconds in fixes emptyGenerics shows native emitter work in dynamicProperties, shapeWith and walkExpressions. This observes C emission; it does not establish an infinite loop. No compiler source was changed.

The tool now defaults classification to --compile-timeout 45s, independent of --timeout 10s for backend builds and execution. JSON records both limits and every command wall_seconds. Timeout, crash and error classes remain failures. Runtime sampling still reserves five runtime command limits per selected program.

Validation: go test ./cmd/adamic-admission-delta -count=1 -timeout 60s -v passed in 0.327 s; touched TestHeadInputIsolation took 0.11 s. TestCallTargetReaders passed in 0.984 s. A real source mutant returning refused for a timeout failed TestCompilerTimeoutIsError in 0.10 s, exit 1; restored targeted tests passed.

Remaining: automatic diff corpus, compiler cache by revision SHA, parallel lowering-only classification, phase timings and main versus ced32bf9 full generated-corpus rerun.

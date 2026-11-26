# Mutants for stock-check.cjs

stock-check.cjs serves the loader's Node view as Load does: prelude_node.d.ts, then prelude.d.ts. It throws unless prelude.d.ts declares no console and prelude_node.d.ts exactly one. The mutant undoes this change in the checker alone, so the prelude stays split and the checker reads it as main did. It ran on Loom job J-rxs0btf at 496b73d1 (main 8fb2722c), after the head's own file passed the same test, then was put back and checked with `cmp`. The command is `go test -count=1 -run '^TestHostConfigurationScriptReaders$' -v ./internal/hostconfiguration`, which runs readers_test.cjs.

| Mutant | Edit | Red |
|---|---|---|
| The checker reads the prelude as main did | stock-check.cjs replaced whole by main 8fb2722c's (sha256 4e1361cb7d9e92f1e426132d074e83174ae58e3393669001cf59e9463913db98), which reads prelude.d.ts alone and rewrites its one console into Node's | TestHostConfigurationScriptReaders/node: "Stock prelude rewrite: expected exactly one console declaration, found 0", since the console now lives in prelude_console.d.ts and prelude_node.d.ts |

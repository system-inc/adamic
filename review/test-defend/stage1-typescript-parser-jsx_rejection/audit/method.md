# u156 audit method

Starting origin/main: 23a19b8b48662dfb6a19baa00b8efbb52ab5275e. nproc: 5. Warm Go 1.27.1 and clang tools; no setup installation. npm ci stage3/api reports 435ms.

Code under test: stage1/typescript/parser TypeScript port, its imported scanner, and native compilation of that port. Oracle: Go TypeScript parser in cohere/TypeScript/tsc; Node runs the same port and provides native diagnostic agreement. Never mutate Go oracle or oracle/node.mjs.

Requested names all exist. TestExpressionsAgree moved to expressions_products_test.go. The other requested wrappers remain in their named files. TestPerformance and TestWholePerformance remain separate because whole adds full-tree preflight assertions. TestExpressionsAgree and TestJsxScannerMutants have built-in mutation witnesses and clean agreement preconditions; production failures in those preconditions will not prove their witness checks. TestNodeCountCheckCatchesMutant is also a witness. TestCompilerExpressionsAgree_Setup constructs products without a semantic output assertion.

Inventory.json is a conservative static function inventory, not measured dynamic reachability. The clean package exceeded its 90 second test-binary budget during a built-in native rebuild. Requested rows are subsequently run alone. Package uniqueness outside this bounded slice is unknown.

Four production mutants were chosen from port guards, JSX metadata and recursive counting before observing any mutant results. Each independent native rebuild is bounded by 90 seconds. P1 is exclusively an empty-answer probe. No verdict rests on it. Harness weakening is reserved for witness and setup rows.

The full-history corpus clone exceeded 90 seconds and was terminated. A depth-one fetch of the pinned commit 050880ce59e30b356b686bd3144efe24f875ebc8 succeeded. ADAMIC_PARSER_BENCH=1 and ADAMIC_TYPESCRIPT_SOURCE=/tmp/u156-typescript enable opt-in rows. Scope/list and baseline logs are saved. No other package is tested.

The unconditional empty-entry return invalidated TypeScript flow narrowing in unreachable code and failed to build. P1-build-failed.log records that failure, with no kills credited. The valid P1 uses path.length >= 0, true for every string path, to return zero at entry while retaining a type-checkable body. This is a probe, not a production mutation.

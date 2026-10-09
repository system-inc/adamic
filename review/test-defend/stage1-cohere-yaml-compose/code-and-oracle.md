Starting commit: bfe0553300773c0b37db2c10df97adeb909705f8.

Code under test: the YAML TypeScript port, executed natively through Adamic, as source on Node, and as emitted JavaScript. D1 changes Composer.warning in composer.ts; D2 changes document-end construction in CSTParser.step; D3 changes the standalone file path in main.ts. Go cohere adapters and original libraries are oracles and were not changed.

Oracles: TestComposeMatchGo compares serialized composed trees, document directives, and diagnostics with Go cohere. TestCSTMatchesGo compares serialized CST tokens, absent fields, UTF-16 positions, and line starts with Go cohere. Both also compare original yaml@2.9.0 when enabled. The eight TestFileDriver shards compare exact standalone formatter output with Go cohere, through native, source Node, and emitted JavaScript executors. TestFormatterMatchesGo compares batch formatter output and separately checks original Prettier's known differences.

Semantic differences, rather than claimed exclusive TypeScript lines:

* Composition reports warning end offsets. Formatting and unist output do not report warnings. D1 changes that metadata while retaining the warning code and message.
* CST output distinguishes absent indentation from zero indentation on document-end tokens. Composed trees and formatter output do not serialize that presence flag. D2 changes the construction option, not the serialization checker.
* File mode passes the full file to format. Batch mode decodes case lines and uses a different call to format. D3 makes the file-mode input bound one UTF-16 unit short. Plain YAML may normalize the missing final newline, while keep-chomp and header boundary cases can change.

There are no separate Node/Native top-level twins among these rows: each target row already executes several backends.

Scope: all 72 current top-level names match the prior whole-package listing. D1 runs composition, unist, batch formatter, eight file shards, and FileDriverUnion. D2 additionally runs CST, props, and eight scalar agreement shards. D3 runs batch formatter, eight file shards, and FileDriverUnion. Each leaf/group has a separate 90-second test-binary budget and the requested 120-second outer backstop. Every mutant uses its own ADAMIC_BUILD_CACHE_DIR and ADAMIC_NATIVE_SPLIT=1.

This is a bounded functional agreement matrix. Production-mutant observations in suite-product setup rows or built-in mutant witnesses would not prove agreement-check strength; those rows are not counted as competing functional guards. Unrelated lexer, width, schema, gap, and cost-probe entries do not execute the changed behavior. No repository-wide uniqueness is claimed.

The requested per-row Go coverage profiles use -coverpkg=github.com/system-inc/adamic/internal/native. They measure compiler execution, not TypeScript lines. File-driver setup executes in a child and may hit cached products, so its parent-only profile understates compilation. No port-line exclusivity is inferred from these profiles. The defense rests on the observable semantic differences and the recorded mutation runs.

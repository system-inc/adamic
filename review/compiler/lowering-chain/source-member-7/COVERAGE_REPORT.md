Built declaration-only capture context and exact outside-use reasons for the 271-site probe toward roadmap step 09, task #akdvc40.
Delivery is compiler/checked-any, with main 031a1259b merged in 2ef0d19ef; the delivery SHA accompanies the push report.
Coverage is 3 checked, 88 refused inside their declarations, 180 not reached; focused tests, extraction verification, merge oracles and counts pass.
Four independent mutants are caught: wrong source attribution, wrong span attribution, changed stock text, and a substituted context body.
This push adds no contract support; recursive, dictionary and nullable-reference work remains Part 2, and MapLike<any> awaits compiler/records-maplike.

The previous measurement was 3 checked, 61 refused, 207 not reached. Loader stops fall from 100 to 42. The remaining 138 not-reached records carry specific lowering stops or outside-use requirements. Counts reflect current main and the improved context together; this is not a count of newly checked any uses. Successful type erasure or unchanged transport is never promoted to checked. All 80 ledger hashes match stock 050880ce59e30b356b686bd3144efe24f875ebc8. Inventory remains ea1b2359, with 271 records. The old table is retained as evidence/coverage-structural-before.json.

The extractor supplies exact stock import-type queries, private interface/type/enum contracts, lexical captures with their upstream checker types and mutable/const status, namespace type exports only when needed, and Node declaration contracts. Overload signatures include their unchanged implementation family. Source text and offsets are unchanged. The validator checks all 271 original spans and rejects executable context bodies. Imported contracts have declarations only, with no substituted implementation. No stock/cohere implementation was copied. Outer type parameters without concrete arguments and stock strict-checker failures remain exact stops rather than invented contracts or relaxed loader options.

Largest outside-stop groups, with one stock-site example each:

| Stop | Count | Example | Exact first requirement or diagnostic |
| --- | ---: | --- | --- |
| consumer_outside_declaration | 88 | `src/compiler/builder.ts:1919:25` | src/compiler/builder.ts:1919:25: emitSignature has no attributed guard in declaration emitSignature; its first outside reference is src/compiler/builder.ts:1932:61 (BinaryExpression); probing that consumer requires its enclosing declaration |
| loader_diagnostic | 42 | `src/compiler/checker.ts:6834:40` | probe/006/src/compiler/__any_probe.ts:6921:60: error TS2345: Argument of type 'TypeNode | undefined' is not assignable to parameter of type 'TypeNode'. |
| type_only_consumer_outside | 19 | `src/compiler/commandLineParser.ts:2430:9` | src/compiler/commandLineParser.ts:2430:9: type-only declaration JsonConversionNotifier has no executable use; its concrete consumers are outside the extracted declaration |
| consumer_not_in_stock_file | 14 | `src/compiler/commandLineParser.ts:3013:44` | src/compiler/commandLineParser.ts:3013:44: declaration parseJsonConfigFileContent lowered without a guard attributed to json; no outside reference is present in this stock file, so a caller or dataflow consumer must be supplied before this site can be classified checked |
| lowering_outside_declaration | 9 | `src/compiler/commandLineParser.ts:3835:83` | probe/089/src/compiler/__any_probe.ts:3850:65: Adamic 0.1 refuses a type predicate whose return is not proven (there is no body proving this parameter); inline the check where you use it, or return a discriminant comparison on the unmodified parameter (adamic/no-type-predicate) |
| generic_instantiation_outside | 8 | `src/compiler/core.ts:1482:37` | src/compiler/core.ts:1482:37: generic declaration clone is registered but has no concrete caller type arguments in the extracted program; its body requires an instantiation outside the declaration |

Every record in evidence/coverage.json retains its complete diagnostic or named requirement, span, hash, captured contracts, stock diagnostics and first outside reference. A refusal counts only when its source path and line fall inside the selected original declaration. Ambient predicates without a body remain unproven; the probe does not trust them. Callable contracts stay refused.

Validation commands (source /workspace/adamic-tools/env.sh and export GOPROXY='https://proxy.golang.org|direct'):

    node stage3/checked-any/extract-sites.cjs /tmp/checked-any-stock /tmp/checked-any-next-sites-v8 stage3/checked-any/evidence/coverage-entry-before.json
    node stage3/checked-any/validate-extraction.cjs /tmp/checked-any-stock /tmp/checked-any-next-sites-v8/manifest.json
    go run ./stage3/checked-any/probe /tmp/checked-any-next-sites-v8/manifest.json /tmp/checked-any-next-coverage-v8.json
    go test ./stage3/checked-any/probe -run '^(TestRefusalMustBeInsideDeclaration|TestUnobservedUseNamesItsOutsideRequirement)$' -count=1 -v
    go test ./internal/oracle -run '^(TestCheckedAny|TestCheckedJSON|TestCheckedAnyUnsupportedContracts)$' -count=1 -timeout 30m -v
    go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts

All output went directly to log files, retained in evidence/next-coverage/. Extraction verification passes all 271 declarations; probe tests pass in 0.007s; 16 scalar, 21 structural and 10 refusal witnesses pass in 15.377s with the existing sanitizer/leak oracle; counts pass in 63.568s. No whole package or full gate ran. The only removed old refusal expectation, staged_field, was superseded by main's checked nonnull assertion: undefined! now stops before the property use. Main's counts rows are retained; the 37 checked-any rows are restored. Compared with main, nested_destructured and switch_case_declarations/neighbors each retain/release one additional boxed string pair, as measured in the previous slice; allocations, frees and peak counts are unchanged. Part 1 adds no runtime fixtures or counts rows.

Mutants run independently. Dropping source identity makes capture.d.ts:12:4 falsely count inside entry.ts; expanding the span makes entry.ts:9:4 falsely count inside lines 10..20. Both fail TestRefusalMustBeInsideDeclaration with the intended assertion. Patches are retained under mutants/next-classification-*.patch. Changing emitSignature to EmitSignature in record 0 fails the extraction validator's unchanged-text check. Appending function substitutedBody(){return 0} after its exact stock bytes fails the declaration-only context check. Neither catcher is a compiler build failure. Production sources are restored and focused checks pass.

Toolchain setup first failed because merge conflict markers were still present; evidence/next-coverage/setup-first-failed.log.txt records the syntax errors. After exact conflict resolution, setup passes: node 0.020s, Go 0.025s, dependency step 0.069s, submodules 0.076s, clang 0.169s, build 26.693s, tests deferred 26.943s, cache warm 26.944s, done 26.969s. nproc=5, CPU quota=4; Node 24.19.0, Go 1.27.1, clang 20.1.8. Markdown validation is skipped by setup. No protected lower.go, native/emit.go, native/native.go or oracle/oracle_test.go edits were made.

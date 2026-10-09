# Code under test and oracle

Base: 76c59c81e8617cea1892a01841494895927a712c (origin/main when this defense began).

`TestWholeGeneratedAgrees`: the production TypeScript parser port, executed directly on Node and compiled to sanitized native code. The oracle is the unmodified typescript-go parser built by `goOracle`; the test compares complete canonical AST bytes. The mutation target is `Statements.statement`, including its WithKeyword branch at statements.ts:743-748. The generated fixture at whole_generated_test.go:23 includes `with (x) y();`, plus its Unicode-comment/CRLF-prefixed variant.

`TestEveryTypeNodeKindAgrees`: it feeds `wholeCases()` to the Go oracle to inventory type kinds, but feeds the port only two JSDoc inputs. It does not run the port's statement parser over wholeCases. Production statement-kind corruption can therefore distinguish the two rows, even though their host compilation coverage is identical.

`TestProduct_ParserOracle`: its code under test is the preparation recipe `wholeMutantBuildOracleProduct`, defined in whole_mutants_split_test.go:67. Its self-written success criterion is recipe completion without Fatal; it discards the returned path (setup_deadline_products_test.go:29). It runs Go build, but does not run the built parser or compare parser output, and does not assert that the returned artifact exists. The Go parser it builds is an oracle, not an allowed mutation target. The same recipe is called by TestProduct_ExpressionsOracle at expressions_products_test.go:141, also discarding the path.

The defender brief prohibits mutations of the oracle, harness, and tests. Thus no construction mutation from the earlier audit is allowed here. No prohibited mutant was planted to manufacture a result for TestProduct_ParserOracle. This row remains cannot-judge under the allowed mutation scope, with a concrete missing artifact-validation finding for its owner. The inability to defend is not permission to delete the row.

# Coverage

Go `-coverpkg` cannot instrument TypeScript. Requested Go profiles cover internal/buildcache for both product wrappers, including independent cold cache misses, and internal/lower plus internal/native for both semantic rows. There are zero exclusive covered blocks in either comparison. These are host-pipeline profiles, not parser-port line coverage.

Fresh Node V8 coverage supplies actual production-port execution evidence. The generated row calls Statements.statement 354 times and enters the WithKeyword branch twice; the type-kind row calls it zero times. The exact transformed-source UTF-16 offset of the emitted kind is 28164, within V8 range [27932,28274) with count 2. The TypeScript origin location is statements.ts:748. Filtered V8 records and coverage-difference.json preserve the measurements.

Both warm and cold Go profiles show no exclusive internal/buildcache blocks for the product wrappers. On cold misses both cover the same 57 blocks. Their identical recipe and discarded return value provide no distinct allowed production behavior to target.

# Predeclared aimed mutant

D1 changes the existing WithStatement kind constant to WhileStatement at statements.ts:748. It preserves traversal and child count. This is the menu's change-constant operation. It is aimed at the exclusive WithKeyword branch, and was declared in mutant-plan.json before running any mutant. Its standalone diff is against the base above. An explicit native compile probe uses the actual load/lower/native pipeline with Sanitize:true; both native and Node emit WhileStatement for `with (x) y();` under D1.

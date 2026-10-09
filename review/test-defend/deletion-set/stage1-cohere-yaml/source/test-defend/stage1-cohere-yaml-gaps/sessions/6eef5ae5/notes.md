# YAML gaps defense notes

Starting commit: 6eef5ae586af0354d3032d36535e54cefb2da43a. Audit base: 3bf0a5d9e74d197a38f61982e43b2d791f17b563. All 72 current top-level functions match audit discovery. The three requested rows remain in gaps_test.go and lexer_test.go.

## Code and oracle

TestLexerGaps calls Adamic Lower on gaps/multiplePush.ts. The code under test is arrayMethodArguments in internal/lower/object.go and its NotYet diagnostic. Node runs the original fixture and must print 2; self-written pins require NotYet and the push-arity diagnostic substring. It does not compare YAML lexer output or parse GAPS.md.

TestStructuralPositionRefusal calls Adamic Lower on gaps/structuralPosition.ts. The code under test is cycleFinder.slotsOf and the structural reachability/fresh-write proof that leads to its Refused return. Node must print 1; self-written pins require Refused and the Span[] array diagnostic substring. Span has start/end fields; Properties has those fields plus errors, so structural compatibility admits a back-reference even though the particular fixture does not construct a runtime cycle.

TestLexerMatchesGo checks the TypeScript Lexer port compiled by Adamic and run under Node, plus Adamic's emitted JavaScript. Live Go cohere and yaml@2.9.0 supply exact token-stream bytes. TestPropsMatchGo uses the same complete/chunked input corpus but compares resolved properties after CST parsing. These are different projections, not Node/native twin rows. Each row already combines several executors.

## Coverage

Each requested row and its subsumer was run independently with -count=1, anchored -run, -coverprofile and -coverpkg covering internal/lower and internal/native. The row-exclusive covered Go blocks number 15 for LexerGaps against StructuralPositionRefusal, 760 in the opposite direction, and two for LexerMatchesGo against PropsMatchGo. The push diagnostic block is exclusive in the first pair. Most structural exclusivity reflects LexerGaps stopping before the cycle pass.

Go coverage does not instrument TypeScript or C. Separate NODE_V8_COVERAGE runs provide actual source-port coverage for LexerMatchesGo and PropsMatchGo. Both reach 34 Lexer script functions; neither the range comparison nor the line-start approximation identifies a lexer-only covered range. Raw V8 reports and the extracted lexer scripts are retained. Shared coverage permits semantic defenses; the attempted lexer changes exercise empty raw tokens, explicit block indentation and CRLF token boundaries.

## Baseline and scope

npm ci ran in stage3/api. Tools were already usable; setup was skipped, nproc is 5. External npm oracles are pinned to yaml 2.9.0, prettier 3.9.6 and yaml-unist-parser 3.2.0.

The whole-package baseline exhausted the test binary's 90-second budget. Bounded baseline batches then exposed a missing yaml-unist-parser import in TestUnistMatchesGo. This was a dependency setup failure, not a production mutation. Installing the latest parser initially upgraded yaml to 2.9.1; the port's comments identify parser 3.2.0, so all three packages were then explicitly pinned. The affected baseline was rerun; every current row has passing enabled baseline evidence before any mutant was applied. The original oversized last batch also exhausted 90 seconds and was split into scalar/schema/width, Unist agreement, and Unist witness batches.

D1 and D2 replay every current top-level function across seven bounded batches, with a distinct ADAMIC_BUILD_CACHE_DIR for each mutant. D3-D5 initially replay five reached rows: LexerGaps, StructuralPositionRefusal, LexerMatchesGo, PropsMatchGo and CSTMatchesGo. Their outside rows are unknown unless separately replayed. A shared catch disproves exclusivity for that attempt without needing an absence claim outside this set. These lexer results are a limited mutation-set defense attempt, not a general redundancy proof.

All mutation files are standalone diffs against the starting commit. D1 changes only the push diagnostic constant. D2 drops the mutable-array Refused return together with the declaration used solely by that return. D3-D5 change lexer.ts itself, not the comparison driver, oracle, harness or test. Every source mutation is reversed after its run. Tests are untouched.

## Brief friction and owner findings

1. The enabled-library skip hints mention only yaml and prettier, while Unist also imports yaml-unist-parser. The missing dependency caused a real baseline failure and required pin discovery from port comments. All failed setup logs are retained; no defense runs took place on that red baseline.
2. A 90-second whole-package limit cannot cover this package's serial agreement and mutant witnesses. Batching preserves the limit and discovers current rows beyond the audit slice. TestMain additionally performs a file-driver setup subprocess before the parent test timer.
3. Go -coverpkg cannot measure the TypeScript port. V8 source coverage supplements it, with an explicit line-range approximation rather than a claim of exact statement coverage.
4. The twin instruction allows defense=twin although the JSON enum omits it. None of the requested rows is an executor twin, so no schema exception is needed here.
5. The raw lexer and property rows share lexical inputs and coverage but assert different answer representations. Comparing covered lines alone cannot settle their relative value. D3 and D4 are shared completed output disagreements. D5 exhausts the binary budget, first in CST and then in a narrowed raw-lexer replay; these are unknown row results, not zero kills. Two native children outlive their timed-out parents and require session-owned process-group cleanup.
6. TestLexerGaps pins a diagnostic, rather than reading GAPS.md. A diagnostic-only exclusive mutant defends that contract, not the semantics of arity rejection. TestStructuralPositionRefusal checks both refusal class and diagnostic; removal of the refusal directly tests its claimed safety guard.
7. TestLexerMatchesGo's name promises agreement with Go, and its assertions check exact output bytes against live Go on complete and chunked inputs. There is no missing performance threshold or other unasserted promise in this row. Failure to find an exclusive mutant in three attempts is not a deletion recommendation.

The third lexer defense remains cannot-judge if its narrowed CRLF replay also exhausts the budget. Two shared catches do not meet the requested three-completed-attempt criterion for not defended. The row stays pending rather than being marked redundant. Its name does not overpromise: exact live-Go token agreement is asserted.

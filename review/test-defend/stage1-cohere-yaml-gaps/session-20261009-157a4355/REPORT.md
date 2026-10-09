TestLexerMatchesGo not defended after3 targeted production-port attempts.
All3 fail both LexerMatchesGo and PropsMatchGo with completed byte disagreements.
Production restored; evidence includes standalone diffs, logs and compiler coverage.

Starting origin/main:157a43552015f41a79331949c2e82b6f8c7caaab. Old audit components retained as audit-* files. Current list has72 tests, unchanged against audit-branch source inventory; no added tests missed. Full package baseline cooked at90seconds during built-in ComposeMutants, with no assertion failure recorded. Narrowed clean Lexer and Props passed7.201s and10.823s with yaml2.9.0 enabled.

CODE UNDER TEST: production YAML TypeScript lexer.ts port, compiled through Adamic and also run on Node. ORACLE: unchanged live Go cohere token stream plus independent yaml2.9.0. Native sanitized, original Node and emitted JS output are compared byte for byte. No oracle, harness, test or driver changes.

Coverage:2 compiler Go blocks exclusive to Lexer compared with Props. See coverage-exclusive.json and coverage-diff.md. Neither gives a production TS line unique to Lexer. The targeted semantic difference was raw token output versus the reduced property outline; both rows use the same complete/chunked corpus.

D1 lexer.ts359: reset blockScalarKeep=false to true. Extra trailing blank line moves into scalar token, producing Lexer byte39040 disagreement; Props detects changed spaceBefore at byte4335. Both fail. Matrix16.802s.
D2 lexer.ts483: allowEmpty=true to false for block-scalar body. Empty raw token omitted, Lexer byte860641 disagreement; Props detects offset9 becoming0 at byte245694. Both fail. Matrix17.289s.
D3 lexer.ts635: quoted-key flowKey=true to false. Colon after quoted flow key becomes scalar content, Lexer byte4163499 disagreement; Props detects lost map-value indicator at byte1169863. Both fail. Matrix17.307s.

All observed failed rows for each mutant: TestLexerMatchesGo, TestPropsMatchGo. No observed survivors. Matrix limited to these2 rows; other package outcomes unknown. Because Props catches every attempt, no attempted mutant can be package-unique to Lexer irrespective of unsampled outcomes. This does not establish that every possible production bug is subsumed. We recommend no deletion based on this finite result.

Each switch-free standalone diff applies to starting origin/main. Separate sanitized native port validation built lex_main.ts with D1 in0.985s, D2 in1.014s, D3 in1.048s, all exit0. Matrix failures are runtime byte comparisons, not lowering, clang, sanitizer or timeout failures. Production restored finally; final clean bounded run passed20.44s. Tests unchanged.

Name/assertion finding: MatchesGo promises agreement, and asserts exact nonempty corpus token bytes against Go. No mismatch between name and assertions identified. It additionally checks source Node, emitted JS and yaml2.9.0. These are integration checks, not only isolated lexer logic. Its independent library step can skip if the environment is missing; all selected runs here enabled it and none skipped.

Brief friction and cost:
1. Whole package is over budget. TestMain setup and built-in native mutant builds consume much of the clock. Baseline timed out during ComposeMutants/implicit_key_limit_skipped at90seconds, not on an assertion. We narrowed to the target and its named subsumer rather than spending the budget rerunning unrelated native products. Other downstream consumers are documented in reached-imports.txt.
2. Requested Go coverpkg can cover lowering/emission, but cannot measure dynamic TypeScript port lines executed by a separately compiled native/Node program. We used compiler profiles only as leads and described the semantic token differences explicitly.
3. The audit supplied only1 shared mutant for this row, and it failed a lowering precondition. Our3 port mutations provide stronger evidence of real byte disagreements, still shared with Props. The audit also had a port mutation that hung; we avoided the known nontermination condition.
4. The2-row bounded matrix is enough to disprove uniqueness of these3 attempts, but not enough to prove package subsumption in general. No unobserved test is counted as passing.
5. A pre-existing defense branch contains other work. This run is kept in its own session directory and existing history will be merged, never overwritten or force-pushed.
6. Compiler tools and the YAML library were warm. npm ci stage3/api was rerun as required. Fresh cache per mutant prevents stale products. Exact npm/compiler-build timing was not separately measured; fresh native validation timings are additional warm validation, not cold matrix build timings.

Timing: setup skipped, nproc5, warm Go1.27.1 Node24.19.0. Clean coverage7.201s+10.823s; production matrices16.802s+17.289s+17.307s; native validations0.985s+1.014s+1.048s; restored matrix20.44s. About10minutes of session work. Scope72 tests unchanged. Not covered: completed whole package baseline, mutation outcomes outside the named pair, dynamic TS/C coverage, repo-wide uniqueness or additional mutants beyond the requested3.

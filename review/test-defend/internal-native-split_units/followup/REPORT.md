All three rows are defended within the six-row runtime caller matrix.
Each of D2, D3 and D4 fails exactly one target while the other five rows pass.
Full-package uniqueness remains unknown; production and tests are restored.

Starting origin/main: 619e7a4cf33741cc04bc78dc4c0c8ba0e31d73fb. The requested defense branch already existed from the preceding target. A fresh follow-up branch starts at origin/main; its evidence will be merged into the existing remote defense history without rewriting it. Prior evidence is preserved at the parent directory, new evidence here in followup/.

Code under test: C string runtime, especially adamic_string_repeat, adamic_string_append and the UTF-16 view builder in string_index.c. Oracles: executed Node operations compared across full output. WTF-8 serialization is handwritten support. ASan/UBSan additionally guard memory. The view-append row also has a self-written pointer-reuse expectation of 1. Neither tests, harness nor oracle was changed.

Semantic differences and aimed mutants:
- Building versus ViewAfterAppend: the building history invokes repeat to make long ASCII, Latin, emoji and lone-surrogate strings, then interleaves random reads, slices and appends. Its subsumer does not invoke runtime repeat. D2 subtracts one from the public repeat count passed to repeat_unchecked. Padding's internal repeat is unchanged. The building row fails at operation 99.
- Index versus CacheStates: the index row's mixed pattern includes U+10FFFF, as well as isolated surrogate halves and BMP values. CacheStates uses only é, U+1F600, A, Z and !. D4 changes the four-byte decoding lead mask 0x07 to 0x03 in the indexed view builder, losing the upper Unicode-plane bit. The mixed index subcase fails at output line 1; the other patterns pass. This is a constant change on shared code, not an input-specific inserted statement.
- ViewAfterAppend versus Building: ViewAfterAppend compares the native pointer before and after appending a low surrogate and explicitly expects reuse. Building compares values only. D3 changes the in-place ownership condition from references == 1 to == 2. In the observed histories this disables reuse and retains correct value outputs, but the view row fails its expected pointer output. This mutation could be unsafe for other shared-reference inputs; no global answer-preserving claim is made.

Go coverage profiles were produced separately for each target and subsumer with:
  timeout 120 go test -json -count=1 -timeout 90s -coverpkg ./internal/native -coverprofile <test>.cover ./internal/native/ -run '^<test>$' > <test>-coverage.log 2>&1
coverage-comparison.json records zero exclusive Go blocks for each requested pair. The Go build paths overlap. Go coverage cannot instrument these compiled C operations. Mutant selection therefore uses the distinct semantic histories and assertions above, not exclusive source-preparation lines.

Matrix:
- TestRuntimeStringViews
- TestStringBuildingMatchesNode
- TestStringIndexCacheStatesMatchNode
- TestStringIndexMatchesNode
- TestStringViewAfterAppendMatchesNode
- TestStringsMatchJavaScript

Current tests were inventoried with go test -list . ./internal/native/. Direct C operation occurrences select string_test.go, string_index_test.go, string_build_test.go and string_views_test.go. All six tests in those files were included. The three requested rows still exist, and none skipped. grep also finds a string_slice call in the RegExp corpus harness; it and compiler-generated callers outside the listed matrix are explicitly unknown. This is a bounded runtime comparison, not package-wide uniqueness.

Baseline: the full package timed out at 90.088 binary seconds without observed assertion failures. The six-row clean baseline passed at 6.566 seconds; the restored baseline passed at 6.630 seconds. No audit was performed on an assertion-red baseline. D2 binary 14.547 seconds; D3 binary 14.776 seconds. Per-mutant wall and exact commands are in D*-results.json; those durations include Go compilation and runtime builds, not separately isolated rebuild timing.

All standalone diffs apply to the starting origin/main. Each compiled under clang -std=c11 -Wall -Wextra -Werror -pedantic -fsyntax-only -Iinternal/native/runtime against its translation unit. string_repeat_impl.h is validated through string.c, its owning translation unit. Native sanitizer test products also built and ran. Each run sets its own ADAMIC_BUILD_CACHE_DIR; runtime archives use source/compiler/flag content keys in the user cache, so changed C sources rebuilt. Production files were restored between mutations and after the matrix.

Friction and limits:
- All audit evidence fields in the user prompt were empty. REPORT.md, scope.md, results.json and menu.json were fetched with the full refspec and retained here as audit-*.
- The literal requested branch creation command cannot create a branch that already exists. A fresh follow-up checkout and history-preserving merge were used instead; no branch or remote history was overwritten.
- /tmp totals 8.8 GB, making 15 GB free impossible. Prior /tmp/native-defense was deleted. /tmp remained at 5.2 GB free and /workspace at 14 GB; deleting scratch on the separate /tmp filesystem cannot raise workspace free space. No disk-full failure occurred.
- Whole package did not fit the 90-second budget. Results outside the six-row matrix, including RegExp corpus and generated-code callers, remain unknown. No package-wide uniqueness is claimed.
- Coverage is Go build preparation, not C runtime instrumentation. Shared code can be defended by semantic differences.
- Whole-file reads can be large; all requested test bodies and mutant production files were read, with audit summaries retained in full.

Name/assertion findings: all three names promise string-value agreement against Node, which their comparisons enforce. ViewAfterAppend additionally requires native pointer reuse. None asserts a timing or instruction budget. All three are retained on the bounded evidence; no deletion, rewrite or weakened assertion is proposed.

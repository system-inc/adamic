Neither target gained a unique kill after three aimed attempts each. Keep both pending the owner's decision; this is not deletion evidence.
Clean baseline passed in 21.039 seconds, nproc 5, warm toolchain. All six matrices completed without skips or unknown rows.
Uniqueness is bounded to the keeper-authorized 47 TestRegExp functions, grouped into eight rows. Full package and repo-wide catches remain unknown.

Code under test: internal/regexp parser, compiler and native declarations, and the C runtime's RegExp.test/exec, UTF-16 candidate scanning, captures, lastIndex and heap-frame ownership. Oracle: live Node 24 RegExp.exec with d flag. The harness compares indices, capture strings, named groups and lastIndex and checks sanitizer exits, in finite-budget and unlimited modes. Neither target is a Node/Native twin; both execute the native runtime, against Node, on different inputs.

Coverage: separate clean profiles used -coverpkg=./internal/native,./internal/regexp. Search has 40 exclusive blocks and lint has 67. Profiles measure Go production code, including parsing and native declaration generation. They do not instrument the subprocess C runtime. Runtime leads below come from the read production branches and input histories, not a claim of dynamic C coverage. Full profiles and exclusive-coverage.json are included.

D1: 200 repetitions can overflow the 8192-byte choice-point arena; drop only the heap overflow frame free. Binary seconds 21.028. Failed rows: TestRegExpBytecodeTest262, TestRegExpLintPatternsNode, TestRegExpSearchNode.

D2: Search explicitly begins three Unicode/global patterns at lastIndex=1 inside a surrogate pair; lint cases always start at zero. Binary seconds 26.186. Failed rows: TestRegExpBytecodeRandomNode family, TestRegExpSearchNode.

D3: Nullable Unicode string-set first-character filter; search explicitly uses [\q{|ab}] on nonmatching x. Binary seconds 20.247. Failed rows: TestRegExpBytecodeRandomNode family, TestRegExpSearchNode.

D4: Lint patterns contain escaped tab classes and tabbed example text; search has no escaped tab pattern. Binary seconds 22.557. Failed rows: none.

D5: Lint has long uppercase CANONICALIZE text under iu. Remove the guard that disables exact literal-prefix filtering under case folding. Binary seconds 19.142. Failed rows: TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpLintPatternsNode, TestRegExpSearchNode.

D6: Lint includes a negative lookbehind dollar pattern. Search uses only positive lookaround. Binary seconds 25.479. Failed rows: TestRegExpBytecodeRandomNode family, TestRegExpBytecodeTest262, TestRegExpLintPatternsNode.

D4 survivor witness: tab-probe parses the same literal \t before and after D4. Clean output is parsed tab Value=9 Raw="\\t"; mutant output is Value=10. The matrix still passes because sets.go characterValue recovers known two-byte escapes from Raw and returns 9 independently of this parsed Value. This is a changed parser result masked at the observed native execution contract, not a demonstrated missing native tab guard. Logs and probe source are included.

Validation: every standalone diff passed git apply --check against the restored starting checkout. D3-D5 passed go vet ./internal/regexp/. All six compiled and executed native products under the existing C11 warning and sanitizer flags: failures are runtime disagreements or LeakSanitizer reports, not clang build failures. Each matrix had a distinct ADAMIC_BUILD_CACHE_DIR. Production sources were restored byte-for-byte; no tests, oracle or harness were modified.

Name findings: TestRegExpSearchNode checks candidate searching inside RegExp.test/exec. It does not call adamic_regex_search or test String.prototype.search's lastIndex restoration. Its comment accurately promises generic VM and generated runner capture agreement. TestRegExpLintPatternsNode checks every benchmark pattern/input result against live Node, as its name/comment promise; it does not promise a performance threshold. No unsupported performance claim is inferred.

Friction and limits: the audit started at 83f3940e; this session starts at bfe0553300773c0b37db2c10df97adeb909705f8. TestRegExpNativeStepLimitBoundary is a new neighbor and ran in every matrix. The keeper's 300-second binary allowance conflicts with the generic outer timeout120; timeout330 was used so compilation did not prematurely cut off the authorized run. No run approached 300 seconds. The keeper's exact ^TestRegExp scope excludes TestRegex-prefixed neighbors and non-regexp consumers; catches there are unknown. Go coverage cannot prove C line exclusivity. D5's first textual anchor appeared twice and was rejected before mutation; a longer unique anchor resolved it. A large audit/benchmark read was truncated, requiring focused reads. Existing defender branch evidence was found; this session uses its own directory and merges that branch without rewriting it.

Setup.sh was skipped; npm ci was rerun. Setup and native clang compile time were not separately instrumented. Binary elapsed times for baseline, coverage and every matrix are in logs and matrix.json. The six mutant binaries total 134.639 seconds. Go vet and command wall times for D3-D6 are in their run JSON files. No whole-package test, test deletion, rewrite, weakened comparison, main push or PR was performed.

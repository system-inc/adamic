Built: tailwind.ClassLiteralReader.ClassLiteralsIn, tailwind.SplitClasses and react.IsNamespacedMember in three separate .a files.
Commits: claim a90f5e9 pushed before implementation; implementation 9221c9b; previous three helpers already pushed at 8fe9c44.
Commands and outputs: setup 23s, nproc 5; isolated new batch PASS 41.024s; full helper package PASS 155.304s; filtered uncached oracle PASS 1.572s; go vet ./... PASS.
Mutants: wrong namespace, retained partial fields, missing nonbreaking-space whitespace and copied literal slice all compile, execute and are caught by real Go comparisons.
Not covered: full repository gate, integrated rule findings/fixes, live Tailwind design-system/corpus gates, arbitrary invalid UTF-8 or malformed AST graphs, and prerequisite implementations owned by other slots.

# Selection and ownership

The original three slot 02 helpers were ported, tested and pushed before this continuation. Fetched all six origin codex/lint-helpers* branches and read every claim. All symbols with more than ten remaining consumers were reserved, with the original comments bundle reserved outside claims/. These three were the unclaimed symbols tied at the highest count, ten. Claim update a90f5e9 was pushed before any implementation. A subsequent fetch and complete claim inspection found these three only in slot 02's claim. The snapshot is evidence/claims.log. No further helper is claimed.

Changes stay under this slot's helper files and standalone tests. No tracked cohere file, shared registration generator, shared rule harness or protected compiler file is changed. Temporary Go overlays add oracle exporters and capture calls without editing the cohere worktree. New Adamic files, including temporary mutant copies of inherited options_json, are .a. No PR opened.

# Observed behavior

IsNamespacedMember: the actual Go helper is called on every node, plus nil, with three predicates: accept all, prefix use, and reject all. Compare both the boolean result and the predicate's invocation count and argument. A property access requires Identifier React after skipping receiver parentheses, then an Identifier member. Other namespaces, computed access, asserted receivers, private members, wrong node kinds and nil nodes decline before invoking the predicate. The outer node is not unwrapped. Controls include nested parentheses, optional chaining, decoded escaped identifiers, Unicode identifiers and factory-built missing/empty member names. A malformed property access with nil receiver panics inside Go SkipParentheses; the Adamic port explicitly panics rather than silently declining. Such malformed factory projections are excluded from the normal comparison, and exact Go internal panic prose is not claimed.

SplitClasses: actual Go SplitClasses is called on 232 distinct texts, including every unique string/no-substitution-template literal in the parsed consumer corpus and controls for holes, duplicated delimiters, empty text, NUL, astral Unicode and whitespace. All 1,114,112 integers from zero through 0x10ffff are tested as separators between a and b; surrogate integers are converted to replacement rune exactly as Go string(rune) does. A compact result records every accepted separator and the accepted count, so equal totals cannot hide a wrong separator. All 25 Go whitespace code points match. U+0085 and U+00A0 split fields; U+FEFF and U+200B do not. Fields containing ${ are omitted in their entirety, with other order, duplicates and text preserved.

ClassLiteralsIn: an oracle-only exporter calls the actual public method and its private classValuesIn prerequisite on a bound default-settings reader for every parser node plus nil. It compares backing slice pointers, serializes actual literal text, byte ranges, origin and leading/trailing edges, and projects template presence separately. The Adamic function receives this values projection through an explicit callback, invokes it once, and returns the exact read-only literals array. Literal objects and node identity are preserved by returning the supplied slice; traversal, memoization and custom settings are external prerequisites. A nil Go slice is represented as an empty array, as documented in README.md. This comparison concerns the accessor, not the prerequisite traversal port.

The combined captured corpus has 813 deduplicated rule/file/source inputs across all 20 consumers, 559 distinct parsed sources including controls, 10,303 projected nodes, 10,304 literal-value projections and 30,912 namespace verdict/callback observations. Source extensions preserve Go TS/TSX/JS/JSX parser mode. File paths are canonicalized to absolute /fixture plus extension because these helpers do not inspect paths. Every consumer's actual runtime source is present; tests refuse missing coverage against the frozen readiness ledger.

Go's expected-output field is removed before the runtime corpus reaches Adamic. Baseline and mutants run on Node source, emitted JavaScript and ASan/UBSan native. Their outputs must match one another byte-for-byte before comparison with actual Go cohere. Every successful run must exit zero without stderr or sanitizer reports. The complete helper package passed in 155.304s. The new isolated package test passed in 41.024s; final full-package new-batch subtest passed in 35.14s.

# Capture and external failures

Capture workflow adapted from slot 03's overlay capture. Pinned cohere: 715ba94f3608a6500086b1076ce5cb7e51b836db. Actual React and structure rule packages passed in 3.053s and 0.281s on the final capture. Tailwind capture returned exit 1 in 8.748s: unavailable external engines and empty external corpora caused these eight live gate failures:

- TestClassOrderFixturesActuallyRan
- TestUnknownClassFixturesActuallyRan
- TestConflictFixturesActuallyRan
- TestConflictingClassesPlacementIsAccountedFor
- TestCanonicalFixturesActuallyRan
- TestCanonicalClassesPlacementIsAccountedFor
- TestUnknownClassesPlacesEveryCorpusClass
- TestClassOrderLiveMatchesTheEngineOverTheCorpus

This is not a passing Tailwind rule gate. The overlays record inputs before live-engine checks/skips; all ten Tailwind consumers are captured and can be compared against the real Go helpers without external engines. The script permits only this bounded known-failure set and requires every selected rule to be observed. It does not hide unexpected failures. Source and coverage artifacts regenerated byte for byte on a second independent capture. SHA256 of sources.jsonl.gz: 26a9af3d9a3a231b32bbbdbb0d7a20c955de66abcc8a410bae8aebf57ee1d4ee. SHA256 of coverage.json: 00962281973908e43dc654dc0341355debdc1335d2c6e763f0641cf76a6a85d4.

# Compiled semantic mutants

All four compile and execute successfully on all three backends. Compilation failures and sanitizer/runtime failures are not credited as semantic witnesses.

| Helper | Mutation | First independent Go mismatch |
|---|---|---|
| IsNamespacedMember | omit receiver text equals React guard | line 433: true:1:useState versus false:0: |
| SplitClasses | retain fields containing ${ | line 30915: split:2 versus split:1 |
| SplitClasses | omit U+00A0 from whitespace | line 30927: split:1 versus split:2 |
| ClassLiteralsIn | copy literals with slice() | line 31591: literals:1:false:0 versus literals:1:true:0 |

The whole helper package also reruns all seven previous semantic mutants: JSON raw control acceptance (15157), schema overlapping oneOf acceptance (15166), strict unknown-field acceptance (15178), policy missing interpolation (22166), JSX namespaced-name acceptance (1140), property computed-identifier acceptance (8290), and reader missing cache namespace (2). Each compiles and is caught by the existing Go comparator; prior reports describe their full contracts. No production file is mutated; variants are created only in temporary directories.

# Commands and limits

All tests write directly to logs, never through a pipe. Run from /workspace/adamic, source /workspace/adamic-tools/env.sh for each toolchain command.

```
bash cloud/setup.sh > stage1/cohere/lint/helpers/slot02/batch2/evidence/setup.log 2>&1
nproc > stage1/cohere/lint/helpers/slot02/batch2/evidence/nproc.log
python3 stage1/cohere/lint/helpers/slot02/batch2/testdata/regenerate.py > stage1/cohere/lint/helpers/slot02/batch2/evidence/regenerate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch2$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot02/batch2/evidence/third.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot02/batch2/evidence/final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot02/batch2/evidence/oracle.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot02/batch2/evidence/vet.log 2>&1
git diff --check > stage1/cohere/lint/helpers/slot02/batch2/evidence/format.log 2>&1
```

Setup: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 23s, total 23s on nproc 5. Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup succeeded. The filtered uncached oracle passed all six fixtures in 1.572s, with zero result cache hits and six probe misses. Vet and whitespace checks exit zero with empty logs. first.log records the failed relative parser filename attempt; second.log records the failed malformed nil receiver control. Both were investigated and corrected in the fixture contract, not credited as successes or mutants.

RULES.md lists all consumers and readiness.json retains each rule's remaining dependencies after these three helpers. Observation: 30 helper dependency entries are removed across 20 distinct rules. Inference from the frozen ledger: zero rules lose their final listed blocker from this batch alone. The shared ledger and rule statuses are not rewritten. Helpers remain conditional on faithful production adapters and the other slots' prerequisite ports.

The full repository test gate, all possible programs/configurations, integrated rule diagnostics/fixes/suggestions, custom settings traversal and live Tailwind designs remain outside this slice. The complete touched helper package and filtered external oracle are the bounded worker gate. Malformed ownership graphs and invalid UTF-8 byte strings are not claimed equivalent; missing/cyclic arena indices refuse explicitly.

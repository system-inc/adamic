TestSharedSliceAppendMatchesNode is defended by D02, a production C runtime mutant.
All 72 current top-level tests were replayed in bounded groups; only the target failed and 71 others passed.
Starting main: 2102ee6b93e0281dc8593abe96c8311aa66ea594; standalone diffs, clang checks, coverage profiles, matrices and logs are preserved here.

CODE UNDER TEST: Adamic lowering/native emission and the C runtime's adamic_string_share, adamic_string_append, and string slice operations. The mutated implementation is string_share.c; no test input, oracle, harness, Go cohere implementation, or port source was changed. ORACLE: live Node stdout pinned to a\nx\n for the target, plus successful native execution under ASan/UBSan/LSan. TestPropsMatchGo compares the port with live Go cohere, Node, emitted JavaScript and yaml@2.9.0. Source Node stays unchanged.

The target and subsumer test files were read whole after reading the prior audit REPORT.md, report.json and M3.diff. Discovery verifies both names exist and records 72 names, unchanged from the audit's complete inventory. The audit itself ran only its assigned slice; this defense replays every discovered current name. Families do not change D02 uniqueness because exactly one top-level function fails.

Coverage commands: ADAMIC_YAML_LIBRARY=/tmp/u152/library timeout 120 go test -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^NAME$' -coverpkg=./internal/lower,./internal/native -coverprofile=review/test-defend/stage1-cohere-yaml-gaps/NAME.cover > NAME.coverage.log 2>&1, for the target and TestPropsMatchGo. Both pass. There are two target-only covered blocks: emit_strings.go:63 (repeat emission) and native.go:106 (unsanitized O2 flag). All other covered compiler/native blocks are shared. Go profiles cannot instrument the embedded C runtime, so no exclusive C-line coverage claim is made.

Semantic lead: sharedSliceAppend.ts builds a 128-byte heap owner, shares an 80-byte view, and appends one byte. At offset 0 it checks an observable owner byte remains 'a'. At offset 48 the view ends exactly at the owner's byte allocation boundary. A slice header's capacity must be zero even when its own reference count is one. The property resolver builds different string histories and does not exercise this owner-end append boundary. The production mutation outcomes below establish that difference empirically; shared compiler lines alone do not prove it.

D01 replaces zero shared capacity with size + 1. Target offset 0 prints x\nx\n instead of Node's a\nx\n; offset 48 triggers ASan. TestPropsMatchGo passes. TestFormatterMatchesGo also fails, so the broad mutant is not unique. Its remaining replay was stopped after that decisive counterexample; D01 has unknown results outside its completed groups. D01-stop.json records task-owned process cleanup and source restoration. D01's partial matrix never supports uniqueness.

D02 makes that capacity mistake only for a slice whose end pointer equals its ultimate owner's end pointer. This is a change to the capacity option, aimed at the semantic owner-end boundary, without matching filenames, test names or literal input bytes. Offset 0 passes; offset 48 reaches the illegal append and ASan reports heap-buffer-overflow in adamic_string_append. Only TestSharedSliceAppendMatchesNode fails. TestPropsMatchGo, TestFormatterMatchesGo, all scalar/file execution shards and every other current row pass. D02-passed-rows.json gives the complete list; D02-matrix.json records top-level and subcase outcomes; D02-runs.json gives every exact command and duration.

Each mutant uses ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml/cache/ID. Runtime library caching additionally hashes runtime contents, flags and compiler identity, so each changed C implementation rebuilds. D01.clang.log and D02.clang.log preserve successful compilation of each modified translation unit using native.Flags' ordinary C11 warning/error, floating-point and optimization flags. The matrix also builds and executes native products with ordinary and sanitizer options. After restoring production, both standalone diffs pass git apply --check against the starting main worktree. No production changes remain.

Brief costs and ambiguities:
- The mandatory whole-package baseline exhausted the 90-second binary budget without an assertion failure. Grouping all rows together is unsuitable for this package. Additional aggregate groups also timed out, so the final matrix selects rows or complete subcase sets separately. groups.json lists all selections; all 72 names have observed final results. The successful uniqueness claim is package-wide over those observations, not a slice extrapolation.
- The audit's warm library contained only yaml@2.9.0 and prettier@3.9.6. Enabling newer oracles exposed missing yaml-unist-parser, with ERR_MODULE_NOT_FOUND. No mutation was planted while this baseline was red. Installing yaml-unist-parser@3.2.1 resolved it; the repaired unist control and every mutant-witness subcase pass cleanly. Exact dependency manifests are saved.
- TestMain preflight and build-only product tests complicate the outer timeout. Listing suppresses preflight. File-driver setup, shard execution, formatter witness products and scalar shard products are all included rather than assuming a union test executes every shard.
- Go -coverprofile with -coverpkg measures compiler and emitter blocks, not C runtime execution. The defense therefore uses an explicit input-history/boundary argument and actual native mutation evidence. LLVM C coverage tools were unavailable; no C exclusive-line claim was fabricated.
- The broad D01 replay was stopped as soon as a second catcher settled that attempt. This saves expensive native rebuilds but means its failed-row list is observed, not exhaustive. The successful D02 replay is complete.
- An attempted read of the clang executable produced an exec-server output-recovery error. No source or evidence was changed by that read; normal clang compilation subsequently succeeded.
- This unit required substantially more baseline work than the audit's original assigned slice. Native C changes invalidate cached integration products, and witness/shard families need separate 90-second runs. Commands and wall durations are saved rather than conflating setup, build and test time.

Owner finding: the target's name is supported. It compares native and Node output for both offsets and also requires sanitized native execution to succeed. D02 is detected by the memory-safety leg rather than a changed unsanitized stdout; that is the demonstrated limit and strength of this defense. The test is not being defended merely by an admission check or a snapshot.

Toolchain setup skipped, env.sh works, nproc=5. API npm ci ran before baseline. No skipped final matrix rows, no test/oracle/harness changes, no other packages tested, no main push and no pull request. A unique defense was found on the second attempt, so a third mutant was unnecessary.

```json
[
  {
    "test": "TestSharedSliceAppendMatchesNode",
    "package": "stage1/cohere/yaml",
    "prior_verdict": "subsumed",
    "subsumed_by": [
      "TestPropsMatchGo"
    ],
    "defense": "defended",
    "unique_mutant": "D02 internal/native/runtime/string_share.c:59",
    "attempts": [
      {
        "mutant": "D01",
        "file_line": "internal/native/runtime/string_share.c:59",
        "change": "Give every shared slice capacity size + 1.",
        "rows_failed": [
          "TestSharedSliceAppendMatchesNode",
          "TestFormatterMatchesGo"
        ]
      },
      {
        "mutant": "D02",
        "file_line": "internal/native/runtime/string_share.c:59",
        "change": "Give a shared slice capacity size + 1 only when it ends at its ultimate owner byte boundary.",
        "rows_failed": [
          "TestSharedSliceAppendMatchesNode"
        ]
      }
    ],
    "evidence": "ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml/cache/D02 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestPropsMatchGo|TestSharedSliceAppendMatchesNode|TestLexerMutants|TestPropsMutants|TestScalarMutants)$' > D02-core.log 2>&1; scalar_runtime_gap_test.go:52: /tmp/adamic-gate/TestSharedSliceAppendMatchesNode2444970310/002/shared-sanitized: exit status 1; AddressSanitizer: heap-buffer-overflow"
  }
]
```

Timing update: complete D02 replay commands totaled 555.965 wall seconds in 24 selections. Whole-unit work took approximately 33 minutes, including dependency repair, baseline narrowing, two attempts, and publication. This exceeded the approximate 30-minute port budget while completing the expanded package matrix.

Positive pin for the former optional string indexing gap

The obsolete `?.[] on a string` row was removed in 2011e1f0167a503a195092cd721bd60dc0308db0. This follow-up adds its exact first(word) function body to optional_indexing_string_gap.a and calls it with undefined, an empty string and a present string. Node prints none, none, f on separate lines, exits 0 and writes no stderr. The JavaScript backend and native backend match byte for byte under sanitizers and leak checks.

Validation output is preserved in logs/string-gap-*.log:

- Focused uncached TestNativeAgreesWithNode for optional_indexing_string_gap.a: PASS, 0.471s after restoring the mutant.
- TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat: PASS, 0.522s, keeping the other gap rows unchanged.
- TestCountsAreRecorded with -update-counts: PASS, 36.436s. Only the new fixture row is added: 0 locals, 0 kept, 9 expressions, 10 statements, 0 functions, 0 omitted.
- run-string-gap-mutant.py restores only the former optional string indexing refusal temporarily. The new positive fixture fails with that exact NotYet diagnostic, exit 1; the production file is restored in finally. string-gap-mutant.json records the command and catcher.

No production compiler bytes or other scout changes are altered. The preserved compiler hash manifest still matches, so no census rerun was needed for this test-only follow-up. Only codex/optional-indexing is pushed.

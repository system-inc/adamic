Built: cooked-to-raw UTF-8 byte offset mapping for the current misleading-character-class claim.
Commits: this component follows pushed aa8aca95 on codex/typeaware-wave-09; no further claims.
Commands/output: verify_offsets.py PASS, 460 cases and 3411 identical bytes across Go, native, sanitizers, source Node and emitted JavaScript.
Mutant: increment each cooked byte's raw offset; builds and exits 0 with empty stderr, caught by Go comparison at byte 0.
Not covered: constructor tracking, complete regex pattern compilation and native combined-driver parity remain incomplete.

The mapper ports the production literal.CookedToRaw behavior, rather than
attempting to fix its deliberately permissive escape comparison. Its input
and output positions are UTF-8 bytes. The existing native PatternBytes helper
supplies encoding and Go-compatible rune widths. Cases include identity,
multibyte and astral characters, hex and Unicode escapes, surrogate pairs,
legacy octal, malformed escapes, mismatched texts, empty strings and trailing
line continuations/backslashes. Each mapping entry, including the end offset
and nil result, is compared to the unchanged production Go helper.

Reproduce after sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_offsets.py > /workspace/wave-09-offsets-test.log 2>&1
```

The Go oracle is built with an overlay placing only the owned testdata main
inside cohere, to satisfy Go's internal package boundary. It directly imports
the production mapper. No production Go, shared harness, registration or
compiler source is changed. Native sanitization covers ASan, UBSan and
LeakSanitizer. Source Node and emitted JavaScript both match Go. Single helper
process times: native 0.003520 seconds, Go 0.003534 seconds. Native embeds
fixtures and Go reads JSON, so this is not a full-rule throughput benchmark.
The compiler setup from this workspace remains 88 seconds and nproc 5.

All test streams and the successful mutant's differing stream are retained
in validation-offsets. This is an independently held component, not complete
constructor diagnostics. Integrating it requires the still-missing constructor
reference/constant-expression adapter. The full invalid-regexp pattern engine
also remains absent. The combined native literal parser/listener still lacks
certification because the shared freshness/lowering analysis did not finish
within the previously measured 120-second timeout. Its exact stacks and attempts
remain in validation-resume; this component does not claim to repair that
blocker. Prior released-handle checks remain in REPORT.md and were not rerun.

Origin was refreshed again. No new rules were selected because these current
claims remain incomplete. The supported native components are pushed for
review while the shared compiler boundary remains outside this unit's allowed
territory. The full corpus/fixes/suggestions bar has not been met for either
complete regex rule.

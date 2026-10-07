# Slot 02 eighth helper batch

| File | Actual Go contract |
|---|---|
| escape_class_rune.a | regexp.EscapeClassRune: escape exactly backslash, close bracket, caret, dash and open bracket; preserve Go string(rune) conversion. |
| identity_escape.a | regexp.identityEscape: Unicode syntax-character restriction, slash exemption and class-only dash exemption; preserve decoded fields, width, error text and wrapped syntax-error cause. |
| literal_rune.a | regexp.literalRune: ASCII alphanumeric passthrough, raw supplementary runes, minimum-width-four lower-case signed hexadecimal for every other rune. |

Each helper has four remaining consumers. RULES.md lists every rule; readiness.json removes twelve prerequisite entries across four rules and preserves their remaining blockers. This is a helper handoff, not four completed rules.

Rune inputs are Go signed int32 values. Fractional values and values outside that range refuse through panic. Caller-provided sizes are exactly representable integer Go widths; tests also check zero, negative and larger synthetic widths. Context carries Unicode/class/group-count/named-group fields; identityEscape consults only Unicode and class. The caller must translate the returned error/cause pair into its error arena if it needs object identity. A nonempty cause names the actual ErrUnsupportedSyntax sentinel; the Go exporter independently checks both errors.Is and errors.Unwrap. Successful results carry no error or cause. Error results preserve Go's zero decoded struct.

EscapeClassRune replaces negative, surrogate and out-of-range scalar values with U+FFFD when converting a rune to text, as Go string(r) does. It decides escaping from the original rune number. IdentityEscape preserves the original numeric rune on success and uses replacement text only for formatting an error. Without Unicode mode every identity escape succeeds. With Unicode mode only syntax characters, slash and an in-class dash succeed. All returned kind/set/negated fields match the real private decodedEscape struct.

LiteralRune deliberately follows a different conversion path. ASCII letters/digits and values above U+FFFF are written using Go rune-to-string conversion. All other values use Go fmt's signed %04x: surrogates remain numeric escape text, and negative values include a sign within the minimum width. The port implements hexadecimal formatting explicitly, with no truncation and no Unicode normalization.

The independent oracle calls actual exported/private Go helpers at pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db through temporary overlay exports. Helper bodies remain untouched. The corpus contains 641 runtime source/options captures from all four consumers, 1,735 string option values, 272 rune values, eight contexts and six size controls. All bytes through U+00FF, invalid scalar boundaries, supplementary values, all syntax characters, slash/dash and negative formatting are included. All decoded fields, errors and causes are compared. Output strings use decimal UTF-16 units so NUL and other controls do not alter output framing. Want is removed before Adamic receives the corpus.

Capture overlays observe ordinary and typed fixture entry points in temporary Go files. Options are marshaled at runtime; typed-fixture temporary directory prefixes are normalized to <fixture> for reproducibility. Every consumer must be observed and the Go Core, Next and TypeScript rule packages must exit zero. Sources and coverage regenerate byte for byte. No shared registration, Adamic rule harness or cohere worktree source is edited.

Source Node, emitted JavaScript and ASan/UBSan native must all compile and exit zero without stderr. They must agree with each other before a mismatch with Go can count as a semantic mutant caught. There are thirteen retained semantic mutants, with every witness in REPORT.md and evidence/. Compiling failures, crashes and sanitizer reports are not credited.

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot02/batch8/testdata/regenerate.py > /tmp/slot02-batch8-capture.log 2>&1
go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch8$' -count=1 -v -timeout=20m > /tmp/slot02-batch8-isolated.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > /tmp/slot02-batch8-full.log 2>&1
```

Ownership: decodeControlEscape was withdrawn to slot 05's earlier 94e0c374 claim and its duplicate .a port was deleted. It is not delivered, counted as readiness or credited for mutants. The replacement literalRune claim was pushed in 0a261996 before code. Only three helpers are retained.

The corpus is bounded, not an exhaustive Unicode sweep or arbitrary-regexp fuzzer. Native regexp compilation/matching, whole-rule findings/fixes/suggestions and malformed adapter inputs remain outside scope. No new rule is introduced, so no rule.json kinds registration or shared Diagnostic integration is changed.

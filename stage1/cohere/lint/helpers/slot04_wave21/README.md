# Class tokens and bare utility arguments

classTokensOf and classTokensIn accept UTF-8/raw byte arrays and a byte whitespace predicate matching Go isSpace. Token text stays a byte array; start/end remain Go byte offsets. classTokensIn validates bounds and exact source/value byte equality, returning ok false and present false on refusal. Successful empty input returns an allocated empty token list, matching Go. classTokensOf accepts an arbitrary offset and groups consecutive bytes with the same separator classification. It does not decode Unicode or convert positions to UTF-16. A future finding builder performs that conversion once at its boundary.

resolveBareArgument takes an argument name, candidate value and fraction plus explicit inference, slash segmentation, positive-integer, spacing-multiplier, Go trim and value-parser callbacks. Only number, integer, ratio and percentage are permitted. Ratio reads Fraction, requires exactly two positive integer components and parses the space-padded slash spelling. Number must satisfy the canonical quarter-multiple guard; percentage must have an integer part. Its generic parser result preserves the caller's node representation; the test projection compares canonical CSS text, not every parser node field.

The private Go overlays call actual classTokensOf, classTokensIn and utilityEvaluation.resolveBareArgument. The Go oracle supplies byte whitespace, InferDataType, segment, IsPositiveInteger, isValidSpacingMultiplier, strings.TrimSpace and ParseValue observations for the Adamic callbacks. These are explicit dependencies, not newly claimed implementations. No regex is translated or replaced by a hand-written matcher in these helpers.

Run from the repository root after sourcing the setup environment:

```
python3 stage1/cohere/lint/helpers/slot04_wave21/testdata/regenerate.py > /tmp/slot04-wave21-regeneration.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave21/testdata/capture.py > /tmp/slot04-wave21-capture.log 2>&1
go test -count=1 -v -timeout=15m ./stage1/cohere/lint/helpers/slot04_wave21 > /tmp/slot04-wave21-tests.log 2>&1
```

Capture uses temporary Go overlays and a temporary pinned Tailwind 4.3.3 fixture directory. All seven inventory consumers have asserted fixtures. Token helper inputs are captured live and replayed separately. Bare resolution was not reached in this selected consumer suite; direct actual-Go controls cover its branches. Consumer source texts are extra byte-token controls, not whole native rule executions. REPORT.md lists every mutant, dependency removal and coverage limit.

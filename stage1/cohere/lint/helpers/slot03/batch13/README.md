# Data-type inference and family names

Each helper has its own `.a` file. Predicates and CSS segmentation are explicit dependencies:

- `inferDataType(value, types, matches)` returns empty immediately for the exact lowercase var( prefix. Otherwise it checks types in the caller's order and returns the first matching string, or empty. The matcher must preserve Go's pure predicate behavior and must not mutate the supplied type list.
- `matchesDataType(value, dataType, check)` routes seventeen exact data-type names to the predicate index below. Unknown names return false without calling any predicate. The check callback implements the corresponding Go predicate on unchanged input. This is CSS data-type dispatch, not lint listener or AST-kind routing.
- `isFamilyName(value, segment)` obtains comma parts from a separately owned CSS segmenter, rejects any part beginning with an ASCII digit, skips parts beginning with exact lowercase var(, counts all other parts including empty strings, and succeeds only when at least one part is counted. The segmenter must preserve Go's final empty part and quote/escape/nesting behavior; no trim is added.

Fixed prefix and digit tests are JS RegExp literals /^var\(/u and /^[0-9]/u. They reproduce Go string-prefix and byte-range operations. There is no Go regexp compile site in data_type.go and no applicable row in the shared regex table checked at 071fb012. No hand-rolled matcher or runtime option regex is introduced. No finding byte offset is constructed. Invalid UTF-8 and unpaired UTF-16 are outside the Unicode adapter boundary.

| Predicate index | Data type | Go dependency |
|---:|---|---|
| 0 | color | IsColor |
| 1 | length | IsLength |
| 2 | percentage | isPercentage |
| 3 | ratio | isFraction |
| 4 | number | isNumber |
| 5 | integer | IsPositiveInteger |
| 6 | url | isURL |
| 7 | position | isBackgroundPosition |
| 8 | bg-size | isBackgroundSize |
| 9 | line-width | isLineWidth |
| 10 | image | isImage |
| 11 | family-name | isFamilyName |
| 12 | generic-name | isGenericName |
| 13 | absolute-size | isAbsoluteSize |
| 14 | relative-size | isRelativeSize |
| 15 | angle | isAngle |
| 16 | vector | isVector |

The temporary Go overlay preserves helper logic and adds top-level observations at actual switch returns and the family segment call. Actual Go supplies all seventeen predicate verdicts and comma segments. Test-only Adamic adapters return those observations and record unchanged input and correct route/segment requests. This proves these compositions and family logic, not the native implementations of every dependency. Consumers must wire the real predicate functions and segmenter before rule integration.

All four consumers supply 118 runtime sources. Complete sources, derived tokens, number/unit/math controls and family-name prefix/comma/quote/escape/nesting controls yield 727 strings. Each is checked against 21 known/unknown type selectors, 26 caller type orders and the family predicate: 48 observations per string, 34,896 verdict/trace lines. Go, source Node, emitted JavaScript and ASan/UBSan native must agree. Twenty-nine variants compile and run without stderr before their wrong behavior is caught: one per dispatcher arm, unknown routing, unchanged input, var guard/case/order/input and six family digit/skip/count/argument changes.

Source /workspace/adamic-tools/env.sh, then run `go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m` with output redirected directly to a log. Regenerate with `python3 stage1/cohere/lint/helpers/slot03/batch13/testdata/regenerate.py`. The known external Tailwind/corpus capture guards still fail; helper parity is not a passing whole-rule gate. Exact consumers and residual blockers are in readiness.json.

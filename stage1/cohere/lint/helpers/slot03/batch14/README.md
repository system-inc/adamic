# Line widths, images and background positions

Each helper has its own `.a` file. CSS segmentation and numeric/URL predicates are explicit dependencies supplied by consumers.

- `isLineWidth(value, dependencies)` segments on a space. It first accepts a length or number, in that short-circuit order, then the exact keywords thin, medium and thick. Any other part rejects the value; finishing the loop accepts it.
- `isImage(value, dependencies)` segments on commas, skips exact lowercase var( prefixes, then accepts URLs, six gradient prefixes and element/image/cross-fade/image-set prefixes. It rejects any other part and requires at least one accepted non-variable part. Prefix matching intentionally accepts unclosed function prefixes as Go does.
- `isBackgroundPosition(value, dependencies)` segments on spaces, counts exact center/top/right/bottom/left keywords, skips exact lowercase var( prefixes, then accepts lengths or percentages in that short-circuit order. An invalid part rejects the value; at least one non-variable part must count.

Fixed prefix checks use JS RegExp literals with the u flag. The source Go operations are string-prefix comparisons, not regex compile sites; the shared regex table at 071fb012 has no data_type.go row. Exact keyword sets use Array.includes. No option pattern, hand-written regex engine or finding position is introduced.

The owned temporary Go overlay keeps the three helper bodies unchanged and wraps actual segment/length/number/percentage/URL dependencies to record their calls. Go supplies actual segments and predicate results. Test-only Adamic callbacks replay those observations and record unchanged arguments, separators and evaluation order. This verifies helper composition; native implementations of every dependency must be wired separately before whole-rule integration. Duplicate parts use the first equal part index in both trace adapters because the observed predicates are pure.

The four consumers supply 118 runtime sources. Full sources, extracted tokens, numeric/unit/math controls, 32-by-32 space/comma part pairs and image-prefix controls produce 2,870 unique strings, each checked with all three helpers: 8,610 verdict/trace lines. Real Go, source Node, emitted JavaScript and ASan/UBSan native agree. Twenty-three semantic mutants must compile and run without stderr before their changed stdout counts as a catch.

Source /workspace/adamic-tools/env.sh and run `go test ./stage1/cohere/lint/helpers/slot03 -run TestBatch14 -count=1 -v -timeout=20m`, redirecting output directly to a log. Regenerate with `python3 stage1/cohere/lint/helpers/slot03/batch14/testdata/regenerate.py`. The capture records known upstream external-engine/corpus guard failures; it is not a passing whole-rule gate. Exact consumers and remaining blockers are in readiness.json.

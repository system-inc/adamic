# Angle, number and percentage composition

Each helper lives in its own `.a` file. Matcher dependencies are explicit callbacks, not additional implementations delivered here:

- `isAngle(value, numberWithSuffix)` passes the unchanged value and the exact ordered units deg, rad, grad, turn. It returns that matcher result without any math fallback.
- `isNumber(value, dependencies)` calls scanNumber once, accepts a positive scan consuming the entire spelling, otherwise calls hasMathFunction once. The scanner must implement Go's ASCII numeric-prefix contract. Its count is an ASCII prefix length, identical in bytes and UTF-16 units; for any remaining non-ASCII tail, whole-input comparison fails in both representations. No finding offset is constructed here.
- `isPercentage(value, dependencies)` passes the unchanged value and the one-element percent suffix list to numberWithSuffix. Only a false result invokes hasMathFunction.

No Go regex is ported and no new hand-rolled matcher is added. Number scanning, suffix matching and math detection retain separate ownership; these helpers compose them. Consumers must wire the real implementations before whole-rule integration. All string inputs are valid UTF-8-derived Unicode; invalid UTF-8 and unpaired UTF-16 are outside the boundary.

The owned Go overlay renames only the dependency entry points, leaves actual predicate bodies unchanged, and wraps those dependencies to capture top-level calls. It executes the original dependency bodies and suppresses nested scan tracing inside numberWithSuffix. The case file records actual Go scanner/suffix/math observations; the Adamic test adapter returns them after checking value and suffix arguments. This independently checks both predicate verdicts and lazy dependency behavior, but does not prove matcher implementations run in native.

All four consuming rules supply 118 runtime sources. Complete sources, derived tokens, ASCII numeric grammar controls, every unit, every math-function name, Unicode tails, case, whitespace and NUL variants yield 690 strings and 2,070 verdict/trace lines. Go, source Node, emitted JavaScript and sanitized native must agree. Eight compiling semantic variants are caught only after successful execution with no stderr. The adapter's first closure version was refused for a possible ownership cycle; file-level functions and explicit state resolve it inside this directory.

Run `source /workspace/adamic-tools/env.sh`, then `go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m`, writing output directly to a log. Regenerate with `python3 stage1/cohere/lint/helpers/slot03/batch12/testdata/regenerate.py`. Capture still records known unavailable external Tailwind/corpus failures; it is not a passing whole-rule gate. Exact consumers, residual blockers and all witnesses are in readiness.json and REPORT.md.

# ecmascript/literal package claim

Branch: lint-helpers/ecmascript-literal. Base: origin/area/stage1-lint.
Triage d7ab0bc4 rank 22; no retained port, four missing helpers: CookedToRaw,
cookedBytesProducedBy, escapeWidthInStringLiteral, producesNoCookedBytes.
All 1,516 origin refs checked for ports and all helper branch claims inspected.
Earlier eligible packages are reserved, excluded tonight, or require runtime regex.
Yielded rules/next to earlier remote dcaf0a30 (local claim never pushed).
No literal package reservation or implementation found; regexpattern's dependency
mention is not a literal reservation. This package scans literals, without regex compilation.

Consumer: no-regex-spaces. Forecast: zero alone; one additional with preceding
complete packages (131 cumulative). Its other dependency ecmascript/regexpattern
is reserved: use its completed shared port if available, otherwise stop dependent
rule work explicitly. Counts are conditional, not findings parity claims.

Claim before code; fetch again and yield to earlier competing claim. Build fresh,
one helper per file; actual Go consumer captures compared across Node, emitted
JavaScript and sanitized native; one caught Node/native mutant per helper.
Prove the consuming rule if its shared prerequisites are available. Run helpers
and lint with every input set before the finished-unit implementation push.


## Delivered independent slice and explicit stops

Built fresh: cookedBytesProducedBy and producesNoCookedBytes, under ../literal/.
Live unchanged-Go capture: no-regex-spaces (459 produced / 1 tail call),
no-misleading-character-class (1,182 produced / 4 tail calls); 82 controls.
No donor implementation was used. Capture script and coverage are in ../literal/testdata/.

Stopped escapeWidthInStringLiteral: cohere/internal/lint/ecmascript/literal/offsets.go:117
calls ecmascript/regexsyntax.AllHexDigits, and lines 122-123 call IsHexDigit.
That package remains claim-only on origin/lint-helpers/ecmascript-regexsyntax;
its package owner holds the smaller dependency. CookedToRaw (offsets.go:33)
is stopped on escapeWidthInStringLiteral. No private duplicate is supplied.
Rule no-regex-spaces is stopped on that closure and ecmascript/regexpattern,
also claim-only on its owner's branch. No consuming rule has been ported and
no new complete rule unblock is claimed by this independent slice.

## Verification

Cohere pin: 7945d102a6c18dd36adf9114a758ce646e8b2359.
Both independent helpers match fresh Go on 1,646 consuming-rule calls and 82
controls: 3,509 identical bytes on source Node, emitted JavaScript and ASan/UBSan
native. One compiling output mutant per helper is caught on all three backends.
The tests enforce both consumer capture counts and compare recorded answers to
fresh private Go functions, not a handwritten oracle. See ../literal/helpers_test.go
and ../literal/testdata/coverage.json; regenerate through testdata/capture.py.

Complete helpers tree passed with no skips. Complete lint coverage used the pinned
TypeScript compiler input, benchmarks, profile generation and profile snapshots.
Upstream parity: 4,576 source/rule/options combinations, 13,980,970 identical bytes.
All 83 registered mutants passed on Node, emitted JavaScript and native.
The parallel full run's sole failure was TestCompilerAndStage1Agree exceeding the
harness's ten-minute native-command limit. Its isolated -parallel=1 retry passed
all 918 files in 346 seconds: 30,293,066 identical bytes across all four runtimes.
The existing TestCheckerBridgeRefusalPending skip remains; no input set was omitted.

Dependency branches were fetched again before finishing: regexsyntax remains at
f997f145cceca508f0c2767778fe503463c31354 and regexpattern at
df0e695283113b123754908ca456330a16d20b41, both claim-only.
This is a verified two-helper slice, not a complete four-helper package or rule proof.

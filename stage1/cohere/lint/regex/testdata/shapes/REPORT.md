Fixtures branch from area/stage1-lint db2ecc004; approved table brought forward from 071fb0128.
107 fixtures: plain 76, (?i) 4, (?s) 1, (?m) 1, dynamic 25; 840 inputs and 341 whole matches.
Outside the port's shapes: zero selected instances; string arguments, no listed V8-divergence shapes.
PaginationInput to PaginationInpuX mutant executes cleanly and fails the fixture naming its Go line.
Native is pending on codex/regex-runtime-compiler; forced native fails, automatic required-pass transition tested structurally.

Observed verification:

- check.py --runtime-js-compiler /tmp/regex-library-adamic passes 107 source Node,
  107 area constant-string JavaScript, 107 runtime source Node and 107 runtime
  JavaScript comparisons from the unmodified library 6e47d677e compiler.
- Area runtime lowering reports 107 named nonconstant-pattern refusals. The first
  fixture forced to native fails with that refusal, not a comparison/compiler
  substitution. Once lowering accepts strings, native.Build with Sanitize true
  and the runtime JavaScript backend must pass. No native runtime pass is claimed.
- The refreshed gate independently recomputes every Go regexp match and UTF-16 span.
- The original TestFixedPatterns and TestInventoryMatchesPinnedSource pass in
  21.676s: all 82 table literals still agree on Go, Node, JavaScript and sanitized
  native. Its original mutant is caught on all three backends. The new planted
  mutant is one character and is separately caught by its per-site source comparison.
- go vet on the testdata gate is clean. All test output is in evidence logs.

Commands (toolchain environment sourced):

```
python3 stage1/cohere/lint/regex/testdata/shapes/check.py --runtime-js-compiler /tmp/regex-library-adamic > /tmp/regex-shapes-check.log 2>&1
go test ./stage1/cohere/lint/regex -run 'TestFixedPatterns|TestInventoryMatchesPinnedSource' -count=1 -v -timeout=20m > /tmp/regex-shapes-table.log 2>&1
go vet stage1/cohere/lint/regex/testdata/shapes/gate.go > /tmp/regex-shapes-vet.log 2>&1
```

Toolchain setup: go ready 0.072s, node ready 0.077s, clang ready 0.551s,
submodules ready 32.128s, go build ready 344.253s, cache warm 344.529s,
done 344.605s; nproc 5, CPU quota 4. Setup's timing lines are cumulative.

Limits: one bounded binding per dynamic expression, not exhaustive option strings;
whole matches only, not captures/replacements/split APIs. Runtime flags are exercised
as strings but omitted/undefined/RegExp-object arguments and SyntaxErrors are not part
of this valid-pattern census. No runtime/compiler changes, no rule migrations, no
shared harness changes, no submodule bump and no full repository gate. The area
runtime JavaScript path shares the native lowerer refusal; library JavaScript proof
uses a separately built, unmodified 6e47d67 checkout, not a merged dependency.

Built real runtime RegExp constructors for three rule consumers; removed the production warning hand matcher.
Commit: see branch codex/lint-regex; constant table unchanged from 071fb0128.
Go esregexp and source Node agree on 1,621 matcher observations and 340 fixture finding outputs.
Three clean-running mutants were caught only by the Go comparison.
Emitted JavaScript and sanitized native are pending codex/regex-runtime-compiler; shared consumer integration remains pending.

The option oracle is cohere's own esregexp package, not Go regexp. The capture script
uses a temporary Go overlay that changes these three rule consumers to that package
and observes the real upstream tests. It does not edit cohere or the shared harness.
This explicitly models #7mztrdd before that change lands. All selected upstream
rule tests pass their own assertions. Regenerate with:

```
python3 stage1/cohere/lint/regex/testdata/capture_migrations.py > /tmp/regex-capture.log 2>&1
```

Matcher observations: id-length 135, no-inline-comments 136, no-warning-comments
1,350; 16,360 identical trace bytes. Mutants respectively invert the successful
exception result, negate ignorePattern.test, and remove the warning `i` flag.
Each loads, executes cleanly and differs from the oracle. No matcher fallback exists.

Finding comparisons include IDs, messages, UTF-16 spans and unchanged whole fixed
sources (these rules have no fixes): id-length 202 fixtures, 108 findings, 30,552 bytes;
no-warning-comments 84 fixtures, 67 findings, 25,980 bytes; no-inline-comments
54 fixtures, 37 findings, 23,442 bytes. Id and warning use the real stage1 parser.
Inline uses source and comment geometry projected from Go, including the empty-JSX
fact; its predicate is independently evaluated, but parser geometry is not independently
validated. Finding byte offsets become UTF-16 once at the Go capture boundary.

Production comments.ts consumes the warning constructor and fixed self-directive
literal. The id consumer is absent on main: integration/id-length.patch wires the
parked consumer, and a credited private snapshot verifies its complete findings.
Inline's complete predicate and numeric SourceFile listener are owned here; its
shared comment adapter and registry integration remain pending #zmh9v36. No shared
harness or compiler files were changed. See integration/README.md for exact limits.

Both emitted JavaScript and native currently fail in the shared lowerer with
`stage 0 can't lower RegExp with a nonconstant pattern yet`. They are pending,
not skipped or replaced by source Node. The tests attempt lowering and accept only
this exact blocker; when it clears they enforce both emitted JavaScript and sanitized
native, including mutants. Dependency: codex/regex-runtime-compiler, Codex 01a114e3,
approved with dynamic_gap.a as acceptance, due October 9 18:00 MDT.

The constant table stays unchanged: 107 rows, 82 fixed and 25 dynamic; fixed classes
plain 76, (?i) 4, (?s) 1, (?m) 1, and zero property/end-z/named-group sites.
All 82 fixed translations retain the three-backend gate. The separately pinned
esregexp Unicode property escape gap remains in gaps.md. No throughput measurement
for dynamic native paths is possible before lowering support lands.

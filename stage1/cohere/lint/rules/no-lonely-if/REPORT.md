Built: no-lonely-if in owned .a modules, with a directory rule descriptor.
Commits: claim 38b0f50e was pushed before implementation; port commit follows this report.
Checks: original Go assertions and Node source, emitted JavaScript and ASan/UBSan native comparisons passed.
Mutant: An extra space inserted into the independent block replacement. Compiled and exited zero on all three runtimes; only Go byte comparison caught it.
Limits: current shared registration still requires .ts; shared infrastructure was not edited.

Fixture findings, messages, UTF-16 ranges, every fix and converged source match Go byte for byte: 33 files, 24 findings, 10967 serialized bytes. The pinned TypeScript revision is 050880ce59e30b356b686bd3144efe24f875ebc8. Its 77 src/compiler files plus all 159 stage1 sources yielded 44 findings and 12446755 identical serialized bytes across all four paths.

Best of five interleaved end-to-end count runs, including startup, read and parse: native 37.22, Node 53.53, Go 230.30 findings/second. Native timing uses an unsanitized release build; correctness uses sanitized native with clean stderr. Timing uses the 44 natural corpus findings.

Reproduction from the repository root:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned python3 stage1/cohere/lint/rules/no-lonely-if/validate_complete.py > /tmp/wave14-fifth-run.log 2>&1
python3 stage1/cohere/lint/rules/no-loss-of-precision/validate_edges.py /tmp/adamic-gate/wave14-fifth-htxq30aj > /tmp/wave14-fifth-edges-run.log 2>&1
```

The validator invokes the real upstream Go tests first, then captures their asserted cases through an overlay of those upstream test files. Owned runners serialize the full repair model independently of the current shared harness. No shared Adamic harness, registration generator or compiler files are changed. See [complete evidence](../no-lonely-if/evidence/complete.log) and [edge evidence](../no-loss-of-precision/evidence/edges.log).

Independent fix range covers the enclosing else block while the diagnostic covers its inner if. The current shared fixer instead uses the diagnostic range. The inspected harness branch adds .a and suggestions but does not resolve this range mismatch. The factory explicitly refuses enabled shared execution instead of returning a misleading fix. Owned execution includes Unicode trivia, overlapping fixes and the ten-pass convergence budget. An additional eleven-pass mutant compiles and runs cleanly but differs from Go on all three runtimes.

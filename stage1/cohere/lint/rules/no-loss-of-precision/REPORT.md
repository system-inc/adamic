Built: no-loss-of-precision in owned .a modules, with a directory rule descriptor.
Commits: claim 38b0f50e was pushed before implementation; port commit follows this report.
Checks: original Go assertions and Node source, emitted JavaScript and ASan/UBSan native comparisons passed.
Mutant: Rounded parser text substituted for the raw numeric source lexeme. Compiled and exited zero on all three runtimes; only Go byte comparison caught it.
Limits: current shared registration still requires .ts; shared infrastructure was not edited.

Fixture findings, messages, UTF-16 ranges, every fix and converged source match Go byte for byte: 152 files, 57 findings, 38186 serialized bytes. The pinned TypeScript revision is 050880ce59e30b356b686bd3144efe24f875ebc8. Its 77 src/compiler files plus all 159 stage1 sources yielded 0 findings and 12409596 identical serialized bytes across all four paths.

Best of five interleaved end-to-end count runs, including startup, read and parse: native 402.15, Node 541.73, Go 2335.23 findings/second. Native timing uses an unsanitized release build; correctness uses sanitized native with clean stderr. Timing uses 500 witness copies because the natural corpus has no findings.

Reproduction from the repository root:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned python3 stage1/cohere/lint/rules/no-lonely-if/validate_complete.py > /tmp/wave14-fifth-run.log 2>&1
python3 stage1/cohere/lint/rules/no-loss-of-precision/validate_edges.py /tmp/adamic-gate/wave14-fifth-htxq30aj > /tmp/wave14-fifth-edges-run.log 2>&1
```

The validator invokes the real upstream Go tests first, then captures their asserted cases through an overlay of those upstream test files. Owned runners serialize the full repair model independently of the current shared harness. No shared Adamic harness, registration generator or compiler files are changed. See [complete evidence](../no-lonely-if/evidence/complete.log) and [edge evidence](../no-loss-of-precision/evidence/edges.log).

Additional fixtures compare 1,200 deterministic decimal lexemes and 300 binary/octal/hex lexemes. Together with explicit precision boundaries, edge comparison yields 1,002 findings and 578,507 identical bytes. This includes Go signed-64 exponent bounds, unsigned-64 integer parsing limits, significant-digit limits, trailing zeros and BigInt exemption.

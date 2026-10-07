Built: no-invalid-regexp's flag frontend now visits and reports its handed node; named listener metadata declares node:true.
Commits: follows 3ebe2040db7e on wave-09, based on area d65a8f931 and main 39638d9e2; no new claims.
Commands/output: frontend, handed-node and listener comparisons PASS; required expression and whole-AST compiler comparisons PASS with real pinned inputs.
Mutants: a clean-running report-refetch mutant differs from Go at byte 58; the named listener wrong-kind mutant differs at byte 276.
Not covered: complete regex pattern validation, shared rule installation, remaining visitor migration, unrelated required checks or the full gate.

NoInvalidRegExp.visit receives ParseNode directly. Its report path uses that
same node for the finding range. The rule no longer scans every parser node
or checks its handed node's kind for relevance. The isolated driver fetches
each candidate once and dispatches CallExpression/NewExpression; child kinds
are still read to interpret the callee and arguments. No shared files changed.

After sourcing /workspace/adamic-tools/env.sh, the commands were:

```
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_frontend.py > /workspace/wave-09-handed-flags-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_handed.py > /workspace/wave-09-handed-flags-contract.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/listeners/verify.py > /workspace/wave-09-handed-flags-listeners.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-09-required-typescript go test ./stage1/typescript/parser -run '^TestCompilerExpressionsAgree$' -count=1 -timeout=30m -v > /workspace/wave-09-required-parser-pinned.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-09-required-typescript go test ./stage1/typescript/parser -run '^TestWholeCompilerAgrees$' -count=1 -timeout=30m -v > /workspace/wave-09-required-whole-pinned.log 2>&1
```

The frontend matches production Go on 40 controls, 35 findings and 6020
serialized bytes, including fixes and suggestions. Native normal and
ASan/UBSan/LeakSanitizer runs pass. Single-run native/Go execution times are
0.065247/0.070671 seconds; sanitized execution is 0.077687 seconds. These
are process timings, not a throughput claim. Setup was reused; nproc is 5.

The contract test supplies index zero while preserving the correct node for
every callback. Both normal and sanitized runs retain the Go bytes. A mutant
refetches parser.node(0) only in report; it builds and exits zero with empty
stderr, and only the Go comparison catches its incorrect range at byte 58.
Listener checks match production Go's named kinds (291 bytes), native,
sanitizers, source Node and emitted JavaScript. Wrong-kind output is caught
at byte 276. Existing released-handle evidence remains in the preceding
landing report; those checks were not repeated for this migration.

The first required expression-corpus run failed explicitly with "corpus pin
differs: \"\" exit status128": the existing corpus was a tar extraction
without Git metadata. This was corrected by fetching a real shallow Git
checkout of microsoft/TypeScript at
050880ce59e30b356b686bd3144efe24f875ebc8 into
/workspace/wave-09-required-typescript. No input pin or check was weakened.
Expression comparison then passed on 77 sources and 28836875 identical
bytes (24.88 seconds); whole-AST comparison passed on the same sources and
44766682 identical bytes (21.26 seconds). Both checks compare independent
Go, source Node and sanitized native outputs. Other required stage1 checks
were not run, and no full gate is claimed.

Both regex claims remain incomplete. The pattern compiler callback still
fails explicitly when exercised; native new RegExp(nonconstant, 'u') is
independently refused in REGEX_POLICY_REPORT.md. No hand-rolled matcher or
new regex parser was added. Compiler/repository lint parity from the preceding
landing is retained, not claimed as newly rerun here. No new claims or React
parking. Publication is only to codex/typeaware-wave-09.

Logs, comparison streams and archived mutant probes are retained in
validation-handed-flags. The initial missing-input failure is included beside
the successful pinned-corpus runs.

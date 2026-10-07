Built: six owned per-rule rule.json declarations with numeric kinds matching production Go listener maps.
Commits: follows 2d30e7bc on codex/typeaware-wave-09; origin/main e8ba3d5d remains an ancestor; no new claims.
Commands/output: listeners/verify.py PASS, six declarations and 168 identical bytes across Go, native, sanitizers, source Node and emitted JavaScript.
Mutants: radix kind 214 changed to 215 in JSON and the .a declaration; Go comparison catches each at byte 166; native mutant builds and exits 0 with empty stderr.
Not covered: shared numeric dispatch, handed-node execution, complete regex engine and constructor tracking, or a new full corpus sweep.

The requested rule.json files live in listeners/<rule>/rule.json and contain
only name and numeric kinds. They do not invent visitor/factory registration
for the incomplete components. Existing .a declarations remain independently
checked against the same production Go map. The registration branch inspected
read-only still uses string kind names; it was neither modified nor merged.

Reproduce with the existing toolchain environment:

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave09_core/listeners/verify.py > /workspace/wave-09-rule-json-test.log 2>&1
```

Native declaration printing took 0.001488 seconds; Go took 0.015783 seconds.
These workloads print metadata, so this is not a full lint speed comparison.
Setup is reused from the successful 88-second setup, nproc 5.
Evidence is retained in validation-rule-json, including both mutant streams.

The shared blocker remains exact: ParseNode.kind is a string and Rules.ask
refetches nodes by index; no numeric handed-node listener contract is exposed.
Legacy rule execution is consequently not claimed speed-compliant. Shared
parser, harness, registration generator and compiler files remain untouched.
No batch-8 Diagnostic SHA has been supplied in this conversation. The complete
regex rules remain unfinished, so no further rules were claimed. Prior corpus,
released-handle and sanitizer evidence is recorded in LANDING_REPORT.md and
the component reports; no checker ABI or finding implementation changed here.

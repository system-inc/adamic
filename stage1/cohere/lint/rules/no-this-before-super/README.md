# no-this-before-super

Base: lint-helpers/rules-core at 22965fd8217a1fdc27e242e117140ec933a1bd46. All fetched origin branches were checked for an existing directory and rule claim before implementation; a fresh 1,464-ref scan immediately before committing found neither. constructor-super remains reserved by codex/lint-wave1-07 and is not included here.

The listener preserves Go's per-constructor report, parameter-default order, separate evaluation contexts, branch credit, short-circuit behavior, loop handling, try/finally entry state and switch fallthrough. Its local structural projection feeds core.isSeparateEvaluationContext and core.bodyDefinitelyExits (including their shared try/switch helpers). No shared source, compiler or harness is changed.

Pinned Go: 7945d102a6c18dd36adf9114a758ce646e8b2359. The unchanged upstream tests capture 96 unique source/file/options cases from 195 runs. TestRulesAgree passes all 4,010 discovered upstream cases: 13,813,010 identical bytes on Go, source Node, emitted JavaScript and ASan/UBSan native. TestCompilerAndStage1Agree passes 898 compiler/stage1 files: 30,217,023 identical bytes. The compiler checkout is TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. TestOwnedWitnesses passes, including this directory's finding-producing witness.

With the setup environment sourced:

```
ADAMIC_TYPESCRIPT_SOURCE=/path/to/pinned/TypeScript go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestCompilerAndStage1Agree)$' -count=1 -v -timeout=20m
python3 stage1/cohere/lint/rules/no-this-before-super/testdata/native_mutant.py
```

The rule-local runner selects this rule as the native mutant canary through a temporary Go overlay. The compiling violation-reversal mutant disagrees with unchanged Go on Node and emitted JavaScript; sanitized native equals mutated Node (502,129 bytes). Inherited parser-recovery refusals remain explicit; none belongs to this rule's captured cases.

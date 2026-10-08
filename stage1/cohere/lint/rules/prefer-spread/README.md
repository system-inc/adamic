# prefer-spread

Base: lint-helpers/rules-core at 22965fd8217a1fdc27e242e117140ec933a1bd46. All origin branches were checked before implementation; the fresh 1,505-ref pre-push scan found no rule directory or rule claim.

The listener preserves Go's apply spelling, argument count, array/spread exclusions, receiver unwrapping and token comparison. Its local projection reuses core.memberExpression, core.isNullOrUndefined and core.hasSameTokens through the shared signatures adapter. Projection is lazy after the candidate-call gates. No shared compiler or harness source is changed.

Pinned Go: 7945d102a6c18dd36adf9114a758ce646e8b2359. Unchanged upstream tests capture 61 unique source/file/options cases from 122 runs. TestRulesAgree passes all 3,991 discovered cases: 13,807,216 identical bytes on Go, source Node, emitted JavaScript and ASan/UBSan native. TestCompilerAndStage1Agree passes 898 files: 30,288,946 identical bytes, using TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. TestOwnedWitnesses passes, including this rule's finding-producing witness. TestJsxLintTrees passes 63 captured sources with 51,432 identical whole-tree bytes; its frozen map needed no edit or merge.

With the setup environment sourced:

```
ADAMIC_TYPESCRIPT_SOURCE=/path/to/pinned/TypeScript go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestOwnedWitnesses|TestCompilerAndStage1Agree|TestJsxLintTrees)$' -count=1 -v -timeout=20m
python3 stage1/cohere/lint/rules/prefer-spread/testdata/native_mutant.py
```

This scratch worktree used GOFLAGS=-buildvcs=false because its oracle submodule was a read-only symlink; only build metadata stamping is disabled. The rule-local mutant runner uses a temporary overlay to select this rule as native canary. The compiling null-receiver-reversal mutant disagrees with unchanged Go on Node and emitted JavaScript; sanitized native equals mutated Node (499,013 bytes). Inherited parser-recovery refusals remain explicit; none belongs to this rule's captured cases.

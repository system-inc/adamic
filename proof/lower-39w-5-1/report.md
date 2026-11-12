# Measurement 39w5 — task #39wtqdj

The measurement supports the hypothesis: accepted assignment `KindBinaryExpression` and `KindVariableDeclaration` nodes have nil `FlowNodeData`, and `predicateFalseRead` returns `yes` for those probes. The implementation's nil-data branch treats such accepted nodes as reads. This is a logging-only agreement check, not an assertion-based correctness test.

1. Loaded `/workspace/adamic-tools/env.sh`; fetched `origin compiler/predicate-false-read-4`, checked out `63eb56cd701c744a603f3c694295d749ba9d46dc` detached, and ran `git submodule update --init --recursive`. Exit code: 0. HEAD is one commit after `730e7c26e590af0c57c545ba0cb905b09e4bee54`.
2. Wrote the supplied test unchanged to `/tmp/measure39w5/measure39w5_test.go`, never inside the repository. SHA-256: `e16a447453ee87b3d2297fd047b44b07241b69972745842e5fc49e1252f52cb9`. Wrote `/tmp/measure39w5/overlay.json` as:

```json
{"Replace": {"/workspace/adamic/internal/lower/measure39w5_test.go": "/tmp/measure39w5/measure39w5_test.go"}}
```

3. Ran:

```sh
go test -overlay /tmp/measure39w5/overlay.json ./internal/lower -run '^TestMeasure39w5$' -count=1 -v -timeout 20m
```

The initial invocation exited 1 before compilation:

```text
go: creating work dir: stat /tmp/adamic-gate: no such file or directory
```

Created `/tmp/adamic-gate` outside the repository and reran the same command. Exit code: **0**. No compiler errors and no changes to the throwaway test. Complete successful output, including every `t.Logf` line verbatim:

```text
=== RUN   TestMeasure39w5
    measure39w5_test.go:76: no_read call 32:7 argument KindPropertyAccessExpression 32:12
    measure39w5_test.go:91:   accepted KindBinaryExpression at 27:3 parent KindExpressionStatement declarationName no FlowNodeData-nil yes FlowNode-nil n/a reaches -
    measure39w5_test.go:91:   accepted KindPropertyAccessExpression at 32:12 parent KindCallExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:91:   accepted KindPropertyAccessExpression at 32:41 parent KindPropertyAccessExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:97:   predicateUseDirections false no, predicateFalseRead yes, checkedPredicateDirections 3
    measure39w5_test.go:76: shorthand call 23:7 argument KindIdentifier 23:12
    measure39w5_test.go:91:   accepted KindIdentifier at 23:12 parent KindCallExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:91:   accepted KindIdentifier at 24:20 parent KindShorthandPropertyAssignment declarationName yes FlowNodeData-nil no FlowNode-nil no reaches yes
    measure39w5_test.go:91:   accepted KindIdentifier at 25:24 parent KindPropertyAccessExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches yes
    measure39w5_test.go:97:   predicateUseDirections false yes, predicateFalseRead yes, checkedPredicateDirections 3
    measure39w5_test.go:109: shorthand shorthand name items at 24:20: FlowNodeData-nil no FlowNode-nil no; the assignment's own FlowNodeData-nil yes
    measure39w5_test.go:76: logical_assignment call 23:7 argument KindIdentifier 23:12
    measure39w5_test.go:91:   accepted KindIdentifier at 23:12 parent KindCallExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:91:   accepted KindBinaryExpression at 24:3 parent KindExpressionStatement declarationName no FlowNodeData-nil yes FlowNode-nil n/a reaches -
    measure39w5_test.go:91:   accepted KindIdentifier at 24:3 parent KindBinaryExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches yes
    measure39w5_test.go:91:   accepted KindIdentifier at 25:17 parent KindTypeOfExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches yes
    measure39w5_test.go:97:   predicateUseDirections false no, predicateFalseRead yes, checkedPredicateDirections 3
    measure39w5_test.go:76: replaced_before_the_read call 4:62 argument KindIdentifier 4:67
    measure39w5_test.go:91:   accepted KindIdentifier at 4:67 parent KindCallExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:91:   accepted KindBinaryExpression at 4:74 parent KindExpressionStatement declarationName no FlowNodeData-nil yes FlowNode-nil n/a reaches -
    measure39w5_test.go:91:   accepted KindIdentifier at 4:105 parent KindPropertyAccessExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:97:   predicateUseDirections false no, predicateFalseRead yes, checkedPredicateDirections 3
    measure39w5_test.go:76: overload_some_empty call 7:20 argument KindIdentifier 7:33
    measure39w5_test.go:91:   accepted KindVariableDeclaration at 6:7 parent KindVariableDeclarationList declarationName no FlowNodeData-nil yes FlowNode-nil n/a reaches -
    measure39w5_test.go:91:   accepted KindIdentifier at 7:33 parent KindCallExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:97:   predicateUseDirections false no, predicateFalseRead yes, checkedPredicateDirections 3
    measure39w5_test.go:76: overload_every call 16:5 argument KindIdentifier 16:11
    measure39w5_test.go:91:   accepted KindVariableDeclaration at 15:7 parent KindVariableDeclarationList declarationName no FlowNodeData-nil yes FlowNode-nil n/a reaches -
    measure39w5_test.go:91:   accepted KindIdentifier at 16:11 parent KindCallExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:91:   accepted KindIdentifier at 16:45 parent KindElementAccessExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:97:   predicateUseDirections false no, predicateFalseRead yes, checkedPredicateDirections 1
    measure39w5_test.go:76: overload_every call 18:16 argument KindIdentifier 18:22
    measure39w5_test.go:91:   accepted KindVariableDeclaration at 17:7 parent KindVariableDeclarationList declarationName no FlowNodeData-nil yes FlowNode-nil n/a reaches -
    measure39w5_test.go:91:   accepted KindIdentifier at 18:22 parent KindCallExpression declarationName no FlowNodeData-nil no FlowNode-nil no reaches no
    measure39w5_test.go:97:   predicateUseDirections false no, predicateFalseRead yes, checkedPredicateDirections 3
--- PASS: TestMeasure39w5 (0.24s)
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.250s
```

4. Before writing this report, `git status --porcelain` produced no output: the repository was untouched by the measurement.

Report publication: only `proof/lower-39w-5-1/report.md` is committed on the new branch `proof/lower-39w-5-1-report`, based on the measured HEAD, and that branch is pushed to origin. No other branch is pushed.

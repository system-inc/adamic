# structure/react-component-no-separate-named-export

The port uses the existing structure file-context, React component-name,
hook-call, namespaced-member, AST projection, and bounded JSX-search helpers.
The declaration inventory and export-moving repair are rule-local upstream logic.
There are no options, suggestions, new syntax, private shared-helper copies,
or changes outside this rule directory.

The descriptor's `Test` prefix captures both the rule-specific tests and the
structure package's absent-optional-node guard. The harness filters captured
rows by the exact public rule name. `evidence/upstream-cases.json` records 57
unique file/source/options combinations; the selected upstream comparison
passes across Go, Node, emitted JavaScript and sanitized native.

Five witnesses cover hook detection, a mixed export list with no repair,
multiple declarations with the export before them and an async declaration,
type-only declarations/specifiers, and an aliased specifier.

Go behavior retained: `testdata/alias.tsx.txt` exports `helper as Button`.
Go matches the exported name `Button` against declared component names and
reports even though the local exported binding is `helper`. It offers no fix.
The port does the same. Type-only exports also report without a fix.

The repair is one edit spanning every insertion and the export-list removal.
It preserves documentation, trivia, async modifiers, and Go's declines for
mixed lists, aliases, type-only exports, overloads, duplicate declarations,
already exported/default declarations and multiple variable declarators.

## Validation

Toolchain setup and timing lines: `evidence/setup.log` and
`evidence/wasi-setup.log`. Required inputs: `evidence/inputs.log`.
The TypeScript source checkout is clean at v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`.
Both profile variables name the same fresh directory.

Commands from the repository root, with GOPROXY set to
`https://proxy.golang.org|direct` and `/workspace/adamic-tools/env.sh` sourced:

```
bash cloud/setup.sh
bash cloud/setup.sh --wasi-sdk
go run ./cmd/lint-registry
go test ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants/separate-named-export-ignored' -count=1 -v -timeout=30m
go test ./stage1/cohere/lint -run 'TestOwnedWitnesses|TestMutants/separate-named-export-ignored' -count=1 -v -timeout=30m
go test ./stage1/cohere/lint -count=1 -json -timeout=90m
```

The final selected run includes all five witnesses. The mutant changes the
component-name membership check to false. It compiles and is caught on Node
and emitted JavaScript. See `evidence/selected.log` for its exact log lines.
The original port's sanitized-native output is included in the witness and
upstream comparisons.

Full-package results, wall time, processor count and load are recorded in
`evidence/counts.json` and `evidence/whole.log`. The package was run once with
all inputs enabled; Python measured elapsed wall time because this environment
has no `/usr/bin/time` executable.

Full result: 152 passing checks (including subtests), 0 failures, 1 skip,
1168.547 seconds of wall time, nproc 5. Load before: 0.13 / 1.60 / 2.02;
after: 4.95 / 5.51 / 4.34.

The skip is `TestCheckerBridgeRefusalPending`, at
`stage1/cohere/lint/checker_pending_test.go:51`. It requires the pending
`TSGoError` bridge in `internal/load/prelude.d.ts`. Every requested input was
configured; this hard-coded prerequisite skip is outside this rule's owned
directory and cannot be removed within this unit. There is no helper or
language blocker for the rule itself.

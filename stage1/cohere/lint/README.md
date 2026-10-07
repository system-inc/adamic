# Syntax-only lint slice

Fifteen rules from the pinned cohere submodule, using the existing Adamic
TypeScript parser and scanner:

| Rule                | Visitor shape                                                          | Repair                            |
| ------------------- | ---------------------------------------------------------------------- | --------------------------------- |
| `no-debugger`       | statement with parent-kind test                                        | remove only from a statement list |
| `no-empty`          | empty block, function parent, interior trivia, empty switch            | none                              |
| `eqeqeq`            | binary operator, unwrapped operands, mode and null policy              | safe fix or human suggestion      |
| `no-var`            | declaration flags, ambient modifier, module ancestry, loop initializer | none                              |
| `no-duplicate-case` | switch-local recursive syntax/token signatures                         | none                              |

Batch 6 adds inventory candidates 1 through 10, each in its own file with one
registry call: `@next/next/no-assign-module-variable`,
`@typescript-eslint/default-param-last`, `@typescript-eslint/no-confusing-non-null-assertion`,
`@typescript-eslint/no-duplicate-enum-values`, `@typescript-eslint/no-dynamic-delete`,
`@typescript-eslint/no-extra-non-null-assertion`, `@typescript-eslint/no-misused-new`,
`@typescript-eslint/no-unnecessary-parameter-property-assignment`,
`@typescript-eslint/prefer-as-const`, and `default-case-last`.
See [BATCH6.md](BATCH6.md) for the selection pin and original evidence, and
[BATCH6_PERF.md](BATCH6_PERF.md) for Linux profiles and node-kind dispatch.

The runner indexes parents before dispatching the fifteen listeners in one preorder traversal. Child
indexes and a parallel parent-index array retain Go's ancestry queries without
owning parent/child cycles. Findings carry rule name, message ID and description,
trimmed range, and a repair category. Suggestions are never applied. There is
no type checker, binder, inferred type, or symbol lookup.

## Driver and answer protocol

Build `main.ts` with stage 0, or run the same source through `oracle/node.mjs`.
The driver accepts a source path, or `--manifest <path> [--count]`.
A manifest row is tab-separated:

```
source-path    rule-or-all    mode    null-policy    allow-empty-catch
```

The last four fields are optional. Defaults are all fifteen rules, `Always`,
`Always`, and false. Modes and null policies are cohere's decoded option values,
not ESLint configuration syntax. `Smart` forces the null policy to `Ignore`.
The driver is a corpus runner, not cohere's config, suppression, file-selection,
or CLI integration layer.

Each case starts with `case N`. Findings are stably sorted by start, then print
cohere's human finding lines (filename, one-based line and byte column, rule,
exact description). The oracle calls `report.Write` and excludes its timing and
coverage footer. A following record checks the byte start and end, message ID,
repair category, replacement and suggestion description. Finding spans and edit spans are independent. Every suggestion ID, description,
ordered edit payload and range is compared; additional automatic edits are
compared separately.

```
range START END ID REPAIR<TAB>REPLACEMENT<TAB>SUGGESTION<TAB>EDIT_START EDIT_END
suggestion ID<TAB>DESCRIPTION
suggestion-edit START END<TAB>REPLACEMENT
edit START END<TAB>REPLACEMENT
rejected RULE START END WINNER REASON
fixed<TAB>WHOLE-FIXED-SOURCE
```

String fields use printable ASCII with non-ASCII, backslash and control UTF-16
units escaped as `\uNNNN`. The final source is compared in full, including
unchanged files, comments, indentation and missing final newlines. Internally
Adamic positions are UTF-16; the driver maps them to Go's UTF-8 byte offsets.
Human line/column formatting follows `report.Write`, including its LF-only
line count. Count mode prints the finding count and skips formatting and fixing.

The native repair phase flattens automatic edits, resolves overlaps in cohere's
position/end/rule order, applies the surviving edits back to front, and reparses
each pass. It repeats to convergence with a ten-pass limit and prints rejected
proposals. Suggestions are never applied. The Go oracle uses cohere's actual
`edit.FixText`; both final source and rejections must agree.

## Tests

`TestRulesAgree` replays captured cases from all fifteen rules, including the
non-null absent-node regression fixtures, and 34 generated runs. A Go overlay records every test
harness `Run`, including tests that inspect repair fields without the ordinary
Expect helpers. It changes no rule. The original assertions still run; a failed
upstream test stops capture. Both positive and clean cases are retained.

`TestCompilerAndStage1Agree` compares every `.ts` file recursively under the
pinned TypeScript v6.0.3 `src/compiler` and this repository's `stage1`. The corpus
checkout stays in scratch. Both tests compare Go, the same TS on Node, and
native under ASan/UBSan with leak checking. The Go oracle rejects input parse diagnostics. One exact constructor parameter
property fixture needs parser recovery the stage1 parser refuses; the test
requires the documented refusal on Node and native instead of silently skipping it.

`TestMutants` changes only a scratch copy of the port. Each mutant must compile,
run successfully on Node and sanitized native, and differ from the Go answer:

- Suggestion promoted to an automatic fix.
- Empty function body exemption removed.
- Duplicate-case membership condition inverted.

`TestBatch6Mutants` checks ten rule-family mutants and two multi-suggestion
mutants, all compiling and running on Node and sanitized native. A separate
count-only mutant must pass ordinary byte parity before its count differs.

`TestNestedConstructorGap` holds the new compiler gap to Node and stage 0.
`TestThroughput` runs compiler count mode in five interleaved rounds when
`ADAMIC_LINT_BENCH=1`. Set `ADAMIC_TYPESCRIPT_SOURCE` to the pinned checkout.
See [REPORT.md](REPORT.md) for commands, results and measurement limits, and
[GAPS.md](GAPS.md) for the proving program and inherited representation gaps.

# Syntax-only lint slice

Fifteen rules from the pinned cohere submodule, using the existing Adamic
TypeScript parser and scanner:

| Rule | Visitor shape | Repair |
| --- | --- | --- |
| `no-debugger` | statement with parent-kind test | remove only from a statement list |
| `no-empty` | empty block, function parent, interior trivia, empty switch | none |
| `eqeqeq` | binary operator, unwrapped operands, mode and null policy | safe fix or human suggestion |
| `no-var` | declaration flags, ambient modifier, module ancestry, loop initializer | none |
| `no-duplicate-case` | switch-local recursive syntax/token signatures | none |

Batch 5 adds ten separately registered rule files; the names, claim commit,
branch checks and final evidence are in [BATCH5.md](BATCH5.md).

The runner dispatches these fifteen listeners in one preorder traversal. Child
indexes and a parallel parent-index array retain Go's ancestry queries without
owning parent/child cycles. Findings carry rule name, message ID and description,
trimmed range, and a repair category. Suggestions are never applied. There is
no type checker, binder, inferred type, or symbol lookup.

## Driver and answer protocol

Build `main.ts` with stage 0, or run the same source through `oracle/node.mjs`.
The driver accepts a source path, or `--manifest <path> [--count]`.
A manifest row is tab-separated:

```
source-path    rule-or-all    mode    null-policy    allow-empty-catch    decoded-options-json    recovery-mode
```

The last six fields are optional. Defaults are all fifteen rules, `Always`,
`Always`, and false. Modes and null policies are cohere's decoded option values,
not ESLint configuration syntax. `Smart` forces the null policy to `Ignore`.
The driver is a corpus runner, not cohere's config, suppression, file-selection,
or CLI integration layer.

Each case starts with `case N`. Findings are stably sorted by start, then print
cohere's human finding lines (filename, one-based line and byte column, rule,
exact description). The oracle calls `report.Write` and excludes its timing and
coverage footer. A following record checks the byte start and end, message ID,
repair category, replacement and suggestion description. Every automatic edit and every suggestion, including its message ID, description
and complete edit list, is compared. Edit ranges may differ from finding ranges.

```
range START END ID REPAIR<TAB>REPLACEMENT<TAB>SUGGESTION<TAB>EDIT_START EDIT_END
extra-fix START END<TAB>TEXT
suggestion ID<TAB>DESCRIPTION
suggestion-edit START END<TAB>TEXT
fixed<TAB>WHOLE-FIXED-SOURCE
```

String fields use printable ASCII with non-ASCII, backslash and control UTF-16
units escaped as `\uNNNN`. The final source is compared in full, including
unchanged files, comments, indentation and missing final newlines. Internally
Adamic positions are UTF-16; the driver maps them to Go's UTF-8 byte offsets.
Human line/column formatting follows `report.Write`, including its LF-only
line count. Count mode prints the finding count and skips formatting and fixing.

The native repair phase flattens automatic edits, sorts by start, end and rule,
resolves overlaps in Go's order, reparses and runs again for up to ten passes.
Budget exhaustion returns the original source and reports the unconverged rules.
Suggestions are never applied. The Go oracle uses cohere's actual `edit.FixText`;
its rejection records and final source must agree as well.

## Tests

`TestRulesAgree` captures 815 distinct upstream source/rule/decoded-options/file
combinations, retaining original upstream assertions. Two illegal method-body
cases are tested as explicit parser recovery gaps. The remaining 813 are compared
with 24 baseline generated runs and 131 batch 5 controls, including independent
Unicode, trivia, escape, ancestry and composed-repair cases.

`TestCompilerAndStage1Agree` compares every `.ts` file recursively under the
pinned TypeScript v6.0.3 `src/compiler` and this repository's `stage1`. The corpus
checkout stays in scratch. Both tests compare Go, the same TS on Node, and
native under ASan/UBSan with leak checking. The Go oracle rejects ordinary input parse diagnostics. Two invalid-source
proving programs are exercised separately by `TestBatch5RecoveryGaps`.

`TestMutants` changes only a scratch copy of the port. Each mutant must compile,
run successfully on Node and sanitized native, and differ from the Go answer:

- Suggestion promoted to an automatic fix.
- Empty function body exemption removed.
- Duplicate-case membership condition inverted.

`TestBatch5Mutants` suppresses each new rule in its own scratch file and requires
both Node and sanitized native to execute successfully and differ from Go.
`TestBatch5RepairPayloadMutants` loses an extra fix, loses the second suggestion,
and unsafely promotes suggestions to automatic edits.

`TestNestedConstructorGap` holds the new compiler gap to Node and stage 0.
`TestThroughput` runs compiler count mode in five interleaved rounds when
`ADAMIC_LINT_BENCH=1`. Set `ADAMIC_TYPESCRIPT_SOURCE` to the pinned checkout.
See [REPORT.md](REPORT.md) for commands, results and measurement limits, and
[GAPS.md](GAPS.md) for the proving program and inherited representation gaps.

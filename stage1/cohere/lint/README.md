# Syntax-only lint slice

Five rules from the pinned cohere submodule, using the existing Adamic
TypeScript parser and scanner:

| Rule | Visitor shape | Repair |
| --- | --- | --- |
| `no-debugger` | statement with parent-kind test | remove only from a statement list |
| `no-empty` | empty block, function parent, interior trivia, empty switch | none |
| `eqeqeq` | binary operator, unwrapped operands, mode and null policy | safe fix or human suggestion |
| `no-var` | declaration flags, ambient modifier, module ancestry, loop initializer | none |
| `no-duplicate-case` | switch-local recursive syntax/token signatures | none |

The runner discovers descriptors in `rules/*/rule.json` and generates static
node-kind dispatch for one preorder traversal. Each rule owns its factory,
implementation, messages, Go oracle adapter, witnesses and mutant. Child
indexes and a parallel parent-index array retain Go's ancestry queries without
owning parent/child cycles. Findings carry rule name, message ID and description,
trimmed range, and a repair category. Suggestions are never applied. There is
no type checker, binder, inferred type, or symbol lookup.

## Driver and answer protocol

Regenerate imports before building `main.ts` with stage 0 or running it on Node:

```sh
go run ./cmd/lint-registry
go run ./cmd/adamic build stage1/cohere/lint/main.ts
```

Tests perform regeneration automatically. Generated output is ignored. See
[the registration contract](../../../docs/lint-registration.md) to add a rule
without changing any shared file.
The driver accepts a source path, or `--manifest <path> [--count]`.
A manifest row is tab-separated:

```
source-path    rule-or-all    mode    null-policy    allow-empty-catch
```

The last four fields are optional. Defaults are all five rules, `Always`,
`Always`, and false. Modes and null policies are cohere's decoded option values,
not ESLint configuration syntax. `Smart` forces the null policy to `Ignore`.
The driver is a corpus runner, not cohere's config, suppression, file-selection,
or CLI integration layer.

Each case starts with `case N`. Findings are stably sorted by start, then print
cohere's human finding lines (filename, one-based line and byte column, rule,
exact description). The oracle calls `report.Write` and excludes its timing and
coverage footer. A following record checks the byte start and end, message ID,
repair category, replacement and suggestion description. Every selected repair
has exactly one edit whose range equals its finding; the Go oracle asserts this
shape before canonicalizing it.

```
range START END ID REPAIR<TAB>REPLACEMENT<TAB>SUGGESTION
fixed<TAB>WHOLE-FIXED-SOURCE
```

String fields use printable ASCII with non-ASCII, backslash and control UTF-16
units escaped as `\uNNNN`. The final source is compared in full, including
unchanged files, comments, indentation and missing final newlines. Internally
Adamic positions are UTF-16; the driver maps them to Go's UTF-8 byte offsets.
Human line/column formatting follows `report.Write`, including its LF-only
line count. Count mode prints the finding count and skips formatting and fixing.

The native repair phase applies disjoint edits back to front and reparses the
result. The Go oracle uses cohere's actual `edit.FixText`, including overlap
resolution, parse guards and repeated passes to convergence. Their resulting
sources must agree. This slice does not implement a general competing-fixer
engine: overlapping repairs stop explicitly.

## Tests

`TestRulesAgree` replays 217 distinct source/rule/decoded-options combinations
from cohere's own tests and 24 generated runs. A Go overlay records every test
harness `Run`, including tests that inspect repair fields without the ordinary
Expect helpers. It changes no rule. The original assertions still run; a failed
upstream test stops capture. Both positive and clean cases are retained.

`TestCompilerAndStage1Agree` compares every `.ts` file recursively under the
pinned TypeScript v6.0.3 `src/compiler` and this repository's `stage1`. The corpus
checkout stays in scratch. Both tests compare Go, the same TS on Node, and
native under ASan/UBSan with leak checking. The Go oracle also rejects input
parse diagnostics; parser recovery is outside this slice.

`TestMutants` discovers each directory's `mutant.json` and changes only a scratch copy of the port. Each mutant must compile,
run successfully on Node and sanitized native, and differ from the Go answer:

- Debugger removal suppressed.
- Suggestion promoted to an automatic fix.
- Empty function body exemption removed.
- Duplicate-case membership condition inverted.
- Variable declaration selection inverted.

`TestOwnedWitnesses` discovers raw TypeScript witnesses for every rule.
`TestRegistrationMutant` changes a valid node subscription and proves both
backends disagree with upstream Go. `TestFactoryHooks` verifies stateful factory
instances and prepare/visit/finish ordering, then kills a removed finish hook.
The generator tests reject malformed descriptors and duplicate names and verify
deterministic regeneration without rewriting unchanged files.

`TestNestedConstructorGap` holds the new compiler gap to Node and stage 0.
`TestThroughput` runs compiler count mode in five interleaved rounds when
`ADAMIC_LINT_BENCH=1`. Set `ADAMIC_TYPESCRIPT_SOURCE` to the pinned checkout.
See [REPORT.md](REPORT.md) for commands, results and measurement limits, and
[GAPS.md](GAPS.md) for the proving program and inherited representation gaps.

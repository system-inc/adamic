# Syntax-only lint slice

Twenty baseline rules from the pinned cohere submodule, plus the frequency-ranked
continuation in [VOLUME.md](VOLUME.md), using the existing Adamic TypeScript
parser and scanner:

| Rule | Visitor shape | Repair |
| --- | --- | --- |
| `no-debugger` | statement with parent-kind test | remove only from a statement list |
| `no-empty` | empty block, function parent, interior trivia, empty switch | none |
| `eqeqeq` | binary operator, unwrapped operands, mode and null policy | safe fix or human suggestion |
| `no-var` | declaration flags, ambient modifier, module ancestry, loop initializer | none |
| `no-duplicate-case` | switch-local recursive syntax/token signatures | none |
| `no-continue`, `no-with`, `no-new` | control statement / keyword span / expression parent | none |
| `no-sparse-arrays` | omitted element with assignment-target ancestry | none |
| `require-yield` | generator ownership across nested functions and classes | none |
| `no-await-in-loop` | repeated loop fields with function and for-await boundaries | none |
| `vars-on-top` | function, static and program declaration prefix | none |
| `no-template-curly-in-string` | cooked string placeholder scan | none |
| `no-labels` | label resolution with loop/switch options | none |
| `no-bitwise` | 13 operators, allow list and directional int32 hint | none |
| `no-sequences` | comma chains, for fields, parenthesis option | none |
| `unicode-bom` | file prefix with absent-option default | insert/delete one mark per pass |
| `no-div-regex` | raw regex prefix | replace one character inside the finding |
| `no-warning-comments` | comments, terms/location/decoration and self-directives | none |
| `no-unneeded-ternary` | boolean branches, precedence, source-preserving inversion/default | rewrite whole expression |
| `base/consistency-no-console` | unwrapped member receiver | none |
| `nexus/consistency-require-type-suffix` | declarations and const-enum-shaped objects | none |
| `adamic/no-type-predicate` | type predicate | none |
| `no-plusplus` | update operators and optional for-incrementor exemption | none |
| `@typescript-eslint/method-signature-style` | type-member signatures, overload groups, module ancestry | rewrite member, merge overloads, readonly suggestion; valid-source coverage |
| `@typescript-eslint/no-wrapper-object-types` | type references with whole-file shadow guard | lowercase primitive except implements |
| `@typescript-eslint/prefer-literal-enum-member` | enum initializer recursion with bitwise option | none |
| `nexus/consistency-no-enum` | enum declaration name | none |
| `no-negated-condition` | if and ternary conditions, else-if exemption | none |
| `no-return-assign` | assignments, return/arrow ancestry, positional parentheses | none |


The runner dispatches these listeners in one preorder traversal. Child
indexes and a parallel parent-index array retain Go's ancestry queries without
owning parent/child cycles. Findings carry rule name, message ID and description,
trimmed range, and a repair category. Suggestions are never applied. There is
no type checker, binder, inferred type, or symbol lookup.

## Driver and answer protocol

Build `main.ts` with stage 0, or run the same source through `oracle/node.mjs`.
The driver accepts a source path, or `--manifest <path> [--count]`.
A manifest row is tab-separated:

```
source-path    rule-or-all    mode    null-policy    allow-empty-catch    decoded-options-json    recovery
```

All fields after the path are optional. Defaults are all thirty implemented rules (method-signature-style has the
explicit recovery limit in VOLUME.md), `Always`,
`Always`, and false. Modes and null policies are cohere's decoded option values,
not ESLint configuration syntax. `Smart` forces the null policy to `Ignore`.
The driver is a corpus runner, not cohere's config, suppression, file-selection,
or CLI integration layer.

Each case starts with `case N`. Findings are stably sorted by start, then print
cohere's human finding lines (filename, one-based line and byte column, rule,
exact description). The oracle calls `report.Write` and excludes its timing and
coverage footer. A following record checks the byte start and end, message ID,
repair category, replacement and suggestion description. Every selected repair has exactly one edit; the Go oracle asserts this shape.
Finding and edit ranges are checked separately. Decoded option JSON preserves
nil versus false and nil versus an empty list. Settings reads this restricted
data grammar through the existing parser, without evaluating it.

```
range START END ID REPAIR<TAB>REPLACEMENT<TAB>SUGGESTION<TAB>EDIT-START EDIT-END
rejected RULE START END WINNER REASON
fixed<TAB>WHOLE-FIXED-SOURCE
```

String fields use printable ASCII with non-ASCII, backslash and control UTF-16
units escaped as `\uNNNN`. The final source is compared in full, including
unchanged files, comments, indentation and missing final newlines. Internally
Adamic positions are UTF-16; the driver maps them to Go's UTF-8 byte offsets.
Human line/column formatting follows `report.Write`, including its LF-only
line count. Count mode prints the finding count and skips formatting and fixing.

The native repair phase sorts proposals by start, end and rule, reports
overlap refusals, applies surviving edits back to front, reparses and reruns
until convergence (at most ten passes). The Go oracle uses cohere's actual `edit.FixText`, including overlap
resolution, parse guards and repeated passes to convergence. Their resulting
sources must agree. Invalid ranges, nonprogress and exhausted passes stop explicitly. No general
configuration, suppression, filesystem write or formatter layer is ported.

## Tests

`TestRulesAgree` replays 873 distinct source/rule/decoded-options combinations
from cohere's own tests and 140 generated runs. A Go overlay records every test
harness `Run`, including tests that inspect repair fields without the ordinary
Expect helpers. It changes no rule. The original assertions still run; a failed
upstream test stops capture. Both positive and clean cases are retained.

`TestCompilerAndStage1Agree` compares every `.ts` file recursively under the
pinned TypeScript v6.0.3 `src/compiler` and this repository's `stage1`. The corpus
checkout stays in scratch. Both tests compare Go, the same TS on Node, and
native under ASan/UBSan with leak checking. The Go oracle rejects input parse diagnostics except for five explicit
`no-div-regex` recovery fixtures. Their findings and proposed edits are held
byte for byte, with a `recovery findings only` marker in place of fixed output.
Cohere's fix engine refuses invalid input before collecting proposals; those
five cases therefore do not claim converged fixes or general parser recovery.

`TestMutants` changes only a scratch copy of the port. Each mutant must compile,
run successfully on Node and sanitized native, and differ from the Go answer:

- Suggestion promoted to an automatic fix.
- Empty function body exemption removed.
- Duplicate-case membership condition inverted.

`TestNestedConstructorGap` and `TestOptionAndComparatorGaps` hold all three
compiler gaps to Node and stage 0. Further rule-family mutants exercise
control, assignment targets, generator/await scope, options, regex edit ranges,
boolean inversion, comment directives and BOM edits.
`TestDecorationOptionMutant` proves hyphen-range options are checked.
`TestCountGuardMutant` changes only count mode: ordinary output stays identical,
but the throughput count check catches the wrong count on Node and native.
`TestThroughput` runs compiler count mode in five interleaved rounds when
`ADAMIC_LINT_BENCH=1`. Set `ADAMIC_TYPESCRIPT_SOURCE` to the pinned checkout.
See [REPORT.md](REPORT.md) for commands, results and measurement limits, and
[GAPS.md](GAPS.md) for the proving program and inherited representation gaps.

Performance follow-up: [PERFORMANCE.md](PERFORMANCE.md) records Callgrind cost
attribution and port changes. `TestProfileArtifacts` saves reproducible snapshots
when `ADAMIC_LINT_PROFILE_DIR` is set; `TestProfileSnapshotsAgree` checks release,
debug and Node snapshots when `ADAMIC_LINT_PROFILE_SNAPSHOTS` lists directories.
`profile.py` measures five interleaved count-mode rounds and checks deterministic
instruction accounting. `TestCommentFoldMutant` covers scalar case folding.

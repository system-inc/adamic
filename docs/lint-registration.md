# Lint registration without shared lines

Current main (`5d4c801`) has five syntax rules. This migration moves all five,
including their messages, options adapters, witnesses and mutants, into
`stage1/cohere/lint/rules/<slug>/`. The thirty-rule scanner branch described in
the accepted proposal is another worker's branch; this unit changes main's five.

Add only your own rule directory. Do not append imports, dispatch conditions,
Go oracle selections, corpus filters, file-copy lists or mutant lists elsewhere.
The rule directory contains:

- `rule.ts`: an exported concrete listener class and named factory.
- `messages.ts`: exact descriptions and any message builders.
- `rule.json`: public name, interested node kinds, factory/class names and provenance.
- `oracle.go`: the unmodified upstream cohere rule and its typed options adapter.
- `testdata/*.ts.txt`: raw TypeScript witnesses, outside the module graph.
- `mutant.json`: one source mutation that compiles, runs, and must disagree with Go.
  Fields are `name`, `from`, `to`, and optional `file` (default `rule.ts`). The
  file must stay inside this rule directory. Anchors are scoped to that file,
  so another worker can use the same snippet without changing your test.

For example, the debugger descriptor is:

```json
{
  "name": "no-debugger",
  "order": 1,
  "kinds": ["DebuggerStatement"],
  "visit": "visit",
  "parent": true,
  "factory": "create",
  "class": "Rule",
  "oracle": "oracleNoDebugger",
  "upstreamPackage": "core",
  "upstreamTest": "TestNoDebugger"
}
```

`order` preserves the five migrated rules' prior dispatch and stable finding tie
order. **New rules omit it** and sort lexically by public name after migrated
rules. There is no shared ordinal allocator. Slugs contain lowercase letters,
digits and hyphens; scoped public names such as `@typescript-eslint/example`
stay in `name`. Upstream test names are prefixes because the tests have suffixes.
Only captured cases with the descriptor's exact rule name enter its corpus.

`create(context: RuleContext)` returns the exported concrete class. The class
implements `visit(index: number, parent: number)` when `parent: true`, or
`visit(index: number)` otherwise. Optional `prepare` and `finish` descriptor
fields name methods taking the source root index. File listeners subscribe to
`SourceFile`. Factories create one instance per source file; instances can own
rule-local state. Each file's full numeric ancestry is ready before `prepare`.
All prepare hooks run before preorder visitation; all finish hooks run afterward.
Hooks and visits run only for `all` or the exact selected name.

The shared context owns source, initialized parser/scanner, read-only parent
indexes, the finding destination and main's existing equality/catch options.
It has no link back to Linter or RuleSet. Concrete named imports and typed fields
avoid interface-method casts and reference cycles. Keep new options decoding
inside the rule adapter; a genuinely new driver option format is infrastructure
work, rather than another per-rule switch in a shared file.

The adapter is built through a Go overlay inside cohere, with the generated
selection file and **all** discovered adapter files passed to `go build`. It
returns the upstream `rule.Rule` from a uniquely named function and has a second
function named `<oracle>Options(fields []string) any`. Preserve upstream defaults
and legacy manifest fields: equality mode/null are fields 2/3, empty-catch is field
4. Mark the adapter `//go:build lintoracle`: ordinary Adamic package discovery
must not build cohere's internal imports. The explicit-file overlay build includes
it. Cohere's rule bodies, messages, formatter and converging fixer stay independent
from the TS port and unchanged in the submodule.

From the repository root, before every manual build or Node run:

```sh
go run ./cmd/lint-registry
go run ./cmd/adamic build stage1/cohere/lint/main.ts
```

The generator validates the entire discovered set: strict JSON, unique public
names and adapter functions, complete directories, named factories/classes/hooks,
pinned AST kind names, required adapter exports, witnesses and mutant metadata.
It writes bucketed TS dispatch and Go selection into ignored `.generated/` files.
Generation uses sorted descriptors and node kinds, atomic replacement and unchanged
byte checks. A removed descriptor cannot leave a stale registration behind.
Do not commit these outputs. CI tests validate and regenerate before filtered tests
as well as before each port build and Node run.

Tests derive upstream package/test filters and copied module files by discovery.
Every owned witness must cause an upstream finding, and Go/Node/sanitized-native
outputs must match for selected and all-rule runs. Each discovered mutant is tested
against its own witnesses plus the inherited corner-case corpus. The existing
217 upstream cases, decoded defaults, exact descriptions, UTF-8 ranges, stable
findings, safe fixes, suggestions, fixed-source reparsing and overlap refusal are
preserved. This unit does not expand main's fix engine or parser recovery contract.

For a new rule, run the generator and package tests, review the upstream comparison
and mutant log, and commit only that directory. Two workers choosing distinct slugs
share no changed lines. Trial commits, merge output and tests are recorded in
[the evidence report](lint-registration-evidence/REPORT.md).

Oracle fixtures have their own companion migration on `codex/no-shared-lists`,
commit `7f958de`. Its source/options/counts sidecars and discovery command remove
fixture registry and table appends. This lint branch starts from main and does not
include that unit's changes; follow that unit's fixture instructions once merged.

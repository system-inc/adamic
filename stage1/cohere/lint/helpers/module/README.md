# Module export predicates

The retained `ecmascript/module` slice selected by triage d7ab0bc4 comes from
`origin/codex/lint-helpers-04` at `8e0d98a645ae73cc4cf72561c779f257b51b8c9a`,
originally `helpers/slot04_wave10/`. There is no retained duplicate of these symbols.
Each helper has its own file; the original helper bodies are unchanged.

- `HasExportModifier`, `has_export_modifier.a`: an export keyword in a modifier list.
- `IsExported`, `is_exported.a`: a present declaration carrying that modifier.
- `IsExportedByName`, `is_exported_by_name.a`: export without default.

`modifier_view.a` holds immutable projected facts. `exportNodeView(parser, index)`
in `node_view.a` adapts stage1 modifier children for these predicates. Its presence
flag records relevant export/default modifiers, not every unrelated modifier; all
three predicates give the same answer when the other modifiers are omitted.

```ts
IsExported(exportNodeView(context.parser, index))
IsExportedByName(exportNodeView(context.parser, index).modifiers)
```

The Go oracle reads real AST modifier lists independently. The baseline compares
all three predicates on every AST node in 426 freshly captured inputs from all
10 consuming rules, plus direct witnesses: 8,875 consumer rows and 2,319 witness
rows on source Node, emitted JavaScript, and ASan/UBSan native. Six semantic
mutants (two per helper) compile, run, and disagree with Go on all three runtimes.
Consumer-omission tests reject removal of any consuming rule's fixture coverage.
The oracle adapter uses the current pinned `PathKey` API; Go helper bodies are
unchanged. Regenerate from the repository root with
`python3 stage1/cohere/lint/helpers/module/testdata/capture.py`.

`go test ./stage1/cohere/lint/helpers` invokes this package through
`module_package_test.go`; its tests can also run directly.

The proof rule is `@typescript-eslint/no-useless-empty-export`, with 64 captured
upstream source/file/options cases, witnesses, automatic deletions, and a semantic
mutant, in `rules/typescript-no-useless-empty-export/`. The other independently
helper-ready rule in the triage ledger is
`nexus/consistency-no-screaming-snake-case`; readiness is not a complete port claim.

This slice does not supply the five other module symbols absent from the source
branch: AllDeclaredTypeNames, DeclaresTypeNamed, HasDefaultModifier,
IsDefaultExported, and forEachDeclaredTypeName. No dependency or language gap
blocks the three retained predicates or the proof rule.

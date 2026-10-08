# Merged wave branch build blocker

The requested merge of area c4bdc23fa86d55cf7e579989201c11258f4d3a62 into the existing pushed wave branch succeeded at 02ff654e2ab6e0b83709ad7c6af2f754e1ea6ad9. This preserves the old private bridge extensions already on that branch, without importing them into a new area-based rule port.

Reproducer from the merged wave branch:

```
source /workspace/adamic-tools/env.sh
GOFLAGS=-buildvcs=false GOMAXPROCS=4 go vet ./stage1/cohere/lint/...
```

First compiler error: `bridge/tsgo/checker/binding_origin.go:39:13`: `source.FileName()` now has type `tspath.RootedFilePath`, but `out.text` takes string. Further errors include `declaration_chain.go:38:56` and `declaration_facts.go:22:55`: `SourceFile.Path` was removed; and `module_records.go:65:59`: `GetSourceFileForResolvedModule` now requires `*module.ResolvedModule` rather than its rooted filename.

No bridge compatibility edits were made: the current instruction restricts new commits to owned unified rule directories. No private checker, shared helper copy, descriptor change or gate relaxation was added. This merged branch is not landing-ready. Its full lint invocation and vet diagnostics are archived beside this note after completion; a package build failure cannot be reported as passing rule tests or caught mutants.

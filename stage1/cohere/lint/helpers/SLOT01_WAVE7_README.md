# Slot 01 seventh helper batch

Three separate .a helpers complete the collector's utility and theme block surfaces and assemble a live descriptor table. Each serves the six canonical/order/shorthand/conflict/unknown Tailwind rules in slot01_wave7_readiness.json. Eighteen remaining prerequisite entries are removed, with zero additional complete rule readiness. This slot now supplies twenty-one retained helpers.

`collapseIngestUtilityBlock` returns an empty error string on success and Go's exact error text on rejection. Supply Go TrimSpace behavior explicitly. Roots store bit 1 for static and bit 2 for functional, retaining both. Functional definitions append in order; static definitions overwrite only the static body. Body slices carry values, length and capacity. The helper copies each header and shares the backing values, matching Go assignment. Callers must preserve valid external arena indices and slice bounds.

`collapseIngestThemeBlock` takes a prefix state, flat arena nodes, root indices and dependencies for options parsing, prefix validation, Go quoting, Walk, IsContainer, identifier unescaping and Theme.Add. Walk invokes the supplied visitor using continue=0, skip=1 and stop=2, in Go's document order. A keyframe subtree is skipped; comments and containers continue; custom properties delegate to Add. Errors stop the walk and retain any earlier changes. A nonempty valid prefix updates state before walking; an empty prefix preserves prior state. An invalid arena index explicitly refuses rather than returning clean output.

`collapseNewTable` takes a bound system with version/theme handle and generated base descriptor handles, framework static registrations, a reading factory, shared PropertyOrder, and the three repository-composition callbacks. A missing system produces a bound=false projection with empty data; callers must decline it. A present system gets fresh descriptor/static maps, preserving descriptor handles and reading backing storage. Reading length/capacity headers and counts are copied. PropertyOrder and theme identity are shared. Callbacks run theme, repository statics, repository functional roots, in that order.

Descriptor numbers identify complete externally owned descriptor objects; the helper neither truncates their fields nor reconstructs them. Theme handles similarly identify the caller's theme arena. Reading headers describe Go's slice view over externally supplied backing values. This unit does not implement append/reslicing operations or the dependency callbacks.

The slot-owned Go overlays invoke the actual private helpers on every consumer's complete statically extractable fixture string, plus parsed/manual CSS and alias/error controls. The theme overlay records actual visitor actions and Add arguments; the table overlay records actual composition calls and actual reading results. It preserves original branches and delegates to the original dependencies. No cohere worktree source, shared rule harness, registration generator or compiler ownership file is edited.

Run from repository root after sourcing /workspace/adamic-tools/env.sh:

```
go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave7' -count=1 -v -timeout=20m > /tmp/slot01-wave7.log 2>&1
```

Read SLOT01_WAVE7_REPORT.md for every mutant, command, observation and limit. The frozen readiness ledger is unchanged. New Adamic files use .a; the inherited OptionsJson .ts import remains existing infrastructure and is renamed only in temporary mutant copies.

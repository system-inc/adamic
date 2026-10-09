# testgrain wave 1

Base: `devtools/testgrain-only` at `0efe2255`, itself based on main `ad631129`.
The helper-only branch leaves both proof files exactly as main has them.

Run the mechanical edit from that base (the program rejects already-converted files):

```sh
go run ./review/devtools/testgrain-codemod \
  stage1/cohere/lint/node_table_split_test.go \
  stage1/cohere/markdownblocks/text_shards_test.go \
  stage1/cohere/markdownblocks/list_layout_shards_test.go \
  stage1/typescript/parser/compiler_expressions_shards_test.go \
  stage1/cohere/lint/grain_rules_agree_shards_test.go \
  stage1/cohere/markdownblocks/quote_layout_shards_test.go
```

Mechanical edits use `go/ast` and `go/printer`: wrap the selected preparation
functions in `testgrain.Setup`, replace shard timers with `Unit`, remove grain
clock/verdict statements, remove obsolete deadline functions, delegate the
existing text/list/quote command factories, and add the helper import. Recipes
are restricted to these six files. Existing build recipes are preserved.

Hand edits remove the redundant process-once wrappers and state, convert
identity lists to `Assign`, replace source parsing and function-list union checks
with `Union`, and convert planted-failure ownership checks to `CaughtByExactly`.
The parser's tuple-returning setup is backed by one immutable prepared value.
Text's sibling dispatcher still calls `textDeadline`; that compatibility
function now starts `Unit` and returns a no-op cleanup.

Lint's scoped execution helpers preserve file-backed stdout and child CPU guards
while routing every explicit command through testgrain. The parser has scoped
manifest, oracle and execution helpers for the same reason; other test families'
harnesses remain unchanged. Both lint capture recipes use `-timeout=0`, including
their cache flags. The only explicit context deadlines left in these files are
10-second planted-failure probes, which run after preparation.

## Case counts

Counts are sums of the executed shard logs, checked against the corresponding
union where available. They count corpus cases, not Go test functions.

| File/family | Shards | Before | After |
| --- | ---: | ---: | ---: |
| node_table_split | 8 | 3,914 | 3,914 |
| grain_rules_agree | 16 | 3,922 | 3,922 |
| quote_layout | 12 | 4,070 | 4,070 |
| list_layout | 16 | 4,073 | 4,073 |
| text | 64 | 1,115,666 | 1,115,666 |
| compiler_expressions | 16 | 77 | 77 |

Every listed family passes before and after, with the final runs warm. Parser
runs use TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.
FNV-64a preserves existing node-table, rules-agree and quote assignments. It
replaces the text/list SHA-256 and parser FNV-32a partitions; their totals and
corpora are unchanged, while individual shard sizes change as expected.

`gofmt`, `go vet` for all affected packages and the codemod, and `git diff --check`
are clean. `git grep -n 'AfterFunc\|cooked'` returns no matches in the six files.
Each test command has a 90-second hard cap. Whole-package results are recorded
below; focused family passes do not imply that all other package tests passed.

## API fit and limits

Overlapping `Setup` calls for the same `testing.TB` cannot identify separate
command owners: the helper registers a command with every active grain for that
TB. A finishing setup could kill a sibling preparation's child. Node-table and
rules-agree therefore prepare their stages sequentially, preserving preparation
results, caching and the deadline policy. Quote's native builds can stay parallel:
they are children of one preparation, without overlapping nested setup stages.

`native.Build` creates private compiler processes and exposes no command-factory
hook. The four files calling it (node-table, rules-agree, text, compiler-expressions)
retain that existing builder; its internal compiler commands cannot be registered
by the testgrain API without changing the native package. The explicit commands
in these files are all registered. Runtime-library compilation inside the native
library is likewise outside this codemod's scope. Loom must still supervise the
complete preparation unit; Setup intentionally has no test-side deadline.

## Whole-package verification (incomplete)

The complete lint, markdownblocks and parser binaries were each run with
`-test.timeout=90s -test.count=1`, under an external `timeout -s KILL 90`.
All three reached the external cap and were killed (exit -9). None produced a
whole-package verdict. These kills count as failed verification attempts;
this branch does not claim that the whole packages pass. The six focused warm
families above all completed successfully under the same cap. The monitored
parser process left no live descendant processes when it exited.

## Main merge repair

Merged main `163086eb`. The conflicted list-layout file was replaced in full
with main's version before applying the AST codemod again; no conflict hunks
were hand-merged. The conversion preserves main's setup-on-cache-miss behavior,
its readiness check and the absence of a setup-side deadline. All 16 list-layout
shards pass warm in 12.6 seconds, with 4,073 cases before and after. Package vet,
gofmt and the timer/verdict grep are clean.

The requested complete markdownblocks `go test -timeout 90s -v -count=1` run
was killed at its external 90-second limit during
`TestMarkdownStructureLayout_Setup`; recorded child processes were killed too.
This remains a failed whole-package verification attempt, not a package pass.
Main then advanced to `b3f83786` (scanner-only changes); list-layout's blob was
identical to `163086eb`, so the conversion was retained and that tip merged too.

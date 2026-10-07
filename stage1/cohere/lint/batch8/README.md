# Inventory syntax batch 8

Candidates 21 through 30 from the frozen inventory at `73ac2eb`. Each listener
owns one file and one dispatch line in `registry.ts`. This runner is separate
from main's original five-rule runner so the inventory batches can merge
without changing its findings or answer protocol.

Run `main.ts` through `oracle/node.mjs`, or build it with stage 0. It accepts a
source path or `--manifest <path> [--count]`. Ordinary manifest rows contain a
source path and an optional rule name, separated by a tab. Omit the rule to run
all ten listeners. There is no checker, binder, configuration decoder or
suppression layer.

Canonical output contains cohere's human finding lines, followed by:

```text
range BYTE_START BYTE_END MESSAGE_ID
fix BYTE_START BYTE_END<TAB>REPLACEMENT
suggestion MESSAGE_ID<TAB>DESCRIPTION
edit BYTE_START BYTE_END<TAB>REPLACEMENT
fixed<TAB>WHOLE_FIXED_SOURCE
```

Every automatic edit and every suggestion's ordered edits are retained.
Suggestions are never applied. String payloads escape control characters,
backslashes and non-ASCII UTF-16 units as `\uNNNN`. Positions are converted from
UTF-16 to Go's UTF-8 byte offsets. Line and offset tables are built once.
Findings are stably sorted by their start position.

Automatic edits are applied backwards and reparsed, repeating up to ten passes.
The Go answer uses cohere's converging edit engine when automatic proposals
exist. The native runner stops on overlapping edits, parser failure or failure
to converge. It does not implement cohere's general rejection protocol.
Fixtures with legacy escapes still compare the Go rule's findings and absence
of fixes; the rule harness accepts them despite lexical diagnostics. No edit
engine is invoked when there is no automatic proposal.

## JSX boundary

The stage1 parser cannot parse JSX. The `jsx_no_comment_textnodes.ts` listener
is implemented, but production parser integration is blocked. Its test adapter
accepts a third manifest field containing Go's `JsxText` boundaries as
`UTF16_START:UTF16_END,...`, or `empty` for a Go JSX tree without text nodes.
These ranges come from the independent Go parser, not a source-text heuristic.
This mode tests the listener only and is not a claim of JSX parsing support.

The Go oracle preserves each fixture's TS, TSX, JS or JSX parser mode. Fourteen
JSX-bearing fixtures for the other React listeners are explicitly recorded by
rule and source hash in `../batch8_evidence/parser_gaps.json`. No other fixture
is skipped. A valid JSX probe produces a Go finding and the same parser refusal
on Node and sanitized native. Removing the JSX gap needs a parser unit.

## Verification

Set `ADAMIC_TYPESCRIPT_SOURCE` to the TypeScript checkout pinned in
`../lint_test.go`. `TestBatch8` captures the selected upstream tests afresh,
checks supported cases and the compiler/stage1 corpus against Go, then runs a
compiled mutant for each family and two extra repair mutants. All positive
mutant controls must match first. Visitors and mutant builds use a fixed source
snapshot. ASan, UBSan and leak checks hold the native observations.

`TestBatch8PositionsAndRelease` checks Unicode positions, line separators,
private writes in `for await`, constructor defaults with nested generic types,
and the Go division rule's raw-byte boundary behavior. It also compares the
release binary on the complete compiler/stage1 corpus. `ADAMIC_LINT_BENCH=1`
enables five fresh-process, interleaved count rounds. See [BATCH8.md](../BATCH8.md)
for selection, commands, measurements, mutants and coverage limits.

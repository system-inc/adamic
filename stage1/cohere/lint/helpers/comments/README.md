# Parser-anchored comment helpers

This unit ports the next bounded high-yield helper dependency set from the 152-rule residual ledger: `comments.All`, `comments.ForFile`, `canBeginAt`, `collectListInteriors`, and `sortByPosition`. The initial claim was pushed as `29990b4` before implementation. One helper per TypeScript file; `comment.ts` only defines the immutable result value. No existing rule entry point is changed.

| File | Contract | Additional fully helper-ready rules in dependency order |
|---|---|---:|
| `can_begin_at.ts` | Go ASCII guard and shebang exception | 0 |
| `collect_list_interiors.ts` | Parser-owned empty and trailing-comma interiors, including empty JSX expressions | 0 |
| `sort_by_position.ts` | Stable source order | 0 |
| `all.ts` | Deduplicated comments, full text, block kind, UTF-8 ranges and byte columns | 0 until the cache is available |
| `for_file.ts` | One shared scan per immutable file, with an uncached path | 16 with the four prerequisites |

The five-symbol bundle serves 24 consumer rules. Removing these symbols from the previous ledger yields **16 additional helper-ready candidates, 62 total, 136 still helper-blocked**. These are dependency claims under the inventory's common AST adapter assumption, not completed rules or full independent parser support. The exact rule list and residual dependencies are in `readiness.json`. No first-unit helper ledger or other worker's rule files are edited.

## Call contract

`all(parser, root)` accepts the parsed file's flat node arena, its root index, and its immutable source text through `parser.scanner.text`. Internal positions are UTF-16 code units; returned `Comment.start`, `end`, and `startColumn` are UTF-8 bytes. Lines are zero-based and follow ECMAScript line breaks, including CRLF and U+2028/U+2029. `all(undefined, -1)` returns an empty result, corresponding to Go's nil slice. Do not pass these byte offsets directly to `source.slice` or the existing lint `Finding` constructor, which use UTF-16 positions. Convert through the file's byte-to-unit map in the common reporting adapter.

Create exactly one `FileCommentCache(parser, root)` per immutable file and share it across all of that file's rule contexts. Call `forFile(parser, root, cache)` from each consumer. Do not reuse a cache for another file or mutate the parsed arena/source after its first scan. `forFile(parser, root, undefined)` computes without caching, as a nil Go file cache does. The cache has no global map retaining completed files. Treat its returned comments as read-only, including when sorting a separate working copy.

The stage-1 arena discards Go's `NodeList.Pos/End`. `collectListInteriors(parser, index)` returns an owned list of candidate positions. It reconstructs anchors only inside the enumerated parser list owners, skips child spans, and skips existing comments while finding delimiters. It does not sweep arbitrary slashes. The proof compares resulting comment ranges, text and metadata against the actual Go helper, rather than requiring identical redundant anchor sets. The stable sort intentionally keeps Go's in-place slice contract; its two caller writes carry explicit, explained lint directives. A stage-1 parser refusal must remain a refusal; a caller must never catch it and lint a partial tree.

## Validation boundary

Two independent comparisons are kept distinct:

1. Forty-seven boundary sources are parsed independently by Go and the current Adamic TypeScript parser, then held on Go, Node, and ASan/UBSan native. They cover empty delimiters, trailing commas, shebangs, EOF comments, literal slashes, deduplication, Unicode byte positions and source separators.
2. All 4,513 captured input sources from the 24 Go consumer rules are compared with their actual Go AST geometry projected into the flat arena. The oracle supplies only node kinds, positions, ends and child indexes, not comment answers. The same helper TypeScript runs on Node and sanitized native. This validates the shared helper contract assumed by the inventory, including JSX and recovery trees; it does not port Go's parser into Adamic.

The independent-parser adapter now refuses all four valid JSX boundary sources with `NotYet: stage-1 JSX parser adapter`. Before that guard, the underlying parser refused three and silently accepted `<div>{/* c */}</div>` as a type assertion/object literal followed by a less-than/regex expression. `TestJsxParserGapIsExplicit` and its compiling adapter-guard mutant hold the refusal boundary; the wrong tree is preserved in the evidence. This guard is in the owned adapter driver, not a change to another worker's parser. In particular, the new `react/jsx-curly-brace-presence` dependency claim remains conditional on a JSX-capable AST adapter. Rule-local behavior, defaults, fixes, full fixture parity, checker questions and integration are each rule worker's responsibility.

The Go oracle uses overlays to expose the two private functions and build its driver without changing the pinned cohere checkout. `testdata/capture.py` executes the original consumer tests with capture enabled, preserves Unicode line separators, deduplicates by rule/file/source, and asserts all 24 consumers were observed. Both cached and uncached comment content are compared; sharing identity is compared separately. All five helper files have compiling semantic mutants, whose output differences are recorded in `evidence/`. A sixth compiling mutant removes the TSX adapter refusal and finishes the known wrong parse with exit 0, which the refusal check catches.

## Reproduce

Source `/workspace/adamic-tools/env.sh` first. From the repository root:

```sh
python3 stage1/cohere/lint/helpers/comments/testdata/capture.py > /tmp/comments-capture.log 2>&1
go test -count=1 -timeout 15m -v ./stage1/cohere/lint/helpers/comments > /tmp/comments-tests.log 2>&1
```

Read the logs after completion. Generation needs Go cohere's original rule tests; normal helper tests use the committed consumer corpus. No throughput improvement is claimed from these comparisons.

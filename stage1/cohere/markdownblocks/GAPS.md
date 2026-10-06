# Markdown block/parser preflight: stopped before a full port

The requested complete native Markdown formatter was not built. This branch
contains a reproducible whole-document oracle audit and parser-representation
probes, alongside the preceding inline slice as an unchanged dependency.
There is no Adamic Markdown file formatter or native full-document throughput
claim here. The audit establishes the target before importing its implementation.

## The default full-document oracles disagree

Go cohere's real `native.Formatter` invokes the Markdown parser/printer and its
native embedded-language dispatch. The JavaScript library it follows is its
vendored Prettier fork, based on version 3.9.6, not unmodified npm Prettier.
`testdata/library.mjs` calls those exact standalone/plugin bundles on Node.
The audit also measured stock npm Prettier 3.9.6 separately.

On the initial census of 873 physical repository/submodule Markdown files and
75 generated block cases, Go and the pinned fork matched 941/948 documents with
embedding enabled. Seven real files disagreed: one JSONC fence and six Flow
fixtures. Stock npm Prettier matched 788/948. With embedding disabled, both
JavaScript libraries matched Go on all 948. The test repeats the census and
checks a named list of observed default-mode gaps; closing or adding a gap fails
that assertion so the report cannot silently become stale.

`gaps/embedded_jsonc.md` is a minimal proving document. Go preserves the contents
because `native/json.go` registers only `.json`, while embedded `jsonc` maps to
`embedded.jsonc`. The fork formats it and inserts a trailing comma. One output
cannot equal both. `gaps/embedded_flow.md` proves the Flow embedding difference:
the native JavaScript path does not format that Flow program as the Babel/Flow
printer in the fork does. These dependencies are outside this unit's territory.
Front matter likewise invokes YAML/TOML embedding when recognized; it cannot be
silently treated as a raw block in a claim about the default complete formatter.

The audit checks both `auto` and `off` without filtering out embedded files.
The off baseline calls Go's actual `markdown.Format` with a nil embedding
callback, and normalizes/restores BOM/line endings exactly at the core boundary.
The auto baseline calls its actual `native.Formatter`. Neither adapter duplicates
Markdown parsing or printing. The fixed options on both sides are tab width 4,
print width 120, single quotes, spaces, semicolons and preserved prose, with the
remaining Prettier defaults. This is an explicit common option set, not each
nested repository's configuration or ignore rules. Every physical Markdown file
is included, even files formatting enumeration would skip.

## The literal parser representation also needs a redesign

`micromark/types.go` represents states as `func(code Code) State`, with nil meaning
finished. The document/container/flow machines are mutually referring closures
with captured mutable frames. These are additional prerequisites for a direct
Adamic port, not evidence that Markdown cannot ever be implemented in Adamic:

- `gaps/1_recursive_state.ts` prints `35` twice on Node. Stage 0 returns exactly
  `lower.NotYet("a function inside a function (a closure)")` for the nested
  function declaration corresponding to a state with its captured frame.
- `gaps/2_state_arrow_cycle.ts` prints `35` twice on Node. Rewriting the nested
  declaration as a mutable arrow slot is refused by Adamic 0.1 as a closure/cell
  reference-counting cycle. This is a `Refused`, not a `NotYet`.
- `gaps/3_recursive_callable.ts` is closed: a noncapturing, top-level recursive
  callable state is supported. Source Node, native under ASan/UBSan, the
  JavaScript backend and LeakSanitizer agree on `1` then `0`. The recursive
  callable type itself is therefore not the blocker.

A full port can instead use explicit state/frame IDs and owned arenas, with
carefully bounded weak back references. That requires porting and checking the
micromark engine, every construct/extension and mdast compilation, followed by
preprocessing, AstPath and document layout. This branch does not pretend that a
line-oriented approximation is the same parser. No compiler files were modified.

## Reproduction and scope of the checks

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_MARKDOWNBLOCKS_FORK=/tmp/adamic-markdown-blocks-fork \
  ADAMIC_MARKDOWNBLOCKS_KEEP=/tmp/markdown-blocks-audit \
  go test -count=1 -v -timeout=30m ./stage1/cohere/markdownblocks > /tmp/markdown-blocks-test.log 2>&1
```

Install the original fork in that scratch directory by copying the exact
`cohere/internal/format/prettier/bundles/` tree from the pinned submodule. The
adapter checks `prettier.version === '3.9.6'`. The unit report records its pin and
bundle digests. This is an installation of the vendored original, not a guessed
patch to npm Prettier. The stock comparison uses the separately installed exact
npm version in `/tmp/adamic-markdown-prettier/node_modules/prettier`.

The census includes `.md`, `.markdown`, `.mdown` and `.mkd`, skipping only `.git`.
Generated cases cover heading forms, all list marker types and five nesting
depths, fenced/indented code, block quotes, tables, HTML, thematic breaks, front
matter and typed code fences. Three Go-printer mutants alter an ATX heading
prefix, unordered list marker and fence length through overlays. Each must build,
exit zero, produce no stderr or formatting errors, and fail the off baseline's
byte comparison. These prove the preflight comparison catches real output
changes; they are Go mutants, not mutants of an unbuilt Adamic block printer.

No new native Markdown parser, preprocessing, recursive block printer, table
layout, HTML block printer, document engine or embedded-language printer was
ported. The block unit remains incomplete. The prior inline implementation is
composed only as a branch dependency, not yet wired into a block formatter.
A full default-mode common target requires resolving the JSONC/Flow oracle gaps
first; an explicit off-mode contract avoids those embedding dependencies but
still requires the complete parser/frame and document-printer work above.

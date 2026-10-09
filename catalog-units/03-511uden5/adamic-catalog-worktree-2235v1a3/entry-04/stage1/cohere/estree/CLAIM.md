# Stage 1 claim: non-JSON ESTree conversion

Branch: `codex/stage1-estree`. Territory: `stage1/cohere/estree/` and this claim.
Base: current main `5d4c8012a0877094134e6c6bac367ff68f9313e8`. Cohere pin: `715ba94f3608a6500086b1076ce5cb7e51b836db`.
Dashboard read: `origin/codex/stage1-progress` at `7ed6d30c3f2d407a183c971ec2ed2577c0e5f615`.

The unit ports the non-JSON pipeline in `cohere/internal/format/estree`: conversion
of TypeScript parser nodes into ESTree, comments, source locations, visitor keys
and Prettier postprocessing. The existing indexed TypeScript parser is a dependency
and will not be edited. The canonical ESTree driver, not a whole-file JavaScript
formatter, is the byte-identical boundary. `parse_json.go` belongs to the JSON slice
and is excluded. JavaScript printers and Markdown/CSS formatters stay with their
existing branches. Compiler and runtime files will not be edited.

## Origin survey and size

Fetched all 142 origin refs visible on 2026-10-06. Read the dashboard's recorded
coverage and current formatter branch contracts. Current YAML implementation is
newer than the dashboard's saved audit-only YAML entry. Markdown parser/layout is
on `codex/stage1-markdown-blocks`; CSS parser/printer on `codex/stage1-css` and
`codex/stage1-css-printer`; JavaScript expression printing on `codex/stage1-ts-printer`;
GraphQL printing on `codex/stage1-graphql-printer`. Those existing units are treated
as occupied, including their unfinished composition work.

No origin branch contains an ESTree converter/postprocess implementation or claim.
The JSON slice owns only `estree/parse_json.go` (532 lines). The separate remaining
ESTree pipeline is 3,674 non-test physical Go lines, larger than the shared printing
package (1,628), options (540), native dispatch (391), comparison (133) and arena
(154). This ranks complete independent package boundaries rather than taking
unimplemented functions out of another worker's active formatter unit.

| Upstream file | Non-test physical lines |
| --- | ---: |
| convert.go | 2337 |
| visitor_keys_generated.go | 347 |
| node.go | 274 |
| postprocess.go | 259 |
| comments.go | 190 |
| location.go | 101 |
| parse.go | 97 |
| arena.go | 69 |
| Total excluding parse_json.go | 3674 |

## Validation contract

Go cohere decides the converted tree, fields, ranges, comments and postprocess
results. Source Node and emitted JavaScript must match sanitized native and Go
byte for byte on repository/submodule TypeScript/JavaScript and generated cases.
The pinned original typescript-estree/Prettier library is a separate oracle where
its contract matches cohere; any disagreement gets a minimal proving program.
Unsupported parser or converter cases must fail explicitly, never return a
plausible incomplete tree. Three successful wrong-output mutants must be caught
by byte comparisons. Compiler/runtime gaps get minimal programs, with observed
behavior separated from inference. Throughput and coverage will be reported for
the implemented boundary without claiming complete JavaScript formatting.

## Current checkpoint status

The follow-up source audit has 15,647 identical frozen files, zero output
mismatches and 5 acceptance disagreements. The local parser adapts the main
parser and origin/codex/parser-recovery b85afdd inside the ESTree territory; only
scanner and node-model dependencies remain shared. No shared parser/compiler/
runtime files are edited. All 549,618,641 matching bytes pass on source Node, sanitized native and emitted
JS; all 11,717 Go refusals explicitly refuse on native. The complete package gate
and 22 driver mutants pass. The claim remains incomplete at the raw-input
boundary. Follow-up evidence lives
in stage1/cohere/estree/FOLLOWUP.md.

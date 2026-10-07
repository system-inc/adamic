# Dedicated bridge questions

Each question rejects suffixes and wrong node kinds before reading compiler state.
`facts.go` contains only five dispatch cases. The existing registration generator
and test harness are unchanged. Go returns compiler and syntax data; native `.a`
files decide all findings and path states.

| Question | Accepted node | Raw result | Native decoder |
| --- | --- | --- | --- |
| platform-symbol | Identifier or StringLiteral | direct and alias-resolved symbol identity, flags and declaration ancestry | platform_symbol.a |
| call-declaration | CallExpression | selected signature declaration, body location, function flags and return type flags | call_declaration.a |
| program-modules | SourceFile | compiler program files and resolved static/dynamic import edges | program_modules.a |
| output-flow | code path root | reachable blocks, AST event sites and successor edges | output_flow.a |
| source-context | SourceFile | external-module and declaration-file flags | source_context.a |

Declaration ancestry includes name text and kind, source path, declaration and
library status, module status, global augmentation flags, node/function flags,
UTF-8 byte spans, and body kind/span. Symbol identity disambiguates spelling from
binding. Native code handles the TypeScript Never flag (262144), overload bodies,
async/generator flags, process/console/timer declarations and module closure.

The syntax graph builder remains Go infrastructure in `checker/outputflow/`.
It is copied from pinned cohere's generic control_flow_graph package; see its
NOTICE.md and LICENSE. It does not import or run any production cohere lint rule.
This is a limitation of the port boundary: graph construction is not native,
while graph traversal, write-history antichains, catch resets, callee following,
blocking-order analysis and all report decisions are native.

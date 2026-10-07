# Module export predicates

Three pure helpers, one per .a file, match the Go module shelf:

- IsExportedByName requires export and excludes default, in any order.
- HasExportModifier requires export, whether or not default is present.
- IsExported guards a nil node then delegates to HasExportModifier.

ModifierListView.present represents a nil versus nonnil Go list. Tags represent
AST kinds: export-keyword, default-keyword, and any other string. The parser
adapter classifies actual Go ast.Kind values, never source text. Unknown tags
are ignored. Order and duplicates are retained. A false presence flag ignores
all backing fields; a present empty list remains distinct in the view but has
the same false verdict. ExportNodeView.present represents a nil Go node and its
modifiers view comes from the actual node.Modifiers() result.

The caller must adapt a real node/list. A nil element inside a nonnil Go
modifier list panics and is outside this nonnil-element arena boundary. The
helper does not decide how a file exports names via a separate export list,
resolve symbols or bindings, or emulate a checker. It answers precisely the
three upstream modifier questions. No semantic compilation of fixture source is
required: Go's TSX parser also supplies its recovery nodes for malformed controls.

The Go oracle uses actual exported module predicates and the real Go TSX parser.
Every parsed AST node from every consumer source is observed, including nodes
without modifiers. Synthetic lists exercise all sequences of export/default/
other of lengths zero through six, both nil and present lists; the node is nil
for synthetic list observations. Source controls exercise actual nonnil export,
default, declaration, nested and nondeclaration nodes. The Adamic source runner,
sanitized native executable and emitted JavaScript all compare to Go. Native uses
ASan/UBSan and Linux leak checking. See REPORT.md for counts, mutants and limits.

The capture script overlays only Go's harness in a temporary directory to record
asserted source fixtures. Shared repository files and registration are unchanged.
All runnable Adamic files in this batch are .a. The existing options_json.ts
reader is used unchanged for test input. Local readiness.json subtracts only
these three dependencies from the frozen original ledger and records residuals.

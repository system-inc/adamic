# transformers/module/system.ts at zero

All-code followup census **3 -> 0** (initial target file was 4).
The owning moduleInfo, exportFunction and contextObject declarations now include
undefined truthfully. Plain scripts can be skipped before cache population, and
the unconditional emit notification restores absent entries. None of those
sparse cache reads is asserted. No initializer changes.

Widening owners exposed exactly 25 required consumers. Each has an individual
required-context ledger entry and an original-point assertion; no other
identifier read is asserted. Twenty-three consumers live in eleven helper
functions reachable from transformation and unreachable from either emission
root, mechanically checked with stock TypeScript symbol resolution. The graph
includes passed callbacks and returned delegates. All three contexts are
initialized before createSystemModuleBody, reset only after its return, and are
not overwritten by a recursive SourceFile transform in that graph.

Two consumers require conditional emission proofs. createExportExpression uses
exportFunction only during initialized transformation or from the export-name
loop in substituteBinaryExpression. That emission call is guarded by getExports.
Exported bindings/specifiers require present moduleInfo; its other branch
requires resolver.getReferencedExportContainer to return a SourceFile. The
checker requires a ValueModule parent symbol and a same-file source container;
foreign UMD exports return undefined. Thus that branch is a transformed module,
and all three caches were populated together under the same original id.
Binder declareSourceFileMember puts ordinary script identifier declarations in
file.locals with parent undefined; only isExternalModule(file) uses module-member
binding. CommonJS scripts can have a SourceFile symbol, but their ordinary
identifier declarations remain parentless, and qualified exports do not enter
the identifier-assignment guard. Imported aliases imply import syntax and JSON
has no such identifier assignment. Thus the eligible SourceFile export parent
requires a transformed external module, rather than merely any SourceFile symbol.

substituteMetaProperty uses contextObject only under isImportMeta. The parser
walks import.meta to mark an external module; all three program module-detection
modes retain isFileProbablyExternalModule. Thus this source file is transformed
and its context cache is present during substitution. Stock 6.0.3 probes confirm
Legacy, Auto and Force all mark a file containing only import.meta external.
The transformation and emission phases retain the original SourceFile id.

verify-system-context.cjs checks the symbol-resolved graph, exact context writes,
initialization/reset order, conditional emission guards, and module-detection
observations. The required moduleInfo ! -> ?? 0 mutant fails emitted-JavaScript
and site-contract checks. Owner and conditional protocol mutants are retained.
All stock JavaScript bytes, idempotence and CRLF are preserved.

Source-only census: /tmp/emit33-system-close-census; before snapshot:
/tmp/emit33-close-source-before-system-close; default oracle:
/tmp/emit33-system-close-oracle. Use the configured Go environment and
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules for stock 6.0.3.

Default oracle: **106367 passing**, zero failing/pending, empty baseline diff,
214.121s. Mechanical API projection accepts exactly 20's 189
and 40's 28 lines; no new API or other baseline change. Removing each optional
owner independently restores exactly its original TS2322 at lines 1773, 1774
or 1776 and fails the owner contract. Calling a required transform helper
from the emission root fails the graph proof. Replacing either import.meta or
export-name guard with true fails its conditional emission proof. All seven
mutations are caught, and source-only owners are restored in finally blocks.
Final JavaScript equality and idempotence pass.

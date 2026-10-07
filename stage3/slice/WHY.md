# Why the parser slice reached the whole compiler

Input for this reproduction: scanner-adapted50, TypeScript 6.0.3 with adaptations
10 and 50, excluding 20. Parser driver is copied from 3cb0846. The parser worker
used a different fully adapted area tree; its 5,103/79-file count is not asserted
for this input. Conservative mode here reaches 5,099 records and 190,832 lines.
Both final outputs match the previously reported full-tree oracle hashes.

The previous module-namespace path expanded every export even for static
property access. Casts also hid the qualified relationship and property symbol.
A shortest historical chain for both createTypeChecker and createProgram is:

createSourceFile -> tracing -> tracingEnabled.stopTracing (typeof namespace)
-> tracingEnabled.dumpTypes -> Debug.formatTypeFlags -> ts namespace expansion.

The last edge is Debug's `(ts as any).TypeFlags`, not dynamic reflection. Another
path reaches Debug.formatSyntaxKind through Parser.parseSourceFile -> convertToJson
-> getTextOfPropertyName -> tryGetTextOfPropertyName -> Debug.assertNever. The
parser worker's Record-cast spelling of SyntaxKind is explicitly tested too;
its full dump has the same hash. The narrow honest rule implemented is to
unwrap identity-preserving casts/parentheses and follow the named namespace
member. Unread facade exports supply bindings without claiming their members
were referenced. Dynamic or bare namespace uses still expand every export.
A control with static access followed by a dynamic access behind || still
retains createTypeChecker and createProgram; the tool does not prune that branch.

tracing is initially undefined and is enabled by startTracing. The driver's
createSourceFile calls do not enable tracing; dumpTypes belongs to stopTracing.
Those functions remain retained as required by the typeof namespace declaration
and the current declaration-granularity contract. This trace is not evidence
that parsing executes type-dump code. We do not remove their reachable code on
a runtime assumption. The static-member rule fixes the false whole-barrel edge
without such an assumption.

Current --why results: createTypeChecker, createProgram and getPreEmitDiagnostics
are not reached. getNodeId remains reached, through Parser.parseSourceFile ->
Parser.parseJsonText -> Parser.parsePrefixUnaryExpression -> Parser.factory ->
createNodeFactory -> getNodeId. Its reference is nodeFactory.ts:1381 in the
getGeneratedNameForNode API. Whole createNodeFactory, its returned method table
and all its statements stay intact. This shortest path goes through the JSON
branch although the driver passes ScriptKind.TS; the same factory is also used
by ordinary TS parsing. One shortest chain is not proof that all paths have an
unmet condition. Narrowing factory methods would require a separately authorized
change in gathering granularity or a source adaptation; it is not done here.
The parser driver separately requests flattenDiagnosticMessageText from program.ts.

Evaluation repair: copying the original side-effect graph alone failed because
rewritten direct bindings introduced additional edges before barrel evaluation
finished. Retain the value binding's original module, copy original facades,
and emit original runtime requests in original order, omitting type-only edges.
The emitted graph includes empty import-only modules, not their unreachable code.
Declaration comments/bodies and namespace-member bytes remain verbatim.

Final code counts: scanner 89 declarations, eight files, 7,127 code-span lines;
parser driver 1,988 declarations, 26 files, 41,677 code-span lines. The original
78-module graph is present in both outputs. The manifest separately counts 77
export facades, namespace wrappers, code files and evaluation-only files.

Commands (each execution redirected to its own log):

```sh
SLICE_TYPESCRIPT=/path/to/typescript/lib/typescript.js bash stage3/slice/run.sh TREE NEW_SLICE src/compiler/parser.ts:createSourceFile --why src/compiler/checker.ts:createTypeChecker --why src/compiler/program.ts:createProgram
node stage3/slice/verify.cjs NEW_SLICE
# Full parser driver entries are recorded in evaluation-proof.json.
PARSER_TYPESCRIPT=... PARSER_RUNTIME=oracle/adamic.mjs node COPIED_PARSER/node.mjs SLICE/parser-proof-main.a FIXED_INPUTS MANIFEST
bash stage3/drivers/scanner/run.sh NEW_RUN --tree SCANNER_SLICE --inputs FIXED_INPUTS --node-only
cmp FULL_TREE_STDOUT SLICE_STDOUT
```

Scanner comparison and parser comparison exit 0, empty stderr, exact hashes.
Audit: scanner 168 spans / 78 ordered lists; parser 2,077 spans / 78 ordered lists.
Reverse-import mutants preserve declaration bytes and fail Node initialization:
scanner reads ScriptTarget.ES3 before initialization; parser reads timestamp
before initialization. Both audits reject specifically evaluation order at the
barrel's first import, after the declaration byte checks pass. Scanner's token
end+1 mutant also fails its comparison. No new native build, compiler gate or
upstream baseline suite is claimed for these import/gathering tool changes.

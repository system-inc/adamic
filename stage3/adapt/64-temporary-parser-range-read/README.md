Temporary: comes out when stable indexed-read invariants are proved for compiler-owned arrays

Checked invariant, scoped to createSourceFile and tsc's own callers: once
from[i] !== undefined succeeds, the second from[i] remains present. The
intervening to.push lookup resolves an ordinary method and cannot change the
source slot. Insert only ! at the second read, retaining both reads and their
order. Adamic must fail loudly if absence reaches that asserted read. Stock
TypeScript erases !, preserving emitted JavaScript byte for byte.

The user explicitly permits this checked invariant on invalid states. It is
not a universal claim about the exported addRange utility or arbitrary client
arrays. Partition 30's getter counterexamples remain real public-API examples;
no compiler-owned caller constructs those states. External factory patches,
custom module-indicator callbacks that mutate arrays, proxies, accessors on
indices/push, or mutation of Array.prototype are outside this proof's domain.

## Complete retained caller inventory and producer evidence

The stock AST audit finds 13 direct calls, no aliases/property references to
addRange, and exactly three internal flatMap/sameFlatMap callbacks. Exact slice
lines, expressions and file hashes are saved in parser/evidence/range-callers.json.

| Owner | Calls | Source and destination provenance |
| --- | ---: | --- |
| Parser.reparseTopLevelAwait | 4 | statements originates in parseList: fresh [], append by push, then createNodeArray; the destination is fresh []. savedParseDiagnostics aliases initializeState's [] populated by diagnostic pushes; the new parseDiagnostics destination is []. All loops bound the range and no callback runs between the two reads. |
| Parser.JSDocParser.parseJSDocComment | 1 | from is the parser diagnostic [] populated by parseErrorAtPosition. jsDocDiagnostics is allocated as [] immediately before its first use. The later truncation of parseDiagnostics occurs after addRange returns. |
| core.flatMap | 1 | result is undefined, a slice from the first array, or append's [] result. The only retained callbacks are filterOwnedJSDocTags and getJSDocTagsWorker; their array results are parsed tag lists, filter results, or [jsDoc]. Each callback completes before addRange starts. |
| core.sameFlatMap | 1 | Only createNodeFactory.createCommaListExpression calls it, with flattenCommaElements. It returns an existing compiler-created node.elements array, [node.left, node.right], or a node. The result is array.slice(0,i); the callback finishes before addRange. |
| utilities.getJSDocCommentsAndTags | 4 | result begins undefined and is thereafter an addRange result. from is filterOwnedJSDocTags or the parameter/template tag functions; these use compiler-owned parsed JSDoc/tag arrays, map/filter/flatMap and fresh array literals. |
| nodeFactory.mergeEmitNode | 2 | destinations are leadingComments.slice()/trailingComments.slice(). Sources are synthetic comment arrays maintained by tsc. Fresh parser nodes have emitNode undefined, so these branches do not run on a fresh createSourceFile parse; source-file update likewise copies an original with no emitNode. They are retained because setOriginalNode is a whole declaration used by update helpers. |

Supporting producer review: Parser.initializeState initializes parseDiagnostics
with []; Parser.parseList uses [] and push; withJSDoc uses mapDefined to build
jsDoc; JSDocParser creates tags via []/push/createNodeArray. filterOwnedJSDocTags
returns filter results or [jsDoc], getJSDocTagsWorker flattens those lists, and
parameter/template queries use Array.filter. createNodeArray retains an
ordinary input array or calls slice and adds metadata fields; it does not
install numeric accessors or override push. Debug.attachNodeArrayDebugInfo
uses Object.create(Array.prototype) and adds only __tsDebuggerDisplay. It adds
no indexed or push getter. The default module indicator scans syntax. tsc's
getSetExternalModuleIndicator callbacks set only externalModuleIndicator;
none installs array accessors or mutates the source/destination arrays.

Undefined elements or holes at the first read are valid and skipped. Density
of every input array is not assumed. Stability of the conditional second read
is the invariant; a global promise that every array element is present would
be wrong. Source strings can describe getters but parsing never evaluates
those strings to execute getters.

## Validation

The adaptation list was committed and pushed as 8dd6805 before implementation.
The guarded adapter accepts exactly one addRange site, is idempotent, and
compares stock TypeScript 6.0.3 emitted JavaScript byte for byte before writing.

Observed results (not substitutes for the producer argument):

- Retained-reference audit: 13 calls, no aliases, three flattening callbacks.
  Adding an unreviewed addRange alias makes the audit fail (exit 1).
- Strict feature-integrated slice checker: zero diagnostics. Removing only !
  restores exactly TS2345 at core.ts:379:21; source restored afterward.
- Parser: 35,456,964 bytes, SHA256
  2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
  The planted Identifier end change alters exactly line 164 and cmp exits 1.
- The runtime checked model also matches the whole parser dump. That corpus
  executes zero conditional second reads. Directed TS/top-level-await and
  JS/JSDoc cases execute five; their tree/diagnostic output is identical.
- Shared-core scanner checked-model run: 509,014 tokens, 27,879,197 bytes,
  SHA256 c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f.
  Control diff exits 0 and token-end mutant diff exits 1.
- 48 ordinary/sparse/undefined/offset controls agree. Indexed-getter and
  push-getter invalid states append own undefined in original addRange and
  throw in the checked model. Removing just the guard admits both again.
- Native range-check.a on non-null-check worker a02613e: control stdout 23,
  exit 0; invalid second-read mutant exit 70 with
  `adamic: panic: non-null assertion failed: from[0]! is null or undefined`.
  The existing six-feature scratch compiler still refuses ! in this isolated
  fixture; the native assertion result is from the worker, not a main gate.
- Default stage3 suite on the full tree with 60-64, with ONLY the second-read
  assertion replaced by a real runtime guard: 106,367 passing, zero failing or
  pending, no baseline diffs; install/build/tests exit 0, 265.696 seconds.
  This exercises checked behavior rather than relying on stock ! erasure.
  The exact partition-32 sanctioned API check passes afterward and verifies
  all 60,930 other reference baselines unchanged; no parser API exception.

Evidence and compressed logs are in ../../drivers/parser/evidence/range-*.
Commands use CENSUS_TYPESCRIPT pointing to the pinned stock 6.0.3 API:
`node adapt.cjs TREE`, `node audit-range-callers.cjs SLICE OUTPUT.json`,
`node check-range-model.cjs SLICE OUTPUT.json`, parser `run.sh SLICE OUT
--inputs CORPUS`, scanner `run.sh OUT --tree TREE --inputs CORPUS --node-only`.
The full-suite checked model is reproducible with parser's
`instrument-range.cjs TREE` followed by the existing stage3/oracle/run.sh.
The same exact sanctioned API preparation from temporaries 60-63 applies.

Lowering is now reached and first refuses `an import cycle` at core.ts:1:1.
The attempt to integrate import-cycles 779ff9d in scratch could not build with
its SDK after conflict resolution: the retained older loader uses pre-typed
path APIs. This is a scratch integration failure, not a second slice lowering
finding. No native parser binary or complete native dump is claimed.

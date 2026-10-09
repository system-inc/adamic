Temporary: comes out when library's Node-types loader lands on main, moving this removal to adaptation 40

Remove `(process as any).browser` from TypeScript 6.0.3's
`src/compiler/core.ts`, `isNodeLikeSystem`. The runtime expression becomes
`process.browser`. This temporary moves into adaptation 40 once the loader is
on main; it is valid only with library's loader and pinned @types/node 25.3.3.

Adaptation 30 currently shadows the global process with a local ambient
unknown-field declaration. Merely dropping the cast would leave that fabricated
local contract in use. This adapter therefore changes that declaration to
`declare const process: (NodeJS.Process & { browser?: unknown }) | undefined;`.
The Node contract comes from the official package. Its Process interface has
no browser member, so the existing optional unknown browser extension remains
explicit. The runtime tests process absence, nextTick truthiness, browser
falsiness and require absence exactly as before. No runtime guard, initializer,
import, fallback Node binding, non-null assertion or caller change is added.
The existing ambient require declaration is unchanged.

The adapter verifies stock TypeScript 6.0.3, the exact Node package version,
one function owner and one browser probe, and rejects partial or drifting
inputs. It is idempotent. Before writing, it compares emitted JavaScript byte
for byte both with comments and without comments. Source maps and declaration
metadata are not runtime JavaScript.

The plan was pushed before implementation as 9bc3fe17. Validation results and
commands will be retained in evidence/. The shared core site was explicitly
authorized by the user. reduceLeft and its arguments.length read stay unchanged.

## Observed validation

- The plan was pushed as 9bc3fe17 before implementation.
- Real newest scratch loader: strict checker accepts the adapted driver slice,
  exit 0, empty stderr. host-types.a exposes Process and browser?: unknown.
- Wrong nextTick number assignment fails TS2322, proving it is the official
  function contract rather than the earlier unknown field. Wrong Node package
  version fails the adapter's pin guard. Neither mutant writes adapted source.
- A planted ! to !! browser test change fails the emitted-JavaScript equality
  guard, exit 1, before writing. An unchanged second application exits 0 and
  preserves the exact source bytes.
- Fresh before/after upstream builds: all ten emitted .js/.mjs/.cjs artifacts
  are byte-identical, including typescript.js, _tsc.js and the test runner.
- Full default stage 3 oracle on the previously sanctioned adaptation-60-to-64
  predecessor, with 65 added: 106,367 passing, zero failing/pending, no baseline
  differences. Install/build/tests exit 0; total 279.367 seconds. No new API
  lines accepted. Existing exact owner attribution confirms only the 40
  sanctioned API line changes and all 60,930 other reference baselines unchanged.
  This is the sanctioned predecessor gate, not a claim about every subsequent
  unsanctioned API change in the newest area tree; the newest slice's strict
  loader-backed checker is a separate result above.
- Parser Node dump remains 36,429,231 bytes, SHA256
  686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
  Node-end, dropped-tags and JSDoc-diagnostic mutants are caught by run.sh.
- Shared-core scanner remains 509,014 tokens, 27,879,197 bytes, SHA256
  c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f,
  cmp 0 against the previous fixed corpus; its token-end mutant is caught.

The initial scratch copy mistakenly excluded nested tracked node_modules and
built test fixtures, causing fourteen project failures. Restoring the 32
missing tracked files gave the full green result above. That failed report and
log remain in evidence, alongside the successful rerun. An old built artifact
still contained adaptation 64's runtime-check instrumentation; the JavaScript
comparison uses a fresh predecessor build instead, with matching source.
No compiler edits, scanner adaptations or native parser result are claimed.

## Reproduce

Set CENSUS_TYPESCRIPT to the pinned stock 6.0.3 API, and CENSUS_NODE_TYPES to
library's stage3/api/node_modules/@types/node 25.3.3 directory. For a disposable
prepared tree, run `node adapt.cjs TREE`. Run
`python3 verify.py BEFORE AFTER COMPILER SCRATCH_COMPILER_ROOT NEW_OUTPUT` to
check idempotence, the runtime and version mutants, and the loader contracts.
Keep compiler_root's official Node API seat installed. The verification script
uses no replacement Node declarations.

Build the sanctioned predecessor and adapted trees separately with
`npm run build --prefix TREE` and compare every built .js/.mjs/.cjs file.
Use `stage3/adapt/32-indexed-reads-program/public-api.cjs PRISTINE TREE
OPTIONAL20_LEDGER stage3/adapt/30-indexed-reads/remaining-readonly-owners.json`
with TSC_ADAPT_TYPESCRIPT set for exact baseline attribution, without --accept-api.
Then run `NODE_OPTIONS=--max-old-space-size=1536 bash stage3/oracle/run.sh TREE
NEW_ORACLE_OUTPUT`, redirecting logs to files. Preserve tracked fixture directories
when copying a predecessor; exclude only its top-level dependency/build folders.
Parser and scanner commands use the fixed /tmp/parser-adapted10 input corpus;
parser run.sh takes --inputs CORPUS, and scanner run.sh takes --inputs CORPUS
--node-only. Evidence retains exact paths, outputs and phase commands.

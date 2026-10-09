# Scanner private owner types

Permanent adaptation 42 makes the scanner's private keyword map explicit as
`Map<string, KeywordSyntaxKind>` and changes the private `createScanner.error`
overload and implementation's `arg0?: any` to `arg0?: string | number`.
It edits types only. It adds no assertion, changes no body, and edits neither
the public `ErrorCallback` nor the `Scanner` interface nor API baselines.

The owning dictionary is `textToKeywordObj: MapLike<KeywordSyntaxKind>`.
Every entry maps keyword text to a keyword enum member, including the computed
constructor key. Its sole map consumer reads `get(tokenValue)`, using scanner
text as a string key and returning a keyword or Identifier. Explicit constructor
type arguments retain that contract even when a discovery placeholder erases
the dictionary's index signature. That placeholder still prevents lowering
Object.entries into typed pairs; this adaptation does not claim to fix it.

Diagnostic substitutions come from every four-argument call of the nested
error helper: string literals, text slices, escape strings, spelling suggestions,
regular-expression names and target names, plus the numeric
`numberOfCapturingGroups` for the backreference diagnostic. Optional target names
and omitted arguments allow undefined. Consequently the own driver callback
uses `string | number | undefined`, rather than its contextual any.

Read patterns: adaptation 40's README, AST owner/body checks and diagnostic
class rules; adaptation 41's README and guarded exact-site adapter at fetched
`origin/codex/stage3-explicit-any-2` `1e920a31`. Adaptation 40 already applies the
same private diagnostic narrowing in the full main pipeline, so 42 accepts
that reviewed type without changing it. The scanner's small source profile
omits 40; 42 performs both private parameter edits there. The public callback
has whatever type preceding adapters established, unchanged by 42.

The adapter requires stock TypeScript 6.0.3, matches parsed private owners,
checks the forwarding helper's unchanged body, and rejects missing, duplicated
or drifted sites. It preserves LF/CRLF source bytes outside its three edits.
A second pass makes zero edits. apply.sh discovers it in numerical order.
The scanner runner and slice helper include it in their source profile.

## Reproduce

Commands write output to logs. Use fresh result directories and source the
setup environment. Main and 42 lanes run apply and the full upstream oracle;
no runner or test filter is used. Their declaration artifacts and baseline
diffs must match byte for byte. Each lane must pass its existing sanction.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > setup.log 2>&1
source /workspace/adamic-tools/env.sh
STAGE3_CACHE=CACHE bash stage3/lane/run.sh RESULTS > lane.log 2>&1
NODE_PATH=CACHE/api/node_modules node stage3/adapt/42-scanner-any/prove.cjs \
  MAIN_TREE RESULTS/adapted-tree stage3/drivers/scanner/main.a PROOF > proof.log 2>&1
```

Run the same lane from the recorded pristine main checkout for MAIN_TREE.
Proof uses a virtual source file inside src/compiler to check the actual driver
callback against the scanner, retaining its body and changing only import paths
and declarations for Adamic prelude operations. No virtual source is written.
Stock checking covers the complete compiler and actual scanner callers.
It records every substitution expression and inferred type, asserts both files'
emitted JavaScript is identical, and checks public callback source identity.

Three source mutants independently make the private payload string-only,
make keyword map values string, and make the driver's payload string-only.
Stock TypeScript must reject each actual owner/consumer contract. Separate
native discovery mutants restore each original any/inference and must reproduce
its previous exact any diagnostic, while concrete controls reach later gates.
Drift and duplicate-owner mutants must fail the adapter's guard. No fixture
bucket or counts.md change is needed for this source adaptation.

See [REPORT.md](REPORT.md) and evidence for measured commits, commands, Node
stream hashes, lane/oracle results, mutant diagnostics and native limitations.
Upstream snippets in evidence are TypeScript 6.0.3, Microsoft copyright,
Apache-2.0 licensed; upstream source is kept in external scratch trees.

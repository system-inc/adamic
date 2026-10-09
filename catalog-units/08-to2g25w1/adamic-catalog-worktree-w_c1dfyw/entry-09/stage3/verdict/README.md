# Stage 3 verdict

Run the native compiler as soon as it links:

```sh
stage3/verdict/run.sh --tsc /absolute/native-tsc /tmp/new-verdict > /tmp/verdict.log 2>&1
```

The output directory must not exist. Python 3, git, and the compiler's standard
libraries are required. The compiler argument is one executable path, including
paths with spaces, not a shell expression. No Node dependency is needed to judge
a native compiler. `summary.json` and `summary.md` contain the whole verdict:
pass/fail counts, exclusions, and the first differing byte with surrounding text.
Exit 0 means every selected case passed, 1 means compiler differences, and 2
means a harness/infrastructure error. All three suites are attempted even when
an earlier suite fails. A harness error is never a pass.

`STAGE3_VERDICT_UPSTREAM=/absolute/pinned-checkout` reuses a TypeScript checkout.
Otherwise the command clones v6.0.3 into the output directory, from the existing
`STAGE3_CACHE/typescript.git` mirror when available, or GitHub. HEAD must be
`050880ce59e30b356b686bd3144efe24f875ebc8`. Source and reference hashes are checked
against the committed census before each selected baseline case is run.

## Suites

| Suite | Comparison |
|---|---|
| acceptance | All 301 existing driver projects, including tiny; exact stdout, stderr, exit goldens |
| tiny | The existing three-file tiny project again; the same three independent byte checks |
| baselines | Selected compiler and conformance cases; upstream diagnostic summary, empty stderr, expected CLI exit |

Acceptance calls the existing driver unchanged. `NATIVE_TSC` is removed from its
child environment so a second compiler cannot accidentally affect the verdict.
Every project retains tsconfig, argv and raw captures. `TSC_JOBS` defaults to 4;
`TSC_TIMEOUT` defaults to 60 seconds per invocation. Baselines retain extracted
expected streams as well. A timeout fails even if its partial output matches.

## What the upstream subset means

`selection.json` enumerates every `.ts` and `.tsx` compiler/conformance case,
recursively, at the pin. It records selected inputs and every excluded filename
with its reason and source SHA256. Each run copies the exclusions and selection
into `baselines/`. Counts are input files; excluded option variants are not
expanded into separate configurations. Exclusions are scope decisions, never
compiler passes.
Selection uses stock TypeScript's parser for syntax eligibility, never the tested
binary's diagnostic results. There is no source size bound. The early version
bounded sources to 80 lines and 8192 bytes; that bound was removed after its runnable smoke proof was pushed.

Supported single-valued compiler directives are the existing driver's
`corpus.CANONICAL` allowlist. `materialize` follows upstream
`src/harness/harnessIO.ts:makeUnitsFromTest`: removes metadata, uses LF, discards
leading empty content, and preserves the resulting diagnostic positions.
The runner uses upstream's `skipDefaultLibCheck: true` and
`noErrorTruncation: true`, overridden by headers; `types: []` prevents ambient
packages in the real filesystem from contaminating the virtual-host equivalent.
`ignoreDeprecations: "6.0"` enables the older test targets. `--noEmit --pretty false`
selects the CLI semantic-diagnostics domain. Libraries come from the supplied
binary's installation; missing or wrong libraries are observable failures.

The upstream harness uses `getPreEmitDiagnostics` and emit diagnostics, and its
`.errors.txt` also includes annotated source. We compare its initial diagnostic
summary block, with only reference CRLF converted to Linux LF and one final
newline. Raw stdout is retained byte for byte. For this baseline suite only,
`actual.diagnostics` removes the exact case scratch directory prefix, matching
upstream `src/harness/util.ts:removeTestPathPrefixes`, which removes `/.src/`
from the summary, including quoted module names. No other paths, filenames,
locations, codes, text, spacing or actual newlines are normalized. Acceptance
and tiny still compare entirely raw bytes. An absent baseline means clean,
following upstream convention. Locations, diagnostic codes, message chains,
ordering and spacing must match. Baseline exit is 0 when clean, 2 on diagnostics.

The first version excludes virtual `@filename` units, all external module
specifiers and triple-slash references, option variants, unsupported directives
(including symlinks, currentDirectory, baselineFile and captureSuggestions),
declaration emit, syntax-error inputs, non-file/global diagnostic summaries,
diagnostics outside the single source (including standard-library `(--,--)`
placeholders), TS18027 emit-resolver errors, non-UTF8
sources. These exclusions avoid claiming CLI
faithfulness for the API/virtual-host scenarios not implemented here. It does
not compare annotated source, emitted JS, types/symbols, suggestions or traces.
`--baseline-limit N` is an explicit smoke mode; it reports eligible but unrun
cases as **deferred**, separately from exclusions. It is not a full measurement.

To regenerate the census, using the stock 6.0.3 API installed by `stage3/apply.sh`:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/census.py /absolute/pinned-tree \
  /absolute/stock/typescript/lib/typescript.js stage3/verdict/selection.json > /tmp/census.log 2>&1
```

## Stand-ins and proof

Build the existing adapted CLI outside the repository:

```sh
source /workspace/adamic-tools/env.sh  # use the path cloud/setup.sh printed
stage3/apply.sh /tmp/verdict-adapted > /tmp/apply.log 2>&1
(cd /tmp/verdict-adapted && npm ci --ignore-scripts --no-audit --no-fund > /tmp/npm.log 2>&1)
(cd /tmp/verdict-adapted && npx hereby tsc --no-typecheck > /tmp/bundle.log 2>&1)
export STAGE3_VERDICT_NODE_TSC=/tmp/verdict-adapted/built/local/tsc.js
export STAGE3_VERDICT_UPSTREAM=/tmp/verdict-adapted
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/prove.py /tmp/new-proof > /tmp/proof.log 2>&1
```

A forwards to Node. B forwards to the same CLI and changes exactly one `e` to
`E` in one diagnostic, preserving stderr, exit and length. The proof restricts
it to `argument.ts` (tiny, also in acceptance) and
`ArrowFunctionExpression1.ts` (upstream). It verifies the newly failing captures
differ from A by exactly one byte and fail only stdout. C exits 1 without output;
the proof requires zero passes in every suite. `prove.py --baseline-limit 40`
provides a smaller first-run demonstration. This is executable plumbing and
Node behavior evidence, not native compiler correctness.

Focused comparison checks:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/test_verdict.py > /tmp/verdict-checks.log 2>&1
```

These independently mutate stdout, stderr and exit, and exercise differences at
EOF. Census mutants change the pin, selected/excluded counts and source uniqueness;
each is rejected. Artifact-pin mutants append a byte to a source and a baseline
in a disposable checkout, prove the SHA256 guard rejects them, and restore them:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/audit_mutants.py /absolute/disposable-pinned-tree > /tmp/pin-mutants.log 2>&1
```

Use a separate checkout for these mutants, never a tree serving another run.
No Adamic oracle fixtures are added, so its counts table is unchanged.

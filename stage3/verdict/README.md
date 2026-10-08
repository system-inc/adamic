# Stage 3 verdict

Run the native compiler as soon as it links:

```sh
stage3/verdict/run.sh --tsc /absolute/native-tsc /tmp/new-verdict > /tmp/verdict.log 2>&1
```

The output directory must not exist. Python 3, git, Linux bubblewrap with unprivileged user namespaces, and the
compiler's standard libraries are required. Absolute-path cases report an
infrastructure error if mount namespaces are unavailable. The compiler argument is one executable path, including
paths with spaces, not a shell expression. No Node dependency is needed to judge
a native compiler. `summary.json` and `summary.md` contain the whole verdict:
pass/fail counts, exclusions, and the first differing byte with surrounding text.
Exit 0 means every selected case passed, 1 means compiler differences, and 2
means a harness/infrastructure error. All three suites are attempted even when
an earlier suite fails. A harness error is never a pass.
Use an isolated output location such as `/tmp`: ancestor `package.json` or
`node_modules` entries would change the real CLI's package and ambient-type
resolution, so the harness rejects those locations.

`STAGE3_VERDICT_UPSTREAM=/absolute/pinned-checkout` reuses a TypeScript checkout.
Otherwise the command clones v6.0.3 into the output directory, from the existing
`STAGE3_CACHE/typescript.git` mirror when available, or GitHub. HEAD must be
`050880ce59e30b356b686bd3144efe24f875ebc8`. Source and reference hashes are checked
against the committed census before each selected baseline case is run.

## Comparing compilers

```sh
stage3/verdict/run.sh --compare /absolute/first-tsc /absolute/second-tsc \
  /tmp/new-comparison > /tmp/comparison.log 2>&1
```

Both binaries run all three suites independently against the unchanged Node
expectations. `left/` and `right/` retain their ordinary verdicts and captures;
`comparison.json` and `comparison.md` add agreement counts, every disagreement,
observed cause groups and up to three examples per group. The command prints
the agreement table. Exit 0 requires agreement and oracle passes for both
compilers; two identical wrong outputs still exit 1. Infrastructure errors exit
2. `--baseline-limit` remains explicit smoke mode for both sides. Baseline
agreement uses the existing diagnostic projection; driver agreement uses raw
stdout, stderr and exit. Classification never changes the byte comparison.

To analyze two completed verdict directories without running the compilers again:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/comparison.py \
  /absolute/first-results /absolute/second-results /tmp/new-agreement \
  > /tmp/agreement.log 2>&1
```

The Go measurement uses the exact Linux x64 binary recorded by performance
commit `afb0651c`. Its original lockfile and binary pin are in
`toolchains/typescript-go/`. Install outside the repository, since package
metadata must not contaminate isolated compiler cases:

```sh
mkdir /tmp/new-tsgo-install
cp stage3/verdict/toolchains/typescript-go/package*.json /tmp/new-tsgo-install/
npm ci --prefix /tmp/new-tsgo-install --ignore-scripts --no-audit --no-fund \
  > /tmp/tsgo-install.log 2>&1
stage3/verdict/run.sh --compare /absolute/node-tsc-wrapper \
  /tmp/new-tsgo-install/node_modules/@typescript/native-preview-linux-x64/lib/tsgo \
  /tmp/new-node-go-comparison > /tmp/node-go.log 2>&1
```

No Go wrapper is needed. The measurement preserves its bundled libraries,
removed-option diagnostics and native exit statuses. The expected TypeScript
6.0.3 outputs stay unchanged. See [TYPESCRIPT_GO.md](TYPESCRIPT_GO.md) for the
measured pass counts, classified disagreements, examples and capture provenance.

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
into `baselines/`. Pass/fail totals count compiler configurations. The census
separately counts selected input files and wholly excluded input files;
`configuration-exclusions.json` records every known excluded configuration,
including variants of partially admitted inputs. Exclusions are scope decisions,
never compiler passes. Unknown configurations of host-only inputs are not counted
as runs. Selection uses stock TypeScript's parser for syntax eligibility, never the tested
binary's diagnostic results. There is no source size bound. The early version
bounded sources to 80 lines and 8192 bytes; that bound was removed after its runnable smoke proof was pushed.

`cases.py` follows upstream `src/harness/harnessIO.ts:makeUnitsFromTest`: splits
`@filename` units, removes metadata, uses LF, discards leading empty content,
decodes UTF-8/UTF-16 BOMs, and preserves diagnostic positions. Root-file selection
follows `compilerRunner.ts`, including its last-unit rule for require, reference
path and noImplicitReferences. JSON units remain on disk but are not source roots.
Compiler directives use the pinned compiler's option declarations in `options.json`.
Variants expand as the harness does: Cartesian products, aliases deduplicated,
wildcards and exclusions, at most 25 configurations, and sorted baseline suffixes.
Lists such as `lib` remain lists. Metadata for other baselines is ignored.
The runner uses upstream's `skipDefaultLibCheck: true` and
`noErrorTruncation: true`, overridden by headers. Provided type packages use
inferred type roots; otherwise the CLI searches only the case's local type root.
Deprecation diagnostics are preserved. Emit, declaration, noEmit and noEmitOnError
follow the case settings; expected exits follow pinned emitSkipped behavior.
Libraries come from the supplied
binary's installation; missing or wrong libraries are observable failures.

The upstream harness uses `getPreEmitDiagnostics` and emit diagnostics, and its
`.errors.txt` also includes annotated source. We compare its initial diagnostic
summary block, with only reference CRLF converted to Linux LF and one final
newline. Raw stdout is retained byte for byte. For this baseline suite only,
`actual.diagnostics` removes the exact case scratch directory prefix, matching
upstream `src/harness/util.ts:removeTestPathPrefixes`, which removes `/.src/`
from the summary, including quoted module names. Library-placeholder cases additionally reproduce the upstream library header
locations and virtual diagnostic ordering as described below. Diagnostic
messages, codes, spacing and actual newlines remain byte checked. Acceptance
and tiny still compare entirely raw bytes. Rooted units, working directories,
path options and symlinks use an isolated filesystem tree. Exact known path
translations reproduce the virtual names; source text remains unchanged.
Embedded projects use `--project`. Config-only options and literal string `null`
use a small generated option config; only its synthetic diagnostic location is
removed, since API options have no config AST. Real test config locations remain
checked. Traces, file listings and performance statistics belong to other
baselines and are disabled for this diagnostic projection.
Pretty cases retain ANSI colors and context: only the CLI's one extra reporter
newline per diagnostic and its separate footer are removed to match the API
formatter block. Internal whitespace remains byte checked.
An absent baseline means clean,
following upstream convention. Locations, diagnostic codes, message chains,
ordering and spacing must match. Baseline exit is 0 when clean, 1 when diagnostics
skip outputs, and 2 when outputs are generated in the presence of diagnostics.

The census selects 11,830 inputs and 13,693 configurations, leaving 614 inputs
excluded. The current reason groups, six new filesystem/formatting groups and
A/B proofs are in [REMAINING.md](REMAINING.md). [UPSTREAM.md](UPSTREAM.md) records
the previous expansion; [TYPESCRIPT_GO.md](TYPESCRIPT_GO.md) records its earlier
comparison scope. Remaining cases require API diagnostic collection, historical
compiler-version overrides, suggestions or canonical case-insensitive aliases.
Syntax-only and global-only cases are supported. The noEmitOnError path also
admits cases where emit collects exhaustive diagnostics.

Absolute references run the unchanged binary in a Linux mount namespace over
temporary case files. Runtime paths are mounted read-only; virtual roots and
emit destinations are private writable temporary paths. Pinned tests/lib inputs
mount at /.lib; shipped declaration input data mounts at /.ts. Resource hashes
are checked before execution. See [resources/NOTICE.txt](resources/NOTICE.txt)
for pin, license and regeneration details. The compiler implementation is not
used as a runtime API by the native verdict. Drive-prefixed names retain compiler
path geometry through colon named directories. Conflicting case-folding aliases
remain excluded. For projects at filesystem root, the exact pinned virtual
input list is appended to the temporary JSONC config, preserving original
option text and locations, so runtime mounts do not enter project root globs.
The output-path-check directive uses a private emit directory for its diagnostic
projection. Default-library summary cases retain messages and codes while
reproducing upstream `lib.*.d.ts(--,--)` placeholders and virtual file ordering.
The harness does
not compare annotated source, emitted JS, types/symbols, suggestions or traces.
`--baseline-limit N` is an explicit smoke mode; it reports eligible but unrun
cases as **deferred**, separately from exclusions. It is not a full measurement.

To regenerate the census, using the stock 6.0.3 API installed by `stage3/apply.sh`:

```sh
node stage3/verdict/options_schema.cjs /absolute/stock/typescript/lib/typescript.js \
  /absolute/pristine-pinned-tree > stage3/verdict/options.json
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

A forwards to Node. B forwards to the same CLI and changes exactly one diagnostic byte. For error diagnostics it changes `e` to
`E` in one diagnostic, preserving stderr, exit and length. The proof restricts
it to `argument.ts` (tiny, also in acceptance) and
`ArrowFunctionExpression1.ts` (upstream). It verifies the newly failing captures
differ from A by exactly one byte and fail only stdout. C exits 1 without output;
the proof requires zero passes in every suite. `prove.py --baseline-limit 40`
provides a smaller first-run demonstration. This is executable plumbing and
Node behavior evidence, not native compiler correctness.

The expanded coverage proof measures every newly admitted configuration with A,
then runs B on one diagnostic configuration per reason group. It requires a
stdout-only failure and checks raw B output against a same-directory Node
control for exactly one changed byte, unchanged stderr and unchanged exit:

```sh
git show 34420929:stage3/verdict/selection.json > /tmp/verdict-before.json
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/prove_groups.py \
  /tmp/verdict-before.json stage3/verdict/selection.json /absolute/pinned-tree \
  /tmp/new-group-proof > /tmp/group-proof.log 2>&1
```

`audit_cli_mutants.py <manifest> <tree> <new output>` deliberately breaks the
literal-null bridge, declaration exit rule and inferred type-root handling;
each must fail its designated byte comparison on a real pinned case.

Focused checks (set up the pinned external checkout for the resource mutants):

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s stage3/verdict -p 'test_*.py' > /tmp/verdict-checks.log 2>&1
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

# Stage 3 verdict

Run the native compiler as soon as it links:

```sh
stage3/verdict/run.sh --tsc /absolute/native-tsc /tmp/new-verdict > /tmp/verdict.log 2>&1
```

Set `STAGE3_VERDICT_NODE_TSC` to the adapted Node `tsc.js` for the source
tree used to build the candidate, or pass `--node-tsc PATH`. The output
directory must be new. The verdict requires the Linux child supervisor and
runs on Loom's Linux boxes. Other hosts exit with a harness error before
argument parsing. Python 3, git, Node and both compilers' standard
libraries are required. `--tsc` takes one executable path, including paths
with spaces. The verdict exits 0 when all selected cases pass, 1 for
case failures, including timeouts, and 2 for harness errors. It attempts all three suites.

The baselines read the pinned TypeScript checkout as apply does: the gate inputs'
`ADAMIC_TYPESCRIPT_SOURCE`, or `STAGE3_VERDICT_UPSTREAM=/absolute/pinned-checkout`
when set. The verdict never clones one, from a mirror or GitHub, so with neither set
the baselines suite is a harness error naming both (exit 2). `cohere/TypeScript` is
Adamic's fork, not the pinned upstream, and can't stand in. HEAD must be
`050880ce59e30b356b686bd3144efe24f875ebc8`. Source and reference hashes are checked
against the committed census before each selected baseline case is run.

## Suites

| Suite | Comparison |
|---|---|
| acceptance | All 301 existing driver projects, including tiny |
| tiny | The same three-file tiny project, run again |
| baselines | Selected upstream compiler and conformance cases |

## What the verdict compares

For each case, the candidate and the live Node reference start in separate,
fresh project directories populated with identical source files and configs.
Each invocation receives the same project arguments. Source directives control
the run. `@noEmit: true` adds `--noEmit` for both sides. With incremental or
composite enabled, the project may add exactly the build-info file, at the
configured `tsBuildInfoFile` or TypeScript's derived path. File directives
resolve against each project's config directory, including `${configDir}`
substitution, before the verdict derives the build-info name. The `rootDir`
calculation compares directory components without regard to case and preserves
the spelling of the emitted suffix. For incremental no-emit cases the two
invocations use the same working path in sequence, with each tree archived
separately. Each compiler runs beneath a child subreaper. The
supervisor terminates and reaps remaining descendants before inventory or path
reuse, including after timeout. It checks adopted child identity and parentage
before sending a signal through a Linux pidfd. If that cleanup cannot complete, the verdict
reports a harness error. The command deadline starts before supervisor
launch, with up to five additional seconds for cleanup. If the supervisor
misses the outer deadline, the verdict fails the invocation and preserves
its scratch path. Before starting the supervisor, the detached cleanup
owner stages a receipt in the case side directory. It records its PID,
start time (field 22 of `/proc/<pid>/stat`), the inode of
`/proc/self/ns/pid`, the boot ID from `/proc/sys/kernel/random/boot_id` and
its acknowledgment time in Unix nanoseconds. The owner checks the command
deadline and publishes with a hard link to `actual.owner.json` in the same
directory and filesystem. At the command deadline, the CLI tries to create
that path exclusively. The first operation to claim the path wins.
If the CLI wins, the receipt says `expired` and records `expired_at` in Unix
nanoseconds. A late owner removes its staged file and exits 3 without
starting a supervisor. If the owner wins, the CLI allows the existing
cleanup grace period.

The owner resumes a stopped supervisor by PID and waits for it to reap its
descendants. An owner's `released` receipt means it verified the
supervisor's cleanup and closed both capture descriptors. A status-read
fault still triggers a supervisor wait. A failed status read or unverified
supervisor exit leaves an `uncertain` receipt. When the CLI reaches its
deadline it fails the invocation. For a pending acknowledgment it reports
the time and PID with the words "liveness not proven". Time alone never
releases a quarantine.

To record a proven release by hand, use
`python3 -B stage3/verdict/run.py release PATH/actual.owner.json`.
An `expired` receipt has no owner. The command reports that there is nothing
to release. For an owner receipt, the reader requires all identity fields
and checks them in this order:

1. Read the local boot ID. If it cannot be read, refuse.
2. If the boot differs, the former owner is gone and hand release can proceed.
3. Within the same boot, refuse if the reader's PID namespace cannot be read
   or differs from the recorded inode.
4. Within the same namespace, open a kernel process handle for the PID.
   A missing PID permits release. Otherwise, require procfs to show that
   handle under the same PID and require the handle to remain live through
   the stat read. If either check fails, refuse. Then compare start times.
   A matching start time blocks release and a changed start time permits it.
   If opening the handle fails with `EINVAL` or `ENOENT`, the reader
   verifies its own procfs view with its own handle before reading
   `/proc/<pid>/status`. A `Tgid` different from the recorded PID proves
   that the PID now names another process's non-leader thread. An
   unreadable status, missing or invalid `Tgid`, unproven view, or `Tgid`
   equal to the PID refuses release.
   Those two errors only trigger the status check, which decides whether
   release can proceed, and any other handle-open error except a missing PID
   makes release refuse.

   Every cleanup owner must open a pidfd on itself and verify its own
   start time before publishing. It stops with exit 1 before publication
   if the host has no pidfds (`ENOSYS`) or the process reaches its
   descriptor limit (`EMFILE`). The diagnostic identifies the cause.

The command honors an existing release after checking its scope. A hand release records
`released_by: hand` and `released_at` in Unix nanoseconds. For an uncertain
receipt it also records `released_from: uncertain` and prints that the
supervisor's subtree cleanup was never verified. The command sends no signals and leaves the case files and receipt in place. The verdict
keeps the quarantined project at its original path and never reuses it.
The verdict checks each tree against its own derived build-info
path, then compares project-relative output paths and their bytes exactly. JavaScript, declarations and all other changes to
the project tree fail the case. Without incremental or composite, the tree
must stay unchanged. Other cases run with emission enabled.
Emission options such as `declaration`, `sourceMap`, `outDir` and `outFile`
reach both compilers through the project config. Unsupported or ambiguous
directives in selected cases raise an error, with the directive named.
Output path validation uses TypeScript's slash normalization. Rooted paths
and parent path components are rejected before either invocation.

The verdict compares stdout, stderr, exit status, and all files under the
project root by relative path. Unchanged inputs match on both sides, and
changes to them are visible. Added and missing files fail the case. A content
difference reports the path, zero-based offset and a short hex window. All
extensions participate, including JavaScript, declarations, maps and
build-info. It compares file bytes as written. Stdout strips only the
equivalent absolute project prefix on each side before comparison. Raw
captures, argv, the launch directory and both project trees remain in each
case directory.
Empty directories, file modes and upstream emitted baseline transcripts
are outside this comparison.

`TSC_JOBS` defaults to 4. `TSC_TIMEOUT` defaults to 180 seconds per invocation.
A timeout fails only its case. The verdict names the case and the side
(`native` for the candidate, `Node` for the reference) and counts it
separately from mismatches. No comparisons use partial output from a
timed-out invocation. A Node timeout skips the candidate for that case.
The remaining cases still run.

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
Both runners use upstream's `skipDefaultLibCheck: true` and
`noErrorTruncation: true`, overridden by headers; `types: []` prevents ambient
packages in the real filesystem from contaminating the virtual-host equivalent.
`ignoreDeprecations: "6.0"` enables the older test targets. `--pretty false` selects plain CLI diagnostics. Libraries come from the supplied
binary's installation; missing or wrong libraries are observable failures.

For acceptance and tiny, the candidate also has to match the committed
stdout, stderr and exit goldens exactly. For upstream baselines, it also has
to match the diagnostic summary from the pinned `.errors.txt` file (empty
when absent), empty stderr, and the historical CLI exit rule. The upstream
summary reader strips CRLF and stops before annotated source. Both
comparisons apply to emit-enabled and no-emit cases. A legacy mismatch
appears under `legacy` in the case's differences even when both live
compilers agree. The upstream emit transcripts remain the domain of
`stage3/oracle`.

The acceptance chooser `choose()` and the upstream census call
`materialize()` with the original eligibility rules. Command construction
calls `materialize_command()` to read the extra emit directives separately.
Main's former acceptance writer placed `path` before `source_sha256`, so
its JSON did not reproduce the reviewed file byte for byte. The writer now
matches the reviewed key order. `TestVerdictAcceptanceRegenerationPython`
and `TestVerdictUpstreamRegenerationPython` independently regenerate their
reviewed selections. Each requires the pinned checkout, compares every
byte, logs its SHA-256, fails on a skip, and has a 30-second unit budget.

The first version excludes virtual `@filename` units, all external module
specifiers and triple-slash references, option variants, unsupported directives
(including symlinks, currentDirectory, baselineFile and captureSuggestions),
cases previously excluded by the diagnostic census for declaration emit, syntax-error inputs, non-file/global diagnostic summaries,
diagnostics outside the single source (including standard-library `(--,--)`
placeholders), TS18027 emit-resolver errors, non-UTF8
sources. These exclusions avoid claiming CLI
faithfulness for the API/virtual-host scenarios not implemented here. Annotated sources, types, symbols, suggestions and traces from the upstream harness remain outside the verdict.
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
ADAMIC_TYPESCRIPT_SOURCE=/absolute/pinned-tree PYTHONDONTWRITEBYTECODE=1 \
  python3 stage3/verdict/test_verdict.py > /tmp/verdict-checks.log 2>&1
```

These independently mutate stdout, stderr, exit and emitted files, and exercise EOF differences and the timeout path. Census mutants change the pin, selected/excluded counts and source uniqueness;
each is rejected. Artifact-pin mutants append a byte to a source and a baseline
in a disposable checkout, prove the SHA256 guard rejects them, and restore them:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/audit_mutants.py /absolute/disposable-pinned-tree > /tmp/pin-mutants.log 2>&1
```

Use a disposable checkout for these mutants. The input pin audit validates sources and baselines before launching any compiler.
No Adamic oracle fixtures are added, so its counts table is unchanged.

The real-case stand-ins in `mutants/` run through `run.sh`:

| Stand-in | Acceptance verdict case | Failing check |
|---|---|---|
| `extra-file.sh` | `001_varianceCantBeStrictWhileStructureIsnt` | Extra `unexpected.js` |
| `missing-file.sh` | `001_varianceCantBeStrictWhileStructureIsnt` | Missing emitted JavaScript |
| `javascript-byte.sh` | `001_varianceCantBeStrictWhileStructureIsnt` | Changed byte in emitted JavaScript |
| `no-emit-write.sh` | `012_expandoFunctionExpressionsWithDynamicNames2` | Extra `unexpected.js` under `@noEmit` |
| `timeout.sh` | `001_varianceCantBeStrictWhileStructureIsnt` | Native timeout at 180 seconds |
| `node-timeout.js` | `001_varianceCantBeStrictWhileStructureIsnt` | Node timeout at 180 seconds |

Run the shell stand-ins as the candidate with `--baseline-limit 2` for a
short upstream smoke sample. For `node-timeout.js`, pass it as
`--node-tsc` and use the normal Node stand-in as the candidate. The runner supplies the suite, case ID and side to the stand-ins.
They target only the specified acceptance case on the intended side,
even when baselines contain the same source filename. Both timeout
stand-ins sleep for 195 seconds in their target case and forward all others.

The targeted selector proof uses the real acceptance target, its committed
same-name baseline row and an unaffected acceptance neighbor:

```sh
TSC_TIMEOUT=12 PYTHONDONTWRITEBYTECODE=1 \
  python3 stage3/verdict/mutants/prove-selection.py /tmp/new-selection-proof
```

The pinned selection has no declaration-enabled or incremental case.
The synthetic proof checks a changed byte in `out/emit.d.ts` in
`synthetic-declaration` and in `out/tsconfig.tsbuildinfo` in
`synthetic-incremental`:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 stage3/verdict/mutants/prove.py /tmp/new-emit-proof
```

The synthetic no-emit incremental proof uses the pinned stock Node CLI. It
checks the one-file output inventory, then rejects an extra JavaScript file
and a missing build-info file:

```sh
STAGE3_VERDICT_NODE_TSC=/absolute/pinned-tree/lib/tsc.js PYTHONDONTWRITEBYTECODE=1 \
  python3 stage3/verdict/mutants/prove-noemit.py /tmp/new-noemit-proof
```

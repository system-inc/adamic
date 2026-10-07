# Host fixtures from the adapted compiler

No unexplained loader diagnostic or silent native miscompile was observed.
All 25 Node stdout/stderr/exit observations remain identical to the previous
status.json. Classification: **a=1, b=23, c=1**. Final stage 0 outcomes:
1 Checker, 8 Refused, 15 NotYet, 1 Compiles.

## Source and loader pins

- Delivery base: origin/area/stage3, `03ccf222dfeedef4bfe5cc3e219e4586d14a5c13`.
- TypeScript source: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Actual source: `bash stage3/apply.sh /workspace/scratch/host-adapted-tree`.
  All 20 adaptation directories applied successfully (78 files changed).
- Loader: newest fetched origin/area/library,
  `53f44e05c3f4efd163d055c9502dbc32ac717605`, which includes the
  `library/qualified-as-const` QualifiedName fix at `8280fd` and current fs work.
- Scratch integration merge: `5497727222e4f388c98509fc72ef1b4322e17c63`.
  This merge is not part of the delivery branch and is not pushed.
- Stock checker: TypeScript 6.0.3 and @types/node 25.3.3 from stage3/api's lockfile.

Each source header gives the pristine tsc line, the adapted line, and every
adaptation that changed that copied declaration. Generated hostErrors.ts is
explicitly identified as adaptation 47's source. source-spans.json records the
canonical declaration tokens and their provenance. Intermediate apply trees
were checked against every patch-set.md diff count; the final tree's blobs were
checked against the generated files before extracting spans. Source bodies,
including source-owned assertions and any types, are copied as adapted.
Static node imports, scratch setup, enum-value scaffolding and input/output
drivers remain explicit reduction scaffolding. The audit rejects casts,
non-null assertions and any keywords introduced outside copied declarations.

## Per-fixture changes and comparison

All rows also receive precise span/adaptation headers. “Unchanged” means copied
code was re-extracted from the adapted tree and its tokens did not change;
indentation, line endings and header order are not behavior changes.
Category b names the actual Adamic refusal or unimplemented lowering rule;
these are not loader checker errors. Stock tsc accepts every b and c row.

| Fixture | Source change and reason | Class | Stage 0 diagnostic (stock accepts unless stated) |
| --- | --- | --- | --- |
| 01_readFile_utf8.a | Adapted readFile indexed reads (32) and failure handling (47); driver iterates cases directly instead of adding cases[i]!. | b | stage3/fixtures/host/01_readFile_utf8.a:20:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 02_readFile_utf16le.a | Same adapted readFile and driver correction as 01; preserves the UTF-16LE decoder. | b | stage3/fixtures/host/02_readFile_utf16le.a:20:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 03_readFile_utf16be.a | Same adapted readFile and driver correction as 01; preserves the UTF-16BE byte swaps. | b | stage3/fixtures/host/03_readFile_utf16be.a:20:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 04_readFile_missing.a | Adapted readFile indexed reads (32) and failure handling (47); missing-file driver unchanged. | b | stage3/fixtures/host/04_readFile_missing.a:20:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 05_writeFile.a | writeFile statements unchanged. Removed the reduction-only ErrnoException cast from the driver; copied adaptation 47 errorCode to print the failure code. | b | stage3/fixtures/host/05_writeFile.a:35:66: Adamic 0.1 refuses in; an object's shape is known; use a discriminant, or a Map |
| 06_fileExists.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/06_fileExists.a:53:5: stage 0 can't lower node:fs.symlinkSync yet |
| 07_directoryExists.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/07_directoryExists.a:53:5: stage 0 can't lower node:fs.symlinkSync yet |
| 08_getDirectories.a | Function bodies unchanged. Restored ensureTrailingDirectorySeparator overloads and their adapted Path brand (40); restored exported separator declarations. | b | stage3/fixtures/host/08_getDirectories.a:235:5: stage 0 can't lower node:fs.symlinkSync yet |
| 09_realpath.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/09_realpath.a:32:5: stage 0 can't lower node:fs.symlinkSync yet |
| 10_getModifiedTime.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/10_getModifiedTime.a:36:5: stage 0 can't lower node:fs.symlinkSync yet |
| 11_setModifiedTime.a | Copied implementation unchanged; refreshed source provenance. | c | Compiles; native stdout, stderr and exit match Node byte for byte. |
| 12_deleteFile.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/12_deleteFile.a:22:5: stage 0 can't lower node:fs.symlinkSync yet |
| 13_createDirectory.a | Copied adapted createDirectory, which calls errorCode (47); copied that helper and removed the driver ErrnoException cast. | b | stage3/fixtures/host/13_createDirectory.a:72:66: Adamic 0.1 refuses in; an object's shape is known; use a discriminant, or a Map |
| 14_getCurrentDirectory.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/14_getCurrentDirectory.a:17:24: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 15_getExecutingFilePath.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/15_getExecutingFilePath.a:23:21: stage 0 can't lower node:path.basename yet |
| 16_getEnvironmentVariable.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/16_getEnvironmentVariable.a:12:12: stage 0 can't lower a BinaryExpression with a string and a string yet |
| 17_write.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/17_write.a:17:5: stage 0 can't lower a literal method call with an unrepresented result yet |
| 18_exit_0.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/18_exit_0.a:14:5: stage 0 can't lower a value of type undefined yet |
| 19_exit_1.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/19_exit_1.a:14:5: stage 0 can't lower a value of type undefined yet |
| 20_exit_2.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/20_exit_2.a:14:5: stage 0 can't lower a value of type undefined yet |
| 21_createHash.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/21_createHash.a:16:18: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 22_createHash_fallback.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/22_createHash_fallback.a:19:7: stage 0 can't lower a value of type undefined yet |
| 23_newLine.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/23_newLine.a:11:31: stage 0 can't lower reading _os yet |
| 24_useCaseSensitiveFileNames.a | Copied implementation unchanged; refreshed source provenance. | b | stage3/fixtures/host/24_useCaseSensitiveFileNames.a:28:16: stage 0 can't lower statSync: fs options other than a fixed literal or its plain const binding yet |
| 25_readDirectory.a | Copied adapted generic indexed reads (30), addRange access (64), SortedReadonlyArray/toOffset brands (40), restored source overloads and their Path/PathPathComponents brands (40). Gave the reduced Debug object an explicit type derived from its source assertions to retain the namespace assertion target type without a cast. | a | stage3/fixtures/host/25_readDirectory.a:864:10: error TS7030: Not all code paths return a value. Stock tsc: stage3/fixtures/host/25_readDirectory.a:864:10: error TS7030: Not all code paths return a value. |

Fixture 25's former TS2345 T | undefined to T failures disappear when the
adapted indexed reads and source overloads are retained. Its remaining TS7030
is faithful: both compilers enforce noImplicitReturns on the unchanged
tryGetExtensionFromPath at fixture line 864. No return was added to hide it.
The non-null assertions still refused in 01-04, 14 and 21 come from the source,
not from reductions. The errorCode in-operator refusal in 05 and 13 is the
adapted host error helper; it is not a missing Node declaration.

## Verification

Commands run with output redirected to log files:

```sh
bash cloud/setup.sh > /tmp/host-adapted-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /workspace/scratch/host-adapted-tree > /tmp/host-adapted-apply.log 2>&1
# In the scratch stage3 plus library merge:
go test ./stage3/fixtures -run '^TestFixtures/host$' -count=1 -timeout 30m -parallel 4 -v -args -fixtures /workspace/scratch/host-adapted/fixtures > /tmp/host-adapted-runner.log 2>&1
# In the delivery checkout:
python3 stage3/fixtures/host/check.py --compiler-repo /workspace/host-adapted-loader --logs /tmp/host-adapted-cli-check > /tmp/host-adapted-cli-check.log 2>&1
python3 stage3/fixtures/host/check.py --mutants-only --logs /tmp/host-adapted-mutants > /tmp/host-adapted-mutants.log 2>&1
node stage3/fixtures/host/stock-check.cjs /workspace/host-adapted-loader /tmp/host-adapted-stock-final > /tmp/host-adapted-stock-final.log 2>&1
node stage3/fixtures/host/stock-check.cjs /workspace/host-adapted-loader /tmp/host-adapted-stock-final --mutant > /tmp/host-adapted-stock-mutant.log 2>&1
NODE_PATH=stage3/api/node_modules node stage3/fixtures/host/source-audit.cjs /workspace/scratch/host-adapted-tree > /tmp/host-adapted-source-audit.log 2>&1
```

The shared runner passed all 25 fixtures (26.194s). Its native oracle hook
passed fixture 11 with byte comparison, address/undefined-behavior sanitizers
and the Linux leak check. The source audit passed 164 copied declarations and
126 function/method spans. All 25 semantic Node mutants in check.py were caught
by changed stdout or exit status; the exact compiler-record comparison catches
an invented diagnostic. The source audit catches both SHA256 changed to SHA1
and a reduction-added ErrnoException cast. The stock checker mutant assigns
_os.EOL to number and must report TS2322 rather than accept fixture 23.
Historical validation.md lists each semantic mutant; check.py contains the
current adapted-source mutation strings.

Stock options match the loader: strict, noUncheckedIndexedAccess,
exactOptionalPropertyTypes, noImplicitReturns, noFallthroughCasesInSwitch,
erasableSyntaxOnly, verbatimModuleSyntax, allowImportingTsExtensions, noEmit,
ESNext modules, forced module detection, Bundler resolution and ES2024 lib.
The loader's Console and Process prelude merges are reproduced to avoid
conflicting global declarations; Node declarations themselves are stock.
stock-check.cjs writes the options and every diagnostic to stock.json.

Setup: nproc=5, CPU quota=4; Node ready 0.079s, Go 0.084s, submodules 0.299s,
clang 0.635s, markdown dependencies 1.751s, Go build 68.616s, warm cache
68.809s, done 68.955s. Node 24.19.0, Go 1.27.1, clang 20.1.8.

## Limits and scratch integration

The scratch merge needed conflict resolution in compiler/runtime and oracle
files. The stage3 string view implementation was retained and a library UTF-16
view bridge added. Library oracle test inputLeaks callers were wrapped in the
new preparation callback API to build the host hook; those unrelated tests
were not executed. These scratch compatibility changes are not delivered.
No full repository gate or full TypeScript suite was run: the delivered change
is fixture extraction, not a source adaptation. Windows, active profiling and
the omitted host surfaces listed in README.md remain untested. Exit fixtures
retain the source's inactive profiler arm; the active arm remains explicitly
out of scope. Fixture 25 retains reduced Debug namespace scaffolding, with its
source-derived assertion target type made explicit.

@system_adamic_library: these adapted host fixtures preserve all Node goldens;
setModifiedTime now has native evidence, and every remaining refusal is named.

## Fixture 08 filesystem stat follow-up

Merged codex/stage3-any-sys-stat at
06971f0574b6cd408c6291973c685941ea6457ee into this branch with a merge commit,
without rebasing. The fixture branch started at origin's 21ef072e. Re-extracted
getAccessibleFileSystemEntries from /tmp/sys-stat-lane/adapted-tree, the actual
full adaptation output proved in that branch. The only copied source change
is line 199's `any` to `import("fs").Stats | import("fs").Dirent | undefined`.
The provenance header and only fixture 08's source-spans.json record are updated.
Its per-fixture stage3 pin supersedes the manifest's historical shared pin.

All 18 copied declarations in 08 match the adapted source canonically. Node
v24.19.0 retains stdout `a,link,z` twice, an empty line, then `a,link,z`; stderr
is empty and exit is 0. Before/after stdout and stderr compare byte for byte,
and the observation matches the unchanged status.json golden. A mutant returning
files instead of directories prints `file` in place of `a,link,z` and is caught
by stdout comparison. A changed copied declaration is caught by the canonical
source comparison. See 08-stat-proof.json.

Reproduce from the Adamic root with stock TypeScript 6.0.3 on NODE_PATH:

```sh
node stage3/fixtures/host/reextract-08.cjs . <adapted-tree> 06971f0574b6cd408c6291973c685941ea6457ee > /tmp/sys-stat-fixture-extract.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs stage3/fixtures/host/08_getDirectories.a > /tmp/sys-stat-fixture-after.stdout 2> /tmp/sys-stat-fixture-after.stderr
```

The README's full source-audit command was also run, logged at
/tmp/sys-stat-host-full-audit.log, and exits 1 solely at fixture 25's copied
getAccessibleFileSystemEntries: that fixture still carries the old any at
line 1238. Re-extraction of 25 is outside this unit's fixture-08 territory.
This is a reported incomplete full-bucket audit, not a passing full-bucket claim.
The scoped fixture-08 audit passes. Native/library outcomes and the full Go
gate were not rerun; the existing stage 0 status remains its historical record.

# Host fixtures from the adapted compiler

The October 7 fixture 25 follow-up below supersedes the historical checker
options and outcomes in this initial report.

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

## Fixture 25 and checker policy follow-up, October 7

Re-extracted getAccessibleFileSystemEntries from the same actual adapted tree
as fixture 08, pinned at 06971f0574b6cd408c6291973c685941ea6457ee.
Fixture 25 line 1238 now carries Stats | Dirent | undefined. Only the local
annotation and its provenance header changed. All 99 copied declarations in
25 and the full bucket's 164 declarations / 126 function and method spans
match the adapted tree. Node stdout/stderr/exit are byte-identical before and
after extraction, and all 25 recorded Node goldens remain unchanged.

Main's docs/0.1.md October 7 ruling removes noImplicitReturns and
noFallthroughCasesInSwitch. The shared runner inherits internal/load options;
the host CLI harness now supplies a Go overlay for older host-capable library
compilers, dropping both flags and matching main's erasableSyntaxOnly false.
Stock checker options match this policy. Production compiler files are not
edited. Main 48c05d091f0a43c31cbe051b1d6578d99eeedf19's
TestAdamicOptionsAreOn passes. STYLE-OPTIONS.md inventories the remaining live
adaptation/triage flags and census option inheritance; none was changed.

The README's scratch library integration uses fetched area/library
c13622a02363dc577508c0e5a49b774aba714e8b merged with fixture branch
17c5385aff74448c124b8cf245b0337b771e8da3 using --no-commit -X ours.
Its pinned cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db.
This scratch integration is not delivered or pushed. Current results:
**0 Checker, 14 Refused, 11 NotYet, 0 Compiles**. Stock accepts all 25.
Fixture 25 is Refused at its source-owned non-null assertion, line 522.

An independent all-25 run under this SAME library compiler's original options
isolates the policy change: only 25 moves, TS7030 -> Refused. Compared with the
historical ledger, outcome changes also occur in 06, 07, 08, 10 and 24
(NotYet -> Refused), and 11 (Compiles -> NotYet at utimesSync). Diagnostics
change without an outcome change in 09, 12, 13, 15, 16, 17 and 23.
These additional movements are compiler-pin/integration changes, not effects
of dropping style options. Every old/new diagnostic is in 25-followup-proof.json.
No current fixture reaches native execution; historical fixture 11's native,
sanitizer and leak evidence is retained but not claimed as a current pass.

Commands (Node 24.19.0 first on PATH), each logged:

```sh
source /workspace/adamic-tools/env.sh
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/fixtures/host/reextract-08.cjs . /tmp/sys-stat-lane/adapted-tree 06971f0574b6cd408c6291973c685941ea6457ee 25_readDirectory.a > /tmp/host-followup-extract.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/fixtures/host/source-audit.cjs /tmp/sys-stat-lane/adapted-tree > /tmp/host-followup-audit.log 2>&1
python3 stage3/fixtures/host/check.py --compiler-repo /workspace/host-followup-library --record stage3/fixtures/host/status.json --mutants --logs /tmp/host-followup-library-check > /tmp/host-followup-library-check.log 2>&1
node stage3/fixtures/host/stock-check.cjs /workspace/host-followup-library /tmp/host-followup-stock-final > /tmp/host-followup-stock-final.log 2>&1
node stage3/fixtures/host/stock-check.cjs /workspace/host-followup-library /tmp/host-followup-stock-final --mutant > /tmp/host-followup-stock-final-mutant.log 2>&1
# In the independent current-main checkout:
go test ./internal/load -run '^TestAdamicOptionsAreOn$' -count=1 > /tmp/host-followup-options.log 2>&1
```

All 25 semantic Node mutants and the invented diagnostic were caught. Source
audit mutants reject a reduction-added cast and SHA256 -> SHA1. Extraction
rejects a changed canonical declaration and is idempotent. The stock type
mutant produces TS2322; restoring both style flags produces TS7030 in stock
and the real stage 0 build. Evidence logs and the same-compiler original-option
ledger are in followup-evidence. No full repository gate, full tsc suite,
Windows run or current native/sanitizer/leak check was run.

Setup on the reference main checkout used GOPROXY=https://proxy.golang.org|direct
before cloud/setup.sh: Go ready 0.016s, Node 0.017s, markdown 0.060s,
clang 0.162s, submodules 327.151s, Go build 577.798s, tests deferred 577.910s,
cache warm 577.912s, total 577.955s; nproc=5, CPU quota=4.

## Phantom brands and array predicate re-extraction

Merged codex/stage3-any-brands-void at
2bcff092155fc19c6519b27891e53163d6191ae7 with a merge commit,
bb89005ad09d72cc5474df5e60c30ac7b141f2cd, retaining both parents.
The documentation conflicts retain the earlier filesystem stat proof and the
new brand/isArray proof. Combined adaptation 40 progress is 70 removed,
140 remaining; each branch's 69/141 observation remains historical.

A fresh `bash stage3/apply.sh /tmp/host-brands-adapted-tree` completed with
exit 0. The re-extractor inspects every recorded non-partial declaration in
all 25 host fixtures against this actual composed tree, plans all changes
before writing, and permits only the reviewed annotation changes. Exactly
two fixture sources change:

| Fixture | Re-extracted declarations and reason |
| --- | --- |
| 08_getDirectories.a | Path.__pathBrand: undefined -> void, following the phantom-brand ruling. |
| 25_readDirectory.a | Path.__pathBrand, PathPathComponents.__pathComponensBrand and SortedReadonlyArray's literal `" __sortedArrayBrand"`: undefined -> void; isArray's input: any -> unknown, preserving its predicate and body. |

The existing stat union is preserved in both fixtures. Fixture 25's isArray
provenance header now identifies adaptation 40; both affected manifest rows
carry the full composed merge SHA and refreshed canonical tokens. No other
fixture source or manifest row changes. A second extraction reports no edits.

All 25 Node stdout/stderr/exit observations match both the pre-extraction
source and the existing goldens, using Node 24.19.0 first on PATH. The final
re-extracted 08 and 25 were independently compared again. The full audit
passes 164 copied declarations and 126 function/method spans. Actual full-audit
mutants restoring 08's stale undefined brand and 25's stale any input are
rejected as missing or changed copied declarations. The existing cast and
SHA256 -> SHA1 audit mutants are also caught. All 25 semantic Node mutants
and the invented recorded diagnostic are caught. Stock tsc accepts all 25
fixtures. See brands-followup-proof.json and brands-followup-evidence/.

Reproduce from the fixture checkout after sourcing the setup environment:

```sh
bash stage3/apply.sh <fresh-tree> > apply.log 2>&1
NODE_PATH=<stock-typescript-6.0.3-node_modules> node stage3/fixtures/host/reextract-brands.cjs <fresh-tree> bb89005ad09d72cc5474df5e60c30ac7b141f2cd > extraction.log 2>&1
NODE_PATH=<stock-typescript-6.0.3-node_modules> node stage3/fixtures/host/source-audit.cjs <fresh-tree> > audit.log 2>&1
python3 stage3/fixtures/host/check.py --mutants-only --logs <mutant-logs> > mutants.log 2>&1
node stage3/fixtures/host/stock-check.cjs /workspace/host-followup-library <stock-logs> > stock.log 2>&1
```

Setup initially failed with `error: expected submodule path 'cohere' not to be
a symbolic link`. The unsupported scratch link was removed and setup rerun
in /workspace/adamic, whose dependency checkout has the identical pin.
With GOPROXY=https://proxy.golang.org|direct: Go 0.018s, Node 0.021s,
markdown 0.067s, submodules 0.090s, clang 0.196s, build 32.549s,
tests deferred 32.661s, cache warm 32.662s, total 32.690s; nproc=5,
CPU quota=4. The printed /workspace/adamic-tools/env.sh was sourced.

The existing stage 0 ledger is unchanged; stage 0/native, sanitizers, leak
checks, the full Go gate and full upstream oracle/lane were not rerun for
this fixture extraction. The adaptation branch's own proof and gate artifacts
arrive through the merge and are not presented as new executions here.

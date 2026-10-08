# Host stage0 rerecord with the pinned Node declarations

Loader bugs / unexplained checker differences: **0**. The six rejected checker
results match stock tsc exactly, including location, code, message, elaboration
and sorted order. The other nineteen pass both checkers and stop in lowering.
Counts: **a = 6, b = 19, c = 0**. All 25 stage0 records change. Node stdout,
stderr and exit observations, provenance, platform metadata and every byte
outside the stage0 JSON values remain unchanged.

## Source and classification

- Stage3 base: `origin/area/stage3`, `f85306ea8913a08324daaca3a5429747fe094a5c`.
- Loader: `origin/library/qualified-as-const`, `8280fd0873a3d5c425f8a1a342b44af20cf883d9`.
- Detached scratch merge: `84de0d4f14a6299ef91032785810933959d2cdf4`.
  This merge is not pushed or included in the rerecord branch.
- Official packages installed by `npm ci --prefix stage3/api`: TypeScript 6.0.3,
  @types/node 25.3.3. Node runtime: 24.19.0. Linux, five visible processors,
  four-CPU cgroup quota.

The first scratch merge used area/library and hit independent additions to
counts.md. A narrowly checked resolution preserved both sets of rows. It was
superseded before observations were recorded: the final merge uses the requested
QualifiedName fix and merges cleanly. Audits of the superseded tree were stopped.

Case a means the complete checker diagnostic set is identical to stock tsc.
Case b means stock accepts while stage0 does not compile. In this run **all 19
b cases are implementation gaps, not checker-option strictness**: load.Load
accepts each file, then `internal/lower/library_node.go:nodeLibraryRefusal`
returns a named NotYet because `node:fs.mkdtempSync` is not registered as an
implemented runtime member. This is the documented Node-member support rule,
not a loader diagnostic discrepancy or a permanent language prohibition. There
is no compiler option that can enable this missing implementation. Every driver
creates its scratch directory with mkdtempSync, so that first blocker masks later
host lowering gaps. Case c would require native stdout/stderr/exit to equal Node
byte for byte, with the runner's sanitizer and leak checks. No case reaches c.

## Per-fixture table

| Fixture | Case | Stock tsc | Stage0 | Diagnostic or named rule |
| --- | --- | --- | --- | --- |
| 01_readFile_utf8.a | a | 2 diagnostics: TS2322 | Checker | Exact diagnostics below; stock and stage0 agree. |
| 02_readFile_utf16le.a | a | 2 diagnostics: TS2322 | Checker | Exact diagnostics below; stock and stage0 agree. |
| 03_readFile_utf16be.a | a | 2 diagnostics: TS2322 | Checker | Exact diagnostics below; stock and stage0 agree. |
| 04_readFile_missing.a | a | 2 diagnostics: TS2322 | Checker | Exact diagnostics below; stock and stage0 agree. |
| 05_writeFile.a | b | Accepts, 0 diagnostics | NotYet | `05_writeFile.a:9:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 06_fileExists.a | b | Accepts, 0 diagnostics | NotYet | `06_fileExists.a:12:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 07_directoryExists.a | b | Accepts, 0 diagnostics | NotYet | `07_directoryExists.a:12:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 08_getDirectories.a | b | Accepts, 0 diagnostics | NotYet | `08_getDirectories.a:22:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 09_realpath.a | b | Accepts, 0 diagnostics | NotYet | `09_realpath.a:10:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 10_getModifiedTime.a | b | Accepts, 0 diagnostics | NotYet | `10_getModifiedTime.a:10:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 11_setModifiedTime.a | b | Accepts, 0 diagnostics | NotYet | `11_setModifiedTime.a:8:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 12_deleteFile.a | b | Accepts, 0 diagnostics | NotYet | `12_deleteFile.a:8:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 13_createDirectory.a | a | 1 diagnostics: TS18046 | Checker | Exact diagnostics below; stock and stage0 agree. |
| 14_getCurrentDirectory.a | b | Accepts, 0 diagnostics | NotYet | `14_getCurrentDirectory.a:9:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 15_getExecutingFilePath.a | b | Accepts, 0 diagnostics | NotYet | `15_getExecutingFilePath.a:9:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 16_getEnvironmentVariable.a | b | Accepts, 0 diagnostics | NotYet | `16_getEnvironmentVariable.a:8:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 17_write.a | b | Accepts, 0 diagnostics | NotYet | `17_write.a:8:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 18_exit_0.a | b | Accepts, 0 diagnostics | NotYet | `18_exit_0.a:9:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 19_exit_1.a | b | Accepts, 0 diagnostics | NotYet | `19_exit_1.a:9:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 20_exit_2.a | b | Accepts, 0 diagnostics | NotYet | `20_exit_2.a:9:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 21_createHash.a | b | Accepts, 0 diagnostics | NotYet | `21_createHash.a:10:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 22_createHash_fallback.a | b | Accepts, 0 diagnostics | NotYet | `22_createHash_fallback.a:9:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 23_newLine.a | b | Accepts, 0 diagnostics | NotYet | `23_newLine.a:8:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 24_useCaseSensitiveFileNames.a | b | Accepts, 0 diagnostics | NotYet | `24_useCaseSensitiveFileNames.a:15:17: stage 0 can't lower node:fs.mkdtempSync yet`; rule `nodeLibraryRefusal`. |
| 25_readDirectory.a | a | 15 diagnostics: TS2322, TS2345, TS2532, TS2775, TS7030 | Checker | Exact diagnostics below; stock and stage0 agree. |

## Exact checker diagnostics for case a

Both stock tsc and stage0 produce each block below. These are genuine stock
errors under Adamic's unchanged options, not extra Adamic strictness.
Fixtures 01-04 index Buffer under `noUncheckedIndexedAccess`; 13 accesses an
unknown catch binding under `strict` / `useUnknownInCatchVariables`; 25 includes
unchecked indexes, an unannotated assertion-function target (TS2775), and
`noImplicitReturns` (TS7030). Full content is also retained in status.json.

### 01_readFile_utf8.a

```text
stage3/fixtures/host/01_readFile_utf8.a:25:13: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
stage3/fixtures/host/01_readFile_utf8.a:26:13: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
```

### 02_readFile_utf16le.a

```text
stage3/fixtures/host/02_readFile_utf16le.a:25:13: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
stage3/fixtures/host/02_readFile_utf16le.a:26:13: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
```

### 03_readFile_utf16be.a

```text
stage3/fixtures/host/03_readFile_utf16be.a:25:13: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
stage3/fixtures/host/03_readFile_utf16be.a:26:13: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
```

### 04_readFile_missing.a

```text
stage3/fixtures/host/04_readFile_missing.a:25:13: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
stage3/fixtures/host/04_readFile_missing.a:26:13: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
```

### 13_createDirectory.a

```text
stage3/fixtures/host/13_createDirectory.a:54:17: error TS18046: 'e' is of type 'unknown'.
```

### 25_readDirectory.a

```text
stage3/fixtures/host/25_readDirectory.a:1001:23: error TS2345: Argument of type 'T | undefined' is not assignable to parameter of type 'T'.
  'T' could be instantiated with an arbitrary type which could be unrelated to 'T | undefined'.
stage3/fixtures/host/25_readDirectory.a:1041:17: error TS2532: Object is possibly 'undefined'.
stage3/fixtures/host/25_readDirectory.a:1046:21: error TS2532: Object is possibly 'undefined'.
stage3/fixtures/host/25_readDirectory.a:408:27: error TS2345: Argument of type 'T | undefined' is not assignable to parameter of type 'T'.
  'T' could be instantiated with an arbitrary type which could be unrelated to 'T | undefined'.
stage3/fixtures/host/25_readDirectory.a:430:21: error TS2345: Argument of type 'T | undefined' is not assignable to parameter of type 'T'.
  'T' could be instantiated with an arbitrary type which could be unrelated to 'T | undefined'.
stage3/fixtures/host/25_readDirectory.a:449:29: error TS2345: Argument of type 'T | undefined' is not assignable to parameter of type 'T'.
  'T' could be instantiated with an arbitrary type which could be unrelated to 'T | undefined'.
stage3/fixtures/host/25_readDirectory.a:513:31: error TS2345: Argument of type 'T | undefined' is not assignable to parameter of type 'T'.
  'T' could be instantiated with an arbitrary type which could be unrelated to 'T | undefined'.
stage3/fixtures/host/25_readDirectory.a:543:5: error TS2322: Type '(string | undefined)[]' is not assignable to type 'string[]'.
  Type 'string | undefined' is not assignable to type 'string'.
    Type 'undefined' is not assignable to type 'string'.
stage3/fixtures/host/25_readDirectory.a:587:5: error TS2775: Assertions require every name in the call target to be declared with an explicit type annotation.
stage3/fixtures/host/25_readDirectory.a:588:5: error TS2322: Type 'T | undefined' is not assignable to type 'T'.
  'T' could be instantiated with an arbitrary type which could be unrelated to 'T | undefined'.
stage3/fixtures/host/25_readDirectory.a:613:54: error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
stage3/fixtures/host/25_readDirectory.a:709:34: error TS2345: Argument of type 'T | undefined' is not assignable to parameter of type 'T'.
  'T' could be instantiated with an arbitrary type which could be unrelated to 'T | undefined'.
stage3/fixtures/host/25_readDirectory.a:741:10: error TS7030: Not all code paths return a value.
stage3/fixtures/host/25_readDirectory.a:851:27: error TS2345: Argument of type 'T | undefined' is not assignable to parameter of type 'T'.
  'T' could be instantiated with an arbitrary type which could be unrelated to 'T | undefined'.
stage3/fixtures/host/25_readDirectory.a:879:31: error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

## Commands and observations

All test output went to log files. Setup: `bash cloud/setup.sh` completed in
63s: Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache
warm 63s. The scratch submodules were initialized recursively; the attempted
local reference failed because the reference repository was shallow, so normal
shallow initialization was used successfully.

```sh
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api > /tmp/host-rerecord-npm.log 2>&1
go run ./stage3/rerecord-tool > /workspace/scratch/host-rerecord/stage0.json 2> /tmp/host-rerecord-stage0.stderr
node /workspace/scratch/host-rerecord/stock.cjs > /tmp/host-rerecord-stock.log 2>&1
go test ./stage3/fixtures -run '^TestFixtures/host$' -count=1 -timeout 30m -parallel 4 -v > /tmp/host-rerecord-runner-before.log 2>&1
go test ./stage3/fixtures -run '^TestFixtures/host$' -count=1 -timeout 30m -parallel 4 -v > /tmp/host-rerecord-runner-after.log 2>&1
```

The temporary Go collector calls load.Load and then lower.Lower exactly as the
shared runner does, classifies typed CheckError/Refused/NotYet errors, and removes
only the scratch repository path prefix from diagnostic text. It does not use
the CLI prefix, trailing newline or go-run exit-status suffix. The collector
and stock comparison script exist only in scratch, not the pushed branch.

Before updating, the shared runner exits 1: all 25 Node comparisons pass and
all 25 stage0 comparisons detect the changed records (20.695s). After replacing
only stage0 values, it exits 0: all 25 Node and 25 stage0 comparisons pass
(13.517s). There are six Checker records and nineteen NotYet records. The
independent collector and stock comparison find six exact diagnostic matches,
nineteen stock accepts, zero unexplained checker differences. Native execution
and the full repository gate were not run because no host fixture compiles and
only status/documentation files change.

### Stock tsc configuration and source view

The stock check uses the official pinned TypeScript compiler API, createProgram
and getPreEmitDiagnostics, once per unchanged fixture. No fixture statements or
declaration package bytes are edited. As in Adamic's loader, the host exposes
X.a as virtual X.a.ts with identical text, includes the official pinned
@types/node/index.d.ts explicitly, and includes Adamic's prelude with only its
console declaration replaced by `declare var console: Console`, exactly as
nodePrelude does. It uses stock ES2024 declarations; Adamic's RegExp library
corrections are not applied to stock. All resulting checker diagnostics still
match, so there is no hidden RegExp strictness difference in these six cases.
Paths are returned to their original .a spelling, messages are flattened with
newlines, and the resulting diagnostic strings are sorted as load does.

Options, taken from the merged internal/load/load.go rather than inferred from
TypeScript defaults:

```json
{
  "strict": true,
  "noUncheckedIndexedAccess": true,
  "exactOptionalPropertyTypes": true,
  "noImplicitReturns": true,
  "noFallthroughCasesInSwitch": true,
  "erasableSyntaxOnly": true,
  "verbatimModuleSyntax": true,
  "allowImportingTsExtensions": true,
  "noEmit": true,
  "module": "esnext",
  "moduleDetection": "force",
  "moduleResolution": "bundler",
  "target": "es2024",
  "lib": ["es2024"],
  "types": []
}
```

### Mutants and preservation proof

One real-source mutant appends `const rerecordMutant: number = _os.EOL;` to a
scratch copy of 23_newLine.a. Stock tsc detects TS2322 at 20:7, string is not
assignable to number. The shared runner on that same mutant exits 1: its Node
byte comparison passes, and only its stage0 comparison fails because it sees
Checker/TS2322 instead of the recorded NotYet/mkdtempSync. This independently
proves both checks can catch a changed source type contract without relying on
a Node crash or a native build error.

```sh
node /workspace/scratch/host-rerecord/stock.cjs --mutant > /tmp/host-rerecord-stock-mutant.log 2>&1
go test ./stage3/fixtures -run '^TestFixtures/host/23_newLine.a$' -count=1 -timeout 30m -v -args -fixtures /workspace/scratch/host-rerecord/runner-mutant > /tmp/host-rerecord-runner-mutant.log 2>&1
```

A separate byte check replaces each stage0 JSON span with the same placeholder
in the old and new files and asserts that the remaining complete file bytes
are identical. It also compares every parsed field other than stage0 and proves
exactly 25 stage0 values changed. No Node observation is rerecorded or weakened.

## Refresh on library/merge-p2b, October 7, 2026

Baseline `047e857207be`, cohere `7945d102`, Linux, Node v24.19.0. The same
scratch typed collector workflow above calls load.Load and lower.Lower, then
replaces only stage0 JSON spans. Twenty stage0 values change. A byte comparison
masks just those spans and proves that every remaining byte, including each
Node field, is unchanged. No fixture source is edited.

The current outcomes are six Checker, fifteen NotYet, three Refused and one
Compiles. 05_writeFile.a now refuses adamic/no-unchecked-cast at 37:30;
14_getCurrentDirectory.a and 21_createHash.a refuse non-null assertions.
The remaining gaps name the next actual unsupported member, not mkdtempSync.
11_setModifiedTime.a now compiles and its Node/native/sanitizer/leak comparison
passes. 25_readDirectory.a's current complete checker diagnostics are recorded.

`go test ./stage3/... -count=1 -v -timeout 30m` passes, including all 25 exact
host Node observations and current stage0 outcomes (fixtures package 301.361s).
Log: `/tmp/p2b-stage3.log`. An inaccurate 05_writeFile.a diagnostic in a scratch
status file fails only stage0; its Node comparison passes, proving that the
rerecorded compiler contract remains strict (`/tmp/p2b-status-mutant.log`).

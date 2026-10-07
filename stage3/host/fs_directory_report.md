Built: merged main and upstream fixtures, added path.relative, and fixed native realPath field specialization that corrupted source enumeration.
Commits: 5fbaa4d main merge, aa987b3 fixture merge, 0ad74bf relative, 97b9f64 acceptance gate, 8f6a0b2 runtime layout fix; all pushed.
Checks: independent loader, native package, load/lower and vet pass; all five owned fixtures match recorded Node output but stop at Checker on both backends.
Mutants: five upstream source mutants and two compiled component mutants are caught only by Node stdout comparison; all three independent loader mutants are caught by Go cohere comparison.
Uncovered: shared @types/node loader hook and sibling host dependencies are pending; no owned upstream acceptance fixture is green yet.

## Upstream acceptance fixtures on current main

Run from the repository root after sourcing /workspace/adamic-tools/env.sh:

```sh
python3 stage3/host/check_fs_directory.py --mutants --logs /tmp/fs-directory-acceptance > /tmp/fs-directory-acceptance.log 2>&1
```

This runner preserves stage3/fixtures/host/status.json and all five upstream sources. It compares source Node observations to the recorded bytes, independently attempts native and JavaScript compilation, and requires both compiled observations to match. Native runs enable ASan/UBSan and leak detection. Checker, NotYet, Refused and unexpected tool failure are distinct outcomes. The runner exits 1 until every owned fixture agrees on both backends; it never accepts the historical Checker status as success.

| Fixture | Source on Node | Native stage | JavaScript stage |
| --- | --- | --- | --- |
| 07_directoryExists | Agrees with status.json | Checker | Checker |
| 08_getDirectories | Agrees with status.json | Checker | Checker |
| 09_realpath | Agrees with status.json | Checker | Checker |
| 24_useCaseSensitiveFileNames | Agrees with status.json | Checker | Checker |
| 25_readDirectory | Agrees with status.json | Checker | Checker |

Exact stdout, stderr, exit codes and results.json are committed under logs/fs_directory_acceptance/. The five source mutants from the supplied check.py each exit cleanly and differ only in stdout: directoryExists uses isFile; directory sort reverses; failed realpath resolves the input; case sensitivity inverts its probe; readDirectory bypasses extension filtering. These are fixture sensitivity evidence, not evidence of compiled acceptance mutants, since declaration loading is still blocked.

Both merges were explicitly requested and preserve history. The main conflict in internal/lower/object.go was resolved by retaining Node host dispatch and main's userMethodCall dispatch. No main-side functionality was discarded. fs_file's inspected branch tip 080789f still used unit-local declaration copies; it was not merged in place of the corrected shared hook. No shared hook SHA has yet been supplied.

## Independent language blocker in readDirectory

Current compilation of fixture 25 reports TS2345 for bounds-guarded generic indexed reads, TS2322/TS2532 for potentially absent array values, TS2775 for the unannotated Debug assertion object, and TS7030 for an implicit undefined return. The missing Node imports are additional blockers. The source audit's upstream function bodies are unchanged.

One-line reproducer, checked independently of Node declarations:

```typescript
export function map<T, U>(xs: readonly T[], f: (x: T) => U): U[] { const out: U[] = []; for (let i = 0; i < xs.length; i++) out.push(f(xs[i])); return out; }
```

`go run ./cmd/adamic types /workspace/scratch/fs-directory/index-proof.a` reports TS2345: T | undefined is not assignable to T. This needs the coordinated sound indexed-read adaptation/proof support; suppressing the diagnostic or weakening NoUncheckedIndexedAccess would not be a sound fix. The diagnostic is saved in logs/fs_directory_index-proof.log.

## Relative-path component

The actual readDirectory driver calls node:path.relative, so this unit now builds it as well as resolve, dirname and join. Its fixture node_path_relative.a covers 144 pairs, including empty segments, roots, divergent name prefixes, dot components, Unicode and NUL. TestNodePathRelativeRuntime constructs the equivalent host-call IR and holds both emitters to the unlowered source on Node. It deliberately bypasses the unavailable declaration loader and therefore does not claim that node_path_relative.a compiles through the checker or lowering. Register its normal input/counts gate after the shared declarations land.

```sh
go test ./internal/oracle -run TestNodePathRelativeRuntime -count=1 -timeout 15m -v > /tmp/fs-directory-relative-runtime.log 2>&1
```

Observed: PASS, 47.755s. Both emitters agree, sanitizer and leak checks pass, and the native join mutant runs cleanly and is caught only by stdout comparison with Node. POSIX relative compares whole normalized components and preserves Node's equal-input shortcut before cwd resolution.

Setup on the merged main: Go go1.27.1 ready 0s, clang 20.1.8 ready 0s, Node v24.19.0 ready 0s, submodules ready 0s, build cache warm 321s, total 321s. nproc reports 5, cgroup cpu.max is 400000 100000. The printed environment is /workspace/adamic-tools/env.sh.

## Current-main source enumeration and field proof

The initial independent run failed only for native: at line 895 it printed a project reference where Go cohere expected cohere/TypeScript/packages/typescript/test/async/api.bench.ts. Source on Node and the JavaScript backend did not report a discrepancy. The initial native package gate also failed TestRuntimeFieldLayoutsAreIncluded because directory.c's runtime path field was absent from main's uniformFieldOffsets proof.

Main added uniform field specialization after base 035999e. Its proof did not know the C realPath result shape (kind at slot 0, path at slot 1). Program object literals could put path at slot 0. Omitting the runtime shape made that field appear uniformly at slot 0, so native read the runtime discriminator Ok as a resolved path, causing unrelated directories to collide in the traversal's visited map. Commit 8f6a0b2 includes that runtime shape, Error's code slot and Dirent's name/type shape in the existing proof. This is an additive proof correction, not a change to the realpath walk.

The minimal node_fs_directory_layout.a fixture combines an own slot-0 path with realPath('.'). It compiles through the checker and lowering using the existing adamic declarations, and agrees on both backends. Its native mutant restores the erroneous slot-0 read, exits cleanly under sanitizers and leak checks, and fails only stdout comparison with Node. Observed TestNodeFSDirectoryRuntimeLayouts PASS, 0.815s; TestRuntimeFieldLayoutsAreIncluded PASS, 26 runtime layouts checked.

The independent loader then passed over 143 settings, 130 tsconfigs and 1354 source files on native, source Node and JavaScript. Its three cleanly executing source mutants were caught by comparison with Go cohere on both native and Node:

| Mutant | Observed comparison failure |
| --- | --- |
| strict JSON comments | A commented config was accepted where Go rejected JSON |
| inherited rule options | Inherited option array became empty |
| tsconfig excludes | Excluded src/b.ts entered the source list |

Commands and final observations:

| Command | Result |
| --- | --- |
| go test ./stage1/cohere/config -run TestLoadersMatchGoCohere -count=1 -timeout 30m -v | PASS, 186.757s, all three mutants caught |
| go test ./internal/native -count=1 -timeout 30m | PASS, 125.509s |
| go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh -count=1 -timeout 30m | load PASS 7.962s; lower PASS 62.624s; JavaScript has no tests; initial native proof failure fixed and retested as above; flow/fresh fail because their fixture-wide scans cannot load pending node:* declarations |
| go test ./internal/oracle -run 'TestInputAgreesWithNode/internal/oracle/testdata/^realpath[.]a$' -count=1 -timeout 15m -v | PASS, 39.899s, original realpath fixture on both backends plus leaks |
| go vet ./... | PASS, empty output |
| gofmt -l cmd internal; git diff --check | Empty output |

Each command's output is committed under logs/. The initial failed independent-loader log is preserved alongside the passing rerun. The earlier unanchored realpath filter also selected the declaration-blocked host fixture and failed; the exact anchored rerun above is the passing original-port evidence.

The verbatim independent-loader failure on the old 035999e base has not yet been supplied. fields.go did not exist at that base. The observed merged-main failure is fixed and independently validated, but this does not establish the cause of a different old-base failure. The full repository gate and fixture counts were not rerun: declaration-dependent fixtures still fail loading. Add the relative and runtime-layout fixtures to the normal input/counts gate and regenerate the complete ledger when the shared hook lands. No temporary declaration copies or independent loader hook were reintroduced.

Earlier sections below are historical, pre-merge evidence.

---

Built: runtime and lowering retained; local Node declarations and this unit's loader hook removed per the corrected integration brief.
Commits: correction follows eef83ac, afe01f1 and f09e4c3; see branch history for the correction SHA.
Checks: load and lower package tests pass after removal; host fixture checks now need the shared fs_file declaration hook.
Mutants: the earlier 19 passing Node-comparison mutants are historical evidence, not a claim of validation with the pending pinned declarations.
Uncovered: @types/node 25.3.3 integration and named-refusal fixture validation await the shared hook SHA and forwarded TypeScript fixture branch.

## Corrected integration contract

sys.ts uses static node:* imports from the TypeScript patch set. Later literal require calls will use the same import route. No require implementation is provided here. fs_file owns the shared loader for @types/node 25.3.3; this branch has no node_<unit>.d.ts copies or independent loader hook. Merge its hook commit when the user sends the SHA, then rerun the host fixtures and mutant gate against those declarations.

Lowering follows resolved symbols in node/fs.d.ts and node/path.d.ts, including the fs/path ambient declarations re-exported by node:fs/node:path. It distinguishes Dirent predicates from Stats predicates and realpathSync.native from unrelated native names. Unimplemented fs/path calls reach a named NotYet fallback after builtin dispatch, allowing the other host units to implement their census members. Unsupported values and non-census overloads are refused rather than silently interpreted.

Error fixtures now use the pinned declarations' NodeJS.ErrnoException structural view instead of globally adding code to Error. The refreshed census branch fetched during this correction was 429c1177f0130f785c19cf590d1860513b2ddbfc; stage3/api was absent at that tip. Host fixture compilation is intentionally blocked pending the shared hook, with its diagnostic captured in logs/fs_directory_correction-hook-blocked.log.

The earlier report and logs below describe the pre-correction declaration setup. They remain useful runtime evidence, but their successful oracle and counts results must be rerun after the shared hook lands. New package evidence is in logs/fs_directory_correction-packages.log.

---

Built: catchable Node fs directory/realpath calls, POSIX path.resolve/dirname/join, and the System directory adapter.
Commit: eef83ac19215884ab83872f96510d17a2c03d44a, based on 035999ea5a55bafd69f88e9445cf488fdbfc92c8.
Checks: five dual-backend fixtures, permission probe, 19 mutants, touched packages, vet, formatting and counts passed.
Mutants: all 19 ran cleanly under ASan/UBSan and leak detection; only stdout comparison with Node caught them.
Uncovered: compiler matchFiles integration and native tsc proof, fs_file statSync replacement, macOS and Windows execution.

## Delivered behavior

The declarations merge into node:fs and node:path beside the prelude load. Lowering, JavaScript emission, native runtime and fixtures implement the census calls. The path census uses only resolve, dirname and join; normalize, relative, basename, extname, isAbsolute, sep and delimiter are not exposed by this unit.

readdirSync returns strings or Dirent objects, orders raw names as Node does on Linux, and preserves the three requested predicates. Both realpathSync functions throw catchable errors with code and message; the walk reports its failing component and native uses libc realpath on the original argument. NUL paths receive Node-style TypeError inspection. Ordinary Error.code remains undefined. Fixtures cover empty paths, missing entries, ENOTDIR, EACCES, dangling links, cycles, symlink/.. traversal, Unicode, dot components and trailing slashes.

stage3/host/fs_directory.a supplies getDirectories, getAccessibleFileSystemEntries, directoryExists, realpath, resolvePath, useCaseSensitiveFileNames and readDirectory. createDirectorySystem receives the executing filename, current-directory provider and compiler's own matchFiles. Its fixture verifies all nine forwarded arguments, omitted options and the callbacks. It does not prove wildcard matching end to end. The private statusType bridge uses existing fileStatus and is explicitly marked for replacement by fs_file statSync; no second statSync was built. POSIX sys.ts prefers native realpath and returns the input on error. Case sensitivity probes the swapped ASCII-case executing filename against the filesystem.

Optional closure calls now supply typed undefined for omitted parameters. The readDirectory(path) fixture exposed an actual native stack-buffer-overflow before that correction. Error-code layout, throwing-call analysis and ownership handling are wired through the existing mechanisms. None of internal/native/emit.go, internal/lower/lower.go, internal/native/native.go or internal/oracle/oracle_test.go was edited.

## Commands and observations

All test output was redirected to log files. Source /workspace/adamic-tools/env.sh before Go commands.

| Command | Observed result |
| --- | --- |
| bash cloud/setup.sh | Go go1.27.1, clang 20.1.8, Node v24.19.0; ready timings 0s each, submodules 0s, cache warm 133s, total 133s; nproc 5, cgroup quota 4 CPUs |
| go vet ./... | exit 0, empty log |
| go test ./internal/load ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh -count=1 -timeout 30m | load 1.706s; lower 40.987s; native 189.551s; flow 130.180s; fresh 70.592s; JavaScript has no package tests |
| go test ./internal/oracle -run 'TestNodeFSDirectory|TestInputAgreesWithNode/internal/oracle/testdata/node_' -count=1 -timeout 30m -v | PASS, 50.740s; five fixtures both backends, permission probe, 19 mutants |
| go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts | PASS, 101.307s |
| go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m | PASS, 16.295s |
| gofmt -l cmd internal; git diff --check | empty output |

The full uncached repository gate was not run. Raw final logs are in logs/fs_directory_*.log. The permission test drops root to uid/gid 65534 and requires Node to observe EACCES before comparison.

## Mutant evidence

Every row below was executed by TestNodeFSDirectoryMutants. Every mutated binary compiled, exited zero with empty stderr under ASan/UBSan and leak checks, and was rejected only because its stdout differed from Node.

| Mutant | Deliberate wrong behavior |
| --- | --- |
| resolve | Use join |
| join | Use resolve |
| dirname | Use join on the input |
| readdir order | Reverse listing |
| Dirent name | Replace names with wrong |
| isFile | Use isDirectory |
| isDirectory | Use isFile |
| isSymbolicLink | Use isFile |
| realpath walk | Use native |
| realpath native | Use walk |
| error code | Replace code with WRONG |
| error message | Replace message with wrong |
| getAccessibleFileSystemEntries | Return empty lists |
| getDirectories | Return empty list |
| directoryExists | Return false |
| realpath | Return input |
| resolvePath | Return input |
| isFileSystemCaseSensitive | Return false |
| readDirectory | Swap extensions and excludes passed to matchFiles |

## Remaining integration and platform limits

The actual compiler matchFiles and native tiny-project tsc --noEmit diagnostic comparison remain integration work. require routing belongs to the separately coordinated patch set. The fs_file unit was absent from this base, so its shared statSync must replace the marked bridge when available.

Only census string-path UTF-8 calls and requested Dirent methods are declared; Buffer/URL paths, other encodings and unrelated fs/path exports are outside this implementation. Resource-exhaustion/race errno paths and deleted-current-directory failures have implementation support but were not fault-injected. Linux is the tested gate. macOS was not run; case sensitivity depends on the mounted filesystem and filename normalization can differ. Windows path semantics are deliberately outside this POSIX unit.

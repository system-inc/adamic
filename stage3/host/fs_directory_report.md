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

Restored cmd/adamic runtime products and replaced their Git-dependent ownership census with a source-tree walk.
Base 062ba8b7ec31ca855248511f80614d99c3a915a5; restores 04d0a5b3/e54dc0a7; submission SHA is in the final worker report.
All 21 cmd tests passed in build and read modes; ordered traces, no-Git census, dependency census, reader guard (39.908 s) and vet passed.
All ten mutants were caught, including second-builder, traversal exclusions and reintroduced Git dependency; actual catchers are recorded below.
Not covered: full packages/full gate, macOS, remote publication, or a complete Loom assembled-source run; the explicit no-Git proof uses PATH without git.

Loom's chg_9tneegyn failure was a test harness assumption: git ls-files requires a checkout, while a runner has an assembled tree. runtimeBuilderPackages now delegates to a filepath.WalkDir traversal from the repository root. No Git command or executable is used to enumerate sources. The remaining os/exec use in this file belongs to the unchanged missing-owner child-product guard.

The walk skips these directories before descending:

* Any directory named testdata, vendor, node_modules or .git, at any depth.
* Root .gitignore directory entries: .cache/ and .idea/, at any depth because both are unanchored basename rules. There are no other build/cache directory rules in current main's root .gitignore. Ignored file entries such as /adamic, /go.work.sum, *.log and editor files do not name directories and are not directory exclusions.
* The cohere submodule path read from root .gitmodules. The old git listing did not enumerate its contents; they are another repository's dependencies, not possible Adamic runtime product owners. No code was copied from it.

Root directory rules are read on every census, not copied into the test. The matcher handles basename globs, root-anchored directory paths and slash-separated directory paths, exercised by a small synthetic tree. Directory negations and ** are refused explicitly if introduced, rather than silently interpreting them incorrectly. WalkDir does not follow directory symlinks. Conservative assumption: nonignored ordinary Go files on disk count even if a developer has not tracked them yet; this both works in assembled source and refuses a newly introduced producer. No build directory or machine path is embedded in product/test declarations.

TestCommandRuntimeCensusWalksSourceTree puts deliberately malformed Go decoys in all excluded categories, and valid aliased producers in ordinary, .cache-source and nested/build directories. /build/ must exclude only root build, not nested/build; a Find caller must remain a reader. Removing an exclusion or dropping an ordinary producer must fail this test. It adds no Adamic fixture and requires no counts.md change.

The complete local tree census read 1,354 Go files. TestCommandRuntimeArchivesHaveOneBuilder passed in 5.96 s including its three runtime-request/key setups. The synthetic traversal witness passed in 0.00 s. no-git.py verified shutil.which('git', path=PATH) is None and ran the real ownership test with that PATH: exit 0, 5.94 s test time, 5.954 s complete invocation. The PATH retains clang/llvm-ar and Go tools but no Git. no-git.json and no-git.log preserve the exact environment selector and output. This uses the user's allowed PATH alternative, not a claim that a full checkout-free tree was run locally.

Archive behavior is unchanged from -3. commandRuntimeInputs still uses current runtimeproducts.Inputs; shared products still call BuiltElsewhere after exactly two owner waits. All three requests use feature set []. Oracle alone builds sanitized and plain WASI, and cmd/adamic alone builds counted WASI. Request affects linking only. No runtime C, existing behavior assertion, runtime feature set, or compiler API changed. Main's product-order.json is retained apart from the same two waits:

* cmd/adamic TestProduct_CommandRuntimeSanitized after internal/oracle TestProduct_NativeRuntimesSanitized0.
* cmd/adamic TestProduct_CommandRuntimeWASI after internal/oracle TestProduct_NativeRuntimesWASI0.

Old and new builders, relative to -3: unchanged. Base sanitized 3ac8c29ccf42 and base WASI 6e08d061bac9 belong to internal/oracle; base counted WASI 2b3ce1fb7e30 belongs to cmd/adamic. These are observations under the ordinary setup PATH; the restricted no-Git PATH run reports sanitized f22bfa733ca5, still with oracle as sole builder (we do not treat different environment selectors as identical inputs). Keys are computed from real inputs, never hardcoded. The original pre-revert duplicate cmd builders remain removed. A future six-group runtime shard layout keeps the base set in each variant's group zero, so neither cmd owner wait changes.

previous-REPORT.md and previous-after.json retain -3's full historical cold/warm table and observations. Those October 14/17 numbers are history, not fresh before/after measurements for this fix. The earlier same-run comparison remains 185.523 -> 43.673 behavior-test core-seconds, saving 141.849; no new causal speedup is claimed here.

Final fresh verification: all 21 cmd/adamic top-level tests were selected individually in both build and read-only modes, 42 passing invocations. Each invocation has timeout 60 and -test.timeout=55s; request/header setup is included. No whole-package test command was used. Largest cmd invocation: 15.414 s. Timings for every test are in validation.json and build-/read- logs. Restored and new runtime tests:

| Test | Build s | Read-only s |
|---|---:|---:|
| TestCommandRuntimeArchivesHaveOneBuilder | 15.414 | 7.662 |
| TestCommandRuntimeProductsWaitForOwners | 0.057 | 0.016 |
| TestSharedCommandRuntimeCannotBuildItsOwnersArchive | 3.713 | 3.877 |
| TestCommandRuntimeCensusWalksSourceTree | 0.012 | 0.292 |
| TestProduct_CommandRuntimeSanitized | 2.368 | 4.780 |
| TestProduct_CommandRuntimeWASI | 2.231 | 2.364 |
| TestProduct_CommandRuntimeWASICounted | 2.150 | 2.477 |
| TestCommandRuntimeProductKeyHoldsEveryPart | 1.949 | 1.831 |

Ordered real tracing used fresh products, oracle owners first. All five trace invocations passed below 60 seconds with zero refusals:

| Product | Total traced s | Result |
|---|---:|---|
| TestProduct_NativeRuntimesSanitized0 | 41.347 | four settled |
| TestProduct_NativeRuntimesWASI0 | 44.404 | four settled |
| TestProduct_CommandRuntimeSanitized | 24.817 | found, zero settled |
| TestProduct_CommandRuntimeWASI | 20.604 | found, zero settled |
| TestProduct_CommandRuntimeWASICounted | 13.648 | one settled |

check-traces.py passed: nine observed runtime keys have one builder each; shared cmd finders read exactly the oracle settlement keys. trace-ownership.log includes their full keys. The six incidental owner-group archives remain oracle-only before and after this change: 19877fc5e50b, 74c7e3895326, 9c3a08bf55d2, 7f0d666b1361, 21bf4c5826f5, a954e124b033.

All ten independently applied mutants failed for the intended test assertion or parser refusal, not a compiler warning. Source files were restored in finally, and the final binary and vet were run on restored sources. Patches are noncompilable .patch evidence.

| Mutant | Catcher / observed failure |
|---|---|
| second-owner | TestCommandRuntimeArchivesHaveOneBuilder: two producer packages [internal/oracle cmd/adamic] for the actual base sanitized name key |
| second-builder | TestSharedCommandRuntimeCannotBuildItsOwnersArchive: child builds successfully instead of failing for the missing oracle owner |
| missing-wait | TestCommandRuntimeProductsWaitForOwners: missing sanitized owner wait |
| missing-key-part | TestCommandRuntimeProductKeyHoldsEveryPart: adamic-runtime-v6 absent from identity |
| undeclared-sanitized | unchanged TestExplainChecksOutput/overload_some_empty/build: no runtime product declared |
| ignored-directory | TestCommandRuntimeCensusWalksSourceTree: ignored .cache/decoy.go is incorrectly parsed |
| fixture-directory | TestCommandRuntimeCensusWalksSourceTree: testdata/decoy.go is incorrectly parsed |
| submodule-directory | TestCommandRuntimeCensusWalksSourceTree: cohere/decoy.go is incorrectly parsed |
| missing-producer | TestCommandRuntimeCensusWalksSourceTree: ordinary producer absent from actual package list |
| git-required | real TestCommandRuntimeArchivesHaveOneBuilder with no Git PATH: exec: "git": executable file not found in $PATH |

Commands and outputs (all test output goes directly to logs):

* timeout 60 git fetch origin; git checkout -b compiler/loom-cmd-archives-4 origin/main, base 062ba8b7ec31ca855248511f80614d99c3a915a5. A final precommit fetch still names that tip. Read CLAUDE.md, docs/front-door.md, root .gitignore/.gitmodules, the restored test files and current runtimeproducts/Find/BuiltElsewhere. Binding language/memory docs and product order are unchanged from the version already read for -3. Only the same two product waits are added to main's file.
* export GOPROXY='https://proxy.golang.org|direct'; timeout 240 bash cloud/setup.sh --wasi-sdk; source /workspace/adamic-tools/env.sh. Passed, no workaround. nproc 5; cpu.max 400000 100000, four CPU share. Timing lines: Node 0.050 s, Go 0.058 s, submodules 0.155 s, markdown 0.206 s, clang 0.491 s, WASI 0.787 s, shared cache 3.571 s, go build 32.709 s, cache warm 32.946 s, done 33.043 s. Node 24.19.0, Go 1.27.1, clang 20.1.8. Full setup.log retained.
* timeout 120 go test -c -o /tmp/adamic-recut4.test ./cmd/adamic (initial compile timeout 180); timeout 120 go test -c -o /tmp/oracle4.test ./internal/oracle; timeout 90 go build -o /tmp/traced4 ./internal/buildcache/cmd/traced. Passed. Cmd binary rebuilt after mutants on final restored sources. Go compilation/toolchain installation is outside measured test leaves; per-test setup is included.
* timeout 60 /tmp/adamic-recut4.test -test.run='^TestCommandRuntime(ArchivesHaveOneBuilder|CensusWalksSourceTree)$' -test.v -test.timeout=55s from cmd/adamic: both pass, 5.96 and 0.00 s. Full-tree scan 1,354 Go files. timeout 90 python3 no-git.py: real full-tree ownership test passes with shutil.which git equal to None, 5.94 s test time. no-git.json preserves invocation time and PATH.
* timeout 450 python3 mutants.py: five existing mutants caught. Each selected go test ./cmd/adamic -run SELECTED -count=1 -timeout 60s has outer timeout 90. timeout 480 python3 walk-mutants.py: four traversal mutants caught via selected go test -timeout 55s / timeout 90; the Git-regression mutant is compiled under timeout 120 and run as a binary under timeout 60 / -test.timeout=55s with no Git in PATH. All expected failures and restores succeeded.
* timeout 540 python3 validate.py: the five traced products listed above, then every cmd Test function separately in build and ADAMIC_BUILD_CACHE=read modes. Product trace command is timeout 60 /tmp/traced4 -- BINARY -test.timeout=55s -test.v -test.run=^NAME$, with cwd internal/oracle or cmd/adamic, ADAMIC_ORACLE_WASI=1 and ADAMIC_BUILD_STORE=off. All live trace reports/logs were outside the checkout until settlement completed; no unstable-tree retry was needed. Every 21 + 21 test invocation passed. python3 check-traces.py passed its owner-key and one-builder assertions.
* timeout 90 go test ./internal/compilercensus -run '^TestDependencies(FollowTheMarkers|CoverTheDeclaredMap)$' -count=1 -timeout 60s: passed, 4.921 s, exit 0. This covers main's compiler-dependency census and cmd consumers. No new package needs declaration.
* timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s: passed, 39.908 s, exit 0. timeout 90 go vet ./cmd/adamic on final sources: passed without diagnostics, exit 0. gofmt, source diff checks and main product-order preservation checks passed.

No test leaf exceeded 60 seconds. No full packages or gate were run. No PR or Loom submission was made; the user submits the pushed branch.

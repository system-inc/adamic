Built atomic publication for every custom executable in admission-delta tests, without retries.
Commit: this branch's one commit; base 6dbab365fe696b93fffc681bbe71c63d2f5bdb29.
Stress before: 20 and 100 cohort repetitions plus 1,000 target repetitions all passed; stress after: the same 100 and 1,000 commands passed; all 31 package tests pass in build and read-only modes.
Direct-write mutant fails the new regression with text file busy; serial-lowering mutant fails the retained barrier test with compiler-timeout.
Not covered: the stochastic Loom failure did not reproduce locally; no whole gate or new compiler admission measurement was run. The reader guard passed in 2.47 s after its export product built in 44.03 s.

Changes and assumptions

Only cmd/adamic-admission-delta's tests changed. writeExecutable creates a temporary file in the executable's own directory, writes and chmods it, checks Close, then renames it to the final name. commandFixture's base/head scripts, the dirty-input head replacement, the negative-witness head replacement and the parallel-lowering barrier script all use it. All three shell build modes also write to their unique output's .tmp sibling, finish the redirection, chmod the temporary executable, then mv it into place. The old copy-then-overwrite in commandFixture is removed. There are no retry loops.

The specialized scripts are needed to encode a two-worker barrier, refusal diagnostics, a deliberate native-output mismatch, dirty-source detection, and ruled negative-witness exits. The plain fixture compiler product cannot supply those behaviors. TestBuildRevisions, the test needing only the fixture compiler, continues to find TestProduct_AdmissionCompiler's product. No additional compiler product or product-order change was needed.

The conservative interpretation of shell temporary names is that each generated native output has exactly one writer inside one command invocation; its existing per-program scratch path is unique, so its .tmp sibling needs no shared allocation or retry. The Go helper uses CreateTemp, so replacement does not truncate an existing executable. New TestExecutableReplacement retains a write descriptor to the previous executable, publishes its replacement, and executes it. This models the retained writer that a sibling fork can leave behind and deterministically distinguishes in-place writes from replacement.

Audit

- speed_test.go: barrier executable now uses writeExecutable.
- command_test.go: both fixture compilers and dirty-input replacement now use writeExecutable; generated native mismatch and native agreement scripts use close-and-rename shell publication.
- negative_test.go: replacement head uses writeExecutable; its generated exit-70 native program uses shell publication.
- Other WriteFile calls write JSON, program source, Node modules/declarations or Go adapter/source files, which are read by tools rather than exec'd. buildCompiler's Go binary is still a buildcache product built by Go's linker, not a test's direct executable write. The audit log retains the complete package search.

Toolchain

Ran export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh, then sourced /workspace/adamic-tools/env.sh. Setup succeeded. Timing lines: Node 0.044 s; Go 0.058 s; submodules 0.106 s; markdown skipped step-duration=0.013 s / ready 0.135 s; clang 0.307 s; shared cache 1.374 s. Go 1.27.1, Node 24.19.0, house clang 20.1.8. nproc=5; cgroup cpu.max is 400000 100000, so four CPUs. Tests used GOMAXPROCS=4.

Red before and green after

The reported stochastic race did not reproduce. Before editing, this selector ran at -count=20, then -count=100, both with -parallel 16:

    ^Test(ParallelLoweringOnly|HeadInputIsolation|NewAdmissionMismatch|ManifestMismatch|NewlyRefusedIsInformational|EmptyCorpus|DiffCoverage|ShardCommandScope|NegativeCommandAccepted|NegativeTypeCorrectMutant|NegativeAlreadyAcceptedRepairMutant|NegativeEmptyListPrinted)$

Commands were timeout 90 go test ./cmd/adamic-admission-delta -run '<selector>' -count=20 -parallel 16 -timeout 90s -json, then timeout 360 go test ./cmd/adamic-admission-delta -run '<selector>' -count=100 -parallel 16 -timeout 350s -json. Each output went directly to its stress-before-*.jsonl file. Both exit 0: 240 and 1,200 leaf executions, 19.451 and 85.528 s respectively. While the 100-count cohort ran, timeout 90 go test ./cmd/adamic-admission-delta -run '^TestParallelLoweringOnly$' -count=1000 -parallel 16 -timeout 90s -json also exited 0, 1,000 leaves in 11.064 s. Thus there were 1,120 target-test attempts before the fix, and 2,440 total cohort/target leaf executions, with zero ETXTBSY failures.

After the change, repeated exactly the 100-count cohort command and the 1,000-count target command, together as before. Both exit 0: 1,200 leaves in 161.167 s and 1,000 leaves in 16.545 s. The slower cohort overlapped cold guard compilation and product tracing. Maximum cohort leaf was 4.41 s; maximum standalone target leaf was 0.24 s. No touched or new leaf exceeded 60 s. The initial 20-count baseline was not separately repeated after the fix; the exact raised-count commands were repeated.

Deterministic mutants

1. Replaced writeExecutable with the previous os.WriteFile-to-destination behavior. go test ./cmd/adamic-admission-delta -run '^TestExecutableReplacement$' -count=1 -timeout 60s -v exited 1 in 0.012 s, reporting fork/exec .../compiler: text file busy. The new test was 0.00 s. See direct-write-mutant.patch and .log. This is a deterministic red witness of the old publication behavior, not a claim that the stochastic stress run went red.
2. Changed the barrier test's worker count from 2 to 1. go test ./cmd/adamic-admission-delta -run '^TestParallelLoweringOnly$' -count=1 -timeout 60s -v exited 1 in 1.015 s with compiler-timeout. See serial-lowering-mutant.patch and .log. This proves the custom script still checks concurrent lowering after the publication change.

Both source mutations were restored exactly before final stress and cache verification. The Python mutant drivers imposed 90 s wall limits.

Build and read-only package verification

Built one binary with timeout 90 go test -c -trimpath -buildvcs=false -o /workspace/scratch/admission-busy.test ./cmd/adamic-admission-delta. Used this same binary in every mode, preserving the binary identity in traced product reads.

With GOCACHEPROG unset, TZ=UTC, GOMAXPROCS=4, ADAMIC_BUILD_STORE=off and a dedicated ADAMIC_BUILD_CACHE_DIR:

- timeout 120 go run ./internal/buildcache/cmd/traced -- <binary> -test.count=1 -test.timeout=90s -test.parallel=4 -test.run='^TestProduct_AdmissionCompiler$' -test.v. Exit 0. Producer 2.86 s; 1 settled / 0 refused / 95 reads, key 08d3d91e2125. See product-traced.log and product-trace.jsonl.
- Same command with -test.run='^TestProduct_AdmissionCompilerReady$'. Exit 0; 33.32 s; 0 settled / 0 refused, so the reader did not build a product.
- timeout 90 <binary> -test.count=1 -test.timeout=90s -test.parallel=4 -test.v. Exit 0; all 31 package tests passed. New test 0.02 s; maximum leaf 4.38 s. See build-mode-final.log.
- Same command with ADAMIC_BUILD_CACHE=read GOTOOLCHAIN=local GOFLAGS=-buildvcs=false. Exit 0; all 31 tests passed, including dirty-input isolation and all negative-witness tests. New test 0.00 s; maximum leaf 3.07 s. See read-mode-final.log.

An earlier traced run overlapped still-changing untracked JSON stress evidence. The producer itself passed, but settlement correctly refused the changed tree; its dependent reader and read-only run then failed for lack of a settled product. Those exploratory logs are retained. The final trace ran after stress evidence stopped changing and settled successfully. Only the *-final.log files and product-traced.log / reader-traced.log are final green evidence.

Required checks

The first timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -v command exhausted its wall limit and exited 124 without test output. Separating compilation with GOCACHEPROG unset exposed a cold standard Go cache; timeout 180 go test -c -o /workspace/scratch/admission-busy-ir.test ./internal/ir also exited 124 before producing a binary. The attempted invocation of that absent binary exited 127 and ran no test. These are setup failures, not observed test leaves over 60 s.

Restored the warmed shared Go cache through env.sh, with ADAMIC_GOCACHE_STORE=off. timeout 240 go test -c -o /workspace/scratch/admission-busy-ir.test ./internal/ir exited 0. From internal/ir, ran timeout 120 <ir-binary> -test.run='^TestProduct_CallTargetExports$' -test.count=1 -test.timeout=90s -test.v: exit 0, product 44.03 s (build 42.27 s). Then timeout 120 <ir-binary> -test.run='^TestCallTargetReaders$' -test.count=1 -test.timeout=90s -test.v: exit 0, guard 2.47 s, export product hit 0.44 s. See reader-product.log and readers-verified.log. No completed test leaf exceeded 60 s, and no over-budget leaf was rerun.

timeout 90 go vet ./cmd/adamic-admission-delta exited 0; vet.log is empty. git diff --check passes. No other package's tests were edited. There are no new oracle fixtures, runtime feature sets or counts changes. No cohere source was copied. No PR or Loom submission was made.

Built ForkLock exclusion from executable open through close, retaining rename outside the lock.
Commit: this branch's follow-up commit, parent 8c7b5fe49820d7ccdce8a54382143a45d1a084cd; main base 6dbab365fe696b93fffc681bbe71c63d2f5bdb29.
The locked fork-pressure test passed 100/100 runs; both package cache modes pass 31/31 tests; reader guard and vet pass.
Removing only the lock causes text file busy on run 1, executable index 12, and in 97/100 runs overall.
Not covered: full gate, load-60 scheduling, other operating systems, or new compiler behavior. No new oracle fixture or counts change.

The earlier rename-only claim was incomplete: a child can inherit a writer to the temporary inode before rename. writeExecutable now takes syscall.ForkLock for writing before CreateTemp and releases it after Close. The anonymous closure defers Unlock, so CreateTemp, WriteString, Chmod and Close failures also release it, including t.Fatal's Goexit. The writer is closed before unlock; rename and temporary-file cleanup happen outside the lock. No retries were added.

Replaced TestExecutableReplacement with TestExecutableForkIsolation. Eight goroutines continuously exec the host's true command in the same process while the test publishes and executes 200 scripts in sequence. It resolves true with exec.LookPath instead of embedding a machine path. A 45-second context bounds the work, is checked before every publication, cancels fork workers on every exit and joins them. Each executable has a one-second execution timeout. The replacement models concurrent fork inheritance, not a deliberately retained writer to the old destination.

Kept the already published branch and added a follow-up commit, as explicitly authorized, rather than rewriting it. Fetched origin; main remained 6dbab365. Read CLAUDE.md, docs/front-door.md and the two changed test files before editing. No other package's tests changed.

Setup

Ran export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh; sourced /workspace/adamic-tools/env.sh. Setup succeeded. Node ready 0.074 s, Go ready 0.151 s, markdown skipped step-duration=0.031 s / ready 0.249 s, submodules ready 0.290 s, clang ready 0.488 s, shared cache ready 1.668 s, build cache warm 20.551 s, done 20.593 s. Go 1.27.1, Node 24.19.0, clang 20.1.8. nproc=5, cpu.max=400000 100000; verification used GOMAXPROCS=4 and TZ=UTC. Normal Go invocations used the warmed shared cache with ADAMIC_GOCACHE_STORE=off. Traced verification unset GOCACHEPROG, as in the preceding unit.

Stress and mutant

Initial focused locked test: timeout 90 go test ./cmd/adamic-admission-delta -run '^TestExecutableForkIsolation$' -count=1 -timeout 90s -v, output in focus.log. Exit 0; 1.98 s. The final version also checks its context deadline before each publication.

The mutant removes syscall.ForkLock.Lock, deferred Unlock and the now-unused import, keeping the new test and rename intact. Ran go test ./cmd/adamic-admission-delta -run '^TestExecutableForkIsolation$' -count=100 -timeout 300s -json with a 330-second subprocess wall limit. Output went directly to no-fork-lock.jsonl; source was restored in finally. Exit 1, 100 completed runs, 97 failures, all 97 reporting text file busy; maximum leaf 1.12 s, package 21.971 s. First failure was run 1, executable index 12. See no-fork-lock.patch and mutant-results.json.

After restoring the lock, ran the identical go test command under timeout 330. Output went directly to locked-100.jsonl. Exit 0, 100/100 passed, package 216.106 s, maximum leaf 4.23 s. There were no ETXTBSY failures. This is the same-process fork-pressure comparison requested, not an inference from the old low-pressure stress runs.

Package products and cache modes

Built one final binary: timeout 90 go test -c -trimpath -buildvcs=false -o /workspace/scratch/admission-fork-lock.test ./cmd/adamic-admission-delta. Exit 0. Every following invocation used that same binary.

With ADAMIC_BUILD_STORE=off, a dedicated ADAMIC_BUILD_CACHE_DIR, TZ=UTC, GOMAXPROCS=4 and GOCACHEPROG unset, ran:

- timeout 120 go run ./internal/buildcache/cmd/traced -- <binary> -test.count=1 -test.timeout=90s -test.parallel=4 -test.run='^TestProduct_AdmissionCompiler$' -test.v. Exit 0, producer 1.17 s; 1 settled / 0 refused / 95 reads, key 568fd1f81894. Logs: product-traced.log, product-trace.jsonl.
- Same command with -test.run='^TestProduct_AdmissionCompilerReady$'. Exit 0, reader 11.87 s; 0 settled / 0 refused. Logs: reader-traced.log, reader-trace.jsonl.
- timeout 90 <binary> -test.count=1 -test.timeout=90s -test.parallel=4 -test.v. Exit 0, 31/31 pass, fork-pressure test 2.88 s. Log: build-mode.log.
- Same command with ADAMIC_BUILD_CACHE=read GOTOOLCHAIN=local GOFLAGS=-buildvcs=false. Exit 0, 31/31 pass, fork-pressure test 2.76 s. Log: read-mode.log.

Stress output finished changing before the trace started. All final package leaves stayed below 60 s; results.json contains their individual times. Dirty-input isolation, negative witnesses and the original parallel-lowering barrier remain green.

Required checks

- timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -v. Exit 0; guard 4.67 s, export product hit 0.96 s. Output: readers.log.
- timeout 90 go vet ./cmd/adamic-admission-delta. Exit 0; vet.log is empty.
- gofmt -l cmd/adamic-admission-delta and git diff --check -- cmd/adamic-admission-delta: clean.

No cohere source was copied. No PR or Loom submission was made.

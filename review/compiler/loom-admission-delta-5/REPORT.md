Built the admission fixture compiler as one traced product with an ordered reader.
Commit: this branch's single commit; base 41fa274899cfdc2152f6ae41ffc77ebd748ef066; source 3d79c445.
All 30 package tests pass in build and read modes; traced producer settles 1 product with 95 reads, 0 refused; vet passes.
Seven mutants caught: missing product, policy key, crash classification, output agreement, witness sampling, blob verification, dirty runtime input.
Not covered: full gate, new real compiler admission measurements, a different base compiler revision; the unchanged IR guard passed in 70.60 s and was not rerun.

Applied the complete requested triple-dot diff from origin/compiler/loom-admission-delta-4 to current main, including its historical evidence. The package remains declared under packages because it builds and exercises compiler libraries through its revision adapters, rather than only invoking independently supplied executables.

TestProduct_AdmissionCompiler owns admission-compiler. The revision fixture is now a root Git commit with fixed author, committer, dates and contents. Its exact SHA is 40b25f0411929152605c138cdca8123d8aa0b2d0. The name key is now e27c33959815, rather than the timestamp-dependent key in the red report. The production build body is shared with the producer. Inputs generated for the product are created inside its build interval.

TestProduct_AdmissionCompilerReady finds that product without a build callback and waits for its owner in product-order.json. TestBuildRevisions finds it in the real source tree before invoking the helper in its synthetic repository. It passes both runtime and lowering binaries explicitly. Both revisions are the same fixture SHA, so the conservative choice is one shared product, not a second base product. Arbitrary user-selected revisions remain separate SHA-keyed tool products; this unit does not claim to prebuild arbitrary history.

Git now runs with toolchain.GitEnv, excluding global and system Git configuration. The compiler key drops the GOTOOLCHAIN selection policy and redundant GOFLAGS options that the build's explicit -trimpath and -buildvcs=false already determine. Output-changing flags remain keyed; TestCompilerCacheIgnoresBuildPolicy checks both claims.

Setup: GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. nproc=5; verification GOMAXPROCS=4 and test.parallel=4. Timing lines: node 0.061 s, go 0.195 s, markdown skipped step-duration=0.035 s / ready 0.221 s, submodules 0.250 s, clang 0.668 s, shared cache 2.544 s. Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup succeeded.

Verification commands, all output redirected to evidence files:

- Built one test binary: timeout 90 go test -c -trimpath -buildvcs=false -o /workspace/scratch/admission-delta-5-final.test ./cmd/adamic-admission-delta.
- With TZ=UTC, GOMAXPROCS=4, ADAMIC_BUILD_STORE=off, a dedicated ADAMIC_BUILD_CACHE_DIR and GOCACHEPROG unset: timeout 120 go run ./internal/buildcache/cmd/traced -- <same-test-binary> -test.count=1 -test.timeout=90s -test.parallel=4 -test.run='^TestProduct_AdmissionCompiler$' -test.v. Exit 0, producer 1.04 s, 1 settled / 0 refused / 95 reads. See product-traced.log and product-trace.jsonl.
- After that settlement, ran the identical binary with -test.count=1 -test.timeout=90s -test.parallel=4 -test.v, once normally (final-build.log), and once with ADAMIC_BUILD_CACHE=read GOTOOLCHAIN=local GOFLAGS=-buildvcs=false (final-read.log). Both exit 0, all 30 leaves pass. Read mode: owner 0.94 s, ordered reader 0.10 s, TestBuildRevisions 0.18 s, dirty isolation 0.42 s, new policy test 0.00 s; maximum leaf 1.03 s. Build mode maximum 0.89 s. Full per-leaf times are in leaf-seconds.json.
- Separately traced the same binary with -test.run='^TestProduct_AdmissionCompilerReady$'. Exit 0, 19.92 s; 0 settled / 0 refused, proving it only finds the owner product. See reader-traced.log.
- Empty-cache mutant: identical binary, ADAMIC_BUILD_CACHE=read, a new empty cache, -test.run='^TestBuildRevisions$'. Exit 1 in 0.05 s: admission-compiler is not built, no traced build has recorded what it reads. See final-missing-product-mutant.log. No producer was run against this cache.
- Restored raw GOFLAGS in compilerInputs. TestCompilerCacheIgnoresBuildPolicy failed in 0.00 s because builder and reader keys differed. Restored the exact original source afterward. See build-policy-mutant.patch and .log.
- timeout 420 python3 review/compiler/loom-admission-delta-5/run-mutants.py. Each child go test has -timeout 60s and subprocess timeout 90s. All five retained safety mutants exit 1 at their intended checks: crash-as-refusal/TestClassification; ignore-output/TestOutputsAgree and TestNewAdmissionMismatch; omit-witness/TestSampling; ignore-blob/TestManifestMismatch; dirty-input/TestHeadInputIsolation runtime comparison. Patches and logs are alongside this report.
- timeout 90 go test -count=1 -timeout 90s -run '^TestCallTargetReaders$' -v ./internal/ir. Exit 0, 70.60 s, including 66.50 s building export data. This unchanged guard exceeds the ordinary 60 s leaf budget on this cold local run; it was not rerun or altered. The requested 90 s guard limit held.
- timeout 90 go vet ./cmd/adamic-admission-delta. Exit 0, including after the final test change; vet-final.log is empty.
- Parsed the imported Python manifest generator and validated the product-order edge; git diff --check for cmd and cloud passes.

Exploratory failures are retained. The first trace refused reads of the local shared Go cache and /etc/localtime; disabling GOCACHEPROG for traced verification and setting TZ=UTC fixed these. The next refused ~/.gitconfig; toolchain.GitEnv fixed that. An early read-mode run overlapped the safety mutant harness and is not valid final evidence. Another read-mode attempt rebuilt a differently configured test binary and missed its settings identity. The same-binary check then exposed synthetic-repository read-set resolution; resolving the product before entering the helper fixed that. Only final-read.log, final-build.log, product-traced.log and reader-traced.log are the final green proof.

No new oracle fixture or runtime feature set was added; synthetic temporary corpus inputs do not change internal/oracle/counts.md. No cohere source was copied. No PR or Loom submission was made.

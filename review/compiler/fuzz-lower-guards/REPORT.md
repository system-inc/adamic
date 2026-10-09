Added reducer refusal-identity, exact-signature, byte-sharing and lowered argument-reader guards for task #wdzm8r1.
Implementation commit: 59a11d10edbd979841794a37d348128a68a55020; delivery SHA is reported with the push.
All four focused tests pass before mutation and after restoration; focused neighbor tests and lane checks pass.
Exact M16, M05, M14 and M04 audit diffs fail their intended named assertions in the requested order.
M12 is skipped as equivalent; no new oracle fixture, full package run, full gate or backend run is included.

The M16 guard exercises the real Reduce parser and reduction loop with controlled observations. It starts with originalFailure(); otherFailure(); and supplies the same verdict and checker diagnostics for both refusals. It confirms that the smaller candidate carrying only the different refusal was actually tried. The final source must still match the derived original signature, and the unrelated statement must be removed. With M16, the result is otherFailure(); while Reduction.Signature still says original refusal, and the guard fails on that drift.

Scope choice: use controlled observations to pin the reducer's contract independently of changes to compiler admission. This does not test how the compiler produces refusal diagnostics. The brief assigns internal/fuzz here, so no ownership redirection was needed.

M05 pins exact abc as matching abc and rejecting abc-suffix. M14 pins bytesShared(63, 128) false with a 64-byte positive neighbor. M04 lowers an ordinary function reading arguments.length and requires ArgumentsCount == 1 together with ReadsArguments == true. The mutant reaches the reader-fact assertion while the count assertion still passes.

| Audit diff | Evidence source | Named assertion failure |
| --- | --- | --- |
| M16 | test-audit/internal-fuzz at 53d97313 | original signature reported for different-refusal source |
| M05 | same | exact signature abc matched abc-suffix |
| M14 | same | 63 bytes must not share a 128-byte owner |
| M04 | test-audit/internal-lower-arguments_length at e17cd28b | ArgumentsCount exists but ReadsArguments is false |

The exact diffs are stored as .diff evidence. verify.py applies each diff, requires the named test to fail, reverses the diff in finally, then requires the restored test to pass. Production mutations were never staged or committed. A build failure does not count as catching a mutant.

| New test | Recorded Go test seconds | Restored invocation including setup seconds |
| --- | ---: | ---: |
| TestReduceRejectsDifferentRefusal | 0.00 | 2.08 |
| TestExactSignatureRejectsSuffix | 0.00 | 1.97 |
| TestBytesSharedRejects63Bytes | 0.00 | 1.92 |
| TestArgumentsLengthReadKeepsReaderFact | 0.03 | 2.13 |

The focused final runs pass 23 leaves, with maximum leaf time 0.31s. timings.json records every final leaf. All new leaves and complete focused invocations are below 60 seconds on four CPU equivalents, with GOMAXPROCS=4 and -parallel=4.

Commands:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 180 bash cloud/setup.sh > /tmp/fuzz-guards-setup.txt 2>&1
source /workspace/adamic-tools/env.sh
timeout 600 python3 review/compiler/fuzz-lower-guards/verify.py > /tmp/fuzz-guards-progress.txt 2>&1
timeout 120 env GOMAXPROCS=4 go test ./internal/fuzz -run '^(TestReduceRejectsDifferentRefusal|TestExactSignatureRejectsSuffix|TestBytesSharedRejects63Bytes)$' -count=1 -parallel=4 -timeout 90s -json > review/compiler/fuzz-lower-guards/fuzz-final.jsonl 2>&1
timeout 120 env GOMAXPROCS=4 go test ./internal/lower -run '^TestArgumentsLength' -count=1 -parallel=4 -timeout 90s -json > review/compiler/fuzz-lower-guards/lower-final.jsonl 2>&1
```

results.json records each baseline, mutant and restored command, exit code, elapsed time and named test result. All tests run with -timeout 90s, and the driver supplies an additional subprocess timeout. An initial test compile used a nonexistent Refused verdict; it was corrected to the existing NotYet verdict before any mutant verification.

Setup succeeded: Go ready 0.025s, Node ready 0.030s, submodules and markdown ready 0.085s, clang ready 0.204s, Go build ready 42.536s, build cache warm 42.756s, done 42.789s. nproc is 5, cpu.max 400000 100000 gives four CPU equivalents. setup.txt records the complete output.

The branch began at current main f91994f019703ba25d2918cf529c0e0b0c05d93c. The later main refresh merged its existing test guards without conflicts or changes to internal/fuzz or internal/lower. Own changes are only two _test.go files and review evidence. No new oracle fixture was added, so no counts refresh was needed.

Required lane checks, after committing:

```sh
timeout 60 git fetch -q origin main devtools/fast-gate cloud/merge-tree
timeout 120 bash -c 'git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'
```

Initial lane output: lane checks 1.7 s: gofmt and tools on 2 Go files, t.Parallel on 2 test packages; vet 2 packages. Final output is recorded in lane.txt.

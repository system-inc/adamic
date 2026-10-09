Added four fresh-proof effect guards and a nonempty-write floor for the existing corpus fixture, addressing task #0na94yc.
Implementation commit: 39a6c8a48aa462ff74db8415a4472479f5bb42c0; delivery SHA is reported with the push.
Focused tests and 808 selected corpus/guard leaves pass; corpus family restoration takes 14.92s including invocation setup.
M6, M11, M14, M15 and M18 fail their intended assertions with the exact audit diffs, then pass after restoration.
No production changes, new fixtures, whole package runs, full gate or backend execution are included.

The four guards pin the audit's minimal IR behaviors: stat.atime retains its identity through a property read, replacement callbacks escape their operands, an unrecognized buffer operation records WriteUnknown, and class method parameters stay outside the direct-call summary proof. The three field-store guards require one identified write at site 1, Proven false and a nonempty reason. The buffer guard requires its unknown write rather than accepting an empty proof.

The corpus floor is deliberately attached to ../flow/testdata/mutations.a, an existing fixture with Map.set and Array.push reference writes. Other fixtures can legitimately have zero writes or be refused by lowering. This fixture must lower and report at least one write; the restored baseline reports 13 writes, all proven. No fixture was added, so counts.md needs no refresh. The corpus helper still checks unknown writes and missing lowering sites.

All five diffs were read from test-audit/fresh at 5f3fad41, applied with git apply, and reversed before any commit. No production mutation enters this branch's history. Evidence diffs retain their .diff extensions under review/compiler/fresh-guards/.

| Mutant | Guard | Observed failure |
| --- | --- | --- |
| M6 | TestStatAtimeSelfCycleRemainsUnproven | next field write becomes Proven true |
| M11 | TestReplacementCallbackEscapeRemainsUnproven | saved field write becomes Proven true |
| M14 | TestFutureBufferOperationRemainsUnknown | writes become empty |
| M15 | TestClassMethodOutsideSelfCycleRemainsUnproven | saved field write becomes Proven true |
| M18 | TestFreshWrites____flow_testdata_mutations_a_5514557d1b47 | known-write corpus fixture reported zero writes |

M18 was also applied to the exact audit selector TestFreshWrites|TestFreshCorpus: exit 1 at the floor, then exit 0 after reversing the diff. Both runs compile and reach the assertions. family-results.json records 14.33s for the mutant and 14.92s for the restored family.

Every measured leaf is below 60s on a worker with nproc 5 and cpu.max 400000 100000, four CPU equivalents. GOMAXPROCS=4 and -parallel=4 are used. The four new tests report 0.00s at Go JSON's elapsed precision; their separately selected invocation times include Go command setup:

| Test | Recorded test seconds | Complete restored invocation seconds |
| --- | ---: | ---: |
| TestStatAtimeSelfCycleRemainsUnproven | 0.00 | 2.02 |
| TestReplacementCallbackEscapeRemainsUnproven | 0.00 | 1.92 |
| TestFutureBufferOperationRemainsUnknown | 0.00 | 1.97 |
| TestClassMethodOutsideSelfCycleRemainsUnproven | 0.00 | 1.92 |
| TestFreshWrites____flow_testdata_mutations_a_5514557d1b47 | 0.05 | 2.53 |

The combined run selects the corpus family, four new guards and the existing TestMethodKeepsArgument. It passes 808 leaves, with maximum leaf time 1.58s. timings.json records every affected corpus leaf and guard. Refused fixtures remain expected passes except for the known-write floor fixture, which must lower.

Commands and outputs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 180 bash cloud/setup.sh > /tmp/fresh-setup.log 2>&1
source /workspace/adamic-tools/env.sh
timeout 600 python3 review/compiler/fresh-guards/verify.py > /tmp/fresh-mutant-progress.log 2>&1
timeout 240 python3 review/compiler/fresh-guards/verify-family.py > /tmp/fresh-family-progress.log 2>&1
timeout 120 env GOMAXPROCS=4 ADAMIC_GATE_UNCACHED=1 go test ./internal/fresh -run 'TestFreshWrites|TestFreshCorpus|TestStatAtimeSelfCycleRemainsUnproven|TestReplacementCallbackEscapeRemainsUnproven|TestFutureBufferOperationRemainsUnknown|TestClassMethodOutsideSelfCycleRemainsUnproven|TestMethodKeepsArgument' -count=1 -parallel=4 -timeout 90s -json > review/compiler/fresh-guards/corpus.jsonl 2>&1
```

Each verification driver records its exact focused commands in results JSON and writes Go output directly to JSONL files. Every command carries a hard timeout. verify.py additionally requires the intended named test to reach failure, so a build failure cannot count as killing a mutant.

Setup succeeded: Node ready 0.029s, Go ready 0.031s, markdown dependencies ready 0.078s, submodules ready 0.079s, clang ready 0.182s, Go build ready 44.618s, build cache warm 44.829s and done 44.862s. Full output is setup.txt; nproc.txt records 5.

The branch starts from main 553ad06abb069736161b2d46532d9036e92f30b3. Current main's GraphQL test changes merged without conflicts and without changes to internal/fresh. All own commits contain only _test.go and review evidence. Production code remains identical to main, preserving the test-only lane.

Required lane command, after committing:

```sh
timeout 60 git fetch -q origin main devtools/fast-gate cloud/merge-tree
timeout 120 bash -c 'git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'
```

Initial lane result: lane checks 1.2 s: gofmt and tools on 2 Go files, t.Parallel on 1 test packages; vet 1 packages. Final lane output is in lane.txt.

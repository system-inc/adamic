Closed the lowering chain's six named reds toward integration task #wj4pmt1, including restoring the structural Map soundness boundary.
Implementation a6cb5660; initial main merge 5d0d69b7; subsequent main merges a36b48f6 and 9a51ddac, the latter incorporating PR #250.
All named checks pass with Go timeout 90s; full lowering and IR pass within 60s; scoped Node/backend oracles, scoped vet and integration lane checks pass.
Six independent runner mutants are caught, plus the enum cleanup mutant caught by LeakSanitizer with Node stdout unchanged.
No full gate, whole stage-1 package suites, performance measurements, or new oracle fixtures; counts.md is unchanged because no fixture was added.

The base is origin/compiler/lowering-chain-fixed 71972e180956b9a5d82700f31994ab01988a6efc. The initial main merge conflicted in internal/lower/import_cycle_ready_test.go. The final resolution preserves main's computed-answer probe and source Node comparison, the chain's native/JavaScript comparisons, and Step 21's already ruled uncaught stdout/exit-1 contract. It also rejects unexpected native stderr so a sanitizer finding cannot masquerade as the expected exit 1. The latest merge includes main cf79ecec and conflicted only in the call-target table.

Per red:

- Parser TestPushSpreadGap now requires the unchanged proving program to lower and agree with source Node on sanitized native with leak detection and JavaScript. GAPS.md records closure. Its child-index append loops remain because this test did not request removing them; no parser port recut is claimed.
- CSS strings TestMultiPushGap now requires the unchanged proving program to agree across Node and both backends. GAPS.md records closure. The sub-printer's adjacent single pushes become multi-value pushes as its old test required. TestMultiPushPortMatchesGo holds all 9,591 corpus texts, both quote preferences, to the actual Go cohere implementation on source Node, emitted JavaScript, and native ASan/UBSan/LSan. This includes every changed push site; no Go cohere code was copied.
- JSON TestDocumentedStageZeroGaps now requires multiplePush.ts to agree across Node and both backends, including the leak check. It is a top-level leaf rather than a one-row refusal subtest. GAPS.md records closure. The JSON port retains its single-value pushes because its old test did not request removing that workaround; no JSON port recut is claimed.
- TestOptionalIndexingMapShapeRefused remains a refusal row. git log -S on the refusal identifies cfa294609c27c921047417bbc4a14b5e371da510, which explicitly kept structural Map dispatch unsupported. git log -S chainMethodCall identifies 28bf5bfbea924356e4e82c7de5848c053f4707f6, the optional-calls implementation delivered through compiler/optional-calls-main b6ef2cb4 (side-stack merge c8ab7ae2). Its commit explicitly says no acceptance/refusal policy is relaxed. The later continuous-chain intrinsic path bypassed optionalMapGet's boundary. A temporary acceptance experiment observed source Node stdout "7\n" and JavaScript backend stdout "". That is observed disagreement, not successful acceptance evidence. The conservative sound reading restores the existing structural receiver check in chainIntrinsic, before MapGet construction. A real native Map get still passes the Node/backend oracle. This is a soundness regression, not an authorized acceptance.
- TestCallTargetReaders: throwsOutReadiness uses Program.ClosureMayThrow, which routes through ClosureTargets, instead of reading ArraySort.Comparator. The obsolete throwsOut allowlist entry is removed. Following the user's PR #250 steering, main's approved compiler-owned class-static reader and aligned table are retained, and this unit's initial class-test CallTargets change is reverted. The fixed chain's preexisting optional_callable and discarded_closure_result runtime-reader entries remain because the chain still contains those readers; they are not new exemptions from this unit.
- TestEnumCleanupMutant anchors to the final main-scope release and global clear, distinguished by its newline/indentation from nested uncaught-error exits. Only the release is removed; clearing remains. The mutant preserves Node stdout and is caught only by LeakSanitizer.

The requested full lowering run exposed additional stale tests: the two optional-index diagnostic strings now name the continuous-chain path, and TestFSStatThrowsByDefault expected retired exit 70 rather than the ruled exit 1. These test expectations were reconciled without changing their production behavior. The first full run also lacked stage3/api's pinned @types/node 25.3.3; npm ci --prefix stage3/api supplied the existing locked dependency. The setup script itself succeeded.

Toolchain setup used GOPROXY='https://proxy.golang.org|direct', then timeout 480 bash cloud/setup.sh, then source /workspace/adamic-tools/env.sh (the printed path; /opt/adamic-tools/env.sh is absent). Timing lines: Node ready 0.032s; Go ready 0.036s; clang ready 0.221s; markdown dependencies ready 1.395s; submodules ready 19.509s; go build ready 311.735s; test binaries deferred 311.861s; build cache warm 311.862s; done 311.893s. nproc=5; cgroup CPU quota=4. Go 1.27.1, clang 20.1.8, Node 24.19.0. setup.log preserves all lines.

Commands, all test output redirected to evidence logs:

```sh
source /workspace/adamic-tools/env.sh
# Every named and additionally reconciled leaf; final-pr250.log.
timeout 120 go test ./stage1/typescript/parser ./stage1/cohere/cssstrings ./stage1/cohere/json ./internal/lower ./internal/ir ./internal/oracle -run '^(TestPushSpreadGap|TestMultiPushGap|TestMultiPushPortMatchesGo|TestDocumentedStageZeroGaps|TestOptionalIndexingMapShapeRefused|TestClassStaticInitializerCallIsEmitted|TestCallTargetReaders|TestEnumCleanupMutant|TestUndecidedCycleReadsUseReadyChecks|TestOptionalIndexingKeepsUnsupportedStorageNotYet|TestFSStatThrowsByDefault)$' -count=1 -v -timeout 90s
# Full packages, each independently bounded to 60s; lower-pr250.log, ir-pr250.log.
timeout 60 go test ./internal/lower -count=1 -timeout 90s
timeout 60 go test ./internal/ir -count=1 -timeout 90s
# Scoped ordinary differential fixtures; scoped-oracle.log, PASS 10.411s.
timeout 120 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(optional_indexing_map|closures_throw|closures_throw_uncaught|call_targets_sort|sorting|enums)\.a$' -count=1 -v -timeout 90s
# Six independent mutants; each subprocess timeout 120s and Go timeout 90s.
timeout 720 python3 review/compiler/lowering-chain-reds/run-mutants.py
# Scoped vet, exit 0, empty output; vet.log.
timeout 90 go vet ./internal/lower ./internal/ir ./internal/oracle ./stage1/typescript/parser ./stage1/cohere/cssstrings ./stage1/cohere/json
# Required integration checks after committing.
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

The fixture selector is Go's matching regex, not an exact path census; it also selected enums-related fixture names. scoped-oracle.log lists every actual selected row. No fixture in that run failed. An early cold-cache aggregate build hit its external 120s cap. A combined package rerun overlapping other validation hit its external 60s cap. Those are not credited as passes. Subsequent uncontended full-package checks passed in 27.411s (lowering) and 0.830s (IR), before the PR #250 composition; final independent results are in lower-pr250.log and ir-pr250.log.

Mutants on the PR #250 composition, from mutants-pr250.log:

| Mutant | Catcher | Process seconds |
| --- | --- | ---: |
| Restore the old multi-value/spread push stop | All three named push gap acceptance tests fail their lowering requirement | 2.77 |
| Bypass the restored structural Map boundary | TestOptionalIndexingMapShapeRefused receives nil rather than NotYet | 8.22 |
| Restore the direct sort-comparator read | TestCallTargetReaders reports throwsOutReadiness's unapproved read | 13.88 |
| Rename the static-reader test, leaving its direct read outside the approved function | TestCallTargetReaders reports the new function's unapproved read | 13.15 |
| Restore the stale throwsOut allowlist row | TestCallTargetReaders reports the stale entry | 12.06 |
| Omit the first item from the sub-printer's combined push | TestMultiPushPortMatchesGo detects native stdout's first differing byte against Go | 4.85 |
| Omit the unique final enum release while retaining its global clear | TestEnumCleanupMutant observes LeakSanitizer; Node stdout is unchanged | See final-pr250.log |

Each of the six external mutants exits 1 from its intended assertion. None is credited for C build warnings, compiler errors or missing files. The runner restores source after every disk mutation, including on failure; overlay sources are .go.txt or .ts.txt, never compilable Go. Preliminary invalid experiments are retained and explicitly excluded: a Go overlay alone cannot mutate the files go/types scans from disk, and a scratch .a CSS copy could not satisfy a helper that hashes strings.ts. The corrected runner mutates and restores only the scanned/source-data file for those cases. The pre-PR #250 direct static-reader mutant is superseded by the renamed-reader mutant because that exact direct read is now approved on main.

The final leaf timing list and latest lane output follow below. Earlier lane checks passed: "lane checks 17.9 s: gofmt and tools on 247 Go files, t.Parallel on 20 test packages; vet skipped, over 10 s". Explicit scoped vet supplied the skipped coverage. The checkout tracks only main by default, so the two integration tool branches were also fetched with explicit remote-tracking destinations before running the exact required command.

Final composition: full internal/lower passed in 28.244 s; full internal/ir passed in 0.689 s. Named leaves: parser push 0.57 s, CSS push 0.76 s, CSS Go parity 4.08 s, JSON push 0.74 s, Map refusal 0.11 s, call-target readers 1.35 s, enum cleanup mutant 0.57 s, readiness cycles 3.10 s. All use -timeout 90s.

Follow-up: first push succeeded at 3f27d3e5. Source Node confirms failing readiness cycles exit 1, stdout "3\n"; generated JavaScript and native agree. readiness-node.log records all seven rows (2.135 s). The guard-bypass overlay changes generated C conditions to false: source Node and JavaScript still exit 1 with "3\n", native exits 0 with "3\n0\n". TestUndecidedCycleReadsUseReadyChecks/early_value fails both stdout and exit assertions in 0.28 s; readiness-mutant.log records it. This is an eighth caught mutant. The first post-evidence lane invocation omitted the toolchain environment and could not locate gofmt (lane-missing-env.log); the corrected invocation sources /workspace/adamic-tools/env.sh. Logs ignored by the repository default are explicitly force-added to evidence.

Final follow-up composition merges main 3a1c8b57 as 3d872bb1. counts.md had conflicts: preserve chain measured rows and append main additions. Scoped TestCountsAreRecorded with -args -update-counts measured node_buffer_digest_twice, node_buffer_finalized, namespace_callable_properties and pow_fractional; apply only those rows to the full table (the filtered updater writes a partial table). The digest-twice row becomes allocations/frees 5/5, retains/releases 7/10, peak 3, regions 0. Other measured rows already agreed. Each count leaf is 0.21–0.51 s. Full internal/lower rerun passed 30.305 s, full internal/ir passed 17.792 s, each externally bounded 60 s with Go -timeout 90s. Corrected lane output: lane checks 3.3 s: gofmt and tools on 246 Go files, t.Parallel on 20 test packages; vet 20 packages. Final command repeats the exact required lane check after committing this evidence.

Final required lane check passed: lane checks 3.2 s: gofmt and tools on 246 Go files, t.Parallel on 20 test packages; vet 20 packages. Final verification produced no whitespace errors.

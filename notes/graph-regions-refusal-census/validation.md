# Validation on Linux

All commands used the same cloud/setup.sh toolchain and GOPROXY setting. Every
command had an outer hard timeout; long runs were monitored. Test output went to
files. No compiler/runtime source change was needed. Count records and family
wrappers are unchanged from 749e668e. internal/fuzz was not modified or run.

| Command / selection | Result |
| --- | --- |
| cloud/setup.sh --wasi-sdk | pass, 43.179 s after reclaiming disposable cache space |
| retained census.py, exact main versus batch 8 | pass: 771 file-backed promises; exactly 63 Refused to Compiles, no other kind transitions |
| stage3/fixtures independent lowering census | all 84 Refused/NotYet rows examined; exactly host 14 and objects 29 admitted |
| whole stage3/fixtures, count=1, timeout 30m | pass, 211.802 s |
| final stage3 directory/path/shard guards | pass, 0.016 s |
| final host 14 and objects 29 fixture checks | pass, 9.686 s, both Node/native/stage0 leaves |
| whole internal/oracle, ADAMIC_ORACLE_WASI=1, count=1, timeout 40m | 1653.814 s; 57 top-level tests failed from absent stock TypeScript or disk-full count/review builds. All other tests, including all 128 native shards and the refusal/WASI families, passed |
| affected oracle retry: Counts, AssignmentProof, ReviewProgramsAgreeWithNode_006 through _009, CountMeasurements; ADAMIC_ORACLE_WASI=1 | pass, 130.424 s; every one of the original 57 failed top-level tests is present as PASS; no count record changed |
| all 63 admission proofs, including counted balance, count=1, failfast | pass, 61.858 s; all newly found seven rebuilt uncached; 56 former-refusal native cache misses and all 126 backend/source Node observations measured |
| final seven proof guards plus native-summary guard and shared-helper caller | pass, 14.151 s |
| objects native-summary generator, update-graph-object-native | pass, 42.664 s; all six compiled object fixtures freshly measured on Node, JavaScript, ASan/UBSan/LSan and counted native before native.json was written |
| whole internal/lower, parallel=1, count=1, timeout 20m | pass, 178.198 s |
| whole stage1/cohere/markdownblocks, parallel=1, count=1 | pass, 896.596 s |
| whole stage1/cohere/yaml, parallel=1, count=1 | pass, 382.015 s |
| whole stage1/typescript/parser, parallel=1, count=1 | pass, 308.671 s; optional compiler-source shards skip without ADAMIC_TYPESCRIPT_SOURCE |
| other stage1 gap/refusal harness selections | pass; CSS, CSS strings, ESTree, gitignore, GraphQL, GraphQL printer, JSON, lint, mediaquery, selector, static_single_assignment, suppression, TS printer, values, scanner; markdowninline has no matching compiler-gap test |
| lint/regex dynamic-pattern gap and option dialects, after its declared oracle owner | pass, 1.030 s; dynamic RegExp remains NotYet |
| native optional-write and prototype refusal probes | pass, 10.546 s; the intentionally bypassed prototype-cycle C probe leaks as its test requires; it is not an admitted program |
| whole cmd/adamic-test262, ADAMIC_ORACLE_WASI=1, parallel=1 | pass, 62.566 s, after NativeRuntimesWASI0 was prepared (16.486 s) |
| whole internal/acheck, including TestEveryProgramOutsideAPackageKeepsItsPromise | pass, 8.654 s; earlier complete run 19.138 s |
| whole cmd/adamic-gate | pass, 1.916 s |
| whole cmd/adamic-meter | pass, 5.188 s |
| whole cmd/adamic-stage1-progress | pass, 0.085 s |
| whole internal/conformance | pass, 1.212 s |
| stage3/meter Python unittest discovery | pass, 42 tests in 10.157 s, four real-source/binary-prerequisite probes skipped |
| gofmt -l; go vet internal/oracle internal/acheck stage3/fixtures; git diff --check | pass |

Earlier exploratory broad runs exceeded the box's temporary-disk or memory
capacity, and consumer-only stage1/test262 runs found missing oracle-owner
products. The final runs above prepared the declared owners and ran serially.
The four typeaware/lint helper/comment runtime-refusal selections found during
string searching are runtime input-validation contracts, not recorded compiler
outcomes; they are distinguished in the inventory rather than treated as graph
admissions. The actual lint regex compiler gap was run and is unchanged.

The script inventory also includes dated source-ledger recorders and frozen
compiler snapshots. For example real/bytes-revealed/check.py explicitly requires
ed6e2975; it cannot assert current compiler outcomes. Those historical records
were not rewritten. The objects current native-summary schema is now covered by
TestObjectFixtureNativeSummaryMatchesCompiledCases.

## Mutants

- Restoring objects 29's Refused status reds
  TestFixturesObjects/objects/29_parameter_default.a/stage0.
- Restoring host 14's Refused status reds
  TestFixturesHost/host/14_getCurrentDirectory.a/stage0.
  Both mutants still pass their Node and native leaves; the status checks are
  the intended red. Combined mutant command: expected exit 1, 54.831 s.
- Keeping the four memoize known entries and cwd's old refused header reds
  TestEveryProgramOutsideAPackageKeepsItsPromise, naming exactly those five.
- Restoring host 14's old first-line refusal reds
  TestRecordedGraphAdmissionHostCurrentDirectory, expected exit 1, 0.007 s.
- Dropping only objects 29's native observation reds
  TestObjectFixtureNativeSummaryMatchesCompiledCases, expected exit 1, 0.007 s.

See the adjacent compressed logs and the per-bucket graph-admission-mutant.json
files. There are no output disagreements, sanitizer defects or counted leaks in
any admitted program, and no musts.

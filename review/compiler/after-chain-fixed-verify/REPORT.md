Verified the combined chain d3558780 with current main merged, with no combination fixes needed.
Tested SHA: 5235536092fe1190225e480f82822de8f682b63c; the delivery adds evidence only.
The requested red list passed together, full lowering passed, the call-target guard passed, and recorded counts passed unchanged.
Both requested record mutants were caught by Node comparisons; the namespace registration mutant's test passed.
Not covered: full native/oracle packages, the full gate, other WASI shards, and existing skipped opt-in or pending leaves.

Fetched compiler/after-chain-fixed d3558780 and created compiler/after-chain-fixed-verify. Merged main 2102ee6b as 70bffe5e. During verification main advanced to 6eef5ae5, adding only lint sanitized-test sharding; merged it as 52355360 and reran the requested checks. No compiler, test, fixture, or counts source was changed by this verification unit.

Each row below passed at 5235536092fe1190225e480f82822de8f682b63c:

| Test | Result |
| --- | --- |
| TestRecordsAgainstNode | PASS |
| TestRecordMutantsUnit00 | PASS |
| TestRecordMutantsUnit01 | PASS |
| TestEnumInitializationUnknownPinned | PASS |
| TestParserNamespaceClassRegistrationMutant | PASS |
| TestParserConstructionUnsetUse | PASS |
| TestParserStackTinyLimit | PASS |
| TestGenericBodyRelationsMaybeBind | PASS |
| TestCheckedJSONNextRefusals | PASS |
| TestWASIAgreesWithNode/shard-002 | PASS; 146 fixture leaves |
| TestWASIAgreesWithNode/shard-004 | PASS; 146 fixture leaves |
| TestFreshCorpusRemainder | PASS |
| TestParserRepresentationProbes | PASS |
| TestGapStandsWhereGapsMdSays | PASS |
| TestBigintGapStandsWhereGapsMdSays | PASS |
| TestFixtureDirectoriesHaveTopLevelTests | PASS |
| full internal/lower | PASS |
| TestCallTargetReaders | PASS |
| TestCountsAreRecorded | PASS |

The complete red selection ran in one Go invocation with ADAMIC_ORACLE_WASI=1 and ADAMIC_GATE_UNCACHED=1. The six final package outputs were native 18.289s, oracle 57.756s, fresh 9.753s, markdownblocks 5.363s, scanner 8.118s and fixtures 0.077s, all ok. Both WASI shards ran all 146 fixture leaves each. Full lowering passed in 61.404s; TestCallTargetReaders passed in 0.778s; the counts package passed in 26.714s (1245 native observation hits, 3 misses), without update-counts or a counts.md edit. results.json records the tested SHA with every result. commands.txt contains the exact commands; all test output was redirected to the logs here, never piped.

Existing skips are explicit limitations. TestCheckedJSONNextRefusals/json_maplike_pending says t.Skip("awaits compiler/records-maplike: any-valued own-key storage"). Full lowering has three environment-dependent inventory skips (TestOriginalCycleLedger, TestOptionalWideningCensus, TestGenericBodyRelationsCensus) and the existing mixed-union contract leaf skip. Their reasons remain in lower-merged.log; none was added or weakened here.

The record mutants indices-in-insertion-order and uint32-max-as-index both completed cleanly and were caught by comparison to Node; their logs say caught by Node. TestParserNamespaceClassRegistrationMutant passed its missing-registration rejection assertions. No new checks or fixtures were introduced, so no additional mutants or counts refresh were needed.

Setup ran with GOPROXY=https://proxy.golang.org|direct. Timing lines: Node 0.022s, Go 0.031s, submodules 0.065s, markdown dependencies 0.072s, clang 0.169s, WASI SDK 4.151s, go build 52.056s, cache warm 52.180s, done 52.204s. nproc reports 5; cgroup quota is 4 CPUs. Every test shell sourced /workspace/adamic-tools/env.sh.

The initial combined command hit its outer 90s limit after records passed, while cold oracle work was still running; it is not claimed as a full pass. The initial counts check also hit its outer 90s limit while compiling/measuring fixtures, with no assertion failure. Both were rerun successfully, then rerun again after merging the newer main. The interrupted and successful logs are retained separately.

Lane checks on the first merge passed: gofmt/tools on 277 Go files, t.Parallel and vet on 23 test packages, 13.3s. The required lane command is run again on the committed final delivery before pushing; lane-final.log records it and the final response reports its output. This is verification of the lowering chain's step 21/24 fixes and the corpus/gap reconciliation, not new semantics.

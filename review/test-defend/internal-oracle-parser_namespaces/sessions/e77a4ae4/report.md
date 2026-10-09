# Parser namespace family defense

Result: not defended after three production-only attempts. This is a bounded finding and does not authorize a test deletion.

Code under test: Adamic namespace lowering in internal/lower, particularly callableNamespaceUse, namespaceOwnThis and namespaceReadyReads. Oracle: Node runs each original source; both lowered JavaScript and sanitized native output must agree on stdout, stderr and exit; successful originals also require clean leak checks. No oracle, harness or test was changed.

D means TestNativeAgreesWithNode family. Current parser family members are TestParserNamespaceReceiver, TestParserNamespaceClass and TestParserCallableNamespace. They call the same checker with pairs from the same six fixtures registered with D. D compares the same source to the same JavaScript and sanitized native products, checks the same leaks, and additionally compares a release product. The family calls Node before Lower; D calls Lower first. These six originals have no input history or filesystem side effects that give this ordering a distinct expected answer.

Both per-family compiler coverage profiles passed: target covered 2137 blocks, D covered 2137, no exclusive blocks either way. Coverage is only a lead; the whole test bodies and six original sources show no semantic input difference. Therefore the three attempts test different claimed corner cases rather than an invented exclusive input: D1 refuses typeof on callable namespaces; D2 stops exempting ordinary object methods from namespace receiver capture; D3 inverts a readiness guard so initialized namespace member reads panic. Each has a real compiler or runtime behavior change observed in the matrix. Every diff applies to this main base and passed go vet ./internal/lower/; D3 also successfully built native products before the observed runtime failure.

All three were caught by both families. D3 was also caught by TestNamespaceLiveExportBoundary. matrix.json lists all observed passes and failures, including subtests. There were no survivors among these three attempts. Witness/setup rows in the expanded run are recorded, but their production failures would not prove witness strength.

The whole clean package timed out at 90.073 binary seconds without an individual test failure before timeout. We narrowed to the three family members, D on six exact shared inputs, TestModuleNamespaceReadsMatchNode and TestNamespaceLiveExportBoundary. The package grew from 193 to 198 top-level tests; all five additions are in the expanded matrix: FractionalPowersReachRuntime, ReviewProgramsNoLooseFiles, ReviewProgramsRefuse, ReviewProgramsSelfTest, and ReviewProgramsAgreeWithNode (smoke.a only). There are 263 review .a inputs; none contains the word namespace. Remaining review-agreement inputs, other D inputs and other package rows are unknown. The bounded clean runs all passed before mutation; the restored family passes again.

Issues and time costs:
- The brief's unexplained D required resolving the audit's row names and matrix commands. It denotes a partially exercised family, not a test named D.
- The whole package exceeds 90 seconds, so package-wide uniqueness was unavailable. Another observed catcher is sufficient to reject a unique-catch claim for every attempted mutant; unknown rows cannot undo those observed catches.
- /tmp has only 8.8 GB total capacity. Disk-first cleanup removed only the preceding unit's /tmp/defend-unicode-alias scratch/cache. It had 8.8 GB free afterward; /workspace had 20 GB free. The requested 15 GB threshold cannot be met on /tmp, and there was no full-disk baseline failure.
- One review selector was over-escaped and matched no smoke subtest. The first runs are preserved as review-unmatched logs and are not evidence. Corrected replays assert the exact smoke subtest passed.
- Twins and cost exceptions do not apply: these rows are not separate executors or performance checks.

Owner finding: the family's name and assertions match its namespace correctness claim. It actually checks receiver, class and callable behavior against Node; no unasserted performance promise was found. The defense found duplicate inputs and covered behavior, not a vacuous assertion. Do not infer repository-wide redundancy from this bounded three-mutant result.

Warm tool setup was skipped; nproc=5. npm ci ran before baseline. Baseline binary 90.073 s; target coverage 0.909 s; D coverage 0.966 s; expanded clean 3.751 s; smoke clean 0.386 s. commands.json records each go-vet/rebuild/run wall time and cache. Each mutant used its own ADAMIC_BUILD_CACHE_DIR and ADAMIC_GATE_UNCACHED=1. No test was deleted, rewritten or weakened; all production sources were restored before committing evidence.

The starting commit is 7b18d0576930caca4e22ce2eef92fcf563af52d0, fetched from origin/main. All fifteen requested names exist in their named files. The older commit printed in the brief is not the starting commit.

The whole package has 37 top-level tests. The baseline passed with ADAMIC_TEST262_MEASURE=1 and no skipped rows. Each mutation ran the whole package, including the rows outside this unit. Repo-wide uniqueness was not checked.

The scoped tests have distinct assertions. Shared preparation and execution helpers do not make these tests one family under the brief's exception for distinct assertions. TestRunnerLocationHelper is only a subprocess entry for TestRunnerLocationIdentity. Its three default timings measure immediate return.

The code under test is the Go test262 runner, not Node, clang, internal/lower or internal/native.C. CompilerStartupMeasurement compares subprocess Adamic against in-process Adamic. That reference is self, despite the subprocess invocation. Mutating the shared lowering or C emitter would change both answers, so this audit left them intact.

Classification, verdict JSON, diagnostic normalization, rewriting and frontmatter expectations are handwritten. Diagnostic text containing a TS code does not make normalization an external-authority oracle. No external authority value was checked in this unit.

The integration tests use Node agreement and also handwritten counts, labels or cache invariants. WorkerLazyFallback's fallback assertion checks exit zero and nonempty C, without comparing that C against another implementation. Classification reasons use substring matching. These are limits of the assertions, not inferred correctness claims.

The mutation plan was frozen before any kill result. It spans production classification, metadata, rewriting, comparison, normalization, identities, buffering and fallback. M06 and M13 mistakenly drop predicates, outside the requested menu. They are supplemental. Their kills are reported separately and do not support a verdict. NodeHarnessIdentity consequently has no admissible demonstrated kill and is cannot-judge, rather than claiming a sacred verdict from the supplemental result.

The initially written switch generator treated Go byte offsets as Python character offsets and omitted variadic forwarding. Both caused build failures before mutant test runs. The corrected generator preserves the original functions and dispatches to code-derived variants through ADAMIC_MUTANT. Failed build attempts remain logged. These implementation mistakes cost time; they were not test failures.

The before-mutation function inventory conservatively listed all production functions. reached-functions.json adds a static transitive graph from immutable source. Method selectors are resolved by name, so this graph is an overapproximation. It is not a coverage profile or proof that every listed branch ran.

Timings use the test binary's package elapsed field from three independent -count=1 runs. This excludes Go command compilation time. Command wall times are separately recorded for the baseline and matrix. Warm medians can hide cold costs: LargeCompilerOutputIsComplete took 12.526, 4.026 and 4.137 seconds in its isolated runs.

The opt-in startup measurement was enabled throughout. Its 30 interleaved compilations have no asserted performance threshold. It still checks C equivalence and can fail under output capture mutations and its empty-answer probe.

Empty-answer probes are separate from mutation verdicts. They run only the scoped rows that call their respective entries. NodeHarnessIdentity's compiler-only edit assertion accepts an empty identity before its runner-edit assertion rejects it. RunnerLocationIdentity rejects an empty prepareCache answer through its helper's error path. Such probe failures show rejection of no answer, not semantic correctness of a working answer.

There are no compiler or port-source mutants, so this unit does not require fresh native product cache keys or per-port rebuilds. Go runner variants are compiled into one switched test package. All standalone production diffs have no switch, apply to the starting index, and are individually checked with go vet.

The approximate 20-minute budget was exceeded. Full-package columns took tens of seconds each, on top of 45 isolated timing invocations, the clean baseline, reading the production code and repairing instrumentation. The matrix was not narrowed simply to manufacture uniqueness or meet the unit budget. Exact measured totals are saved in the final report.

Subsumption is only a hint from fifteen admissible mutations. Supplemental mutations and probes do not count. No test deletion is proposed. No other packages or repo-wide replay were run, and no external oracle implementation was mutated.

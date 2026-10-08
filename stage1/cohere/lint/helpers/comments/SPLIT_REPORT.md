# TestCommentMutants split, October 8, 2026

Branch: `stage1-split/lint-helpers`.
Base: `origin/area/stage1-lint`, `bd5490b6ff36cfc6013d549f78ffa4f9f9a6e075`.
The initial workspace was main `73352e874ddbda5a78c95c5c670860d43275e4d0`.
Only the comments package was changed. Eleven requested tests remain.

`nproc`: 5; cgroup `cpu.max`: `400000 100000` (4 CPUs).
Initial load: 0.00/0.00/0.00. Baseline command sampled one-minute load up to
6.02 during Go harness compilation, ending at 2.37. Isolated after run started
at 0.85/1.28/0.93 and ended near 0.62/1.12/0.91. No other test workload ran
alongside that isolated measurement.

| Test | Before own wall | After longest shard | Shards | After parent wall including build inputs |
| --- | ---: | ---: | ---: | ---: |
| TestCommentMutants | 69.38 s | 0.03 s | 5 | 64.67 s |
| TestCommentMutantShardCatchesSurvivor (new proof) | n/a | 0.03 s (whole test) | 5 subprocess selections | 0.03 s |

The untouched baseline ran first with `go test -json -run
'^TestCommentMutants$' -count=1 -timeout 3h`. All five mutants ran against all
47 witnesses. Go harness compilation precedes the first JSON test event and
is excluded from the test's own wall.

A second run of the original test used a temporary Go overlay adding only
build timers. Its own wall was 62.94 s: 62.826386 s builds and 0.113614 s
execution/comparison/other overhead. Oracle build: 1.451882 s; native builds
in original mutant order: 16.016000, 11.263704, 11.300111, 11.457366,
11.337323 s. The first native build includes the cold runtime archive.
This separate profile attributes the work; it does not replace the untouched
69.38 s baseline.

## Build inputs and cold product measurements

`internal/buildcache` is absent on the base. Each product is built once before
parallel subtests and shared. Build callbacks write into their supplied directory;
inputs are documented beside the callback. No cache was added to this package.
The compiler's existing hashed runtime archive remains an input.

The isolated after run used a new `XDG_CACHE_HOME` for a cold runtime archive
and fresh output directories for every product. The Go dependency compilation
cache was retained from harness compilation; these are cold output-product
measurements, not measurements with every Go dependency archive deleted.

| Build product | Cold output-product wall |
| --- | ---: |
| Sanitized runtime archive, shared | 4.960440 s |
| Go oracle executable | 1.281182 s |
| can_begin_at.ts sanitized native | 11.997511 s |
| collect_list_interiors.ts sanitized native | 12.039306 s |
| sort_by_position.ts sanitized native | 11.874037 s |
| all.ts sanitized native | 11.474325 s |
| for_file.ts sanitized native | 11.023239 s |

| Mutant | Load | Lower | Emit C | Clang |
| --- | ---: | ---: | ---: | ---: |
| can_begin_at.ts | 0.066608 s | 0.523026 s | 0.354459 s | 11.053149 s |
| collect_list_interiors.ts | 0.071998 s | 0.559255 s | 0.414919 s | 10.992945 s |
| sort_by_position.ts | 0.073132 s | 0.510991 s | 0.357010 s | 10.932731 s |
| all.ts | 0.071291 s | 0.529261 s | 0.375788 s | 10.497767 s |
| for_file.ts | 0.072258 s | 0.482792 s | 0.381633 s | 10.086367 s |

Every unit retains the full witness corpus, Go output comparison, original
ASan/UBSan flags, nonzero-exit checks, and empty-stderr check (including Linux
LeakSanitizer reports). Each parallel unit enforces a 30 s execution ceiling.
`ADAMIC_TEST_SHARD=i/n` uses zero-based indices and modulo assignment;
unset runs all five units.

## Coverage and failure proof

Union: 5 mutants x 47 witnesses = **235 unique case IDs** across five shards.
IDs include the corpus row index because filenames legitimately repeat.
Code checks counts, duplicate IDs, unexpected IDs, and missing IDs before gate
selection. An initial filename-only implementation failed this check and was
corrected before validation.

The new proof runs the same parallel runner and fatal survivor check in five
subprocesses, selecting each shard separately. Prepared oracle-equal output
plants a survivor for `sort_by_position.ts`. Exactly
`shard-02-sort_by_position.ts` fails with `compiled semantic mutant survived`;
shards 00, 01, 03 and 04 pass. No builds occur in this new proof unit.
The real compiled mutants are also killed by all five full-corpus units.

## Whole package validation

`go test -json ./stage1/cohere/lint/helpers/comments -count=1 -timeout 3h`:
**PASS, 132.116 s; 6 top-level passes, 0 failures, 0 skips**.
All input sets ran, including witnesses, consumers and parser gaps.

| Top-level test | Wall |
| --- | ---: |
| TestCommentsMatchCohere | 17.73 s |
| TestCommentMutants (including builds) | 59.79 s |
| TestConsumerCommentHelpers | 30.03 s |
| TestJsxParserGapIsExplicit | 13.28 s |
| TestJsxAdapterGuardMutant | 11.23 s |
| TestCommentMutantShardCatchesSurvivor | 0.03 s |

`go vet ./stage1/cohere/lint/helpers/comments`: exit 0.
`gofmt -l` on the changed Go file and `git diff --check`: clean.
The other requested packages were not run in this session.

Full local logs: `/tmp/comment-mutants-before.json`,
`/tmp/comment-mutants-baseline-profile.json`,
`/tmp/comment-mutants-cold-final.json`, `/tmp/comments-package.json`,
`/tmp/comments-vet.log`.

## Remaining, longest first

1. lint/regex TestShapeFixtures
2. lint/helpers TestHelperMutants
3. lint/rules/no-underscore-dangle TestCompileProfiles
4. lint/helpers TestHelpersMatchCohere
5. lint/rules/no-unsafe-negation TestCompileProfiles
6. lint/rules/no-unsafe-optional-chaining TestCompileProfiles
7. lint/rules/typescript-no-this-alias TestCompileProfiles
8. lint/helpers/comments TestCommentsMatchCohere
9. lint/helpers/comments TestConsumerCommentHelpers
10. lint/regex TestFixedPatterns
11. lint/helpers/comments TestJsxParserGapIsExplicit

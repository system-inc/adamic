# Prettier fixture selection defense

TestProvisionedPrettierFixtures is defended by D01. The entire current package ran, with the opt-in fixture checkout enabled. Only this row failed; its former subsumer and the other three rows passed. No tests were altered.

## Starting state and prior evidence

Started clean from origin/main b8bcadb2c493173855f19d7e5c508b34f5eeb5b6 on test-defend/internal-corpusfiles. Fetched test-audit/internal-corpusfiles with its full remote refspec and read REPORT.md and rows.json, including the embedded oracle notes and mutation conclusions. Read CLAUDE.md, supporting repository documentation, all of files_test.go, and files.go. The audit started at 7b18d0576930caca4e22ce2eef92fcf563af52d0. Current go test -list confirms the same five top-level tests; none moved or vanished. See list.log.

## Code under test and oracle

Code under test: production corpus file collection through Upstream, checked, selectFiles, git, names, and matches in internal/corpusfiles/files.go. The subsumer additionally calls Repository. The collector is the audited product for this unit. Git and the actual pinned fixture checkout are external inputs and were not mutated.

Oracle: Git tracked-state queries and the test's handwritten requirement that Upstream succeeds on the provisioned checkout. The test does not compare returned file identities or counts with an external answer. It computes and logs extension counts, but those counts do not decide pass or fail. The audit's empty-Upstream finding remains a limitation, not an obstacle to defending rejection behavior.

## Coverage and semantic difference

Commands, each after sourcing /workspace/adamic-tools/env.sh:

```
ADAMIC_CSS_FIXTURES=/tmp/defendcorpus/prettier timeout 120 go test -count=1 -timeout 90s ./internal/corpusfiles/ -run '^TestProvisionedPrettierFixtures$' -coverpkg=./internal/corpusfiles -coverprofile=TestProvisionedPrettierFixtures.cover > TestProvisionedPrettierFixtures.coverage.log 2>&1
ADAMIC_CSS_FIXTURES=/tmp/defendcorpus/prettier timeout 120 go test -count=1 -timeout 90s ./internal/corpusfiles/ -run '^TestConvertedPackageContracts$' -coverpkg=./internal/corpusfiles -coverprofile=TestConvertedPackageContracts.cover > TestConvertedPackageContracts.coverage.log 2>&1
```

There are zero covered blocks exclusive to the requested row relative to its subsumer. See coverage-differences.json and both raw profiles. Semantic coverage differs: the requested row passes three heterogeneous alternatives, *.css, *.scss, *.less, across three roots. Every subsumer subcase passes one pattern. TestDeterministicRootsPatternsAndMissingGit passes two alternatives, but they are case aliases (*.ts and *.TS) under the matcher's case folding. Removing the final alternative still leaves those inputs fully matched.

## Aimed mutant and full matrix

D01 changes the implicit matcher iteration bound from len(patterns) to max(1, len(patterns)-1). It omits the last alternative when multiple alternatives exist and preserves singleton behavior. This is a production off-by-one bound mutation from the allowed menu; it contains no test-name, checkout-name, or pin special case. The plan was saved before the run. Standalone D01.diff applies to starting origin/main and passed go vet ./internal/corpusfiles/ with the mutant applied. Production source was restored afterward.

ADAMIC_CSS_FIXTURES=/tmp/defendcorpus/prettier ADAMIC_BUILD_CACHE_DIR=/tmp/defendcorpus/cache/D01 timeout 120 go test -json -count=1 -timeout 90s ./internal/corpusfiles/ -run . > D01.log 2>&1

Observed failure:

```
files_test.go:170: corpus-files: /tmp/defendcorpus/prettier root tests/format/less: no tracked files match ["*.css" "*.scss" "*.less"]
```

The missing *.less alternative makes selectFiles reject an otherwise valid tracked Less root. All five current top-level rows completed. matrix.json and D01.log preserve all outcomes; rows.json lists the four rows that passed. No narrowing, timeout, panic, skipped row, or unknown outcome. Stopped after one attempt because a unique catch satisfies the defense criterion. No claim of repository-wide uniqueness.

## Provisioning and timing

Warm Go 1.27.1 tools worked; cloud setup was skipped. nproc: 5. npm ci in stage3/api completed with three packages added in 863 ms. Before baseline, fetched Prettier commit cb4b33fba24a8428d00e54be85fc886288a374ea from https://github.com/prettier/prettier.git into /tmp/defendcorpus/prettier, with a cone sparse checkout containing tests/format/css, tests/format/scss, tests/format/less. provision.log records the fetch. Baseline observed counts: .css 157, .scss 90, .less 43. baseline.log proves a green full-package baseline, 0.515 seconds on the binary's ok line. Coverage binaries took 0.090 and 0.413 seconds. D01 full matrix binary took 0.512 seconds. Vet and matrix combined shell tool elapsed about 0.92 seconds; provisioning/build times were not separately instrumented. Each run uses its own specified Adamic build cache where applicable; this package is Go-only and no native product rebuild is required.

## Brief friction and owner finding

The /tmp filesystem has only 8.8 GB total capacity, so the requested 15 GB free threshold is unattainable there. Removed named earlier scratch/cache trees under /tmp only, then checked again: /tmp had 8.8 GB free and /workspace 16 GB free. No repository or tool directory was deleted.

The fixture opt-in was unset and the pinned checkout absent. Provisioning it was necessary to avoid defending a skipped row. The audit's empty-answer result demonstrates a limit in the assertion, but shared line coverage alone did not establish semantic subsumption. All current package rows were included.

The row's name accurately describes exercising provisioned fixtures. It guards successful selection, including heterogeneous alternatives. It does not assert corpus cardinality or returned contents: the count logging must not be described as a completeness guarantee. That limitation remains even though this row is defended. No deletion or rewrite proposed.

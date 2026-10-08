Built: the shared capture retains every compiled option source; no rule matcher or compiler code changed locally.
Commits: library-only 6c5b7f03 merged at 57738619e; both compiler merge attempts aborted with 53 conflicts.
Checks: 350 captured cases, 14 compiled sources; all 350 production cases and 216 findings agree on Go, Node, emitted JavaScript and sanitized canonical native; wall 1024.152 seconds.
Mutant: capture_options.go:35 Source() -> {} fails go test's JSON comparison; vet and formatting checks pass.
Uncovered: the shared split gate waits on compiler integration; Greek property waits on #7zm085y; no full lint package or quiet-hundred rerun.

The only shared file edited is stage1/cohere/lint/shared_test.go. Its captureUpstream overlay calls the owned testdata/capture_options.go encoder. It starts from encoding/json's representation, replaces compiled matcher leaves with Source() strings, and retains JSON tags, omissions, custom enum encoding, nulls and exact numeric values. Cohere remains pinned at f5d1934a; no property-fix branch pointer was taken.

| Rule | Upstream cases | Findings | Identical output bytes | Gate seconds |
| --- | ---: | ---: | ---: | ---: |
| id-length | 204 | 108 | 30,636 | 313.71 |
| no-inline-comments | 58 | 38 | 24,123 | 338.09 |
| no-warning-comments | 88 | 70 | 27,027 | 353.31 |

Every row compares IDs, messages, UTF-16 spans and whole fixed sources. The canonical BuildTSGo path executes native with ASan and UBSan, without a link shim or fallback. The unchanged dynamic_gap.a also passes the checker-linked native builder. Five selected regex tests passed with zero skips; one of those is the explicitly named Greek-property divergence assertion, not an agreement claim. nproc is 5. Observed load was 2.66/1.16/0.48 during concurrent builds, near 1.00 during the native compiler runs, and 0.93/1.04/0.88 afterward.

Commands (all test output redirected to files; logs retained under evidence/link-capture):

```sh
source /workspace/adamic-tools/env.sh
# Every command uses GOCACHE=/tmp/scout-go-cache GOFLAGS=-ldflags=-w GOMAXPROCS=4.
# Native, shared capture and split controls use TMPDIR=/workspace/scratch/scout-native.
go test ./stage1/cohere/lint -run '^TestRegexCompiledOptionCapture$' -v -count=1 -timeout 30m
go test ./stage1/cohere/lint/regex -run '^(TestNativeCheckerRegexLinkGap|Test(Id|Inline|Warning)MigrationFindings|TestOptionPropertyGap)$' -v -count=1 -timeout 3h
go test ./stage1/cohere/lint/regex -run '^(TestNativeSplitAttributeGap|TestCompiledOptionCaptureSources)$' -v -count=1 -timeout 30m
ADAMIC_CAPTURE_SOURCE_MUTANT=1 go test ./stage1/cohere/lint/regex -run '^TestCompiledOptionCaptureSources$' -v -count=1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/regex
git diff --check
```

The source-loss mutant compiles and reaches the comparison: actual exceptionPatterns is [{},null], expected is ["^_",null]. The failed go test names the source-loss comparison, rather than a compiler diagnostic. The original three per-rule matcher mutants and their evidence remain unchanged from the preceding scout unit and were not rerun here.

Both origin/compiler/area-next-fixtures (547551cb) and origin/compiler/split-attribute (3bea093d) conflict in 53 files across compiler, runtime, oracle and fixture ownership. Both merges were attempted without rebasing and aborted. No unrelated conflicts were resolved in this lint unit. The shortest split witness, testdata/split_gap.a with two literals, still fails duplicate definition __attribute__. The canonical finding leg is green; the shared lint package's split leg awaits integration of the compiler fix. The Greek witness remains \p{Script=Greek} with u and input α: pinned Go reports SyntaxError while Node matches. That pattern is parked on #7zm085y, without a fallback.

The first library merge attempt ran out of disk space. Every partial merge file was backed up under /workspace/scratch/scout-interrupted-merge-backup before restoring those paths. Clearing disposable default Go build cache recovered 22 GB; the merge then succeeded. Automatic review rejected the initial broad reset/cleanup proposal; that command was not executed. Recovery preserved all partial files and avoided broad cleanup. The temporarily archived unused Fortran binary was restored after disk recovery.

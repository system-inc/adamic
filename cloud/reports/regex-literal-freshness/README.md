# RegExp literal freshness

Base: `df959ee978da7cfcac45fe0e9ff941b80a6fda58` on `codex/regex-matcher`.
Only a fixture, its registration/count check, and these observations are added.
No matcher/runtime implementation change is needed.

Each of three evaluations of `/a/g` in a loop matches `aa` at index 0 and leaves
lastIndex 1. The function containing `/x/g` is called twice: each starts at 0,
exec advances to 1, and the two returned objects compare unequal. The analogous
sticky `/a/y` loop matches at 0 and leaves 1 every iteration. `node.stdout`
records all eleven observed lines and equals `expected.txt`.

Node 24.19.0 decides the result. The oracle compares the original typed source,
JavaScript backend, sanitized native, and release native; the native leak check
passes. Counted execution allocates/frees 69 values, retains 34, releases 65,
peaks at 10, and has zero values left to regions.

The real mutant inserts a static cache into `adamic_regex_new`, keyed by the
compiled-program pointer. The cache retains objects, so the witness is semantic
sharing rather than use after free. Both native builds exit 0 but print index 1
on their second loop iteration, miss on their third, start the second function
call at lastIndex 1, and compare the function results equal. The sticky loop
also advances/misses. The fixture fails with `stdout differs`; the mutant test
exits 1. The driver restores the runtime in a finally block.

Commands (output goes to files):

```sh
source /workspace/adamic-tools/env.sh
export GOFLAGS=-buildvcs=false
export GOWORK=/tmp/regex-literal-freshness.go.work
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/regexp_literal_freshness.a > cloud/reports/regex-literal-freshness/node.stdout 2> cloud/reports/regex-literal-freshness/node.stderr
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_literal_freshness|^TestRegExpLiteralFreshnessCounts$' -update-regexp-literal-freshness-counts -count=1 -v -timeout 15m > /tmp/regex-literal-freshness-green.log 2>&1
python3 cloud/reports/regex-literal-freshness/run-mutant.py > /tmp/regex-literal-freshness-mutant-driver.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$|^TestRegExpLiteralFreshnessCounts$|TestNativeAgreesWithNode/internal/oracle/testdata/regexp_literal_freshness' -count=1 -v -timeout 20m > /tmp/regex-literal-freshness-final.log 2>&1
```

The scratch Go workspace uses this matcher checkout and the unchanged installed
`/workspace/adamic/cohere/TypeScript/tsc`, reusing its warm build cache. VCS
stamping is disabled because the isolated checkout links the installed
submodule; it does not change compiler semantics. No workspace/toolchain files
are committed. The unit's setup previously completed in 119 seconds (`nproc=5`,
CPU quota 4).

This is a literal-allocation control, not a new survey or a performance change.
The wider protocol unit resumes after this standalone matcher commit is pushed.

Final observations: initial oracle/count control passed in 2.644s; the hoisting
mutant failed in 60.681s with the required stdout witness; restored fixture and
complete recorded-counts gate passed in 67.494s. Runtime diff after restoration
is empty. Logs are checked in alongside this report.

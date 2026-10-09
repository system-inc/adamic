# Reproduce the Array measurements and checker evidence

Use Linux, source `/workspace/adamic-tools/env.sh`, and keep test output in files. The test262 checkout is `/workspace/test262` at `c8c798898646638cd0c24879f8e0374e847e7d74`. The baseline is Adamic `5d4c801`; use a detached worktree at that commit when regenerating baseline evidence.

```sh
source /workspace/adamic-tools/env.sh
git worktree add --detach /tmp/library-array-baseline 5d4c801
cd /tmp/library-array-baseline
go run ./cmd/adamic-test262 -adapt -json -test262 /workspace/test262 -work /tmp/library-array-before built-ins/Array > /tmp/library-array-before.json 2> /tmp/library-array-before.log
```

The archived exporter uses unchanged runner functions, rather than reimplementing adaptation. From `/workspace/adamic`:

```sh
mkdir -p /tmp/library-array-classifier
for part in adapt classify frontmatter prelude rewrite; do
  git show 5d4c801:cmd/adamic-test262/$part.go > /tmp/library-array-classifier/$part.go
done
cp docs/library-array/export-adapted.go.txt /tmp/library-array-classifier/main.go
go run /tmp/library-array-classifier/*.go > /tmp/library-array-adapted.jsonl 2> /tmp/library-array-export.log
mkdir -p /tmp/library-array-tsc
npm pack typescript@6.0.3 --pack-destination /tmp/library-array-tsc > /tmp/library-array-npm.log 2>&1
tar -xzf /tmp/library-array-tsc/typescript-6.0.3.tgz -C /tmp/library-array-tsc
node docs/library-array/check-refusals.cjs > /tmp/library-array-tsc.stdout.log 2> /tmp/library-array-tsc.log
```

The classifier resumes `/tmp/library-array-classified.jsonl` if it exists. For a fresh classification, archive that result file first. The raw JSONL contains real TypeScript diagnostic messages and Adamic diagnostics for every attempted adapted input. `refusals.json` is the per-refusal ledger, including matching diagnostic codes and the adapted source SHA256. Bucket and feature grouping uses the first reported Adamic blocker, as described in `refusals.md`.

Measure the current implementation from its branch:

```sh
go run ./cmd/adamic-test262 -adapt -json -test262 /workspace/test262 -work /tmp/library-array-complete built-ins/Array > /tmp/library-array-record.json 2> /tmp/library-array-record.log
go test ./internal/lower -run TestLibraryArray -count=1 -v -timeout 10m > /tmp/library-array-tests-final.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_' -count=1 -timeout 30m > /tmp/library-array-oracles-final.log 2>&1
go vet ./... > /tmp/library-array-vet-final.log 2>&1
TMPDIR=/tmp/adamic-array-gate go test -count=1 -timeout 30m ./... > /tmp/library-array-gate-final.log 2>&1
TMPDIR=/tmp/adamic-array-gate go test -count=1 -timeout 30m ./internal/lower ./internal/flow ./internal/oracle > /tmp/library-array-gate-guard-final.log 2>&1
```

Create `/tmp/adamic-array-gate` with mode 1777 before the full gate: input tests drop privileges when running as root. The negative crash fixture is in `internal/oracle/testdata/library_array_refused/`, following the existing refusal-fixture convention, and `TestLibraryArrayHeterogeneousCrashIsRefused` reads and checks it directly. No runner or shared oracle registry changes are needed.

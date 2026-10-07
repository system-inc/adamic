# Commands run

Run from /workspace/adamic unless the block says otherwise. Logs are under
/tmp/library-array-coverage; final evidence is copied into this directory.

## Setup and branch

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
git fetch origin main codex/library-array
git fetch origin refs/heads/codex/library-array:refs/remotes/origin/codex/library-array
git switch -c coverage/library-array origin/codex/library-array
git log --oneline origin/main..origin/codex/library-array
git log origin/main..origin/codex/library-array --format=fuller
git diff --stat origin/main...origin/codex/library-array
git diff origin/main...origin/codex/library-array
bash cloud/setup.sh > /tmp/library-array-coverage/setup.log 2>&1
```

The first setup warmed its cache while the branch checkout was changing and failed to compile
some lower/flow tests because it saw new dispatch hooks before the new helper files. Repeating
setup on the stable checkout passed. The first fetch did not create the remote-tracking branch
under the checkout's configured refspec; the second fetch used an explicit destination.
CLAUDE.md, README.md, docs/0.1.md, docs/memory.md, the changed lowering and oracle files,
and the branch's existing report/reproduction/refusal evidence were inspected before writing.

## Oracle and counts

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_' -count=1 -v -timeout 30m > /tmp/library-array-coverage/oracle-first.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_' -count=1 -v -timeout 30m > /tmp/library-array-coverage/oracle-second.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_' -count=1 -v -timeout 30m > /tmp/library-array-coverage/oracle-third.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_shapes.a' -count=1 -v -timeout 30m > /tmp/library-array-coverage/oracle-shapes.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_' -count=1 -v -timeout 30m > /tmp/library-array-coverage/oracle-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_' -count=1 -v -timeout 30m > /tmp/library-array-coverage/oracle-final-bounded.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/library-array-coverage/counts.log 2>&1
```

The first two passes rejected probes at type checking/lowering; they were revised to supported
forms, retaining minimal refusal evidence. The shapes-only pass refused a spread-created object
isArray operand. The initial final pass included an unbounded huge sparse Node search. Its Node
processes were identified with ps and stopped with `kill 4201 4642` (4201 had already exited).
The final bounded oracle pass is the retained seven-program pass.

## Standalone builds

The loop below ran on the final seven programs. An initial loop without set -e ran the initial
probes and recorded their refusals. A second loop on the six corrected programs plus the first
shape probe was stopped during the unbounded Node execution. The final loop completed.

```sh
source /workspace/adamic-tools/env.sh
set -e
for file in internal/oracle/testdata/library_array_coverage_*.a; do
 name=$(basename "$file" .a)
 cp "$file" "/tmp/library-array-coverage/$name.mts"
 node --disable-warning=ExperimentalWarning "/tmp/library-array-coverage/$name.mts" > "/tmp/library-array-coverage/$name.node" 2> "/tmp/library-array-coverage/$name.node.err"
 go run ./cmd/adamic build "$file" -o "/tmp/library-array-coverage/$name" > "/tmp/library-array-coverage/$name.build.log" 2>&1
 "/tmp/library-array-coverage/$name" > "/tmp/library-array-coverage/$name.native" 2> "/tmp/library-array-coverage/$name.native.err"
 diff -u "/tmp/library-array-coverage/$name.node" "/tmp/library-array-coverage/$name.native" > "/tmp/library-array-coverage/$name.diff"
 test ! -s "/tmp/library-array-coverage/$name.node.err"
 test ! -s "/tmp/library-array-coverage/$name.native.err"
 echo "$name agrees"
done
go run ./cmd/adamic build notes/library-array/refused_bounds.a -o /tmp/library-array-coverage/refused_bounds > /tmp/library-array-coverage/refused_bounds.log 2>&1
go run ./cmd/adamic build notes/library-array/refused_is_array_view.a -o /tmp/library-array-coverage/refused_is_array_view > /tmp/library-array-coverage/refused_is_array_view.log 2>&1
```

## Gate

```sh
source /workspace/adamic-tools/env.sh
gofmt -l cmd internal > /tmp/library-array-coverage/gofmt.log
go vet ./... > /tmp/library-array-coverage/vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/library-array-coverage/gate.log 2>&1
```

## Mutation

```sh
git worktree add --detach /tmp/library-array-coverage-mutant HEAD
rmdir /tmp/library-array-coverage-mutant/cohere
ln -s /workspace/adamic/cohere /tmp/library-array-coverage-mutant/cohere
cp internal/oracle/testdata/library_array_coverage_identity.a /tmp/library-array-coverage-mutant/internal/oracle/testdata/library_array_coverage_identity.a
```

Python inserted the new identity fixture's registration into the scratch oracle_test.go and
changed exactly one implementation line: `ir.BooleanConstant{Value: isArray}` to
`ir.BooleanConstant{Value: !isArray}` in internal/lower/library_array_is_array.go.

```sh
source /workspace/adamic-tools/env.sh
cd /tmp/library-array-coverage-mutant
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_identity.a' -count=1 -v -timeout 30m > /tmp/library-array-coverage/mutant.log 2>&1
git diff -- internal/lower/library_array_is_array.go > /tmp/library-array-coverage/mutant.diff
```

Python restored that exact line, then:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_identity.a' -count=1 -v -timeout 30m > /tmp/library-array-coverage/restored.log 2>&1
git diff --exit-code -- internal/lower/library_array_is_array.go
```

## Final scalar cases

After the full gate had completed its oracle package, the final identity cases were added.
These commands rechecked the completed seven-program set and refreshed its identity row:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/library_array_coverage_' -count=1 -v -timeout 30m > /tmp/library-array-coverage/oracle-final-facts.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/library-array-coverage/counts-final.log 2>&1
```

The standalone loop's identity-file commands were rerun separately after that count update,
using the same output paths and checking identical stdout and empty stderr.

## Commit and push

```sh
git diff --exit-code origin/codex/library-array -- internal/lower internal/native internal/javascript
git diff --cached --check
git add internal/oracle/oracle_test.go internal/oracle/counts.md internal/oracle/testdata/library_array_coverage_*.a notes/library-array
git add -f notes/library-array/*.log
git commit -m "Add oracle coverage for the library Array branch"
git push -u origin coverage/library-array
```

The stored mutant patch omits unchanged context (`-U0` form) to avoid whitespace warnings
when a patch containing a space-prefixed tab-indented context line is itself added as a file.

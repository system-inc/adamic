# Verification

Five added programs agree with Node. No differing program was retained.
The focused uncached oracle checked source Node, generated JavaScript, sanitized
native, release native, and a separate LeakSanitizer run for every program.
Each also built and ran with the CLI; cmp found every complete stdout identical.
Counts regeneration added exactly five rows and moved no existing row.
Formatting and vet passed. The restored full uncached repository gate exited 0;
all packages passed (full-gate.txt). Native took 264.314s, oracle 208.243s,
and the longest package, Unicode properties, 721.374s.

## Mutation

Changed internal/native/runtime/integer.h:28 from
`return remainder == 0 ? copysign(0.0, left) : (double)remainder;`
to `return remainder == 0 ? 0.0 : (double)remainder;`.
Ran only integer_coverage_remainder.a with the uncached oracle. Test exit 1:
stdout differed. Both executions exited 0, with empty stderr, and the leak check
passed. The first Node line was `target -2147483648 0 true`; native printed
`target -2147483648 0 false`. Full outputs are in mutant.txt. Restored the exact
original line and verified git diff for integer.h was empty before the full gate.
This is a deliberate mutant, not a discrepancy in the unmodified feature branch.

## Commands

Commands below ran from /workspace/adamic. Test output went to logs, never into
a truncating pipe. Source file names use the integer_coverage_ prefix.

```bash
git fetch origin main codex/integer-fast-paths
git fetch origin codex/integer-fast-paths:refs/remotes/origin/codex/integer-fast-paths
git log --format='%h %s%n%b' origin/main..origin/codex/integer-fast-paths
git diff origin/main...origin/codex/integer-fast-paths
git switch -c coverage/integer-fast-paths origin/codex/integer-fast-paths
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/integer_coverage_' -count=1 -v -timeout 10m > /tmp/adamic-gate/integer-coverage.log 2>&1
mkdir -p /tmp/adamic-gate/integer-standalone
for file in internal/oracle/testdata/integer_coverage_*.a; do
 name=$(basename "$file" .a)
 go run ./cmd/adamic build "$file" -o "/tmp/adamic-gate/integer-standalone/$name" > "/tmp/adamic-gate/integer-standalone/$name.build.log" 2>&1 || exit 1
 node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" > "/tmp/adamic-gate/integer-standalone/$name.node"
 "/tmp/adamic-gate/integer-standalone/$name" > "/tmp/adamic-gate/integer-standalone/$name.native"
 cmp "/tmp/adamic-gate/integer-standalone/$name.node" "/tmp/adamic-gate/integer-standalone/$name.native" || exit 1
 echo "$name agrees"
done
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/adamic-gate/integer-counts.log 2>&1
gofmt -l cmd internal > /tmp/adamic-gate/integer-gofmt.log
go vet ./... > /tmp/adamic-gate/integer-vet.log 2>&1
# With the one-line mutation described above:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^integer_coverage_remainder.a$' -count=1 -v -timeout 10m > /tmp/adamic-gate/integer-mutant.log 2>&1
# After restoration:
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/adamic-gate/integer-full-gate.log 2>&1
for file in internal/oracle/testdata/integer_coverage_*.a; do
 name=$(basename "$file" .a)
 go run ./cmd/adamic c "$file" > "/tmp/adamic-gate/integer-standalone/$name.c" || exit 1
done
git diff --check
```

Generated C confirms calls.a reaches adamic_bitwise_and/or/xor/not,
adamic_shift_left/right/right_unsigned and adamic_remainder. Nested.a reaches
adamic_to_uint32, adamic_signed_bits and adamic_shift_right_bits, with arithmetic
barriers returning to double evaluation.

The first setup attempt overlapped the branch switch and failed while compiling
files removed by that switch. Setup was rerun on the stable checkout and passed.
An initial oracle invocation before fixture registration selected no programs;
it is not counted as coverage. The later oracle.txt names all five programs.

Setup timing: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build cache warm 73s, done in 73s. nproc=5, cgroup cpu.max=400000 100000.
Go 1.27.1, clang 20.1.8, Node v24.19.0, Linux amd64. Full timing lines are in setup.txt.

Commit and publication commands:

```bash
git add internal/oracle/oracle_test.go internal/oracle/counts.md internal/oracle/testdata/integer_coverage_*.a notes/integer-fast-paths
git diff --cached --check
git commit -m "Add oracle coverage for integer fast paths"
git push -u origin coverage/integer-fast-paths
git status --short
git rev-parse HEAD
git ls-remote origin refs/heads/coverage/integer-fast-paths
```

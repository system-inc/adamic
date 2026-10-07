# Commands and observations

Repository: /workspace/adamic. All Go/Node commands below source
/workspace/adamic-tools/env.sh first (the setup used ADAMIC_TOOLS=/workspace/adamic-tools).
Logs and independent build outputs are under /tmp/adamic-gate on this worker.

```sh
git fetch origin main codex/string-build-speed
git fetch origin refs/heads/codex/string-build-speed:refs/remotes/origin/codex/string-build-speed
git log --oneline origin/main..origin/codex/string-build-speed
git log -3 --format=fuller origin/codex/string-build-speed
git diff origin/main...origin/codex/string-build-speed
git switch -c coverage/string-build-speed origin/codex/string-build-speed
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
```

Setup printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (95s)
setup: done in 95s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc printed 5. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/string_build_' -count=1 -v > /tmp/adamic-gate/string-build-oracle.log 2>&1
# Repeated after adding all-unit codePointAt observations and the remaining cached-read guards:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/string_build_' -count=1 -v > /tmp/adamic-gate/string-build-oracle-final.log 2>&1
# Sixth program added afterward:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/string_build_calls.a$' -count=1 -v > /tmp/adamic-gate/string-build-calls.log 2>&1
```

All passed. The harness compares Node source, JavaScript backend, sanitized native,
and release native; programs that finish also run its leak check.

The first five were built and compared twice, before and after the expanded
observations. This is the final loop, with calls subsequently built individually:

```sh
mkdir -p /tmp/adamic-gate/string-build
for file in internal/oracle/testdata/string_build_{padding,join,repeat,cached_reads,boundaries}.a; do
 name=$(basename "$file" .a)
 go run ./cmd/adamic build "$file" -o "/tmp/adamic-gate/string-build/$name" || exit
 "/tmp/adamic-gate/string-build/$name" > "/tmp/adamic-gate/string-build/$name.native"
 node --disable-warning=ExperimentalWarning oracle/node.mjs "$PWD/$file" > "/tmp/adamic-gate/string-build/$name.node"
 diff -u "/tmp/adamic-gate/string-build/$name.node" "/tmp/adamic-gate/string-build/$name.native" || exit
 echo "$name: final native build and Node agree"
done
go run ./cmd/adamic build internal/oracle/testdata/string_build_calls.a -o /tmp/adamic-gate/string-build/string_build_calls
/tmp/adamic-gate/string-build/string_build_calls > /tmp/adamic-gate/string-build/string_build_calls.native
node --disable-warning=ExperimentalWarning oracle/node.mjs "$PWD/internal/oracle/testdata/string_build_calls.a" > /tmp/adamic-gate/string-build/string_build_calls.node
diff -u /tmp/adamic-gate/string-build/string_build_calls.node /tmp/adamic-gate/string-build/string_build_calls.native
```

All six exited zero and their complete stdout files matched byte for byte.

A Python script changed the single line described in coverage.md, ran the
following command with stdout/stderr redirected to string-build-mutant.log,
asserted the nonzero test exit, and restored the file in finally:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/string_build_cached_reads.a$' -count=1 -v
```

It failed by stdout comparison with the C compiling successfully. The restored
program then passed the final uncached oracle above. The mutation is not a
branch disagreement and is not counted as a differing program.

```sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
# Run after the initial five, after their expanded observations, and after calls.
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/adamic-gate/string-build-gate.log 2>&1
git diff --check
```

Counts updates, formatting, vet and diff checks passed. Formatting printed no
paths. Only the six new counts rows changed.

The cloud environment interrupted the full gate before its final results were
reported. The saved log records passing results through stage1/cohere/lint,
including internal/native, internal/oracle and all earlier packages. Go buffers
later package results in package order; the missing results were recovered with:

```sh
go test -count=1 -timeout 30m ./stage1/cohere/typeaware > /tmp/adamic-gate/string-build-typeaware-recovery.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/markdownblocks/... ./stage1/cohere/markdowninline ./stage1/cohere/mediaquery ./stage1/cohere/selector ./stage1/cohere/suppression ./stage1/cohere/values ./stage1/typescript/parser ./stage1/typescript/scanner > /tmp/adamic-gate/string-build-remaining-recovery.log 2>&1
```

The typeaware tests do not use oracle result caching.

Recovery results: typeaware passed (732.806s). Every package in the remaining
recovery run passed except TestMarkdownUnicodeWidths, which failed because the
scratch npm width dependencies were absent, before any width comparison ran.
Installed the exact versions named in widthTables.ts and REPORT.txt, then ran
only that failed check:

```sh
npm install --prefix /tmp/adamic-markdown-width --ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0 narrow-emojis@0.0.3
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$' > /tmp/adamic-gate/string-build-width-recovery.log 2>&1
```

That test passed (139.357s). Every package/check in the full gate has a passing
result across the initial run and these recovery runs. The original full-gate
invocation itself was interrupted, not a single uninterrupted green run.
No compiler changes or test skips were needed.

Publication commands:

```sh
git commit -m 'Cover string building and cached UTF-16 reads in the oracle'
git push -u origin coverage/string-build-speed
git commit -m 'Record gate recovery and pinned dependency checks'
git push origin coverage/string-build-speed
```

The first push succeeded without a retry.

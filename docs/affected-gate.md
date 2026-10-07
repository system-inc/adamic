# Affected gate input records

Status: implementation and small real-process proofs, not a proven integration gate.
The five full-repository branch proofs and the affected-versus-whole timing comparison
are blocked by the observer's incompatibility with LeakSanitizer. Do not use this
prototype to omit integration packages until those proofs have been completed.

`cmd/adamic-affected` records inputs and selects packages. It does not reuse test
results. Every recorded test process runs with `ADAMIC_GATE_UNCACHED=1`; selected
packages must likewise run with that variable and `go test -count=1 -timeout 30m`.
The existing sharded gate was not edited.

Build the command from the worker branch into a path outside the repository:

```sh
source /workspace/adamic-tools/env.sh
go build -o /tmp/adamic-affected ./cmd/adamic-affected
```

At a clean, green main checkout, with `HEAD = origin/main` and initialized submodules:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -json ./... > /tmp/main-gate.jsonl 2> /tmp/main-gate.stderr
ADAMIC_GATE_UNCACHED=1 /tmp/adamic-affected record -out /tmp/main-inputs.json > /tmp/record.stdout 2> /tmp/record.stderr
```

Keep the record and its `.logs` directory together with integration's gate output.
They are not repository files. A failed test, failed tracer, dirty checkout, changed
input or changed toolchain prevents publishing a new record. A leak-free sanitizer
control checks observer compatibility before tracing Adamic packages. An older destination
is not overwritten on failure, so integration must check the command's exit status
and the record's commit before publishing it as the new green-main artifact.

On a branch:

```sh
ADAMIC_GATE_UNCACHED=1 /tmp/adamic-affected select -record /tmp/main-inputs.json > /tmp/selected-packages.txt 2> /tmp/select.stderr
```

Standard output contains one Go import path per selected package, sorted. No output
means no selected package only when the command exited successfully. A missing,
unreadable, malformed or unknown-version record selects all current packages. An
unreadable current package graph exits unsuccessfully: it cannot justify any skip.
A new package, changed static closure, changed observed input, incomplete trace,
Git-metadata read, undecoded multi-path operation or repository CGo dependency
selects that package. A changed
submodule or toolchain identity selects everything.

## Inputs

For each current package the command asks `go list -deps -test -json` for the actual
Go dependency graph, including its internal and external tests. It hashes Go files,
CGo files, embedded files, test files, C/C++ sources, headers, assembly and object
files, plus every dependency directory's complete `testdata` tree and module file.
It also records root `go.mod`, `go.sum`, `go.work`, `go.work.sum`, `.gitignore` and
`.gitmodules`, including their absence. Selection recomputes the static closure, so
adding a source file or an embedded file cannot hide behind the old filenames.
Generated test-main files returned as absolute paths stay absolute.

Observed inputs are collected separately for each test binary and all of its children:

```sh
strace -ff -yy -s 0 -e trace=%file,getdents64 -o PACKAGE.trace \
  go tool test2json -t -p IMPORT_PATH PACKAGE.test \
  -test.v=test2json -test.timeout=30m
```

`-ff` gives one log per process without interleaved syscall fragments. `-yy` resolves
cwd and directory descriptors, including after fork or chdir. `%file` includes
`open`, `openat`, `execve`, old and modern stat variants, access and failed existence
probes. `getdents64` supplies directory listings, which the requested four-syscall
filter alone would miss. Directory fingerprints include names and entry types;
file fingerprints include mode and bytes, and symbolic links include their target
and the target's fingerprint. Missing paths and their parent directories are explicit inputs. Undecodable paths,
truncated records and unsupported input types prevent skipping the package. Writes
into repository inputs also prevent skipping it. Syscall logs do not request expanded
`execve` environments; the environment identity stores a digest rather than values.

CGo source lists do not establish a complete closure for arbitrary `#include` paths
used while building a cached binary. Repository CGo dependencies therefore force
selection until their build input tracing is proven as well.

The shared submodule identity includes cohere's HEAD, its binary working-tree diff
against HEAD, recursive submodule commit status, and every non-Git working-tree
file and directory below cohere. This includes staged, unstaged, untracked and nested
working-tree bytes; it intentionally selects more packages than the minimum when
cohere changes.

Toolchain identity contains `go version`, full `clang --version`, `node --version`,
stable Go configuration and the invoking environment's digest. The audit found
literal and dynamic `os.Getenv` calls and `os.Environ` forwarded to children, so
hashing only a list of literal names would omit inputs. Stable Go configuration
excludes `GOGCCFLAGS`, whose generated temporary prefix changes on each `go env`
invocation. No environment values are written into the JSON record.

This first implementation keeps absolute external Go input paths and exact environment
identity. Changing checkout location, `PWD`, cache paths or workspace configuration
can select everything. That conservative behavior is observed in the design, not
claimed as a worktree speedup. External corpus/library directories named by environment
variables require an additional closure audit before production use. The parser records
repository paths, not arbitrary external data or network responses.

## Tracer blocker, observed on this box

A standalone leak-free C program, compiled with `clang -fsanitize=address,undefined`,
exited 0 normally and 1 under the requested `strace -f` filter. Its output was:

```text
LeakSanitizer has encountered a fatal error.
HINT: LeakSanitizer does not work under ptrace (strace, gdb, etc)
```

The actual repository test `TestFreedValuesAreCaughtWithSlabs` likewise passed
normally and failed under tracing, for both clean control binaries. These checks
were left enabled. The recorder refuses a failed traced package rather than turning
that failed observation into a green record.

The environment has no effective capabilities, no mounted tracefs, and no installed
`bpftrace` or `perf`. A non-ptrace observer needs additional platform support and a
separate proof of PID-tree attribution, directory listings, failed probes, path
resolution, and lost-event detection. Merely running once without tracing and once
with leak checks disabled would not establish that both runs opened the same files.

## Evidence and mutants

Fixed main: `e011f8f60899586d6373a5ccb07335ad82cfbf3c`.
All changes are confined to `cmd/adamic-affected/` and this document.

The whole-test-file path audit is `/tmp/affected-path-audit-all.log`. It contains
4,488 matches, 1,702 outside the cohere submodule, at the time of the audit. Examples
that defeat a package-directory-only model include native tests reading
`bench/regex/cases.json`, lowering tests reading oracle fixtures, flow/fresh tests
globbing oracle directories, stage1 CSS tests reading sibling slices, markdown tests
reading markdowninline, and scanner tests reading cohere's generated identifier tables.
The environment audit also includes optional TypeScript corpora and library paths.

The real process-tree proof builds a small uncached Go package outside the repository,
traces it and its Node child, then changes real files and runs the same test binary
again without tracing. This is a controlled observer proof, not one of the five
requested full-main branch proofs.

| Mutant | Observed bad decision | Independent check that catches it |
|---|---|---|
| Remove observed inputs | Relative fixture change wrongly skipped | Fresh Go test and Node child both fail on changed fixture bytes |
| Remove directory inputs | New fixture wrongly skipped | Fresh test fails because enumeration finds two files instead of one |
| Compile a selector without toolchain comparison, then change the recorded Node version | Both packages wrongly skipped | Independent `node --version` requires both packages; the original selector prints both import paths |

A supplemental real-repository mutant used the clean `/tmp/affected-main` worktree
at the fixed SHA. A deliberately static-only record for `internal/native` omitted
`bench/regex/cases.json`; adding invalid JSON bytes made `select` wrongly omit that
package. The same untraced `TestRegExpLintPatternsNode` binary first passed with
102 patterns and 890 inputs, then failed with `invalid character 'n' after top-level
value`. The fixture was restored and the worktree returned clean. This was an
isolated synthetic mutant record, not a certified whole-main record. Logs are
`/tmp/affected-real-relative-proof.log`, `/tmp/affected-real-relative-baseline.log`
and `/tmp/affected-real-relative-mutant.log`.

The controlled CLI test creates a clean Git main and a real submodule, records both
packages, skips an unchanged checkout and unread documentation, selects the reader
when its relative fixture changes, and selects everything when the Node version in
the record changes or the record is missing. Additional tests cover missing paths,
file modes, symbolic-link target bytes, recursive testdata, and trace uncertainty.
The toolchain mutant is a separately compiled command with the toolchain comparison
removed. The test changes Node's version string in the record, runs that mutant and
observes an empty selection; the independent actual Node version requires both
packages to run. This is not an altered Node installation or a completed whole-Adamic
branch proof. Observed-input and directory-input mutants also run through the real
CLI; a fresh `go test -count=1 ./pkg` catches each wrongful skip.

Validation commands and complete logs:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -v ./cmd/adamic-affected > /tmp/affected-command-tests.log 2>&1
go vet ./cmd/adamic-affected > /tmp/affected-vet.log 2>&1
gofmt -l cmd internal > /tmp/affected-gofmt.log
go vet ./... > /tmp/affected-full-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/numbers.a$' > /tmp/affected-filtered-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -json ./... > /tmp/affected-main-gate.jsonl 2> /tmp/affected-main-gate.stderr
```

Other evidence: `/tmp/affected-setup.log`, `/tmp/affected-lsan-plain.log`,
`/tmp/affected-lsan-traced.log`, `/tmp/affected-native-plain.log`,
`/tmp/affected-native-traced.log`, and `/tmp/affected-native.trace`.

Setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build cache warm 280s, done 280s. `nproc` was 5; cgroup `cpu.max` was
`400000 100000`. Go was 1.27.1 linux/amd64, clang was 20.1.8, Node was v24.19.0.
Setup overlapped the initial full-gate invocation, so that invocation is validation,
not a clean performance measurement. Setup build flags: main commit above,
`go build ./...`, `go test -count=1 -run '^$' ./...`, build cache warmed, no fixture
result-cache run. Load averages immediately before and after setup were not recorded.
The script-reported setup duration is not a paired benchmark. No affected-only
speedup is claimed.

| Loop | Before | After | Instrument |
|---|---|---|---|
| Toolchain setup | Not measured | 280s, reported by script | `bash cloud/setup.sh > /tmp/affected-setup.log 2>&1` |

## Uncompleted integration proof

No full-main input record was certified. Therefore none of the five scratch branch
selection sets, skipped-package event comparisons or paired timing measurements
is reported as passing. Their status is explicit below; the instrument column gives
the required command shape, not a command claimed to have completed.

| Loop | Before: whole gate | After: affected gate | Instrument |
|---|---|---|---|
| stage3-only | Not paired | Blocked by uncertified record | `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m -json ./...` versus the same invocation on selected import paths |
| docs-only | Not paired | Blocked by uncertified record | Same whole and selected commands |
| one stage1/cohere slice | Not paired | Blocked by uncertified record | Same whole and selected commands |
| native runtime C | Not paired | Blocked by uncertified record | Same whole and selected commands |
| oracle fixture | Not paired | Blocked by uncertified record | Same whole and selected commands |

Each eventual measurement must log the fixed commit, `nproc`, cgroup quota, versions,
full build flags, load before and after, exact command and uncached status. Run on the
same box, interleaved, best of three unless a run exceeds five minutes.

Literal byte equality of raw `go test -json` streams also cannot be assumed: timestamps,
elapsed durations and parallel scheduling vary even for identical inputs. A future
proof must specify which test events are compared, preserve all test names and verdicts
and genuine test output, and make any normalization explicit. This implementation
keeps original logs and does not claim a normalized comparison as raw byte equality.

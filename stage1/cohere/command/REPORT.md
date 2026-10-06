Command front door, branch codex/stage1-command. Full CLI execution parity remains
open; the deterministic front door and project planning are ported and compared.
See [GAPS.md](GAPS.md) for the exact boundary and proving programs.

Started at origin/main 5d4c8012a0877094134e6c6bac367ff68f9313e8 and fast-forward
merged codex/stage1-config at 7ccbcd17e97d0bd74af28149aee8f23539d44b8f.
Cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db; TypeScript at
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. No protected compiler file was edited.
An initial merge could not find the remote-tracking config branch; fetching
explicit main/config refspecs fixed it before any source work.

Setup: bash cloud/setup.sh succeeded. Its timing lines were:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (13s)
setup: done in 13s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc printed 5. Go 1.27.1, clang 20.1.8, Node 24.19.0. Each command sources
/workspace/adamic-tools/env.sh; tests write logs rather than piping their output.

Built:

- flags.ts and definitions.ts: all 31 main flags and five rename flags; Go flag
  parsing, help/defaults, exact errors and int64 normalization without rounding.
  testdata/definitions.go extracts declarations from pinned Go source. It also
  supplies an independent Go flag oracle for parsed values and visited flags.
- location.ts and discovery.ts: root precedence, marker selection, config-path
  anchoring, root notes and repository discovery eligibility/root selection.
- scope.ts: named-path enumeration, Unicode byte order, links and duplicates,
  population/formatter narrowing, descriptions and Go value-receiver semantics.
- ownership.ts: reference expansion, solutions, nearest owners, all-file yields,
  shared-directory cache decisions and child lint-config argument rewriting.
- main.ts: deterministic early CLI responses and an explicit execution gap.
  Host adapters bind argv[0] and the typed exit result; they contain no flag,
  enumeration or settings logic. Ordinary adamic build is not a standalone CLI.

The Go overlay renames production helper implementations and wraps their entry
points. It runs 24 original command test functions with their original assertions,
comparing the port while each original fixture still exists. Native under
ASan/UBSan/LeakSanitizer, Node source and the JavaScript backend must produce the
same canonical stdout, empty stderr and exit 0. Returned fields, descriptions,
ordered file names, ownership, yields and original scope descriptions are held.
The Go checkout itself is never modified by this instrumentation.

Additional fixtures cover linked/dangling directories, Unicode ordering,
duplicate named files, gitignored files within an explicitly named directory,
root-dot short circuiting, reference cycles/missing files/dependency directories,
and escaped references keys. Config and formatter walks retain their separate
contracts. Full upstream checker/formatter/process CLI tests are outside the
ported boundary; they are not reported as passing native tests.

36 deterministic actual CLI cases compare separate stdout, stderr and exit status
against a freshly built Go cohere, on native, Node source and the JS backend.
They include help, syntax/unknown/missing flags, invalid boolean/integer values,
verbose/json and rename's help/count/write contradictions. Another 74 cases
compare every parsed value, flag presence and positional argument with Go flag,
including failed setters, int64 endpoints, overflow/syntax/underscore precedence,
explicit false, positional stopping and empty-program-name help.

Five executable mutants, each run on native, Node and the backend:

| Mutant | Comparison that catches it |
| --- | --- |
| Swift wins when both root markers exist | Original Go root tests, engine/config fields |
| Dot at a subdirectory becomes the whole project | Original root/scope tests, files and Everything |
| Narrowing changes the original description | Original scope tests, unchanged input description |
| Shared-owner tie goes to the last includer | Original ownership test, file yields |
| A bare boolean sets false | Go flag oracle, parsed types value |

The original copy mutant inferred a this type and was refused with
"stage 0 can't lower a value of type this yet". It is excluded as evidence.
The corrected mutant changes the original description while retaining the
correct returned description; it executes cleanly and fails specifically on
the new original-description comparison. The mutation suite checks a baseline
first. Compiler rejection, sanitizer failure or a process error never counts as
an output-comparison kill.

Commands and observations:

```sh
go test -count=1 -timeout 15m -v ./stage1/cohere/command > /tmp/stage1-command-release.log 2>&1
go test -count=1 -timeout 15m ./stage1/cohere/config ./stage1/cohere/formatfiles ./stage1/cohere/gitignore > /tmp/stage1-command-dependencies.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 15m -v ./internal/oracle -run '^TestRealPathAgreesWithNode$' > /tmp/stage1-command-oracle.log 2>&1
go vet ./... > /tmp/stage1-command-vet-final.log 2>&1
gofmt -l stage1/cohere/command stage1/cohere/command/testdata/boundaries.go.txt > /tmp/stage1-command-gofmt.log
go -C cohere run ./command/cohere --format-only --no-cache --directory /workspace/adamic stage1/cohere/command > /tmp/stage1-command-format-final.log 2>&1
ADAMIC_COMMAND_TIMING=1 go test -count=1 -timeout 15m -v ./stage1/cohere/command -run Original > /tmp/stage1-command-timing.log 2>&1
```

Final command gate PASS in 142.981 s: 36 actual CLI cases, 74 parsed cases,
24 original Go command helper tests, generated edges, three gap proofs and all
five clean executable mutants. Dependency suites passed: config 319.335 s,
formatfiles 382.540 s, gitignore 40.617 s while run together. Uncached realpath
oracle passed in 0.672 s, including its two runtime mutants. Vet and gofmt produced
no findings. Cohere formatted eight new TS files, then the final flag change.
The full repository gate was not run; no compiler implementation changed.

Timing on a 512-file named-directory fixture, with ordered-output equality checked:

| Round | Go scope plus canonical output | Sanitized native plus canonical output and process startup |
| --- | --- | --- |
| 0 | 0.003500 s, 146,295 files/s | 0.031616 s, 16,194 files/s |
| 1 | 0.004065 s, 125,942 files/s | 0.030678 s, 16,690 files/s |
| 2 | 0.004703 s, 108,864 files/s | 0.031291 s, 16,363 files/s |

These are different timing boundaries: Go stays inside the test process; native
starts a new sanitized process. Both serialize the full scope. No claim about
unsanitized enumeration speed or whole-engine performance follows from them.

Gap evidence: Node's child-process fixture prints 0; Adamic refuses the missing
node:child_process binding with TS2591. Node's process-status fixture exits 7;
unadapted native panics with exit 70. On a one-file strict TypeScript project,
Go --types --no-cache --directory ROOT exits 0, while this driver deliberately
prints "cohere: stage 1 command execution is not yet ported" and exits 1.
The graph/checker and full process runner must land before full CLI parity can be
claimed. CohereSettings remains strict JSON and tsconfig remains JSONC.

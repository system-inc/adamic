Numeric process.exitCode and process.exit are compiled, with Node 24's chosen-status semantics.
Stdout/stderr terminal observations and read-only environment lookup are compiled.
Seven fixtures, eight exit-argument cases and twenty terminal/environment cases agree with Node.
Five executable mutants were caught by Node comparisons; the compiler branch is pushed.
The lint output port follows on top; immediate-exit pipe buffering remains a stated limitation.

Base: origin/main, 5d4c8012a0877094134e6c6bac367ff68f9313e8.
Implementation: fd2ef77, on codex/process-exit-and-tty. No pull request.

Setup command: bash cloud/setup.sh, output /tmp/process-exit-setup.log.
Timing lines: setup go: 0s; setup clang: 0s; setup node: 0s; setup submodules: 0s;
setup build warm: 83s; setup done: 83s. nproc: 5; cgroup CPU quota: 4.
Environment: source /workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node 24.19.0.

Validation, with output written to logs:

- go vet ./...: exit 0, /tmp/process-vet.log.
- ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 10m ./internal/oracle
  -run '^TestNativeAgreesWithNode/internal/oracle/testdata/process_': PASS, 4.109s,
  /tmp/process-fixtures-final.log. Source Node, JavaScript backend, sanitized native and
  release native agree on stdout, stderr and exit codes. Normal completion with status 259
  releases all 21 allocated values; the OS observes status 3. Uncaught throw remains 70.
- ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 10m ./internal/oracle
  -run '^TestProcess': PASS, 28.144s, /tmp/process-probes-complete.log. Includes all independent
  stdout/stderr PTY combinations, unset/empty/populated NO_COLOR and FORCE_COLOR, UTF-8 values,
  NUL names, missing values, and explicit NotYet for escaping the process objects/function.
- go test -count=1 -timeout 30m ./internal/oracle -run '^TestCountsAreRecorded$'
  -args -update-counts: PASS, 54.604s, /tmp/process-counts-controlled.log. Environment-dependent
  counts use fixed input values, not the caller's color preferences. Only seven new rows moved.
- go test -count=1 -timeout 30m ./...: attempted the complete gate, /tmp/process-gate.log.
  It failed only internal/fresh's unknown-IR-node registration check. Every other package
  passed, including internal/native 391.601s, internal/oracle 278.583s, stage1/cohere/lint
  174.330s, and internal/unicodeproperties 860.055s. The process node was then registered as
  scalar state with immutable string results and evaluated operand effects. The corrected
  go test -count=1 -timeout 10m ./internal/fresh passed in 59.239s, /tmp/process-fresh.log.
  The 14-minute complete gate was not repeated; the focused process oracle rerun was uncached.

Mutants in TestProcessMutants all compiled under the ordinary -Werror flags and ASan/UBSan,
ran without sanitizer failures, and reached the actual Node comparison:

| Mutant | What caught it |
| --- | --- |
| Return 0 at normal completion instead of the stored status | Exit codes differ |
| Immediate exit uses 0 instead of its argument | Exit codes differ |
| Missing environment values become empty strings | Stdout differs |
| Nonterminal streams report true | Stdout differs |
| Setter silently ignores fractional/nonfinite codes | Stdout differs |

Node first revealed that process.exit(undefined) means 0 while process.exit() preserves exitCode.
The failed fixture comparison caused the implementation correction, and both forms now have separate
oracle cases. A pre-existing console recognizer also needed symbol.Name == console: declaring process
in the prelude exposed its old assumption that every prelude object was console.

Scope and limits are in [process.md](process.md). Numeric strings and escaping external objects are
excluded explicitly. Immediate exit skips finally and frame cleanup, including LeakSanitizer's exit
hook. Large asynchronous pipe writes on immediate exit were not held to Node; the renderer uses
exitCode with normal completion instead. Go cohere's character-device color test also differs from
Node isTTY on devices such as /dev/null. These observations are not a claim of complete Node process
or stream emulation.

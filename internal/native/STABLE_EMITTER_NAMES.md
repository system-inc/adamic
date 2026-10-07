Built function-local temporary and inline-cache counters, including region variants; this unit is partial.
Base f8013f0; merged emitter-speed 52f9bae; implementation commit is the commit adding this report.
Full uncached native and oracle pass (115.597s and 115.310s); vet and formatting pass; own diff checks pass.
Whole-program temporary/cache mutants fail the counter regression; stale region facts fail regions.a under ASan.
Not covered: stable module-qualified symbols, per-module main counters, split-mode oracle, or complete one-module acceptance.

The small emit.go hook is resetCounters, called before each function variant and main.
It resets temporaries, a separate cache counter, and regionValues. Region facts are
keyed by C names, so shortening their lifetime with the names is necessary. The first
full oracle exposed this dependency; simply resetting numbers without clearing facts
caused use-after-free and invalid C. That failed initial log remains at
/tmp/stable-emitter-native-oracle.log. The corrected full log is saved beside this report.
Cache declarations retain file scope and existing behavior; their names contain the
owning function and variant so local counter values cannot collide across bodies.
Main currently has one counter namespace, because IR merges all module-level code.

No field-store, switch or check-elision code was restructured. The completed
codex/emitter-speed branch was merged before testing. Its field-store changes are
not attributed to this unit. Current origin/main was fetched and merged before
landing; it was already included. No cohere file was copied or changed.

Observed lint unit changes using developer-tools splitter
14e8372816b1d3bf5bcc37ca57336e46d41b7a41, with the lint entrypoint
stage1/cohere/lint/main.ts and edits overlaid on
stage1/typescript/parser/grammar.ts:

| Edit | Unit bodies/state before | Unit bodies/state after | Header before/after | Effective invalidations before/after |
|---|---:|---:|---|---|
| Add isolatedProbe function before precedence | 29/29 | 29/29 | changed / changed | 29 / 29 |
| Add const isolatedProbe = [14].length in precedence | 20/29 | 19/29 | changed / changed | 29 / 29 |
| Replace return 14 with return [14].length | 20/29 | 1/29 | changed / unchanged | 29 / 1 |

The baseline includes the merged emitter-speed branch but lacks these counter changes.
The reporting probe uses load.LoadOverlay with absolute normalized filenames; original
port files are never edited. The probe extracts the unmodified splitter through splitC
from developer-tools into a temporary test file. It is saved as
stable_emitter_evidence/unit_changes_probe.go.txt, with both raw logs. It reports churn;
it is not an acceptance check asserting that all three edits change only one unit.
The counters_test.go regression compares the later function body, cache declaration,
and main across edits. It catches downstream text changes from either whole-program
counter. The requested complete splitter-based mutant acceptance is not claimed.

Why the remaining work needs a scope exception

ir.Program carries only the entry basename, functions carry only Name, and locals
carry whole-program IDs and their function ID. Module path, declaration ancestry,
anonymous-function ordinal within its parent, specialization identity, and the
boundaries between module-level Main statements are lost before native.C is called.
Native cannot recover identical names declared in different modules from that IR.
Generic lowering also puts a whole-program ordinal in Function.Name itself.

A minimal follow-up needs source identity metadata in internal/ir, hooks in
internal/lower/functions.go and local/closure/generic registration, and a module-range
bookkeeping hook in lower.go for Main. These are outside the user's native territory;
permission was requested before touching them. None was changed in this patch.

Separately, developer-tools' splitter groups every sixteen consecutive functions and
puts all prototypes and cache declarations in one shared header. Adding a function
necessarily changes that header and shifts groups. Stable names alone cannot make
that splitter invalidate only one module unit. Achieving the stated acceptance needs
module grouping and narrower declaration dependencies as well as stable local/global,
shape, method, class and literal identities. That requires coordination or explicit
scope expansion beyond renaming and counter allocation. The measured table shows
this remaining gap; this branch must not be treated as acceptance-complete.

Verification and reproducibility

Source /workspace/adamic-tools/env.sh before each command. Setup printed Go, clang,
Node and submodules ready at 1s; build cache warm and done at 97s. nproc=5,
cgroup cpu.max=400000 100000. The toolchain is Go 1.27.1, clang 20.1.8, Node 24.19.0.

Commands, with all test output saved to logs:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/stable-emitter-native-oracle-final.log 2>&1
go vet ./... > /tmp/stable-emitter-vet-final.log 2>&1
gofmt -l cmd internal
git diff 52f9bae --check
python3 internal/native/stable_emitter_evidence/reproduce.py > /tmp/stable-emitter-reproduce.log 2>&1
```

Own changes pass git diff 52f9bae --check. The merged emitter-speed raw logs contain
pre-existing diagnostic whitespace flagged by git diff origin/main --check; their
bytes were left as supplied by that worker.

The complete native and oracle packages pass, including counts, Node parity and
sanitizers. The full repository gate was not run. The developer-tools branch was
used as a measurement instrument, not merged, and split-mode execution of every
fixture was not verified. Counter changes do not justify claiming that unrun mode.

Mutants were applied independently through Go overlays, never left in production:

| Mutant | Command selection | Observed failure |
|---|---|---|
| Remove temporary reset | native TestFunctionCountersDoNotMove | second function temporaries move; exit 1 |
| Remove cache reset | native TestFunctionCountersDoNotMove | expected second-function cache absent; exit 1 |
| Remove regionValues reset | uncached oracle TestNativeAgreesWithNode/internal/oracle/testdata/regions.a$ | ASan heap-use-after-free; exit 1 |

Each mutant compiles its Go test binary. The region mutant reaches an executable;
ASan kills it, not clang warnings. Raw logs with diagnostic whitespace are preserved as gzip files. Saved logs and a reproduction runner are in
stable_emitter_evidence. The runner also repeats both churn measurements against
the exact baseline emitter files from 52f9bae. It removes only its own temporary
probe test file after execution. No PR was opened; only codex/stable-emitter-names
is the push destination.

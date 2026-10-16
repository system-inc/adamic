Built: step (1), a class-naming refusal for library storage spread through structural views, with explicit-field copying as the fix.
Commits: implementation bf2fa67219a0e88c357bb30c8b72046084c56cd8, based on main d79302e50c022c997c00f7c68ddb7475f9ee65fc; evidence is a following commit.
Commands and outputs: focused lowering and oracle pass; counts refresh unchanged; admission 1,356 files, 0 newly admitted and 2 newly refused.
Mutants: dropping the refusal fails the exact task fixture; dropping the plain-producer proof fails the record/class admission control.
Not covered: step (2), library enumerability; no whole-package suites, full gate, PR, or Loom submission.

The compiler waits until call targets are complete, then checks object spreads. Known plain-record and Adamic-class producers remain admitted, including arguments and returned values joined by the existing allocation-flow graph. Unresolved producers are compared with the closed world's library-origin instance types. Constructor objects are excluded. Identity comes from bundled library, prelude, and pinned Node declarations, rather than a list of class names. Non-object representations that were already rejected also receive the library-specific reason when applicable.

Conservative assumption, also recorded in the implementation commit: without enumerability metadata, a library-origin object shape compatible with an unresolved spread source is unsafe. This can refuse a program whose actual path only reaches records when the available producer proof cannot establish that fact. It does not infer enumerability or change runtime copying. Ordinary compiler-proven records and Adamic classes remain admitted.

The only imported draft file is `internal/oracle/testdata/coverage_regexp_view.a`, byte-for-byte from `77a1197104:internal/oracle/testdata/coverage_regexp_view.a`. It is also exactly the task's program, so there is one fixture rather than two duplicates. Runtime's unlanded draft has a copy that must agree once both land. No draft branch was merged. `TestRegExpViewSpreadIsRefused` checks source Node's exit 0, stdout `"\n"`, empty stderr, and this whole diagnostic, including location:

```text
internal/oracle/testdata/coverage_regexp_view.a:3:19: Adamic 0.1 refuses an object spread through a view that may hold library class RegExp (its internal fields are not own enumerable properties); copy the fields you mean by name
```

Admission comparison used main's exact `lower.go` and `object.go` through a Go overlay, with the newly added refusal file replaced by an empty package. Both binaries examined the same corpus, including the imported fixture: every `.a` and `.ts` file under `internal/oracle/testdata`, excluding declaration files, used as its own entry. Main admitted 1,138; final admitted 1,136. All 1,356 admission bits from the final compiler agree with the earlier implementation census. `before.jsonl`, `final.jsonl`, and `delta.json` hold the results.

Every newly refused program:

| Program | Why main's admission was unsound |
|---|---|
| `coverage_regexp_view.a` | The interface hides a RegExp. Node copies no keys; native copied its internal regex program and metadata. The task documents the native mismatch, and the fixture independently observed Node's empty line. |
| `review/agree/fxspptb_9984394_error_spread.a` | The Error has no own enumerable `name` or `message`. Node prints `no name no message`; a compiler built from main prints `Error bad 1`. Both outputs were observed here. Its existing `.pending` already identifies it as a silent miscompile. That other unit's fixture and marker were left unchanged. |

There are **0 newly admitted** programs. Two existing entries panic on both compilers with `Unhandled case in Node.Text: *ast.ComputedPropertyName`: `review/agree/fxspptb_iterators_derived_symbol.a` and `review/agree/fxspptb_iterators_override_source.a`. The first sequential census stopped there; the scanner was changed to record each panic as a non-admission and resumed, then the final compiler was scanned afresh. No compiler panic was fixed or hidden as an ordinary refusal.

Toolchain setup, with `GOPROXY='https://proxy.golang.org|direct'`: the resumed `timeout 300 bash cloud/setup.sh` reached submodules at 266.083s and shared cache at 272.239s, then expired during warming (exit 124). An explicit bounded shallow fetch of the pinned TypeScript commit succeeded; the submodule update had completed by then. `timeout 180 bash cloud/setup.sh` subsequently passed: Node 0.096s, Go 0.116s, Markdown dependencies 0.246s, submodules 0.340s, clang 0.900s, shared cache 19.224s, cache warm 153.325s, done 153.385s. Its printed environment file, `/workspace/adamic-tools/env.sh`, was sourced for builds and tests. `nproc` is 5; `cpu.max` is `400000 100000`, a 4-CPU quota. Go 1.27.1, clang 20.1.8, Node v24.19.0.

Final validation commands and observed outputs (all test output went directly to files):

```sh
# Lowering: each newly added test is a separate top-level leaf.
timeout 90 go test ./internal/lower -run '^TestLibrarySpread' -timeout 60s -count=1 -v
# PASS; package 0.320s

timeout 90 go test ./internal/oracle -run '^TestRegExpViewSpreadIsRefused$' -timeout 60s -count=1 -v
# PASS; leaf 2.92s, package 2.930s

# Precompiled oracle binary, from internal/oracle, only these four existing fixtures.
ADAMIC_GATE_UNCACHED=1 timeout 90 /tmp/regexp-spread-oracle.test \
  -test.run '^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/(spreads|plain_data_spread|class_as_interface|spread_undefined)\.a$' \
  -test.timeout 90s -test.parallel 4 -test.v
# PASS: plain_data_spread 5.52s, spreads 5.63s,
# spread_undefined 5.68s, class_as_interface 6.05s.
# Source Node, JS backend, release native, ASan, UBSan, and leak checks.

# Precompiled oracle binary, from internal/oracle, only the counts family.
timeout 300 /tmp/regexp-spread-oracle.test -test.run '^TestCountsAreRecorded' \
  -test.timeout 300s -test.parallel 4 -update-counts
# PASS; git diff --quiet -- internal/oracle/counts.md exits 0.
# The new fixture is refused, so it has no counted runtime row.

timeout 120 go test ./internal/lower ./internal/native ./internal/fresh ./internal/flow \
  -run '^(TestLibrarySpread.*|TestOptionalWideningSpreadOverwrite|TestClassFeaturesThrowingAccessorSpread|TestRuntimeFieldLayoutsAreIncluded|TestRegexProgramsKeepCheckedFieldReads|TestRegexOperationsDoNotPoisonTreeWrites|TestRegexOperandsStillJudgeCycleWrites|TestLower.*)$' \
  -timeout 90s -count=1 -v
# PASS for lower, native, fresh; no matching flow tests in this command.
# Native's longest leaf was 14.94s.

timeout 90 go test ./internal/flow \
  -run '^TestFlowProgram_.*(spreads_a|spread_undefined_a|plain_data_spread_a|class_as_interface_a)_' \
  -timeout 60s -count=1 -v
# PASS: three matching flow leaves, 0.20s, 0.21s, 0.26s; package 0.270s.

timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s -count=1
# PASS; final package 28.426s.
timeout 120 go vet ./internal/lower ./internal/oracle
# exit 0, no diagnostics.
gofmt -l internal/lower/library_spread.go internal/lower/library_spread_test.go \
  internal/lower/lower.go internal/lower/object.go internal/oracle/library_spread_test.go
git diff --check
# Both exit 0, no output.
```

New lowering leaf seconds in the final focused run: RegExpUnion 0.22; ErrorView 0.09; PlainProducers 0.20; MapView 0.08; SetView 0.18; DateView 0.22; TypedArrayView 0.19. The oracle leaf's first run, including its cold harness setup, was 16.23s. All are below 60s. Early bounded Go commands expired during cold compilation, before any test leaf started; compilation was then completed separately. The initial typed-array test failed with the existing generic representation rejection, prompting the named diagnostic refinement; its final run passes.

Mutants were built using Go overlays, without modifying the working tree. Their source evidence ends in `.go.txt`:

- `drop-refusal`: remove the call to `checkLibrarySpreads`. The final-source oracle mutant fails `TestRegExpViewSpreadIsRefused` in 4.97s with `want library spread refusal, got <nil>`. It restores admission of the task program. This is an assertion kill, not a compiler warning or a different refusal.
- `drop-plain-proof`: stop accepting the complete plain-producer proof. `TestLibrarySpreadPlainProducers` fails in 0.18s because the unrelated RegExp now wrongly refuses the record/class spread. This independently holds the promised admission control.

`prepare-mutants.py` recreates the overlays from the current source. `admission.go.txt` and `build-admission.py` build the census tool without adding a product package. With setup's environment sourced, run `timeout 120 python3 review/compiler/regexp-view-spread/build-admission.py --baseline --output /tmp/regexp-main-admission` and the same command without `--baseline` for the final compiler; each executable writes JSONL for the complete corpus under `timeout 300`. The baseline is pinned to the recorded main SHA, not a moving branch.

Reader-package compilation is a filtered `go test -run '^$'` check, not a claim that reader behavior suites ran. The 46 direct reader packages are listed in `reader-packages.txt`. The first `timeout 180 go test -p 2 -run '^$' -timeout 90s $(cat .../reader-packages.txt)` completed 34 packages before its command limit; the remaining 12 passed separately with `-p 4 -vet=off` and the same test filter. All 46 packages completed successfully across those two logs. No test leaf ran or was retried in those compilation-only checks. Focused behavioral coverage is exactly the commands above; the full gate is left to Loom.

No enumerability metadata, native spread implementation, library package, other packages' fixture expectations, cohere sources, or runtime feature sets were changed. No cohere code was copied. No PR was opened and nothing was submitted to Loom.

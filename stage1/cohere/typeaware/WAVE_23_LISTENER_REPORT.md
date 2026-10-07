Built: numeric parser SyntaxKind listener declarations for all eighteen claimed rules, plus an independent Go registration oracle and native probe.
Commits: previous landing-ready pushed tip e0408d59, already rebased onto fetched origin/main e8ba3d5d; this metadata change is committed separately.
Commands and outputs: TestWave23NumericListeners PASS in 6.343s; eighteen declarations match 614 bytes across Go, native, sanitized native and emitted JavaScript on Node; vet and formatting checks pass.
Mutants: eighteen independent first-listener-kind corruptions compile, exit 0 with empty stderr and are caught only by Go byte comparison; one per declared rule.
Not covered: numeric dispatch integration, refetch elimination, rule decision changes or speedup; three React ports remain unfinished pending shared JSX frontend integration.

The speed instruction requests declarations before the shared kind-indexed driver
lands. [wave_23_listener_kinds.a](wave_23_listener_kinds.a) provides a readonly
name/kinds table for the fifteen implemented rules and the three reserved React
rules. Entries use actual numeric ast.Kind values from the pinned typescript-go
parser. This file reads no node and contains no string syntax-kind comparison.
All new Adamic sources are .a. No shared registration generator or harness changed.

[oracle_wave_23_listeners.go](testdata/oracle_wave_23_listeners.go) initializes
an independent cohere program and calls each unmodified production rule's Run
with its actual rule context. It sorts the returned listener keys and serializes
their numeric values. It imports no bridge code. Default settings are used;
option-dependent listener subsets are not separately certified by this new test.
Existing decision tests already cover their documented options.

The Go registrations include whole-SourceFile (307) callbacks for rules that
perform internal analyses, including all three React rules. These are preserved
as registered, rather than guessing expression kinds from their names. The full
production inventory observed here is:

```
@typescript-eslint/only-throw-error	258
@typescript-eslint/prefer-promise-reject-errors	214,215
@typescript-eslint/prefer-reduce-type-parameter	214
nexus/correctness-no-collection-misuse	213,214,227
nexus/correctness-no-discarded-outcome	245
nexus/correctness-no-discarded-pure-result	245
no-eval	79,212,213,214
no-extend-native	214,227
no-func-assign	307
no-new-func	214,215
no-new-native-nonconstructor	215
no-new-wrappers	215
no-throw-literal	258
no-useless-backreference	13,307
prefer-arrow-callback	307
react-hooks/set-state-in-effect	307
react-hooks/set-state-in-render	307
react-hooks/static-components	307

```

[wave_23_listener_test.go](wave_23_listener_test.go) compares the native declaration
probe with the Go registration stream. Native is also built with ASan/UBSan and
leak checking; both native runs exit 0 with empty stderr. The emitted JavaScript
runs under Node using the repository's unchanged oracle/adamic.mjs runtime and
also exits 0 with empty stderr. No output bytes are normalized.

For each rule, a separate copy changes its first declared numeric kind to 0.
Every mutant compiles and exits 0, with empty stderr, while its output differs
from Go. These are listener metadata mutants, not React decision mutants or a
replacement for the previously pushed fifteen rule-decision mutants:

```
=== RUN   TestWave23NumericListeners
    wave_23_listener_test.go:113: @typescript-eslint/only-throw-error listener mutant: exit 0, empty stderr, Go byte oracle catches byte 36
    wave_23_listener_test.go:113: @typescript-eslint/prefer-promise-reject-errors listener mutant: exit 0, empty stderr, Go byte oracle catches byte 88
    wave_23_listener_test.go:113: @typescript-eslint/prefer-reduce-type-parameter listener mutant: exit 0, empty stderr, Go byte oracle catches byte 144
    wave_23_listener_test.go:113: nexus/correctness-no-collection-misuse listener mutant: exit 0, empty stderr, Go byte oracle catches byte 187
    wave_23_listener_test.go:113: nexus/correctness-no-discarded-outcome listener mutant: exit 0, empty stderr, Go byte oracle catches byte 238
    wave_23_listener_test.go:113: nexus/correctness-no-discarded-pure-result listener mutant: exit 0, empty stderr, Go byte oracle catches byte 285
    wave_23_listener_test.go:113: no-eval listener mutant: exit 0, empty stderr, Go byte oracle catches byte 297
    wave_23_listener_test.go:113: no-extend-native listener mutant: exit 0, empty stderr, Go byte oracle catches byte 329
    wave_23_listener_test.go:113: no-func-assign listener mutant: exit 0, empty stderr, Go byte oracle catches byte 352
    wave_23_listener_test.go:113: no-new-func listener mutant: exit 0, empty stderr, Go byte oracle catches byte 368
    wave_23_listener_test.go:113: no-new-native-nonconstructor listener mutant: exit 0, empty stderr, Go byte oracle catches byte 405
    wave_23_listener_test.go:113: no-new-wrappers listener mutant: exit 0, empty stderr, Go byte oracle catches byte 425
    wave_23_listener_test.go:113: no-throw-literal listener mutant: exit 0, empty stderr, Go byte oracle catches byte 446
    wave_23_listener_test.go:113: no-useless-backreference listener mutant: exit 0, empty stderr, Go byte oracle catches byte 475
    wave_23_listener_test.go:113: prefer-arrow-callback listener mutant: exit 0, empty stderr, Go byte oracle catches byte 504
    wave_23_listener_test.go:113: react-hooks/set-state-in-effect listener mutant: exit 0, empty stderr, Go byte oracle catches byte 540
    wave_23_listener_test.go:113: react-hooks/set-state-in-render listener mutant: exit 0, empty stderr, Go byte oracle catches byte 576
    wave_23_listener_test.go:113: react-hooks/static-components listener mutant: exit 0, empty stderr, Go byte oracle catches byte 610
    wave_23_listener_test.go:115: eighteen production registrations: 614 identical Go/native/sanitized/Node bytes
--- PASS: TestWave23NumericListeners (6.33s)
PASS
ok  	github.com/system-inc/adamic/stage1/cohere/typeaware	6.343s

```

The initial Node attempt lacked the adamic runtime package. It was initially
misattributed to an .a dependency import; inspecting emitted JavaScript showed
that source dependency was already bundled. The final test installs the existing
runtime in its scratch node_modules/adamic package, without changing generated
JavaScript or shared test code. The initial failure is retained as earlier
evidence and is not counted as a mutant kill or final pass.

Exact commands, after source /workspace/adamic-tools/env.sh:

```sh
ADAMIC_WAVE23_LISTENER_ARTIFACTS=/workspace/wave-23/listeners/gate-final go test ./stage1/cohere/typeaware -run '^TestWave23NumericListeners$' -count=1 -timeout=10m -v > /workspace/wave-23/listeners/gate-final.log 2>&1
go vet ./... > /workspace/wave-23/listeners/vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_23_listener_test.go stage1/cohere/typeaware/testdata/oracle_wave_23_listeners.go > /workspace/wave-23/listeners/gofmt.log 2>&1
git diff --check > /workspace/wave-23/listeners/diff.log 2>&1
```

Current main and the worker's prior remote tip were freshly fetched and verified
unchanged. The previous fifteen rule gates, sanitizer checks, released-handle
checks, corpora and performance observations remain in WAVE_23_LANDING_CURRENT_REPORT.md;
they were not rerun for unused metadata that does not alter rule decisions.
This new test compiles current stage0, uses an independent Go rule registry,
and adds the meaningful oracle needed by this change. Setup remains the earlier
successful run: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 83s,
done 83s; nproc 5. No setup rerun is claimed.

The current shared ParseNode exposes only kind:string, and no numeric node kind
or shared numeric driver API. Legacy rule loops still read string kinds and
refetch nodes; this pass does not claim to satisfy those migration requirements.
Changing parser nodes or the driver is outside this worker's shared-file scope.
These declarations are ready for that driver; this pass claims no lint speedup.
The JSX parser implementation on codex/stage1-jsx-lint remains unmerged on main
and this branch. Static-components, set-state-in-effect and set-state-in-render
have metadata but no completed native decision port. No new claims were made.
No main or area branch is written and no PR is opened.

[validation-wave23-listeners](validation-wave23-listeners) preserves compressed
raw outputs, complete mutant evidence, the earlier failed run and source hashes.

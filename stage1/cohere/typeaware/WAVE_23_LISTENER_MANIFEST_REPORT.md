Built: eighteen worker-owned rule.json manifests declaring numeric parser SyntaxKind listeners; no new rules claimed.
Commits: previous landing-ready pushed tip 9f0e7fe8, on unchanged fetched origin/main e8ba3d5d; this manifest change is committed separately.
Commands and outputs: TestWave23NumericListeners PASS in 9.779s; eighteen JSON manifests match Go and 614 declaration bytes match Go/native/sanitized/Node; vet and whitespace checks exit 0.
Mutants: eighteen numeric declaration mutants compile and run cleanly before byte comparison catches them; a separate JSON kind-to-zero mutation fails the intended Go manifest comparison, exit 1.
Not covered: numeric callback dispatch integration, string-kind/refetch migration, new Diagnostic adoption, speedup or native completion of three React ports.

[ listeners-wave23 ](listeners-wave23) contains one name/kinds rule.json per
claimed rule, preserving namespaced rule names in directory paths. This is a
worker-owned declaration directory, not a new runtime rule registry. Its fifteen
implemented and three blocked entries correspond to wave_23_listener_kinds.a.
The three React manifests describe production listener registrations only;
they do not mark those rules ported or create new claims.

The updated worker-owned wave_23_listener_test.go reads every JSON manifest and
compares its name and ordered numeric kinds against the independent production
Go registry. That oracle calls actual unmodified rule Run methods with a real
checker context, and sorts numeric ast.Kind keys. Default settings are used.
The existing native, ASan/UBSan/LeakSanitizer, emitted JavaScript and eighteen
compiled declaration mutants are rerun. Their complete final output is below:

```
=== RUN   TestWave23NumericListeners
    wave_23_listener_test.go:65: eighteen rule.json name/kinds declarations match the independent Go registry
    wave_23_listener_test.go:140: @typescript-eslint/only-throw-error listener mutant: exit 0, empty stderr, Go byte oracle catches byte 36
    wave_23_listener_test.go:140: @typescript-eslint/prefer-promise-reject-errors listener mutant: exit 0, empty stderr, Go byte oracle catches byte 88
    wave_23_listener_test.go:140: @typescript-eslint/prefer-reduce-type-parameter listener mutant: exit 0, empty stderr, Go byte oracle catches byte 144
    wave_23_listener_test.go:140: nexus/correctness-no-collection-misuse listener mutant: exit 0, empty stderr, Go byte oracle catches byte 187
    wave_23_listener_test.go:140: nexus/correctness-no-discarded-outcome listener mutant: exit 0, empty stderr, Go byte oracle catches byte 238
    wave_23_listener_test.go:140: nexus/correctness-no-discarded-pure-result listener mutant: exit 0, empty stderr, Go byte oracle catches byte 285
    wave_23_listener_test.go:140: no-eval listener mutant: exit 0, empty stderr, Go byte oracle catches byte 297
    wave_23_listener_test.go:140: no-extend-native listener mutant: exit 0, empty stderr, Go byte oracle catches byte 329
    wave_23_listener_test.go:140: no-func-assign listener mutant: exit 0, empty stderr, Go byte oracle catches byte 352
    wave_23_listener_test.go:140: no-new-func listener mutant: exit 0, empty stderr, Go byte oracle catches byte 368
    wave_23_listener_test.go:140: no-new-native-nonconstructor listener mutant: exit 0, empty stderr, Go byte oracle catches byte 405
    wave_23_listener_test.go:140: no-new-wrappers listener mutant: exit 0, empty stderr, Go byte oracle catches byte 425
    wave_23_listener_test.go:140: no-throw-literal listener mutant: exit 0, empty stderr, Go byte oracle catches byte 446
    wave_23_listener_test.go:140: no-useless-backreference listener mutant: exit 0, empty stderr, Go byte oracle catches byte 475
    wave_23_listener_test.go:140: prefer-arrow-callback listener mutant: exit 0, empty stderr, Go byte oracle catches byte 504
    wave_23_listener_test.go:140: react-hooks/set-state-in-effect listener mutant: exit 0, empty stderr, Go byte oracle catches byte 540
    wave_23_listener_test.go:140: react-hooks/set-state-in-render listener mutant: exit 0, empty stderr, Go byte oracle catches byte 576
    wave_23_listener_test.go:140: react-hooks/static-components listener mutant: exit 0, empty stderr, Go byte oracle catches byte 610
    wave_23_listener_test.go:142: eighteen production registrations: 614 identical Go/native/sanitized/Node bytes
--- PASS: TestWave23NumericListeners (9.77s)
PASS
ok  	github.com/system-inc/adamic/stage1/cohere/typeaware	9.779s

```

The new manifest check is separately proven able to fail: changing only
@typescript-eslint/only-throw-error/rule.json from kinds [258] to [0] causes:

```
=== RUN   TestWave23NumericListeners
    wave_23_listener_test.go:62: rule.json kinds differ from Go: got @typescript-eslint/only-throw-error	0, want @typescript-eslint/only-throw-error	258
--- FAIL: TestWave23NumericListeners (5.23s)
FAIL
FAIL	github.com/system-inc/adamic/stage1/cohere/typeaware	5.236s
FAIL

```

The JSON mutant is accepted JSON and numeric metadata. The compiler builds and
independent Go oracle run succeed; only the new manifest comparison fails.
The original bytes are restored in a finally block and confirmed as [258].
The final commit fingerprints every restored manifest. No production source
or Go rule implementation was mutated.

Exact commands after source /workspace/adamic-tools/env.sh:

```sh
ADAMIC_WAVE23_LISTENER_ARTIFACTS=/workspace/wave-23/listener-manifests/gate go test ./stage1/cohere/typeaware -run '^TestWave23NumericListeners$' -count=1 -timeout=10m -v > /workspace/wave-23/listener-manifests/gate.log 2>&1
# With only-throw-error/rule.json temporarily changed to kinds [0]:
ADAMIC_WAVE23_LISTENER_ARTIFACTS=/workspace/wave-23/listener-manifests/json-mutant go test ./stage1/cohere/typeaware -run '^TestWave23NumericListeners$' -count=1 -timeout=10m -v > /workspace/wave-23/listener-manifests/json-mutant.log 2>&1
# Restore original manifest bytes. The mutant command intentionally exits 1.
go vet ./... > /workspace/wave-23/listener-manifests/vet.log 2>&1
git diff --check > /workspace/wave-23/listener-manifests/diff.log 2>&1
```

The all-heads fetch confirms main e8ba3d5d81de4d3773c723914fccd4c76248b965
is unchanged and is an ancestor of this branch. Its previous remote tip
9f0e7fe8d53831d3d6e4e04b1bb9c91b18a7345c was verified. This metadata-only change
alters no rule decisions or driver dispatch, so the preceding fifteen-rule
landing gates remain applicable; their corpus, sanitizer, released-handle and
performance evidence is in WAVE_23_LANDING_CURRENT_REPORT.md. They are not
represented as rerun here. The new declaration oracle runs against current main.

Setup remains the prior successful run: Go 0s, clang 0s, Node 0s, submodules 0s,
cache warm 83s, done 83s; nproc 5. No setup rerun is claimed.

The shared parser still exposes string kinds, and the numeric driver and JSX
frontend integration remain outside this unit's shared-file territory. The
three reserved React ports remain unfinished. No callbacks or new rules were
implemented here, so no new every-rule/every-node dispatch or string relevance
checks were introduced. Legacy rule migration is still pending the shared API.
The user has not supplied the new batch-8 Diagnostic landing sha; no adoption
or rebase onto that unnamed change is claimed. No new claims or PR were made.
Only codex/typeaware-wave-23 is pushed; no main or area branch is written.

[validation-wave23-listener-manifests](validation-wave23-listener-manifests)
contains raw compressed outputs, complete mutant logs and restored source hashes.

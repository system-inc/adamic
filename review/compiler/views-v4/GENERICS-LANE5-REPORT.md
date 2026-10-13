Built checker-resolved generic call contracts, concrete monomorphized producer domains, and lane 5 factory/boxed call-time rewrites toward V4.
Generic commit: 06fe16ac4, pushed separately before the lane 5 point. The second commit contains this report and the lane 5 changes.
Final V4 tests passed in 41.960 s; focused lower/native/JavaScript passed in 8.938/14.270/2.611 s; call-target guard passed in 12.779 s.
Producer/view instantiation omission mutants, the trust-erased compiler mutant, and the skip-adapter mutant were caught in both backends and sanitized native; finishing mutants match Node and are leak-clean.
Pending: rank165's ruled wider-write case, its readonly higher-order control, runtime-callback source-site blame, and existing unsupported first-class named generic producer dispatch. No aggregate sweep or construction implementation was attempted.

Generic relations

A reached generic view call uses GetResolvedSignature only when its declaration belongs to the view's declared signature. Parameters/results are runtime domains of that call's actual instantiation. Producer metadata is recorded from l.concrete under the existing monomorphizer, preserving literal domains rather than binder names. The producer fixture instantiates a factory at 'ok', returns its callable, then calls it with 'bad'; exit70 precedes producer execution. The generic view fixture invokes run<'ok'> and checks the producer's returned string against 'ok'. Dropping either literal check produces Node's unchecked output in native, JavaScript and sanitized native with no leaks.

An escaping quantified relation without a retained runtime witness remains refused in .ts with path and fix. The trust-erased compiler mutant marks that unsupported read as non-escaping. It cleanly executes like Node in native, JavaScript and ASan/UBSan/LSan, then fails the fixture's required refusal assertion with nil. The exact mutant patch and log are committed, and compiler source was restored before final checks. First-class named generic producer values remain at the pre-existing unsupported boundary; generic template dispatch is not claimed.

Lane 5 rewrite

The thirteen original source witnesses under lane5/ are byte-identical to their files in 5d30aca9 (factory) and cb6adb25 (boxing). Tests create temporary .ts controls from them, preserving .a's unproven escaping-read refusal. This branch previously had neither the lane 5 source tree nor its factory tests. No worker branch was merged, and no code was copied from cohere.

Compatible calls in old read-time negatives now succeed like Node. Separate misfit programs require exit70 at their first call: identifier/string-literal/unique-name/import/yield overload arguments; identifier/clone/modifier results; boxed parameter and parameter-members arguments; boxed result and result-members results. Overloaded calls use the selected checker signature only after validating declaration membership. Callable storage in readonly unknown object-literal fields is admitted only for real runtime-checkable producer signatures; this grants no signature certificate. The .a unproven-read refusal remains pinned.

The adapter bypass mutant uses a scalar literal-domain fixture with matching producer/view ABIs. Native bypasses the adapter by normalizing the invoked closure; JavaScript bypasses the adapter in adamicCall. Both finish with Node's 'bad' output instead of required exit70, with native ASan/UBSan/LSan and allocation counts clean. Existing V4 argument/result omission and adapter identity/no-stacking mutants also ran in the final V4 test set.

Rank165 remains a pinned refusal at wider-write.ts:6:48: the readonly onEmitNode field becomes writable. Its fix is exactly keep onEmitNode readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable). Its Node-output test has exactly t.Skip("awaits compiler/checked-wider-writes: checked write through a wider view"). The separate readonly-target control has t.Skip("awaits compiler/views-v4: runtime-checkable higher-order callable parameter contracts"); it needs real nested callable domains and producer/callback boundary checks, not a fabricated matching signature.

Runtime-callback source-site blame follow-up needs source positions on callback-bearing IR (including array sort/visits and RegExp replacement), emission of that invocation position in both backends, and balanced context restoration across nested calls and exceptions. Runtime callback entries must receive the library call's source position rather than inheriting an unrelated caller or substituting the adapter's read position. Existing two-site read/call blame for ordinary, call/apply and bound invocations remains green.

Commands and evidence

Every Go test used -count=1 -timeout 90s and outer timeout 90, writing directly to a review log.

- go test ./internal/oracle -run TestV4Generic -v: 1.095 s (generic leaves 0.29/1.01/1.08 s in the focused run).
- go test ./internal/oracle -run TestV4 -v after generic work: 13.952 s.
- go test ./internal/lower ./internal/native ./internal/javascript -run 'View|Generic': 5.055/9.552/1.875 s.
- go test ./internal/ir -run TestCallTargetReaders after generic work: 13.434 s.
- Trust-erased mutant: failing refusal assertion in 0.553 s after clean semantic admission; compiler restored.
- go test ./internal/oracle -run TestV4Lane5 -v: 3.897 s before the final exact refusal pin and pending-control declaration.
- go test ./internal/oracle -run TestV4 -v final: 41.960 s, all active leaves passed, three explicit pending tests.
- go test ./internal/lower ./internal/native ./internal/javascript -run 'View|Generic|Unknown': 8.938/14.270/2.611 s.
- go test ./internal/ir -run TestCallTargetReaders final: 12.779 s.
- Integration lane command follows each commit; generic lane: 3.2 s, gofmt/tools 136 Go files, t.Parallel 8 test packages, vet 8 packages.

Final new lane 5 leaf seconds: identifier 3.41, string-literal 3.15, unique-name 3.01, import 2.92, yield 2.63, identifier-result 1.44, clone 2.82, modifier 3.10, boxed-parameter 2.43, parameter-members 2.94, boxed-result 1.29, result-members 1.13, skip-adapter 2.20, .a refusal 0.33, wider-write refusal 0.28. Generic leaves in the final combined run: erased refusal 0.47, view instantiation 3.80, producer instantiation 4.92. All remain below 60 s.

Setup: Go 0.020 s, Node 0.021 s, markdown ready 0.082 s, submodules 0.086 s, clang 0.198 s, Go build 43.297 s, total 43.525 s. nproc=5, CPU quota=4. No registered oracle fixture count rows changed; evidence sources are under review and new tests use temporary inputs. The count table is not expanded or claimed as refreshed across unrelated fixtures.

# V4 function surface: call and apply

Escaped .ts adapters invoked with call or apply keep their producer argument and view result checks. Evaluation order is callable, thisArg, arguments, argument/receiver checks, producer, result check. An explicit, runtime-checkable object this parameter receives thisArg. .a escaping reads still require proof, including the receiver relation. Named dynamic receiver functions and inferred or uncheckable receivers remain refused.

Apply currently accepts dense argument literals. Spread/hole literals, dynamic argument arrays, optional surface invocations, and primitive thisArg representations remain explicitly unsupported. This conservatively retains the compiler's existing ABI frontier. Bind and new are separate operations.

Native adapters receive the receiver through the existing closure receiver convention. Ordinary invocations pass an absent receiver. Sort and RegExp callback invocation paths preserve that convention. JavaScript forwards the receiver as an adapter invocation argument and preserves the wrapper's own reflection length so existing surface mutants still discriminate.

Test commands (each -count=1 -timeout 90s, with an external timeout 90):

- go test ./internal/oracle -run '^TestV4EscapeAdapter(Call|Apply|Surface(Sort|RegExp)Callback)' -v: 2.019 seconds. Twelve new leaves, each 0.33 to 1.16 seconds in this run. Node truth is asserted before compiler outcomes.
- go test ./internal/oracle -run '^TestV4' -v: 8.337 seconds before the final receiver-proof addition; final focused rerun recorded separately.
- go test ./internal/lower -run 'View|Call|FunctionExpression|Arguments|LibraryLanguage': 2.366 seconds.
- go test ./internal/native -run 'View|Closure|Arguments': 5.776 seconds, green. This is the requested focused rerun of the broader native suite that previously reached 90 seconds; no claim of full native-package coverage.
- go test ./internal/javascript -run 'View|Closure|Arguments': 0.796 seconds.
- go test ./internal/ir -run TestCallTargetReaders: 20.886 seconds.

Four surface invocation omission mutants (call argument/result, apply argument/result) bypass the adapter and must finish exactly like unchecked source Node, with native ASan/UBSan/LeakSanitizer clean. The receiver-domain omission mutant removes its literal constraint and likewise reaches Node's successful body/output. Every mutant is caught by the required exit-70 behavior and output order. Sort and RegExp controls finish sanitizer/count clean.

No persistent fixture was added. Oracle count rows remain unchanged; new oracle programs are temporary language-policy inputs. Setup: node 0.024s, Go 0.031s, markdown dependencies 0.082s, submodules 0.098s, clang 0.195s, build 40.728s, total 40.918s. nproc=5, cgroup quota=4 CPUs.

Final package validation: full lower 27.268s; full JavaScript 0.510s; native focused View|Closure|Arguments|RegExp 24.000s, all green. See call-apply-packages-final.log and call-apply-native-final.log.
Final TestCallTargetReaders: 15.002s, green.

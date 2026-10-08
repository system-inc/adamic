Built: callback and generic NonNullable assertion admission, plus ordinary call direction reporting in --explain-checks.
Commits: callback d8f6241210b229c49e327310e28a986185989a8d pushed; assertion implementation and final evidence are in the containing commit.
Commands/results: complete lower, CLI, IR and load packages pass; uncached filtered predicate/generic/regex oracle passes; vet, formatting and diff checks pass.
Mutants: six independent compiler mutants caught; null-test and scalar-evaluation erasures fail both backends; undefined/null confusion fails native output.
Limits: mixed-union closure ABI is under investigation; the full repository gate is not covered.

The requested base already admits the exact unused callback signature probe. It is
pinned byte for byte from 5d777de3b9506d2a303a4a69d66d21f37f608ede,
`stage3/drivers/parser/native-predicate-callback-parameter.a`, as
`internal/lower/testdata/predicates/parser_callback_parameter.a`.
The executable witnesses exercise narrowing inside the callee with named and
inferred arrow argument functions, including a generic `TOut extends TIn` callback
contract. Existing production admission independently proves each argument body.
No annotation or inferred checker claim substitutes for that proof.

The first attempted execution witnesses used number/string mixed unions. Both
stopped at the existing closure ABI refusal for differently held members. The
conservative scope assumption is to test the predicate lane using the supported
string/undefined representation and retain that independent capability gap.
The original probe remains unchanged, including its unused generic `any` default.

The callback mutant inserts unconditional `return true` into
`proveInferredPredicate`. TestParserCallbackLie then fails with
`want callback body proof refusal, got <nil>`, not a compilation or warning error.
The runner restores the production file in a finally block. The original source
lie passes checker inference via an overload and fails on Node with
`adamic: panic: TypeError: Cannot read properties of undefined (reading 'toUpperCase')`,
exit 70. The independent compiler argument-body proof refuses it before emission.

`--explain-checks` for the original callback probe and the executed generic callback
both prints exactly:

```text
adamic: predicate checks: proven 0 checked 0 unobservable 0
```

These are overload-reporting counts, not a claim that the callback's proof is absent.
The original declaration-only witness emits no callback invocation. Per-direction
reporting for ordinary predicate and callback calls is not implemented by this unit.

Setup used `GOPROXY=https://proxy.golang.org|direct` and
`source /workspace/adamic-tools/env.sh`; no setup failure occurred. An early probe
attempt before the checker checkout finished failed with missing
`cohere/TypeScript/tsc/go.mod`; it was rerun once the dependency was available.
Timing lines: Node 0.085s, Go 0.122s, clang 0.727s, markdown dependencies 1.161s,
submodules 162.088s, Go build 480.451s, deferred test binaries 480.587s,
build cache warm 480.589s, done 480.625s. nproc 5, cgroup quota 4 CPUs.
Go 1.27.1, Node 24.19.0, clang 20.1.8. Cold test compilation overlapped setup.

Commands wrote output directly to logs:

```sh
go test ./internal/lower -run 'TestParserCallback|TestPredicateOverloadRuntime/(parser_callback_parameter|callback_parameter_read|generic_callback_parameter_read)$' -count=1 -timeout 10m
python3 stage3/census/predicates/parser-stops/mutants.py callback
go test ./internal/lower -count=1 -timeout 10m
go vet ./internal/lower
```

Evidence is in `evidence/`. The full repository gate was not run.

The base is `a58ba402b428280d4932afe6c98c1f47a5d3c495`. Merge
`e02efc1fd6782de2c387c211c82e04496ecc6706` has that base and current main
`f4efdd2369311d1420aa53fdf5c1a55bdda811d4` as its two parents. Main was
explicitly fetched again after the mutants; it remains an ancestor of this branch.
Only `codex/proven-predicates-2` is pushed. No pull request was opened.

The exact assertion probe is pinned from the same probe commit as the callback,
`stage3/drivers/parser/native-generic-assert-non-nullable.a`, as
`internal/lower/testdata/predicates/parser_assert_defined.a`. Before this step,
it reproduces the normal-return refusal recorded in assertion-baseline.log.
The flow verifier now removes an expression-statement call to a direct failure
helper only when the existing verifier independently inspects its implementation
and establishes that it never returns. A never annotation alone supplies no fact;
recursive and unsupported helper bodies remain conservative. The normal-return
flow then independently narrows the unchanged argument to `NonNullable<T>`.
No runtime predicate check is necessary for this proved body.

Execution exposed two companion lowering gaps. A scalar/null comparison used to
stop as NotYet, even though its scalar representation cannot contain null. It now
uses the existing IsNull node with AlwaysFalse and preserves operand evaluation.
Both null and undefined comparison selection must inspect the concrete generic
instantiation. The former must retain a real null test for nullable regex results;
the latter must distinguish a native null pointer from undefined. The previous
undefined selection hid a missing null test in native execution by matching null
as undefined. This is now held by an observable generic `isUndefined` call.
General scalar null storage and mixed nullable/undefined unions are not added.

The executable assertion witness prints exactly this on source Node, generated
JavaScript and ASan/UBSan native (exit 0, empty stderr):

```text
WORD
missing
word
no match
1
no number
evaluated
false
null evaluated
false
```

It covers successful normal returns, actual undefined and null failures caught by
the caller, string and numeric reads narrowed after the call, and an impossible
scalar/null comparison whose function argument must still run once. Native uses
ASAN_OPTIONS=detect_leaks=1. These witnesses use the existing lower-package
runtime harness; they do not add rows to the registered oracle allocation corpus.

Seven source erasures are separately required to fail with the original assertion
normal-return proof diagnostic: empty body, missing undefined exclusion, missing
null exclusion, returning failure helper, annotation-only failure helper,
recursive failure helper, and early normal return. Erasing the executable
assertion body on Node changes the caught errors and drops the numeric failure message,
as recorded in assertion-erasure-node.log. This observation is separate from the
compiler's compile-time rejection of those claims.

The final compiler mutants all restore their files and fail a semantic assertion,
never just a build or clang warning:

| Mutant | Catcher and observation |
|---|---|
| Accept inferred callback without proving its body | TestParserCallbackLie: admission unexpectedly returns nil |
| Accept assertion normal returns without proof | TestParserAssertionProofErasure: erased or incomplete assertion admitted |
| Treat every direct failure helper as non-returning | TestParserAssertionProofErasure: returning helper admitted |
| Erase specialized null test | TestParserAssertionBackends/native and /javascript: panic at the null array read instead of caught `no match` |
| Erase specialized undefined/null distinction | TestParserAssertionBackends/native: prints `true` instead of Node's `false`; JavaScript keeps its strict comparison |
| Replace impossible scalar/null comparison with a constant and skip operand evaluation | TestParserAssertionBackends/native and /javascript: missing `evaluated` line |

The initial failure-helper mutant left a Go local unused. The runner rejected
that build-only outcome; it was corrected to preserve the syntactic use while
bypassing the proof, and its semantic failure was rerun. The initial specialized
null mutant failed JavaScript but was masked in native by the companion undefined
bug. The runner rejected the incomplete two-backend kill, leading to the second
comparison fix and its independent mutant. Those rejected outcomes are preserved
in rejected-helper-build-only.log and masked-null-check-before-undefined-fix.log;
neither is counted among the final six caught mutants.

Final `--explain-checks` output for each exact original probe, in C, JavaScript and
sanitized build modes, is exactly:

```text
adamic: predicate checks: proven 0 checked 0 unobservable 0
```

Both exact probes only load declarations; neither invokes its predicate. The
ordinary-predicate reporter limitation stated above still applies. C and JavaScript
output are nonempty, and native builds pass. Build mode was rerun with the CLI's
required `--explain-checks -o OUTPUT --sanitize` ordering after an initial usage
error. The executable witnesses are independently held to Node by runtime tests.

Final unmutated commands and complete package results:

```text
go test ./internal/lower ./cmd/adamic ./internal/ir ./internal/load -count=1 -timeout 10m
ok github.com/system-inc/adamic/internal/lower 72.238s
ok github.com/system-inc/adamic/cmd/adamic 17.103s
ok github.com/system-inc/adamic/internal/ir 39.073s
ok github.com/system-inc/adamic/internal/load 2.885s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestPredicateDirectionCountsAreRecorded|TestNativeAgreesWithNode/internal/oracle/testdata/(proven_|census_|regexp|maybe_number_slots|weak_narrowed|visits|generic)' -count=1 -timeout 30m
ok github.com/system-inc/adamic/internal/oracle 46.342s
go vet ./...
exit 0, no output
gofmt -l cmd internal
exit 0, no output
git diff --check
exit 0, no output
python3 stage3/census/predicates/parser-stops/mutants.py
six caught, exit 0
```

The full repository gate and full native/oracle packages were not run. Unsupported
callback argument shapes, forwarding through opaque function values, method
assertion helpers and ordinary-predicate per-direction reporting remain outside
this unit. No code was copied from cohere. The implementation does not edit any
of the four prohibited compiler/oracle ownership files. The report and evidence
keep original probe admission, actual execution, and reporting limits separate.

## Ordinary call reporting follow-up

Ordinary annotated predicate calls now record the same direction states as overload
calls. Calls used as statements are included, so a normally returning assertion is
reported. Callback parameter invocations report their validated closed-world argument
contract. Ordinary tag predicates report checked directions where narrowed reads use
checked views, and unobservable directions where no such read occurs. Reporting adds
no new runtime checks or admission rules. Overload numbering and output are unchanged.

Exact CLI stderr goldens cover C, JavaScript and sanitized build for
`callback_parameter_read.a`, `assert_defined_read.a`, and `ordinary_regions.a`.
The callback reports proven 2, checked 0, unobservable 0. The assertion witness has
three specializations/call sites, each with only a proven true direction, totaling
proven 3. The direct predicate witness reports proven 5, checked 1, unobservable 3.
`ordinary_regions.a` also runs on original Node and both backends with stdout
`word`, `true`, `true`, each on its own line, and exit 0 with empty stderr.

Validation: `go test ./internal/lower ./cmd/adamic -count=1 -timeout 10m` passed
(lower 28.721s; CLI 8.126s). `go test ./internal/oracle -run
TestPredicateDirectionCountsAreRecorded -count=1 -timeout 5m` passed (1.221s).
Erasing ordinary expression reporting and separately erasing assertion statement
reporting each fails `TestExplainChecksOutput` by an exact stderr mismatch; both
mutants built successfully and were restored. Logs are in `evidence/ordinary-*`
and `evidence/mutant-ordinary-report.log`, `evidence/mutant-assertion-report.log`.

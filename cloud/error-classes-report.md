Error subclasses retain own fields, JavaScript defaults, cause and nominal hierarchy identity.
Generated TypeError and ReferenceError failures are ordinary catchable throws with cleanup.
Every uncaught throw keeps Adamic's panic prefix and exit 70, including Node oracle normalization.
The branch is rebased onto main 5d4c801; all six original convention regressions agree with Node.
Five error fixtures, 35 compiled mutants and 21 source mutants cover the implementation.

Branch: `codex/error-classes`, rebased onto `origin/main` at `5d4c801`.
Implementation commits: `be5a509`, `bb37125` and `7e40a8b` (rebased earlier commits), followed by `8465ed20663345932c7f40595286ac255b38bd9c` for the convention correction and generated-error integration. No pull request.

The implementation uses ordinary class layouts and the existing cleanup paths. Built-in errors share a name/message/cause prefix; derived fields follow it. Error values can be returned, stored, passed and thrown in another function. Catch uses actual instanceof identity, including negative tests between TypeError, RangeError and user classes. Explicit and inherited default constructors forward message and cause correctly, including through several subclass levels.

Error.prototype.toString follows the standard empty-name, empty-message and undefined-field rules. Its explicit .call form supports errors and plain objects with known required fields. Incompatible primitive receivers throw nominal TypeError objects. Number-format bounds, repeat count/length and normalize form checks throw nominal RangeError objects. Invalidated property narrowings throw TypeError, with the receiver and right side evaluated before a failed write. All these throws use the existing exception cleanup paths. Library methods and inherited constructor properties are refused before an own-field load.

Uncaught supported errors flush earlier stdout, run finally, write `adamic: panic: ` plus the standard name/message line and exit 70. The Node oracle maps every uncaught throw to this same convention, including primitive throws. It formats synthetic error objects with Error.prototype.toString and omits frames and source excerpts. Raw Node exits 1 for uncaught errors; that raw convention is deliberately normalized to Adamic's documented exit 70. Explicit panic remains uncatchable.

GraphQL gap 2 is closed: its minimal returned-error fixture passes Node, native and the leak check. The GraphQL port itself still has its 35 throw-site construction workarounds; this unit did not rewrite those sources.

The six original regressions are fixed: narrowed_fields.a, narrowed_methods.a, narrowed_writes.a, narrowed_reads.a, reuse_narrowed.a and dead_zone.a all pass the three-backend comparison. The exit convention was corrected without moving generated failures to exit 1. The obsolete, unapplied exit-1 emitter proposal was deleted.

The new error_checks.a fixture holds temporal-dead-zone reads and writes through functions, finally and a closure; positive/negative ReferenceError instanceof checks; nominal TypeErrors after invalidated undefined and null narrowings; assignment right-side evaluation; and stopping later operands after a failed read. Main's new intrinsic Number and String prototype .call forms use the same checked wrappers as direct methods. Array.with constructs a nominal RangeError instead of renaming a generic Error. Error subclass descriptor observations stay explicitly NotYet, so nonenumerable fields never masquerade as enumerable own fields.

ReadyErrors registers ordinary ReferenceError constructor calls for checked globals before computing the exception fixed point. The small emitter edit in emit_locals.go constructs that error and follows existing pending-exception cleanup, retaining the surrounding statement's ownership bookkeeping on the successful branch. The JavaScript backend uses the same nominal constructor. The flow graph records ready-check exception edges. MayThrow grows monotonically: an unrelated unchecked read cannot erase a throw already discovered earlier in the body.

Setup was `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8 and Node 24.19.0 were available; setup succeeded without a workaround. `nproc` printed `5`:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (88s)
setup: done in 88s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Verification commands and observations, with test output written to logs:

| Command | Observed result | Log |
| --- | --- | --- |
| `bash cloud/setup.sh` and `nproc` | exit 0; 88s, 5 processors | `/tmp/adamic-errors-setup-rebased.log` |
| `gofmt -l cmd internal` | exit 0, no output | `/tmp/adamic-errors-gofmt-corrected-final.log` |
| `go vet ./...` | exit 0, no output | `/tmp/adamic-errors-vet-corrected-final.log` |
| `git diff --check` | exit 0, no output | terminal |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts` | exit 0, 19.753s; table updated | `/tmp/adamic-errors-counts-corrected-final.log` |
| Uncached error unit, exact command below | exit 0; oracle 10.156s, lower 0.713s; five fixtures, six regressions, 35 compiled mutants, 21 refusal probes and five primitive-throw convention probes | `/tmp/adamic-errors-unit-corrected-final2.log` |
| `go test ./internal/lower -count=1` after the final descriptor guard | exit 0, 7.784s | `/tmp/adamic-errors-lower-last.log` |
| `python3 cloud/error-classes-mutants.py --logs /tmp/adamic-errors-source-mutants-corrected-final` | exit 0; all 21 compiled, then failed their intended assertions | `/tmp/adamic-errors-source-mutants-corrected-final.log` |
| `python3 cloud/error-classes-mutants.py --only derived-error-descriptors --logs /tmp/adamic-errors-descriptor-mutant-final` after adding Object.hasOwn's guard | exit 0; all four descriptor assertions caught restored acceptance | `/tmp/adamic-errors-descriptor-mutant-final.log` |
| Scoped Node flow trace for generated errors | exit 0; 173 points, 296 events | `/tmp/adamic-errors-library-identity.log` |

Exact uncached error-unit command:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle ./internal/lower -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(error_checks|error_classes|narrowed_fields|narrowed_methods|narrowed_writes|narrowed_reads|reuse_narrowed|dead_zone)|TestError(ClassMutants|ReportingMutants|Boundaries)|TestGeneratedErrorMutants|TestOracleUncaughtThrowConvention|TestStockNodeErrorExit' -count=1 -v > /tmp/adamic-errors-unit-corrected-final2.log 2>&1
```

The initial complete gate in the previous pass took roughly fourteen minutes, including a 694.505s Unicode-property sweep. This correction runs the complete affected packages instead of repeating that unchanged sweep: ir, lower, javascript, flow, fresh, native, oracle and stage1/cohere/graphql. The first affected run found two obsolete NotYet expectations, a nonmonotonic exception-discovery bug affecting flatMap/Map.groupBy, and Array.with failures made as generic errors. All were corrected. The final affected gate passed:

```text
go test -count=1 -timeout 30m ./internal/ir ./internal/lower ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql
? ir: no test files
ok lower 38.392s
? javascript: no test files
ok flow 158.348s
ok fresh 67.494s
ok native 247.510s
ok oracle 241.132s
ok graphql 104.718s
exit 0
```

Log: `/tmp/adamic-errors-gate-corrected-final.log`. A final narrow Object.hasOwn guard for error subclasses was added afterward and held by the complete lower package, its refusal/mutant probes and the uncached error unit. It prevents inherited name defaults from becoming apparent own fields.

Cohere was built with `go build -o /tmp/adamic-errors-cohere ./command/cohere` in the pinned submodule. The repository-wide `--no-fix` check failed on existing formatting findings and a nested parser failure in cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.bug-hir_loc_diff_2.js. The pinned binary does not recognize .a files; the early selected .a formatting command checked nothing. The workaround was byte-identical .ts copies of the five fixtures under `/tmp/adamic-errors-cohere-fixtures`, using the same compiler options and original prelude. Formatting was applied back as a patch. The hierarchy fixture has a documented max-classes-per-file exception. The generated-check fixture has a documented no-use-before-define exception because it deliberately reaches the temporal dead zone.

Final scoped command:

```text
/tmp/adamic-errors-cohere --directory /tmp/adamic-errors-cohere-fixtures --no-fix error_checks.ts error_classes.ts error_classes_uncaught.ts error_classes_uncaught_empty.ts error_classes_uncaught_name.ts
exit 0
276 rules, 5 checked, 100% Adamic-ready (5 of 5), only 5 of 6 files
```

The sixth file is the existing prelude, which was loaded for type checking and excluded from the selected-file lint/format scope. Log: `/tmp/adamic-errors-cohere-all-final.log`.

Every compiled mutant built under the native compiler's normal strict flags. Semantic mutants disagreed with Node in stdout or exit status; reporting mutants changed status or dropped earlier stdout. Build errors do not count. The permanent tests are TestErrorClassMutants, TestErrorReportingMutants and TestGeneratedErrorMutants:

| Compiled mutant | What caught it |
| --- | --- |
| Lose MyError's Error ancestry | Node identity/output comparison |
| Lose TypeError's Error ancestry | Node identity/output comparison |
| Give RangeError TypeError's definition identity | Negative instanceof comparison |
| Lose message in super | Node message/output comparison |
| Change default name to MyError | Node name/output comparison |
| Ignore cause | Narrowed cause/output comparison |
| Ignore empty name | Standard toString/output comparison |
| Ignore empty message | Standard toString/output comparison |
| Change separator from colon-space | toString/output comparison |
| Fold instanceof TypeError true | Negative user-error identity comparison |
| Remove toFixed guard | Catch output and exit comparison |
| Remove repeat guard | Catch output and exit comparison |
| Remove normalize-form guard | Catch output and exit comparison |
| Remove precision guard | Catch output and exit comparison |
| Remove radix guard | Catch output and exit comparison |
| Remove undefined-property read guard | Catch identity/output and exit comparison |
| Remove undefined-property write guard | Catch identity/evaluation-order/output and exit comparison |
| Remove exponential guard | Catch output and exit comparison |
| Lose cause through default constructor | Forwarded cause/output comparison |
| Throw RangeError for incompatible prototype receiver | TypeError catch/identity comparison |
| Lose undefined primitive receiver's text default | Exact TypeError message comparison |
| Lose undefined generic name's default | Generic standard-method output comparison |
| Lose undefined generic message's default | Generic standard-method output comparison |
| Change only uncaught status to 1 | Node exit 70 versus native 1, stdout/stderr unchanged |
| Skip buffered stdout flushing | Node before/finally output versus missing native output |
| Restore panic ready checks | A catch cannot observe ReferenceError; exit/output comparison |
| Give ready failures TypeError identity | ReferenceError catch and negative instanceof output |
| Ignore ready reads | Output after an invalid read |
| Ignore ready writes | Missing catch, right-side and later-write output |
| Suppress ready callee MayThrow | Work after a failed call becomes observable |
| Suppress closure throw propagation | Closure catch/output comparison |
| Remove null-property read guard | Nominal TypeError catch and exit comparison |
| Remove intrinsic Number prototype guard | RangeError catch/exit comparison |
| Remove intrinsic String prototype guard | RangeError catch/exit comparison |
| Give Array.with a generic Error | Nominal RangeError catch/identity comparison |

The source mutants are Go overlays, never edits to the working tree. Each compiled, then failed its intended assertion. Refusal mutants made Lower accept the refused probe; the registration mutant made the ownership coverage test find unrecorded writes. Reproduce all 21 with `python3 cloud/error-classes-mutants.py --logs /tmp/adamic-error-source-mutants` after sourcing the tools environment:

| Source mutant | Check that failed |
| --- | --- |
| Allow Error.stack read | stack read refusal |
| Allow Error.stack write | stack write refusal |
| Stop following concrete boxed cause types | cause closes a cycle refusal |
| Allow an unchecked toString override | override ABI refusal |
| Treat optional generic field as present | optional generic field refusal |
| Accept erased unknown cause | erased cause refusal |
| Drop nonliteral options | nonliteral ErrorOptions refusal |
| Mark unbounded normalization fully guarded | normalization expansion refusal |
| Restore structural error views, direct and nested | structural view refusals |
| Restore fake structural Error values | nominal ancestry refusal |
| Enumerate Error by native shape spread | error spread refusal |
| Restore inherited method own-field read | inherited method refusal |
| Accept mutable unknown aliases | unknown representation refusal |
| Give synthesized initializer writes Site 0 | fresh.TestEveryWriteIsRecordedAndKnown |
| Restore mutable boxed cause writes | mutable cause refusal |
| Allow nullable generic object receiver | nullable receiver refusal |
| Restore inherited constructor own-field read | inherited constructor refusal |
| Drop a ready-read exception edge | Node trace ends a call in the middle of its graph block |
| Drop a ready-write exception edge | Node trace ends a call in the middle of its graph block |
| Let an unchecked read overwrite MayThrow | Node comparison sees work after a failed call |
| Treat derived errors as plain descriptor-bearing objects | Four derived-error descriptor refusal assertions, including Object.hasOwn |

The ready-write flow mutant initially survived because its right side could also throw, independently providing an exception edge. The final probe uses a right side proven not to throw, then performs observable work after the write. The ready-read probe likewise performs work after its read. The final flow mutants each fail with "the call ended ... in the middle of bb1", proving the specific missing edge. The initial caller-propagation mutant also needed work after the failed call; that isolation was added before the passing final mutant run.

Early failures were real findings, not counted mutants: the whole gate caught missing synthetic write registrations; a stdout regression caught missing managed-buffer flushing. Intermediate mutant harness bugs (nil spread cloning, byte/string conversion and shared slice change detection) were fixed before the final passing mutant run.

Not covered or deliberately refused:

- Stack reads and writes are NotYet because native frames cannot report a JavaScript source stack. Full raw Node stderr, source locations, frames and extra-property inspection are not reproduced.
- ErrorOptions variables, erased unknown causes, mutable cause links and mutable unknown aliases are refused with ownership/representation reasons. The tested cause values are string, number, boolean and nested Error objects.
- Error spreading and structural views are refused because JavaScript descriptors and property absence are not represented. Error.toString overrides are refused until the synthetic built-in method's override ABI is checked.
- Generic prototype calls with optional/erased fields, nullable object receivers, arrays/functions or object coercion are NotYet. Required string fields containing undefined are supported and tested.
- Error, TypeError and RangeError constructors are directly held by these fixtures. ReferenceError is exercised by generated checks. SyntaxError, URIError and EvalError use the same implementation but were not directly exercised by this unit's fixtures. AggregateError and bare Error(...) calls were not implemented.
- Non-Error throws remain refused. Other unguarded array-length and code-point failures cannot unwind to catch and are refused inside try when arguments are not proven safe. Normalization inside try requires a bounded constant input. Stack exhaustion and other unguarded string-size failures still cannot unwind to catch.
- No performance measurements or GraphQL source rewrite were made. The full repository Cohere gate is not green. The complete uncached repository gate remains for integration.

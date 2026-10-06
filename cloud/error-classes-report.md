Error subclasses now retain own fields, JavaScript name/message defaults and nominal identity.
Literal ES2022 cause options retain boxed values under the existing ownership and cycle proof.
Four uncached Node fixtures pass; 42 mutants are caught (25 compiled, 17 source-check mutants).
The branch is reviewable, but one whole-oracle failure remains in the reserved native emitter.
The tested correction is in cloud/error-classes-emitter.patch and has not been applied.

Branch: `codex/error-classes`, from `origin/main` at `d799ede`.
Implementation commits: `45e7cb5e053adb6eecd9f688c11def3fd283005f` and
`9a50870601ba155723ea088de7dba903aea933bd`. No pull request.

The implementation uses ordinary class layouts and the existing cleanup paths. Built-in errors share a name/message/cause prefix; derived fields follow it. Error values can be returned, stored, passed and thrown in another function. Catch uses actual instanceof identity, including negative tests between TypeError, RangeError and user classes. Explicit and inherited default constructors forward message and cause correctly, including through several subclass levels.

Error.prototype.toString follows the standard empty-name, empty-message and undefined-field rules. Its explicit .call form supports errors and plain objects with known required fields. Incompatible primitive receivers throw nominal TypeError objects. Number-format bounds, repeat count/length and normalize form checks throw nominal RangeError objects. Invalidated property narrowings throw TypeError, with the receiver and right side evaluated before a failed write. All these throws use the existing exception cleanup paths. Library methods and inherited constructor properties are refused before an own-field load.

Uncaught supported errors flush earlier stdout, run finally, write the standard name/message line and exit 1. Explicit panic remains prefixed and exits 70. Node's oracle handler normalizes the error with Error.prototype.toString and omits frames and source excerpts. This is not a reproduction of full raw Node diagnostic stderr. A separate stock Node probe confirmed exit 1 and the ordinary Error message line.

GraphQL gap 2 is closed: its minimal returned-error fixture passes Node, native and the leak check. The GraphQL port itself still has its 35 throw-site construction workarounds; this unit did not rewrite those sources.

The remaining gate failure is `internal/oracle/testdata/dead_zone.a`:

```text
Node:   exit 1, stdout "before\n", stderr "ReferenceError: Cannot access 'later' before initialization\n"
Native: exit 70, stdout "before\n", stderr "adamic: panic: ReferenceError: Cannot access 'later' before initialization\n"
```

The standing unit instruction reserves internal/native/emit.go for another worker. The proposed patch changes only checkReady's generated failure call and exposes the existing ordinary-error helper in the runtime header. It has been tested by changing generated C in an isolated test overlay, without editing the reserved file:

```text
proposal agrees with Node: exit 1, stdout "before\n", stderr "ReferenceError: Cannot access 'later' before initialization\n"
PASS
ok github.com/system-inc/adamic/internal/oracle 0.352s
```

This proposal corrects uncaught reporting; it does not make temporal-dead-zone failures unwind to catch. None of the four reserved files was edited. Do not merge this branch with the known oracle failure still present.

Setup was `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8 and Node 24.19.0 were installed; setup succeeded without a workaround. `nproc` printed `5`:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (72s)
setup: done in 72s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Verification commands and observations, with test output written to logs:

| Command | Observed result | Log |
| --- | --- | --- |
| `gofmt -l cmd internal` | exit 0, no output | `/tmp/adamic-errors-gofmt-final.log` |
| `go vet ./...` | exit 0, no output | `/tmp/adamic-errors-vet-final.log` |
| `git diff --check` | exit 0, no output | terminal |
| `go test -count=1 -timeout 30m ./...` | exit 1; initially missing synthesized write records and six diagnostic/status differences | `/tmp/adamic-errors-gate.log` |
| `go test -count=1 -timeout 30m ./internal/ir ./internal/lower ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql` | lower 15.210s, fresh 54.338s, native 118.192s, GraphQL 93.866s passed; oracle failed on TDZ and an intermediate mutant harness bug | `/tmp/adamic-errors-affected-final.log` |
| `go test ./internal/oracle -count=1 -timeout 30m` after the harness fix | exit 1, 63.715s; only dead_zone.a failed | `/tmp/adamic-errors-oracle-final.log` |
| `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle ./internal/lower -run 'TestNativeAgreesWithNode/internal/oracle/testdata/error_classes|TestError(ClassMutants|ReportingMutants|Boundaries)|TestStockNodeErrorExit' -count=1 -v` | exit 0; oracle 7.884s, lower 0.653s; four fixtures, 25 compiled mutants and refusal probes passed | `/tmp/adamic-errors-unit-final.log` |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts` | exit 0, 9.884s; counts recorded | `/tmp/adamic-errors-counts-final.log` |
| `go test ./internal/lower -count=1` after the constructor guard | exit 0, 2.661s | `/tmp/adamic-errors-lower-final.log` |
| `python3 cloud/error-classes-mutants.py --logs /tmp/adamic-errors-source-mutants-reproduction` at the 15-mutant version | exit 0; all 15 assertion checks caught their mutants | `/tmp/adamic-errors-source-mutants-reproduction.log` |
| `python3 cloud/error-classes-mutants.py --only nullable-generic-receiver --logs /tmp/adamic-errors-source-mutant-nullable` | exit 0; refusal assertion caught removal of receiver guard | `/tmp/adamic-errors-source-mutant-nullable.log` |
| `python3 cloud/error-classes-mutants.py --only inherited-constructor-field --logs /tmp/adamic-errors-source-mutant-constructor` | exit 0; refusal assertion caught restored own-field read | `/tmp/adamic-errors-source-mutant-constructor.log` |

The complete gate finished; its Unicode-property package took 694.505s. Every stage-1 package passed in that initial run. After the implementation fixes, the affected packages were rerun instead of repeating that long sweep. The final generic receiver/default and constructor probes were added after the most recent whole-oracle run, then tested by the uncached error unit and final lowering run respectively. A final complete uncached gate remains for integration after the reserved emitter correction.

Cohere was built with `go build -o /tmp/adamic-errors-cohere ./command/cohere` in the pinned submodule. The repository-wide `--no-fix` check failed on existing formatting findings and a nested parser failure in cohere/internal/lint/rules/react/conformance/testdata/fixtures/error.bug-hir_loc_diff_2.js. The pinned binary does not recognize .a files; the early selected .a formatting command checked nothing. The workaround was byte-identical .ts copies of the four fixtures under `/tmp/adamic-errors-cohere-fixtures`, using the same compiler options and original prelude. Formatting was applied back as a patch. The fixture has a documented max-classes-per-file exception because one program tests a shared hierarchy.

Final scoped command:

```text
/tmp/adamic-errors-cohere --directory /tmp/adamic-errors-cohere-fixtures --no-fix error_classes.ts error_classes_uncaught.ts error_classes_uncaught_empty.ts error_classes_uncaught_name.ts
exit 0
276 rules, 4 checked, 100% Adamic-ready (4 of 4), only 4 of 5 files
```

The fifth file is the existing prelude, which was loaded for type checking and excluded from the selected-file lint/format scope. Log: `/tmp/adamic-errors-cohere-ts-fixtures.log`.

Every compiled mutant built under the native compiler's normal strict flags. Semantic mutants disagreed with Node in stdout or exit status; reporting mutants changed status or dropped earlier stdout. Build errors do not count. The permanent tests are TestErrorClassMutants and TestErrorReportingMutants:

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
| Restore uncaught panic status | Node exit 1 versus native 70 |
| Skip buffered stdout flushing | Node before/finally output versus missing native output |

The source mutants are Go overlays, never edits to the working tree. Each compiled, then failed its intended assertion. Refusal mutants made Lower accept the refused probe; the registration mutant made the ownership coverage test find unrecorded writes. Reproduce all 17 with `python3 cloud/error-classes-mutants.py --logs /tmp/adamic-error-source-mutants` after sourcing the tools environment:

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

Early failures were real findings, not counted mutants: the whole gate caught missing synthetic write registrations; a stdout regression caught missing managed-buffer flushing. Intermediate mutant harness bugs (nil spread cloning, byte/string conversion and shared slice change detection) were fixed before the final passing mutant run.

Not covered or deliberately refused:

- Stack reads and writes are NotYet because native frames cannot report a JavaScript source stack. Full raw Node stderr, source locations, frames and extra-property inspection are not reproduced.
- ErrorOptions variables, erased unknown causes, mutable cause links and mutable unknown aliases are refused with ownership/representation reasons. The tested cause values are string, number, boolean and nested Error objects.
- Error spreading and structural views are refused because JavaScript descriptors and property absence are not represented. Error.toString overrides are refused until the synthetic built-in method's override ABI is checked.
- Generic prototype calls with optional/erased fields, nullable object receivers, arrays/functions or object coercion are NotYet. Required string fields containing undefined are supported and tested.
- Error, TypeError and RangeError constructors are directly held by these fixtures. ReferenceError, SyntaxError, URIError and EvalError use the same implementation but were not directly exercised by this unit's fixtures. AggregateError and bare Error(...) calls were not implemented.
- Non-Error throws remain refused. Other unguarded array-length and code-point failures cannot unwind to catch and are refused inside try when arguments are not proven safe. Normalization inside try requires a bounded constant input. Stack exhaustion, other string-size failures and temporal-dead-zone failures still cannot unwind to catch.
- No performance measurements or GraphQL source rewrite were made. The full repository Cohere gate is not green. The ordinary whole oracle is not green until the reserved emitter patch is applied and verified.

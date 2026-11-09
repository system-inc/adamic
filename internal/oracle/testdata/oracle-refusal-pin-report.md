# Oracle refusal pins, task 5jnptpn

Base: origin/main d7b05dc830678b13734bcf0bc4c809aa9b75c8fe.
Initial measurement base: fd8b74b9bd4fa5d8012dfcc0c0252c3f1964cd4d.

## Method and provenance

Read each fixture, its registration, its path history and the adding commit.
The requested adding commits include moves into the current directories.
The nullable logical-and witness originally appeared in 967b1f05f; that
commit explicitly records "a value of type boolean | null" at 12:16.

Each measurement uses a temporary parallel test in the oracle package calling
its existing lowered helper: load.Load followed by lower.Lower in process.
The temporary test extracts lower.NotYet or lower.Refused with errors.As and
logs Where and What. Each historical compiler is built in its own detached
worktree, with that commit's exact cohere and TypeScript gitlinks. No current
compiler is substituted for a historical compiler.

| Adding commit | cohere | TypeScript |
| --- | --- | --- |
| 5b682673 | 715ba94f3608a6500086b1076ce5cb7e51b836db | 8d550c837c90bd1805b047b7eeccc2baac2d5e7a |
| 56d33551, cdf62f5f, 506441dd | 7945d102a6c18dd36adf9114a758ce646e8b2359 | d92d9bfee114c80be2c375d72edae966176e3a4f |
| 141fec77 | 0183ccf8c57a775f874ca73be1adfc06e9ad1562 | d92d9bfee114c80be2c375d72edae966176e3a4f |
| fd8b74b9 and d7b05dc8 | 06af2e399c6c1fb93fd570b774027942747400b4 | d92d9bfee114c80be2c375d72edae966176e3a4f |

All fixture paths below are relative to internal/oracle/testdata.

## Registered intent

| Fixture | Registration | Adding commit | Intended construct and reason |
| --- | --- | --- | --- |
| census_small_stopped/census_small_default_stopped.a | census_small_test.go | 5b682673 | Computed wildcardMatchers[usage] read inside the destructured parameter default at line 28. The header requires a proved complete finite key set and compatible selected-field representations. |
| census_small_stopped/census_small_optional_stopped.a | census_small_test.go | 5b682673 | Dynamic this in the function expression at 11:22 and bind at 8:12. Bound closures must retain the receiver and erase this from runtime parameters. |
| logical_and_reference_refused/logical_and_reference_null.a | logical_and_reference_null_test.go | 56d33551 | boolean &#124; null produced by none && none.ready at 12:16. A Maybe with boolean &#124; undefined would print the wrong absent value. |
| notyet_element_access/fields.a | element_access_fields_test.go | cdf62f5f | Computed field assignment value[key] at 12:5. The move commit calls this the prior unlowered write witness; the preceding finite reads are supported. |
| presence_refused/optional_field_write.a | presence_refused_test.go | 506441dd | Writing state.pending at 9:1 after its literal omitted the optional own slot. Without the guard native execution could panic on a missing field. |
| records_refuse/r24.a | records_test.go | 141fec77 | String(values) at 4:20 on an index-signature object. Object string conversion requires ToPrimitive, which is not lowered. |

Logs are saved under /tmp/5jnptpn-*.log. Commands report their child's exit
code; logs are never piped into head or tail.

## Mutant method

The four new pins have no shared assertion helper or changed harness path.
Each calls lowered directly; it starts no Go or compiler subprocess.
Each top-level test calls t.Parallel. A repository-wide rg search before
adding the tests returned exit 1, confirming all four names were unused.

An external Go overlay replaces only the return in lower.notYet, adding
the following code and an os import:

```go
where := l.program.Where(node)
if target := os.Getenv("ORACLE_PIN_MUTANT_TARGET"); target != "" && strings.Contains(where, target) {
    switch os.Getenv("ORACLE_PIN_MUTANT_MODE") {
    case "reason":
        what = "mutant changed refusal reason"
    case "location":
        where = strings.Replace(where, target, strings.Split(target, ":")[0]+":999:", 1)
    }
}
return &NotYet{Where: where, What: what}
```

Build once with:

```sh
go test -c -overlay /tmp/5jnptpn-mutants/overlay.json \
  -o /tmp/5jnptpn-mutants/oracle-mutant.test ./internal/oracle
```

The overlay maps the absolute internal/lower/diagnostics.go path to
/tmp/5jnptpn-mutants/diagnostics.go. It names no test source.
Run the binary from internal/oracle with -test.v and -test.run selecting all
four new pins. With both variables unset, all four pass and exit is 0.
For each target below, set ORACLE_PIN_MUTANT_TARGET and set
ORACLE_PIN_MUTANT_MODE to reason, then location. Each run exits 1 and has
exactly the indicated test fail; the other three pass.

| Target | Failing test for both mutants | Evidence |
| --- | --- | --- |
| census_small_optional_stopped.a:11: | TestCensusSmallDynamicReceiverRefusalPin | [census mutants](census_small_stopped/pin-mutants.md) |
| fields.a:12: | TestComputedFieldsWriteRefusalPin | [field mutants](notyet_element_access/pin-mutants.md) |
| optional_field_write.a:9: | TestAbsentOptionalOwnFieldWriteRefusalPin | [presence mutants](presence_refused/pin-mutants.md) |
| r24.a:4: | TestRecordR24StringConversionRefusalPin | [record mutants](records_refuse/pin-mutants.md) |

The inactive-overlay control and all eight mutant outcomes were checked
programmatically, including the exact set of failing tests. Per-directory
evidence preserves each failure output line.

## Measured comparison

All twelve measurements returned *lower.NotYet; none returned *lower.Refused.
Locations below are relative to the fixture named in the first column.

| Fixture | Adding commit stop: location and exact What | Main stop: location and exact What | Pinned |
| --- | --- | --- | --- |
| census_small_stopped/census_small_default_stopped.a | 5b682673 at 28:5: a parameter that isn't a plain name | d7b05dc8 at 28:5: a parameter that isn't a plain name | No, finding |
| census_small_stopped/census_small_optional_stopped.a | 5b682673 at 11:22: a function expression with a this parameter (dynamic receivers are not implemented) | d7b05dc8 at 11:22: a function expression with a this parameter (dynamic receivers are not implemented) | Yes |
| logical_and_reference_refused/logical_and_reference_null.a | 56d33551 at 12:16: a value of type boolean &#124; null | d7b05dc8 at 12:16: a template interpolating a union with an object, an array, a map or a function in it | No, finding |
| notyet_element_access/fields.a | cdf62f5f at 12:5: assigning an element of a value | d7b05dc8 at 12:5: assigning an element of a value | Yes |
| presence_refused/optional_field_write.a | 506441dd at 9:1: writing a possibly absent optional own field | d7b05dc8 at 9:1: writing a possibly absent optional own field | Yes |
| records_refuse/r24.a | 141fec77 at 4:20: String conversion of an object, array, map or function (ToPrimitive is not lowered) | d7b05dc8 at 4:20: String conversion of an object, array, map or function (ToPrimitive is not lowered) | Yes |


## Findings left unpinned

1. census_small_stopped/census_small_default_stopped.a was already stopping
   before its declared target when registered in 5b682673. Its intended
   computed read wildcardMatchers[usage] is at 28:111. Both historical and
   current measurements stop instead at the destructured parameter, 28:5,
   with "a parameter that isn't a plain name". This is an intent mismatch
   already present at registration, not evidence of subsequent diagnostic drift.
2. logical_and_reference_refused/logical_and_reference_null.a has drifted.
   Its header and original commit name the boolean | null refusal at 12:16.
   The requested adding/move commit 56d33551 measures exactly
   "a value of type boolean | null" there. Main instead measures
   "a template interpolating a union with an object, an array, a map or a function in it"
   at the same 12:16. Matching coordinates do not make the construct or reason
   equivalent. The intended type refusal is no longer what this witness tests.

No fixture or compiler behavior was changed to hide these findings. The two
registrations remain lowers: false, and neither mismatch is cemented by a pin.

## Validation

Go checks ran with /workspace/adamic-tools/env.sh sourced. Setup ran as
bash cloud/setup.sh and exited 0. Full logs are files.

| Check | Result | Log |
| --- | --- | --- |
| Fixture provenance and adding-commit diffs | Exit 0 for each command | /tmp/5jnptpn-source-history.log |
| Main measurements, six cases | Exit 0 | /tmp/5jnptpn-measure-main.log |
| Historical measurements | Exit 0 for each of five builds/runs | /tmp/5jnptpn-measure-{5b682673,56d33551,cdf62f5f,506441dd,141fec77}.log |
| All tests declared in the four changed Go files, including existing record mutants and source Node witnesses | Exit 0, 12 top-level tests, no skips | /tmp/5jnptpn-touched-tests.log |
| TestNativeAgreesWithNode numbered shards filtered to exactly the six fixtures | Exit 0, six fixture subtests passed | /tmp/5jnptpn-native-six.log |
| Mutant binary build and inactive control | Exit 0 | /tmp/5jnptpn-mutant-build.log and /tmp/5jnptpn-mutant-control.log |
| Four reason mutants and four location mutants | Each exit 1, exactly its targeted pin failed | Per-directory pin-mutants.md |
| go vet ./internal/oracle | Exit 0, empty log | /tmp/5jnptpn-vet.log |
| gofmt -l internal/oracle/*.go | Exit 0, empty log | /tmp/5jnptpn-gofmt-list.log |

The affected-test selector was:

```text
^(TestCensusSmallStoppedSourceOnNode|TestCensusSmallDynamicReceiverRefusalPin|TestComputedFieldsWriteRefusalPin|TestInUnionNarrowingIsRefused|TestAbsentOptionalOwnFieldWriteRefusalPin|TestRecordJavaScriptAgreesWithNode|TestRecordOperationMutants|TestRecordPrototypeMutant|TestRecordOwnershipMutant|TestRecordCoalesceMutant|TestRecordNarrowingMutant|TestRecordR24StringConversionRefusalPin)$
```

The native fixture selector was:

```text
^TestNativeAgreesWithNode_[0-9]+$/internal/oracle/testdata/(census_small_stopped|logical_and_reference_refused|notyet_element_access|presence_refused|records_refuse)/(census_small_(default|optional)_stopped[.]a|logical_and_reference_null[.]a|fields[.]a|optional_field_write[.]a|r24[.]a)$
```

Both selectors ran via go test ./internal/oracle -run SELECTOR -count=1 -v,
with stdout and stderr redirected to the indicated log. Each command's wrapper
reported that child's exit code. The main measurement test was removed after
measurement; historical measurement tests remain only in detached worktrees.


## Main-tip refresh

The final fetch advanced origin/main from fd8b74b9 to d7b05dc8.
The four new commits change rest-pack graph metadata and native allocation,
with no checker gitlink or oracle source changes. The isolated patch was
preserved, the branch fast-forwarded, and the patch restored. All six current
measurements, affected-file tests, the six native fixture checks, mutant runs,
vet and formatting checks were refreshed on d7b05dc8. The initial measurement
log remains /tmp/5jnptpn-measure-initial-main.log; the main log named above is
the final-tip measurement. All six stops are identical at the two main tips.
Historical measurements remain those of the requested adding commits.

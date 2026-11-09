# Census default agreement mutant evidence

Parent: 6a1f7e79fa0c3d17b7bcc3ee9de4e4c698bb3a10,
the first refusal pins head. This change carries only the census agreement
move, registration, runtime count record, and this evidence. The nullable
logical-and fixture is unchanged from the parent.

## Controls

Setup runs with /workspace/adamic-tools/env.sh sourced and bash cloud/setup.sh.
The two tests declared in census_small_test.go pass:
TestCensusSmallStoppedSourceOnNode and TestCensusSmallDynamicReceiverRefusalPin.
The unmodified census fixture passes
TestNativeAgreesWithNode_109/internal/oracle/testdata/census_small_default.a.
That agreement test compares Node, JavaScript, sanitized native, release
native and the leak check through the existing harness.

Logs: /tmp/census-agreement-target-setup.log,
/tmp/census-agreement-file-tests.log,
/tmp/census-agreement-agreement-control.log.

## Default matcher output mutant

Failing test:
TestNativeAgreesWithNode_109/internal/oracle/testdata/census_small_default.a.

Node reads the committed fixture. For the source read by lowering, an external
Go overlay of internal/load/load.go replaces only this fixture's default
wildcardMatchers[usage] with:

```typescript
{ singleAsteriskRegexFragment: '[^/]*', doubleAsteriskRegexFragment: '.*', replaceWildcardCharacter: (match: string): string => 'MUTANT' }
```

The replacement callback is used by the third defaulted call. Its lowered
native program prints one/MUTANT where the unchanged Node source prints
one/[^/]*. Both processes finish with exit 0 and empty stderr. The explicit
matcher argument on the fifth call remains unchanged.

The temporary overlay replaces Load with the following code. It does not
change any committed loader, compiler, or harness source.

```go
func Load(paths []string) (*Program, error) {
    overlay := map[string]string{}
    for _, path := range paths {
        if filepath.Base(path) != "census_small_default.a" { continue }
        data, err := os.ReadFile(path)
        if err != nil { return nil, err }
        before := "selectedMatcher: WildcardMatcher = wildcardMatchers[usage],"
        after := "selectedMatcher: WildcardMatcher = { singleAsteriskRegexFragment: '[^/]*', doubleAsteriskRegexFragment: '.*', replaceWildcardCharacter: (match: string): string => 'MUTANT' },"
        if strings.Count(string(data), before) != 1 { return nil, fmt.Errorf("census mutant: default not found exactly once") }
        overlay[path] = strings.Replace(string(data), before, after, 1)
    }
    return load(paths, overlay, nil)
}
```

The external overlay file maps the worktree's absolute internal/load/load.go
to /tmp/census-agreement-mutants/load.go. Run from the repository root:

```sh
go test -overlay /tmp/census-agreement-mutants/overlay.json ./internal/oracle   -run '^TestNativeAgreesWithNode_109$/internal/oracle/testdata/census_small_default[.]a$'   -count=1 -v
```

Child exit code: 1. Log: /tmp/census-agreement-default-mutant.log.
The only failing fixture subtest is the census default. The JavaScript backend
also disagrees because it consumes the same mutated lowering input.

Exact native comparison failure and output lines:

```text
oracle_test.go:834: stdout differs
node:   exit 0, stdout "one/two\nmissing\none/[^/]*\n(?:one(?:/[^./])?)?\none/(?:[^./][^/]*)?\n", stderr ""
native: exit 0, stdout "one/two\nmissing\none/MUTANT\n(?:one(?:/[^./])?)?\none/(?:[^./][^/]*)?\n", stderr ""
```

An exploratory mutant of only the single-asterisk fragment stayed green:
none of the defaulted calls uses that field. The callback mutation above
targets the defaulted call that actually exercises replacement. The exploratory
log is /tmp/census-agreement-default-fragment-survivor.log.

## Recorded-count value mutant

Control test:
TestCountsAreRecorded_013/internal/oracle/testdata/census_small_default.a.
It passes with the committed count row (child exit 0).

The actual .counts file was temporarily edited, changing only the first numeric
value, allocations, from 67 to 68. The row remains syntactically valid.
Run without -update-counts so the comparison reads the edited record:

```sh
go test ./internal/oracle   -run '^TestCountsAreRecorded_013$/internal/oracle/testdata/census_small_default[.]a$'   -count=1 -v
```

Child exit code: 1. The only fixture subtest that fails is
TestCountsAreRecorded_013/internal/oracle/testdata/census_small_default.a.
Exact failure and comparison lines:

```text
counts_test.go:459: the counts changed and testdata/counts/runtime/internal/oracle/testdata/census_small_default.a.counts wasn't updated with them (go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts):
recorded: | internal/oracle/testdata/census_small_default.a | 68 | 67 | 120 | 164 | 13 | 0 |
measured: | internal/oracle/testdata/census_small_default.a | 67 | 67 | 120 | 164 | 13 | 0 |
```

The original file was restored in a finally block and checked byte for byte
before staging. The committed value is 67, not the mutant's 68.

Logs: /tmp/census-agreement-counts-control.log,
/tmp/census-agreement-counts-mutant.log,
/tmp/census-agreement-mutants/count-values.txt.

## Verification

- The tests in census_small_test.go pass.
- The scoped census default native agreement passes.
- The owning count shard's scoped census default control passes.
- The count updater measured this .counts row on the exact parent compiler.
- Both output and numeric-count mutants exit 1 with the expected red fixture.
- go vet ./internal/oracle passes; /tmp/census-agreement-vet.log is empty.
- gofmt -l internal/oracle/*.go passes; /tmp/census-agreement-gofmt.log is empty.

A broad count-product warm-up also ran all 32 count wrappers. The census
subtest passed, but two unrelated program-region count products required an
unset ADAMIC_TYPESCRIPT_SOURCE. This exploratory run is recorded in
/tmp/census-agreement-counts-broad-warmup.log. The control and mutant above
select census's owning shard 013 and pass/fail for the census record itself.
No workaround or harness change for the unrelated products is included.

# Census default agreement mutant evidence

Parent: 6a1f7e79fa0c3d17b7bcc3ee9de4e4c698bb3a10,
the first refusal pins head. This is one amended census-only commit.
The nullable logical-and fixture is unchanged from that parent.

## Selection witness and controls

wildcardMatchers now holds three separate objects: filesMatcher,
directoriesMatcher and excludeMatcher. The files matcher keeps [^/]* and .*;
the other matchers have separately named single-asterisk and double-asterisk
fragments and distinct replacement callbacks.

Each of the three defaulted calls uses one/*/x*?/**/leaf, exercising a leading
asterisk, embedded asterisk/question replacements and an interior double
asterisk. The returned strings identify the selected matcher through their
distinct fragments and callbacks. The fourth call explicitly supplies the
original matcher and is a control for mutations of default lookup.

Direct Node stdout (exit 0, empty stderr):

```text
files:one/(?:[^./][^/]*)?/x[^/]*[^/].*/leaf
directories:(?:one(?:/(?:[^./](directories-single))?(?:/x(directories-star)(directories-question)(directories-double)(?:/leaf)?)?)?)?
exclude:one/(exclude-star)/x(exclude-star)(exclude-question)(exclude-double)/leaf
explicit:one/(?:[^./][^/]*)?
```

The unmodified fixture passes
TestNativeAgreesWithNode_109/internal/oracle/testdata/census_small_default.a.
That existing test compares Node, JavaScript, sanitized native, release native
and the leak check. The external mutant binary also passes with its selector
environment variable unset.

## Wrong-key read mutant

Failing test:
TestNativeAgreesWithNode_109/internal/oracle/testdata/census_small_default.a.

For lowering only, change the default read from wildcardMatchers[usage] to
wildcardMatchers.files. Node reads the unchanged committed fixture. The
native files line remains correct, but directories and exclude lose their
distinct matcher fragments. The explicit argument line remains byte-identical.
These properties were asserted when checking the mutant's captured stdout.

Child exit code: 1. Exact failure and comparison lines:

```text
oracle_test.go:834: stdout differs
node:   exit 0, stdout "files:one/(?:[^./][^/]*)?/x[^/]*[^/].*/leaf\ndirectories:(?:one(?:/(?:[^./](directories-single))?(?:/x(directories-star)(directories-question)(directories-double)(?:/leaf)?)?)?)?\nexclude:one/(exclude-star)/x(exclude-star)(exclude-question)(exclude-double)/leaf\nexplicit:one/(?:[^./][^/]*)?\n", stderr ""
native: exit 0, stdout "files:one/(?:[^./][^/]*)?/x[^/]*[^/].*/leaf\ndirectories:(?:one(?:/(?:[^./][^/]*)?(?:/x[^/]*[^/].*(?:/leaf)?)?)?)?\nexclude:one/[^/]*/x[^/]*[^/].*/leaf\nexplicit:one/(?:[^./][^/]*)?\n", stderr ""
```

The JavaScript backend also disagrees because it consumes the same mutated
lowering input. The only failing fixture subtest is census_small_default.a.
Log: /tmp/census-agreement-2-fixed-key-mutant.log.

## Default callback output mutant

Failing test:
TestNativeAgreesWithNode_109/internal/oracle/testdata/census_small_default.a.

Replace the default value for lowering only with a matcher whose callback
always returns MUTANT. Node continues to read the committed source. The
defaulted output changes; the explicit argument line remains unchanged.

Child exit code: 1. Exact failure and comparison lines:

```text
oracle_test.go:834: stdout differs
node:   exit 0, stdout "files:one/(?:[^./][^/]*)?/x[^/]*[^/].*/leaf\ndirectories:(?:one(?:/(?:[^./](directories-single))?(?:/x(directories-star)(directories-question)(directories-double)(?:/leaf)?)?)?)?\nexclude:one/(exclude-star)/x(exclude-star)(exclude-question)(exclude-double)/leaf\nexplicit:one/(?:[^./][^/]*)?\n", stderr ""
native: exit 0, stdout "files:one/(?:[^./][^/]*)?/xMUTANTMUTANT.*/leaf\ndirectories:(?:one(?:/(?:[^./][^/]*)?(?:/xMUTANTMUTANT.*(?:/leaf)?)?)?)?\nexclude:one/MUTANT/xMUTANTMUTANT.*/leaf\nexplicit:one/(?:[^./][^/]*)?\n", stderr ""
```

Log: /tmp/census-agreement-2-callback-mutant.log.

## Reproducing the source mutants

A temporary external Go overlay replaces Load in internal/load/load.go with
the following function. Existing imports supply os, filepath, strings and fmt.
No loader, compiler, test helper or harness change is committed.

```go
func Load(paths []string) (*Program, error) {
 mode := os.Getenv("CENSUS_DEFAULT_MUTANT")
 if mode == "" { return load(paths, nil, nil) }
 overlay := map[string]string{}
 for _, path := range paths {
  if filepath.Base(path) != "census_small_default.a" { continue }
  data, err := os.ReadFile(path)
  if err != nil { return nil, err }
  before := "selectedMatcher: WildcardMatcher = wildcardMatchers[usage],"
  after := ""
  switch mode {
  case "fixed-key":
   after = "selectedMatcher: WildcardMatcher = wildcardMatchers.files,"
  case "callback":
   after = "selectedMatcher: WildcardMatcher = { singleAsteriskRegexFragment: '[^/]*', doubleAsteriskRegexFragment: '.*', replaceWildcardCharacter: (match: string): string => 'MUTANT' },"
  default:
   return nil, fmt.Errorf("unknown census mutant %q", mode)
  }
  if strings.Count(string(data), before) != 1 { return nil, fmt.Errorf("census default not found exactly once") }
  overlay[path] = strings.Replace(string(data), before, after, 1)
 }
 return load(paths, overlay, nil)
}
```

The external overlay file maps the repository's absolute internal/load/load.go
to /tmp/census-agreement-2-mutants/load.go. Build the observation binary:

```sh
go test -c -overlay /tmp/census-agreement-2-mutants/overlay.json -o /tmp/census-agreement-2-mutants/oracle-mutants.test ./internal/oracle
```

Run from internal/oracle, with CENSUS_DEFAULT_MUTANT unset for the green control,
set to fixed-key for the wrong-key mutant, or set to callback for the callback
mutant:

```sh
CENSUS_DEFAULT_MUTANT=fixed-key /tmp/census-agreement-2-mutants/oracle-mutants.test -test.run '^TestNativeAgreesWithNode_109$/internal/oracle/testdata/census_small_default[.]a$' -test.v
```

Control child exit code: 0. Both mutant child exit codes: 1.
Logs: /tmp/census-agreement-2-mutant-build.log and
/tmp/census-agreement-2-control-mutant.log, plus the mutant logs above.

## Recorded-count value mutant

The fixture's strengthened calls changed the measured row from
67 / 67 / 120 / 164 / 13 / 0 to:

```text
| internal/oracle/testdata/census_small_default.a | 136 | 136 | 192 | 249 | 26 | 0 |
```

The existing scoped updater recorded the new row on the parent compiler.
The scoped unmodified count test passes (child exit 0).

The actual .counts file was temporarily edited, changing only the first
numeric value, allocations, from 136 to 137. Its syntax stays valid.
The existing comparison test runs without -update-counts:

```sh
go test ./internal/oracle -run '^TestCountsAreRecorded_013$/internal/oracle/testdata/census_small_default[.]a$' -count=1 -v
```

Failing test:
TestCountsAreRecorded_013/internal/oracle/testdata/census_small_default.a.
Child exit code: 1. Exact failure and comparison lines:

```text
counts_test.go:459: the counts changed and testdata/counts/runtime/internal/oracle/testdata/census_small_default.a.counts wasn't updated with them (go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts):
recorded: | internal/oracle/testdata/census_small_default.a | 137 | 136 | 192 | 249 | 26 | 0 |
measured: | internal/oracle/testdata/census_small_default.a | 136 | 136 | 192 | 249 | 26 | 0 |
```

The original file was restored in a finally block and verified byte for byte.
The committed allocation count is 136.

Logs: /tmp/census-agreement-2-counts-update.log,
/tmp/census-agreement-2-counts-control.log,
/tmp/census-agreement-2-counts-mutant.log,
/tmp/census-agreement-2-mutants/count-values.txt.

## Verification

All commands source /workspace/adamic-tools/env.sh. Setup uses bash
cloud/setup.sh. Every process wrapper reports its child's own exit code.
All logs go to files, never through head or tail.

| Check | Result | Log |
| --- | --- | --- |
| Setup | Exit 0 | /tmp/census-agreement-2-setup.log |
| Direct Node | Exit 0, empty stderr | /tmp/census-agreement-2-node.stdout and .stderr |
| Both tests in census_small_test.go | Exit 0 | /tmp/census-agreement-2-file-tests.log |
| Scoped native agreement | Exit 0 | /tmp/census-agreement-2-agreement-control.log |
| Scoped count update and count comparison | Exit 0 each | /tmp/census-agreement-2-counts-update.log and -counts-control.log |
| Count-file census and family wrapper currency | Exit 0 | /tmp/census-agreement-2-registration-checks.log |
| Wrong-key and callback mutants | Exit 1 each, expected output difference | Mutant logs above |
| Numeric-count mutant | Exit 1, expected count difference | /tmp/census-agreement-2-counts-mutant.log |
| go vet ./internal/oracle | Exit 0, empty log | /tmp/census-agreement-2-vet.log |
| gofmt -l internal/oracle/*.go | Exit 0, empty log | /tmp/census-agreement-2-gofmt.log |

The file-test selector is
^(TestCensusSmallStoppedSourceOnNode|TestCensusSmallDynamicReceiverRefusalPin)$.
Agreement and count selectors are the exact paths shown above; all go test
runs use -count=1 -v.

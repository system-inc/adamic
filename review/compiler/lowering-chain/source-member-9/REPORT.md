Built both refusal rulings: Function annotations and uninitialized merged fields.
Function commit c3782de63; merged-field commit is the branch tip reported with delivery.
Fifteen focused leaves, Node and both backends, ASan/UBSan, leaks, vet and counts passed.
Six independent mutants failed their intended behavioral tests.
Constructor reflection, other Function properties and the full package gate were not covered.

## Main observations

The physical temporary .ts adapter uses the same bytes as each checked-in .a witness.
The corrected baseline probe is /tmp/refusal-rulings-baseline.log.

| Program | Main .a | Main .ts | Ruling |
|---|---|---|---|
| Unused Function parameter | Accepted, prints ok | Accepted, prints ok | Refused in .a; accepted in .ts |
| Closure passed to Function parameter, .length observed | JavaScript prints undefined; clang rejects incompatible closure/object pointers | Same | .a annotation refused; .ts holding and arity accepted |
| Call through Function | NotYet, a call to an Identifier | Same | Permanent refusal with call-signature fix |
| Class/interface phantom field, unused | Accepted, prints ok | Same | .a declaration refused; .ts holding accepted |
| Phantom field read | JavaScript prints undefined; native exits 70 with generic compiler-bug panic | Same | .a declaration refused; .ts checked read naming the field |
| Initialized merged field | Accepted, prints 7 | Same | Accepted |

## Function implementation

The refusal pass checks library Function type references throughout .a files, before
branch pruning. The diagnostic includes its path and tells the author to write a
call signature with parameters and a result. TypeScript values use closure storage;
the refusal pass forbids invocation, construction and call/apply/bind through the
opaque type. Closure capture traversal still participates in cycle checking.

Function.length reads closure metadata. SourceLength counts parameters before the
first default or rest; optional parameters count, erased this parameters do not.
Named-function adapters retain their source length, as do canonical nested values.
Both backends preserve identity and the captured cells.

The old branch 5fbf1ab7 supplied the annotation refusal's starting point. Its old
base and unrelated changes were not merged. Constructor reflection and other
Function properties are not certified here; unsupported observations remain stops.
No unchecked any representation was added.

## Verification

Setup used GOPROXY=https://proxy.golang.org|direct. nproc=5, CPU quota=4.
Setup timing lines: node 0.022s; Go 0.025s; markdown skip 0.008s;
markdown 0.069s; submodules 0.080s; clang 0.189s; build 41.056s;
test binaries deferred 41.176s; cache warm 41.178s; done 41.211s.

Commands and logs:

- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestFunction(Adamic|TypeScript|CallSignature)|^TestNativeAgreesWithNode$/internal/oracle/testdata/entries_scanner.a$' -count=1 -v: /tmp/refusal-rulings-function-tests.log.
- python3 review/compiler/refusal-rulings-main/mutants.py: /tmp/refusal-rulings-function-mutants.log and /tmp/refusal-rulings-mutants/*.log. Mutants use Go overlays, never altered source artifacts. Removing the annotation guard makes the unused .a parameter accepted; its pinned refusal test fails. Removing invocation refusal changes Refused to NotYet; the call diagnostic test fails. Setting SourceLength to zero changes backend stdout against Node; the arity oracle fails. No mutant is credited for a compiler failure.
- go vet ./internal/lower ./internal/ir ./internal/javascript ./internal/native ./internal/oracle: /tmp/refusal-rulings-function-vet.log.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts: /tmp/refusal-rulings-function-counts.log.

All six new test leaves are top-level parallel tests. Latest uncached durations:
unused .a refusal 0.21s; length .a refusal 0.22s; .ts call refusal 0.26s;
.ts unused 0.52s; .ts length 0.46s; call-signature alternative 0.41s.
The first cold arity leaf took 9.18s. All are below 60s.

Counts adds the call-signature fixture and the two .ts-mode adapted witnesses.
The signature fixture counts its closure and printed strings; the unused function
allocates nothing; the arity fixture counts function values, captured storage,
boxed arity observations and printed strings. Existing rows stay unchanged.
A temporary boxed Function representation changed entries_scanner's constructor
record slot and failed its Node oracle. Closure storage corrected that regression;
the existing scanner oracle was rerun uncached. The failed candidate's counts are
not the delivered counts.

## Merged-field implementation

Task h3tshdq. A full declaration walk rejects a required merged field in .a when
no class member supplies it, including unused classes. The path points to the
class declaration, with advice to initialize the member in that class or use a
separate interface with a checked view. Actual inherited class members and
constructor parameter properties count; an interface declaration does not.
The checker's strict initialization proof covers ordinary required class fields;
declare and definite-assignment assertions do not supply initialization evidence.
Optional fields do not promise presence. Namespace merging is left to its own rule.

In .ts, the declaration and unused value remain accepted. A phantom member read
registers the existing checked-view contract for its declared type. Dotted and
constant-string bracket reads use the same presence, initialization and runtime
representation check. No second runtime checking mechanism was added.

The explicit ruled divergence list is mergedFieldRuledDivergences in
internal/oracle/merged_fields_main_test.go. For each missing witness, Node prints
undefined and exits 0. Both backends stop with exit 70 and the pinned text:

    adamic: panic: field read failed: new Missing().promised is not initialized; expected number, found missing

The bracket witness names new Missing()['promised'] in the same message.
Initialized merged fields print 7 on Node and both backends. ASan/UBSan runs cover
both stops and successful programs; successful programs also pass leak checks.
A loud exit deliberately abandons live state, so stopped programs have no leak
success claim.

The declaration mutant removes the .a declaration refusal and makes the unused
class accepted. The checked-read mutant clears CheckedFields and loses the named
runtime check. The bracket mutant drops the contract-bearing expression path.
Each fails its focused pinned test without a compiler failure.

## Final records

Evidence, including the mutant runner, lives in this review directory. The first
commit's report and runner were moved here in the second ruling's commit, following
the evidence-layout addendum. No evidence was put under cloud/reports/.

Final focused command:

    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestFunction(Adamic|TypeScript|CallSignature)|^TestMergedField' -count=1 -v

It passed in 1.299s. logs/focused-tests.log records every leaf's seconds;
results.json lists them by name. All fifteen new leaves are top-level parallel
functions and each is below 60s. logs/scanner-and-rulings.log also records the
existing entries_scanner oracle run uncached against Node, both backends and leaks.

    python3 review/compiler/refusal-rulings-main/mutants.py
    go vet ./internal/lower ./internal/ir ./internal/javascript ./internal/native ./internal/oracle
    go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts

All passed; the final counts run took 49.244s. These commands ran focused oracles,
not whole packages or the full gate. logs/mutants.log names all six catchers and
individual logs preserve their failures. Refusals pin file, line, column, reason
and replacement advice. TypeScript mode materializes temporary .ts files only;
every new source witness committed here has the .a extension.

Counts changes, in allocations/frees/retains/releases/peak/regions order:

| New row | Counts | Explanation |
|---|---|---|
| function_signature.a | 3/3/0/3/3/0 | Closure plus printed strings |
| function_unused.a, TypeScript mode | 0/0/0/0/0/0 | Unused parameter; literal output |
| function_length.a, TypeScript mode | 20/20/4/29/13/0 | Callable values, capture storage, boxed arity and strings |
| merge_present.a | 3/3/1/4/3/0 | Instance and printed strings |
| merge_unused.a, TypeScript mode | 0/0/0/0/0/0 | No instance constructed |
| merge_missing.a, TypeScript mode | 1/0/1/1/1/0 | Stops holding its instance |
| merge_missing_computed.a, TypeScript mode | 1/0/1/1/1/0 | Same stop through brackets |
| merge_present.a, TypeScript mode | 3/3/1/4/3/0 | Same initialized instance and output |

No existing counts row changed. The required committed-HEAD lane checker is run
before push; its output is included in the delivery response. No main or area
branch was merged into or pushed to. This unit implements the two refined refusal
contracts toward the refusal roadmap work, without importing the old branch's base.

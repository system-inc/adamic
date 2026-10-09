Built Function annotation refusals and TypeScript callable holding with source arity.
Commit: recorded in the branch log; base 7459656e052d7172479a7ce14e35c314b3e622f1.
Focused Function oracles, ASan/UBSan, leak checks and refusal pins passed.
Function annotation, invocation refusal and zero-arity mutants were caught.
The merged-field ruling is recorded below when its separate commit is verified.

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
- python3 devtools/refusal-rulings/mutants.py: /tmp/refusal-rulings-function-mutants.log and /tmp/refusal-rulings-mutants/*.log. Mutants use Go overlays, never altered source artifacts. Removing the annotation guard makes the unused .a parameter accepted; its pinned refusal test fails. Removing invocation refusal changes Refused to NotYet; the call diagnostic test fails. Setting SourceLength to zero changes backend stdout against Node; the arity oracle fails. No mutant is credited for a compiler failure.
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

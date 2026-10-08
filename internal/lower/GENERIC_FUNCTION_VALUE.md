# Generic function values at concrete instantiations

October 8, 2026.

Built: contextual generic function values use the existing monomorphized function
and closure forwarder in both backends, without a backend or IR change.
Implementation: 9d1d672c; merge edbd3bd6 brings origin/main c6761c24.
Boundary follow-up: 12fa3b23; final package outputs are recorded below.
Commands: uncached lowerer and oracle packages, Node, counts, vet and formatting.
Mutants: seven source mutants and one wrong-result IR mutant, detailed below.
Limits: polymorphic values and function identity observations remain NotYet;
the library's combined fixture 25 is not certified by this isolated unit.

## What changed

A module generic such as TypeScript core.ts's `identity<T>(x: T): T` can be passed
where a concrete callback is expected, or stored in a concretely typed variable.
The contextual signature supplies the inference position. Adamic references the
pinned checker's `instantiateSignatureInContextOf` through a linkname shim; no
cohere implementation is copied. The resulting resolved signature enters the
same specialization routine as a generic call. Existing calls retain their
explicit-type-argument mutation checks.

The specialization cache retains the full checker type identity, not just a C
representation or the source declaration. Each specialized target gets its own
existing counted closure forwarder. Repeated string uses share the string
instance; number, string, object and array uses each have a distinct instance
and forwarder. Distinct object shapes also remain distinct even though both
have the same pointer representation. The tests cover the reference collision
reported in #wbxc0xz.

The fixture [generic_function_value.a](../oracle/testdata/generic_function_value.a)
holds typed variables for all four representations, string and number
`Array.prototype.map(identity)`, and `identity` passed to a concrete callback
parameter. Source Node prints:

```text
Kirk
42
Ahra
3,4
one,two
2,5
callback
```

The native and JavaScript backends match it. The native oracle includes release
execution, ASan, UBSan and the leak check. The fixture's counts are
15 allocations, 15 frees, 27 retains, 45 releases, peak 7, regions 0.
Counts regeneration changes only this new row.

Generic function values with no concrete contextual signature remain NotYet.
The former gap test now holds that boundary instead of refusing a supported
number instance. A context with multiple call signatures or generic type
parameters is also NotYet. Existing closure restrictions still apply, including
optional/rest parameters and results that cannot fit the closure convention.

JavaScript has one function identity across instantiations, while these
specialized forwarders have different pointers. Function equality (including boxed unions), Object.is
on functions, and function-valued collection identity observations therefore
stop with NotYet in programs that read a generic function as a value. This is a
conservative program-wide boundary, including ordinary closures in that program.
The source inspection sees earlier function bodies, before initializers create
forwarders. Set and Map construction also stop, since they deduplicate keys before a lookup
can be guarded. A shared runtime source identity is future work.

## Mutation proof

Run with the setup environment sourced:

```sh
python3 internal/lower/testdata/run-generic-function-value-mutants.py > /tmp/generic-value-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestGenericFunctionValueWrongResultMutant -count=1 -v -timeout 30m > /tmp/generic-value-result-mutant.log 2>&1
```

The source runner restores every file in a finally block. Each failure is an
assertion after the program loaded and lowered, not a compiler warning.

| Mutant | Catcher |
|---|---|
| Reuse one specialization key for object and array instances | `TestGenericFunctionValueDifferentReferenceRepresentations`: got 1 instance, want 2 |
| Key only by representation, erasing distinct object shape identities | `TestGenericFunctionValueReferenceInstances`: got 1 instance, want 2 |
| Look for created forwarders before guarding an earlier function comparison | `TestGenericFunctionValueEarlierIdentityComparisonIsNotYet` |
| Allow Object.is and collection searches to observe specialized pointers | `TestGenericFunctionValueIdentityCallsAreNotYet` |
| Hash and deduplicate specialized function pointers in Set/Map construction | `TestGenericFunctionValueIdentityCallsAreNotYet` |
| Compare boxed functions while guarding only plain closures | `TestGenericFunctionValueUnionIdentityIsNotYet` |
| Allow equality of specialized closure pointers | `TestGenericFunctionValueIdentityComparisonIsNotYet` |
| Replace the numeric instance's body with return -1 | Native and JavaScript source-Node comparisons: stdout differs, clean exit 0, empty stderr |

The shared-instance mutants are caught by the IR instance assertions: an identity
body can return the same pointer even when the wrong reference specialization was
reused, so output alone does not prove separate instances. No malformed-C kill is
counted. The wrong-result mutant independently proves both Node comparisons can
fail: numeric output changes from 42 to 0 and the mapped numbers from 2,5 to -1,-1.

## Commands and scope

```sh
bash cloud/setup.sh > /tmp/generic-value-setup.log 2>&1
source /workspace/adamic-tools/env.sh
node --input-type=module-typescript < internal/oracle/testdata/generic_function_value.a > /tmp/generic-value-node.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_value' -count=1 -timeout 30m > /tmp/generic-value-focus.log 2>&1
go test ./internal/lower -run 'TestGenericFunctionValue|TestWhatStageZeroCannotLower' -count=1 > /tmp/generic-value-lower-focus.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/generic-value-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/generic-value-boundary-packages.log 2>&1
go vet ./... > /tmp/generic-value-boundary-vet.log 2>&1
gofmt -l cmd internal > /tmp/generic-value-boundary-fmt.log
git diff --check
```

Setup succeeded: Go 0.029s, Node 0.030s, submodules 0.068s, markdown 0.080s,
clang 0.181s, Go build 36.619s, cache warm 36.782s, total 36.808s.
`nproc=5`, cgroup quota 4 CPUs; Go 1.27.1, Node 24.19.0, clang 20.1.8.
The generated environment is `/workspace/adamic-tools/env.sh`.
No setup failure or workaround.

The user reports host proof 24/25, stopped at fixture 25's concrete identity
value. This unit verifies that reduced shape, not the combined library branch's
entire host surface. No 25/25 claim is made. Library must merge this branch into
`codex/host-proof-combined` and rerun its fixture. No full repository gate, WASI
execution specific to this fixture, first-class polymorphic function storage,
nested generic declarations or shared runtime function identity is added here.

Final results on the merged branch, with the union and collection boundaries:

```text
ok  	github.com/system-inc/adamic/internal/lower	25.447s
ok  	github.com/system-inc/adamic/internal/oracle	123.264s
go vet ./...: exit 0, no output
gofmt -l cmd internal: exit 0, no output
git diff --check: exit 0, no output
reuse-one-instance: caught by TestGenericFunctionValueDifferentReferenceRepresentations
erase-reference-identity: caught by TestGenericFunctionValueReferenceInstances
miss-earlier-comparison: caught by TestGenericFunctionValueEarlierIdentityComparisonIsNotYet
observe-specialized-pointers: caught by TestGenericFunctionValueIdentityCallsAreNotYet
deduplicate-specialized-pointers: caught by TestGenericFunctionValueIdentityCallsAreNotYet
compare-boxed-functions: caught by TestGenericFunctionValueUnionIdentityIsNotYet
compare-specialized-pointers: caught by TestGenericFunctionValueIdentityComparisonIsNotYet
```

Before the boundary follow-up, the merged oracle package also passed uncached
in 118.571s. The final run above supersedes that earlier validation. The full
repository test gate was not run under the worker-gate exception; all touched
packages were tested, and the entire repository was vetted.

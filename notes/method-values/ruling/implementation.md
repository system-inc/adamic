# Method values implementation, October 8

60 this-free implementation-family sites were measured among 108 pinned refusal rows.
Measurement commit: b79dd6c4. Implementation commit: df6ea3ce8bcc32132ec7486fc3e9b54343a7a678.
Both commits were pushed to codex/method-values, based on compiler SHA 4721b9e180691f4f5f42b5a71166714a0c68d103.
The pinned refusal table is d35a81d36fdafccf827bad0f572d311b2a0d4deb.
The final report commit follows the implementation on the same branch.

## Observations and scope

The census records all 108 source locations in census.json. The 60 this-free
family sites are 11 factory accesses, 45 parenthesizer accesses, and four
parenthesizerRules() accesses. Five sites resolve directly to this-reading
bodies; 43 have unresolved runtime origins. Family observations are not a
closed program proof that every runtime implementation has that body.
No end-to-end compilation of these 108 tsc sites was performed.

Concrete class and object methods now produce ordinary unbound function values.
The .a proof covers bodies, default expressions, nested lexical arrows,
overrides, and concrete structural interface implementations in the program.
Ordinary nested functions have their own receiver and are excluded from the
lexical receiver scan. A this-reading .a extraction remains refused and names
both an arrow wrapper and obj.m.bind(obj) as fixes.

A .ts extraction supplies undefined to the method. A direct property read
through that receiver checks at the read, preserves preceding side effects,
and throws Adamic's TypeError with Node's message. The exception analysis
propagates the throw to try/catch. Methods with no unbound extraction keep
unchecked receiver reads; a direct bind alone does not require an undefined
check. Bind evaluates the method and bound receiver in source order, owns the
receiver, and participates in the cycle proof.

Class method values preserve identity across instances and inherited shapes by
sharing an immortal callable per lowered method. Literal methods have an
unbound wrapper owning the original closure and a weak identity cache. The
cache is cleared when its wrapper dies. Borrow and region analyses distinguish
an allocated extracted callable from a borrowed field and track bound receivers.

Explicitly unsupported: static method extraction, optional extraction,
overloaded or rest method values, callable parameters/results without a
represented slot, partial bind arguments, and receivers outside the represented
object type. An extracted receiver escaping into an alias, being observed
outside a direct property read, used in optional property access, or used in a
plain property assignment remains NotYet. Those forms need further lowering
work; this unit does not implement every JavaScript receiver operation.
Intrinsic library aliases retain their existing specialized lowering and
restrictions. In particular mutable Math.abs aliases remain refused. Unknown
interface implementations with no concrete body remain unproved in .a.
Whole-program scans have not been benchmarked on tsc.

## Verification

Every test command redirected all output to its named log. No whole-package
test run or full gate was run.

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lower ./internal/oracle -run 'TestMethodBindCycleIsRefused|TestMethodValuesProof|TestMethodValuesSelectiveChecks|TestMethodValuesTypeScript|TestNativeAgreesWithNode/internal/oracle/testdata/(method_values|library_method_values|method_closures|class_as_interface|user_iterators)' -count=1 -timeout 10m > /tmp/method-values-final-tests-3.log 2>&1
go test ./internal/lower -run 'TestAMethodReadAsAValueIsRefused|TestWhatZeroOneRefusesIsRefusedWithAFix' -count=1 > /tmp/method-values-refusals-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-counts-final.log 2>&1
python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-mutants-verified.log 2>&1
```

Observed final output:

```text
focused lower: ok 0.332s
focused oracle: ok 3.724s
refusals: ok 0.608s
counts refresh: ok 26.348s
bound-extraction: caught, exit 1
missing-this-check: caught, exit 1
receiver-without-retain: caught, exit 1
stale-unbound-cache: caught, exit 1
bound-cycle-hidden: caught, exit 1
```

The four source fixtures are free.a, bind.a, reading.a, and uncaught.a. The first
two run as .a oracle fixtures. All four also run as temporary .ts inputs, without
adding a .ts source file. Node decides stdout, stderr, and exit status byte for
byte for the JavaScript backend, native with ASan/UBSan, and release native.
Successful runs also pass leak verification. Free extraction covers callbacks,
nested arrows, callable identity, repeated weak-cache expiry, and structural
interface dispatch. Bind covers a receiver outliving its factory, callbacks,
and a nested arrow using this. Reading covers both class methods and lexical
arrows, catchability, error name/message, and side effects before the property
read. Uncaught covers an object literal and termination after its first read.

The mutants each restore exact original source bytes in finally:

| Mutant | Evidence catching it |
|---|---|
| Bind every unbound extraction to its original object | Node disagreement in reading/uncaught |
| Remove the extracted receiver check | ASan/UBSan in reading/uncaught |
| Do not retain the bound receiver | ASan heap-use-after-free in bind.a |
| Do not clear the weak unbound cache | ASan heap-use-after-free in free.a |
| Hide bound receiver edges from the cycle proof | TestMethodBindCycleIsRefused |

Mutant logs are /tmp/method-values-mutants/<name>.log. Two new fixture rows were
added to counts.md. The required full counts refresh initially exposed a
computed-name scan panic in user_iterators.a. Checker symbols replaced raw
method-name Text calls, the focused iterator oracle passed, and the complete
counts refresh subsequently passed. A static extraction experiment was removed
because constructor storage and own-method identity require separate support;
static extraction now reports NotYet rather than emitting an invalid program.

## Notes recheck

The compiler was built with:

```sh
go build -o /workspace/scratch/method-values/adamic ./cmd/adamic > /tmp/method-values-build-final.log 2>&1
```

Each of the 32 files under notes/method-values/unsupported was run with
`/workspace/scratch/method-values/adamic c <file.a>`. C stdout and diagnostic
stderr went to separate files under /tmp/method-values-notes. Every invocation
returned 1. notes-recheck.json preserves each outcome and existing header.
No example now checks, so no a-check header was removed or changed.
The four method-value headers are instance_method.a, mutable_alias.a,
optional_chain_alias.a, and string_instance_method.a. The first reads this;
the others retain the intrinsic library restrictions described above.

## Toolchain

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/method-values-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
```

Setup completed. Timing lines in seconds: node 0.077, go 0.080, clang 0.633,
markdown 1.954, submodules 24.059, go build 230.254, warm 230.406, done 230.485.
nproc returned 5; the quota is 4 CPUs. Versions: Go 1.27.1, clang 20.1.8,
Node 24.19.0. No source was copied from cohere and no protected orchestration or
oracle_test.go file was edited.

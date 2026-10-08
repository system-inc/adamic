# Method values continuation

The continuation starts at f7de26db on codex/method-values. Static, optional, and rest steps were committed and pushed before the standing
rule changed. The remaining receiver, intrinsic, and audit work is one unit,
with one push after its local fixtures, tests, and mutants pass. No protected orchestration file is edited.

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh, then
source /workspace/adamic-tools/env.sh. The setup log is
/tmp/method-values-followup-setup.log. Times in seconds: go .029, Node .036,
submodules .098, markdown .099, clang .249, build 45.016, warm 45.130,
done 45.215. nproc is 5, quota 4 CPUs.

## Static extraction

Static allocation now installs an inherited callable table and distinguishes own,
non-enumerable methods from inherited ones. Function identity is shared with the
parent when the method is inherited; an override gets its own callable. JavaScript
constructor objects install only own methods and inherit the parent constructor.
The exception analysis includes static methods that escape as callables.

The Node oracle covers free extraction, bind with a derived constructor receiver,
identity, own-property reflection, enumeration, overrides, and a caught unbound
static this read. The initial oracle caught omitted static exception propagation:
its pending error had been ignored and printed 0. Static methods now participate
in callable throw propagation. All final commands passed:

```sh
go test ./internal/oracle -run 'TestMethodValuesTypeScript/static' -count=1 -timeout 10m > /tmp/method-values-static-typescript-fixed.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(method_values|class_static)' -count=1 -timeout 10m > /tmp/method-values-static-adamic-fixed.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-static-counts.log 2>&1
python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-static-mutants.log 2>&1
```

Outputs: TypeScript oracle ok .406s, Adamic and static regressions ok 1.065s,
counts refresh ok 26.720s. All seven mutants exit 1 with intended catchers:
the original five plus static-throw-hidden (Node stdout disagreement) and
static-own-hidden (Node stdout disagreement). Source is restored after every
mutant. The fixtures run native with ASan/UBSan, release native, and the
JavaScript backend against source Node, with successful-run leak checks.

## Optional extraction

Optional receiver extraction evaluates its receiver once and skips the complete
binding expression, including receiver-producing arguments, when absent.
Optional method signatures can read absent own fields as undefined; class
implementations still use their callable table. Both .a and temporary .ts
fixtures cover present and absent receivers, lazy bind arguments, missing
optional members, and a caught unbound this-reading optional extraction.

```sh
go test ./internal/oracle -run 'TestMethodValuesTypeScript/optional' -count=1 -timeout 10m > /tmp/method-values-optional-typescript-2.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/optional' -count=1 -timeout 10m > /tmp/method-values-optional-adamic-2.log 2>&1
go test ./internal/lower -run 'TestMethodValuesProof|TestMethodValuesSelectiveChecks|TestMethodBindCycleIsRefused' -count=1 > /tmp/method-values-optional-lower.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-optional-counts.log 2>&1
METHOD_VALUES_MUTANTS=optional-receiver-guard,optional-member-hidden python3 notes/method-values/ruling/run-mutants.py
```

TypeScript and Adamic oracles each passed .495s, lower passed .204s,
counts refresh passed 26.170s. Optional-receiver-guard exited 1 under sanitizers.
Optional-member-hidden exited 1 with a Node exit-code disagreement. The first
runner expectation incorrectly looked for stdout instead of the earlier exit-code
comparison; after correcting that expectation, the mutant was rerun and caught
(/tmp/method-values-optional-member-mutant.log). Source was restored.
Optional binding of an absent callable (m?.bind) remains a separate boundary.

## Rest extraction

Method callables pack a fresh rest array at their boundary, rather than treating
an array slot as one ordinary incoming argument. Class thunks and literal method
closures share the same runtime packer; JavaScript mirrors that convention.
Direct class method calls use that callable convention when the method has rest.
Spreads copy their slots at the source position and retain reference elements
before later arguments can mutate the source. No source array is reused as rest.

Fixtures cover fixed parameters plus rest, empty rest, class/static/literal
methods, direct calls, bind, callbacks, spreads, array independence, dynamic
strings, and later argument mutation. Temporary .ts inputs also check caught
this-reading rest methods. Generic rest methods, super rest calls, and spreads
filling fixed parameters remain explicit boundaries.

```sh
go test ./internal/oracle -run 'TestMethodValuesTypeScript/rest' -count=1 -timeout 10m > /tmp/method-values-rest-typescript-2.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/rest' -count=1 -timeout 10m > /tmp/method-values-rest-adamic-2.log 2>&1
go test ./internal/lower -run 'TestMethodValuesProof|TestMethodValuesSelectiveChecks|TestMethodBindCycleIsRefused|TestCensusRestMutableElements' -count=1 > /tmp/method-values-rest-lower.log 2>&1
go test ./internal/oracle -run TestMethodValuesTypeScript -count=1 -timeout 10m > /tmp/method-values-rest-regression.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-rest-counts.log 2>&1
METHOD_VALUES_MUTANTS=rest-reference-without-retain,rest-spread-without-owner python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-rest-mutants.log 2>&1
```

All passed: rest TypeScript .574s, rest Adamic .589s, lower .233s, all method-value
TypeScript cases 1.343s, counts refresh 26.918s. Both mutants exit 1 with ASan
heap-use-after-free on dynamic strings. Each source was restored.


## Receiver forms and intrinsic aliases

Receiver provenance is followed through locals, plain object fields, array slots,
returns, helper calls, and callable arguments. Bare this retains undefined in
TypeScript; typeof and equality can observe it. A later property read raises the
TypeError at that read. A plain write evaluates its RHS before the TypeError;
a compound write first attempts its property read. Optional scalar reads require
an undefined-compatible result representation. This-type aliases use their
checker constraint instead of losing the represented class. Bind validates nominal
receiver compatibility, including layout, rather than accepting an unrelated
structurally similar class.

Only extracted methods seed absent receiver provenance. The selective-check test
still confirms that an unextracted method's this reads stay unchecked. Helpers
receiving a possibly absent extracted receiver can require checks independently.
Slot names and callable parameter flows join conservatively. Unsupported consumers
of that receiver, including object reflection and union erasure without an
undefined adapter, remain NotYet. Refusing them is necessary: a native panic or
an object-tagged NULL would not reproduce Node's undefined behavior.

All 34 existing deterministic numeric Math operations now have ordinary, stable
callable identities, including imul, clz32, and fround. They can inhabit mutable
variables and object fields, return from functions, and serve numeric callbacks.
Variadic max, min, and hypot use the fresh rest convention. Every read of the same
intrinsic shares one global callable; mutable aliases read their actual local.
The old specialized adapters continue to handle supported typed call/apply uses.
The existing callable variance proof refuses rest signatures erased into a
narrower single-argument callback; no additional redundant check was retained.

Three new .a source fixtures cover receiver forms, their bound variants, and
intrinsic aliases. The this-reading receiver fixture is materialized as .ts only;
its bound variant and intrinsic fixture also enter the .a oracle table. In total,
the method-value suite has thirteen temporary .ts fixtures and seven .a fixtures.
Source Node decides their output, exit status, and stderr for sanitized native,
release native, and JavaScript backends. Successful sanitized runs check leaks.

Final commands and logs:

```sh
go test ./internal/lower -run 'TestMethodValuesProof|TestMethodValuesSelectiveChecks|TestMethodBindCycleIsRefused|TestMethodReceiverUnsupportedConsumers|TestMethodReceiverRepresentationRefusals' -count=1 > /tmp/method-values-final-lower-5.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(method_values|user_iterators_rest_tdz)' -count=1 -timeout 10m > /tmp/method-values-final-adamic-3.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestMethodValuesTypeScript -count=1 -timeout 10m > /tmp/method-values-final-typescript-4.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestMethodValuesTypeScript/intrinsics' -count=1 -timeout 10m > /tmp/method-values-intrinsics-expanded-typescript.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/intrinsics' -count=1 -timeout 10m > /tmp/method-values-intrinsics-expanded-adamic.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(method_coverage_(alias_order|math)|library_method_values)' -count=1 -timeout 10m > /tmp/method-values-intrinsic-regressions.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-final-counts-3.log 2>&1
python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-final-all-mutants.log 2>&1
METHOD_VALUES_MUTANTS=optional-scalar-erased,bound-nominal-erased python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-final-representation-mutants.log 2>&1
```

The full mutant invocation ran its then-current sixteen mutants. Four more were added afterward and run separately; receiver-consumer
mutants were also rerun after tightening collection and private-name boundaries. All twenty fail with the intended
catcher and restore exact source bytes. Final passing outputs: lower .234s,
uncached .a/iterator oracle 3.089s, uncached TypeScript oracle 3.044s, expanded
intrinsics TypeScript .599s and .a .945s, affected alias/Math regressions .771s,
and final counts refresh 26.506s. No whole-package test or full gate was run.

The initial counts refresh exposed a panic caused by treating a computed method
name such as Symbol.iterator as plain text. The metadata now skips computed
names; the existing iterator fixture passes and counts were refreshed again.
The optional scalar and nominal receiver mutants each turn a required refusal
into successful lowering, confirmed by their "got <nil>" failures.

Every mutant and its catcher:

| Mutant | Catcher |
| --- | --- |
| bound-extraction | Node output/exit disagreement for this-reading extraction |
| missing-this-check | UBSan/ASan on an undefined receiver |
| receiver-without-retain | ASan heap-use-after-free |
| stale-unbound-cache | ASan heap-use-after-free |
| bound-cycle-hidden | Required cycle refusal disappears |
| static-throw-hidden | Node stdout disagreement |
| static-own-hidden | Node stdout disagreement |
| optional-receiver-guard | UBSan/ASan |
| optional-member-hidden | Node exit-code disagreement |
| rest-reference-without-retain | ASan heap-use-after-free |
| rest-spread-without-owner | ASan heap-use-after-free |
| receiver-consumer-erased | Unsupported receiver refusal disappears |
| receiver-alias-unchecked | UBSan/ASan |
| receiver-write-unchecked | Node exit-code disagreement / sanitizer failure |
| intrinsic-identity-split | Node stdout disagreement |
| intrinsic-rest-dropped | Node stdout disagreement |
| optional-scalar-erased | Required result-representation refusal disappears |
| bound-nominal-erased | Required receiver-layout refusal disappears |
| intrinsic-alias-tdz-erased | Node output/exit disagreement when an alias is read before initialization |
| private-receiver-adapter-erased | Required private source-name/brand adapter refusal disappears |

## All 43 unresolved sites

The original pinned census remains 108 rows: 60 this-free implementation-family
sites, five direct this-reading bodies, and 43 unresolved runtime origins.
The lexical-this census controls still pass. audit-unresolved.cjs records every
one of the 43 rows in unresolved-audit.json, with declaration locations,
receiver initializer observations, and named candidates kept separate.

Six program-method reads have observed this-free local initial bodies:
getCommonSourceDirectory, getCompilerOptions, getSourceFile, getSourceFileByPath,
getSourceFiles, and getFileIncludeReasons. The other 37 rows depend on runtime
contracts for host methods, cancellation tokens, trackers, or writers. The six
local bodies do not establish all runtime targets either: the program object
escapes. No whole-program nonreplacement proof is claimed, so none of these
43 rows is added to the proven this-free count. Determining their actual targets
requires the complete calling program and its supplied implementations. A .ts
interface declaration or a same-named function cannot establish that proof.

Reproduction:

```sh
METHOD_VALUES_TYPESCRIPT=/workspace/scratch/method-values/node_modules/typescript node notes/method-values/ruling/census.cjs > /tmp/method-values-followup-census.log 2>&1
METHOD_VALUES_TYPESCRIPT=/workspace/scratch/method-values/node_modules/typescript node notes/method-values/ruling/audit-unresolved.cjs > /tmp/method-values-unresolved-audit-final.log 2>&1
```

## Notes and remaining boundaries

The final compiler build is logged in /tmp/method-values-followup-build-4.log.
All 32 unsupported examples were rechecked individually with adamic c; stdout
and diagnostics are under /tmp/method-values-followup-notes, with their summary
in /tmp/method-values-followup-notes-summary.log and committed outcomes in
notes-recheck.json. mutable_alias.a, math_variadic.a, and math_new_statics.a now
compile. The stale a-check refusal header was removed from mutable_alias.a;
the other two had no header. The remaining 29 still refuse or report NotYet.

Remaining explicit boundaries: optional binding of an absent callable (m?.bind),
binding an explicitly absent receiver, partial bind arguments, generic rest
methods, super rest calls, spreads filling fixed parameters, and receiver
consumers without an undefined adapter. This includes object collection callbacks
that may receive an absent extracted receiver, consuming escaped absent arrays,
maps, strings, or callables, and aliased private reads/writes lacking source-name
and brand adapters. Numeric collection callbacks remain supported. Non-Math intrinsic values continue to
use their limited specialized adapters; arbitrary escaping String/Number/Object
or Array methods require callable adapters that retain their receiver and
argument/result representations. Unsupported .ts intrinsic extraction is still
NotYet rather than Node-equivalent execution. The 43 audited runtime origins
remain unresolved; no full tsc compilation or end-to-end resolution is claimed.

No protected orchestration file was edited, no source was copied from cohere,
and no other worker branch was merged. The dependency for a stronger site count
is the actual host/tracker/writer/cancellation implementation and a whole-program
nonreplacement proof, not an unlanded worker branch.


Final review preserved Math alias initialization checks as catchable ReferenceErrors,
including alias calls. A Node fixture caught the initial raw native TDZ panic;
checked aliases now use an ordinary throwing helper. The new mutant removes its
undefined check and is caught by the oracle. Alias reads keep their local value
rather than replacing it with the canonical intrinsic before initialization.
Collection provenance now also covers reference locals in destructured loops and
map-value arrays; unsupported callback/consumer paths remain refused. Numeric
collection callbacks and array lengths keep their ordinary supported path.
Private aliased receivers remain refused rather than displaying qualified native
storage keys in JavaScript error messages. Its mutant turns that refusal into
successful lowering, reported as "got <nil>".

Additional final commands:

```sh
go test ./internal/lower -run 'TestMethodValuesProof|TestMethodValuesSelectiveChecks|TestMethodBindCycleIsRefused|TestMethodReceiverUnsupportedConsumers|TestMethodReceiverRepresentationRefusals' -count=1 > /tmp/method-values-final-lower-10.log 2>&1
go test ./internal/oracle -run TestMethodValuesTypeScript -count=1 -timeout 10m > /tmp/method-values-final-typescript-7.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values' -count=1 -timeout 10m > /tmp/method-values-final-adamic-6.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestMethodValuesTypeScript/intrinsics' -count=1 -timeout 10m > /tmp/method-values-intrinsic-tdz-typescript-3.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/intrinsics' -count=1 -timeout 10m > /tmp/method-values-intrinsic-tdz-adamic-3.log 2>&1
METHOD_VALUES_MUTANTS=intrinsic-alias-tdz-erased python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-intrinsic-tdz-mutant.log 2>&1
METHOD_VALUES_MUTANTS=receiver-consumer-erased,receiver-alias-unchecked,receiver-write-unchecked python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-final-receiver-mutants-2.log 2>&1
METHOD_VALUES_MUTANTS=receiver-consumer-erased python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-final-consumer-mutant-4.log 2>&1
METHOD_VALUES_MUTANTS=private-receiver-adapter-erased python3 notes/method-values/ruling/run-mutants.py > /tmp/method-values-private-adapter-mutant.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/method-values-final-counts-7.log 2>&1
```

Final lower output ok .327s, method-value TypeScript ok 1.144s, .a ok 1.511s,
uncached TDZ intrinsics TypeScript ok .610s and .a ok 1.756s. All additional
mutant invocations exit 1 with intended catchers and exact source restoration.
The complete twenty-mutant accounting combines the sixteen-mutant run, the two
representation mutants, the TDZ mutant, and the private adapter mutant.
An intermediate verification had used a nonexistent IR name ArrayLength;
compilation caught it, it was corrected to Length, and the final targeted suite,
consumer mutant, and counts refresh were rerun successfully.

Final counts refresh: ok 35.949s, /tmp/method-values-final-counts-7.log.
The final rebuilt compiler rechecked all 32 notes with the same three successes
and 29 refusals/NotYet outcomes. Source and mutation restoration passed git diff --check.

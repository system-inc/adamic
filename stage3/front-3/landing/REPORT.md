# Current-main landing

Merge exact origin/main b39d8146901f3cab86f4595af07ce95db0e875e4 into compiler/stage3-front-3, after green debugger integration 4b07fc712d1ce39f72b3b9b1f57a2282b849a15e. Automatic merge is conflict-free.

Keep front-3 namespace, enum-readiness, overload, branding, census, method-presence and debugger behavior. Add main's initializer-read-path, iterator receiver-origin, symbol-key-view and static-private refusals, with original Node observations and safe repair probes unchanged. Held item17, item18 and views-integration e555d67e are not ancestors of the incoming main pin.

The frozen-target reproducer was rerun against that detached exact main. Node prints caught/exit0; main rejects at checker TS2591 node:fs and TS2339 captureStackTrace. It does not print completed. Raw evidence remains under item18/.

Linux oracle counts and stage3 outcomes are regenerated; every existing status byte outside stage0 is compared with pre-merge front-3 HEAD. Package gates and four refusal mutation controls are recorded here when complete.

This integration runs the requested six packages and vets the repository. It does not claim whole-repository test execution, full TypeScript compiler acceptance or performance parity.

First full landing gate: lower fails TestParameterPropertiesSoundness/inherited_initializer, /early_default and /field_initializer with got nil, want refusal initialized. Incoming initializerReads checked only ast.KindPropertyDeclaration and missed parameter properties represented by ast.KindParameter. Small compiler repair: treat parameterProperty(member) as field storage in the same availability check. Existing parameter-property refusal tests pass and a removal mutant reproduces all three failures. The corrected full six-package gate is rerun rather than claiming the first run was green.

Flow corpus inventory repair: main's eight original class/iterator wrong-output probes are intentionally refused by TestClassWrongOutput103/106/107/108, which preserve exact Node observations and diagnostics. Add only those eight filenames to refusedNonNullFixture; do not ignore arbitrary lowering failures or remove the dedicated refusal/repair checks. This classification is included before the corrected full gate compiles its flow tests.

The first run's lower and flow failures are preserved in first-gate.log. Native614.220s, IR19.312s and fixtures73.836s passed that run; its remaining oracle was deliberately terminated after the known failures to free CPU for the corrected complete rerun. That interrupted oracle is not a pass. See first-interrupted.log.

Corrected complete gate is green: lower221.167s, native658.655s, IR50.299s, flow477.626s, oracle785.912s, fixtures112.429s. Vet, gofmt and whitespace checks are clean. Five valid refusal controls are caught. Every existing byte outside stage0 is unchanged from pre-merge HEAD; no Compiles downgrades.

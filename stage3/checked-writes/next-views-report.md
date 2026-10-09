Built short-circuit census repair, generic relation proofs, optional boolean writes, conditional allocation tracking and overload result checks for step 10, #63x2441.
Base 674b68b5; origin/main 54cbc125 is already merged through 830a46e1; the delivery SHA is reported with the push.
The final census is 471 checked, 18 proven, 26 refused and zero unmatched; Node oracles, sanitizers, local a-check and allocation counts pass.
Thirty-six mutants are caught, including eleven new contract mutants; deleting the original write check is caught by flags-misfit.ts.
Not covered: twenty-two waiting views contracts, four independent rule refusals, unrestricted overload ABIs, generic function values, or native execution of complete stock tsc.

The && walk follows the result-bearing right operand. Both spurious refusals now
prove. Four previously checked rows also prove after the same correction: they
were truth-tested left operands, not returned views. No runtime check is claimed
for those six corrected relations.

Producing signatures instantiate against receiving signatures through the existing
checker shim. A receiving generic binder remains rigid, rather than being replaced
by its constraint. Contextual generic method signatures supply binder identity for
their return expressions. Twelve of the sixteen generic rows prove: exact binder
or concrete signature identity, undefined admitted by the receiving union, and
three homomorphic mapped types that remove only readonly. The mapped proof checks
the same binder, the complete keyof domain, unchanged indexed values and unchanged
optionality; optional mapped transformations do not receive that proof. Concrete
generic calls have fitting/misfitting runtime witnesses. Generic function values
still have their existing separate lowering limit, so no backend execution of
those uninstantiated source functions is claimed.

Four generic rows retain independent refusals. Records 429 and 430 instantiate to
producers that cannot accept undefined admitted by the receiving parameter. Their
fix is to take the admitted input domain. Records 413 and 613 contain value-level
any handed to rigid T through erased callback/table relations. Their fix is to
retain the typed relation or use unknown and narrow it. These are not missing
field contracts; no-unsafe-assignment and method-signature-style remain enforced.

Optional boolean contracts validate presence and the finite true/false domain.
Native writes convert to the actual slot's plain or tagged representation and
retain that representation for later reads. The three new pairs cover true or
undefined, required true, and false or undefined. A fitting sequence writes
undefined and then true back into the same optional slot. The unproven Adamic
control refuses with its exact path and readonly/copy fix.

Conditional relations are judged against each original branch, even when the
checker reduces the whole expression to its wider branch. Three diagnostic pairs
pin file, start and length, and two bottom-array pairs pin number and string
views. Allocation contracts remain attached to each original object or array.
Never has no fitting element: the fitting witnesses perform an empty operation
on the bottom branch and insert through the same guarded writer into a broad
allocation. Node finishes after each misfitting store; Adamic stops before it is
stored, with exit 70 and the pinned field/domain message.

The createTempVariable relation validates every receiving overload against the
producer, including missing optional arguments. A narrower object-result overload
checks unproved fields at the resolved call. The call runs once, and its returned
allocation keeps its original contracts. Direct and function-value callback pairs
pin emitNode's required autoGenerate relation. Function values are admitted only
when their overloads share the implementation's incoming representations and object
result ABI. Existing scalar overload-value diagnostics remain. --explain-checks
lists the optional boolean and both overload-result checks for C, JavaScript and
sanitized builds. Adamic's unproven overload result control stays refused.

| Census stage | Checked | Proven | Refused | Unmatched | Total |
| --- | ---: | ---: | ---: | ---: | ---: |
| Base | 466 | 0 | 49 | 0 | 515 |
| Short-circuit walk | 462 | 6 | 47 | 0 | 515 |
| Initial signature instantiation | 462 | 15 | 38 | 0 | 515 |
| Mapped identity, booleans and branches | 470 | 18 | 27 | 0 | 515 |
| Overloaded callback relation | 471 | 18 | 26 | 0 | 515 |

These are measured relation decisions on checker-rejected adapted TypeScript, with
all 603 pinned source hashes verified. The latent loader and Lower cannot produce
backend IR. next-views-results.json retains each requested slice's records, every
remaining source site, its complete first refusal and its waits-on entry. Twelve
of the sixteen generic rows prove; none is declared checked merely because a
constraint happens to match one possible type argument.

| Remaining reason | Sites | Waits on |
| --- | ---: | --- |
| Boxed-union fields and elements | 14 | Views V2 and V5: tagged slot contracts and safe conversion |
| Tuple slots | 4 | Views V3: positional and callable element contracts |
| Recursive references | 2 | Views V5/V6: recursive allocation contracts |
| Set element | 1 | Views V2 and V5: nullable element insertion contracts |
| Constrained generic array | 1 | Views V3: concrete member and element contracts |
| Method parameter variance | 2 | Producer signatures accepting undefined; independent method-signature-style rule |
| Erased any-to-generic callback relation | 2 | Typed callback/table relations replacing any; independent no-unsafe-assignment rule |

The fixture split is 105 checked .ts witnesses, eleven refused .a controls, and
four proven .a controls. The twenty-two new TypeScript witnesses are held to Node
in both backends, sanitized and release native. Fitting executions and successful
mutants also pass native leak checks. Four new refusal sidecars pin the complete
path, reason and fix. proven-generic.a has no inserted write checks. counts.md adds
exactly twenty-three rows: twenty-two TypeScript witnesses and that proven control.

| New mutant | Fixture catching the missing exit-70 check |
| --- | --- |
| Remove optional true literal set | optional-boolean-misfit.ts |
| Remove optional false literal set | optional-false-misfit.ts |
| Remove required boolean presence | required-boolean-misfit.ts |
| Remove branch file presence | branch-file-misfit.ts |
| Remove branch start presence | branch-start-misfit.ts |
| Remove branch length presence | branch-length-misfit.ts |
| Remove bottom number allocation contract | branch-bottom-number-misfit.ts |
| Remove bottom string allocation contract | branch-bottom-string-misfit.ts |
| Remove concrete generic slot literal set | generic-site-misfit.ts |
| Remove direct overload result check | overload-result-misfit.ts |
| Remove callback overload result check | overload-callback-misfit.ts |

Every new mutant matches Node with exit 0 in both backends, so its pinned exit-70
oracle fails. The bottom string mutant also changes the now-unrestricted empty
allocation's element representation to String, allowing the admitted reference
store to run with normal reference ownership. It therefore fails only the
contract oracle, not clang or LeakSanitizer. The earlier twenty-five mutants were
rerun in the focused regression suite, including removal of the write check and
NodeFlags.Synthesized set in flags-misfit.ts.

The focused regression suite passes for oracle, driver, lowering and native; its
exact commands and output log paths are in next-views-commands.txt. The required
TestCountsAreRecorded update passes. Local a-check exits 0 with no failures,
using the same gate policy pinned at 3e339bbf. A final boolean conversion check
and exact refusal pin check pass after the focused suite. No whole package test
or full gate was run.

The stock Node tsc controls are rerun from 3255eb1e with TypeScript 6.0.3. All 301
CLI inputs match their goldens, 61 flags writes pass, and the parent write executes
zero times. Both original small programs exit 70 before storage. The native and
JavaScript flags/parent reductions also pass the focused oracle. This is explicit
pre-store Node instrumentation of stock tsc; no native whole-compiler run is claimed.
next-views-latent-results.json records the observed commands, package hashes and
write counts.

Setup used GOPROXY=https://proxy.golang.org|direct, then cloud/setup.sh, then the
printed /workspace/adamic-tools/env.sh. Timing lines: Go 0.045s, Node 0.045s,
submodules 0.111s, markdown 0.131s, clang 0.257s, build 70.192s, total 70.552s.
nproc reports 5, with a four-CPU quota. Setup had no failure. Current origin/main
54cbc125, including the requested e96b7d43, is already an ancestor of the delivery
branch. No cohere source was copied. The only delivery push is to
codex/checked-wider-writes.

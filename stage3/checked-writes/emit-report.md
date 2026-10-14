Built checks for all 38 pinned EmitNode.autoGenerate sites toward step 10, #63x2441.
Base f84d5e45; merged origin/main 54cbc125, including e96b7d43, in merge commit 830a46e1.
Targeted Node oracles, local a-check, allocation counts and the 515-site census pass.
Twenty-five mutants are caught; deleting the EmitNode store check is caught by emit-node-misfit.ts.
Not covered: 49 refused relations, native execution of complete stock tsc, and unrestricted generic/overloaded callback contracts.

Plain object intersections now use the representations ordinary object fields use.
EmitNode & { autoGenerate: AutoGenerateInfo } therefore keeps its real allocation
contract. Replacing emitNode through a wider view validates that contract. Replacing
its nested autoGenerate validates that slot's required reference contract. Both
checks precede storage and stop with exit 70, the field expression, declared type
and incoming value or kind. Compatible callbacks carry field relations backward
through their parameters and forward through their results. The rule covers one
non-generic signature. Unproven parameter variance and overloads stay refused.

Five fitting/misfitting pairs cover EmitNode replacement, nested autoGenerate,
a callback receiving GeneratedIdentifier, SynthesizedComment positions, and a
callback returning a resolved filename through its wider result type. The latter
two also close pinned sites 414 and 521. Fitting cases match Node in JavaScript,
sanitized native and release native, with native leak checks. Misfits stop with the
pinned message while Node stores silently. --explain-checks includes the EmitNode
store. emit-refused.a pins its path and readonly/copy fix. proven-emit.a stores a
statically fitting AutoGenerateInfo with zero write checks. The unit now has 83
checked TypeScript witnesses, seven refused Adamic controls and three proven
Adamic controls; spread-override.ts remains a separate refusal control.

| Census stage | Checked | Proven | Refused | Unmatched | Total |
| --- | ---: | ---: | ---: | ---: | ---: |
| After main merge | 426 | 0 | 78 | 11 | 515 |
| Plain intersection fields | 462 | 0 | 42 | 11 | 515 |
| Compatible callbacks | 466 | 0 | 38 | 11 | 515 |
| Resolved receiving expressions | 466 | 0 | 49 | 0 | 515 |

The 38 EmitNode sites are all checked. Forty of the previous 78 refusals are now
checked, including the two adjacent callback sites. The eleven unmatched records
now have a unique receiving expression or shorthand at their pinned location.
Ten refuse an uninstantiated writable generic relation; record 482 refuses an
overloaded callback result. census.json records their actual expression/type and
complete first refusal. The printed nested generic type was not the enclosing
expression's type, which caused the old matcher to miss them. Where the normal
relation walk skips a generic function pending instantiation, this census explicitly
measures its uninstantiated receiving relation and retains the refusal. These are
relation decisions on checker-rejected source, with all 603 source hashes verified;
the latent loader and Lower cannot produce backend IR. No unresolved record was
promoted to a static proof or an executable whole-compiler claim.

The prior 78 refusals group as follows. emit-results.json contains every site and
its complete first reason, the new decision and all eleven recovered records.
The contract probe log separately records the first unsupported field component;
its source is retained as scratch-only evidence.

| Prior refusal reason | Sites | Now checked | Now refused |
| --- | ---: | ---: | ---: |
| EmitNode.autoGenerate | 38 | 38 | 0 |
| Boxed union fields or elements | 14 | 0 | 14 |
| Uninstantiated writable generic relation | 6 | 0 | 6 |
| Tuple slots, including callable elements | 4 | 0 | 4 |
| Inferred diagnostic branch loses its narrow relation | 3 | 0 | 3 |
| Optional boolean field | 3 | 0 | 3 |
| Array versus Map field representation | 2 | 0 | 2 |
| Compatible callback fields | 2 | 2 | 0 |
| Inferred bottom-array branch loses its narrow relation | 2 | 0 | 2 |
| Recursive structural reference | 2 | 0 | 2 |
| Set element | 1 | 0 | 1 |
| Constrained generic array union | 1 | 0 | 1 |

The largest remaining group is 16 uninstantiated writable generic relations, then
14 boxed-union contracts. Further generic relation work must instantiate the
producing and receiving signatures before proving their concrete writable slots.
Further step 10 work supplies boxed-union contracts and checked representation
conversion, packed boolean-or-undefined contracts, positional tuple/callable
contracts, Set insertion contracts, cycle-safe recursive references and preserved
per-branch allocation relations. Existing union, tuple, generic and Set language
support provides their representation but does not establish these slot contracts.
The remaining overloaded callback needs every applicable signature proved. The two
array-to-Map field views need a proven copy or representation-safe reads and writes;
a write check alone cannot make their incompatible member layouts safe. No later
numbered roadmap dependency for these gaps is pinned in this repository; the ledger
names the required capability and keeps the refusal until it exists. Adamic's
unproven wider writes keep their refusal under the source ruling.

| Mutant | Fixture catching the lost exit-70 check |
| --- | --- |
| Delete scalar store check | flags-misfit.ts |
| Delete NodeFlags.Synthesized literal set | flags-misfit.ts |
| Delete alias store check | diagnostic-alias-misfit.ts |
| Delete reference allocation proof | diagnostic-proof-misfit.ts |
| Delete nested reference contract | diagnostic-nested-misfit.ts |
| Delete never-element contract | never-number-misfit.ts |
| Delete FlowNode union contract | flow-node-misfit.ts |
| Delete undefined-object contract | flow-undefined-misfit.ts |
| Delete undefined-array contract | flow-array-misfit.ts |
| Delete numeric element contract | container-number-misfit.ts |
| Delete string element contract | container-string-misfit.ts |
| Delete Map value contract | container-map-misfit.ts |
| Delete splice element contract | container-splice-misfit.ts |
| Delete object element contract | container-object-misfit.ts |
| Delete nested-array contract | container-nested-misfit.ts |
| Delete boolean element contract | container-boolean-misfit.ts |
| Delete fill element contract | container-fill-misfit.ts |
| Delete Map object-value contract | container-map-object-misfit.ts |
| Delete required nested numeric presence | container-alias-misfit.ts |
| Delete EmitNode store check | emit-node-misfit.ts |
| Delete EmitNode allocation proof | emit-node-misfit.ts |
| Delete required autoGenerate presence | emit-auto-misfit.ts |
| Delete callback parameter allocation proof | emit-callback-misfit.ts |
| Delete comment-position literal set | emit-comment-misfit.ts |
| Delete callback result filename presence | emit-resolution-misfit.ts |

Every mutant executes successfully and matches Node in both backends, so its
pinned exit-70 oracle fails. Sanitized native mutants pass leak checks. The new
six mutants pass in emit-mutants-final.log; the affected suite retains the earlier
nineteen mutants. emit-final-tests.log records passing oracle 163.660s, driver
4.140s, lower 1.399s and native 1.593s. The required count refresh passes in
166.549s and adds exactly eleven rows. Local a-check exits 0 with no failures;
its gate policy is pinned at 3e339bbf, since origin/main does not contain the gate
tools' cloud/fast-gate/run.py. Exact commands are in emit-commands.txt.

The stock Node tsc controls from 3255eb1e are rerun with TypeScript 6.0.3 and the
same explicit pre-store contract instrumentation. All 301 CLI inputs match their
goldens; 61 flags stores pass and the parent store executes zero times. Both original
small programs stop with exit 70 before storing. The flags and parent reductions
also pass the native and JavaScript Node-held oracles. This is Node instrumentation
of stock tsc; no native stock compiler run is claimed. emit-latent-results.json
records commands, package hashes, exit codes and observed counts.

Setup used GOPROXY=https://proxy.golang.org|direct and sourced the printed
/workspace/adamic-tools/env.sh. Timing lines: Node 0.067s, Go 0.086s, markdown
0.215s, submodules 0.229s, clang 0.622s, build 72.107s, total 72.356s. nproc is 5,
with a four-CPU cgroup quota. No setup failure occurred. No whole-package test,
full gate, Windows run or cohere source copy was performed. The only delivery push
is to codex/checked-wider-writes.

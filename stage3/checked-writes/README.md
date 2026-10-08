Checked wider writes preserve field declarations in actual object shapes. TypeScript input may expose a wider writable view; each affected named store checks the incoming value before changing ownership or storage. A rejected value exits 70 with the expression, actual declaration and value. `--explain-checks` lists the stores. Adamic input keeps its existing invariant-view refusals.

Contracts preserve scalar representation and finite number, enum-member, string and boolean domains. Broad numeric enums retain the existing open-number doctrine. Reference contracts now retain the complete plain allocation declaration and directional checker proofs for incoming values and allocation fields. A reference arriving through a broad view can use its actual allocation declaration or field declarations to fit a narrower destination. Compatible callable and container-bearing payloads need no erased type guess: their complete allocation proof is recorded. Map key variance, Set elements, tuple and class element views, unreifiable callable variance, general representation conversions and explicit replacement of fields in checked programs' spreads remain refused. Required numeric slots accept present packed numbers and reject undefined; NaN and negative zero retain their values. Unproven reference payloads fail closed. Plain spread copies preserve their source contracts. Shapes with identical names but different declarations remain distinct. Programs without checked writes emit no contract tables.

Runtime witnesses are `.ts` files and the oracle loads those files directly. Six `.a` refusal controls pin the source path, unsafe view and readonly or copy fix. `proven-number.a` and `proven-containers.a` hold exactly typed writes that compile without inserted write checks. This directory is excluded from ordinary Cohere source lint because its negative witnesses intentionally exercise unsound views. Fitting inputs match Node in JavaScript, sanitized native and release native. Misfits pin exit 70 and the complete diagnostic; Node silently stores those values. Successful sanitized native executions are also checked for leaks.

The reductions cover scalar literal unions, enum flags, boolean literals, diagnostic file presence, compound writes, nullable numbers, TextRange positions, receiver evaluation order and cloned parent presence. The diagnostic group adds six fitting/misfitting pairs: rich SourceFile replacement, DiagnosticRelatedInformation numeric presence, nested allocation domains, nested field writes, full allocation proofs containing callable and container fields, and generic alias lifetime. The former reference-structure refusal witness now stops at the write. Final allocation contracts protect earlier named writers too, so a broad alias cannot invalidate a narrower referenced object after a fitting store. These additional checks are listed by --explain-checks. The generic flags reduction performs 61 broad in-type writes. The parent reductions check preservation through a clone and the undefined counterexample. These are reduced compiler functions, not a native execution of the complete stock TypeScript compiler or its input suite. The stock Node controls are remeasured by `latent-controls.py`: all 301 CLI inputs match their golden outputs, 61 flags stores pass the pre-store check and the parent store executes zero times. The original small programs from `3255eb1e` each stop with exit 70. This is Node instrumentation of the stock stores, with explicit erased-domain registration for the narrow flags witness; the complete stock compiler is not compiled natively.

Six mutants run valid Node behavior in both backends. Removing the original store check or enum literal set produces exit 0 with `after 0`. Removing the allocation proof requirement, deleting the nested allocation contract, or removing the alias write check produces exit 0 with `0`. Dropping the never-element allocation contract admits the numeric insertion and prints `1`. Each pinned exit-70 oracle catches its mutant.

`census.json` records every pinned site's relation decision and first refusal. The census verifies all 603 source hashes from `9b8ebd77`. It uses the existing scratch latent overlay and asserts that ordinary Load and Lower cannot produce backend IR. Expression previews in the audit are truncated, so matching uses an exact location and checker type plus a source-text prefix. This measures relation admission on checker-rejected source; it does not claim complete compiler lowering or executable backend coverage.

Reproduce the adapted tree with `stage3/apply.sh` at `9b8ebd77`, then run `bash stage3/checked-writes/census.sh ADAPTED_TREE PINNED_RESULTS_JSON LOG_PATH` with the Adamic tools environment sourced. The delivery branch contains no commits from that evidence branch.

After the diagnostic group, 227 of the 515 class-d sites admitted checked field relations, 263 retained a refusal, and 25 had no current relation refusal and no runtime admission. That group added 200 checked relations: 146 diagnostics and 54 other. All 84 DiagnosticWithLocation-to-Diagnostic sites and all 58 DiagnosticWithLocation-to-DiagnosticRelatedInformation sites now admit checks. The 263 refusals include 166 other, 15 diagnostics and 82 shared-never cases. Their complete first reasons are in the JSON. Class-a and class-b rows are included for comparison; the unit does not claim to close all their remaining compiler limitations.

The next largest exact first reason is 19 FlowNode-to-FlowNode-or-undefined sites, where a wider node field can hold BindingElement, Expression or VariableDeclaration but the actual slot reads BinaryExpression or CallExpression. These union-bearing structural contracts remain refused. The shared-never group below closes all 82 of that family's relation refusals. A never-element container has no fitting element; its .a refusal control preserves the stronger Adamic source promise. Class allocation identities and unrestricted recursive or union-bearing schema fallback are not provided by this group. This delivers the diagnostic reference group toward roadmap step 10, task #63x2441.

Shared-never group: the allocation records that its element type is never. Push,
index assignment and splice check that allocation before storing. Runtime insertion
helpers also fail closed for bottom allocations. The receiver stays alive while
arguments are evaluated. Number, string and object views have empty-operation
controls; no value inhabits never, so a fitting element-write witness is impossible.
A separate broad-allocation control shows the guarded writer accepts a number slot.
Removing the allocation contract admits the numeric insertion and reproduces Node.

All 82 shared-never sites now have checked relation admission, including the nullable
Symbol array view. The class-d split is 309 checked relations, 181 refusals, and 25
sites without a current relation refusal. The next largest exact first reason is
19 FlowNode structural-union sites. never-group-results.json records every remaining
first reason and its sites. This census measures relations on checker-rejected
adapted sources; it does not claim usable backend IR for the whole compiler.

Fixture extension repair: 43 checked TypeScript witnesses, three refused Adamic
witnesses and one proven Adamic witness. spread-override.ts is a separate TypeScript
refusal control. The compiler area revision f0c6e6fc is merged. Runtime-check mutants
continue to use the original TypeScript files; dropping the store check in
flags-misfit.ts admits the misfit and reproduces Node's after-0 output.


FlowNode and direct-container group: union relations now retain the field contract
of the member allocated at run time. An undefined-only field accepts undefined
and rejects a present object or array. Arrays retain their element declaration at
allocation; Maps retain their value declaration. Index, push, splice, fill and
Map.set validate before changing ownership or storage. Bulk writes validate every
incoming item. Earlier writers and nested aliases keep the actual field checks.
Fresh array producers record fresh contracts, and checked programs avoid reusing
an old array while constructing a new allocation with a different declaration.
Map keys must stay invariant. Tuple and class element views remain refused.

The unit has 73 checked TypeScript runtime witnesses, six refused Adamic controls
and two proven Adamic controls. Fitting cases match Node in both backends;
misfits pin exit 70 and the full diagnostic. Nineteen mutants are caught,
including thirteen new contract mutants. See flow-container-report.md for the
commands, tsc controls, per-group counts and limits.

The final 515-site census records 426 checked relations, zero proven relations,
78 compiler refusals and 11 unresolved source matches. The unresolved matches
remain outside coverage; they are not static proofs or observed compiler refusals.
The largest remaining contract family has 38 EmitNode.autoGenerate sites.
flow-container-results.json records every remaining first reason and original
census reason. Earlier numbers above are historical measurements with the earlier
matcher; the new comparison uses the same corrected matcher for all groups.

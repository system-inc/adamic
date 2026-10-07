# Taste fixtures and stress coverage

These 24 programs retain TypeScript 6.0.3 functions, arrows, export statements,
or selected scanner/parser statements, plus explicitly marked receiver/index and
finally stress drivers. A 25-site primary sample apportions coverage by the source
counts; additional cases stress the low-frequency forms. No compiler was changed.
The remote taste tip was fetched and independently checked with `git ls-remote`;
its newest commit remains `aa896b5` for this follow-up.

## Pins and discovery

- Main: `ef3d907ecdc4c771b016f7d9c52372def057a340`.
- `codex/taste-not-soundness`: `aa896b5d5ccc82210184fd01b8fe4d0ce0730a50`.
- Upstream v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Stock npm `typescript@6.0.3`; Node v24.19.0; Go go1.27.1; clang 20.1.8.

Read the complete census REPORT and stage3-base README before extraction.
The census `sites.json`, `files.json`, and `rankings.json` were read as complete
JSON documents. `inventory.cjs` uses the stock compiler API to walk the pinned
source and select real syntax; it uses no regex for site discovery. It verifies
every original compiler source SHA256 against the census before scanning.
The non-boolean typed count comes from the census, not a fresh checker census.
Complete functions were initially extracted from compiler API nodes. CRLF was
normalized to LF. Exceptions made to isolate unrelated representation blockers
are declared below; the primary sample identifies exact original source sites. Dependencies
are small declared shapes or explicit stubs, not a claim to implement the omitted
compiler subsystem. The drivers include normal tsc inputs and boundary probes.

Census reasons answered here: `non-boolean control condition` (6,697 sites,
6,635 lines, 62 files), `||=` (110 sites, 109 lines), `the comma operator`
(47 sites, 46 lines), `a label` (10 sites/lines), `the void operator` (15),
and `an ExportDeclaration` (77).
The full AST walk also finds 173 `??=` and one `&&=`: the census's 110 is
specifically `||=`, not the entire logical assignment family. The nullish-assignment
fixtures cover additional usage of that family, with the census family heading
in their provenance comments. Likewise operand-valued logical operations and
negation are supplemental to the typed control-condition count.
The walk counts 75 `export *` and **2 named export lists**, both in
`_namespaces/ts.ts`. A named-export success does not establish star-barrel support.
The full recount and source selection are reproducible with the commands below.

## Proportional primary sample

`sample.cjs` uses the stock compiler API and exact census start/end offsets to
identify the original conditions. Operators, labels, void expressions and exports
are selected by their AST kinds. `sample.json` records all 25 distinct source
sites and the runnable fixture containing each; supporting probes do not inflate
these counts. Generated sources are excluded.

Allocation: reserve one site per form, then distribute the remaining 19 of the
25 slots by Hamilton largest remainders against the 7,130 sites in these forms.
This gives 19 conditions, 2 assignments and one of each smaller form. The
coverage floor deliberately oversamples rare forms; this is proportional
allocation with a minimum coverage constraint, not an unconstrained random draw.
Selection within each form is purposive (strings, objects, flags, array callbacks,
and all three loop conditions), not a statistical estimate of compiler behavior.
The assignment denominator includes the API's 173 nullish assignments, as well as
110 OR assignments and one AND assignment; the original 110 census heading alone
would undercount the family requested here.

| Form | Pinned count | Corpus share | Primary sites | Sample share | Supporting coverage |
| --- | ---: | ---: | ---: | ---: | --- |
| Non-boolean conditions | 6,697 | 93.927% | 19 | 76% | operand values, negations and mixed types |
| Logical assignment family | 284 | 3.983% | 2 | 8% | all four AND/nullish field/index combinations, exact scanner AND assignment |
| Comma | 47 | 0.659% | 1 | 4% | literal cache and scanner return |
| Labels | 10 | 0.140% | 1 | 4% | nested continues; four jumps through finally |
| Void | 15 | 0.210% | 1 | 4% | callback returns undefined |
| Export declarations | 77 | 1.080% | 1 | 4% | named list and star barrel |
| Total | 7,130 | 100% | 25 | 100% | 24 runnable programs |

The 19 conditions comprise five in 03/04/07/08, four in 21 (while, for,
indexed-access entry test and do), and ten in 24. Primary assignments are the
actual binder OR assignment and checker indexed nullish assignment. Every
primary site is attempted on both compilers, including sites whose enclosing
program is blocked. The sample table is coverage, not a claim that every site
runs natively.

## Observed outcomes

Every entry was run using the unchanged `oracle/node.mjs`, with no enum or
namespace runner changes. Every entry was then compiled using `go run ./cmd/adamic
build` on main and on the taste checkout. Every successful binary was executed.
All 24 Node runs exit 0 with empty stderr. Main compiles none. Taste compiles
twelve, and **all twelve match stdout, stderr, and exit exactly**. No silent
miscompile was observed among binaries that compiled; no claim is made about
programs blocked before execution.

`status.json` has exactly the requested schema and records main; `evidence.json`
records both builds and every executed native binary. Nonzero go-run diagnostics
retain their exact stdout/stderr, including `exit status 1`. `Checker` means a
checker error, `NotYet` follows the compiler's actual “can't lower ... yet”
message, and `Refused` follows its actual “Adamic 0.1 refuses ...” message.

| Fixture | Real form / driver coverage | Main | Taste |
| --- | --- | --- | --- |
| `01_diagnostic_code.a` | numeric operand values including both zeros and NaN | NotYet | Compiles (Node match) |
| `02_diagnostic_message.a` | string/object union operand values | NotYet | NotYet |
| `03_diagnostic_file.a` | object/undefined ternary | Refused | Compiles (Node match) |
| `04_jsx_runtime.a` | empty string/undefined ternary | Refused | Compiles (Node match) |
| `05_this_type.a` | double negation of flags and boolean | NotYet | Compiles (Node match) |
| `06_localized_message.a` | localized dictionary operand values | Refused | Refused |
| `07_compact.a` | mixed array/object truthiness and nullish assignment | NotYet | NotYet |
| `08_for_each.a` | callback result truthiness | Checker | Checker |
| `09_literal_cache.a` | literal cache comma expression | Refused | NotYet |
| `10_global_import_meta.a` | lazy OR assignment cache | Refused | Compiles (Node match) |
| `11_build_info_pending.a` | field nullish assignment | NotYet | NotYet |
| `12_symbol_links.a` | indexed nullish assignment | Refused | NotYet |
| `13_void_callback.a` | void callback | Refused | NotYet |
| `14_relative_complement.a` | nested labeled continue | Refused | NotYet |
| `15_barrel.a` | star barrel | Refused | Refused |
| `16_scan_exclamation.a` | scanner comma return | Refused | NotYet |
| `17_binder_flow.a` | binder OR assignment | Refused | Compiles (Node match) |
| `18_named_export.a` | named export list | Refused | Compiles (Node match) |
| `19_deprecated_flags.a` | numeric double negation | NotYet | Compiles (Node match) |
| `20_jsdoc_terminate.a` | JSDoc labeled break | Refused | NotYet |
| `21_truthy_loops.a` | non-boolean while, for and do | Refused | Compiles (Node match) |
| `22_assignment_once.a` | AND/nullish field and index assignments with counters | Refused | Compiles (Node match) |
| `23_labels_finally.a` | inner/outer break and continue through finally | Refused | Compiles (Node match) |
| `24_proportional_conditions.a` | ten original sampled string/object conditions | Refused | Compiles (Node match) |

Notable blockers: string/object/undefined fields in 02, optional boolean fields
in 11, unrestricted string index signatures in 06, generic array lowering
in 07, checked generic indexed reads in 08, inferred `any` in 09, an array
initialized only with undefined in 12, and a function returning undefined in 13.
Fixture 16 reaches an unsupported compound assignment used as an expression;
fixture 20 reaches a switch case expressed as an enum-shaped property. The latter
uses a supporting constant object because this unit does not implement enums.
Fixture 15 remains refused for `export *` on taste.

The numbered files are the runnable entries; `support/` is dependency scaffolding,
not additional standalone observations. 15 redirects the original core barrel
specifier to the minimal core support file. 18 preserves `export { performance };`
in a support barrel with an opaque performance-shaped value, and its entry imports
that exported binding. This tests export syntax/binding, not native performance.
16 retains `scan`'s signature and its exact exclamation-case statements, removing
the unrelated token cases and outer dispatch. 20 retains `parseJSDocType`'s
signature, exact `terminate` loop, scanner-restoration call, and return; other
parsing branches are omitted, with token/factory/scanner support supplied by the
driver. 17 supplies deterministic binder collaborators while preserving the
complete real binder function. The new supporting adaptations are explicit:
05's local Type model now makes isThisType a required boolean, and its absent
case is represented by false, preserving the original double-negation output.
The actual function and expression remain unchanged. 14 specializes T to
number-or-undefined to accommodate sound indexed reads, allows undefined in its
comparer, and makes its initial array-presence guard explicit. Its two loops,
switch and labeled continues are retained. This removes the checker mask; its
label mutant now exposes a separate enum-shaped-case NotYet.
21 preserves complete getAncestor/findActiveLabel functions and the indexed-access
branch of getRecursionIdentity. The latter's return declaration is narrowed from
object to its concrete Type model; its original if/do statements are unchanged.
24 retains ten complete functions and supplies deterministic collaborators;
Debug.fail's supporting declaration is void and its implementation still throws.

22 is a supplemental counter driver, not a claim that tsc writes those exact
side-effectful receivers. It also retains the real scanner's boolean AND-assignment
statements with unrelated token branches cut. 23 is supplemental finally stress
using tsc's two-loop label topology; no original tsc label-through-finally site is
claimed. Each jump prints its coordinates, then finally prints the same pair.
Inner continue, outer continue, inner break and outer break all execute cleanup.

The first ten lines from 22 each start `1:1:1:` (field receiver, array receiver,
index), followed by RHS count and resulting field/slot values. AND short-circuits
undefined, 0, -0 and NaN; nullish assignment evaluates RHS only for undefined.
23 prints four `try`/`finally` pairs and the expected outer-tail/done markers.
`audit.py` asserts those exact outputs, rather than treating compilation as proof.


## Rewrite mutants and what caught them

These are explicit source rewrites on main, not compiler edits. `mutants.json`
retains every result and the full diagnostic. All fifteen rewrites preserve the
original Node stdout, stderr, and exit. All fifteen flip the main outcome, proving
those fixtures reach the corresponding form. Eleven rewritten programs compile and their
native observations match Node. A flip to NotYet establishes removal of the form's
refusal and exposure of another blocker; it does not establish native correctness.

| Rewrite / fixture | Main before and after | Evidence |
| --- | --- | --- |
| `operand` / `01_diagnostic_code.a` | NotYet to Compiles | Node unchanged and native matches |
| `object_condition` / `03_diagnostic_file.a` | Refused to Compiles | Node unchanged and native matches |
| `string_condition` / `04_jsx_runtime.a` | Refused to Compiles | Node unchanged and native matches |
| `optional_double_not` / `05_this_type.a` | NotYet to Compiles | Node unchanged and native matches |
| `logical_assignment` / `17_binder_flow.a` | Refused to Compiles | Node unchanged and native matches |
| `comma` / `16_scan_exclamation.a` | Refused to NotYet | Node unchanged; further blocker exposed |
| `void` / `13_void_callback.a` | Refused to NotYet | Node unchanged; further blocker exposed |
| `label` / `14_relative_complement.a` | Refused to NotYet | Node unchanged; further blocker exposed |
| `double_not` / `19_deprecated_flags.a` | NotYet to Compiles | Node unchanged and native matches |
| `label_break` / `20_jsdoc_terminate.a` | Refused to NotYet | Node unchanged; further blocker exposed |
| `named_export` / `18_named_export.a` | Refused to Compiles | Node unchanged and native matches |
| `assignment_once` / `22_assignment_once.a` | Refused to Compiles | Node unchanged and native matches |
| `finally_jumps` / `23_labels_finally.a` | Refused to Compiles | Node unchanged and native matches |
| `truthy_loops` / `21_truthy_loops.a` | Refused to Compiles | Node unchanged and native matches |
| `star_export` / `15_barrel.a` | Refused to Compiles | Node unchanged and native matches |

The object and string rewrites make truthiness explicit. The numeric `||`
rewrite handles absence, both zeros, and NaN explicitly. Numeric `!!` becomes a
nonzero bitmask comparison. `||=` becomes its explicit boolean if. The comma
rewrite sequences the position update before returning the token assignment.
The void rewrite becomes a block that discards push's result. The labeled-break
rewrite restores scanner state and returns directly. The nested-continue rewrite
uses an explicit outer-iteration flag; it now flips Refused to NotYet, preserving
Node. The formerly masked optional double negation now flips NotYet to Compiles. Star export removal leaves the driver dependency import; named export
moves export to the supporting declaration. These export rewrites prove that the
barrel statement reaches main's refusal, not full star/live-binding semantics.

Four additional deliberate stress mutants were compiled with taste: reevaluate
the field receiver on the RHS, reevaluate the index on the RHS, omit finally's
output, and retarget the outer continue to the inner loop. All finish normally
and differ from the original source Node reference. Thus the output check catches
the intended semantic failure, rather than a checker/compiler rejection.
`stress-mutants.json` retains commands and native outputs. These intentional
source changes are checked against the unmodified program's reference output;
they are distinct from the semantics-preserving main rewrites.

`audit.py` separately proves its byte comparison can fail by deleting one stdout
byte from a copied successful observation, and its schema assertion can fail by
adding an extra status key. Both artifact mutants are caught. They do not change
saved evidence. Exact caught messages are in `logs/audit.log`.

## Reproduction and validation

Obtain the three census JSON files from `origin/codex/tsc-census` in a scratch
CENSUS_DATA directory, clone the pinned upstream tree, and install stock
`typescript@6.0.3` in a scratch API directory. Create a detached taste worktree;
initialize its pinned cohere dependencies. This run shared the identical pinned
cohere directories with symlinks in the scratch worktree, without editing them.

```sh
bash cloud/setup.sh > /tmp/taste-setup.log 2>&1
source /workspace/adamic-tools/env.sh
node stage3/fixtures/taste/inventory.cjs /tmp/taste-typescript /tmp/taste-api/node_modules/typescript/lib/typescript.js /tmp/taste-census /tmp/taste-inventory-followup.json > stage3/fixtures/taste/logs/inventory.log 2>&1
node stage3/fixtures/taste/sample.cjs /tmp/taste-typescript /tmp/taste-api/node_modules/typescript/lib/typescript.js /tmp/taste-census stage3/fixtures/taste/sample.json > stage3/fixtures/taste/logs/sample.log 2>&1
python3 stage3/fixtures/taste/observe.py --main /workspace/adamic --taste /tmp/taste-feature --logs /tmp/taste-followup3 > /tmp/taste-followup3.log 2>&1
python3 stage3/fixtures/taste/mutants.py --main /workspace/adamic --output /tmp/taste-followup-mutants-final > /tmp/taste-followup-mutants-final.log 2>&1
python3 stage3/fixtures/taste/stress_mutants.py --taste /tmp/taste-feature --output /tmp/taste-stress-mutants > stage3/fixtures/taste/logs/stress-mutants.log 2>&1
python3 stage3/fixtures/taste/audit.py > stage3/fixtures/taste/logs/audit.log 2>&1
```

All commands exit 0; individual expected build failures are captured as data.
Initial setup: Go ready 0s, clang/Node/submodules ready 1s each, build cache warm
99s, total 99s. Follow-up setup: Go/clang/Node/submodules ready 0s each, build
cache warm 23s, total 23s; `nproc` 5, cgroup CPU quota 4. Both setup logs are
retained under `logs/`.
The final validation ran 24 source Node executions, 48 compiler builds, twelve
compiled native executions, fifteen source-rewrite Node/build comparisons, eleven
main mutant native executions, and four taste stress-mutant builds/executions. The whole uncached Go gate was not run: this
unit changes only fixtures and their observation tooling. The separate worker's
`fixtures_test.go` was neither created nor edited.

## Remaining limits

The follow-up adds the requested stress forms, unmasked rewrite proofs, and
minimum-coverage proportional sample. Some older exact-source fixtures still have
the independently recorded checker/representation blockers; source adaptations
or compiler implementation beyond this territory are required to make all of
them native. In particular star export remains refused on the newest taste tip.
No compiler code was edited, no full native tsc or full uncached Go gate was run,
and no claim of exhaustive sampling across all 62 files or random within-form
selection is made. The concrete allocation and every selected source site are
recorded rather than inferred from fixture filenames.

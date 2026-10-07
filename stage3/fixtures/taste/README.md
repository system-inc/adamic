# Taste fixtures: partial coverage

These 20 programs retain TypeScript 6.0.3 functions, arrows, export statements,
or selected scanner/parser statements. They are a curated partial set, **not the
requested proportional sample and not complete coverage**. No compiler was changed.

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
normalized to LF; their statements and signatures remain unchanged. Dependencies
are small declared shapes or explicit stubs, not a claim to implement the omitted
compiler subsystem. The drivers include normal tsc inputs and boundary probes.

Census reasons answered here: `non-boolean control condition` (6,697 sites,
6,635 lines, 62 files), `||=` (110 sites, 109 lines), `the comma operator`
(47 sites, 46 lines), `a label` (10 sites/lines), `the void operator` (15),
and `an ExportDeclaration` (77).
The full AST walk also finds 173 `??=` and one `&&=`: the census's 110 is
specifically `||=`, not the entire logical assignment family. The two `??=`
fixtures are supplemental usage of that family, with the census family heading
in their provenance comments. Likewise operand-valued logical operations and
negation are supplemental to the typed control-condition count.
The walk counts 75 `export *` and **2 named export lists**, both in
`_namespaces/ts.ts`. A named-export success does not establish star-barrel support.
The full recount and source selection are reproducible with the commands below.

## Observed outcomes

Every entry was run using the unchanged `oracle/node.mjs`, with no enum or
namespace runner changes. Every entry was then compiled using `go run ./cmd/adamic
build` on main and on the taste checkout. Every successful binary was executed.
All 20 Node runs exit 0 with empty stderr. Main compiles none. Taste compiles
seven, and **all seven match stdout, stderr, and exit exactly**. No silent
miscompile was observed among binaries that compiled; no claim is made about
programs blocked before execution.

`status.json` has exactly the requested schema and records main; `evidence.json`
records both builds and every executed native binary. Nonzero go-run diagnostics
retain their exact stdout/stderr, including `exit status 1`. `Checker` means a
checker error, `NotYet` follows the compiler's actual “can't lower ... yet”
message, and `Refused` follows its actual “Adamic 0.1 refuses ...” message.

| Fixture | Real form / driver coverage | Main | Taste |
| --- | --- | --- | --- |
| `01_diagnostic_code.a` | numeric operand-valued ||: 0, -0, NaN, diagnostic codes | NotYet | Compiles (Node match) |
| `02_diagnostic_message.a` | string/object union operand-valued || | NotYet | NotYet |
| `03_diagnostic_file.a` | object/undefined conditional expression | Refused | Compiles (Node match) |
| `04_jsx_runtime.a` | empty string/undefined conditional expression | Refused | Compiles (Node match) |
| `05_this_type.a` | !! over numeric flags and optional boolean | NotYet | NotYet |
| `06_localized_message.a` | &&/|| values from localized dictionary lookup | Refused | Refused |
| `07_compact.a` | !value, array/object truthiness, mixed ??, ??= and if | NotYet | NotYet |
| `08_for_each.a` | callback result truthiness: number, string, object | Checker | Checker |
| `09_literal_cache.a` | comma cache fill and operand-valued || | Refused | NotYet |
| `10_global_import_meta.a` | lazy ||= cache and operand-valued || | Refused | Compiles (Node match) |
| `11_build_info_pending.a` | field ??= and !! of optional booleans | NotYet | NotYet |
| `12_symbol_links.a` | indexed ??= and bitmask condition | Refused | NotYet |
| `13_void_callback.a` | void push callback returns undefined | Refused | NotYet |
| `14_relative_complement.a` | two labels and continue across nested loops | Checker | Checker |
| `15_barrel.a` | actual star barrel re-export | Refused | Refused |
| `16_scan_exclamation.a` | scanner comma returns advancing position and setting token | Refused | NotYet |
| `17_binder_flow.a` | binder ||= preserves earlier flow effects | Refused | Compiles (Node match) |
| `18_named_export.a` | actual named export list from a barrel | Refused | Compiles (Node match) |
| `19_deprecated_flags.a` | numeric flags coerced by !! | NotYet | Compiles (Node match) |
| `20_jsdoc_terminate.a` | JSDoc labeled break terminates at four real delimiters | Refused | NotYet |

Notable blockers: string/object/undefined fields in 02, optional boolean fields
in 05 and 11, unrestricted string index signatures in 06, generic array lowering
in 07, checked generic indexed reads in 08 and 14, inferred `any` in 09, an array
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
complete real binder function. These cuts are explicit, not rewritten tsc logic.

## Rewrite mutants and what caught them

These are explicit source rewrites on main, not compiler edits. `mutants.json`
retains every result and the full diagnostic. All twelve rewrites preserve the
original Node stdout, stderr, and exit. Ten flip the main outcome, proving those
fixtures reach the corresponding form. Seven rewritten programs compile and their
native observations match Node. A flip to NotYet establishes removal of the form's
refusal and exposure of another blocker; it does not establish native correctness.

| Rewrite / fixture | Main before and after | Evidence |
| --- | --- | --- |
| `operand` / `01_diagnostic_code.a` | NotYet to Compiles | Outcome flip; Node unchanged and native matches |
| `object_condition` / `03_diagnostic_file.a` | Refused to Compiles | Outcome flip; Node unchanged and native matches |
| `string_condition` / `04_jsx_runtime.a` | Refused to Compiles | Outcome flip; Node unchanged and native matches |
| `masked_double_not` / `05_this_type.a` | NotYet to NotYet | Masked, no proof; Node unchanged |
| `logical_assignment` / `17_binder_flow.a` | Refused to Compiles | Outcome flip; Node unchanged and native matches |
| `comma` / `16_scan_exclamation.a` | Refused to NotYet | Outcome flip; Node unchanged |
| `void` / `13_void_callback.a` | Refused to NotYet | Outcome flip; Node unchanged |
| `label` / `14_relative_complement.a` | Checker to Checker | Masked, no proof; Node unchanged |
| `double_not` / `19_deprecated_flags.a` | NotYet to Compiles | Outcome flip; Node unchanged and native matches |
| `label_break` / `20_jsdoc_terminate.a` | Refused to NotYet | Outcome flip; Node unchanged |
| `named_export` / `18_named_export.a` | Refused to Compiles | Outcome flip; Node unchanged and native matches |
| `star_export` / `15_barrel.a` | Refused to Compiles | Outcome flip; Node unchanged and native matches |

The object and string rewrites make truthiness explicit. The numeric `||`
rewrite handles absence, both zeros, and NaN explicitly. Numeric `!!` becomes a
nonzero bitmask comparison. `||=` becomes its explicit boolean if. The comma
rewrite sequences the position update before returning the token assignment.
The void rewrite becomes a block that discards push's result. The labeled-break
rewrite restores scanner state and returns directly. The nested-continue rewrite
uses an explicit outer-iteration flag; it preserves Node but remains checker
masked. Star export removal leaves the driver dependency import; named export
moves export to the supporting declaration. These export rewrites prove that the
barrel statement reaches main's refusal, not full star/live-binding semantics.

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
node stage3/fixtures/taste/inventory.cjs /tmp/taste-typescript /tmp/taste-api/node_modules/typescript/lib/typescript.js /tmp/taste-census /tmp/taste-inventory.json > /tmp/taste-inventory.log 2>&1
python3 stage3/fixtures/taste/observe.py --main /workspace/adamic --taste /tmp/taste-feature --logs /tmp/taste-runs-final4 > /tmp/taste-observe-final4.log 2>&1
python3 stage3/fixtures/taste/mutants.py --main /workspace/adamic --output /tmp/taste-mutants-final > /tmp/taste-mutants-final.log 2>&1
python3 stage3/fixtures/taste/audit.py > stage3/fixtures/taste/logs/audit.log 2>&1
```

All commands exit 0; individual expected build failures are captured as data.
Setup: Go ready 0s, clang/Node/submodules ready 1s each, build cache warm 99s,
total 99s; `nproc` 5, cgroup CPU quota 4. Logs are retained under `logs/`.
The final validation ran 20 source Node executions, 40 compiler builds, seven
compiled native executions, twelve source-rewrite Node/build comparisons, and
seven mutant native executions. The whole uncached Go gate was not run: this
unit changes only fixtures and their observation tooling. The separate worker's
`fixtures_test.go` was neither created nor edited.

## Not completed

I did not produce a proportional sample. Small self-contained functions were
selected for reviewable extraction and blockers, rather than distributing draws
across all 62 files and 6,697 typed conditions. I did not claim all hardest cases:
compact and relativeComplement are included, but compiler/generic/array blockers
prevent their native evaluation on either revision.

Non-boolean conditions in while/for/do, labeled jumps through finally,
side-effectful receiver/index single-evaluation stress, field `||=`/`&&=`, and the
single real scanner `&&=` are not covered. The field and indexed `??=` samples
are blocked and do not prove receiver-once semantics. Ordinary loops in the
fixtures do not substitute for non-boolean loop conditions. The optional-boolean
`!!` and nested-label mutants are masked, and no full native proof is claimed
for comma, void, or labeled control. These omissions need additional fixture
work and, for several preserved whole functions, the independent array/type
representation features. The set is useful input for that work, not completion
of the requested bucket.

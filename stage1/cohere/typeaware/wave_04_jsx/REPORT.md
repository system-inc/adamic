Built: partial node-local JSX kernels/reporters for three new claimed rules; the earlier HIR React claims are parked.
Commits: claim da5982d4 was pushed before implementation; this partial implementation/evidence commit follows.
Commands: validate_partial.py PASS 56 records / 4,941 bytes across Go/native/sanitizers/source Node/emitted JS; positive source Go controls PASS.
Mutants: three successful-exit comparison mutants and three removed-source-refusal mutants caught as described below.
Uncovered: full source/corpus findings, fixes/suggestions parity, complete rule mutants, checker-handle integration, timings and full repository gate.

## Scope and evidence

Code stays in this owned directory, with separate rule directories containing
`rule.a` and `rule.json`. All listeners use pinned numeric SyntaxKinds:
fragments 285/286/289; constructed-context-values and no-undef 286/287.
The Go oracle provides those enum values independently. Rule entry points take
a handed node, and construction classification uses its numeric kind directly.
No per-rule parser lookup or string-kind relevance dispatch was added.

`jsx-fragments`: both production message IDs and exact text, token spans and
zero repairs/suggestions. `jsx-no-undef`: production component-name split,
including empty names, dashes, ASCII lowercase, `$`, `_` and non-ASCII names,
plus exact diagnostic rendering. `jsx-no-constructed-context-values`: direct
construction classification for ten AST cases, function-remedy classification,
and both direct-message arms with exact text, line, span and zero repairs.
These are partial kernels, not complete source analyses. `run(node)` explicitly
refuses unavailable source integration rather than returning an empty verdict.

The independent owned Go overlay calls production private helpers and message
objects. Positive source controls separately invoke all three unchanged
production rules with `RunTypedFiles`, and report one finding each. Those
source findings are retained as blocker evidence, not native parity results.
No Go lint verdict is used as a native implementation input.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_jsx/validate_partial.py /workspace/typeaware-wave-04-jsx/final > /workspace/typeaware-wave-04-jsx/final.log 2>&1
```

The validator uses the fresh compiler built on main f8013f0b. All normal and
sanitized backend stderr is empty. Mutants compile and exit zero, and only
independent Go output bytes catch them: fragment end+1; no-undef's lowercase
lower bound 97 changed to 98; object construction classified as array. Three
additional mutants remove the `run(node)` refusal, compile and exit zero,
and are caught by the required exit-70 assertion. They are kernel/reporting and
refusal mutants, not complete-rule mutants. The first isolated fragment mutant
failed nominal ancestry before comparison; it was not counted. The boundary was
changed to a structural handed-node interface and all final mutants compile.

Exact streams, expected bytes, source hashes, inputs and subprocess results
are in `evidence/`. No shared parser, harness, driver, registration generator,
bridge dispatch or protected compiler file was changed. Earlier complete ports
retain their current-main landing evidence; these new modules are not imported
by their runners.

## Exact blockers and remaining port work

The valid JSX probe `const x = <Missing />;` compiles as an owned parser runner,
then exits 70 with:

```
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 19 in control.tsx
```

Current shared `ParseNode.kind` is still a string. The isolated wave-16 JSX
parser also exposes string kinds; substituting it alone does not provide the
required handed numeric-node interface. Shared parser/driver integration is
outside this unit's authorized territory. JSX support is landing on
`area/stage1-lint`; no integration branch was modified or pushed here.

Beyond that shared blocker, these own-rule parts remain unported:

- Fragments: fragment binding resolution across import aliases, destructuring,
  initializer declarations, attributes, default-mode judgment and optional mode.
- No-undef: numeric tag-tree reference selection, symbol declaration file facts,
  `.cjs`/allowGlobals handling and final token-span placement.
- Constructed context values: provider resolution, value extraction, component
  ancestry, recursive construction/declaration following, indirect-message Go
  quoting, and the bounded AST stability/escape walk (including memo dependency
  and no-dependency messages). This rule does not require the parked HIR pipeline.

The latter stability analysis is rule-local work, not a claim that all remaining
work is blocked externally. Current source execution is refused until these
parts and numeric JSX integration are complete. No corpus timing or full source
agreement is claimed for these modules. Prior completed continuation measurements
remain compiler native 2.130260 s / Go 0.387003 s and repository native 0.318505 s /
Go 0.127645 s; concurrent validation means these are single-run observations.

The new claims remain unfinished and reserved to wave-04. No fourth rule was
claimed. The parked HIR claims remain reserved, with #dnv6f2c and JSX prerequisites
named in claims/wave-04.md.

## Numeric JSX tag-reference and factory helpers

The three existing claims now include additional production-Go-held helpers.
`NumericTag` holds numeric SyntaxKind, identity, child links, text and generic
call argument links. Helpers act on the root they receive and descend only into
its children; they do not refetch that root from a parser or compare kind strings.

- No-undef reference selection accepts a bare component name, or the leftmost
  identifier of a member chain at any case. Namespaced and `this`-rooted tags
  decline. Identity is preserved rather than replaced with the property's span.
- Qualified fragment recognition requires exactly `React.Fragment`, with both
  identifier kinds checked. A deeper qualifier declines. Fragment initializer
  recognition accepts bare React, that qualified name and a bare require call
  whose first string-like argument is exactly react (including a no-substitution
  template); preact, missing arguments and unrelated callees decline.
- Context factory recognition accepts bare createContext and React.createContext,
  unwrapping parentheses around the callee and receiver, matching Go's helper.

The owned Go overlay constructs real pinned AST nodes using NodeFactory and
calls production `resolvableJsxReference`, `jsxFragmentsNameIsFragment` (qualified
routes only), `jsxFragmentsInitializerIsFragmentSource` and
`jsxNoConstructedContextValuesIsCreateContextCallee`. It does not copy predicates.
Forty-two numeric AST nodes produce 152 records / 2,045 identical bytes across
native, ASan/UBSan/leaks, source Node and emitted JavaScript, all with empty
backend stderr. `validate_references.py` prints PASS. The earlier partial
reporting validator was rerun and PASS: 56 / 4,941 bytes, its three comparison
mutants and three removed-refusal mutants.

Five sanitized reference mutants compile, exit zero and have empty stderr; only
Go bytes catch them: qualified React receiver changed to Other; require module
react changed to preact; template literal numeric kind 14 changed to 8;
member-root reference incorrectly subjected to the bare-name case predicate;
receiver-parentheses unwrapping disabled by changing kind 218 to 8. The initial
unwrapping mutant used kind 109 and panicked on a this receiver. That failure
was not counted and is retained; the replacement mutant fails only comparison.
The first generated empty-array call arguments hit stage 0's array-of-never
refusal. Explicitly typed numeric argument arrays resolved it without compiler
edits.

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave_04_jsx/validate_references.py /workspace/typeaware-wave-04-jsx/references/final > /workspace/typeaware-wave-04-jsx/references/final.log 2>&1
python3 stage1/cohere/typeaware/wave_04_jsx/validate_partial.py /workspace/typeaware-wave-04-jsx/references/reporting > /workspace/typeaware-wave-04-jsx/references/reporting.log 2>&1
```

Exact controls, hashes, backend and mutant streams are under
`evidence/references`. No Go regular expression occurs in these ported helpers;
no hand-rolled regex matcher was introduced. Character classification follows
Go's explicit non-regex component-name predicate. No position conversion was
added: source span placement remains outside these partial kernels.

Remaining blockers are unchanged shared numeric JSX integration (the current
parser probe still exits 70) and own rule work: symbol/declaration binding,
attribute judgment, final source spans and constructed-context component,
recursive construction and stability/escape decisions. Numeric tag selection
and the structural/initializer helpers listed above are now covered, superseding
those entries in the earlier unported list. Source `run(node)` still explicitly
refuses execution. No full corpus parity, full-rule mutants, checker-handle
integration, fresh throughput or full repository gate is claimed. No new claims.
Main remains f8013f0b, with the earlier six completed-rule landing oracles green.

## Named harness adoption

Current evidence supersedes the earlier JSX parser failure and main revision: JSX parsing now succeeds after adopting ab70f38d4; current main is c01907a7. Numeric registry integration and the own source-analysis work remain incomplete. See [HARNESS_AB70.md](HARNESS_AB70.md) for green completed-rule oracles, partial-kernel mutants, measured timings and exact limits.

## Value attribute extraction

Ported and independently compared ordered numeric value-attribute extraction, including invalid-first duplicate semantics. Expanded references pass 302 records and eight comparison-only mutants across five backends. See [ATTRIBUTES.md](ATTRIBUTES.md); full source rules remain incomplete and no new claims were taken.

## Named kind contract correction

The latest user instruction replaces numeric listener metadata with upstream ast.Kind names. All twelve declarations are corrected and independently checked. Actual Descriptor deserialization now succeeds, superseding the earlier numeric registry blocker; source implementation remains unfinished. See [NAMED_KINDS.md](NAMED_KINDS.md). No new claims.

## Landing on b8fb957aa

Preserving merge history resolved the earlier rebase conflict without shared edits. Current main and named harness 41eb6eab2 are adopted, and all six completed-rule and partial-kernel oracles are green again. See [LANDING_B8.md](LANDING_B8.md). JSX source claims remain unfinished; no new claims.

## Integration area landing

Rebased onto fetched area/stage1-lint 7481e0324, containing 50a5f105 and current main 39638d9e2. Six completed-rule and partial helper/listener oracles are green again. See [LANDING_AREA.md](LANDING_AREA.md) for exact tested base, timings and limits. No new claims; JSX source rules remain unfinished.

## Restored environment landing

Rebased onto current area d65a8f931 with current main 39638d9e2. Fresh rule/sanitizer/mutant and runtime checks are green after recovering disk capacity and retrying every affected test. See [LANDING_D65.md](LANDING_D65.md). No new claims; JSX source work remains unfinished.

## Fragment declarations

Ported fragment import/alias/destructure/variable declaration judgments and exact module ancestry; production-Go comparisons pass 1291 records with twelve clean reference mutants across five backends. See [FRAGMENT_DECLARATIONS.md](FRAGMENT_DECLARATIONS.md). Source symbol resolution/adapters remain unfinished; no new claims.

## Native fragment and undefined-name source visitors

See [SOURCE_RULES.md](SOURCE_RULES.md). Both reserved source visitors now match production Go on positive controls and both frozen corpora under sanitizers, with comparison-only mutants and released-handle refusal. Constructed-context source analysis and shared checker/factory integration remain unfinished. No new claims.

## Current main and lint-area landing

[LANDING_C799.md](LANDING_C799.md) records the rebase onto main c7991b900 and area b84a9d931, eight source-rule oracles, all partial validators, adopted proof/record checks and mutants. The final area rebase preserves exactly the tested Git tree. Constructed-context source analysis and shared checker/factory integration remain unfinished. No new claims.

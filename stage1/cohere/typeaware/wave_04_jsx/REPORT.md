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

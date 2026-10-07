Claimed, not ported: `react-hooks/set-state-in-effect`.
Claim commit: `e21e2919`, pushed before implementation; previous nine ports remain pushed.
Observed: production React control tests pass; valid TSX reaches native parser panic 70.
Mutants: none for this blocked claim; no agreement or sanitizer result is claimed.
Not covered: implementation, full byte comparison, mutants, released handles, timing.

The native parser used by the existing rule runners has no JSX implementation.
A valid input ending in `return <Component />;` is accepted by the independent
Go runner (exit 0), but the existing native runner exits 70:

```
parser slice expected GreaterThanToken, got SlashToken at 94
```

This is an observed parser failure before any rule runs, not a failure to load
`.a` modules. `stage1/typescript/parser/parser.ts` treats angle brackets as
TypeScript type assertions, and its grammar, statements and lookahead sources
contain no JSX productions. `WHOLE_REPORT.md` explicitly excludes JSX.
The production tests exercise JSX elements/fragments in all three selected
React rule families. Shared parser changes are outside the scope authorized
by Ahra's correction, so this worker stops instead of modifying those files.
No placeholder rule that silently returns zero findings has been added.

These production rules also depend on React HIR lowering, SSA propagation,
compilation-unit gates, and, for the setter rules, control/post-dominance and
manual-memoization processing. There is no native Adamic React HIR port under
`stage1/cohere`. This is a separate prerequisite to faithful implementation;
exposing Go HIR lint judgments through the checker bridge would violate the
raw-facts contract. Non-JSX examples alone would not establish the requested
whole-rule agreement. No claim of impossibility in principle is made.

Selection inspected 389 fetched origin refs, 33 distinct Markdown claim
blobs, and native source ports on origin/main and origin/codex/tsgo-c-library.
The first three remaining rules, each with combined selection volume zero,
were set-state-in-effect, set-state-in-render and static-components. Claims
remain reserved and explicitly blocked; no later candidates were claimed.

Required setup succeeded: Go ready 0s, clang ready 0s, Node ready 0s,
submodules ready 0s, cache warm 38s, done 38s; `nproc` is 5 (quota 4 cores).
Environment sourced from `/workspace/adamic-tools/env.sh`.

Commands, with every test output redirected to logs:

```sh
bash cloud/setup.sh > /workspace/wave20-validation/fourth/setup.log 2>&1
source /workspace/adamic-tools/env.sh
cd /workspace/adamic/cohere
go test ./internal/lint/rules/react -run '^(TestStaticComponentsFires|TestSetStateInRenderFires|TestSetStateInEffectFires)$' -count=1 -v > /workspace/wave20-validation/fourth/go-react-controls.log 2>&1
```

Observed production test output: PASS, package 0.082s. The parser probe uses
the unchanged prior batch's independent Go oracle and native suite on the same
TSX source/config/manifest. Their selected core rules are irrelevant to this
probe: native fails during parsing before dispatch. The production React tests
separately prove that these constructs are real inputs to the claimed rules.
No full repository gate was run because no implementation changed.

Common evidence is committed under
`../react-hooks-static-components/validation/`: selection and production test
logs, setup timings, parser probe input/config/manifest and outputs, and push
attempt logs. The first claim push received a generic remote failure; retry
succeeded. No shared harness, generator, parser or compiler file was changed.

## Dependency recheck on October 7

Fetched all 417 origin refs after the next keep-cooking request. Native JSX
support DOES now exist on `origin/codex/stage1-jsx-lint`, implemented in
`e715ef4a2f898230af63c40195dea6586a557899`, with evidence at tip `a8a62d62`.
The earlier absence statement applies to this worker's checkout, not all
origin branches. The change introduces `parser/jsx.ts` and modifies the
shared `parser.ts`, `lookahead.ts`, and `scanner.ts`; it has not been
integrated into this branch. Its lint work ports batch8 syntax rules, not
these React HIR validators. The shared harness tip is now `f4d98cab`.

Ahra explicitly instructed: "Keep your changes inside your own rule
directories." Applying the JSX changes would change those shared files,
so this worker has not cherry-picked or copied them into the checkout.
The integration owner can land that dependency; afterward these three rules
still require native React HIR lowering and their value/control analyses.
No additional rules were claimed, and no new implementation or validation
pass is asserted by this dependency inventory. Existing results above are
historical, not reruns. The current checkout and shared files remain unchanged.

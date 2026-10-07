# Wave 15 parked from unified landing

All wave-15 implementations remain on codex/typeaware-wave-15 as historical standalone bridge work. None is certified on the unified registry. No wave-15 implementation is carried onto codex/lint-port-no-proto, which starts directly from origin/area/stage1-lint. Step 1 has no remaining unblocked owned rule to land on the unified harness. Existing private-suite byte comparisons are not unified certification.

The shared RuleContext has no checker Program or declaration/type fact provider. The following minimal source reproducers require that missing input to match unmodified Go on all cases:

| Rule | Reproducer | Required fact/blocker |
| --- | --- | --- |
| @typescript-eslint/no-useless-default-assignment | `function f(x = 1) { x = 2; return x; }` | checker-backed declaration and usage facts |
| @typescript-eslint/prefer-find | `const a: number[] = []; a.filter(x => x > 0)[0];` | receiver type and array method facts |
| @typescript-eslint/require-array-sort-compare | `const a: number[] = []; a.sort();` | receiver/element types |
| nexus/correctness-no-global-listener-target-assertion | `addEventListener('click', e => (e.target as HTMLElement).click());` | listener and target types/declarations |
| nexus/correctness-no-leaked-number-render | `const n = 1; const view = <div>{n && <span/>}</div>;` | JSX integration and numeric type facts |
| nexus/correctness-no-mock-on-module-namespace | `import * as mod from './m'; jest.spyOn(mod, 'run');` | namespace/module declaration resolution |
| no-invalid-regexp | `new RegExp('[');` | constructor identity and ECMAScript pattern validation |
| no-label-var | `let label = 1; label: while (true) { break label; }` | checker locals and label binding |
| no-misleading-character-class | `/[👍🏽]/u;` | RegexSyntax/reference/constant helpers; existing scanner is not a complete RegExp port |
| no-throw-literal | `function f(undefined: Error) { throw undefined; }` | global-versus-shadowed undefined symbol (even this syntactic rule declares a checker) |
| no-useless-backreference | `/(a\1)/;` | RegexSyntax/capture/reference helpers; hand-written scanner is not complete under the RegExp instruction |
| prefer-arrow-callback | `[1].map(function f(x) { return f(x); });` | recursive callback symbol/declaration identity |
| react-hooks/set-state-in-effect | `useEffect(() => { setState(1); }, []);` | native HIR, capture translation, memo preparation and dominators |
| react-hooks/set-state-in-render | `function C() { setState(1); return null; }` | native HIR, SSA setter/capture propagation and unconditional blocks |
| react-hooks/static-components | `function C() { const Child = () => null; return <Child/>; }` | native HIR, phi/instruction taint, capture/compilation gates |
| react/jsx-fragments | `import { Fragment as F } from 'react'; const v = <F/>;` | checker import/declaration facts |
| react/jsx-no-undef | `function C() { return <Missing/>; }` | symbol and declaration-file facts |
| react/jsx-no-constructed-context-values | `function C() { const dep = {}; const value = useMemo(() => ({}), [dep]); return <Ctx.Provider value={value}/>; }` | checker declarations/types plus memo/capture/reference analysis |

React HIR and capture-dependent claims are parked pending #dnv6f2c. The context rule's quoted-name and raw construction judgments are preserved, but they do not replace live analysis. Reproducers are raw source for upstream/shared-fact integration; they are not substituted Go verdicts. The new no-proto branch carries none of these blocked implementations, descriptors or standalone harness edits.

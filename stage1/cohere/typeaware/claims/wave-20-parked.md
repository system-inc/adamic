# Wave 20 parked work

All twelve private analyses are parked from unified-harness landing. They require checker facts through RuleContext, which currently exposes parser, scanner and settings but no checker program or query lease. Their existing source and private oracle evidence remain on codex/typeaware-wave-20 for recovery; none is registered under lint/rules. No landing branch is proposed for these analyses.

| Rule | Minimal source reproducer | Required checker fact |
| --- | --- | --- |
| @typescript-eslint/no-floating-promises | `declare function f(): Promise<void>; f();` | Promise and callback types |
| @typescript-eslint/no-implied-eval | `setTimeout('work()', 1);` | Resolved global callee |
| @typescript-eslint/no-meaningless-void-operator | `declare function f(): void; void f();` | Operand return type |
| nexus/correctness-no-process-exit-after-output | `console.log('x'); process.exit(0);` | Resolved standard stream/callee |
| nexus/correctness-no-uncleared-race-timeout | `Promise.race([work(), new Promise(r => setTimeout(r, 1))]);` | Resolved Promise and timer |
| nexus/correctness-require-blocking-standard-streams | `process.stdout.write('x');` | Standard stream declaration |
| prefer-promise-reject-errors | `Promise.reject('x');` | Binding origin |
| prefer-regex-literals | `new RegExp('x');` | Binding origin |
| prefer-rest-params | `function f() { return arguments[0]; }` | Binding origin |
| react/jsx-fragments | `import React from 'react'; const x = <React.Fragment />;` | React binding origin |
| react/jsx-no-undef | `const x = <Missing />;` | Declaration lookup and project globals |
| react/jsx-no-constructed-context-values | `const C = React.createContext(null); function App() { return <C.Provider value={{}} />; }` | Context/callee/type facts |

Reproduction of the integration blocker: a unified listener cannot call context.checker or context.program; neither field exists in stage1/cohere/lint/context.ts. Private suites explicitly open the bridge program instead. Exposing that shared lifecycle is outside this rule unit. React-hooks/set-state-in-effect, set-state-in-render and static-components remain separately parked on source-to-HIR, SSA and capture analysis, with existing BLOCKED.md reproducers. This note does not claim unified certification.

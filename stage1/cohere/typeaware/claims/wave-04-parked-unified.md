# Parked from unified-harness landing

All twelve wave-04 claims are parked for unified-harness certification. None has a descriptor under stage1/cohere/lint/rules, so none enters the unified landing registry. The private bridge implementations and evidence remain on codex/typeaware-wave-04 for later integration. The eight private source-oracle passes do not certify the unified harness. Since every owned rule is blocked there, the requested landing-only branch is skipped.

RuleContext on lint area d3a37422c owns source/parser/settings/findings but no checker program, symbol/type questions or raw declaration query surface. Constructing the private source visitors requires such a program; replacing it with an empty answer would invalidate the rule. These are reproducer sources for that unavailable dependency, not new claimed positive-oracle certifications:

| Rule | Reproducer | Unified blocker |
| --- | --- | --- |
| nexus/consistency-require-constant-casing | `const BadName = 1;` | Checker-backed initializer/type classification |
| nexus/consistency-require-matching-return-type | `function getString(): number { return 1; }` | Checker return-type/name classification |
| @typescript-eslint/strict-void-return | `declare function takes(callback:()=>void):void; takes(()=>1);` | Contextual/resolved return types |
| nexus/correctness-no-process-exit-after-output | `import process from 'node:process'; process.stdout.write('x'); process.exit(0);` | Resolved declaration and process-module identity |
| nexus/correctness-no-uncleared-race-timeout | `Promise.race([work(),new Promise(resolve=>setTimeout(resolve,100))]);` | Resolved declaration and callback/flow identities |
| nexus/correctness-require-blocking-standard-streams | `import process from 'node:process'; process.stdout.write('x');` | Process stream identities and source flow |
| react-hooks/preserve-manual-memoization | `function Component({value}) { return useMemo(()=>({value}),[]); }` | Native React HIR/SSA/reactivity analysis |
| react-hooks/purity | `function Component(){ return Date.now(); }` | Native React HIR and capture propagation |
| react-hooks/refs | `function Component(){ const ref=useRef(0); return ref.current; }` | Native HIR/SSA/ref/capture analysis and checker |
| react/jsx-fragments | `import {Fragment} from 'react'; const view=<Fragment/>;` | JSX parser plus checker import/alias declarations |
| react/jsx-no-undef | `const view=<Missing/>;` | JSX parser plus checker binding/scope declarations |
| react/jsx-no-constructed-context-values | `declare const Ctx:any; function Component(){return <Ctx.Provider value={({} satisfies object)}/>;}` | JSX/checker integration; unchanged production Go AsAsExpression cast panics at jsx_no_constructed_context_values.go:484 |

The exact constructed-context failing source is also in wave_04_jsx/jsx_no_constructed_context_values/testdata/satisfies.tsx.txt. Current complete raw comparison/mutant/sanitizer evidence is in wave_04_jsx/LANDING_D3.md and its archive; React HIR dependency reproductions are in wave_04_react/REPORT.md. No parser, checker or harness guard was relaxed. The new explicitly assigned ambiguous-identifier rule proceeds on a separate clean branch from the lint area.

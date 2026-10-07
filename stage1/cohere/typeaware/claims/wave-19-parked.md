# Wave 19 parked on unified checker context

All nine owned algorithms are parked for the unified harness. No wave-19 rule has a registered unified descriptor, so no blocked rule enters a landing registry. Step 1 has no eligible rule and no landing branch is created. Existing isolated implementations and their Go parity evidence remain reference material on this branch.

Shared blocker: stage1/cohere/lint/context.ts RuleContext has no checker program, symbol lookup or raw fact query API; lint_test.go buildPort does not link the tsgo archive. A factory cannot construct the owned checker Context from RuleContext. Do not replace checker judgments with syntax guesses.

| Rule | Minimal source reproducer | Required fact |
| --- | --- | --- |
| @typescript-eslint/consistent-generic-constructors | `class Box<T>{} const x: Box<string> = new Box();` with isolatedDeclarations true | isolated-declarations |
| @typescript-eslint/dot-notation | `declare const x: {[key:string]:number}; x['value'];` with noPropertyAccessFromIndexSignature true | index-signature-access |
| @typescript-eslint/no-array-constructor | `function f(Array:any){ return new Array(1,2); }` | node-symbol-origin |
| nexus/correctness-no-uncleared-race-timeout | `Promise.race([work(), new Promise(resolve => setTimeout(resolve, 10))]);` | declaration-ancestry |
| nexus/correctness-no-process-exit-after-output | `process.stdout.write('x'); process.exit(0);` | declaration-ancestry and wave19-resolved-callee |
| nexus/correctness-require-blocking-standard-streams | `process.stdout.write('x');` in a CLI project | wave19-program-modules and declaration-ancestry |
| require-await | `declare function consume<T extends () => Promise<number>>(callback:T):void; consume(async () => 1);` | wave19-type-signatures, wave19-generic-call, wave19-type-members and wave19-heritage-members |
| symbol-description | `function f(Symbol:any){ return Symbol(); }` | declaration-ancestry distinguishes local and ambient Symbol |
| valid-typeof | `function f(undefined:number){ return typeof x === undefined; }` | declaration-ancestry distinguishes shadowed undefined |

A contract reproducer is an Adamic module importing the unified RuleContext and accessing context.program or context.symbol: neither member exists. The live-production refusal reproducer is TestWave19ThirdProductionPending with ADAMIC_WAVE19_THIRD_RELEASE_ARTIFACTS=/workspace/wave19-f801-scratch/third-release; all four unregistered questions produce panic 70. Seven prepared dispatch lines remain unapplied under the shared-file restriction. See wave_19_third/TYPEOF_LANDING_REPORT.md for the latest isolated parity, sanitizer and mutation evidence.

# Wave 01 parked outside the unified landing branch

All owned type-aware ports are parked for unified-harness integration. Their
standalone Go-captured-fact comparisons are not unified-harness certification.
RuleContext has no checker program handle; Go CFG facts also have no native
driver delivery. Implementations and evidence stay on codex/typeaware-wave-01,
not on codex/lint-port-no-ex-assign. No owned unblocked descriptor is available
for a separate step-1 landing branch.

| Rule | Small input reproducing the needed facts |
| --- | --- |
| nexus/correctness-no-implicit-return | function f(x:boolean){if(x)return 1;} |
| @typescript-eslint/no-deprecated | /** @deprecated */ function old(){}; old(); |
| no-else-return | function f(x:boolean){if(x)return 1;else return 2;} |
| nexus/correctness-require-child-process-error-listener | import {spawn} from 'node:child_process'; spawn('x'); |
| nexus/correctness-require-response-status-check | async function f(){const r=await fetch('/');return r.json();} |
| nexus/performance-no-independent-await-in-loop | async function f(xs:number[]){for(const x of xs){await Promise.resolve(1);}} |
| symbol-description | Symbol(); |
| require-await | const f:()=>Promise<number>=async()=>1; |
| require-atomic-updates | let x=0;async function f(){x+=await Promise.resolve(1);} |

React reservations remain parked separately in wave-01.md with HIR/SSA/capture
blockers. No parked descriptor is installed in the unified registry. These
inputs are minimal integration probes, not claims of native findings.

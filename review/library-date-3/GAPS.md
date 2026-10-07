# Date language boundaries for @system_adamic

The requested runner (af12899) together with the existing Date adaptations leaves
54 refusals. Each exported adapted program was independently checked with stock
tsc 6.0.3 and the runner's strict options/prelude. Matching diagnostic codes are
not-typescript, not language gaps. See triage-after.json and remaining-refusals.json.

| Remaining first refusal | Tests | Boundary |
|---|---:|---|
| Detached toString | 13 | First-class inherited method values |
| Detached getters or ISO method | 19 | Same boundary; wrong-receiver tests subsequently need catchable TypeError and some use arguments |
| Prototype read as an own field | 9 | Observable constructor/prototype values; three subsequently need catchable TypeError |
| Catch around ISO validation | 4 | User catch must handle native library validation exceptions |
| Borrowed ISO on an object lacking Date's slot | 2 | Catchable TypeError instead of native validation panic |
| new String | 1 | Boxed primitive construction |
| isPrototypeOf | 1 | Observable intrinsic prototype chain |
| Clock | 5 | Four Date.now tests and one no-argument new Date; intentionally nondeterministic |

No library implementation target remains among these first refusals. No language
feature was built. The 43 untyped Date locals in the standalone runner baseline
were handled by preserving the already implemented Date-local annotation adapter,
not by introducing any at runtime. General evolving-any locals remain a language
boundary outside that adaptation.

These one-line .a reproducers all pass stock tsc, execute on Node, and are refused
by Adamic. gap-proofs.json records complete commands' observations and diagnostics;
gap-proofs.log records their outputs. Node and compiler used TZ=UTC.

| Feature | One-line reproducer | Node stdout |
|---|---|---|
| Evolving any | `let result; result = new Date(0).getTime(); console.log(String(result));` | 0 |
| Detached method | `const f = new Date(0).toString; console.log(f.call(new Date(0)));` | UTC epoch string |
| Prototype observation | `console.log(typeof Date.prototype);` | object |
| Catchable library validation | `try { new Date(NaN).toISOString(); } catch (e) { console.log(String(e instanceof Error)); }` | true |
| Missing internal slot | `try { Date.prototype.toISOString.call({}); } catch (e) { console.log(String(e instanceof Error)); }` | true |
| Boxed String | `const s = new String("");` | empty |
| Prototype ancestry | `console.log(String(Function.prototype.isPrototypeOf(Date)));` | true |

The files are under gaps/. Stock tsc checks identical bytes under a scratch .ts
filename because tsc does not recognize .a; Node runs identical bytes as .mts.
Adamic compiles the original .a. The stock oracle helper uses exactly the options
in cmd/adamic-test262/typescript.cjs and internal/load/prelude.d.ts.

The prototype TypeError tests cannot be implemented as a NaN-valued prototype
Date: current V8 Date.prototype has no Date internal slot. Clock refusal reasons
explicitly say the wall clock is nondeterministic and cannot be compared with Node.

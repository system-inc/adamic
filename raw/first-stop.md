# ff5eddz stopped

Base: 1bb9663d4a9cbf1466a758c41c2293c9c869e6ab.

The requested literal rewrite removes the single-spread refusal at 103:67,
but exposes JSON.stringify object references at tracing.ts:104:46.
This violates the required no-new-startTracing-findings condition.
The comparison child exits 1 intentionally. All three census children exit 0.
The restoring mutant recovers the before startTracing findings exactly.
Both applies, both builds and harness builds, the Node byte witness, emission
audit, and actual adamic types on 79 roots pass. The sharded lane is unrun.
The separate full-47 replay failure at the sys helper import also reproduces
with untouched origin/main adaptation 47 and was not bypassed.

Candidate and detailed proof: typescript/meta-spread-checkpoint.
No final landing commit or main push.

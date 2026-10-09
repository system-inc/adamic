# Timer family

Three explicit tokens are replaced in the two ambient declarations at sys.ts:51–52.
These globals are captured only by getNodeSystem. Registration returns Node's
Timeout object; cancellation accepts that object or an absent handle. Callback
arguments are not inspected by this owner. The host-supplied timer signatures
remain separate: tests return numeric IDs, and external hosts can supply other
handle domains. Those need a correlated generic, not a Node-only substitution.

Current main is 5fda2d26, merged without rewriting worker history. Its compiler
and adaptation inputs are unchanged from the prior main measurement; only the
lane/oracle worker configuration changed. Setup finished in 38.204s, nproc=5.
Stock TypeScript 6.0.3 reports zero compiler diagnostics. Whole sys.ts emits
byte-identical JavaScript. The actual emitted-source throw mutant fails equality.

The latent census measures 36 -> 35 direct any sites, against main's original
37. Refused=5157 and NotYet=1468 remain unchanged: a removed any blocker can expose
another first error. The scanner removal was in c6ea4131. The timer return mutant
restores `: any` only in an isolated copy of the actual declaration and measures
36 sites again, exactly one extra function-return site. Its raw stream is retained.
The cancellation parameter still produces `a value of type any` at sys.ts:52:31
in this latent tool despite the concrete NodeJS.Timeout source type and a clean
stock checker. This is an observation, not a claim that all timer refusals closed.

The unmodified apply/oracle/lane passes with eight workers: 106366 passing,
one sanctioned API failure, zero pending, only api/typescript.d.ts differs.
The declaration guard has exactly the existing 222 sanctioned changes. This
matches the previous main oracle. No public API or compiler edit was added.

Commands: stock tsc.js -p TREE/src/compiler --noEmit; family-proof.cjs BEFORE AFTER
 timers OUTPUT; LATENT_ASSERT_NO_OUTPUT=1 /tmp/adaptation41-census TREE/src/compiler
 OUTPUT (control and actual-source mutant); census-report.py; bash stage3/lane/run.sh
 /workspace/adaptation41-timers-lane. Every command writes a log; raw logs, counts,
proofs and lane report are under evidence/timers/. No whole-package gate was run.

# Record table amendment, step B

Ruling: task b5w3ycg, Nov 7 00:55Z, road (b), JavaScript-changing.
Owners: 32 for symbolWalker table creation and cleanup; 30 for getOwnValues.
Measurement passed and was pushed before construction: 398 paired shards,
1109 identical calls including 445 formerly sparse calls, zero invalid ids or
keys, and matching Node, sanitized native and JavaScript witnesses.

Base remains origin/main 61880be6dd7884dec15cf17c73fd9d56ee760ee4 with exact
99z efa9c7dd7ffd4504068e6beeae26a2365eb145f1 retained. Its full apply, compiler
build, 79-root type check and ordinary binder stop at 1326:26 were recorded in
the completed measurement. The user's clarification accepts the one internal
declaration hunk, while requiring both public declarations byte-identical.

Construction checkpoint: whole-owner before/after validation, both owner
phases, idempotence and eight half-adapted/foreign-body mutants pass. The
normal pre-30 and pre-32 planner phases pass. A fully completed core already
contains an unrelated later addRange assertion, so full 30 replay has the
same inherited refusal at core.ts:992 as before this amendment. The focused
getOwnValues owner contract is independently checked on completed source.
The initial harness tried that full completed-core planner and recorded the
inherited refusal before switching to the actual owner contract. A second
harness comparison exposed LF in the measurement walker; the production port
retains CRLF and its normalized source exactly matches the measured candidate.

Full apply, replay, emit, declarations, lane, mutants and census are pending.

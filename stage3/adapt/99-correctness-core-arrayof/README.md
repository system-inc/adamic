# arrayOf builds its result by push

JavaScript-changing established-shape adaptation, task b5w3ycg, site F2 at
adapted core.ts:1339:12. The whole arrayOf owner is pinned before and after.
The result is explicitly T[] and grows by push, removing the any[] return
relation without a cast or a signature change.

Ruling: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.

The count gate is complete before adapter construction. COUNT.json retains
all five input classes, the independent raw-value reclassification, exact
value checks, test multiset and baseline diff agreement with stage3/lane/run.sh.
Only non-negative integer counts occur. Negative, fractional, NaN and infinite
counts each occur zero times. Node witnesses for 0 and 3 agree; -1, 1.5 and NaN
show the documented out-of-count divergence. Full after-apply proofs are pending.

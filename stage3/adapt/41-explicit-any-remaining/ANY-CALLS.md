# Any-call coverage

The disposition table has one historical call-returning-any row, sys.ts:378:63,
in scheduleNextPoll. Main and final census both have zero rows with that exact
reason. The body stores host.setTimeout's handle in its polling queue.
This owner is examined but remains with the shared host timer contract:
System and WatchHost erase H; numeric test hosts, Node handles and custom handles
all occur. The queue's actual pollScheduled value is H | false, despite the old
boolean annotation. Writing NodeJS.Timeout everywhere would exclude hosts;
ReturnType of the existing any-returning host would only rename any.

The Node ambient timer declaration is typed separately because getNodeSystem
captures the real Node global. Its actual return-any mutant is caught, but it is
not claimed to be a call-kind-specific mutant. No call-return owner adaptation
or dedicated call-return mutant is claimed. Zero observations do not establish
that any calls compile: the census stops early and skips diagnosed bodies.

The correlated TimerHost<H> minimal program and all declaration/storage locations
are in RESIDUE.md. Public handle correlation requires additional API sanction
and owner/consumer proof. With has no disposition-table row or compiler-source
AST site and is checker-rejected; nothing was silently removed.

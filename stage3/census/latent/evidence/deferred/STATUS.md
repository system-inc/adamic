# Complete

The complete head stream was pushed in checkpoint
`dcae1d2042ec004349e7fc4fd45be3365e15fad7` before pairing.
The verified 018b7c21 baseline and all 711 adapted source hashes were reused.
Both census children and the pair exited 0.

See [counts.md](counts.md) for the complete measured result: zero units newly
stopped by deferred NotYet, 10 changed existing first stops, 17 new findings,
73 cleared findings, two moved findings and zero panics. The eligibility ledger,
504 generic declarations, 112 witnesses and overflow trailers match exactly.
No lowering state field changed. Required package, statecopy, predicate, meter,
pair, overlay and real fixture checks passed; the dropped-read mutant exited 1.

The baseline pooled-checker TS2859 was first observed in the run-end probe,
after the last timed unit src/tsc/_namespaces/ts.ts:3:1. The onset unit is unknown.
Production sources are unchanged, preserving all 73 core.ts:2599:15 first stops;
the roots were not recompiled.

Code and evidence are squashed into one commit on 018b7c21 for normal publication
to typescript/census-deferred. Historical incomplete-run notes remain in
previous-stop.md and incomplete/superseded logs. They do not describe this run.

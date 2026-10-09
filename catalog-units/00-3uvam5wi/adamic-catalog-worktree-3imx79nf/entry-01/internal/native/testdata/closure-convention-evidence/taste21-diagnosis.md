# Taste 21 diagnosis

The unchanged acceptance fixture is stage3/fixtures/taste/21_truthy_loops.a.
It extracts getAncestor, findActiveLabel and getRecursionIdentity from TypeScript
utilities.ts, binder.ts and checker.ts. Its census category is non-boolean
control conditions. Node exits 0 and prints true, true, outer, missing, true.

The branch exits 1 before C generation at 34:25:
`stage 0 can't lower checked view field objectType of type Type yet`.
The checked interface-view implementation supports required scalar fields;
this field is a recursive union/object field. Replacing the check with unchecked
ir.Narrow would discard the promised proof and is not an acceptable fix.
This is a real expected-success admission failure, not a native miscompile.
The full counts gate remains blocked; its incoming expected row was preserved.

Pinned main 48c05d09 also refuses the source, earlier at 31:13:
`Adamic 0.1 refuses a number as a condition; compare it explicitly, like name.length > 0 or count !== 0`.
The merged branch advances past that condition but has not closed the recursive
checked-view gap. No claim that main accepted the complete fixture is made.

A separate diagnostic source taste21-loops-only.a removes getRecursionIdentity
and its final invocation. All ancestor/label loops and their original calls
remain. Branch release compilation succeeds; native and source Node stdout
compare byte for byte: true, true, outer, missing. This probe is not substituted
for the full acceptance fixture and does not weaken its expected success.

Commands: Node 24 --input-type=module-typescript reads each .a via stdin;
adamic c runs the complete source with branch and main compilers;
adamic build runs the loops-only probe with default release flags.
Raw stdout/diagnostics are adjacent taste21-*.log files.

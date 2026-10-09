Built the exact const statSync options used by fixture 24 and refused unsafe optional Stats calls by name.
Parent: c2878c784e31e525fddec038b08a76fe8aa0b683; this piece precedes the requested area/library merge.
Checks: lower 102.573s, flow 263.825s, fresh 105.506s; directory/path/stat oracle 28.006s; full Linux counts 96.710s; final focused lower 10.142s, all passed.
Mutants: forcing throwIfNoEntry true passed sanitizers but failed Node output comparison; removing the optional Stats guard failed the named NotYet check. Earlier failed mutant setup attempts are retained as development logs.
Uncovered: macOS execution; fixture 25 remains a language checker failure; all 25 host acceptance stages will be recorded after merging area/library 53f44e05.

The accepted const binding must have a plain exact object initializer asserted as const. Arbitrary casts, optional fields, bigint stats and unsupported overloads remain refused. Existing fs-file runtime stat semantics are reused.

Fixture 25 reports TS2345 at 1001:23: T | undefined is not assignable to T. One-line reproducer: `function find<T>(a: readonly T[], p: (x:T)=>boolean): void { for(let i=0;i<a.length;i++) p(a[i]); }`. No acceptance source workaround was made.

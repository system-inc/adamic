Temporary: comes out when Node bindings for throwing data arguments lands

The guarded path remains a string. Only the data argument admission is widened; undefined still reaches Node and throws. No native binding is supplied.

Source: TypeScript 6.0.3, `src/compiler/tracing.ts`, `dumpLegend`.

Before: `fs.writeFileSync(legendPath, JSON.stringify(legend))`.

After: `(fs as { writeFileSync(path: string, data: string | undefined): void }).writeFileSync(legendPath, JSON.stringify(legend))`.

The script requires exactly one owner/site, is idempotent, and compares stock
TypeScript emitted JavaScript byte for byte before writing. Full stage 3
oracle and parser comparison results are recorded in the parser evidence.

Validated cumulatively after 60-62: full default stage 3 oracle 106,367
passing, zero failing/pending, empty baseline diff, 234.146 seconds. Only the
existing mechanically sanctioned API snapshot is accepted; its post-run
exact attribution check passes. The parser dump retains its original hash;
removing the view restores its diagnostic and the planted runtime edit fails
the JavaScript equality guard.

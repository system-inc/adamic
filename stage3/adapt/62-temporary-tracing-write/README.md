Temporary: comes out when Node bindings for throwing data arguments lands

The error is on serialized data, not the path. The temporary call view admits undefined, which Node rejects by throwing exactly as it did before. It does not implement a native binding.

Source: TypeScript 6.0.3, `src/compiler/tracing.ts`, `dumpTypes`.

Before: `fs.writeSync(typesFd, JSON.stringify(descriptor))`.

After: `(fs as { writeSync(fd: number, data: string | undefined): number }).writeSync(typesFd, JSON.stringify(descriptor))`.

The script requires exactly one owner/site, is idempotent, and compares stock
TypeScript emitted JavaScript byte for byte before writing. Full stage 3
oracle and parser comparison results are recorded in the parser evidence.

Validated cumulatively after 60 and 61: full default stage 3 oracle 106,367
passing, zero failing/pending, empty baseline diff, 244.528 seconds.
Only the existing mechanically sanctioned API snapshot is accepted; its
post-run exact attribution check passes. The parser dump retains its original
hash; removing the view restores its diagnostic and the planted runtime edit
fails the JavaScript equality guard.

# tracing.ts: 6 to 5, declined remainder

The private TraceRecord.typesPath optional field now truthfully permits explicit
undefined. stopTracing clears this precomputed path when a server trace has no
type catalog, because no types file will be created. This changes only the owning
annotation; it adds no initializer and preserves every read/write and object shape.
The full census is 370 to 369 with no shifted errors; whole files at zero remain
40 of the same 78 roots. This file does not count as zero.

Every remaining finding is recorded in closure-declined.json and proof.json:

| Line | Code | Reason / what would bring it to zero |
| --- | --- | --- |
| 40 | TS2591 | typeof import("fs") needs the compiler's proven Node module binding. |
| 63 | TS2591 | require needs the compiler's require-builtins declaration work. |
| 82 | TS2591 | process.pid needs the proven Node host declaration. |
| 83 | TS2591 | Same host declaration, at the second original read. |
| 66 | TS18046 | A catch can receive arbitrary thrown values. Nothing proves e.message is readable; Error/any casts would guess. A runtime guard changes this existing error path and emitted JavaScript. A proven exception contract or separately authorized runtime correction is needed. |

performanceCore.ts's two require findings remain explicitly outside this wave,
as requested. watch.ts's two findings require truthful owning public optional
host declarations; those would add API changes outside the exact sanctioned set,
so they remain declined. No source shim or unsound presence cast is substituted.

The final emitted-byte check compares directly against untouched input, without
calling the adapter to construct expected output. Both this wave and the entire
closure pass compare byte-identically to their original stock output.
All 26 files preserve stock-emitted JavaScript bytes and CRLF; the second adapter
run makes zero edits. Removing the typesPath undefined union restores TS2412
and fails the owning-declaration guard. Changing eventStack[index]! to ?? 0 is
caught by the emitted-byte check and required-read occurrence contract. The
default oracle passes 106,367 tests with zero failing and an empty baseline diff.
The public API remains exactly adaptation 20's 189 plus 40's 28 sanctioned lines;
70 is absent from this integration input and no other API line is accepted.

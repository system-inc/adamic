Runtime handoff: write_stderr_order.a

Ownership: runtime only. Fault: internal/native/runtime/input.c, adamic_write_text_file(), at the same open(name, O_WRONLY | O_CREAT | O_TRUNC | O_CLOEXEC, 0666) call. Its preceding adamic_output_flush() correctly prints first. With piped stderr, WASI path_open for dev/stderr under preopen fd 3 returns WASI errno 44, ENOENT. The fixture discards the returned Error, so second is absent from stderr.
The inherited stderr fd 2 is already present. The runtime worker owns the stream-alias write path and must preserve output ordering and descriptor lifetime. No runtime fix was authored here.

Exact original program, copied byte for byte:

```typescript
// writeTextFile to /dev/stderr after a console.log: with stdout and stderr on one pipe, the line
// printed first is out before the file's text arrives (output_test.go).
import { writeTextFile } from 'adamic';

console.log('first');
writeTextFile('/dev/stderr', 'second\n');
console.log('third');
```

Node source oracle: exit 0

stdout (UTF-8 JSON string):

```json
"first\nthird\n"
```

stderr (UTF-8 JSON string):

```json
"second\n"
```

WASI with emitted-C fix and runtime 52959fc: exit 0

stdout (UTF-8 JSON string):

```json
"first\nthird\n"
```

stderr (UTF-8 JSON string):

```json
""
```

Raw source, stdout and stderr are retained in this directory; observations.json in the parent records all five fixtures before and after.

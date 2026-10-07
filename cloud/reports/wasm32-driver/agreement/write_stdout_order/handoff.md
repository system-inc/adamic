Runtime handoff: write_stdout_order.a

Ownership: runtime only. Fault: internal/native/runtime/input.c, adamic_write_text_file(), specifically open(name, O_WRONLY | O_CREAT | O_TRUNC | O_CLOEXEC, 0666). Its existing adamic_output_flush() is called correctly before opening. With piped stdout, WASI path_open for dev/stdout under preopen fd 3 returns WASI errno 44, ENOENT. failure() consequently returns 'cannot write /dev/stdout: no such directory'.
The driver already supplies stdout fd 1 and root preopening; no compiler flag or emitted argument is wrong. Reopening a pipe through a filesystem path is the failing runtime/platform operation. The runtime worker owns stream handling through inherited descriptors, including flushing, UTF-8 encoding, borrowed descriptor lifetime and any distinction between pipe writes and regular-file truncation.

Exact original program, copied byte for byte:

```typescript
// writeTextFile to /dev/stdout between two console.logs. Node writes each line at once, so the file's
// text lands between them, and so must native's, whose stdout is buffered: a file is written only
// after what was printed is out (output_test.go runs this with stdout and stderr on one pipe).
import { writeTextFile } from 'adamic';

console.log('first');
const written = writeTextFile('/dev/stdout', 'second\n');
console.log(written.kind === 'Ok' ? 'third' : `third, after ${written.message}`);
```

Node source oracle: exit 0

stdout (UTF-8 JSON string):

```json
"first\nsecond\nthird\n"
```

stderr (UTF-8 JSON string):

```json
""
```

WASI with emitted-C fix and runtime 52959fc: exit 0

stdout (UTF-8 JSON string):

```json
"first\nthird, after cannot write /dev/stdout: no such directory\n"
```

stderr (UTF-8 JSON string):

```json
""
```

Raw source, stdout and stderr are retained in this directory; observations.json in the parent records all five fixtures before and after.

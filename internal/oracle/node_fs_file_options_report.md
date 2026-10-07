Fixed fs constant option reading for both compiler representations, on codex/host-proof-compiler.
The same files pass against old library base 460b9d7 and the compiler proof branch.
All 17 fs fixtures and 34 behavioral mutants pass, including drop-force caught only by Node stdout.
Effective adapted host matrix: 6/25 green, 1 Checker, 3 Refused, 15 NotYet; original matrix: 3 green.
Remaining: table below; no crypto/OS implementation changes, full gate or macOS run.

The precise representation change comes from objectLiteral's fit(value, declaredField):
optional boolean fields become MaybeOf; optional strings can become Box. The fs reader
previously recognized only raw scalar constants, falsely classifying these constant
wrappers as evaluated expressions. It now recursively unpacks only constant leaves.
Calls, reads and other effects remain refused. Old raw constants remain accepted.
An undefined field selects the existing Node default. Neither checker options nor
runtime error semantics were changed.

The rm fixture catches forced-missing failures so a mutant dropping force finishes
normally with no sanitizer/stderr failure. Its extra ENOENT stdout is detected solely
by comparison with the fixture source on Node. The old always-force mutant also remains.
No mutant killed by compile failure or an uncaught runtime error is claimed caught.

Validation:
- /tmp/fs_options_lower.log: fs lower tests PASS 4.666s.
- /tmp/fs_options_oracle_final.log: all fs fixtures and mutants PASS 10.082s.
- Old compiler validation, detached worktree /workspace/fs-options-old at 460b9d7,
  with the identical option reader and rm fixture/mutant files:
  /tmp/fs_options_old_compatible_lower.log PASS 9.691s,
  /tmp/fs_options_old_compatible_oracle.log PASS 19.013s.
  Its cohere commit 715ba94f is identical to the proof branch. The worktree uses
  symlinks to that submodule and the sole official @types/node installation.
  An initial attempt lacked the submodule checkout; the matching symlink fixed it.
- /tmp/fs_options_counts.log: complete Linux counts regeneration PASS 46.643s.
- /tmp/fs_options_counts_verify.log: complete counts verification without update PASS 36.139s.
- Original 25 runner /tmp/fs_options_original_final.log: all recorded Node comparisons
  pass; exit 1 for remaining blockers. The original 11,16,17 match Node on both backends.
- Adaptation-47 runner /tmp/fs_options_47_final.log: all five Node observations and
  all five behavior mutants pass; exit 1 for remaining 01/13 blockers.
  Adapted 02,03,04 match Node on both backends. Its other 20 controls are unchanged
  original sources already executed.
- Every setup/module-fetching Go command exported GOPROXY=https://proxy.golang.org|direct.
  No full go test ./... or macOS execution is claimed.

Stage columns each apply equally to native and JavaScript. Blocker/owner refers to
the adapted form, or original for unchanged controls. Original 01-04 remain the
audited TS2322 checker failures; original 13 remains TS18046.

| Fixture | Original N/J | Adapted N/J | First blocker | Owner |
| --- | --- | --- | --- | --- |
| 01_readFile_utf8.a | Checker | NotYet | an array of never | compiler |
| 02_readFile_utf16le.a | Checker | Compiles | Node agrees on both backends | none |
| 03_readFile_utf16be.a | Checker | Compiles | Node agrees on both backends | none |
| 04_readFile_missing.a | Checker | Compiles | Node agrees on both backends | none |
| 05_writeFile.a | Refused | Refused | a cast the runtime can't check; narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast) | adaptation |
| 06_fileExists.a | NotYet | NotYet | node:fs.symlinkSync | library |
| 07_directoryExists.a | NotYet | NotYet | node:fs.symlinkSync | library |
| 08_getDirectories.a | NotYet | NotYet | node:fs.symlinkSync | library |
| 09_realpath.a | NotYet | NotYet | node:fs.symlinkSync | library |
| 10_getModifiedTime.a | NotYet | NotYet | node:fs.symlinkSync | library |
| 11_setModifiedTime.a | Compiles | Compiles | Node agrees on both backends | none |
| 12_deleteFile.a | NotYet | NotYet | node:fs.symlinkSync | library |
| 13_createDirectory.a | Checker | Refused | in; an object's shape is known; use a discriminant, or a Map | compiler |
| 14_getCurrentDirectory.a | Refused | Refused | 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function callback() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) | compiler |
| 15_getExecutingFilePath.a | NotYet | NotYet | node:path.basename | library |
| 16_getEnvironmentVariable.a | Compiles | Compiles | Node agrees on both backends | none |
| 17_write.a | Compiles | Compiles | Node agrees on both backends | none |
| 18_exit_0.a | NotYet | NotYet | a value of type undefined | compiler |
| 19_exit_1.a | NotYet | NotYet | a value of type undefined | compiler |
| 20_exit_2.a | NotYet | NotYet | a value of type undefined | compiler |
| 21_createHash.a | NotYet | NotYet | node:crypto."node:crypto" read as a value | library (buffer worker) |
| 22_createHash_fallback.a | NotYet | NotYet | a value of type undefined | compiler |
| 23_newLine.a | NotYet | NotYet | reading _os | library (process worker) |
| 24_useCaseSensitiveFileNames.a | NotYet | NotYet | statSync: fs options other than a fixed literal or its plain const binding | library |
| 25_readDirectory.a | Checker | Checker | recorded strict diagnostics | adaptation |

One-line compiler reproducers, actually compiled with /tmp/fs-options-adamic:

Fixture 01:
~~~typescript
import { Buffer } from 'node:buffer'; const cases = [[], [0x61]]; for (const bytes of cases) { console.log(Buffer.from(bytes).toString('utf8')); }
~~~
~~~text
adamic: /tmp/fs-options-repros/01.a:1:54: stage 0 can't lower an array of never yet
~~~

Fixture 18:
~~~typescript
let activeSession: undefined; console.log(String(activeSession));
~~~
~~~text
adamic: /tmp/fs-options-repros/18.a:1:5: stage 0 can't lower a value of type undefined yet
~~~

Fixture 19:
~~~typescript
let activeSession: undefined; console.log(String(activeSession));
~~~
~~~text
adamic: /tmp/fs-options-repros/19.a:1:5: stage 0 can't lower a value of type undefined yet
~~~

Fixture 20:
~~~typescript
let activeSession: undefined; console.log(String(activeSession));
~~~
~~~text
adamic: /tmp/fs-options-repros/20.a:1:5: stage 0 can't lower a value of type undefined yet
~~~

Fixture 22:
~~~typescript
const crypto: undefined = undefined; console.log(crypto ? 'present' : 'absent');
~~~
~~~text
adamic: /tmp/fs-options-repros/22.a:1:7: stage 0 can't lower a value of type undefined yet
~~~

After cleanup no longer masks later lowering, 14 exposes the memoize callback
capture cycle refusal. Its exact diagnostic is retained in the observations.
05 still needs its driver catch cast adapted; 13 still needs unknown/in lowering.
21 and 23 retain their namespace-read blockers and were not implemented here.
This report supersedes the previous proof's zero-green matrix; the earlier report
remains as historical evidence of the representation regression.

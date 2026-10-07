Proof only: both compiler branches merged; zero of 25 original/adapted host drivers compile through cleanup.
Base library/merge-host-fs-file 460b9d7; compiler merge commits 61e9bb4 and 9f9a438.
All 25 original source observations and five changed adaptation-47 sources match recorded Node.
Five adapted behavior mutants caught; final compiler-feature oracle passes.
Remaining: matrix below; counts fail at fs option lowering; no full gate or Mac execution.

This branch is not for the area. No rebase, force push or main push. All three
_DARWIN_C_SOURCE defines remain, and no runtime C file was added. Both compiler
features compile past the old non-null and non-boolean barriers. The initial merge
incorrectly retained the old definite-assignment refusal: the feature oracle caught
it, and the final resolution uses the compiler branch's implemented loud readiness
checks while preserving the area's predicate proof.

Adaptation 47 changes only 01-04 and 13. The other 20 adapted controls are exactly
the original sources executed here. Audited fixtures/status.json and Node truth are
unchanged. Stages agree on both backends. No driver builds, so native host correctness
is not claimed. N is native, J is JavaScript. First blocker and owner below refer
to the adapted row; unchanged controls use their original observations.
The original six checker failures belong to adaptation. The unchecked catch cast
in 05 belongs to adaptation under the ruling requiring checkable casts.

| Fixture | Original N | Original J | Adapted N | Adapted J | First blocker | Owner |
| --- | --- | --- | --- | --- | --- | --- |
| 01_readFile_utf8.a | Checker | Checker | NotYet | NotYet | an array of never | compiler |
| 02_readFile_utf16le.a | Checker | Checker | NotYet | NotYet | rmSync: fs option literals containing evaluated expressions; bind a plain options object first | library |
| 03_readFile_utf16be.a | Checker | Checker | NotYet | NotYet | rmSync: fs option literals containing evaluated expressions; bind a plain options object first | library |
| 04_readFile_missing.a | Checker | Checker | NotYet | NotYet | rmSync: fs option literals containing evaluated expressions; bind a plain options object first | library |
| 05_writeFile.a | Refused | Refused | Refused | Refused | a cast the runtime can't check; narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast) | adaptation |
| 06_fileExists.a | NotYet | NotYet | NotYet | NotYet | node:fs.symlinkSync | library |
| 07_directoryExists.a | NotYet | NotYet | NotYet | NotYet | node:fs.symlinkSync | library |
| 08_getDirectories.a | NotYet | NotYet | NotYet | NotYet | node:fs.symlinkSync | library |
| 09_realpath.a | NotYet | NotYet | NotYet | NotYet | node:fs.symlinkSync | library |
| 10_getModifiedTime.a | NotYet | NotYet | NotYet | NotYet | node:fs.symlinkSync | library |
| 11_setModifiedTime.a | NotYet | NotYet | NotYet | NotYet | rmSync: fs option literals containing evaluated expressions; bind a plain options object first | library |
| 12_deleteFile.a | NotYet | NotYet | NotYet | NotYet | node:fs.symlinkSync | library |
| 13_createDirectory.a | Checker | Checker | Refused | Refused | in; an object's shape is known; use a discriminant, or a Map | compiler |
| 14_getCurrentDirectory.a | NotYet | NotYet | NotYet | NotYet | rmSync: fs option literals containing evaluated expressions; bind a plain options object first | library |
| 15_getExecutingFilePath.a | NotYet | NotYet | NotYet | NotYet | node:path.basename | library |
| 16_getEnvironmentVariable.a | NotYet | NotYet | NotYet | NotYet | rmSync: fs option literals containing evaluated expressions; bind a plain options object first | library |
| 17_write.a | NotYet | NotYet | NotYet | NotYet | rmSync: fs option literals containing evaluated expressions; bind a plain options object first | library |
| 18_exit_0.a | NotYet | NotYet | NotYet | NotYet | a value of type undefined | compiler |
| 19_exit_1.a | NotYet | NotYet | NotYet | NotYet | a value of type undefined | compiler |
| 20_exit_2.a | NotYet | NotYet | NotYet | NotYet | a value of type undefined | compiler |
| 21_createHash.a | NotYet | NotYet | NotYet | NotYet | node:crypto."node:crypto" read as a value | library |
| 22_createHash_fallback.a | NotYet | NotYet | NotYet | NotYet | a value of type undefined | compiler |
| 23_newLine.a | NotYet | NotYet | NotYet | NotYet | reading _os | library |
| 24_useCaseSensitiveFileNames.a | NotYet | NotYet | NotYet | NotYet | statSync: fs options other than a fixed literal or its plain const binding | library |
| 25_readDirectory.a | Checker | Checker | Checker | Checker | recorded TS2322/2345/2532/2775/7030 | adaptation |

Expected movers: adapted 01 now stops at array of never; adapted 02-04 at rmSync.
05 passes ToBoolean and reaches its unchecked catch-variable cast.
14 passes ! and reaches rmSync; 21 passes ! and reaches the crypto namespace read.
No row still refuses ! or a non-boolean condition.

The rmSync regression is a library/IR compatibility gap: the literal effect guard
accepts only raw scalar constants, but the compiler now wraps some boolean fields.
A fixed {recursive:true,force:true} object is incorrectly called an evaluated expression.
The runtime is unchanged. This also removes fixture 11's prior acceptance.

Exact diagnostic for the remaining expected refusal, fixture 05:
~~~text
adamic: /workspace/adamic/stage3/fixtures/host/05_writeFile.a:37:30: Adamic 0.1 refuses a cast the runtime can't check; narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast)
~~~

One-line reproducer actually run:
~~~typescript
import type { Stats } from 'node:fs'; try { throw new Error('x'); } catch (e) { console.log((e as NodeJS.ErrnoException).code); }
~~~
/tmp/fs-proof-repros/cast.log: adamic/no-unchecked-cast at 1:94.

One-line adapted 13 reproducer, actually run:
~~~typescript
function errorCode(e: unknown): string | undefined { return typeof e === 'object' && e !== null && 'code' in e && typeof e.code === 'string' ? e.code : undefined; }
~~~
/tmp/fs-proof-repros/in.log: Refused in at 1:107.

One-line library regression reproducer, actually run:
~~~typescript
import { rmSync } from 'node:fs'; rmSync('missing', { recursive: true, force: true });
~~~
/tmp/fs-proof-repros/rm.log: named rmSync NotYet at 1:35.

Commands and logs:
- go build -o /tmp/fs-proof-adamic ./cmd/adamic: PASS.
- python3 internal/oracle/node_fs_file_host_check.py --all --compiler /tmp/fs-proof-adamic
  --logs /tmp/fs-proof-original-final --report internal/oracle/node_fs_file_proof_original_status.json:
  /tmp/fs_proof_original_final.log, expected exit 1 for blockers.
- NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node
  internal/oracle/node_fs_file_host_47_check.cjs /tmp/fs-adaptation47-tsc
  /tmp/fs-proof-adamic /tmp/fs-proof-47-final internal/oracle/node_fs_file_proof_47_status.json:
  /tmp/fs_proof_47_final.log, expected exit 1; five Node-only semantic mutants caught.
  First attempt lacked TypeScript on NODE_PATH; the existing pinned installation fixed it.
- go test ./internal/lower ./internal/native ./internal/fresh ./internal/flow
  -run 'NonNull|Truth|ToBoolean|Deinitial|NodeFSFile' -count=1 -timeout 30m:
  /tmp/fs_proof_regressions_final.log. Native/flow have no matching tests.
- go test ./internal/oracle -run
  'TestNativeAgreesWithNode/internal/oracle/testdata/(non_null|truth|to_boolean)'
  -count=1 -timeout 30m: /tmp/fs_proof_feature_oracle_final.log, PASS 1.600s.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m
  -args -update-counts: /tmp/fs_proof_counts_final.log, FAIL 40.367s on fs option gaps.
  No partial table was generated and no counts pass is claimed.
No full repository gate, uncached integration gate or macOS run; checker options unchanged.

# 99-wasm: a qualified crypto namespace import

Task #redxes2, head 1.

The main base is `a0ba9dd52b23326880b16a7898d446597cfb5a88`. Both control and after observations
are fresh on that base; no previous control is reused. Upstream is TypeScript
6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`, with every current main
adaptation applied before this one. Platforms and Adamic approved the
JavaScript change on November 3.

## Purpose and exact edit

The original `_crypto` local can be undefined when requiring crypto fails.
The Node host's hash stores test that local, choosing SHA-256 or djb2. Adamic
does not admit a namespace used as a value, including those tests. This
adaptation makes the binding a qualified namespace import, whose one remaining
use is its `createHash` member:

```ts
import * as _fs from "node:fs";
import * as _path from "node:path";
import * as _os from "node:os";
import * as _crypto from "node:crypto";
```

Only the last import is added. The comment about reduced Node installations,
`let _crypto: typeof import("crypto") | undefined;` and its complete try/catch
are removed. The two stores become:

```ts
createHash: createSHA256Hash,
createSHA256Hash,
```

`_crypto!.createHash("sha256")` becomes `_crypto.createHash("sha256")`.
The shorthand typechecks against 98-e's property type; no explicit-property
fallback was needed. The interface remains byte-for-byte as main leaves it:

```ts
createSHA256Hash?: { createSHA256Hash(data: string): string; }["createSHA256Hash"] | undefined;
```

This is the intended JavaScript change: the built CommonJS compiler requires
crypto at module scope instead of inside getNodeSystem. On this Node, crypto
is present and both original conditional stores take their first arm. A Node
built without crypto now fails during module evaluation instead of falling
back to djb2. No other require, source-map-support, plugin load, inspector,
__filename or __dirname is changed. The djb2 function itself is untouched.
Fixture 21, status.json, the meter and owners.json are not part of this head.

## Order and the earlier reader

Main already has `94-axolotl-property-read`; the new directory is deliberately
`99-wasm-crypto-import`. It sorts after `99-correctness-axolotl-i`, last of all.
`apply.py:272` sorts the directories and its sole non-setup adapter
loop at lines 315-321 runs each once. There is no
post-adaptation second pass. The separate 00-setup diagnostic generation runs
before the pristine snapshot. The build path starts with a fresh pinned clone;
restoring a prepared product does not run adapters again. `order.json` records
the complete order and the AST-checked loop locations. No later reader of these
crypto expressions or automatic rerun of 98-e was found in the apply/lane path.

98-e is a reader of the line this adapter removes: its store guard at
`stage3/adapt/98-axolotl-e/adapt.cjs:141` requires exactly
`createSHA256Hash: _crypto ? createSHA256Hash : undefined`. It passes during
the full official apply, before this edit. Rerunning 98-e alone on this output
intentionally exits 1:

```text
the nodeSystem store: want one, got 0

0 !== 1
```

That is its input moving, not a bug in its guard. It writes no source on the
failed rerun; the complete source check passes afterwards. `earlier-reader.md`
retains the command and exact red. No earlier adapter was modified.

## Complete guards and source proof

The adapter copies 93's full-file hashes, validate-before-write discipline and
single conditional write. The before pin is the final official main sys.ts,
including 98-e and 99-i, rather than 93's intermediate snapshot:

- Before: `f9ff61b6a486c60e6738b334fe2e483ced5e53bfb2e462fadc4bdc9749d25336`.
- After: `a52755b55d8998df5fd5045fc0418c33bc7496502a131116a00d9bb4bc04b273`.

The checker resolves all five original references to the one local let: two
assignments, two conditions and the checked receiver. It rejects any other
identifier declaration named `_crypto` anywhere in sys.ts. Before the sole
possible write, a separate virtual program resolves every remaining use to
the namespace import: exactly one qualified property-access receiver, no bare
namespace value and no competing declaration. The SHA-256 call and its argument
are pinned too. A rerun reports zero files and preserves bytes and nanosecond
mtime. An added newline is rejected before writing, also preserving the
mutant's bytes and mtime.

`bindings.cjs` independently binds both trees and plants a module collision,
an enclosing parameter shadow and a bare namespace value. These negative
programs bypass the whole-file digest guard, so it cannot mask a broken binding
check. All three are caught. The separate retained-ternary mutant changes the
actual after source, invokes that same semantic guard without its digest gate,
and is rejected as `collision or bare namespace use: _crypto`. It is restored
in finally.

`TestCryptoSourceScope` inventories all 746 source files; only sys.ts changes.
Independently undoing the import, complete optional-load block and three hash
expressions restores the exact before bytes, including CRLF. Both comments-kept
and comments-removed transpile comparisons independently undo only the new
import, removed load block and two hash stores; the non-null assertion erases.
`source.diff` is an LF-normalized display, not the byte proof.

The erased-type scope mutant adds null to clearTimeout's private ambient
parameter type. It passes both JavaScript comparisons, then fails the exact
source undo proof. An unrelated runtime width fallback edit also fails scope.
The moved-target, three binding, retained-ternary and baseline-line negatives
all run independently. Every on-disk mutant is restored; each `.mutant.md`
names its command, exact red and catching check.

## Official table and upstream builds

The official apply exits 0 on both separate pristine trees. Its generated
reports are retained in proofs/; main's committed `stage3/patch-set.md` and all
lane plan files are untouched. No table was hand-merged.

| Tree or row | Files | Added | Removed |
| --- | ---: | ---: | ---: |
| Main control total | 95 | 6024 | 5731 |
| 99-wasm-crypto-import row | 1 | 4 | 11 |
| With adaptation total | 95 | 6028 | 5742 |

Both upstream `npm run build` commands exit 0: control
26.039 seconds, after 24.015 seconds.
Public `typescript.d.ts` is byte-identical: 593,502 bytes,
SHA-256 `c4edc2ad37cf1afeb4704ccb4bacf7c76b9f538a8e5685e4b276e2e7dd52796b`.

## Full serial oracles

These are two fresh full observations on current main, control first and after
second. They use 92's driver through the shared, unmodified runner: all runners,
no test filter, `--workers 8 --limit-seconds 1800`, a 2,400-second outer bound and
`NODE_OPTIONS=--max-old-space-size=3072`. No other build/test workload runs beside
either suite. The stock generated, untracked .parallelperf.json is reset first.
Before each suite, clean file-cache pages of finished task artifacts are released
without changing their bytes. Every cgroup memory-event counter must remain
unchanged, and the run stops below 3 GiB free disk. No guard is weakened.

| Oracle | Passing | Sanctioned API failure | Pending | Wall seconds |
| --- | ---: | ---: | ---: | ---: |
| Current main control | 106366 | 1 | 0 | 345.003 |
| With 99-wasm-crypto-import | 106366 | 1 | 0 | 347.139 |

The actual unmodified `stage3/lane/check.py` CLI prints this for each:

```text
PASS stage3 landing lane [causes: timeout=0, baseline-content=1, baseline-missing=0, exception=0, other=0]
```

Raw oracle exit is 1 only for main's sanctioned public API baseline test.
The current unchanged normalizer accepts 266 declarations.
The complete baseline.diff is byte-identical: 69,407 bytes,
SHA-256 `dade0a0d86d5f2ba9de4b812f1aa22e16a5c4d1121f488b1a31f7e30ad14f40f`. The public declaration bytes still match.
Appending an extra line makes the same real lane CLI fail its exact baseline
byte check. Restoring those bytes returns PASS, without another suite or a
changed sanction. Both accepted observations leave memory.events unchanged;
minimum free disk is 19.73 GiB. Both trees fit, so neither oracle tree had
to be deleted between observations. Afterwards, the entire source scope,
both transpile comparisons and a zero-write adapter rerun pass again.

## Built Node and incremental witnesses

Each real built tsc.js compiles the same small strict project. stdout, stderr,
exit code, emitted JavaScript and declarations are byte-identical. The project
is then built fresh with `--incremental` by each compiler. A common symlink and
Node's `--preserve-symlinks --preserve-symlinks-main` keep the compiler/lib paths
identical; only the target of that symlink changes between the two untouched
built trees. Outputs and metadata are removed before each run, so the second
cannot reuse the first. No metadata is normalized after compilation.
All recorded file versions are SHA-256, and raw .tsbuildinfo bytes match:
10,456 bytes, SHA-256 `5d36280fdfea10ea7d195d0049f98aa922398195aa38d67e8f7066a21b425f2b`.
This exercises sys.createHash, rather than just comparing a non-incremental emit.

Separate Node processes intercept Module._load while loading each actual built
typescript.js, then restore the loader. Main's crypto require has getNodeSystem
on its stack; the after require is at module scope with no getNodeSystem frame.
Upstream's CommonJS bundler emits the builtin specifier as bare `crypto`;
the source retains `node:crypto`. Full commands, stacks and compiler outputs
are in built-witness.json and its compressed raw log.

## Reproduction and retained evidence

Use Node v24.19.0, the pinned stage3/api install, the usual go-run refusal wrapper
and a TypeScript checkout at the source pin. Set ADAMIC_TYPESCRIPT_SOURCE to it,
NODE_PATH to stage3/api/node_modules, and CRYPTO_PROOF_SCRATCH to a fresh directory
(default /workspace/scratch/wasm-crypto-proof). Install stage3/api with its lockfile.
Use `proofs/control.py` to make a detached main source worktree at the recorded
base and run its official apply into SCRATCH/control, recording bounded commands.
Then run `proofs/prepare.py`: it pins the independently planned before/after
bytes, runs the complete official after apply, guards, source and transpile
checks, mutants and both upstream builds. It never copies the generated table
over main's committed report.

Run `node proofs/built-witness.cjs CONTROL AFTER PROJECT OUTPUT_JSON`, then
`proofs/oracles.py` on a quiet box, followed by `proofs/archive.py`. Paths above
are relative to this adaptation directory. commands.json retains exact argv,
limits, exits and times for the accepted preparation and built witness; oracle
JSON retains every phase and the complete verdict. The evidence is fresh, not
cached test success. All raw logs are gzip-compressed under proofs/logs, as in
92 and 93. The adapter directory is the entire repository diff; no shared lane,
oracle, fixture, meter, owners map or committed patch table is edited.

# 93: qualified builtin namespace imports

Task #he8tysb, head 1 revision. This is a new commit after
bfe98353b7e58dabcaa642b0f9c1f9c007944b48, never an amendment or rebase.
The branch's main base is 7b9c0fe7e52cfe9cdeb774107389bd6870f41423.
Current main was fetched at c4fe7a014b6ff3fc7abf77c69ac29ab7ea04477c.
All stage3 inputs are byte-identical between those main commits. Nevertheless,
because main's tip moved, the old control observation is not reused: this
revision reapplies current main, rebuilds both trees and runs both full oracles.
Upstream remains TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

## Purpose and exact edit

The first revision bound namespace imports to locals, which Adamic refuses as
a namespace object used as a value. This revision makes the original names
module-scope namespace imports instead:

```ts
import * as _fs from "node:fs";
import * as _path from "node:path";
import * as _os from "node:os";
```

The complete three const declarations in getNodeSystem are removed. Every
remaining occurrence is a qualified namespace receiver; there is no namespace
value assignment or bare read. There are 27 qualified AST accesses across 22
source lines: 19 _fs, six _path and two _os. No access expression is edited.
The difference from the previous wording's 22 uses is counting expressions
rather than source lines with multiple accesses.

This remains the JavaScript-changing adaptation decided by Adamic: the built
CommonJS compiler loads these three builtins eagerly at module evaluation.
Crypto, source-map-support, plugin require, inspector, __filename and __dirname
are untouched. Only sys.ts changes in the adapted source tree.

## Complete guards and binding proof

The adapter retains 92's complete before/after hashes, validate-before-write
discipline and sole conditional write. targets.json pins both entire sys.ts
byte sequences. A rerun reports zero files and preserves its timestamp; an
added newline is rejected before writing, preserving the mutant's exact bytes
and timestamp.

Before editing, the TypeScript checker must resolve every qualified access to
its original local. All other identifiers with each of these names are rejected,
so there cannot be another declaration at module scope or in any scope that
encloses a use. The check is deliberately stronger: it rejects any competing
declaration anywhere in sys.ts. Every nonbinding identifier must be a property
access receiver, and its checker symbol must be exactly the expected binding.
The counts are pinned at 19/6/2 and 22 distinct lines.

Before writing, a separate virtual program containing the planned bytes must
resolve all 27 accesses to the three new import bindings, with no locals left.
The same checks apply on a rerun. proofs/bindings.cjs independently runs the
before/after census and directly exercises the semantic guard against a module
collision, an enclosing function-parameter shadow and a bare namespace value.
These probes bypass the file-hash guard, so its rejection cannot mask a broken
binding check. All three must be caught; bindings.json records exact messages.

TestBuiltinSourceScope inventories all 746 source files and finds only sys.ts
changed. Independently removing the three imports and reinstating the three
complete declarations restores the exact original bytes. Both comments-kept
and comments-removed transpile comparisons independently undo only the three
imports and removed declarations. source.diff is an LF-normalized display;
target hashes and the undo check preserve the exact CRLF source bytes.

The erased-type scope mutant still passes both JavaScript comparisons but fails
the exact-source scope check. It is restored in finally. See proofs/scope.mutant.md.
The moved-target and extra runtime-edit negatives are also rerun and restored.

## Official table and builds

Both trees are separate pristine clones with their full official numbered
application; neither is manually merged. The control applicator is current
main's unchanged stage3/apply.py. The generated after patch-set.md is copied
byte for byte to stage3/patch-set.md and proofs/patch-set.md, as in 92.

The official row is `93-builtin-imports | 1 | 3 | 3`. Without 93 the totals are
85 files, 5,526 added / 5,484 removed; with this revision they are 85 files,
5,529 added / 5,487 removed. The current main table's stale adaptation-40 row
is also refreshed by official generation, never by hand merge.

Both upstream npm run build commands exit 0: control 24.023
seconds, revision 24.011 seconds. Public typescript.d.ts bytes
match exactly: 590,996 bytes, SHA-256 `b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`.

## Full serial oracles

Both observations below are new for this revision. No previous control result
is reused. The unmodified shared runner uses all runners, no filter,
--workers 8 --limit-seconds 1800, with 92's 2,400-second outer bound and
NODE_OPTIONS=--max-old-space-size=3072. No other build/test workload runs beside
either suite. The stock generated .parallelperf.json is reset if present.
Disk is monitored throughout and the driver stops below 3 GiB free.

| Oracle | Passing | Sanctioned API failure | Pending | Wall seconds |
| --- | ---: | ---: | ---: | ---: |
| Current main control | 106,366 | 1 | 0 | 344.928 |
| With revised 93 | 106,366 | 1 | 0 | 339.16 |

Both actual unmodified stage3/lane/check.py CLI invocations print PASS stage3
landing lane. The only raw failure is main's sanctioned public API test; its
unchanged normalizer accepts the same 223 declarations. No expected count,
API sanction, fixture or baseline is updated. Wall time follows 92's driver
measurement; report.json retains every raw phase timing.

The complete baseline.diff is byte-identical: 51,123 bytes,
SHA-256 `6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`. Public typescript.d.ts is
also byte-identical, with the size and hash above. Both accepted runs leave
all memory-event counters unchanged. Both trees fit; minimum free disk in the
accepted observations is 7.56 GiB.

The first fresh control also passed the lane with 106,366 / 1 / 0 in
315.005 seconds, but 92's copied memory guard rejected it: memory.events
max rose from 1,016 to 18,090, with no OOM or kill. Its complete observation is
retained as control-attempt1. The previous job's completed adapted control and
after trees were removed; their logs and hashes remain in the preceding commit.
The fresh serial pair was then repeated with the same worker count, heap setting,
packing-cache reset, limits and strict memory guard. No guard was weakened.

The extra baseline-line mutant still fails the same lane byte checker and is
restored exactly. See proofs/oracle.mutant.md. After both suites, exact source scope,
both transpile comparisons and a zero-file adapter rerun pass again.

## Built Node witness

The final built compilers run the same small strict project through their real
tsc.js. stdout, stderr, exit status, emitted JavaScript and emitted declaration
bytes must match. A separate node -e process intercepts Module._load only while
requiring each real built typescript.js: main's builtin loads have getNodeSystem
on the stack; the revision's loads are at module scope with no getNodeSystem
frame. Upstream esbuild normalizes node: to bare builtin require specifiers in
the CommonJS bundle. The source imports retain node:. The witness records
actual stacks, restores the loader, and never rewrites the built compiler.

## Reproduction and retained evidence

Use Node v24.19.0 and the repository's pinned stage3/api install. Set
BUILTIN_PROOF_SCRATCH to a fresh directory (default
/workspace/scratch/adapt-93-revision-proof). Prepare upstream at the pin and
control by current main's official applicator, excluding only adaptation 93.
Run proofs/prepare.py for the fresh after apply, guards, scope, transpile and
binding proofs, mutants and both upstream builds. commands.json records exact
arguments, limits, exits and times. Run scope-mutant.cjs CONTROL AFTER (expected
exit 1), then restored check.cjs and built-witness.cjs CONTROL AFTER PROJECT
OUTPUT_JSON. Run proofs/oracles.py on a quiet box, then proofs/archive.py.
NODE_PATH must name stage3/api/node_modules. No shared oracle or lane source
changes. Compressed raw logs follow 92's layout under proofs/logs.

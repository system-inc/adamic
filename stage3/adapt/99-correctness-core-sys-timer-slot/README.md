# System timer slot: C-noinfer-b

Adamic, task b5w3ycg, Nov 7 21:47:55Z. This head is held for Runtime
#995spev and stacks on the adaptation 41 ambient-head amendment. It uses
TypeScript 6.0.3, source `050880ce59e30b356b686bd3144efe24f875ebc8`, on the
pinned main `0cb816a0ec439f63542b3f83bb13941fd79f8f41`.

```
setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;
clearTimeout?(timeoutId: unknown): void;
```

The two private `pollScheduled` annotations become `unknown`. The measured
inventory has 22 calls and no member read on a returned handle. `NoInfer`
lets the actual extra arguments choose A while callbacks may ignore a suffix,
including the two label-only calls. The adapted typescript.d.ts needs
TypeScript 5.4 or later to read because of `NoInfer`. This is a declared public
API difference, recorded in api-sanctions.json and the upstream ledger.

## Owner contract

owners.json pins the whole System interface and both complete polling
functions before and after, with exact text and SHA-256. All three owners
must be uniformly before or after. The adapter plans and checks the complete
file before writing; a half-applied tree, a partial member change, changed,
missing, duplicate or moved owner, or parse error refuses without writes.
Replay is idempotent and preserves mtimes. LF and CRLF are preserved. Unrelated
source outside the owners is accepted. Full-file standalone JavaScript is
checked equal with comments kept and removed. No runtime statement changes.

## Census gate and the newly accepted initializer

All coordinates below use original fully adapted main. The two deleted
ambient lines shift actual head coordinates down two after line 57.
Head 1 already exposes `reading setTimeout` at main 1501:13, actual 1499:13,
owned by Runtime #995spev. Against Head 1, C-noinfer-b exposes eight sites;
against original main, it exposes nine including that existing read. All are
expected compiler walls, not a claim that native lowering succeeds.

The eighth additional site, main 1519:36 (actual 1517:36), says exactly:

`a member typed any (require()module) related to {} at the initializer site`

It names `const nodeSystem: System = { ... }`. Its `require` member returns
`{ module: require(modulePath), modulePath, error: undefined }`, where
`modulePath = resolveJSModule(moduleName, baseDir, nodeSystem)`. The `any` is
Node's `require(modulePath)` result in that object, not the separate
`require("source-map-support")` or `require("inspector")` calls. System's
`require?(baseDir: string, moduleName: string): ModuleImportResult` uses
`ModuleImportResult<T = {}>`; its success member is `module: T`. Thus the
nested `module` result is related to `{}`. The stock 6.0.3 checker confirms
`module: require(modulePath)` at actual 1613:38 has value type `any` and
contextual type `{}`; head2-stock.json retains that query.

The original slot, C-noinfer-b, and the any-slot-restoration mutant report the
same recording unit. Its deferred relation changes from main 384:79's callback
`any` relation to this initializer relation and back. No initializer finding
is recorded in Head 1 or the mutant. `internal/lower/any_relations.go:24-48`
keeps only the first deferred any relation (`if l.refusal.deferred == nil`);
line 43 constructs the quoted message. Inference, supported by the measured
replacement and reversal: the slot's earlier `any` relation masked the later
initializer relation by occupying that deferred slot. The census does not
trace every intermediate relation candidate, so this is not presented as a
complete scan trace.

The gate requires exactly the approved location, kind, reason and phase set,
including that initializer. Any other new finding stops the proof. The
any-slot-restoration mutant keeps the two private annotations as unknown and
must restore main 384:79.

## Reproduction and receipts

Source `/workspace/adamic-tools/env.sh`, install the pinned stage3/api parser,
and export ADAMIC_TYPESCRIPT_SOURCE to the pinned upstream checkout. Run
`bash stage3/apply.sh --build-product NEW_TREE` from a pristine checkout.
`node contract.cjs HEAD1_TREE` exercises owner refusal and replay. The retained
proof drivers log each child's actual exit and bound each piece at 30 minutes.
The full lane executes the entire unchanged manifest in separate units,
validating test identities, multiplicities, counts and exact sanctioned API
changes. No expected failure, timeout or test inventory is relaxed.

The accepted storage survey found no direct executable rule pinning either
private annotation; storage-pin-survey.json retains its scope and limitations.
See [PROOF.md](PROOF.md) for the complete build, lane and census receipts.

## House API generator correction

The earlier 915a67cf receipt used the raw cjs helper output. The corrected head
uses the documented writer, which removes original/expected and retains the
pretty house schema with baseline_diffs and diff_lines:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" \
  python3 stage3/lane/generate-api.py "$ADAMIC_TYPESCRIPT_SOURCE" \
  > /tmp/lane-generate.log 2>&1
```

The following complete diff is against committed Head 1. In addition to the
two timer members and their owner/proof rows, it includes two pre-existing
proof-hash refreshes. An isolated regeneration on Head 1 reproduces exactly
those two hash refreshes, with no timer adaptation present. Their inclusion
is approved: the house generator output is authoritative. Against regenerated
Head 1, existing normalized rows, owners and proofs are unchanged; additions
belong only to System timers. Counts and baseline_diffs stay unchanged. The patch-line delta contains
only the two before/after timer declarations. delta.json records these checks.

```diff
--- head1/stage3/lane/sanctioned-api.json
+++ head2/stage3/lane/sanctioned-api.json
@@ -1117,6 +1117,11 @@
       "after": "[\"PropertySignature\",[\"ReadonlyKeyword\"],[\"Identifier\",\"tupleNameSource\"],[\"QuestionToken\"],[\"UnionType\",[\"TypeReference\",[\"Identifier\",\"ParameterDeclaration\"]],[\"TypeReference\",[\"Identifier\",\"NamedTupleMember\"]],[\"UndefinedKeyword\"]]]"
     },
     {
+      "declaration": "member:[\"ts\",\"System\",\"clearTimeout\"]#0",
+      "before": "[\"MethodSignature\",[\"Identifier\",\"clearTimeout\"],[\"QuestionToken\"],[\"Parameter\",[\"Identifier\",\"timeoutId\"],[\"AnyKeyword\"]],[\"VoidKeyword\"]]",
+      "after": "[\"MethodSignature\",[\"Identifier\",\"clearTimeout\"],[\"QuestionToken\"],[\"Parameter\",[\"Identifier\",\"timeoutId\"],[\"UnknownKeyword\"]],[\"VoidKeyword\"]]"
+    },
+    {
       "declaration": "member:[\"ts\",\"System\",\"createSHA256Hash\"]#0",
       "before": "[\"MethodSignature\",[\"Identifier\",\"createSHA256Hash\"],[\"QuestionToken\"],[\"Parameter\",[\"Identifier\",\"data\"],[\"StringKeyword\"]],[\"StringKeyword\"]]",
       "after": "[\"PropertySignature\",[\"Identifier\",\"createSHA256Hash\"],[\"QuestionToken\"],[\"UnionType\",[\"IndexedAccessType\",[\"TypeLiteral\",[\"MethodSignature\",[\"Identifier\",\"createSHA256Hash\"],[\"Parameter\",[\"Identifier\",\"data\"],[\"StringKeyword\"]],[\"StringKeyword\"]]],[\"LiteralType\",[\"StringLiteral\",\"createSHA256Hash\"]]],[\"UndefinedKeyword\"]]]"
@@ -1125,6 +1130,11 @@
       "declaration": "member:[\"ts\",\"System\",\"exit\"]#0",
       "before": "[\"MethodSignature\",[\"Identifier\",\"exit\"],[\"Parameter\",[\"Identifier\",\"exitCode\"],[\"QuestionToken\"],[\"NumberKeyword\"]],[\"VoidKeyword\"]]",
       "after": "[\"MethodSignature\",[\"Identifier\",\"exit\"],[\"Parameter\",[\"Identifier\",\"exitCode\"],[\"QuestionToken\"],[\"NumberKeyword\"]],[\"UnionType\",[\"NumberKeyword\"],[\"UndefinedKeyword\"]]]"
+    },
+    {
+      "declaration": "member:[\"ts\",\"System\",\"setTimeout\"]#0",
+      "before": "[\"MethodSignature\",[\"Identifier\",\"setTimeout\"],[\"QuestionToken\"],[\"Parameter\",[\"Identifier\",\"callback\"],[\"FunctionType\",[\"Parameter\",[\"DotDotDotToken\"],[\"Identifier\",\"args\"],[\"ArrayType\",[\"AnyKeyword\"]]],[\"VoidKeyword\"]]],[\"Parameter\",[\"Identifier\",\"ms\"],[\"NumberKeyword\"]],[\"Parameter\",[\"DotDotDotToken\"],[\"Identifier\",\"args\"],[\"ArrayType\",[\"AnyKeyword\"]]],[\"AnyKeyword\"]]",
+      "after": "[\"MethodSignature\",[\"Identifier\",\"setTimeout\"],[\"QuestionToken\"],[\"TypeParameter\",[\"Identifier\",\"A\"],[\"ArrayType\",[\"UnknownKeyword\"]]],[\"Parameter\",[\"Identifier\",\"callback\"],[\"FunctionType\",[\"Parameter\",[\"DotDotDotToken\"],[\"Identifier\",\"args\"],[\"TypeReference\",[\"Identifier\",\"NoInfer\"],[\"TypeReference\",[\"Identifier\",\"A\"]]]],[\"VoidKeyword\"]]],[\"Parameter\",[\"Identifier\",\"ms\"],[\"NumberKeyword\"]],[\"Parameter\",[\"DotDotDotToken\"],[\"Identifier\",\"args\"],[\"TypeReference\",[\"Identifier\",\"A\"]]],[\"UnknownKeyword\"]]"
     },
     {
       "declaration": "member:[\"ts\",\"TaggedTemplateExpression\",\"typeArguments\"]#0",
@@ -3469,6 +3479,22 @@
       "line": 6987
     },
     {
+      "adaptation": "adapt/99-correctness-core-sys-timer-slot/api-sanctions.json row 1 {\"path\":[\"System\",\"setTimeout\"],\"kind\":\"MethodSignature\",\"replace\":\"member\",\"before\":\"setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;\",\"after\":\"setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;\",\"why\":\"Adamic ruling, task b5w3ycg, Nov 7 21:47:55Z, C-noinfer-b: the adapted typescript.d.ts now needs TypeScript 5.4 or later to read because of NoInfer.\"}",
+      "path": [
+        "System",
+        "setTimeout"
+      ],
+      "line": 8491
+    },
+    {
+      "adaptation": "adapt/99-correctness-core-sys-timer-slot/api-sanctions.json row 2 {\"path\":[\"System\",\"clearTimeout\"],\"kind\":\"MethodSignature\",\"replace\":\"member\",\"before\":\"clearTimeout?(timeoutId: any): void;\",\"after\":\"clearTimeout?(timeoutId: unknown): void;\",\"why\":\"Adamic ruling, task b5w3ycg, Nov 7 21:47:55Z, C-noinfer-b: the adapted typescript.d.ts now needs TypeScript 5.4 or later to read because of NoInfer.\"}",
+      "path": [
+        "System",
+        "clearTimeout"
+      ],
+      "line": 8492
+    },
+    {
       "adaptation": "75",
       "path": [
         "JsonSourceFile",
@@ -3484,7 +3510,7 @@
     },
     {
       "file": "stage3/adapt/40-explicit-any/class-rules.json",
-      "sha256": "0879b71c393b0aac9c28676df002b4539a4d7b7d5d9158a167c98e821a8f7567"
+      "sha256": "215ecf3a168203ecdbd24949fecaaf7f8ebb8d9d456ae884c532673b51b2d576"
     },
     {
       "file": "stage3/adapt/75-optional-widening/evidence/wave2/api.json",
@@ -3504,7 +3530,7 @@
     },
     {
       "file": "stage3/adapt/32-indexed-reads-program/public-host-sites.json",
-      "sha256": "775c2ae6d7d1f784ca2d249dd277737526f7bac44de5c368c98db9cf0bd1ae2b"
+      "sha256": "1f641e9402f9cccda75e52209fea74f827e40fe21103ce743062c943ef170776"
     },
     {
       "file": "stage3/adapt/32-indexed-reads-program/resumed-watch/api.json",
@@ -3585,6 +3611,10 @@
     {
       "file": "stage3/adapt/99-correctness-axolotl-c/api-sanctions.json",
       "sha256": "00373a533e7b6436402c35f67d6b7148a7ade064b453cc5cb40950b743156432"
+    },
+    {
+      "file": "stage3/adapt/99-correctness-core-sys-timer-slot/api-sanctions.json",
+      "sha256": "5f0bf9920772e9b8c59350c29c13f3d24775c9a1a9473511d20516c8043fae20"
     }
   ],
   "baseline_diffs": [
@@ -4886,7 +4916,7 @@
     "         readonly allowTextChangesInNewFiles?: boolean;\n",
     "         readonly providePrefixAndSuffixTextForRename?: boolean;\n",
     "         readonly includePackageJsonAutoImports?: \"auto\" | \"on\" | \"off\";\n",
-    "@@ -8484,9 +8482,9 @@\n",
+    "@@ -8484,12 +8482,12 @@\n",
     "          */\n",
     "         createHash?(data: string): string;\n",
     "         /** This must be cryptographically secure. Only implement this method using `crypto.createHash(\"sha256\")`. */\n",
@@ -4896,8 +4926,13 @@
     "-        exit(exitCode?: number): void;\n",
     "+        exit(exitCode?: number): number | undefined;\n",
     "         realpath?(path: string): string;\n",
-    "         setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;\n",
-    "         clearTimeout?(timeoutId: any): void;\n",
+    "-        setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;\n",
+    "-        clearTimeout?(timeoutId: any): void;\n",
+    "+        setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;\n",
+    "+        clearTimeout?(timeoutId: unknown): void;\n",
+    "         clearScreen?(): void;\n",
+    "         base64decode?(input: string): string;\n",
+    "         base64encode?(input: string): string;\n",
     "@@ -8519,7 +8517,7 @@\n",
     "     function isIdentifierStart(ch: number, languageVersion: ScriptTarget | undefined): boolean;\n",
     "     function isIdentifierPart(ch: number, languageVersion: ScriptTarget | undefined, identifierVariant?: LanguageVariant): boolean;\n",
```

### Correction checks and approved hash refreshes

| Check through the Go wrappers | Final child exit |
|---|---:|
| test_generate_api: all eight groups plus coverage test | 0 |
| test_check: TestLaneCheckPython | 0 |
| test_shards: TestLaneShardsPython | 0 |
| go test ./stage3/lane/... whole, unfiltered | 0 |
| generate-api.py, default output | 0 |
| generate_shards.py --check with refreshed runtime | 0 |

The first checker run exited 1 because its fixed declaration count was 266.
The two timer sanctions make 268; the existing test now asserts both new member
keys and the exact total 268. No guard was relaxed. The first whole run exited
1 when its two-worker schedule exhausted a 25-minute package budget. The
unfiltered retry used three workers and a 29-minute package bound, below the
30-minute piece limit; wrapper deadlines and expectations were unchanged.
All 398 shard wrappers passed, as did the remaining fixture and product tests.

The environment restarted while the retry ran. The Go process survived, but
its logging parent did not. Its own kernel wait status was recovered from
/proc/58100/stat field 52: 0, decoded exit 0. The raw stat and recovery record
are retained in generator-fix. The observed elapsed upper bound is 1,420
seconds; package durations remain in the log. No exit was inferred from PASS
text. results.json records every requested command and the initial failures.

shard.py's api_check reads only source_commit and normalized_changes from this
file. Both are identical to 915a67cf, so the correction preserves its approval
decision. In addition, TestShard_Stage3LaneByteComparisons reran successfully in
the whole Go suite against this corrected file. The generated runtime binding
was refreshed before these runs. tested-inputs.json verifies the executed
adapter, owner, sanction, manifest, plan and checker-test bytes stayed unchanged
through receipt collection.

The two generated proof-hash refreshes are approved for inclusion. They were
already stale on main, not caused by this head. No output hash was hand-edited
to hide that drift. The pending labels in the retained generator-fix receipts
record the earlier review pause; this README records its resolution.

### Main history and the missed freshness check

These heads use main at `0cb816a0ec439f63542b3f83bb13941fd79f8f41`.
Comparing `git log` for each proof file with the sanction file shows:

| File | Last change on that main | Commit time (UTC) |
|---|---|---|
| stage3/lane/sanctioned-api.json | `e3d1fb506adff11cb66ad7b1e54222ce944666e6` | 2026-11-03 09:19:14 |
| stage3/adapt/40-explicit-any/class-rules.json | `01813e07a5b062baa5cabd9b5fb7e4ad6c983382` | 2026-11-07 00:10:03 |
| stage3/adapt/32-indexed-reads-program/public-host-sites.json | `0bf7852d140f9e2d081cd2ff9485725187dd1a7e` | 2026-11-06 20:33:54 |

The last sanction-file update, e3d1fb506, is the 98-axolotl-e API update.
The preceding explicitly titled regeneration was
`e977285dcfacb1b494fb43b46dae0dfcaa951b29`, "Regenerate lane API sanction for
99c, and the shard plan's runtime hash that pins it", at 2026-11-03 03:43:03 UTC.
Both proof edits came after these sanction updates and are ancestors of the
pinned main. 01813e07 types adaptation 40's debug enum views as readonly records;
0bf7852d folds adaptation 32's watch.ts maybeBind calls into optional bind calls.
Neither later proof edit refreshed the sanction file.

At e3d1fb506, the two stored hashes match those files' bytes: class-rules.json
is 0879b71c... and public-host-sites.json is 775c2ae6.... On the pinned main,
the files instead hash to 215ecf3a... and 1f641e94..., while the stored hashes
remain old. The isolated Head 1 regeneration reproduces exactly these two
refreshes. Thus this is pre-existing main drift exposed by regeneration, not
a timer change or a newly introduced API declaration difference.

The history queries are reproducible with this pin and each path above:

```sh
git log -1 --format='%H %cI %s' 0cb816a0ec439f63542b3f83bb13941fd79f8f41 -- stage3/lane/sanctioned-api.json
git log -1 --format='%H %cI %s' 0cb816a0ec439f63542b3f83bb13941fd79f8f41 -- stage3/adapt/40-explicit-any/class-rules.json
git log -1 --format='%H %cI %s' 0cb816a0ec439f63542b3f83bb13941fd79f8f41 -- stage3/adapt/32-indexed-reads-program/public-host-sites.json
```

A production-manifest freshness assertion in test_generate_api should have
caught this, but no such assertion exists. What the suite actually compares:

- ApiSanctionGeneratorTests.setUp reads only the checked-in manifest's proof
  paths, ignores their stored sha256 values, and symlinks the current files
  into a scratch stage. It then adds its synthetic zz-sanction-fixture row.
- generate runs the cjs helper on that scratch stage. generated checks exit 0
  and verifies that the synthetic fixture's newly computed SHA-256 appears in
  the new proof list. It does not compare the production proofs' newly computed
  hashes with their checked-in values.
- Positive cases check the requested fixture declaration in expected and pass
  fresh generated original/expected snapshots and that fresh proof to api_check.
  Negative cases undo or corrupt fixture declarations or owners and require
  rejection. These check reconstruction and guard behavior, not whether the
  checked-in production manifest is current.

Consequently, stale hashes on existing listed proof paths pass: fresh bytes
are read, but compared only within the fresh fixture result, not against the
old production hash. The suite also cannot discover an unlisted proof solely
by staging the manifest's existing list. The missing check should regenerate
from the actual production tree with generate-api.py into a temporary output
and compare the complete result with sanctioned-api.json, including proofs.
No test is added in this README-only follow-up.

api_check itself compares the source pin and normalized declarations, not
proof hashes. test_shards.test_plan_runtime_is_this_trees_runtime_hash binds
the committed sanction file's bytes to the plan; changing a proof input without
regenerating that file changes neither of those bytes nor that runtime hash.
Those existing checks therefore do not detect this provenance drift either.

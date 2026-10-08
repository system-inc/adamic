# Scanner generic returns and computed names

The two topic merges unblock the existing concrete callback reduction and the
scanner's computed constructor key. Each matches source Node with ordinary and
split native compilation. The unchanged upstream generic bodies still stop at
other checks. No full scanner native proof is claimed.

Input is TypeScript 6.0.3, commit
050880ce59e30b356b686bd3144efe24f875ebc8, with existing adaptations 00 and 10.
The scanner closure is gathered without temporary scanner adaptations, using
createScanner, ScriptTarget and SyntaxKind as its entries. It preserves the
78-module evaluation graph: 89 code declarations in eight files. Stock
TypeScript 6.0.3's checker audits every module, including the empty evaluation
scaffolds. sites.json records all coordinates and the closure summary.

| Exact construct | Fresh slice | Pinned upstream | Witness |
| --- | --- | --- | --- |
| U or undefined return | core.ts:12 | core.ts:33 | 01-core-forEach.a |
| U or undefined return | utilities.ts:12 | utilities.ts:746 | 02-utilities-forEachEntry.a |
| Computed property name | scanner.ts:113:5 | scanner.ts:150:5 | 04-scanner-constructor.a |

These are all two U-or-undefined function return signatures and the one computed
property name in this closure. The historical utilities.ts:13:17 coordinate
comes from a different import prelude in the discovery slice. The fresh slice's
function name is on line 12; its original upstream location is line 746.
This is an exhaustive structural/checker type audit, not a census of live
lowering failures behind other stops. It does not silently remove earlier stops.

The audit also lists nine related signatures with different return reasons:
core.ts:51 (identity, upstream 1828), core.ts:77 (getSpellingSuggestion, upstream
2169), Scanner's lookAhead/scanRange/tryScan signatures at scanner.ts:84/88/94
(upstream 121/125/131), and scanner implementation functions at 3376/3398/3423/
3427 (upstream 3930/3952/3977/3981). Those return T or T | undefined, rather than
the dispatched U | undefined reason, and are not claimed as tested witnesses.
Append's generic array returns have a different representation and are outside
this exact-reason audit.

## Scratch compiler inputs

Main is 45487a809f89885a3fc651cd590e7dabf31362dc. Separate detached worktrees
merge main with each topic using git merge, without source conflict resolution:

- Generic topic codex/notyet-generic-returns-t-topic,
  6f2c92fc10882b1064222d31642468766b2976ef; scratch merge
  86291ef2962652c577c09d2940a7adf19543aadb.
- Computed-name topic codex/notyet-destructuring-topic, found among origin's
  destructuring branches, d516d4e1dcc960e8a9373e161abd3adb94bb7c6b;
  scratch merge 5cb2461da86fc58b74602e30ae950f1f12a09104.

Neither merge is included in this scout branch or pushed. Each worktree uses
the pinned main cohere dependency through a shared symlink. Build with
`go build -buildvcs=false -o COMPILER ./cmd/adamic` from that worktree; VCS
stamping is disabled because git status rejects a symlink at the submodule
path. Main's binary is built independently as a control.

## Witnesses and results

01 and 02 preserve the complete upstream forEach and forEachEntry bodies,
respectively. Only standalone calls are added. 03 is the existing native3
standalone each reduction, with an empty-map call added. It substitutes an
explicit undefined comparison for upstream's truthiness condition and iterates
the map directly. It is not claimed equivalent for falsy callback values.
04 reduces textToKeywordObj to its computed constructor field, replacing
SyntaxKind.ConstructorKeyword with its stock numeric value, 137, and observes
its own key through Object.keys.

All source Node runs exit zero with empty stderr. Expected output bytes are in
manifest.json and are asserted before any native comparison.

| Witness | Node stdout, line by line | Main | Main plus topic, both native modes |
| --- | --- | --- | --- |
| 01 forEach | hit; missing | TS2345 | TS2345 at array[i]: T or undefined is not T |
| 02 forEachEntry | y; missing | NotYet returning U or undefined | Refused: string as a condition |
| 03 each reduction | x; missing | NotYet returning U or undefined | Matches Node; ASan/UBSan clean |
| 04 constructor key | constructor | NotYet computed field name | Matches Node; ASan/UBSan clean |

Both modes mean ADAMIC_NATIVE_SPLIT=0 and ADAMIC_NATIVE_SPLIT=1, with five jobs.
Four successful native comparisons execute binaries built with --sanitize,
ASAN_OPTIONS=detect_leaks=1 and UBSAN_OPTIONS=halt_on_error=1. Output, stderr and
exit status are compared exactly. Failed builds never count as matches.

Compiler findings: signature support is demonstrated by a concrete call, not
by an uninstantiated generic declaration. forEachEntry progresses to its real
truthiness refusal. forEach cannot establish generic lowering here because its
unchecked indexed read is rejected first. The computed-key topic preserves
constructor as an own enumerable name. An exploratory `keywords.constructor`
read was typed as inherited Function and rejected with TS2345, despite Node
reading the numeric own field. That read is not in the final key-only witness;
no value-read support or complete keyword table execution is claimed.

## Mutants

Each source mutant is generated in the scratch results directory:

- 01 replaces the early `return result` with `return undefined`: Node changes
  hit to missing. The fixed original Node stdout check catches it. Native is
  unavailable because of the original indexed-read error.
- 02 makes the same replacement: Node changes y to missing, caught by the
  original stdout check. Native is unavailable because of the original refusal.
- 03 makes the same replacement: Node and both sanitized native binaries print
  missing instead of x; the original Node comparison rejects all three.
- 04 renames the computed key to wrong: Node and both sanitized native binaries
  print wrong instead of constructor; the original Node comparison rejects all.

All four mutants execute successfully on Node; the two lowerable mutants also
build and execute successfully in both native modes. No compilation failure,
clang error or sanitizer error is credited as a semantic mutant kill. These
are source semantic mutants, not changes to production compiler rules.

## Commands and evidence

Every build and execution writes stdout/stderr to files, without output pipes.
Commands run, after sourcing /workspace/adamic-tools/env.sh:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step12-setup.log 2>&1
SLICE_TYPESCRIPT=/workspace/type-imports/npm/node_modules/typescript/lib/typescript.js \
  bash stage3/slice/run.sh /workspace/type-imports/generator-adapted \
  /workspace/step12-slice src/compiler/scanner.ts:createScanner \
  src/compiler/types.ts:ScriptTarget src/compiler/types.ts:SyntaxKind \
  --no-adapt > /tmp/step12-slice.log 2>&1
SLICE_TYPESCRIPT=/workspace/type-imports/npm/node_modules/typescript/lib/typescript.js \
  node stage3/scouts/step12/generics-names/sites.cjs /workspace/step12-slice \
  /workspace/type-imports/pristine > stage3/scouts/step12/generics-names/sites.json
python3 stage3/scouts/step12/generics-names/run.py \
  --generic /tmp/step12-generic --names /tmp/step12-names \
  --out /workspace/step12-proof > /tmp/step12-proof.log 2>&1
```

The focused run passes: four Node witnesses, four Node mutants, four native
matches and four native mutants. evidence/report.json contains exact statuses,
stdout and diagnostics; evidence/main-report.json records the four main control
builds. Compressed setup, merge, build, slice and run logs are retained.
The focused evidence audit also passes; deleting either one generic-return site
or the computed-name site from a copy of sites.json is caught by its exact-count
assertions (evidence/audit.log.gz).
counts.md describes this standalone corpus; no internal oracle registry was
changed. evidence/witness-sha256.json pins all four authored .a files.

Setup passed: Go .023s, Node .027s, submodules .067s, markdown .079s, clang .169s,
build 8.848s, cache 8.959s, total 8.988s. nproc=5; cgroup quota is four CPUs.
No full packages or full gate ran. No whole-scanner native, generated JavaScript,
release ownership-counter, or unrelated generic-return proof is claimed.

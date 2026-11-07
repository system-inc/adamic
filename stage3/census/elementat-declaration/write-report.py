import json
from pathlib import Path
out=Path(__file__).resolve().parent
read=lambda n:json.loads((out/n).read_text())
p=read('provenance.json');probes=read('probes.json');static=read('static.json');census=read('census.json');lane=read('lane-summary.json');verified=read('reconciliation.json')
assert lane['complete'] and verified['complete']
assert probes['p2']['nativeAgrees'] and probes['p2']['javascriptAgrees'] and probes['expectedValueAndIdentityChecks']==30
assert census['f9Gone'] and verified['differences']==0 and verified['uninitializedReads']==0
assert not static['unresolved'] and not static['strings']
counts=static['counts'];offsetTable='\n'.join(f"| {name} | {verified['offsetClasses'].get(name,0)} |" for name in ['in-range non-negative integer','in-range negative integer','out-of-range integer','non-integer','array undefined']);actual=lane['calls'];pieces=len(lane['pieces']);largest=max(r['seconds'] for r in lane['pieces'])
rows=[('Direct call sites','call'),('Constructor uses','new'),('prototype reads','prototype'),('name reads','name'),('length reads','length'),('Passed, stored, identity-compared, or other value use','value'),('Reassignment','reassignment'),('ts.elementAt reads','tsObjectRead'),('Top-level read before the declaration','topLevelBeforeDeclaration')]
statictable='\n'.join(f'| {label} | {counts[key]} |' for label,key in rows)
text=f'''# elementAt as a generic function declaration

Task k220r1g. Measurement only, for Adamic's ruling. No adaptation amendment,
compiler change, built compiler, or built TypeScript product is included.

Measured base: `{p['base']}`, fetched from origin/main.
TypeScript source: `{p['typescriptSource']}`. Stock checker: 6.0.3.
The full apply exits 0. The proposed tree differs in exactly one declaration
in core.ts; the entire elementAt body and toOffset are identical.
[Source validation](source-validation.json), [provenance](provenance.json),
and [instrumentation scope](instrumentation.json) retain the checks and hashes.

## P1 to P3

| Probe | Measured result |
| --- | --- |
| P1, current const arrow | Type check exits 0. Native and JavaScript lowering each exit 1 at p1.ts:6:5: NotYet, `a function returning T \\| undefined`. Emitter: internal/lower/functions.go:129, `(*lowering).signature`. This is the required control. |
| P2, generic function declaration | Type check, sanitized native build/run, JavaScript build/run, and stock Node all exit 0. All 30 rows agree byte for byte. All 30 expected-value checks pass, including reference identity for interface results. |
| P3, complete core.ts site census | Both guarded censuses exit 0 with zero checker diagnostics. Baseline has F9 at core.ts:1078:5; candidate has no F9 and no findings in elementAt or toOffset. |

[P1/P2 results](probes.json), [P1 source](p1.ts.txt), [P2 source](p2.ts.txt),
and [complete P3 site inventory](census.json) are retained. Each witness has
an interface Item, number, and string; each tests offsets -1, 0, 1, 5, -5
with an array and with an undefined array. Output has a label, value, and
expected-value/identity check per row. The two stock Node variants also agree.
Native used --sanitize --explain-checks. The emitted JavaScript runs with
oracle/node.mjs, the repository's normal Adamic runtime resolver.

P3 uses the reference report's directory census procedure, LATENT_FULL=1,
LATENT_ONLY_FILE set to each tree's core.ts, and LATENT_ASSERT_NO_OUTPUT=1.
The raw recorder retains its legacy checker-rejected measurement label;
its actual metadata has checker_rejected false and no diagnostics.
No speculative or replay flags are enabled, and ordinary Load/Lower output
is disabled in this measurement binary. It is a fresh paired run on the
measured base, not a reuse of the af42d3dc results from
`typescript/next-stops-after-wave` ({p['referenceBranchTip']}). Main now includes
the getAllKeys adaptation and compiler changes absent from that older census.

Both P3 runs have 270 recording units. Baseline has 209 deferred generic
units and 23 distinct Refused/NotYet sites; candidate has 210 and 22.
The only removed blocker kind/reason/phase is F9. Its failed-statement rollback
boundary also disappears. Raw event totals, including boundary events, are
39 and 37; distinct recorder sites including boundaries are 35 and 33.

Candidate elementAt, core.ts:1076:1, is a function declaration with status
`deferred_generic_instantiation`, instantiated false. toOffset, core.ts:975:1,
has the same status. Both have zero findings. This census does not claim
those deferred bodies were instantiated. P2 supplies the concrete interface,
number, and string instantiation and execution evidence.

## Ordinary declaration emit

The uninstrumented declaration candidate's ordinary `npm run build:compiler`
exits 0. `typescript.d.ts` and `tsserverlibrary.d.ts` are byte-identical to
the observed const build. `typescript.internal.d.ts` has two diff hunks
moving the same documented elementAt declaration from the const group to
the function group and changing its spelling to the proposed declaration.
[declaration-emit.json](declaration-emit.json) retains all hashes and the
complete diff. This internal declaration change also needs review in a real
99z amendment. The whole-lane API guards below apply to the observer tree,
which retains the original exported const signature; they do not certify
byte-identical internal declarations for the proposed function form.

## Observers in adapted source

The stock 6.0.3 checker resolves aliases and property symbols over all 711
TypeScript files in the full 746-file adapted src, including compiler,
services, server, harness, and test runner sources. Import/export bindings
are excluded from value-reference counts. There are zero unresolved
identifiers named elementAt and zero string-literal elementAt keys.
[static.json](static.json) retains options and every reference location.

| Observation | Static sites |
| --- | ---: |
{statictable}

All 11 call sites are in deferred function bodies: two in
src/compiler/transformers/es2015.ts and nine in
src/services/codefixes/convertToAsyncFunction.ts. No compiler, services,
server, or harness source passes, stores, constructs, or examines the
function itself. A separate stock-checker namespace audit finds zero computed
member accesses
and zero reflection calls on namespace objects exporting this exact symbol.
It finds three namespace value escapes: plugin initialization at
src/server/project.ts:2182:68, fourslash execution at
src/harness/fourslashImpl.ts:4721:11, and the debugging global at
src/testRunner/runner.ts:264:30. Those expose the namespace without directly
reading elementAt. [namespace-audit.json](namespace-audit.json) records them.
All five emitted elementAt export getters, three in run.js and two in
typescript.js, are therefore observed in the final full runtime inventory.
The observer returns the exact original function identity. Counts concern this
source and the measured lane, not arbitrary external consumers of the
exported internal helper.

[Stock Node observable witnesses](observable-delta.json) confirm the actual
JavaScript differences: early typeof is ReferenceError for the const and
function for the declaration; only the declaration owns a prototype; new
throws TypeError for the arrow and succeeds for the declaration. Both have
name elementAt and length 2. These are semantic differences even though the
resolved adapted source has no observing sites.

## Complete runtime lane count

| Measurement | Result |
| --- | ---: |
| Manifest units, each exactly once | {len(lane['completedUnits'])} / {lane['inventory']} |
| Completed pieces | {pieces} |
| Largest piece, seconds | {largest:.3f} |
| Ordinary test passes | {lane['counts']['pass']} |
| Expected failures | {lane['counts']['fail']} |
| Skips | {lane['counts']['skip']} |
| Named elementAt binding reads | {lane['reads']} |
| Namespace export getter reads | {lane['exportReads']} |
| elementAt calls | {actual} |
| Result or reference-identity differences | {lane['differences']} |
| Dead-zone errors | {lane['tdz']+lane['exportTdz']} |
| Reads before initialization, including bundled var | {verified['uninitializedReads']} |
| Reads through ts.elementAt | {lane['tsObjectReads']} |

| Observed offset/input class | Calls |
| --- | ---: |
{offsetTable}

All unit reports pass the unchanged lane identity, baseline, API, failure,
and count guards. All named binding reads are direct calls. Namespace export
reads are counted separately. The complete
counts equal the unchanged lane plan. Every piece verifies that src and
built/local stayed frozen. [lane-summary.json](lane-summary.json) and
[reconciliation.json](reconciliation.json) retain the totals; lane-NN.tar.gz
retains every raw read/call row, child exit code, unit report, and log.
Adjacent SHA256 files and internal SHA256SUMS cover the archives and members.
The independent reconciler reopens every archive, verifies both hash layers,
recounts the raw rows and reports, and rejects duplicate or missing units.

This is one final full inventory with paired execution within each observed call.
The original const arrow keeps its identity. Its scratch-only observer saves
the incoming offset, invokes a private generic function declaration with the
same array and original offset, and compares both results with Object.is.
The private declaration has the exact proposed body and unchanged toOffset;
the observer returns the original result. Thus objects and function values
must be the same reference, not merely similar serialized values. Hoisting,
constructor, and metadata differences are checked separately above. This is
not a claim that both compiler forms separately ran through the full lane.
The declaration candidate is separately built for the emit comparison above.

The binding observer records actual reads and detects both ReferenceError
and an undefined pre-initializer value in a bundle. Excluded seeds prove
changed primitive results, substituted object identities, cross-realm
dead-zone reads, bundled-var early reads, export-getter early reads, and
identity-preserving export reads are detected. Preparation,
seeds, and the dependency-layout probe are excluded from the lane count.
A preliminary 160-unit direct-read run is retained under preliminary-direct/;
it was stopped at a completed piece with no live children after discovering
the three namespace escapes. Its rows are not added to the final inventory.
The final inventory restarts at unit one with namespace-export coverage and
fresh frozen build hashes.

## Ownership

99z pins the whole exported const statement and rejects the declaration
with `missing or duplicate elementAt owner`; a real change belongs in a
99z amendment. No other adaptation pins its initializer or its callers'
call expressions. The existing core-multimap pin mentioning elementAt is
an unchanged import list.

There is an adjacent ownership contract to preserve: adaptation 30 pins
one declined `array[offset]` expression in elementAt. The expression, its
occurrence/total, and its unasserted form stay unchanged. The documented
`<callback>` enclosing-function label is not an adapter address; 30 keys
this read by expression and occurrence. Adaptation 40 owns toOffset's
generic signature and body, both unchanged. [ownership.json](ownership.json)
records the targeted checks. A whole-core replay of 30 stops identically
on baseline and candidate at its existing core.ts:992:21 declined-site
contract, so no successful whole-adapter replay is claimed here.

## Retained setup and scope

Setup sourced /workspace/adamic-tools/env.sh, ran bash cloud/setup.sh,
installed stage3/api with npm ci, and exported ADAMIC_TYPESCRIPT_SOURCE to
the pinned checkout. Setup exits 0; full apply exits 0. The base was plainly
pushed as the post-apply checkpoint, preserving one final census commit.
The active full apply exceeded the initial short supervisor bound while
making progress; the extension retained its own child status and enforced
30 minutes without log progress as a stop limit.

raw.tar.gz retains exact commands and child exits, both P3 streams and
unit timings, source manifests, source inputs, guarded .go.txt overlay
inputs, compiler-emitter evidence, observer seeds, and setup/probe logs.
It also retains rejected witness-harness attempts, the corrected entry
versus directory census invocation, and both pre-test dependency-layout
failures. None is counted as an accepted runtime unit. No guard was disabled
to repair setup. Native binaries, compiler binaries, node_modules, and full
built TypeScript products are excluded from the submitted artifacts.

## Recommendation

Recommend road (a') for Adamic's ruling as a JavaScript-changing amendment
of 99z: it removes the measured generic-arrow wall, agrees with Node on all
30 concrete witness rows, and gives {actual} identical paired lane results
with no observed early read or function-object observer in the requested
scope. Preserve the exact body, toOffset, and 30's declined indexed-read
contract. Hoisting, prototype ownership, and construction remain real
semantic differences. The internal declaration-kind change and relocation
also belong in that ruling. These measurements support a bounded ruling rather
than a claim of universal JavaScript equivalence. No adaptation is built or
landed by this report.
'''
assert '\u2013' not in text and '\u2014' not in text
(out/'REPORT.md').write_text(text)

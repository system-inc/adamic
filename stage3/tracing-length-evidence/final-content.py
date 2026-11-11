from pathlib import Path
from run import raw,root
import json,hashlib
lane=json.loads((raw/'lane-summary.json').read_text());census=json.loads((raw/'census-summary-report.json').read_text());built=json.loads((raw/'built-comparison-report.json').read_text());source=json.loads((raw/'source-comparison-report.json').read_text())
assert lane['wholeLaneExit']==0 and lane['emptyPopWrites']==0
assert lane['oracleCountsCheckedByEachGoUnit']=={'pass':106366,'fail':1,'skip':0}
for n in ['baseline-clean-apply','guarded-full-apply','before-build','after-build','before-types','after-types','before-census','after-census','mutant-pop-census','mutant-popAll-census','node-witness','reentrant-witness','reentrant-controls','adapter-guards-guarded','baseline-clean-comparison']:
 assert json.loads((raw/(n+'.json')).read_text())['exit']==0,n
assert not census['newlyVisibleFindingsInTracing']
assert census['reports']['after']['pop']==[] and census['reports']['after']['popAll']==[]
def calls(site):return sum(lane['histograms'][site].values())
siteRows='\n'.join(f'| {label} | {calls(site)} | `{json.dumps(lane["histograms"][site],sort_keys=True)}` |' for label,site in [('pop calls','popCall'),('pop decrement/splice','popWrite'),('popAll calls','popAllCall'),('popAll truncation/splice','popAllWrite')])
hunks='\n'.join(f'| {f["name"]} | '+', '.join('`'+h+'`' for h in f['hunks'])+' |' for f in built['files'] if f['classification']!='byte-identical')
proof=f'''# Tracing length adaptation proof

Task ff5eddz. Base: 16429a30452d027809f0ea85bc395479423e493e, origin/main
at setup. TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.
Node v24.19.0, Linux x64. This is a JavaScript-changing adaptation under
docs/rulings.md:306 and :307.

## Owners and scope

Only pop and popAll are pinned whole, before and after, preserving CRLF bytes.
Unknown bodies, partial application, duplicate owners and missing owners are
refused. The second application is byte-identical. The source comparison covers
{source['files']} files: only src/compiler/tracing.ts changes, exactly by these two
owner replacements. All module declarations and all startTracing bytes remain
identical. Its upstream 71 / adapted 72 typeCatalog.length = 0 write belongs to
the concurrent amendment to adaptation 47, not this change.

20 owns eventStack's optional args annotation; 32 owns indexed reads and the
private TraceRecord.typesPath annotation; 40 owns erased any annotations; 47
validates the require("fs") producer in startTracing; 62 and 63 own dumpTypes and
dumpLegend calls; 65 owns builtin declarations; 99-correctness-tracing-assert
pins stopTracing whole. None pins pop or popAll whole. No existing owner is
amended here. The typeCatalog declaration has no conflicting whole-owner pin.

## Empty pop and Node witness

pop first calls Debug.assert(eventStack.length > 0). Plain Debug.assert is
unconditional, even at AssertionLevel.None. If assertions are disabled, an empty
entry still throws TypeError when writeStackEvent(-1, ...) destructures
undefined before the decrement. However, this does not prove the stack remains
nonempty through writeStackEvent: JSON.stringify(args) can invoke a getter that
reenters popAll and empties the stack. A built-tree witness reaches this case
with assertions on. Upstream then throws RangeError("Invalid array length");
the unguarded splice returned normally. The final adaptation therefore keeps
an explicit empty-stack RangeError guard immediately before the splice.
The paired reentrant witness requires the same class, message and empty array
state before and after, with exactly one reentry.

The paired witness loads the two normal built typescript.js bundles in private
VMs. A read-only stack observer exposes array state, and a deterministic host
clock gives byte-comparable timestamps. For assertions off, it explicitly
replaces Debug.assert with a no-op hook; setAssertionLevel(0) alone is not an off
switch for plain assert. The before and after serialized observations are
byte-identical after every operation: three pushes, two pops, popAll on nonempty
and empty stacks, empty pop, and a push/pop after the failure.

| Empty pop mode | Class | Message | Array afterward |
|---|---|---|---|
| Assertions on | Error | Debug Failure. False expression. | [] |
| Reentrant popAll during args serialization | RangeError | Invalid array length | [] |
| Assertions off by explicit hook | TypeError | Cannot destructure property 'phase' of 'eventStack[index]' as it is undefined. | [] |

## Whole sharded lane and positive controls

All {lane['units']} real Go shard units pass, with child exit 0. Each unit checks its
manifest counts and identities: 106366 passing, one sanctioned API failure, and
zero skipped tests in total. The measurement uses the gate's adapted-tree
product path. A measurement-only Go overlay selects a private copied product
and carries the VM-aware preload through the toolchain environment.

The copied payload is byte-identical before instrumentation, including nested
node_modules fixtures. Only private emitted JavaScript copies receive counters.
The normal source and normal built trees have no counter edits. All manifest
units have counter-loaded process coverage, over {lane['processRecords']} completed Node
processes. Histograms give length -> execution count, with an empty histogram
meaning no execution.

| Site | Executions | Length histogram |
|---|---:|---|
{siteRows}

pop's decrement/splice executes on an empty array {lane['emptyPopWrites']} times. A zero
lane count is not direct execution coverage of these truncations. The separate
positive controls supply that coverage and agree before and after: pop has
16 calls, including four empty calls, and 12 reached writes at lengths 1, 2
and 3 (four each); popAll has 12 calls and 12 writes, eight at length 0 and four
at length 1. Empty entry reaches no write in either basic control. A separate
reentrant counter control records exactly one baseline decrement at length 0;
the candidate throws its explicit guard before reaching splice. Both outcomes
are RangeError with the same message and empty array state.

## Full tracing census and independent restoring mutants

All four runs exit 0 with LATENT_FULL=1, LATENT_ONLY_FILE set to that tree's
adapted tracing.ts, and LATENT_ASSERT_NO_OUTPUT=1. The overlay is measurement
only; its no-output guards remain enabled. Both named bodies are attempted.

| Owner | Before | After | Independent restoring mutant |
|---|---|---|---|
| pop | 164:9 NotYet: assigning a field of a value | No finding | Original 164:9 finding returns |
| popAll | 171:9 NotYet: assigning a field of a value | No finding | Same refusal returns at 172:9, shifted by the added guard line |

Any newly visible guard stop is retained in census-summary.json. The exact original finding texts are:

    src/compiler/tracing.ts:164:9: stage 0 can't lower assigning a field of a value yet
    src/compiler/tracing.ts:171:9: stage 0 can't lower assigning a field of a value yet

Each mutant restores only its own write and retains the other adapted owner.
There is no newly visible stop in either owner. For the file-wide delta only,
locations after the added guard are mapped back one line; there is no new stop
elsewhere either. Raw streams and owner reports retain their actual locations. The startTracing finding at 72:9 remains in every run and is not claimed
as removed. Raw streams retain every other finding.

## Normal builds, types and emitted output

The clean baseline and candidate full bash stage3/apply.sh runs exit 0. Both
normal npm run build commands exit 0. Both adamic types commands cover the same
79 roots and exit 0. Declaration emission in the adapter is identical with
stripInternal on and off; both normal public/internal declaration bundles also
match byte for byte.

All ten normal JavaScript files are compared. Seven are byte-identical; exactly
six hunks in three bundles change. Each full changed file equals its baseline
after only the literal pop and popAll write replacements. There are no esbuild
rename hunks and no unexplained bytes.

| Bundle | Hunks |
|---|---|
{hunks}

All {built['exportKeyCount']} runtime ts own export keys are identical, including the
symbol-key check. The built-output comparison uses the normal bundles, before
counter instrumentation.

## Evidence and retained failed attempts

Raw evidence and replay scripts are on
[typescript/tracing-length-evidence](https://github.com/system-inc/adamic/tree/typescript/tracing-length-evidence/stage3/tracing-length-evidence).
Go evidence is saved as .go.txt. Logs stay complete in files, and wrappers record
each child's own exit code. Setup sources /workspace/adamic-tools/env.sh, runs
cloud/setup.sh and locked installs for stage3/api and stage3/typescript, and sets
ADAMIC_TYPESCRIPT_SOURCE to the pinned checkout.

The initial setup began before the branch switch and used the old cohere pin;
the main-tip setup succeeds. The first baseline apply's source adaptations
finished, but a prematurely started build changed files during cache archiving;
that apply is invalid. A clean full replay exits 0 and reproduces every source
byte used by the baseline build. The first off-mode witness incorrectly assumed
AssertionLevel.None disabled plain assert; the corrected explicit-hook witness
passes. A counter-preparation attempt rejected esbuild's 1e3 spelling before any
lane began. The first lane attempt omitted 30 fixture files by recursively
ignoring node_modules, failed module-resolution shards, and was stopped in full;
none of its counts is used. Reusing that copy then hit read-only Git object files
before a lane began. A later verified-copy lane for the unguarded candidate was stopped when the
reentrant RangeError counterexample was found; it is excluded too. The accepted
run uses the guarded candidate, starts from a fresh copy, verifies the entire
payload, preserves all fixtures, and runs every shard anew. Failed logs
and the invalid lane's counts are retained separately.
'''
directory=root/'stage3/adapt/99-correctness-tracing-length';(directory/'PROOF.md').write_text(proof)
p=directory/'README.md';s=p.read_text();s=s.replace('Validation is in progress. No completed lane or census result is claimed yet.\nEvidence is kept outside stage3/adapt and will be published separately.',f'Both full applies, both normal builds, both 79-root type checks, the paired\nNode witness, and all {lane["units"]} sharded lane units pass. The lane counts pop\nand popAll calls and reached writes explicitly; empty pop reaches no write.\nBoth independent restoring mutants reproduce their original full-census\nfindings. See [PROOF.md](PROOF.md) for counts, error messages, every emitted\nhunk, export keys, and retained failed attempts. Raw evidence is on the separate\n[typescript/tracing-length-evidence](https://github.com/system-inc/adamic/tree/typescript/tracing-length-evidence/stage3/tracing-length-evidence) branch.');p.write_text(s)
for p in [directory/'README.md',directory/'PROOF.md']:
 assert '\u2013' not in p.read_text() and '\u2014' not in p.read_text(),p
hashes={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in directory.iterdir() if p.is_file()}
(raw/'final-content.json').write_text(json.dumps({'files':hashes,'laneUnits':lane['units'],'sourceFiles':source['files']},indent=2)+'\n')
print('final README and proof generated only after all acceptance evidence passed')

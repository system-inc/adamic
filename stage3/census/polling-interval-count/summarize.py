#!/usr/bin/env python3
import json,collections
from pathlib import Path
here=Path(__file__).resolve().parent
p=json.loads((here/'pin-census.json').read_text());a=json.loads((here/'adapted-census.json').read_text())
raw=[]
for f in (here/'runtime').glob('*.json'):raw.append(json.loads(f.read_text()))
events={};env={};after_any=collections.Counter();after_enum=collections.Counter()
for r in raw:
    for key,e in r['events'].items():
        dest=events.setdefault(key,{'count':0,'values':{}});dest['count']+=e['count']
        for k,v in e['values'].items():dest['values'][k]=dest['values'].get(k,0)+v
    for k,vs in r['environment'].items():
        d=env.setdefault(k,{})
        for value,n in vs.items():d[value]=d.get(value,0)+n
    after_any.update(r['afterAnyWrite']);after_enum.update(r['afterEnumWrite'])
readers=[]
for r in a['reads']:
    if r['class']=='a':
        key=f"{r['file']}:{r['line']}:{r['expression']}"
        readers.append({**r,'runtimeReads':events.get('read:'+key,{}).get('count',0),'afterAnyWrite':after_any[key],'afterEnumWrite':after_enum[key]})
summary={'configurationCalls':events.get('call:setCustomPollingValues',{}).get('count',0),'enumPropertyWrites':sum(v['count'] for k,v in events.items() if k.startswith('write:enum:')),'tableBindingAssignments':sum(v['count'] for k,v in events.items() if k in ('write:pollingChunkSize','write:unchangedPollThresholds')),'classAReads':sum(r['runtimeReads'] for r in readers),'classAReadsAfterAnyWrite':sum(r['afterAnyWrite'] for r in readers),'classAReadsAfterEnumWrite':sum(r['afterEnumWrite'] for r in readers),'classASitesExecuted':sum(r['runtimeReads']>0 for r in readers)}
result={'processFiles':len(raw),'summary':summary,'events':events,'environment':env,'classAReaders':readers}
(here/'count.json').write_text(json.dumps(result,indent=2)+'\n')
text='''# PollingInterval measurement (baobab)

Base: `464e5daa23a60588f5aaa656e95d93af05d582c3`. Upstream pin: TypeScript 6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`. No adaptation changed.

The compiler-API census uses the locked 6.0.3 checker over every .ts/.tsx file under src,
resolves imported symbols, distinguishes enum objects from enum-valued numbers, and
propagates object identity through casts, variable bindings, assignments and resolved
call parameters to a fixed point. The non-member object-use inventory is retained for
manual escape review. This is a closed-source census, not a proof about external clients
or arbitrary reflective JavaScript. Enum declaration initialization is not counted as
an application write.

'''
text+=f"| Tree | Source files | Value sites | (a) | (b) | (c) | Enum-write sites |\n|---|---:|---:|---:|---:|---:|---:|\n"
for label,d in [('Pin',p),('Adapted',a)]:
    c=d['counts'];text+=f"| {label} | {d['sourceFiles']} | {c['reads']} | {c['classes']['a']} | {c['classes']['b']} | {c['classes']['c']} | {c['writes']} |\n"
text+='''
## Writes and the runtime path

The single enum-writing expression is `levels[level] = customLevels![level] || levels[level]`
in nested `setLevel`. `setCustomPollingValues` passes the actual `PollingInterval` object
as the second argument of `setCustomLevels(baseVariable, levels: Levels)`; `setLevel`
captures that writable view. Its Low/Medium/High calls write all three properties
whenever any nonempty `TSC_WATCH_POLLINGINTERVAL_{LOW,MEDIUM,HIGH}` variable is supplied.
The right side preserves an old value for zero or NaN; other truthy numbers are accepted.

`pollingChunkSize` and `unchangedPollThresholds` are plain numeric-keyed objects,
not enums. `defaultChunkLevels` is a plain `Levels` object `{Low:32,Medium:64,High:256}`;
it is read and spread, never mutated by this path. `setCustomLevels` has just one call.
`getCustomLevels` writes a fresh temporary `Partial<Levels>` for each of the three
variable families. `getCustomPollingBasedLevels` builds a new table if intervals
changed or its own custom levels exist. The two outer table bindings are assigned on
every configuration call with an environment getter, even when their identity and
contents do not change. They are not additional enum-object writes.

'''
for label,d in [('Pin',p),('Adapted',a)]:
    for w in d['writes']:text+=f"- {label}: `{w['file']}:{w['line']}`: `{w['expression']}`; enum `{w['enum']}`.\n"
text+='''
No other enum write was found. In particular the `FileIncludeKind as any` escape is
a reverse read in program.ts, and `DiagnosticCategory` is passed to `Object.keys`
in harness/client.ts. Qualified enum names in the retained object inventory are
namespace access components; the CompletionSource import-equals is an alias.

## Reader classification and scheduling trace

Class (a) means the source-level intent requires a configured interval. Queue
construction stores its interval in `queue.pollingInterval`. Queue insertion and
transitions call `scheduleNextPollIfNotAlreadyScheduled`, then `scheduleNextPoll`,
then `host.setTimeout(..., pollingInterval, ...)`. Those transition arguments therefore
belong to (a), even though they also choose queues. Watch file/directory arguments
flow through `createSystemWatchFunctions`, `pollingWatchFile` and ultimately the host
polling worker or dynamic queue timer. Class (b) means a symbolic level: table keys,
threshold lookups, comparisons and switch labels. These stay symbolic, but a future
adaptation must make the boundary between symbolic queue identity and timer delay
consistent. Merely changing a few timer constants cannot establish that invariant.

The 3 keys in `createPollingIntervalBasedLevels` label the numeric-keyed chunk/threshold
tables. `pollPollingIntervalQueue` reads `pollingChunkSize[queue.pollingInterval]`;
`pollQueue` reads `unchangedPollThresholds[pollingInterval]`. Fixed-chunk polling reads
the low chunk size but schedules at High. Chunk and threshold customization is therefore
read indirectly as table contents, separately from the enum member sites below.

No direct PollingInterval member reads occur in watch.ts, watchUtilities.ts, or
server/types.ts. watchUtilities.ts's watch factory types and callbacks forward the
interval supplied by their callers. Every observed production reader serves watching; no unrelated compiler/checker consumer was found.
Outside compiler watch internals, (a) has four
sites in tsbuildPublic.ts and ten in server/project.ts and server/editorServices.ts.

All sites, including every (a) reader and its enclosing function:

| Class | Pin location | Adapted location | Function | Expression |
|---|---|---|---|---|
'''
assert len(p['reads'])==len(a['reads'])
for r,s in zip(p['reads'],a['reads']):
    assert r['expression']==s['expression'] and r['file']==s['file']
    text+=f"| {r['class']} | {r['file']}:{r['line']} | {s['file']}:{s['line']} | `{s['function']}` | `{s['expression']}` |\n"
text+='\nClass files:\n\n'
for cls in 'abc':text+=f"- ({cls}): "+(', '.join(sorted({r['file'] for r in a['reads'] if r['class']==cls})) or 'none')+'.\n'
text+='''
## Runtime measurement

See `oracle-run.json`, `oracle/report.json`, `count.json`, and per-process counter files.
The disposable counted tree wraps source value expressions without changing their
values, and records the three enum property writes plus both table-binding assignments.
A preload supplies the observer to build/test processes and their forked workers.
Counts are source-expression evaluations: esbuild may inline enum members into constants.
`afterAnyWrite` is process-local after either table-binding assignment or enum mutation;
`afterEnumWrite` is after the first actual enum property assignment. Local temporary
bindings and fresh customLevels construction are not configuration-target writes.
These definitions avoid treating the two unconditional binding assignments as evidence
that the lane exercised custom configuration.

'''
text+=f"Observed configuration calls: {events.get('call:setCustomPollingValues',{}).get('count',0)}.\n\n"
report=json.loads((here/'oracle/report.json').read_text())
text+=f"Oracle status: {report['status']}; test counts: `{json.dumps(report.get('counts',{}),sort_keys=True)}`.\n\n"
custom=any(value not in ('', 'undefined') and n for vs in env.values() for value,n in vs.items())
text+=('Numeric custom environment values were observed; see count.json.\n\n' if custom else 'All observed numeric customization variables were empty. The lane did not exercise custom polling intervals, chunk sizes, or thresholds. This is a coverage gap: the adaptation witness must supply those variables. Tests do set watch-mode variables such as TSC_WATCHFILE and TSC_WATCHDIRECTORY; that does not cover numeric customization.\n\n')
text+='The first instrumented attempt failed during build because an any-valued observer lost literal types. Its exit and logs are retained under oracle-attempt1. The final observer preserves its input type through a generic identity signature; the final run is the only input to count.json.\n\n'
text+=f"Enum property writes: {summary['enumPropertyWrites']}; table-binding assignments: {summary['tableBindingAssignments']}. Class (a): {summary['classAReads']} evaluations, {summary['classAReadsAfterAnyWrite']} after table/enum writes, {summary['classAReadsAfterEnumWrite']} after enum mutation. {summary['classASitesExecuted']} of 36 sites executed; sys.ts:1052 did not execute.\n\n"
text+='| Target write | Count | Assigned values and frequencies |\n|---|---:|---|\n'
for k in ['write:enum:Low','write:enum:Medium','write:enum:High']:
    if k not in events:text+=f"| {k} | 0 | none |\n"
for k,e in events.items():
    if k.startswith('write:'):text+=f"| {k} | {e['count']} | `{json.dumps(e['values'],sort_keys=True)}` |\n"
text+='\n| (a) reader (adapted) | Reads | After any target write | After enum write |\n|---|---:|---:|---:|\n'
for r in readers:text+=f"| {r['file']}:{r['line']} {r['expression']} | {r['runtimeReads']} | {r['afterAnyWrite']} | {r['afterEnumWrite']} |\n"
text+='''
## Node reference and reproduction

`node-reference.cjs` uses the locked, published TypeScript 6.0.3 bundle, whose `sys`
is initialized by getNodeSystem. It intercepts timer registration and selects dynamic
priority polling with Low=17, Medium=41, High=89. Actual output:

```
'''+(here/'node-reference.txt').read_text()+'''```

Direct adapted-source CommonJS output (same getNodeSystem, with compiler-API
transpilation and no bundle enum inlining):

```
'''+(here/'node-source.log').read_text()+'''```

The published bundle inlines watch-code enum member uses to 250/500/2000 while retaining
the mutable enum object. Therefore configured object values and internal queue identities
diverge: default numeric inputs schedule the defaults, and configured numeric inputs
fail queue lookup. This reference must not be reported as successful custom intervals.
A future witness needs an explicit custom-environment case and a decision about this
observable difference between source intent and the published bundle.

Reproduce from this base with stage3/api's locked npm install, the pinned upstream tree,
and `ADAMIC_TYPESCRIPT_SOURCE` pointing to it. Build the adapted tree using stage3/apply.sh.
Run `census.cjs TREE` for each tree. Copy the adapted tree, then run
`instrument.cjs COUNTED_TREE adapted-census.json`. Run stage3/oracle/run.sh on that copy
with `NODE_OPTIONS=--require=ABSOLUTE/preload.cjs` and `POLLING_MEASURE_DIR` set to an
absolute output directory. Use `run.py OUTPUT COMMAND...` to bound the whole invocation
at 1700 seconds, retain its log and report the child's exit code. Never instrument the
production adaptation. `summarize.py` aggregates the retained process files.
'''
(here/'REPORT.md').write_text(text)

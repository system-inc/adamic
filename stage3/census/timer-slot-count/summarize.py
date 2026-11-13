#!/usr/bin/env python3
"""Pair independent API scans; apply reviewed classifications; retain all candidates."""
import collections,hashlib,json,pathlib,subprocess,sys
here=pathlib.Path(__file__).resolve().parent
u,a=(json.loads(pathlib.Path(p).read_text()) for p in sys.argv[1:3])
review=json.loads((here/'classifications.json').read_text())
base='e3f59bf5509b1feef85b405807a732e11dc383b6' # Frozen measurement base, not the future census commit.
def pairs(key):
    assert len(u[key])==len(a[key]),key
    result=[]
    for x,y in zip(u[key],a[key]):
        assert (x['file'],x.get('name'),x.get('kind'))==(y['file'],y.get('name'),y.get('kind')),(key,x,y)
        result.append((x,y))
    return result
def inline(text):return ('`` '+text+' ``') if '`' in text else '`'+text+'`'
def pos(x):return f"{x['file']}:{x['line']}"
def pair(x,y):return {'upstream':pos(x),'adapted':pos(y)}
calls=[]
reads=[]
for x,y in pairs('accesses'):
    if x['file']=='src/testRunner/parallel/worker.ts':
        assert not x['assignableToSystem'] and all(d['file']=='src/testRunner/parallel/worker.ts' for d in x['slotDeclarations'])
        continue # Mocha.Runnable, not System
    assert x['slotDeclarations'] and all(d['file'] in ('src/compiler/sys.ts','src/compiler/watchPublic.ts','src/server/types.ts','src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts') for d in x['slotDeclarations'])
    assert x['receiverType']==y['receiverType']

    if x['kind']=='call':
        assert x['arguments']==y['arguments']
        row={**pair(x,y),'slot':x['name'],'receiverUpstream':x['receiverType'],'receiverAdapted':y['receiverType'],'argumentCount':x['argumentCount'],'arguments':x['arguments'],'scope':'System-compatible receiver' if x['assignableToSystem'] else 'partial System slot / WatchHost','classification':review.get(pos(x))}
        if x['name']=='setTimeout' and x['argumentCount']==2:row['callback']='op: callback parameter of Promise wrapper; no method extraction at this site; arbitrary supplied function may itself depend on this.'
        calls.append(row)
    if x['kind']=='read':reads.append({**pair(x,y),'expression':x['expression'],'shape':'guard' if x['file'].endswith(('watchPublic.ts','tsbuildPublic.ts')) else 'forward/bind'})
for x,y in pairs('bareCalls'):
    if x['file']=='src/compiler/sys.ts':
        assert x['arguments']==y['arguments']
        calls.append({**pair(x,y),'slot':x['name'],'scope':'destructured System slot','receiverUpstream':'RecursiveDirectoryWatcherHost (destructured parameter)','receiverAdapted':'RecursiveDirectoryWatcherHost (destructured parameter)','argumentCount':x['argumentCount'],'arguments':x['arguments'],'classification':review.get(pos(x))})
for x,y in pairs('bindings'):reads.append({**pair(x,y),'expression':x['text'],'shape':'destructure slot from RecursiveDirectoryWatcherHost or CreateSystemWatchFunctions'})
for x,y in pairs('suppliers'):
    if x['file']=='src/compiler/sys.ts' and x['line'] in (1054,1058,1122,1123):reads.append({**pair(x,y),'expression':x['value'],'shape':'forward destructured slot into object'})
impl_specs=[
('src/compiler/sys.ts',[1054],'ensureDynamicPollingWatchFile input object','forwarder','Forwards the supplied setTimeout unchanged; eventual provider receives original rest arguments.'),
('src/compiler/sys.ts',[1058],'ensureFixedChunkSizePollingWatchFile input object','forwarder','Forwards the supplied setTimeout unchanged; eventual provider receives original rest arguments.'),
('src/compiler/sys.ts',[1122,1123],'createRecursiveDirectoryWatcher input object','forwarder','Forwards supplied timers unchanged; callback gets original rest arguments from eventual provider.'),
('src/compiler/sys.ts',[1505,1506],'getNodeSystem createSystemWatchFunctions input object','forwarder','Passes Node global setTimeout/clearTimeout unchanged; Node invokes callback with original rest arguments.'),
('src/compiler/sys.ts',[1603,1604],'getNodeSystem nodeSystem','concrete host','Node globals handed through; callback receives original rest arguments. clearTimeout is Node cancellation.'),
('src/compiler/watch.ts',[679,680],'createWatchHost return object','forwarder','maybeBind binds only this to system (no argument prefix), preserving rest arguments; absent timer uses noop and never invokes callback.'),
('src/harness/harnessLanguageService.ts',[593,597],'SessionServerHost','concrete host','setTimeout empty body: never invokes callback; clearTimeout empty body.'),
('src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts',[468,469],'TestServerHost createSystemWatchFunctions input object','forwarder','bind(this) binds only receiver; TestServerHost later invokes callback with original rest arguments.'),
('src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts',[1103,1111],'TestServerHost','concrete host','Registers callback and args with Callbacks.register; invokeCallback destructures stored args and calls cb(...args). clearTimeout unregisters id.'),
('src/testRunner/unittests/tsserver/session.ts',[42,45],'mockHost','concrete host','setTimeout returns 0 and never invokes callback; clearTimeout is ts.noop.'),
('src/tsserver/nodeServer.ts',[268,269],'initializeNodeSystem existing sys host','concrete host','Assigns Node globals directly; Node invokes callback with original rest arguments; Node clearTimeout cancels.')]
implementers=[]
claimed=set()
for file,lines,name,kind,shape in impl_specs:
    sites=[]
    for x,y in pairs('suppliers'):
        if x['file']==file and x['line'] in lines:
            sites.append({**pair(x,y),'slot':x['name'],'upstreamBody':x['value'],'adaptedBody':y['value']});claimed.add(pos(x))
    assert len(sites)==len(lines),name
    implementers.append({'name':name,'kind':kind,'invocation':shape,'sites':sites,'stopCondition':False})
bulk_contexts=[]
for x,y in pairs('bulkCopies'):
    assert x['callee']=='copyProperties' and x['argumentSlots'][1]==['setTimeout','clearTimeout']
    assert all(d['file']=='src/compiler/core.ts' for d in x['declarations'])
    name='createWatchCompilerHost result' if x['file'].endswith('/watch.ts') else 'createSolutionBuilderWithWatchHost result'
    sites=[{**pair(x,y),'slot':slot,'upstreamBody':x['text'],'adaptedBody':y['text']} for slot in x['argumentSlots'][1]]
    implementers.append({'name':name,'kind':'forwarder','invocation':'copyProperties copies both timer values unchanged from WatchHost; original callback arguments pass through the previously bound System timers, or noop never invokes callback.','sites':sites,'stopCondition':False})
    bulk_contexts.append({**pair(x,y),'expression':x['text'],'slots':x['argumentSlots'][1]})
for x,y in pairs('genericCopyReads'):
    reads.append({**pair(x,y),'expression':x['expression'],'shape':'generic bulk-copy read, instantiated with WatchHost by both listed copyProperties calls'})
excluded_suppliers=[]
for x,y in pairs('suppliers'):
    if pos(x) not in claimed:
        assert x['file'] in ('src/testRunner/parallel/worker.ts','src/testRunner/unittests/tscWatch/watchApi.ts')
        excluded_suppliers.append({**pair(x,y),'reason':'Mocha timeout method' if '/parallel/' in x['file'] else 'slot removed by assignment of undefined, not a supplied timer implementation'})
assert len(calls)==19 and len(reads)==19
for c in calls:
    if c['slot']=='setTimeout' and c['argumentCount']>2:
        assert c['classification'] and c['arguments'][2:]==[t['expression'] for t in c['classification']['tails']]
assert len([c for c in calls if c['scope']!='destructured System slot'])==17
assert len([c for c in calls if c['scope']=='System-compatible receiver'])==9
assert all(d['code'] not in (2307,2688) for scan in (u,a) for d in scan['diagnostics'])
timer_files={c['upstream'].rsplit(':',1)[0] for c in calls}
assert all(d.get('file') not in timer_files for scan in (u,a) for d in scan['diagnostics'])
property_calls=[c for c in calls if c['scope']!='destructured System slot']
strict_calls=[c for c in calls if c['scope']=='System-compatible receiver']
partial_calls=[c for c in calls if c['scope']=='partial System slot / WatchHost']
alias_calls=[c for c in calls if c['scope']=='destructured System slot']
def tail_count(xs):return sum(c['slot']=='setTimeout' and c['argumentCount']>2 for c in xs)
counts={'propertyCalls':len(property_calls),'propertyCallsWithTail':tail_count(property_calls),'systemCompatiblePropertyCalls':len(strict_calls),'systemCompatiblePropertyCallsWithTail':tail_count(strict_calls),'partialHostPropertyCalls':len(partial_calls),'partialHostPropertyCallsWithTail':tail_count(partial_calls),'destructuredSlotCalls':len(alias_calls),'destructuredSlotCallsWithTail':tail_count(alias_calls),'allSlotCalls':len(calls),'allSlotCallsWithTail':tail_count(calls),'implementerSitesIncludingForwarders':len(implementers),'concreteHostSites':sum(i['kind']=='concrete host' for i in implementers),'forwardingObjectSites':sum(i['kind']=='forwarder' for i in implementers),'slotValueSupplies':len(claimed)+sum(len(b['slots']) for b in bulk_contexts),'nonCallSlotReads':len(reads),'stopCondition':any(i['stopCondition'] for i in implementers)}
runtime={'status':'not yet available'}
if len(sys.argv)>3:
    folder=pathlib.Path(sys.argv[3]);rows=[]
    for f in sorted(folder.glob('*.json')):
        d=json.loads(f.read_text());rows.extend(dict(r,pid=d['pid']) for r in d['rows'])
    tails=collections.defaultdict(lambda:{'count':0,'sites':set()})
    by_site=collections.Counter()
    for r in rows:
        by_site[r['site']]+=r['count']
        if not r['tail']:continue
        tail=[dict(v,process=r['pid']) if 'identity' in v else v for v in r['tail']]
        key=json.dumps(tail,sort_keys=True);tails[key]['count']+=r['count'];tails[key]['sites'].add(r['site'])
    runtime={'status':'measured','calls':sum(r['count'] for r in rows),'callsWithTail':sum(r['count'] for r in rows if r['tail']),'distinctTails':len(tails),'bySite':dict(sorted(by_site.items())),'tails':[{'tail':json.loads(k),'count':v['count'],'sites':sorted(v['sites'])} for k,v in sorted(tails.items())], 'identitySemantics':'Primitive values by value; objects/functions by WeakMap identity, scoped by process. Object contents are deliberately not traversed.'}
    runtime['oracle']=json.loads(pathlib.Path('/tmp/timer-slot-oracle/report.json').read_text())
    runtime['boundedRunner']=json.loads(pathlib.Path('/tmp/timer-slot-oracle.log').read_text().strip().splitlines()[-1])
    assert runtime['boundedRunner']['childReturncode']==1 and not runtime['boundedRunner']['timedOut']
    oracle=runtime['oracle']
    assert oracle['counts']=={'passing':106366,'failing':1,'pending':0}
    assert oracle['baseline_diffs']==['api/typescript.d.ts']
    assert [f['title'] for f in oracle['failures']]==['unittests:: Public APIs for typescript.d.ts should be acknowledged when they change']
    assert {k:v['exit'] for k,v in oracle['phases'].items()}=={'install':0,'build':0,'tests':1}
    runtime['matchesExpectedOracleOutcome']=True
result={'task':'995spev condition 1','adamicMain':base,'typescriptPin':'050880ce59e30b356b686bd3144efe24f875ebc8','counts':counts,'calls':calls,'implementers':implementers,'nonCallReads':reads,'bulkReadContexts':bulk_contexts,'excludedSuppliers':excluded_suppliers,'runtime':runtime,'apiScans':{'upstream':u,'adapted':a}}
(here/'count.json').write_text(json.dumps(result,indent=2)+'\n')
lines=['# Timer slot census — 995spev, condition 1','',f'Base: adamic main `{base}`. TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.','', '**17 property calls, 9 with trailing arguments, in each tree.** Of these, **9 calls / 4 tails** have receivers assignable to System (including ServerHost and TestServerHost); **8 calls / 5 tails** use partial System timer-host interfaces. Another **2 calls / 1 tail** invoke destructured System slots. Thus the inclusive slot census is **19 calls / 10 tails per tree**, not 38 distinct sites across the two snapshots.','', '**13 supplier/implementer sites: 5 concrete host supply sites (4 object/class definitions plus 1 overwrite of an existing host) and 8 forwarding object sites; 24 individual slot supplies. Stop condition: NO.** None supplies its own replacement callback arguments. Fakes never invoke callbacks. Counts refer to source sites, not runtime instances.','', '## Method and scope','',f'TypeScript compiler API {u["typescript"]}; one Program over every source file under src, using upstream tsconfig-base options, Node/Mocha/Chai type roots, and no emit. Upstream has {len(u["sourceFiles"])} roots; adapted has {len(a["sourceFiles"])}. All loaded src files are visited. The scanner records receiver types, assignability, inheritance, slot-symbol declarations, callback bodies, every matching property read/write, destructured bindings, bare timer calls, suppliers, spreads, Object.assign and typed assign/extend/clone/copyProperties/Object.create candidates. The saved raw scans retain excluded Mocha/global calls and diagnostics. There are no timer-bearing spreads or Object.assign sites; the typed bulk-copy scan finds exactly two copyProperties calls and no timer-bearing assign/extend/clone/Object.create calls. Classification and pairing assertions fail rather than silently drop new candidates.','', 'WatchHost does not extend System, and the internal polling hosts contain only picked System slots. They are included because createWatchHost/createSystemWatchFunctions forward System timers into them. The strict System-compatible subtotal is separately reported. TestServerHost explicitly implements ServerHost; structural assignability also catches it despite getBaseTypes not representing implements clauses.','',f'The unified programs report {len(u["diagnostics"])} upstream and {len(a["diagnostics"])} adapted diagnostics (including cross-project/global-library merging and existing inference issues); these are retained, not suppressed. No missing module/type-definition diagnostics remain; no source file containing a timer-slot invocation has a diagnostic. This is a census, not a claim that a unified all-src project is an upstream build configuration.','', '## Every call','', 'Locations below are upstream → adapted. The receiver column is identical in both independently checked trees (asserted); both raw checker types and both source locations are also in count.json. A `setTimeout` tail begins after callback and ms. clearTimeout has no callback.','', '| Upstream → adapted | Slot | Receiver | Args | Callback |','|---|---|---|---:|---|']
for c in calls:
    cb=c.get('classification',{}).get('callback') if c.get('classification') else c.get('callback','—')
    lines.append(f"| {c['upstream']} → {c['adapted']} | {c['slot']} | `{c['receiverUpstream']}` | {c['argumentCount']} | {cb} |")
lines+=['','## Every trailing argument','', 'Value stability distinguishes a stable object reference from mutation of the referenced object. Function calls are inspected rather than automatically declared effectful. The classification is the same in both trees.']
for c in calls:
    if not c.get('classification'):continue
    lines+=['',f"### {c['upstream']} → {c['adapted']}",'',c['classification']['callback']]
    for t in c['classification']['tails']:lines+=['',f"- {inline(t['expression'])} — Evaluation: {t['sideEffect']} Stability: {t['canChange']}"]
lines+=['','The PollingInterval enum is writable: sys.ts setCustomPollingValues/setCustomLevels updates its properties. Even where ordinary runs leave it stable, moving those reads into a firing closure is not a generally valid equivalence. The two static method callbacks read their explicit self arguments; neither reads dynamic this. The two arrow callbacks retain lexical this.','', '## Every implementer and forwarding supplier','', 'Each row counts one object/class/host-initialization source site; set and clear locations are listed individually. The two polling input objects supply setTimeout only. clearTimeout never invokes a callback.']
for im in implementers:
    lines+=['',f"- **{im['name']}** ({im['kind']}): "+'; '.join(f"{s['slot']} {s['upstream']} → {s['adapted']}" for s in im['sites'])+'. '+im['invocation']]
lines+=['','TestServerHost execution path (identical in both snapshots):','', '```ts','setTimeout(callback: TimeOutCallback, ms: number, ...args: any[]): number {','    return this.timeoutCallbacks.register(callback, args, ms);','}','// Callbacks.register stores { cb, args, ms, time }. invokeCallback later does:','const { cb, args, ms, time } = data;','// ... removes the timer and advances simulated time ...','cb(...args);','```','', 'No stop-condition implementer exists, so there is no offending body to quote. All supplying member bodies/expressions are retained verbatim in count.json. The typings installer uses the existing sys/installTypingHost, not a distinct timer-slot implementation. watchApi.ts:190–191 (same adapted lines) removes both slots by assigning undefined; those are writes, not reads or implementations. Mocha Timeout.clearTimeout is unrelated and excluded.','', '## Slot reads without invocation','', '| Upstream → adapted | Expression | Use |','|---|---|---|']
for r in reads:lines.append(f"| {r['upstream']} → {r['adapted']} | `{r['expression']}` | {r['shape']} |")
lines+=['', '19 source read sites per tree: 10 named property reads (6 presence guards, 4 passed/bound), 4 destructuring reads, 4 forwarding shorthand reads of those destructured aliases, and the generic second[id] read in copyProperties. Node-global identifier reads in getNodeSystem/nodeServer are provider inputs, not reads of an existing System slot; they are retained under implementers.','', 'The generic bulk-copy read is instantiated for timers at: '+ '; '.join(b['upstream']+' → '+b['adapted']+' (`'+b['expression']+'`)' for b in bulk_contexts)+'. Each copies both slots without calling either.', '', '## Runtime cross-check','']
if runtime['status']=='measured':
    o=runtime['oracle'];lines +=[f"The lane's stage3/oracle/run.sh ran all runners with eight workers on Node {o['node']}, on a disposable instrumented adapted tree. Its install/build/test child exits and durations are retained in count.json. The outer bounded runner also prints its child's actual return code. Measurement instrumentation replaced only the 19 census callees; receiver and callee are obtained before argument evaluation, arguments are evaluated once, and Reflect.apply preserves the original receiver and argument list. No callback is wrapped. Counts are taken at the call site, before forwarding, so internal forwarding is not double-counted. This cross-check is instrumented, not an unmodified lane verdict.",'',f"Observed **{runtime['calls']:,} total calls**, **{runtime['callsWithTail']:,} carrying a tail**, and **{runtime['distinctTails']:,} distinct tails**. count.json runtime.tails contains every distinct tail, occurrence count and adapted source sites. Primitive values use value equality; objects/functions use process-scoped identity, without reading mutable object contents.",'',f"Oracle counts: {o.get('counts')}; status {o['status']}. Baseline differences: {o.get('baseline_diffs')}. This exactly matches the lane's expected counts, phase exits and sole API-baseline failure title; the oracle status fail is expected. The cross-check completed in {o['wall_seconds']} seconds with no timeout. See commands.md for reproduction.",'', '| Adapted site | Runtime calls |','|---|---:|']
    lines += [f'| {k} | {v} |' for k,v in runtime['bySite'].items()]
else:lines+=['Runtime results pending.']
lines+=['', '## Coverage limits','', 'Covers source sites in src (compiler, services, server, harness, testRunner, typingsInstaller, tsserver, jsTyping and all other loaded src files), including type declarations for validation. Does not count external user hosts, dependencies, fixtures under tests/, string-embedded evaluation programs, generated JS/built output as additional source sites, or hypothetical runtime-created/computed unknown property names. The two typed bulk-copy contexts and their generic property read are explicitly covered. Global Node/DOM timer calls and Mocha timeout methods are not System-slot calls. Aliased slot invocations found here are included separately from property-call syntax. No adaptation is changed. Runtime observations cover only executed oracle paths and are not evidence that unexecuted sites cannot receive other values.','']
(here/'REPORT.md').write_text('\n'.join(lines))
print(json.dumps(counts,indent=2))

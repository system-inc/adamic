#!/usr/bin/env python3
"""Render compiler-API use records; no text-based discovery or flow inference."""
import collections, gzip, hashlib, json, pathlib, sys
here=pathlib.Path(__file__).resolve().parent
source=pathlib.Path(sys.argv[1]);data=json.loads((source/'a-star-uses.json').read_text())
origins=data['origins'];events=data['events'];byid={e['id']:e for e in events}
assert len(origins)==15
assert not data.get('unmodeledCarriers'),data.get('unmodeledCarriers')
assert data['headlineSites']==['B10','B22','B26','B27','B47','B48'],data['headlineSites']
assert all(e.get('typed') and not e.get('spread') for e in events if e['kind']=='direct-call')
initial={x['id']:x for x in origins}
# The actual argument vectors at each API-discovered invocation are checked, rather
# than assuming callback compatibility implies a particular runtime vector.
extra=[];unknown=[]
for e in events:
 for oid in e['origins']:
  M=initial[oid]['M']
  if e['kind'] in ['call','direct-call']:
   if e.get('spread'):unknown.append([oid,e['id'],'spread vector'])
   elif e['written']>M:extra.append([oid,e['id'],e['written'],M])
  elif e['kind']=='callback-or-argument':
   cb=e['callback'];actual=[]
   if cb.get('arrayCallbackArguments') is not None:actual.append(cb['arrayCallbackArguments'])
   for impl in cb.get('local',[]):
    for call in impl['calls']:
     if call.get('spread') or call.get('indirect')=='apply':unknown.append([oid,e['id'],'callback vector not fixed'])
     else:actual.append(call['written'])
   if not actual:unknown.append([oid,e['id'],'no callback invocation implementation'])
   elif max(actual)>M:extra.append([oid,e['id'],max(actual),M])
  elif e['kind'] in ['apply','other-CallExpression']:unknown.append([oid,e['id'],'indirect vector'])
assert not extra,extra
assert not unknown,unknown
assert not any(e['kind']=='apply' or e.get('callback',{}).get('arrayCallbackArguments') is not None for e in events)
assert not any(e.get('callback',{}).get('callee','').startswith('Reflect.') for e in events)
guard_sites={o for e in events if e['kind']=='guard-or-comparison' for o in e['origins']}
headline=set(data['headlineSites']);with_guards=sorted(headline|guard_sites)
assert len(with_guards)==11
fmt=lambda v:'`'+str(v).replace('|','\\|').replace('\n',' ')+'`'
where=lambda x:x['file']+':'+str(x['line'])+':'+str(x['column'])
short=lambda x:x['file'].split('/')[-1]+':'+str(x['line'])
q=[]
def p(x=''):q.append(x)
p('## Downstream uses of the fifteen a* bound values')
p()
p('**6 of the 15 a* sites have a downstream callback, forwarding store/copy, rebinding or indirect call:** **B10** `program.ts:1957`, **B22** `watch.ts:678`, **B26** `watch.ts:774`, **B27** `watch.ts:775`, **B47** `watchPublic.ts:548`, and **B48** `watchPublic.ts:551`. **None of the identified caller sites passes more function arguments than that origin’s M.** In particular, no selected bound value is used as an array-method callback, invoked through `.apply`, or invoked through `Reflect` in this audit.')
p()
p('Counting non-invoking presence/comparison guards as well gives a **conservative 11/15**: '+', '.join(fmt(x)+' ('+fmt(short(initial[x]))+')' for x in with_guards)+'. The six-site headline excludes those guards and excludes each bind’s initial storage; otherwise initial storage alone would count every site. Guards are listed separately below. The guard-inclusive count is a may-flow count: the shared `ProgramHost.now`/solution-host path includes the later overwrite at B49, so it is not proof that the B35 wrapper itself executes at `host.now()`.')
p()
p('This follows the bound result and its receiver-only rebindings, not every unrelated function with the same member name. It follows field reads, local aliases, shorthand/destructured cache records, source-call arguments/returns and the source `copyProperties`/`assign`/`clone`/`extend` helpers. A fresh closure which calls the bound function is inspected at that call; its own identity is not silently equated with the captured bound function. Returned host containers are followed to their local consumers; their initial return is not counted as another function-valued store. The scope is the same 79 source roots, with no harness/server/services code and no invented outside caller bodies.')
p()
p('### All fifteen origins')
p()
p('| Origin | Initial storage of the bound result | M | Downstream non-direct uses | Ordinary typed calls found (argument counts) | More than M at an identified caller? |')
p('|---|---|---:|---|---|---|')
for o in origins:
 direct=[byid[u] for u in o['directCalls']]
 calls=', '.join(fmt(short(e['call']))+' ('+str(e['written'])+')' for e in direct) or 'none found'
 nd=', '.join(o['nonDirect']) or 'none found'
 p(f"| {o['id']} {fmt(short(o))} | {fmt(o['storage']['destination'])} | {o['M']} | {nd} | {calls} | No identified case |")
p()
p('The storage contracts and declaration locations are recorded in `evidence/a-star-uses.json.gz`. For B46/B55 the destination is `compilerHost.getModuleResolutionCache`; for B49 it is the solution-builder host’s `now`. B47/B48 first enter the two local `customHasInvalidated...` variables rather than a field.')
p()
p('### Every identified downstream callback/store/copy/rebind/indirect-call use')
p()
p('Each row is an API-discovered read of a bound identity or its receiver-only rebound descendant. `.call`’s first argument is the receiver and is excluded from the function-argument count. Rebinding here supplies no fixed function arguments. Copies/stores do not invoke the value; the caller column names the downstream invocation evidence. The source alias/call fields are retained for all origins, including the nine with no such use.')
p()
p('| Use | Origin(s) | Source quote | Form | Can this use introduce more arguments than M? |')
p('|---|---|---|---|---|')
for e in events:
 if not e['headline']:continue
 use=e.get('call') or e.get('use') or e
 code=use.get('text',e['read'])
 kind=e['kind']
 if kind=='call':answer=f"No: exactly {e['written']} function argument(s); no spread."
 elif kind=='bind':answer=f"No invocation here; {e['written']} fixed function arguments. Descendant calls remain at the original M."
 elif kind=='callback-or-argument':
  parts=[]
  for impl in e['callback'].get('local',[]):
   for call in impl['calls']:parts.append(fmt(short(call))+' '+fmt(call['text'])+f" ({call['written']} argument)")
  answer='No: the source callback consumer invokes it with one argument: '+'; '.join(parts)+'.'
 elif kind=='object-copy':answer='No invocation: own-property copy; the copied timer is called with 3 arguments (M=3). See core.ts:1513 below.'
 elif kind in ['alias-store','field-store','store','destructure-store']:
  answer='No invocation: alias/field movement. The getter remains a zero-argument getter.' if e['origins']==['B10'] else 'No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument.'
 else:answer='No identified extra-argument caller; see the retained use record.'
 p(f"| {e['id']} | {', '.join(e['origins'])} | {fmt(where(use))} {fmt(code)} | {kind} | {answer} |")
p()
p('`copyProperties` performs the actual dynamic read and store at **`src/compiler/core.ts:1513`**, quoted exactly:')
p()
p('```typescript\nfirst[id] = second[id];\n```')
p()
p('For B22, the two call sites above use `second` equal to the watch host, and `id = "setTimeout"` copies that bound value. The receiver/label timer calls are `tsbuildPublic.ts:2074`, `watchPublic.ts:872`, and `watchPublic.ts:900`; each passes exactly `(callback, timeoutDelay, t1)`, three function arguments, matching M=3. Being the *first/target* argument of `copyProperties` does not read an existing `now` or directory method; the key set comes from `second`. This prevents counting unrelated existing target members as spread uses.')
p()
p('For B26, the cache forwarding is included, not stopped at the first alias: `changeCompilerHostLikeToUseCache` captures `originalDirectoryExists`, returns it as a shorthand field, the watch caller destructures that record, and later restores `compilerHost.directoryExists`. Its `originalDirectoryExists.call(host, directory)` passes one argument. The originals record is modeled **per helper call**: the solution builder’s separate call on its `ProgramHost` does not inherit the watch compiler host’s bound method merely because both calls use the same helper body. Any B26 attribution to that separate solution-host restoration is excluded.')
p()
p('B47/B48 are passed as callbacks to `resolutionCache.createHasInvalidatedResolutions` at `watchPublic.ts:628`. The actual implementation is `resolutionCache.ts:739`; it captures them in new one-parameter predicates and invokes them at `:749` and `:754` as `customHasInvalidatedResolutions(path)` and `customHasInvalidatedLibResolutions(libFileName)`. It supplies no array index, array object, rest tail or spread. The later predicate-field aliases in `program.ts:1603` and `:1684` refer to those fresh wrapper predicates, not directly to the original two bound-function identities; their calls into the original identities are the two explicit one-argument calls just listed.')
p()
p('### Presence/comparison reads (no function invocation)')
p()
p('| Read | Origin(s) in the may-flow audit | Quoted value read | Can add function arguments? |')
p('|---|---|---|---|')
for e in events:
 if e['kind']!='guard-or-comparison':continue
 p(f"| {e['id']} {fmt(where(e))} | {', '.join(e['origins'])} | {fmt(e['read'])} | No: presence/comparison only |")
p()
p('### Cross-checks and coverage')
p()
p(f"The API pass scanned **{data.get('readExpressionsScanned','all recorded')}** member/alias read expressions, **{data.get('sourceCallsScanned','all recorded')}** source calls and **{data.get('objectSpreadsScanned','all recorded')}** object-spread syntax nodes. It retained {len(events)} attributed read/use records, {len(data['surfaceReads'])} non-direct callable member-name reads for the independent surface cross-check, and {len(data['carrierCalls'])} conservative host-carrier argument records. The latter are an over-approximation used to inspect copy/index/Reflect paths, not additional function invocation sites. Callable identities are constrained by checker types; generic callback contracts stay open. Object carriers stay conservative. The originals cache record is modeled separately at each call, including its shorthand return fields and destructured aliases; the audit reports any other function-valued literal/array carrier it cannot model. There are zero such unresolved carriers in this result.")
p()
p('The independent non-direct member-read cross-check leaves only three non-guard/non-bind same-name cases outside the attributed identity set: native `Date.now` at `performanceCore.ts:103` (a different symbol and number-valued contract), and the two fresh invalidation-wrapper aliases at `program.ts:1603`/`:1684` explained above. Source index/Reflect carrier reads were inspected; the relevant dynamic function-valued read is `second[id]` in the explicit copy helper. No attributed `.apply`, `Reflect` invocation, array-method callback or untyped direct call remains in the records. All recorded direct calls have fixed argument lists of size ≤ M, and both callback consumers have source bodies with one-argument invocation lists.')
p()
p('**Limit:** this is a compiler-API static may-use audit, not execution tracing or a theorem of exact path feasibility. It is flow/context insensitive except for the cache originals record, retains source field/call contracts and explicit unknown host boundaries, and cannot constrain code outside the 79 roots. The six positive sites and their quoted uses are reviewable; “none found” is scoped to this audit. In particular, the result does not assert that outside JavaScript callers can never supply extra arguments. No native tsc run or production lowering change was used.')
p()
p('Reproduce the additional audit with `node stage3/census/bind-sites-static/uses-run.cjs ADAPTED_TREE NEW_OUTPUT_DIRECTORY`, then render with `python3 stage3/census/bind-sites-static/write-uses.py NEW_OUTPUT_DIRECTORY`. The renderer checks the 15 origins, six affected sites, fixed direct-call vectors, callback body vectors and absence of an attributed apply/Reflect/array-callback use before writing this section. `write-report.py` incorporates the saved section without changing the original measurements.')
(here/'USES.md').write_text('\n'.join(q).rstrip()+'\n')
e=here/'evidence';e.mkdir(exist_ok=True)
(e/'a-star-uses.json.gz').write_bytes(gzip.compress((source/'a-star-uses.json').read_bytes(),mtime=0))
review={'a_star_origins':15,'with_non_direct_use':6,'sites':data['headlineSites'],'guard_inclusive_may_use_count':len(with_guards),'guard_inclusive_sites':with_guards,'identified_extra_argument_callers':extra,'unresolved_recorded_caller_vectors':unknown,'records':len(events),'source_reads_scanned':data.get('readExpressionsScanned'),'source_calls_scanned':data.get('sourceCallsScanned'),'source_object_spreads_scanned':data.get('objectSpreadsScanned'),'unmodeled_carriers':data.get('unmodeledCarriers',[]),'scope':'79 roots; may-use audit; external callers excluded'}
(e/'a-star-uses-review.json').write_text(json.dumps(review,indent=2)+'\n')
(e/'SHA256SUMS').write_text(''.join(hashlib.sha256(f.read_bytes()).hexdigest()+'  '+f.name+'\n' for f in sorted(e.iterdir()) if f.name!='SHA256SUMS'))
print(json.dumps(review,indent=2))

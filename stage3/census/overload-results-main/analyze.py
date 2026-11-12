"""Enumerate every matching observation and both prior census identity views; no pair."""
import pathlib,json,re,collections,subprocess,sys
from merge import load
out=pathlib.Path(__file__).resolve().parent;tree=pathlib.Path('/workspace/orm-adapted');rows=load('/workspace/orm-build-main/latent.jsonl')
assert rows[0]['status']=='measurement' and rows[0]['latent_mode']=='first-error' and not rows[0]['checker_rejected']
patterns={k:re.compile(r'^overload \d+ of (\S+) '+k+r'(?:\s|$)') for k in ['result','parameter']}
expected={'filter':19,'createBuilderProgram':9,'getProperties':3,'arrayFrom':4,'arrayToMap':2,'getParseTreeNode':2,'convertOptionsFromJson':2,'sameMap':1,'getDiagnostics':1,'injectClassNamedEvaluationHelperBlockIfMissing':1,'transformNamedEvaluation':1};assert sum(expected.values())==45
units={u['where']:u for r in rows for u in r.get('units',[])};first={};findings=[]
for r in rows[1:]:
 for f in r.get('findings',[]):
  findings.append(f)
  if f['phase'] in ('lowering','deferred') and f['unit'] in units:first.setdefault(f['unit'],f)
def category(f):
 for kind,pattern in patterns.items():
  m=pattern.match(f['reason'])
  if m:return kind,m.group(1)
 return None
selected=[(i,f,category(f)) for i,f in enumerate(findings) if category(f)]
# Request the same unmodified stock normalizer as stage3/census/pair uses.
files=collections.defaultdict(list);seen=set()
for _,f,_ in selected:
 for field in ['where','unit']:
  w=f[field]
  if not w or w=='program' or w in seen:continue
  seen.add(w);file,line,column=w.rsplit(':',2);relative=pathlib.Path(file).relative_to(tree).as_posix();files[relative].append([w,int(line),int(column)])
request=dict(tree=str(tree),typescript='/workspace/orm-main/stage3/api/node_modules/typescript',files=files)
(out/'normalization-request.json').write_text(json.dumps(request,indent=2)+'\n')
p=subprocess.run(['node','/workspace/orm-main/stage3/census/pair/normalize.cjs'],input=json.dumps(request),capture_output=True,text=True);print('normalization_child_exit='+str(p.returncode),flush=True)
if p.returncode:print(p.stderr);sys.exit(p.returncode)
ids=json.loads(p.stdout);(out/'normalization.json').write_text(json.dumps(ids,indent=2)+'\n')
def site(w,distinct):
 if w=='program':return ('<program>','',0)
 x=ids[w];return x['file'],x['sha256'],x['read_offset' if distinct else 'byte_offset']
ledgers={k:dict(observations=[],raw={},distinct={}) for k in patterns};firstcounts={k:collections.Counter() for k in patterns}
for w,f in first.items():
 c=category(f)
 if c:firstcounts[c[0]][c[1]]+=1
for index,f,(kind,function) in selected:
 u=units.get(f['unit']);event=dict(index=index,function=function,unit_name=u.get('name',u['kind']) if u else 'refusal scan',unit_status=u['status'] if u else None,is_first_stop=first.get(f['unit']) is f,finding=f)
 ledger=ledgers[kind];n=len(ledger['observations']);ledger['observations'].append(event)
 rawkey=(site(f['where'],False),f['kind'],f['reason'],f['text']);distinctkey=(site(f['where'],True),f['reason'])
 for view,key in [('raw',rawkey),('distinct',distinctkey)]:
  if key not in ledger[view]:ledger[view][key]=dict(function=function,identity=key,finding=f,observations=[])
  ledger[view][key]['observations'].append(n)
summary=dict(main=json.loads((out/'pins.json').read_text())['main'],header=rows[0],units=len(units),attempted=sum(u['status'] in ('attempted','panic') for u in units.values()),no_first_stop_units=sum(u['status']=='attempted' and w not in first for w,u in units.items()),canonical_overload_result_first_stops=sum(bool(re.search(r'^overload \d+ of .+ result .+ cannot be served by implementation result ',f['reason'])) for f in first.values()),categories={})
for kind,l in ledgers.items():
 observed=collections.Counter(e['function'] for e in l['observations']);raw=collections.Counter(e['function'] for e in l['raw'].values());distinct=collections.Counter(e['function'] for e in l['distinct'].values());names=set(observed)|(set(expected) if kind=='result' else {'visitNodes','visitArray','visitArrayWorker','visitEachChild','visitParameterList'})
 table={name:dict(observations=observed[name],raw=raw[name],distinct=distinct[name],first_stop_units=firstcounts[kind][name],handoff=expected.get(name) if kind=='result' else {'visitEachChild':6,'visitParameterList':3}.get(name)) for name in sorted(names)}
 totals=dict(observations=sum(observed.values()),raw=sum(raw.values()),distinct=sum(distinct.values()),first_stop_units=sum(firstcounts[kind].values()))
 summary['categories'][kind]=dict(functions=table,totals=totals,present_functions=sorted(observed),listed_but_absent=sorted(set(expected)-set(observed)) if kind=='result' else [],unlisted_present=sorted(set(observed)-set(expected)) if kind=='result' else [],handoff_total=45 if kind=='result' else 27)
 l['raw']=list(l['raw'].values());l['distinct']=list(l['distinct'].values())
(out/'overload-findings.json').write_text(json.dumps(ledgers,indent=2)+'\n');(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:v['totals'] for k,v in summary['categories'].items()},indent=2));print('first stops',summary['categories']['result']['totals']['first_stop_units'],'no first stop',summary['no_first_stop_units'])
for kind in patterns:
 print(kind,'functions')
 for name,v in summary['categories'][kind]['functions'].items():print(name,json.dumps(v))

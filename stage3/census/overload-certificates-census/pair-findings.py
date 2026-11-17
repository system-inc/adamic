"""Pair complete same-input runs by unit and position-free diagnostic tuples."""
import collections,json,pathlib,re,sys,time
scratch=pathlib.Path(sys.argv[1]).resolve();tree=scratch/'adapted';out=scratch/'analysis';out.mkdir(exist_ok=True)
manifest=[]
for manifestPath in (scratch/'results').glob('*manifest.json'):
 for attempt in range(20):
  try:rows=json.loads(manifestPath.read_text());break
  except json.JSONDecodeError:time.sleep(0.05)
 else:raise RuntimeError('manifest still being written: '+str(manifestPath))
 manifest.extend(rows)
manifest += [json.loads(p.read_text()) for p in (scratch/'results').glob('*/full-directory/receipt.json')];callers=json.loads((scratch/'callers.json').read_text())
baseRoot=str(tree)+'/'
def relative(s):return str(s).replace(baseRoot,'')
def strip(s):
 s=relative(s or '')
 return re.sub(r'((?:[^\s:\"\'<>]+/)?[^\s:\"\'<>]+\.(?:ts|tsx|a|go)):\d+(?::\d+)?',r'\1',s)
def fingerprint(f):return tuple([f.get('kind',''),f.get('phase',''),strip(f.get('reason','')),strip(f.get('text',''))])
def family(f):
 reason=strip(f.get('reason',''));m=re.search(r'overload (\d+) of (\S+) (result|parameter \S+|type parameter \S+|type \S+)',reason)
 return (f.get('kind'),f.get('phase'),m.group(2),m.group(1),m.group(3)) if m else (f.get('kind'),f.get('phase'),reason)
def read(row):
 if row['mode']=='scan':return json.loads((scratch/row['output']).read_text())
 return [json.loads(line) for line in (scratch/row['output']).read_text().splitlines() if line.strip()]
def flatten(records):
 findings=[];units={}
 for r in records:
  context=relative(r.get('file','<program>'))
  for u in r.get('units') or []:units[(context,relative(u['where']))]=u
  for f in r.get('findings') or []:
   findings.append({'context':context,'unitKey':[context,relative(f.get('unit','program'))],'tuple':list(fingerprint(f)),'finding':f})
 return findings,units
pairs=[];headFindings=[];overloadLedger=[]
index={(r['label'],r['mode'],r['file']):r for r in manifest if r['state']=='complete' and r.get('exit')==0}
for (label,mode,file),b in sorted(index.items()):
 if label!='base' or ('head',mode,file) not in index:continue
 h=index[('head',mode,file)]
 if mode=='scan':
  before,after=read(b),read(h)
  def sf(s):return (s.get('Kind',''),s.get('Rule',''),strip(s.get('Words','')))
  bc=collections.Counter(sf(s) for s in before.get('Stops') or []);hc=collections.Counter(sf(s) for s in after.get('Stops') or [])
  pairs.append({'mode':mode,'file':file,'scan':{'baseBoundary':before.get('Boundary'),'headBoundary':after.get('Boundary'),'baseCompleted':before.get('Completed'),'headCompleted':after.get('Completed'),'gone':[{'tuple':list(k),'count':n} for k,n in (bc-hc).items()],'new':[{'tuple':list(k),'count':n} for k,n in (hc-bc).items()]}});continue
 if not b.get('streamComplete') or not h.get('streamComplete'):continue
 br,hr=read(b),read(h);bf,bu=flatten(br);hf,hu=flatten(hr)
 headFindings.extend({'mode':mode,'selectedFile':file,**x} for x in hf)
 buckets=collections.defaultdict(lambda:{'base':[],'head':[]})
 for side,rows in [('base',bf),('head',hf)]:
  for row in rows:buckets[tuple(row['unitKey'])][side].append(row)
 changes=[];unchanged=0
 for unit,sides in sorted(buckets.items()):
  old=list(sides['base']);new=list(sides['head']);remaining=[]
  for item in old:
   match=next((i for i,x in enumerate(new) if x['tuple']==item['tuple']),None)
   if match is None:remaining.append(item)
   else:unchanged+=1;new.pop(match)
  old=remaining;used=set();remaining=[]
  for item in old:
   match=next((i for i,x in enumerate(new) if i not in used and family(x['finding'])==family(item['finding'])),None)
   if match is None:remaining.append(item)
   else:used.add(match);changes.append({'grade':'changed','unitKey':list(unit),'base':item,'head':new[match]})
  new=[x for i,x in enumerate(new) if i not in used]
  changes.extend({'grade':'gone','unitKey':list(unit),'base':x} for x in remaining)
  blocked=bool(sides['base'])
  for item in new:
   phase=item['finding'].get('phase','')
   reached=blocked and phase not in ('refusal_scan','contract_scan','deferred_scan')
   changes.append({'grade':'reached' if reached else 'new','unitKey':list(unit),'head':item,'baselineBlocked':blocked})
 pair={'mode':mode,'file':file,'baseHeader':br[0],'headHeader':hr[0],'headerIdentical':br[0]==hr[0],'baseRelationReport':br[-1],'headRelationReport':hr[-1],'baseFindings':len(bf),'headFindings':len(hf),'unchanged':unchanged,'counts':dict(collections.Counter(c['grade'] for c in changes)),'changes':changes,'unitStatuses':{'base':{k[1]:v.get('status') for k,v in bu.items()},'head':{k[1]:v.get('status') for k,v in hu.items()}}}
 assert pair['baseFindings']==unchanged+sum(c['grade'] in ('gone','changed') for c in changes)
 assert pair['headFindings']==unchanged+sum(c['grade'] in ('new','reached','changed') for c in changes)
 pairs.append(pair)
 for side,rows in [('base',bf),('head',hf)]:
  for row in rows:
   f=row['finding'];m=re.search(r'overload (\d+) of (\S+)',f.get('reason',''))
   if m:overloadLedger.append({'side':side,'mode':mode,'selectedFile':file,'set':m.group(2),'ordinal':int(m.group(1)),**row})
primary=[c for c in callers['programs']['allSrc']['calls'] if c['mapped'] and c['file'].startswith(('src/compiler/','src/services/'))]
assert len(primary)==17,len(primary)
location=lambda s:re.match(r'^(.*):(\d+):(\d+)$',relative(s or ''))
rows=[];allStorage=[]
for f in headFindings:
 if 'a certified overload whose implementation stores ' in f['finding'].get('reason',''):allStorage.append(f)
for c in primary:
 observations=[]
 for f in allStorage:
  loc=location(f['finding'].get('where'))
  if loc and (loc.group(1),int(loc.group(2)),int(loc.group(3)))==(c['file'],c['line'],c['column']):observations.append(f)
 words=sorted({f['finding']['reason'] for f in observations})
 rows.append({'call':c,'storageStopObserved':bool(observations),'words':words,'observations':observations})
expected=2+2*len(callers['fullCensusFiles'])+4
completed=sum(r.get('state')=='complete' and r.get('exit')==0 and (r['mode']=='scan' or r.get('streamComplete')) for r in manifest if r['mode']!='full-directory')
result={'coverage':{'expectedRequiredRuns':expected,'completeRequiredRuns':completed,'allRequiredComplete':completed==expected},'comparison':'unit-scoped multiset of (kind, phase, position-stripped reason, position-stripped full text)','grading':'same tuple unchanged; same unit and same overload claim family with different wording changed; remaining base gone; remaining head reached if unit was already blocked outside independent scan phases, otherwise new','pairs':pairs,'mapped17':rows,'storageStopCalls17':sum(r['storageStopObserved'] for r in rows),'allHeadStorageFindings':allStorage,'overloadLedger':overloadLedger}
(out/'paired.json').write_text(json.dumps(result,indent=2)+'\n')
(out/'overload-ledger.json').write_text(json.dumps(overloadLedger,indent=2)+'\n')
summary={'coverage':result['coverage'],'pairedRuns':len(pairs),'pairCounts':[{'mode':p['mode'],'file':p['file'],'base':p.get('baseFindings'),'head':p.get('headFindings'),'counts':p.get('counts',{})} for p in pairs],'storageStopCalls17':result['storageStopCalls17'],'mapped17':[{'file':r['call']['file'],'line':r['call']['line'],'column':r['call']['column'],'stopped':r['storageStopObserved'],'words':r['words']} for r in rows]}
(out/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))

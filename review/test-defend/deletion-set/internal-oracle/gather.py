from pathlib import Path
import subprocess,json,re,hashlib,gzip
root=Path('/tmp/deletion-replay');out=Path('review/test-defend/deletion-set/internal-oracle');out.mkdir(parents=True,exist_ok=True)
candidates=['TestCallTarget family','TestCheckedCastFlushesOutput','TestCheckedCastRunsNoCatchOrFinally','TestCheckedViewUntaggedSourceFlows','TestCheckedViewV2ArrayArmBoundary','TestCheckedViewV2MovedStage3Results','TestFileWritesLandInNodesOrder','TestImportCycleRuntimeCalls','TestInterfaceCastImportedConstruction','TestModuleNamespaceReadsMatchNode','TestNamespaceLiveExportBoundary','TestNumericEnumNeverPathsPinned','TestNumericEnumNeverPinned','TestOmittedOriginalProbePolicy','TestParserNamespace family','TestRegexCycleFixtureHasItsNativeDependency','TestStage3EnumSparseArrayBoundary','TestTypedArrayWriteStopIsPinned','TestWASIEmission']
families={'TestCallTarget family':['TestCallTargetThrowAgreesWithNode','TestDirectClosureCallAgreesWithNode'],'TestParserNamespace family':['TestParserNamespaceReceiver','TestParserNamespaceClass','TestParserCallableNamespace']}
def candidate(t):
 t=t.split('/')[0]
 for c in candidates:
  if t==c or t.startswith(c+' ') or t in families.get(c,[]):return c
 return None
def canon(s):
 m=re.match(r'([MD])0*(\d+)',str(s));return m.group(0) if m else None
def extract(obj,acc,mid=None,row=None):
 if isinstance(obj,list):
  for v in obj:extract(v,acc,mid,row)
 elif isinstance(obj,dict):
  own=next((canon(obj[k]) for k in ['mutant','id','mutant_id'] if k in obj and canon(obj[k])),None);mid=own or mid
  row=obj.get('test',row)
  if isinstance(row,str) and row.startswith('Test'):
   for m in obj.get('kills',[]):
    if canon(m):acc.setdefault(canon(m),set()).add(row)
  if mid:
   for k in ['failed','failed_rows','rows_failed','kills','failing_tests','failing_rows','failures']:
    v=obj.get(k)
    if isinstance(v,list):
     acc.setdefault(mid,set()).update(t for t in v if isinstance(t,str) and t.startswith('Test'))
   if obj.get('Action')=='fail' and str(obj.get('Test','')).startswith('Test'):acc.setdefault(mid,set()).add(obj['Test'])
   for k,v in obj.items():
    if k.startswith('Test') and (v=='fail' or v is False):acc.setdefault(mid,set()).add(k)
  for k,v in obj.items():
   if k in ['passed','passed_rows','passed_subtests','skipped','matrix_rows','events','diagnostics','evidence','last_proven_fail']: 
    if k=='events':extract(v,acc,mid,row)
    continue
   nm=canon(k) if re.match(r'^[MD]\d+(?:$|[-_.])',k) else None
   extract(v,acc,nm or mid,row)
records=[];notes=[]
for bi,b in enumerate(json.loads((root/'branches.json').read_text())):
 prefix='review/'+b+'/'
 files=subprocess.check_output(['git','ls-tree','-r','--name-only','origin/'+b,prefix]).decode().splitlines()
 metadata={}
 for f in files:
  if not f.endswith(('.json','.json.txt')):continue
  if not any(k in Path(f).name.lower() for k in ['matrix','result','row','attempt']):continue
  try:x=json.loads(subprocess.check_output(['git','show','origin/'+b+':'+f]))
  except:continue
  acc={};extract(x,acc);d=str(Path(f).parent)
  for mid,ts in acc.items():metadata.setdefault(d,{}).setdefault(mid,set()).update(ts)
 for f in files:
  n=Path(f).name
  if not re.match(r'^'+('M' if b.startswith('test-audit') else 'D')+r'\d+.*\.diff$',n):continue
  mid=canon(n);d=Path(f).parent;failed=set();sources=[]
  log_failures=set();log_events=False
  for lf in files:
   if Path(lf).parent!=d or not re.match(r'^'+re.escape(mid)+r'(?:[-.]|$)',Path(lf).name):continue
   if not lf.endswith(('.log','.log.gz')) or any(z in Path(lf).name.lower() for z in ['vet','build','witness']):continue
   raw=subprocess.check_output(['git','show','origin/'+b+':'+lf])
   if lf.endswith('.gz'):
    try:raw=gzip.decompress(raw)
    except:continue
   for line in raw.decode(errors='replace').splitlines():
    try:e=json.loads(line)
    except:continue
    if isinstance(e,dict) and e.get('Action') in ('pass','fail','skip') and e.get('Test'):
     log_events=True
     if e['Action']=='fail':log_failures.add(e['Test'])
  if log_events:sources=['matching raw Go JSON logs']

  while str(d).startswith(prefix.rstrip('/')):
   if mid in metadata.get(str(d),{}):failed=metadata[str(d)][mid];break
   alternate=next((k for k in metadata.get(str(d),{}) if k[0]==mid[0] and int(k[1:])==int(mid[1:])),None)
   if alternate:failed=metadata[str(d)][alternate];break
   if str(d)==prefix.rstrip('/'):break
   d=d.parent
  if log_events:failed=log_failures
  cs=sorted({candidate(t) for t in failed if candidate(t)})
  if not cs:
   notes.append({'branch':b,'diff':f,'mutant':mid,'recorded_failed':sorted(failed),'reason':'No candidate recorded failing' if failed else 'No failure mapping found'})
   continue
  data=subprocess.check_output(['git','show','origin/'+b+':'+f]);name=f'b{bi:02d}-{len(records):03d}-{n}'
  (out/'diffs').mkdir(exist_ok=True);(out/'diffs'/name).write_bytes(data)
  m=re.search(rb'^--- a/(.+)\n\+\+\+ b/.+\n@@ -(\d+)',data,re.M)
  check=subprocess.run(['git','apply','--check','-'],input=data,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
  records.append({'mutant':n[:-5],'canonical_id':mid,'branch':b,'source_diff':f,'saved_diff':'diffs/'+name,'file_line':m.group(1).decode()+':'+m.group(2).decode() if m else None,'candidates_failed':cs,'other_rows_failed':sorted(t for t in failed if not candidate(t)),'recorded_failed':sorted(failed),'failure_mapping_source':sources or ['nearest matrix/results/rows JSON'],'stale':check.returncode!=0,'apply_check':check.stdout.decode(),'sha256':hashlib.sha256(data).hexdigest()})
(out/'mutant-list.json').write_text(json.dumps(records,indent=2)+'\n');(out/'excluded-diffs.json').write_text(json.dumps(notes,indent=2)+'\n');(out/'candidates.json').write_text(json.dumps({'rows':candidates,'families':families},indent=2)+'\n');(out/'main.txt').write_text(subprocess.check_output(['git','rev-parse','HEAD']).decode());(out/'branches.json').write_text((root/'branches.json').read_text())
print('gathered',len(records),'stale',sum(x['stale'] for x in records),'excluded',len(notes))
for x in records:print(x['branch'],x['mutant'],x['candidates_failed'],'STALE' if x['stale'] else '')

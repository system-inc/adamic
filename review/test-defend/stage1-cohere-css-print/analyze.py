from pathlib import Path
import json,subprocess,difflib
p=Path('/workspace/adamic/review/test-defend/stage1-cohere-css-print')
menu=json.loads((p/'menu.json').read_text());(p/'planned-menu.json').write_text(json.dumps(menu,indent=2)+'\n')
used=[x for x in menu if x['mutant'] in ['D1','D2','D3','D4','D8']]
for x in used:
 if x['mutant']=='D8':x['line']=24
(p/'mutants.json').write_text(json.dumps(used,indent=2)+'\n')
for n in ['D5','D6','D7']:(p/(n+'.diff')).unlink(missing_ok=True)
def log(name):
 result={'passed':[],'failed':[],'errors':[],'seconds':None,'skipped':[]}
 for l in (p/(name+'.log')).read_text().splitlines():
  try:x=json.loads(l)
  except:continue
  t=x.get('Test');a=x.get('Action')
  if t and '/' not in t and a in ['pass','fail','skip']:result[{'pass':'passed','fail':'failed','skip':'skipped'}[a]].append(t)
  if x.get('OutputType')=='error':result['errors'].append(x.get('Output','').strip())
  if not t and a in ['pass','fail']:result['seconds']=x.get('Elapsed')
 return result
matrix={n:log(n) for n in ['D1','D2','D3','D4']};matrix['D8']={k:v for k,v in log('D8-snapshots').items()};matrix['D8']['passed']+=log('D8-artifacts')['passed']
if (p/'D8-other-rows.log').exists():
 other=log('D8-other-rows');matrix['D8']['passed']+=other['passed'];matrix['D8']['failed']+=other['failed']
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
rows=[]
for name,prior,sub,verdict,ids,unique in [
 ('TestOptionalBooleanPrinterMatchesGo','subsumed','TestCSSProfileSnapshotsAgree','not defended',['D2','D3','D4'],None),
 ('TestCSSProfileArtifacts','untrue',None,'defended',['D1'],'D1 internal/native/native.go:110'),
 ('TestCSSProfileSnapshotsAgree','subsumed','TestCSSPrinterThroughput','defended',['D8'],'D8 internal/native/runtime/count.c:24')]:
 attempts=[]
 for id in ids:
  m=next(x for x in used if x['mutant']==id)
  attempts.append({'mutant':id,'file_line':m['file']+':'+str(m['line']),'change':m['old']+' -> '+m['new'],'rows_failed':matrix[id]['failed']})
 id=ids[-1];err=next((x for x in matrix[id]['errors'] if ('print_test.go:317' if name.startswith('TestOptional') else 'profile_test.go:')in x),'none')
 command={'D4': "printf D4 > /tmp/css-defense/selector; ADAMIC_BUILD_CACHE_DIR=/tmp/css-defense/cache/D4 ADAMIC_CSS_PROFILE_SNAPSHOTS=/tmp/css-defense/profile-switched timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run '^(TestOptionalBooleanPrinterMatchesGo|TestCSSProfileSnapshotsAgree|TestCSSPrinterBoundaryProofs|TestClosedPrinterRegexGap)$'", 'D1': "ADAMIC_DEFENSE_MUTANT=D1 ADAMIC_BUILD_CACHE_DIR=/tmp/css-defense/cache/D1 ADAMIC_CSS_PROFILE_DIR=/tmp/css-defense/profile-D1 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run '^(TestCSSProfileArtifacts|TestOptionalBooleanPrinterMatchesGo|TestSharedSliceAppendAgreesWithNode|TestClosedPrinterRegexGap)$'", 'D8': "ADAMIC_BUILD_CACHE_DIR=/tmp/css-defense/cache/D8 ADAMIC_CSS_PROFILE_SNAPSHOTS=/tmp/css-defense/profile-D8 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/css/ -run '^TestCSSProfileSnapshotsAgree$'"}[id]
 rows.append({'test':name,'package':'stage1/cohere/css','prior_verdict':prior,'subsumed_by':sub,'defense':verdict,'unique_mutant':unique,'attempts':attempts,'evidence':command+'; '+err,'bounded':True,'scope_note':'Whole package exceeded 90 seconds. See matrix.json and counted-callers.txt. Other executions are unknown.'})
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
# V8 native-product code cannot be measured with Go's coverpkg.
def covered(name):
 d={}
 for script in json.loads((p/('v8-'+name+'.json')).read_text()):
  file='/'.join(script['url'].split('/')[-2:])
  for f in script['functions']:
   for b in f['ranges']:
    k=(file,f['functionName'],b['startOffset'],b['endOffset']);d[k]=d.get(k,0)+b['count']
 return d
s=covered('snapshots');t=covered('throughput-corrected')
(p/'v8-snapshots-vs-throughput-PARTIAL.json').write_text(json.dumps({'limitations':'Throughput timed out. These are observed differences, not proven exclusive coverage. TS offsets, not Go lines.','snapshot_only_observed':[dict(zip(['file','function','start','end','count'],[*k,v])) for k,v in s.items() if v>0 and t.get(k,0)==0]},indent=2)+'\n')
functions=sorted({k[:2] for k,v in s.items() if v>0}|{k[:2] for k,v in covered('optional').items() if v>0})
(p/'reached-port-functions.json').write_text(json.dumps([{'file':f,'function':n or '<module/anonymous>'} for f,n in functions],indent=2)+'\n')

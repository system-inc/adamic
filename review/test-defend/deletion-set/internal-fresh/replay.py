import json,subprocess,pathlib,time,os
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-defend/deletion-set/internal-fresh'; out.mkdir(parents=True,exist_ok=True)
c='TestNodeFSFileDoesNotEscapeBorrowedObjects'
def git(*a): return subprocess.check_output(['git',*a],cwd=root)
a=json.load(open('/tmp/replay-fresh/evidence/audit-mutants.json')); d=json.load(open('/tmp/replay-fresh/evidence/defender-matrix.json')); gathered=[]
for branch,base,rows,key in [('test-audit/fresh','review/test-audit/fresh',a,'failed_tests'),('test-defend/internal-fresh','review/test-defend/internal-fresh',d,'rows_failed')]:
 for r in rows:
  if c not in r.get(key,[]): continue
  ident=r.get('id',r.get('mutant')); diff=git('show',f'origin/{branch}:{base}/{ident}.diff'); (out/(ident+'.diff')).write_bytes(diff)
  gathered.append(dict(mutant=ident,branch=branch,source_file=base+'/'+ident+'.diff',candidates_failed=[c],other_rows_failed=[x for x in r[key] if x!=c],file_line=(r.get('file','')+':'+str(r.get('line',''))) if 'file' in r else 'internal/fresh/fresh.go:1333'))
(out/'mutant-list.json').write_text(json.dumps(gathered,indent=2)); print('FROZEN',[(r['mutant'],len(r['other_rows_failed'])) for r in gathered],flush=True)
main=git('rev-parse','HEAD').decode().strip(); results=[]
def run(ident):
 cmd=f"source /workspace/adamic-tools/env.sh; ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/replay-fresh/cache/{ident} go test -json -count=1 -timeout 30m ./internal/fresh/ -skip '^({c})($|_|/)'"
 start=time.monotonic()
 with open(out/(ident+'.log'),'w') as f: p=subprocess.run(['bash','-c',cmd],cwd=root,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for line in (out/(ident+'.log')).read_text().splitlines():
  try: events.append(json.loads(line))
  except: pass
 fail=sorted({e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']}); passed=sorted({e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']})
 r=dict(mutant=ident,command=cmd,status=p.returncode,wall_seconds=round(time.monotonic()-start,3),rows_failed=fail,rows_passed=passed,panics=[e.get('Test') for e in events if e.get('Output','').startswith('panic:')]); print(ident,r['status'],r['wall_seconds'],fail,flush=True); return r
baseline=run('baseline'); (out/'baseline.json').write_text(json.dumps(baseline,indent=2))
if baseline['status']: raise SystemExit('BASELINE FAILED')
for r in gathered:
 path=str(out/(r['mutant']+'.diff')); check=subprocess.run(['git','apply','--check',path],cwd=root,capture_output=True)
 if check.returncode: r['stale']=True; r['apply_error']=check.stderr.decode(); r['still_caught_by']=[]; continue
 r['stale']=False
 subprocess.run(['git','apply',path],cwd=root,check=True)
 try: result=run(r['mutant']); results.append(result); r['still_caught_by']=result['rows_failed']
 finally: subprocess.run(['git','apply','-R',path],cwd=root,check=True)
(out/'matrix.json').write_text(json.dumps(results,indent=2)); report=dict(package='internal/fresh',main=main,skipped=[c],mutants=gathered,keep=[],deletable=[c] if all(x.get('still_caught_by') and not x['stale'] for x in gathered) else [])
(out/'report.json').write_text(json.dumps(report,indent=2))

from pathlib import Path
import subprocess,os,json,time
root=Path('/workspace/adamic');out=Path('/tmp/def-pred');p=root/'internal/lower/predicates.go'
original=p.read_text();old='node.Text() == "undefined" && p.l.checker.GetSymbolAtLocation(node) == p.l.checker.GetUndefinedSymbol()';new='node.Text() == "null" && p.l.checker.GetSymbolAtLocation(node) == p.l.checker.GetUndefinedSymbol()'
assert original.count(old)==1
line=original[:original.index(old)].count('\n')+1
try:
 p.write_text(original.replace(old,new));(out/'D1.diff').write_bytes(subprocess.check_output(['git','diff','--','internal/lower/predicates.go'],cwd=root))
 with open(out/'D1-vet.log','w') as f:subprocess.run(['go','vet','./internal/lower/'],cwd=root,stdout=f,stderr=subprocess.STDOUT,check=True)
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/def-pred/cache/D1';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'];t=time.monotonic()
 with open(out/'D1.log','w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 events=[json.loads(s) for s in (out/'D1.log').read_text().splitlines() if s.startswith('{')]
 result=dict(mutant='D1',file_line='internal/lower/predicates.go:'+str(line),change='Change undefined literal recognition constant from "undefined" to "null".',rows_failed=sorted(set(x['Test'].split('/')[0] for x in events if x.get('Action')=='fail' and x.get('Test'))),rows_passed=sorted(x['Test'] for x in events if x.get('Action')=='pass' and x.get('Test') and '/' not in x['Test']),errors=[x['Output'].strip() for x in events if 'predicates_test.go:' in x.get('Output','')],returncode=r.returncode,seconds=time.monotonic()-t,command='ADAMIC_BUILD_CACHE_DIR=/tmp/def-pred/cache/D1 '+' '.join(cmd))
 (out/'D1-result.json').write_text(json.dumps(result,indent=2))
finally:p.write_text(original)

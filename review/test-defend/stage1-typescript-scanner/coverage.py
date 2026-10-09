from pathlib import Path
import subprocess,os,time,json
p=Path('/tmp/defend-scanner/evidence');env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/defend-scanner/typescript',ADAMIC_SCANNER_BENCH='1',ADAMIC_SCANNER_PROFILE_DIR='/tmp/defend-scanner/artifacts/baseline',ADAMIC_SCANNER_PROFILE_SNAPSHOTS='/tmp/defend-scanner/artifacts/baseline',ADAMIC_BUILD_CACHE_DIR='/tmp/defend-scanner/cache/coverage');runs=[]
for id,rx,cover in [('push','^TestGapStandsWhereGapsMdSays$','github.com/system-inc/adamic/internal/lower'),('bigint','^TestBigintGapStandsWhereGapsMdSays$','github.com/system-inc/adamic/internal/lower'),('snapshots','^TestProfileSnapshotsAgree$','github.com/system-inc/adamic/internal/native'),('agreement','^TestScannerAgreesWithTypescriptGo_','github.com/system-inc/adamic/internal/native')]:
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/scanner/','-run',rx,'-coverpkg='+cover,'-coverprofile='+str(p/(id+'.cover'))];s=time.monotonic()
 with (p/(id+'-coverage.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
 runs.append(dict(id=id,exit=r.returncode,seconds=time.monotonic()-s,command=cmd));(p/'coverage-runs.json').write_text(json.dumps(runs,indent=2));print(runs[-1],flush=True);assert r.returncode==0

def covered(id):
 return {s.split()[0]:s for s in (p/(id+'.cover')).read_text().splitlines()[1:] if int(s.split()[-1])>0}
result={}
for a,b in [('push','bigint'),('bigint','push'),('snapshots','agreement')]:
 ca,cb=covered(a),covered(b);exclusive=[ca[s] for s in sorted(ca.keys()-cb.keys())];(p/(a+'-exclusive.txt')).write_text('\n'.join(exclusive)+'\n');result[a]={'exclusive_blocks':len(exclusive),'row_covered_blocks':len(ca),'subsumer_covered_blocks':len(cb)}
(p/'coverage-summary.json').write_text(json.dumps(result,indent=2));print(result)

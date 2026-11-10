from pathlib import Path
import subprocess,json,sys,shutil,time
out=Path(__file__).resolve().parent;repo=out.parents[2]
def run(name,cmd):
 with (out/(name+'.log')).open('wb') as f:r=subprocess.run(cmd,cwd=repo,stdout=f,stderr=subprocess.STDOUT)
 print(name,'child exit code:',r.returncode,flush=True)
 if r.returncode:raise SystemExit(r.returncode)
assert (out/'recovery-monitor-complete.json').is_file(), 'exit observer has not completed both runs'
for side in ['parent','head']:
 p=out/('tsc-entry-ordered-first-'+side+'.timing.json')
 t=json.loads(p.read_text());assert t['child_exit_code']==0 and t['peak_rss_exact_exit_observation']
for src,name in [('/tmp/never-call-recovery-monitor.log','recovery-monitor.log'),('/tmp/never-call-ordered-entry-suite.log','ordered-launch-diagnostic.log'),('/tmp/never-call-ordered-head-wrapper.log','interrupted-head-launcher.log')]:
 if Path(src).exists():shutil.copyfile(src,out/name)
def stream(side):
 p=out/('tsc-entry-ordered-first-'+side+'.jsonl');return p if p.exists() else Path(str(p)+'.gz')
run('tsc-entry-ordered-first-pair',['python3',str(repo/'stage3/census/pair/pair.py'),str(stream('parent')),str(stream('head')),'/workspace/never-call-adapted','--typescript',str(repo/'stage3/api/node_modules/typescript'),'--json',str(out/'tsc-entry-ordered-first-paired.json'),'--report',str(out/'tsc-entry-ordered-first-paired.md')])
run('audit',['python3',str(out/'audit.py')])
run('render',['python3',str(out/'render.py')])
delta=json.loads((out/'tsc-entry-ordered-first-paired.json').read_text())
summary={}
for case in ['debug-full','semver-full','tsc-entry-ordered-first']:
 d=json.loads((out/(case+'-paired.json')).read_text())
 summary[case]={view:{k:len(d[view][k]) for k in ['lowered','moved','new_findings','new_sites','unit_first_stop_delta','regressions']} for view in ['raw_findings','distinct_reads']}
summary['entry_moved_units']=delta['distinct_reads']['unit_first_stop_delta']
summary['entry_regressions']=delta['distinct_reads']['regressions']
summary['audit_extra_new_site_regressions']=json.loads((out/'audit.json').read_text()).get('extra_new_site_regressions_distinct_reads',[])
(out/'SUMMARY.json').write_text(json.dumps(summary,indent=2)+'\n')
# No process started by this measurement may remain live after both exit observations.
alive=[]
for p in Path('/proc').iterdir():
 if not p.name.isdigit():continue
 try:
  exe=(p/'exe').resolve();stat=(p/'stat').read_text().rsplit(')',1)[1].split()
  if str(exe).startswith('/workspace/never-call-tools/') and stat[0]!='Z':alive.append(int(p.name))
 except (FileNotFoundError,PermissionError,ProcessLookupError):pass
assert not alive,alive
(out/'cleanup.json').write_text(json.dumps(dict(live_census_processes=alive,exit_observations_complete=True),indent=2)+'\n')
run('archive-before-hashes',['python3','/tmp/never-call-archive.py'])
run('stream-hash',['python3',str(out/'hash-streams.py')])
# Logs from these final steps must be closed before the last manifest is written.
with open('/tmp/never-call-final-archive.log','wb') as f:r=subprocess.run(['python3','/tmp/never-call-archive.py'],cwd=repo,stdout=f,stderr=subprocess.STDOUT)
print('final archive child exit code:',r.returncode,flush=True);assert r.returncode==0
with open('/tmp/never-call-final-sha-check.log','wb') as f:r=subprocess.run(['sha256sum','-c','SHA256SUMS'],cwd=out,stdout=f,stderr=subprocess.STDOUT)
print('sha256 verification child exit code:',r.returncode,flush=True);assert r.returncode==0
print(json.dumps(summary,indent=2))

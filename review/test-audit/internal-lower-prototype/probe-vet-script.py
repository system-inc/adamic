import pathlib,subprocess,json,difflib,time,shutil
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/internal-lower-prototype';prs=json.loads((p/'probes.json').read_text());runs=[]
for pr in prs:
 f=r/pr['file'];base=f.read_text();assert base==subprocess.check_output(['git','show','origin/main:'+pr['file']],cwd=r).decode();sig=pr['signature'];insert=sig+'\n\temptyAnswer := true\n\tif emptyAnswer { '+pr['answer']+' }\n';probe=base.replace(sig,insert,1)
 (p/(pr['id']+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),probe.splitlines(True),fromfile='a/'+pr['file'],tofile='b/'+pr['file'])))
 try:
  f.write_text(probe);start=time.monotonic()
  with (p/(pr['id']+'-vet.log')).open('w') as out:q=subprocess.run(['timeout','90','go','vet','./internal/lower/'],cwd=r,stdout=out,stderr=subprocess.STDOUT)
  runs.append(dict(id=pr['id'],exit=q.returncode,wall=time.monotonic()-start));assert q.returncode==0,pr['id']
 finally:f.write_text(base)
(p/'probe-vet-runs.json').write_text(json.dumps(runs,indent=2))
for name in ['finish','probe-vet']:
 shutil.copyfile('/tmp/u043-'+name+'.py',p/(name+'-script.py'))
report=p/'REPORT.md';s=report.read_text();s+='\nStandalone probe diffs were also vetted individually. Their local boolean is always true and exists to retain the original body without static unreachable-code warnings. The runtime-switch probe results implement the same entry return. Probe vet timings are saved in probe-vet-runs.json.\n';report.write_text(s)
print('all five standalone probes vetted')

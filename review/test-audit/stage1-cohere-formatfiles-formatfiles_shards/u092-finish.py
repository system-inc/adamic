import pathlib,json,subprocess,time,difflib
p=pathlib.Path('review/test-audit/stage1-cohere-formatfiles-formatfiles_shards');probes=[]
def run(id,regex):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/formatfiles/','-run',regex];start=time.monotonic()
 with (p/'logs'/(id+'.log')).open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
 es=[json.loads(l) for l in (p/'logs'/(id+'.log')).read_text().splitlines() if l.startswith('{')];return dict(id=id,command=' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-start,events=es)
for id,file,old,new,regex in [('W01','stage1/cohere/formatfiles/formatfiles_test.go','func firstDifference(got string, want string) string {','func firstDifference(got string, want string) string { if true {return ""};','^TestFormatfilesShardPlantedDisagreement$'),('S01','stage1/cohere/formatfiles/formatfiles_shards_test.go','const testThePortParsesAsGoCohereDoesShards = 32','const testThePortParsesAsGoCohereDoesShards = 31','^TestThePortParsesAsGoCohereDoesUnion$')]:
 f=pathlib.Path(file);before=f.read_text();assert before.count(old)==1;after=before.replace(old,new);diff=''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+file,tofile='b/'+file));(p/'probes'/(id+'.diff')).write_text(diff)
 try:
  f.write_text(after)
  with (p/'logs'/('validate-'+id+'.log')).open('w') as out:subprocess.run(['go','vet','./stage1/cohere/formatfiles/'],stdout=out,stderr=subprocess.STDOUT,check=True)
  q=run(id,regex);q.update(file=file,line=before[:before.index(old)].count('\n')+1,old=old,new=new);probes.append(q);print(id,q['exit'],flush=True)
 finally:f.write_text(before)
(p/'harness-probes.json').write_text(json.dumps(probes,indent=2))
r=run('restored','.');assert r['exit']==0;(p/'restored.json').write_text(json.dumps(r,indent=2))

import pathlib,sys,time,json,subprocess,os,signal
sys.dont_write_bytecode=True
follow=pathlib.Path(__file__).resolve().parent
binary=pathlib.Path('/workspace/scratch/baobab-cohere')
start=time.monotonic()
while not binary.exists():
 assert time.monotonic()-start<900,'cohere build did not finish'
 time.sleep(2)
label=sys.argv[1] if len(sys.argv)>1 else 'c'
root=pathlib.Path('/workspace/scratch/baobab-road2-'+label)
target=follow/label;target.mkdir(exist_ok=True)
for name in ['cohere-tsconfig.json','CohereSettings.json']:
 if label!='c':(target/name).write_text((follow/'c'/name).read_text().replace('/baobab-road2-c/','/baobab-road2-'+label+'/'))
names=['scanner.ts']+json.loads((follow/'variant-files.json').read_text())
cmd=[str(binary),'--lint','--no-format','--no-cache','--single-threaded','--json','--tsconfig',str(target/'cohere-tsconfig.json'),'--lint-config',str(target/'CohereSettings.json')]+[str(root/'src/compiler'/n) for n in names]
with (target/'cohere-rules.jsonl').open('wb') as log:
 start=time.monotonic();child=subprocess.Popen(cmd,stdout=log,stderr=subprocess.STDOUT,start_new_session=True);timeout=False
 try:child.wait(timeout=1500)
 except subprocess.TimeoutExpired:
  timeout=True;os.killpg(child.pid,signal.SIGTERM)
  try:child.wait(timeout=10)
  except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait()
status={'command':cmd,'child_exit':child.returncode,'timed_out':timeout,'wall_seconds':time.monotonic()-start,'scope_files':names,'rules':'cohere:adamic','type_graph':'79 compiler roots and transitive imports, exactOptionalPropertyTypes true'}
(target/'cohere-rules.status.json').write_text(json.dumps(status,indent=2)+'\n');print(json.dumps(status),flush=True)
assert not timeout and child.returncode in [0,1]

rows=[]
for line in (target/'cohere-rules.jsonl').read_text().splitlines():
 try:rows.append(json.loads(line))
 except ValueError:pass
assert any(r.get('kind')=='summary' and r.get('rules',0)>0 and r.get('checked',0)>0 and not r.get('error') for r in rows), 'cohere must actually run rules over files'

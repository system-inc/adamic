import pathlib,sys,json,subprocess,time,os,signal,gzip
sys.dont_write_bytecode=True
p=pathlib.Path(__file__).resolve().parent
for label in (sys.argv[1:] or ['original','c-public','c-public-unstripped']):
 target=p/label
 dts=target/'typescript.d.ts'
 if not dts.exists():dts.write_bytes(gzip.decompress(pathlib.Path(str(dts)+'.gz').read_bytes()))
 source=target/'public-adamic-consumer.ts'
 source.write_text('import * as ts from "./typescript.js";\nexport function probe(file: ts.SourceFile): ts.LineAndCharacter {\n    return ts.getLineAndCharacterOfPosition(file, 0);\n}\n')
 options=json.loads((p/'c/cohere-tsconfig.json').read_text())['compilerOptions']
 config=target/'public-adamic-tsconfig.json';config.write_text(json.dumps({'compilerOptions':options,'files':[str(source)]},indent=2)+'\n')
 settings=target/'public-adamic-CohereSettings.json';settings.write_text((p/'c/CohereSettings.json').read_text())
 cmd=['/workspace/scratch/baobab-cohere','--lint','--no-format','--no-cache','--single-threaded','--json','--tsconfig',str(config),'--lint-config',str(settings),str(source)]
 with (target/'public-adamic-cohere.jsonl').open('wb') as log:
  start=time.monotonic();child=subprocess.Popen(cmd,stdout=log,stderr=subprocess.STDOUT,start_new_session=True);timed=False
  try:child.wait(timeout=300)
  except subprocess.TimeoutExpired:
   timed=True;os.killpg(child.pid,signal.SIGTERM)
   try:child.wait(timeout=10)
   except subprocess.TimeoutExpired:os.killpg(child.pid,signal.SIGKILL);child.wait()
 status={'command':cmd,'child_exit':child.returncode,'timed_out':timed,'wall_seconds':time.monotonic()-start};(target/'public-adamic-cohere.status.json').write_text(json.dumps(status,indent=2)+'\n');print(json.dumps(status),flush=True)
 assert not timed and child.returncode in [0,1]
 rows=[]
 for line in (target/'public-adamic-cohere.jsonl').read_text().splitlines():
  try:rows.append(json.loads(line))
  except ValueError:pass
 assert any(r.get('kind')=='summary' and r.get('checked',0)>0 and r.get('rules',0)>0 and not r.get('error') for r in rows)

import pathlib,json,subprocess,time,os
root=pathlib.Path('/workspace/adamic');p=pathlib.Path('/tmp/defend-markdown');out=root/'review/test-defend/stage1-cohere-markdownblocks-structure_layout_shards';plans=json.loads((out/'planned-mutants.json').read_text());runs=[]
def run(label,names,env):
 regex='^('+ '|'.join(names)+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',regex];start=time.monotonic();log=p/(label+'.log')
 with log.open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in log.read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 item=dict(label=label,command=' '.join(cmd),cache=env['ADAMIC_BUILD_CACHE_DIR'],status=r.returncode,wall=time.monotonic()-start,cooked=any('test timed out' in e.get('Output','')for e in events),failed=[e['Test']for e in events if e.get('Action')=='fail' and'Test'in e],passed=[e['Test']for e in events if e.get('Action')=='pass'and'Test'in e]);runs.append(item);(out/'remaining-runs.json').write_text(json.dumps(runs,indent=2));print(item,flush=True);return item
while not (p/'width-done').exists():time.sleep(1)
env=os.environ.copy();env['ADAMIC_NATIVE_SPLIT']='1';env['ADAMIC_MARKDOWNWIDTH_DEPS']=str(p/'width-deps')
seen=set()
for log in p.glob('D1-*.log'):
 for l in log.read_text().splitlines():
  try:e=json.loads(l)
  except:continue
  if e.get('Action')in ['pass','fail','skip']and'Test'in e:seen.add(e['Test'])
names=json.loads((out/'potential-matrix-rows.json').read_text());remaining=[n for n in names if n not in seen];(out/'D1-remaining-before-replay.json').write_text(json.dumps(remaining,indent=2))
m=plans[0];f=root/m['file'];s=f.read_text()
try:
 f.write_text(s.replace(m['old'],m['new']));env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/'D1')
 # Replay completed-cache setups/products separately from any cooked functional leaves.
 for group in [[n for n in remaining if not n.startswith('TestProduct')],[n for n in remaining if n.startswith('TestProduct')]]:
  if group:run('D1-remaining-'+str(len(runs)),group,env)
finally:f.write_text(s)
# Additional current package competitor family for each of the shared layout attempts.
leaf=[n for n in names if n.startswith('TestMarkdownLeafComposition_') and n[-3:].isdigit()]
for m in plans[1:]:
 f=root/m['file'];s=f.read_text();env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/m['id'])
 try:
  f.write_text(s.replace(m['old'],m['new']));run(m['id']+'-new-leaf-family',leaf,env)
 finally:f.write_text(s)
(p/'remaining-done').write_text('done')

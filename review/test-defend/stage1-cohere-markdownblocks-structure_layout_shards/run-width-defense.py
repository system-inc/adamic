import pathlib,json,time,subprocess,os,re
root=pathlib.Path('/workspace/adamic');p=pathlib.Path('/tmp/defend-markdown');out=root/'review/test-defend/stage1-cohere-markdownblocks-structure_layout_shards';runs=[]
def run(label,regex,env,cover=False):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s']
 if cover:cmd+=['-coverpkg=./internal/load,./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(p/'width-isolated.cover')]
 cmd+=['./stage1/cohere/markdownblocks/','-run',regex];start=time.monotonic();log=p/(label+'.log')
 with log.open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in log.read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 result=dict(label=label,command=' '.join(cmd),cache=env.get('ADAMIC_BUILD_CACHE_DIR'),wall=time.monotonic()-start,status=r.returncode,cooked=any('test timed out' in e.get('Output','')for e in events),failed=[e['Test']for e in events if e.get('Action')=='fail' and 'Test'in e],passed=[e['Test']for e in events if e.get('Action')=='pass' and 'Test'in e]);runs.append(result);(out/'width-defense-runs.json').write_text(json.dumps(runs,indent=2));print(result,flush=True);return result
while not (p/'layout-done').exists():time.sleep(1)
env=os.environ.copy();env['ADAMIC_NATIVE_SPLIT']='1';env['ADAMIC_MARKDOWNWIDTH_DEPS']=str(p/'width-deps');env.pop('NODE_V8_COVERAGE',None)
b=run('width-clean-isolated','^TestMarkdownUnicodeWidths$',env,True)
if b['status'] and not b['cooked']:raise RuntimeError('red baseline; stop')
f=root/'stage1/cohere/markdownblocks/width.ts';original=f.read_text();m=json.loads((out/'planned-mutants.json').read_text())[0]
try:
 f.write_text(original.replace(m['old'],m['new']));env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/'D1')
 cmd=['timeout','120','go','run','./cmd/adamic','build','stage1/cohere/markdownblocks/testdata/width_probe.ts','-o',str(p/'D1-native'),'--sanitize'];start=time.monotonic()
 with (p/'D1-native-build.log').open('w') as target:r=subprocess.run(cmd,env=env,stdout=target,stderr=subprocess.STDOUT)
 (out/'D1-native-build.json').write_text(json.dumps(dict(command=' '.join(cmd),status=r.returncode,wall=time.monotonic()-start),indent=2))
 if r.returncode:raise RuntimeError('D1 did not compile')
 run('D1-width','^TestMarkdownUnicodeWidths$',env)
 for family in ['Structure','Table','LeafComposition','Quote','List','Whitespace']:
  regex='^TestMarkdown'+family+'Layout(Union|_[0-9]{3})$' if family!='LeafComposition' else '^TestMarkdownLeafComposition(Union|_[0-9]{3})$'
  r=run('D1-'+family,regex,env)
  if r['cooked']:run('D1-'+family+'-retry',regex,env)
 run('D1-other-layouts','^TestMarkdown(CodeBlockLayout|HTMLBlockLayout|RootLayout)$',env)
finally:f.write_text(original)
(p/'width-done').write_text('done')

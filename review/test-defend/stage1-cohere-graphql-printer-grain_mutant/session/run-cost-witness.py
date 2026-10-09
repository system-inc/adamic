import pathlib,subprocess,os,json
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-cohere-graphql-printer-grain_mutant/session';tmp=pathlib.Path('/tmp/printer-defense/witness');tmp.mkdir(parents=True,exist_ok=True)
files=['stage1/cohere/graphql/'+x.name for x in (root/'stage1/cohere/graphql').glob('*.ts')]+['stage1/cohere/graphql/printer/'+x for x in ['main.ts','doc.ts','printer.ts']]+['stage1/cohere/json/width.ts','stage1/cohere/json/widthTables.ts']
(tmp/'cases.txt').write_text('>query{hello(a:1,b:2)}\n')
results=[]
for mid in ['clean','D1','D2','D3']:
 folder=tmp/mid;folder.mkdir(exist_ok=True)
 for f in files:
  out=folder/f;out.parent.mkdir(parents=True,exist_ok=True);out.write_bytes(subprocess.check_output(['git','show','HEAD:'+f],cwd=root))
 if mid!='clean':subprocess.run(['git','apply','--unsafe-paths',str(p/(mid+'.diff'))],cwd=folder,check=True)
 coverage=p/('witness-v8-'+mid);env=os.environ.copy();env['NODE_V8_COVERAGE']=str(coverage)
 cmd=['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(folder/'stage1/cohere/graphql/printer/main.ts'),'--cases',str(tmp/'cases.txt')]
 r=subprocess.run(cmd,env=env,capture_output=True);(p/(mid+'-witness.stdout')).write_bytes(r.stdout);(p/(mid+'-witness.stderr')).write_bytes(r.stderr)
 functions=[]
 for f in coverage.glob('*.json'):
  for sc in json.loads(f.read_text())['result']:
   if sc['url'].endswith('/cohere/json/width.ts'):
    functions.extend([dict(function=fn['functionName'],count=fn['ranges'][0]['count']) for fn in sc['functions']])
 results.append(dict(mutant=mid,command=cmd,exit=r.returncode,stdout=r.stdout.decode(),width_functions=functions))
(p/'cost-witness.json').write_text(json.dumps(results,indent=2));print([(x['mutant'],x['exit'],[f for f in x['width_functions'] if f['function']=='stringWidth']) for x in results]);assert len(set(x['stdout'] for x in results))==1 and all(x['exit']==0 for x in results)

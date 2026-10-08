import json,subprocess,time,hashlib,os
from pathlib import Path
root=Path('/workspace/wave-23/fixes-problem-cases')
records=json.loads((root/'captured.json').read_text());results=[];start=time.monotonic()
runner='/workspace/adamic/oracle/node.mjs';source='/workspace/adamic/stage1/cohere/lint/main.ts'
def run(name,args,directory):
 before=time.monotonic()
 with (directory/(name+'.stdout')).open('wb') as out,(directory/(name+'.stderr')).open('wb') as err:
  proc=subprocess.run(args,stdout=out,stderr=err,timeout=120)
 data=(directory/(name+'.stdout')).read_bytes();errors=(directory/(name+'.stderr')).read_bytes()
 return {'exit':proc.returncode,'seconds':time.monotonic()-before,'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest(),'stderr_bytes':len(errors)}
with (root/'events.jsonl').open('w') as events:
 for index,row in enumerate(records):
  directory=root/('case-%04d'%index);directory.mkdir(exist_ok=True)
  path=directory/row['file'].lstrip('/');path.parent.mkdir(parents=True,exist_ok=True);path.write_text(row['source'])
  config=directory/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True},'files':[str(path)]}))
  options=row.get('options');raw=json.dumps(options,separators=(',',':'),ensure_ascii=True)
  fields=[str(path),row['rule'],'','','false',raw,'']
  manifest=directory/'manifest.txt';manifest.write_text('program '+str(config)+'\n'+'\t'.join(fields)+'\n')
  result={'index':index,'rule':row['rule'],'source':str(path),'options':options,'status':'pending'}
  try:
   result['Go']=run('go',[str(root/'go-oracle'),'--manifest',str(manifest)],directory)
   if result['Go']['exit']!=0 or result['Go']['stderr_bytes']:
    result['status']='blocked-oracle';results.append(result);events.write(json.dumps(result)+'\n');events.flush();continue
   transcript=str(directory/'transcript')
   result['native']=run('native',[str(root/'native-asan'),'--manifest',str(manifest),'--record',transcript],directory)
   result['Node']=run('node',['node','--disable-warning=ExperimentalWarning',runner,source,'--manifest',str(manifest),'--replay',transcript],directory)
   result['emitted']=run('emitted',['node','--disable-warning=ExperimentalWarning',runner,str(root/'emitted.mjs'),'--manifest',str(manifest),'--replay',transcript],directory)
   expected=(directory/'go.stdout').read_bytes()
   good=all(result[side]['exit']==0 and result[side]['stderr_bytes']==0 and (directory/(file+'.stdout')).read_bytes()==expected for side,file in [('native','native'),('Node','node'),('emitted','emitted')])
   result['status']='pass' if good else 'fail'
  except Exception as error:result['status']='error';result['error']=str(error)
  results.append(result);events.write(json.dumps(result)+'\n');events.flush()
(root/'result.json').write_text(json.dumps({'wall_seconds':time.monotonic()-start,'cases':results},indent=2)+'\n')

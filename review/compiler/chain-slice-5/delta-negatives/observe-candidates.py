import subprocess,json,pathlib,hashlib
root=pathlib.Path('/tmp/delta-negatives-slice5');out=root/'review/compiler/chain-slice-5/delta-negatives';binary='/tmp/delta-negatives-slice-adamic'
rows=json.loads((root/'review/compiler/chain-slice-5/admission-proof.json').read_text())['rows'];rows=[r for r in rows if not r['source_node_equality']]
ledger=json.loads(pathlib.Path('/workspace/adamic/cloud/admission-corpus/negative-witnesses.json').read_text());results=[]
messages={'alias_reset':"placeholder 'value' is unset at assignment via source.value",'assignment_result':"placeholder 'target' is unset at assignment result via target = value",'before_use':"placeholder 'value' is unset at argument via value",'null_before_use':"placeholder 'value' is unset at argument via value",'null_saved_leak':"placeholder 'value' is unset at argument via saved",'return_assignment':"placeholder 'target' is unset at return via target = value",'saved_leak':"placeholder 'value' is unset at argument via saved"}
def run(args):
 p=subprocess.run(args,cwd=root,capture_output=True,text=True,timeout=45);return {'exit':p.returncode,'stdout':p.stdout,'stderr':p.stderr}
for i,row in enumerate(rows):
 path=row['program'];source=(root/path).read_bytes();record={'path':path,'sha256':hashlib.sha256(source).hexdigest()}
 record['node']=run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(root/path)])
 js=run([binary,'js',path]);assert js['exit']==0,js
 module=out/'logs'/f'program-{i}.mjs';module.write_text(js['stdout']);record['javascript']=run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(module)])
 executable=f'/tmp/delta-negatives-observe-{i}';built=run([binary,'build',path,'-o',executable]);assert built['exit']==0,built;record['native']=run([executable])
 n,j,c=(record[k] for k in ['node','javascript','native']);agreement=n['exit']==j['exit']==c['exit'] and n['stdout']==j['stdout']==c['stdout'];record['agreement']=agreement
 if not agreement:
  assert n['exit']==0 and j['exit']==c['exit']==70,(path,record)
  if path.startswith('internal/oracle'):
   suffix=path.rsplit('placeholder_nonnull_',1)[1][:-2];expected='adamic: panic: '+messages[suffix]+'\n';from_text='null!' if 'null_' in suffix else 'undefined!';to_text='"ready"';decl=next(k+1 for k,line in enumerate(source.decode().splitlines()) if ': string' in line and from_text in line);typ='string';task='#9wc5q5j'
  else:
   expected='adamic: panic: speculative result failed: callback misfit at {root}/stage3/parser-next/speculation/misfit.a:3:14; expected number\n';from_text="return 'wrong';";to_text='return 1;';decl=3;typ='number';task='#ktz9fek'
  assert j['stderr']==c['stderr']==expected.replace('{root}',str(root)),record
  assert j['stdout']==c['stdout'],record
  ledger['witnesses'][record['sha256']]={'ruling_task':task,'declaration':f'{path}:{decl}','declared_type':typ,'repair_from':from_text,'repair_to':to_text,'type_correct_sha256':hashlib.sha256(source.replace(from_text.encode(),to_text.encode())).hexdigest(),'exit':70,'stderr':expected}
 record['listed']=not agreement;results.append(record)
 print(path,'agreement' if agreement else 'ruled witness',flush=True)
(out/'candidate-observations.json').write_text(json.dumps(results,indent=2)+'\n')
(root/'cloud/admission-corpus').mkdir(parents=True,exist_ok=True);(root/'cloud/admission-corpus/negative-witnesses.json').write_text(json.dumps(ledger,indent=2)+'\n')
assert len(ledger['witnesses'])==8

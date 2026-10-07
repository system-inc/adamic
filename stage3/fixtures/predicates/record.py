import json,subprocess,pathlib,tempfile
root=pathlib.Path(__file__).resolve().parents[3]; bucket=pathlib.Path(__file__).resolve().parent; logs=bucket/'logs';logs.mkdir(exist_ok=True)
rows=json.loads((bucket/'manifest.json').read_text())
scratch=tempfile.TemporaryDirectory(prefix='predicates-native-')
native_root=pathlib.Path(scratch.name)
for row in rows:
 f='stage3/fixtures/predicates/'+row['file']
 def run(cmd,name):
  with open(logs/(name+'.stdout'),'wb') as out,open(logs/(name+'.stderr'),'wb') as err:r=subprocess.run(cmd,cwd=root,stdout=out,stderr=err)
  return {'stdout':(logs/(name+'.stdout')).read_text(),'stderr':(logs/(name+'.stderr')).read_text(),'exit':r.returncode}
 row['node']=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',f],row['file']+'.node')
 stage=run(['go','run','./cmd/adamic','build',f,'-o',str(native_root/(row['file']+'.bin'))],row['file']+'.stage0')
 what=stage['stdout']+stage['stderr'];outcome='Compiles' if stage['exit']==0 else 'Refused' if 'refus' in what.lower() else 'NotYet' if 'NotYet' in what or 'not yet' in what else 'Checker'
 row['stage0']={'outcome':outcome,'what':what if stage['exit'] else ''}
 if not stage['exit']:
  native=run([str(native_root/(row['file']+'.bin'))],row['file']+'.native')
  if native!=row['node']:raise Exception('SILENT MISCOMPILE '+row['file'])
 print(row['file'],row['node'],outcome,flush=True)
(bucket/'status.json').write_text(json.dumps(rows,indent=2)+'\n')
scratch.cleanup()

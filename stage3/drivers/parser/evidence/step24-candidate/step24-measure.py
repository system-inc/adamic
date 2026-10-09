import pathlib,subprocess,os,json,time,hashlib
out=pathlib.Path('/tmp/step24-candidate-evidence/builds');out.mkdir();entry='/tmp/step24-candidate-slice/parser-proof-main.a';compiler='/tmp/step24-candidate-adamic';scratch='/tmp/step24-candidate';rows=[]
for split in ['0','1']:
 binary=out/('parser-split-'+split);env=dict(os.environ,ADAMIC_NATIVE_SPLIT=split,ADAMIC_NATIVE_JOBS='5');started=time.monotonic()
 with (out/(split+'.stdout')).open('wb') as stdout,(out/(split+'.stderr')).open('wb') as stderr:
  result=subprocess.run([compiler,'build',entry,'-o',str(binary)],cwd=scratch,env=env,stdout=stdout,stderr=stderr,timeout=90)
 row={'split':int(split),'command':[compiler,'build',entry,'-o',str(binary)],'exit':result.returncode,'attempt_wall_seconds':time.monotonic()-started,'binary':str(binary) if binary.exists() else None,'clang_seconds':None};rows.append(row)
 if result.returncode==0:
  with (out/(split+'.native.stdout')).open('wb') as stdout,(out/(split+'.native.stderr')).open('wb') as stderr:
   native=subprocess.run([str(binary),'/tmp/step24-candidate-corpus/corpus','/tmp/step24-candidate-corpus/manifest'],cwd=scratch,stdout=stdout,stderr=stderr,timeout=120)
  compared=subprocess.run(['/tmp/step24-dumpdiff','/tmp/step24-candidate-corpus/node.dump',str(out/(split+'.native.stdout'))],capture_output=True);(out/(split+'.comparison.log')).write_bytes(compared.stdout+compared.stderr)
  row['native_exit']=native.returncode;row['comparison_exit']=compared.returncode;row['stderr_equal']=(out/(split+'.native.stderr')).read_bytes()==b''
  assert native.returncode==0 and compared.returncode==0 and row['stderr_equal'],'NATIVE PARSER MISMATCH'
(out/'report.json').write_text(json.dumps({'compiler':'9d09100b623ce6923c74da78eed81f1a99d28f70','attempts':rows},indent=2)+'\n');print(json.dumps(rows,indent=2))

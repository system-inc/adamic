import pathlib,subprocess,json
root=pathlib.Path('/workspace/adamic/stage3/drivers/parser');out=pathlib.Path('/tmp/step24-candidate-evidence/probes');out.mkdir();previous=json.load(open('/tmp/step24-parser-proof/stage3/drivers/parser/evidence/front34/comparison.json'));rows=[]
for prior in previous['probes']:
 name=prior['file'];file=root/name;binary=out/(name+'.bin');row={'file':name,'previous_pass':prior['current_pass']}
 for stage,command in [('node',['node','--disable-warning=ExperimentalWarning','/workspace/adamic/oracle/node.mjs',str(file)]),('build',['/tmp/step24-candidate-adamic','build',str(file),'-o',str(binary)])]:
  run=subprocess.run(command,cwd='/tmp/step24-candidate',capture_output=True,timeout=60);(out/(name+'.'+stage+'.stdout')).write_bytes(run.stdout);(out/(name+'.'+stage+'.stderr')).write_bytes(run.stderr);row[stage]={'exit':run.returncode,'stdout':run.stdout.decode(),'stderr':run.stderr.decode()}
 assert row['node']['exit']==0 and row['node']['stderr']=='',name
 row['pass']=False
 if row['build']['exit']==0:
  native=subprocess.run([str(binary)],capture_output=True,timeout=30);(out/(name+'.native.stdout')).write_bytes(native.stdout);(out/(name+'.native.stderr')).write_bytes(native.stderr);row['native']={'exit':native.returncode,'stdout':native.stdout.decode(),'stderr':native.stderr.decode()};row['pass']=row['node']==row['native'];assert row['pass'],'SILENT MISCOMPILE '+name
  assert native.stdout;mutant=out/(name+'.mutant.stdout');mutant.write_bytes(bytes([native.stdout[0]^1])+native.stdout[1:]);caught=subprocess.run(['/tmp/step24-dumpdiff',str(out/(name+'.node.stdout')),str(mutant)],capture_output=True);(out/(name+'.mutant-comparison.log')).write_bytes(caught.stdout+caught.stderr);assert caught.returncode==1;row['byte_mutant_exit']=1
 row['status_vs_previous']='gone' if row['pass'] and not row['previous_pass'] else 'new' if not row['pass'] and row['previous_pass'] else 'still there' if not row['pass'] else 'still green';rows.append(row)
(out/'report.json').write_text(json.dumps({'probes':rows,'total':len(rows),'passing':sum(r['pass'] for r in rows)},indent=2)+'\n');print('\n'.join(r['file']+' '+r['status_vs_previous'] for r in rows))

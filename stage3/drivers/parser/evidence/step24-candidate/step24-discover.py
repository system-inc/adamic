import pathlib,subprocess,json,re,shutil
root=pathlib.Path('/tmp/step24-candidate-discovery');shutil.copytree('/tmp/step24-candidate-slice',root);out=pathlib.Path('/tmp/step24-candidate-evidence/discovery');out.mkdir();rows=[]
for index in range(15):
 run=subprocess.run(['/tmp/step24-candidate-adamic','c',str(root/'parser-proof-main.a')],cwd='/tmp/step24-candidate',capture_output=True,timeout=60);(out/f'{index+1:02}.stdout').write_bytes(run.stdout);(out/f'{index+1:02}.stderr').write_bytes(run.stderr);text=run.stderr.decode();row={'order':index+1,'exit':run.returncode,'behind_discovery_placeholders':index>0};rows.append(row)
 if run.returncode==0:row['c_emitted']=True;break
 match=re.search(r'(/.*?):(\d+):(\d+): (.*)',text)
 if not match:row['unresolved']='failure without source location';break
 file,line,col,message=match.groups();row.update(file=str(pathlib.Path(file).relative_to(root)),line=int(line),column=int(col),message=message,diagnostics=text)
 if 'error TS' in message and index>0:row['unresolved']='Discovery introduced checker diagnostics; excluded from original-source blocker count';break
 if 'index signature' in message:
  row['unresolved']='Structural declaration stop, no semantics-preserving throwing placeholder; stop here';break
 command=['node','/tmp/step24-parser-proof/stage3/drivers/parser/evidence/front34/error-placeholder.cjs',file,line,col];stub=subprocess.run(command,capture_output=True);(out/f'{index+1:02}-stub.stdout').write_bytes(stub.stdout);(out/f'{index+1:02}-stub.stderr').write_bytes(stub.stderr);row['stub_exit']=stub.returncode
 if stub.returncode:row['unresolved']='No safe discovery-only function/initializer placeholder; stop here';break
 row['placeholder']=json.loads(stub.stdout)
(out/'report.json').write_text(json.dumps({'warning':'Discovery only. Unmodified native attempts are separate; no Node/native acceptance for this edited copy.','rows':rows},indent=2)+'\n')
for r in rows:print(r['order'],r.get('file'),r.get('line'),r.get('message'),r.get('unresolved',''))

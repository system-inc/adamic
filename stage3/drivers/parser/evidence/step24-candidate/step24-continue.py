import pathlib,json,subprocess,re
root=pathlib.Path('/tmp/step24-candidate-discovery');out=pathlib.Path('/tmp/step24-candidate-evidence/discovery');report=json.loads((out/'report.json').read_text());report['excluded_attempts']=[report['rows'].pop()];p=root/'src/compiler/performanceCore.ts';text=p.read_text();needle='function tryGetPerformance()';assert needle in text;return_type=pathlib.Path('/tmp/step24-candidate-evidence/original-return-type.log').read_text().strip();p.write_text(text.replace(needle,needle+': '+return_type,1));report['discovery_type_preservation']={'function':'tryGetPerformance','original_inferred_return':return_type,'reason':'Retain the original inferred return when replacing its body by throw; no inference from the placeholder.'}
for index in range(4,15):
 run=subprocess.run(['/tmp/step24-candidate-adamic','c',str(root/'parser-proof-main.a')],cwd='/tmp/step24-candidate',capture_output=True,timeout=60);(out/f'{index+1:02}-continued.stdout').write_bytes(run.stdout);(out/f'{index+1:02}-continued.stderr').write_bytes(run.stderr);text=run.stderr.decode();row={'order':index+1,'exit':run.returncode,'behind_discovery_placeholders':True};report['rows'].append(row)
 if run.returncode==0:row['c_emitted']=True;break
 match=re.search(r'(/.*?):(\d+):(\d+): (.*)',text)
 if not match:row['unresolved']='failure without source location';break
 file,line,col,message=match.groups();row.update(file=str(pathlib.Path(file).relative_to(root)),line=int(line),column=int(col),message=message,diagnostics=text)
 if 'error TS' in message:row['unresolved']='Discovery checker artifact; excluded';report.setdefault('excluded_attempts',[]).append(report['rows'].pop());break
 if 'index signature' in message:row['unresolved']='Original structural declaration stop; cannot bypass without changing the representation';break
 stub=subprocess.run(['node','/tmp/step24-parser-proof/stage3/drivers/parser/evidence/front34/error-placeholder.cjs',file,line,col],capture_output=True);(out/f'{index+1:02}-continued-stub.stdout').write_bytes(stub.stdout);(out/f'{index+1:02}-continued-stub.stderr').write_bytes(stub.stderr);row['stub_exit']=stub.returncode
 if stub.returncode:row['unresolved']='No valid discovery placeholder';break
 row['placeholder']=json.loads(stub.stdout)
(out/'report.json').write_text(json.dumps(report,indent=2)+'\n')
for r in report['rows']:print(r['order'],r.get('file'),r.get('line'),r.get('message'),r.get('unresolved',''))

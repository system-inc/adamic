import pathlib,subprocess,time,json,os
p=pathlib.Path('review/test-audit/stage1-cohere-formatfiles-formatfiles_products');work=pathlib.Path('/tmp/u091/validation/control');base=['stage1/cohere/formatfiles/'+n for n in ['main.ts','enumerate.ts','disk.ts','golang.ts']]+['stage1/cohere/gitignore/'+n for n in ['path.ts','glob.ts','gitignore.ts']]
for f in base:
 t=work/f;t.parent.mkdir(parents=True,exist_ok=True);t.write_bytes(subprocess.check_output(['git','show','HEAD:'+f]))
root=pathlib.Path('/tmp/u091/witness-tree');root.mkdir(exist_ok=True);(root/'hello.ts').write_text('const a = 1;\n');case=pathlib.Path('/tmp/u091/witness-cases');case.write_text('handles\t.ts\ntree\twitness\t'+str(root)+'\nhouse\t0\t\n');(p/'witness-cases.txt').write_text(case.read_text());rs=[]
for label,cmd in [('control-build',['go','run','./.audit-u091',str(work/'stage1/cohere/formatfiles/main.ts'),str(work/'port')]),('control-run',[str(work/'port'),str(case)]),('empty-run',['/tmp/u091/validation/P01/port',str(case)])]:
 t=time.monotonic()
 with (p/('witness-'+label+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 rs.append(dict(label=label,command=cmd,exit=r.returncode,wall_seconds=time.monotonic()-t))
(p/'empty-port-witness.json').write_text(json.dumps(rs,indent=2));print(rs)

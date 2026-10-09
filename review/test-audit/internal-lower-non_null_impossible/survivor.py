import pathlib,subprocess,os,json,time
p=pathlib.Path('review/test-audit/internal-lower-non_null_impossible');samples={
'conditional':'const source: { x: number } = { x: 1 }; const wider: { x: number; y?: number } = { x: 2 }; const result = true ? source : wider;',
'coalesce':'function choose(source: { x: number } | undefined, wider: { x: number; y?: number }): { x: number } { const result = source ?? wider; return result; }',
'inferred_array':'const source: { x: number } = { x: 1 }; const wider: { x: number; y?: number } = { x: 2 }; const result = [source, wider];',
'arrow':'const source: { x: number } = { x: 1 }; const wider: { x: number; y?: number } = { x: 2 }; const choose = () => true ? source : wider;'}
records=[]
for name,text in samples.items():
 s=pathlib.Path('/tmp/u040-M10-'+name+'.a');s.write_text(text+'\n');(p/('M10.'+name+'.a.txt')).write_bytes(s.read_bytes())
 for label,id in [('before',''),('after','M10')]:
  env=os.environ.copy();env['ADAMIC_MUTANT']=id;t=time.monotonic()
  with (p/('M10.'+name+'.'+label+'.log')).open('w') as f:r=subprocess.run(['/tmp/u040-adamic','js',str(s)],env=env,stdout=f,stderr=subprocess.STDOUT)
  records.append(dict(command='ADAMIC_MUTANT='+id+' /tmp/u040-adamic js '+str(s),sample=name,label=label,exit=r.returncode,wall=time.monotonic()-t))
 print(name,[(r['label'],r['exit']) for r in records if r['sample']==name],(p/('M10.'+name+'.before.log')).read_text()[:220],(p/('M10.'+name+'.after.log')).read_text()[:220],flush=True)
(p/'M10.witness-commands.json').write_text(json.dumps(records,indent=2))

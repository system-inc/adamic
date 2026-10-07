"""Named declaration contract against production listener maps, without parser adapters."""
import pathlib,json,subprocess,time,hashlib
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-listeners');work.mkdir(exist_ok=True)
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 elapsed=time.monotonic()-start;print(name,r.returncode,round(elapsed,6),flush=True)
 assert r.returncode==0,(name,(work/(name+'.stderr')).read_text()[:2000])
 return (work/(name+'.stdout')).read_bytes(),elapsed
virtual=repo/'cohere/wave09_listener_oracle.go';overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/oracle.go')}}))
go=work/'go';run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere')
truth,go_time=run('go',[go]);data,_=run('maps',[go,'--json']);rows=json.loads(data)
declarations=[]
for row in rows:
 directory=root/(row['Name'].split('/')[-1].replace('-','_'))
 declaration=json.loads((directory/'rule.json').read_text())
 expected={'name':row['Name'],'kinds':row['Kinds']}
 if row['Name'] in ['no-label-var','no-invalid-regexp']:expected['node']=True
 assert declaration==expected,directory
 assert all(type(kind) is str for kind in declaration['kinds']),directory
 declarations.append(declaration)
metadata=''.join(row['name']+'\t'+','.join(map(str,row['kinds']))+'\n' for row in declarations).encode()
assert metadata==truth
mutated=json.loads((root/'radix/rule.json').read_text());mutated['kinds']=['NewExpression']
mutant_metadata=''.join((mutated if row['name']=='radix' else row)['name']+'\t'+','.join(map(str,(mutated if row['name']=='radix' else row)['kinds']))+'\n' for row in declarations).encode()
assert mutant_metadata!=truth
write('rule-json.stdout',metadata.decode());write('rule-json-mutant.stdout',mutant_metadata.decode())
json_difference=next(i for i,(a,b) in enumerate(zip(mutant_metadata,truth)) if a!=b)
source=''
for i,row in enumerate(rows):
 file=root/(row['Name'].split('/')[-1].replace('-','_')+'.a')
 source+='import {ruleName as name'+str(i)+',syntaxKinds as kinds'+str(i)+'} from '+json.dumps(str(file))+';\n'
for i,row in enumerate(rows):source+='console.log(name'+str(i)+'+\'\\t\'+kinds'+str(i)+'.join(\',\'));\n'
path=write('main.a',source);measurements={'rules':len(rows),'bytes':len(truth),'go_seconds':go_time,'sha256':hashlib.sha256(truth).hexdigest(),'rule_json_mutant_difference':json_difference}
for mode in ['normal','sanitized','mutant']:
 current=path
 if mode=='mutant':
  target=root/'radix.a';text=target.read_text();assert text.count('["CallExpression"]')==1
  mutated=write('mutated.a',text.replace('["CallExpression"]','["NewExpression"]'));current=write('mutant-main.a',source.replace(str(target),str(mutated)))
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);actual,elapsed=run(mode,[binary]);assert not (work/(mode+'.stderr')).read_bytes()
 if mode=='mutant':
  assert actual!=truth;measurements['mutant_difference']=next(i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b)
 else:assert actual==truth;measurements[mode+'_seconds']=elapsed
emitted,_=run('js-build',['/workspace/wave-09-core/adamic','js',path]);js=write('main.mjs',emitted.decode())
for name,p in [('source',path),('emitted',js)]:actual,_=run(name,['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',p]);assert actual==truth
write('results.json',json.dumps(measurements,indent=2)+'\n')
print('PASS six named listener declarations, production Go maps, native, sanitizers, both Node modes and wrong-kind mutant; shared dispatch is not exercised',flush=True)

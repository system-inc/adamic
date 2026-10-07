"""Production Go groups and exhaustive rune membership, ordered for comparison."""
import pathlib,json,subprocess,time,hashlib
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-equivalence');work.mkdir(exist_ok=True)
def write(name,text):
 p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 elapsed=time.monotonic()-start;print(name,result.returncode,round(elapsed,6),flush=True)
 assert result.returncode==0,(name,(work/(name+'.stderr')).read_text()[:2000])
 assert not (work/(name+'.stderr')).read_bytes(),name
 return (work/(name+'.stdout')).read_bytes(),elapsed
virtual=repo/'cohere/wave09_equivalence_oracle.go'
overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/equivalence_oracle.go')}}))
go=work/'go';run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere')
truth,go_time=run('go',[go])
source='import {CaseEquivalence} from '+json.dumps(str(root/'case_equivalence.a'))+''';
for(const mode of [false,true]){
 console.log(`${mode}`);
 const table=new CaseEquivalence(mode);
 for(const group of table.groups){console.log('group\\t'+group.join('\\t'));}
 for(let value=-1;value<=1114112;value++){
  const group=table.equivalents(value);
  if(group.length>0){console.log(`member\\t${value}\\t${group.join('\\t')}`);}
 }
}
'''
path=write('main.a',source);results={'answers':2228228,'bytes':len(truth),'sha256':hashlib.sha256(truth).hexdigest(),'go_seconds':go_time}
for mode in ['normal','sanitized','mutant']:
 current=path
 if mode=='mutant':
  text=(root/'case_equivalence.a').read_text()
  for dependency in ['canonicalize.a','unicode_case_data.a']:text=text.replace("'./"+dependency+"'",json.dumps(str(root/dependency)))
  assert text.count('group.length<2')==1
  changed=write('mutant.a',text.replace('group.length<2','group.length<3'))
  current=write('mutant-main.a',source.replace(str(root/'case_equivalence.a'),str(changed)))
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);actual,elapsed=run(mode,[binary])
 if mode=='mutant':
  assert actual!=truth
  results['mutant_difference']=next(i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b)
 else:assert actual==truth;results[mode+'_seconds']=elapsed
emitted,_=run('js-build',['/workspace/wave-09-core/adamic','js',path]);js=write('main.mjs',emitted.decode())
for name,p in [('source',path),('emitted',js)]:
 actual,_=run(name,['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',p]);assert actual==truth
write('results.json',json.dumps(results,indent=2)+'\n')
print('PASS production equivalence groups and exhaustive memberships, native, sanitizers, source/emitted Node and pair-dropping mutant; full regex engine remains incomplete',flush=True)

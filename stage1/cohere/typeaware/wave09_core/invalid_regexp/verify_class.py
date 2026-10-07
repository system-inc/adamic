"""Literal widening and escape spelling against unchanged Go helpers."""
import pathlib,json,subprocess,time,hashlib
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-class');work.mkdir(exist_ok=True)
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 elapsed=time.monotonic()-start;print(name,result.returncode,round(elapsed,6),flush=True)
 assert result.returncode==0,(name,(work/(name+'.stderr')).read_text()[:2000])
 assert not (work/(name+'.stderr')).read_bytes(),name
 return (work/(name+'.stdout')).read_bytes(),elapsed
virtual=repo/'cohere/wave09_class_oracle.go'
overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/class_oracle.go')}}))
go=work/'go';run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere')
truth,go_time=run('go',[go])
prefix='import {CaseClasses,escapeClassRune,escapeTrailingDash} from '+json.dumps(str(root/'case_class.a'))+''';
function spelling(text:string):string{
 const parts:string[]=[];
 for(let at=0;at<text.length;){const value=text.codePointAt(at)??0;parts.push(`${value}`);at+=value>65535?2:1;}
 return parts.join(',');
}
'''
source=prefix+'''
for(const mode of [false,true]){
 console.log(`${mode}`);const classes=new CaseClasses(mode);
 for(let value=-1;value<=1114112;value++){
  const result=classes.widen(value);
  if(result.changed||result.pattern!==''){console.log(`${value}\\t${result.changed}\\t${spelling(result.pattern)}`);}
 }
}
for(const value of [-1,0,8,9,10,13,32,45,91,92,93,94,127,55296,56319,57343,65533,65535,65536,1114111,1114112]){console.log(`escape\\t${value}\\t${spelling(escapeClassRune(value))}`);}
for(let index=0;index<4000;index++){const value=(index*7919)%1114112;console.log(`escape\\t${value}\\t${spelling(escapeClassRune(value))}`);}
'''
dash_virtual=repo/'cohere/wave09_dash_oracle.go';shim=repo/'cohere/internal/lint/ecmascript/regexp/wave09_dash_shim.go'
dash_overlay=write('dash-overlay.json',json.dumps({'Replace':{str(dash_virtual):str(root/'testdata/dash_main.go'),str(shim):str(root/'testdata/dash_oracle.go')}}))
dash_go=work/'dash-go';run('dash-go-build',['go','build','-overlay',dash_overlay,'-o',dash_go,dash_virtual],repo/'cohere')
dash_truth,_=run('dash-go',[dash_go])
bodies=[p+'\\'*n+s for p in ['', 'a','^a','👍','[','-','\x00'] for n in range(13) for s in ['-','a','--']]
dash_source=prefix+'for(const body of '+json.dumps(bodies)+'){console.log(spelling(escapeTrailingDash(body)));}\n'
results={'case_queries':2228228,'escape_queries':4021,'dash_queries':len(bodies),'go_seconds':go_time}
for label,program,expected,anchor,replacement in [('class',source,truth,"'\\\\]^-['","'\\\\]^-'") ,('dash',dash_source,dash_truth,'backslashes%2===1','backslashes%2===0')]:
 path=write(label+'.a',program);results[label+'_bytes']=len(expected);results[label+'_sha256']=hashlib.sha256(expected).hexdigest()
 for mode in ['normal','sanitized','mutant']:
  current=path
  if mode=='mutant':
   text=(root/'case_class.a').read_text().replace("'./case_equivalence.a'",json.dumps(str(root/'case_equivalence.a')))
   assert text.count(anchor)==1,(label,anchor)
   changed=write(label+'-mutant.a',text.replace(anchor,replacement));current=write(label+'-mutant-main.a',program.replace(str(root/'case_class.a'),str(changed)))
  binary=work/(label+'-'+mode);args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
  if mode=='sanitized':args.append('--sanitize')
  run(label+'-'+mode+'-build',args);actual,elapsed=run(label+'-'+mode,[binary])
  if mode=='mutant':
   assert actual!=expected
   results[label+'_mutant_difference']=next(i for i,(a,b) in enumerate(zip(actual,expected)) if a!=b)
  else:assert actual==expected;results[label+'_'+mode+'_seconds']=elapsed
 emitted,_=run(label+'-js-build',['/workspace/wave-09-core/adamic','js',path]);js=write(label+'.mjs',emitted.decode())
 for name,p in [('source',path),('emitted',js)]:
  actual,_=run(label+'-'+name,['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',p]);assert actual==expected
write('results.json',json.dumps(results,indent=2)+'\n')
print('PASS literal case widening, rune escaping and trailing dash, native, sanitizers, both Node modes and two semantic mutants; full regex rewrite remains incomplete',flush=True)

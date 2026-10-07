"""Hold native Go-style rune quoting to fmt's actual output, including Unicode boundaries."""
import pathlib,json,subprocess,time,hashlib
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-quotes');work.mkdir(exist_ok=True)
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,args)),cwd=repo,stdout=out,stderr=err,timeout=120)
 print(name,'exit',r.returncode,'seconds',round(time.monotonic()-start,6),flush=True)
 assert r.returncode==0,(name,(work/(name+'.stderr')).read_text()[:3000])
 return (work/(name+'.stdout')).read_bytes()
go=work/'go';run('go-build',['go','build','-o',go,root/'testdata/rune_oracle.go'])
ranges=json.loads(run('ranges',[go,'--ranges']))
values=set(range(160));values.update([-1,1114112,55296,56319,56320,57343,65533,1114111,2147483647])
for start,end in ranges['Ranges']:
 values.update([start-1,start,start+1,end-1,end,end+1])
values=sorted(values);fixture=write('values.json',json.dumps(values));truth=run('go',[go,fixture])
print('Unicode',ranges['Version'],'quote inputs',len(values),flush=True)
source='import {quoteRune} from '+json.dumps(str(root/'rune_quote.a'))+';\nconst values:readonly number[]='+json.dumps(values)+';for(const value of values){console.log(quoteRune(value));}\n'
path=write('quotes.a',source)
for mode in ['normal','sanitized','mutant']:
 current=path
 if mode=='mutant':
  text=(root/'rune_quote.a').read_text().replace("'./printable_ranges.a'",json.dumps(str(root/'printable_ranges.a')))
  assert text.count('value===7')==1
  mutated=write('mutated.a',text.replace('value===7','value===6'))
  current=write('mutant_main.a',source.replace(str(root/'rune_quote.a'),str(mutated)))
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);actual=run(mode,[binary]);assert not (work/(mode+'.stderr')).read_bytes()
 if mode=='mutant':
  assert actual!=truth
  diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
  print('rune escape mutant caught at byte',diff,flush=True)
 else:
  if actual!=truth:
   diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
   raise AssertionError((mode,diff,actual[max(0,diff-60):diff+80],truth[max(0,diff-60):diff+80]))
  print(mode,'identical bytes',len(truth),flush=True)
emitted=run('js-build',['/workspace/wave-09-core/adamic','js',path]);js=write('quotes.mjs',emitted.decode())
for name,p in [('source',path),('emitted',js)]:
 actual=run(name,['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',p]);assert actual==truth
write('result.json',json.dumps({'unicode_version':ranges['Version'],'inputs':len(values),'bytes':len(truth),'sha256':hashlib.sha256(truth).hexdigest(),'mutant_difference':diff},indent=2)+'\n')
print('PASS rune quoting against Go, source Node, emitted JS, mutant and sanitizers',flush=True)

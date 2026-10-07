"""Compare every mapping entry to the production Go helper, never a decoder copy."""
import pathlib,json,subprocess,time,itertools
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-offsets');work.mkdir(exist_ok=True)
def write(name,text):
 p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 print(name,result.returncode,round(time.monotonic()-start,6),flush=True)
 assert result.returncode==0,(name,(work/(name+'.stderr')).read_text())
 return (work/(name+'.stdout')).read_bytes()
atoms=[('a','a'),('á','á'),('👍','👍'),('\\n','\n'),('\\x41','A'),('\\u0041','A'),('\\u{1f44d}','👍'),('\\uD83D\\uDC4D','👍'),('\\101','A'),('\\41','!'),('\\👍','👍'),('\\z','z'),('\\uZZZZ','uZZZZ'),('',''),('a','b'),('ab','a'),('a','ab'),('\\',''),('\\\n',''),('\\\r\n','')]
rows=[(a+c,b+d) for (a,b),(c,d) in itertools.product(atoms,repeat=2)]
for a,b in atoms:rows.extend([(a+'\\\n',b),(a+'\\\r\n',b),(a+'\\',b)])
fixture=write('rows.json',json.dumps(rows))
virtual=repo/'cohere/wave09_offsets_oracle.go'
overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/offsets_oracle.go')}}))
go=work/'go';run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere')
truth=run('go',[go,fixture])
source='import {cookedToRaw} from '+json.dumps(str(root/'cooked_to_raw.a'))+';\n'
source+='class Input {readonly raw:string;readonly cooked:string;constructor(raw:string,cooked:string){this.raw=raw;this.cooked=cooked;}}\n'
source+='const rows:Input[]=['+','.join('new Input('+json.dumps(a)+','+json.dumps(b)+')' for a,b in rows)+'];\n'
source+='for(const row of rows){const offsets=cookedToRaw(row.raw,row.cooked);console.log(offsets===undefined?\'nil\':offsets.join(\',\'));}\n'
path=write('main.a',source)
for mode in ['normal','sanitized','mutant']:
 current=path
 if mode=='mutant':
  text=(root/'cooked_to_raw.a').read_text().replace("'./character_class.a'",json.dumps(str(root/'character_class.a')))
  anchor='offsets.push(rawIndex);}'
  assert text.count(anchor)==1
  mutant=write('mutated.a',text.replace(anchor,'offsets.push(rawIndex+1);}'))
  current=write('mutant-main.a',source.replace(str(root/'cooked_to_raw.a'),str(mutant)))
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);actual=run(mode,[binary]);assert not (work/(mode+'.stderr')).read_bytes()
 if mode=='mutant':
  assert actual!=truth
  difference=next(i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b)
  print('mutant caught at byte',difference,flush=True)
 else:assert actual==truth,(mode,actual[:100],truth[:100])
emitted=run('js-build',['/workspace/wave-09-core/adamic','js',path]);js=write('main.mjs',emitted.decode())
for name,p in [('source',path),('emitted',js)]:assert run(name,['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',p])==truth
write('result.json',json.dumps({'inputs':len(rows),'bytes':len(truth),'mutant_difference':difference},indent=2)+'\n')
print('PASS cooked/raw Go byte maps, native, sanitizers, source Node, emitted JS',flush=True)

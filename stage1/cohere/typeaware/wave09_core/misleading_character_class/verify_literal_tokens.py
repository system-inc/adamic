"""Native literal judgments on explicit token metadata, not the combined native parser driver."""
import pathlib,json,subprocess,time
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-literal-tokens');work.mkdir(exist_ok=True)
fixture=pathlib.Path('/workspace/wave-09-literals')
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,args)),cwd=repo,stdout=out,stderr=err,timeout=120)
 elapsed=time.monotonic()-start
 print(name,'exit',r.returncode,'seconds',round(elapsed,6),flush=True);assert r.returncode==0,(name,(work/(name+'.stderr')).read_text()[:3000]);return (work/(name+'.stdout')).read_bytes()
rows=[]
for path in (fixture/'manifest').read_text().splitlines():
 source=pathlib.Path(path).read_bytes().decode();start=source.index('const pattern=')+len('const pattern=');end=source.index(';',start)
 rows.append({'path':path,'text':source[start:end],'start':len(source[:start].encode()),'end':len(source[:end].encode())})
write('inputs.json',json.dumps(rows,ensure_ascii=False,indent=2)+'\n')
module=root/'literal_rule.a'
source='import {literalTextFindings} from '+json.dumps(str(module))+';\nimport {written} from '+json.dumps(str(repo/'stage1/typescript/parser/nodes.ts'))+';\n'
source+='class Input {readonly path:string;readonly text:string;readonly start:number;readonly end:number;constructor(path:string,text:string,start:number,end:number){this.path=path;this.text=text;this.start=start;this.end=end;}}\n'
source+='const rows:readonly Input[]=['+','.join('new Input('+','.join([json.dumps(row['path']),json.dumps(row['text'],ensure_ascii=False),str(row['start']),str(row['end'])])+')' for row in rows)+'];\n'
# Two independent default/option runs in one binary avoid needing the checker runtime.
source+="for(const allowed of [false,true]){console.log(allowed?'allow':'default');let count=0;for(const row of rows){console.log(`file\\t${written(row.path)}`);const found=literalTextFindings(row.text,row.start,row.end,allowed);for(const finding of found){finding.sortKey=finding.written();}found.sort((a,b)=>a.sortKey<b.sortKey?-1:a.sortKey>b.sortKey?1:0);for(const finding of found){console.log(finding.written());}count+=found.length;}console.log(`findings ${count}`);}\n"
path=write('main.a',source)
truth=b'default\n'+(fixture/'default-go.stdout').read_bytes()+b'allow\n'+(fixture/'allow-go.stdout').read_bytes()
write('expected.txt',truth.decode())
for mode in ['normal','sanitized','mutant']:
 current=path
 if mode=='mutant':
  text=module.read_text()
  for name in ['pattern_findings.a','no_misleading_character_class.a']:text=text.replace("'./"+name+"'",json.dumps(str(root/name)))
  text=text.replace("'../../", "'"+str(repo/'stage1/cohere/typeaware')+'/').replace(str(repo/'stage1/cohere/typeaware')+'/../../typescript/',str(repo/'stage1/typescript')+'/')
  assert text.count("new Repair(end,end,'u')")==1
  mutant=write('mutated.a',text.replace("new Repair(end,end,'u')","new Repair(end,end+1,'u')"));current=write('mutant_main.a',source.replace(str(module),str(mutant)))
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);actual=run(mode,[binary]);assert not (work/(mode+'.stderr')).read_bytes()
 if mode=='mutant':
  assert actual!=truth;diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)));print('native suggestion edit mutant caught byte',diff,flush=True)
 else:assert actual==truth,(mode,len(actual),len(truth));print(mode,'identical bytes',len(truth),flush=True)
emitted=run('js-build',['/workspace/wave-09-core/adamic','js',path]);js=write('main.mjs',emitted.decode())
for name,p in [('source',path),('emitted',js)]:assert run(name,['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',p])==truth
write('result.json',json.dumps({'files':len(rows),'bytes':len(truth),'mutant_difference':diff,'actual_native_source_parser_tested':False},indent=2)+'\n')
print('PASS native literal token judgments and suggestion edits; source parsing is verified separately by verify_literals.py',flush=True)

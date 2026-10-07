"""Hold resolved constructor-argument judgments to complete production rule bytes."""
import pathlib,json,subprocess,time,itertools,hashlib
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-constructors');work.mkdir(exist_ok=True)
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 elapsed=time.monotonic()-start
 print(name,result.returncode,round(elapsed,6),flush=True)
 assert result.returncode==0,(name,(work/(name+'.stderr')).read_text()[:2000])
 return (work/(name+'.stdout')).read_bytes(),elapsed
patterns=['[Á]','[👍]','[👶🏻]','[🇯🇵]','[👨‍👩‍👦]','[👍][👍]','[Á][B́]',r'[A\u0301]',r'[\uD83D\uDC4D]',r'[\u{d83d}\u{dc4d}]','[a-z👍]','[👍-￿]','[🇯[A]🇵]','[Á] [','{ [Á]','[]']
rows=[]
for pattern,flags,encoding,kind in itertools.product(patterns,['','u','v'],['plain','json','template'],['string','indirect','concat']):
 raw=json.dumps(pattern,ensure_ascii=encoding=='json')
 if encoding=='template':raw='`'+pattern.replace('\\','\\\\')+'`'
 prefix='/* 世界 🌍 */\r\n'
 if kind=='indirect':source=prefix+'const pattern='+raw+';RegExp(pattern,'+json.dumps(flags)+');';argument='pattern';start=len((prefix+'const pattern='+raw+';RegExp(').encode())
 elif kind=='concat':source=prefix+'RegExp('+raw+'+"",'+json.dumps(flags)+');';argument=raw+'+""';start=len((prefix+'RegExp(').encode())
 else:source=prefix+'RegExp('+raw+','+json.dumps(flags)+');';argument=raw;start=len((prefix+'RegExp(').encode())
 rows.append({'kind':kind,'raw':raw,'pattern':pattern,'flags':flags,'known':True,'hasFlags':True,'start':start,'end':start+len(argument.encode()),'source':source+'\r\nexport {};\r\n'})
for pattern,ownFlags,flags in itertools.product(['[👍]','[Á]',r'[\uD83D\uDC4D]'],['','u'],['','u','v','unknown','absent']):
 raw='/'+pattern+'/'+ownFlags;prefix='/* 世界 🌍 */\r\nRegExp('
 tail='' if flags=='absent' else ',dynamicFlags' if flags=='unknown' else ','+json.dumps(flags)
 source=prefix+raw+tail+');\r\nexport {};\r\n'
 rows.append({'kind':'regex','raw':raw,'pattern':pattern,'flags':flags,'known':flags!='unknown','hasFlags':flags!='absent','start':len(prefix.encode()),'end':len((prefix+raw).encode()),'source':source})
paths=[str(write(f'control-{i}.a',row['source'])) for i,row in enumerate(rows)]
manifest=write('manifest','\n'.join(paths)+'\n');write('root.d.ts','declare const dynamicFlags:string;')
config=write('tsconfig.json',json.dumps({'compilerOptions':{'target':'ES2022','module':'NodeNext','moduleResolution':'NodeNext','lib':['es2022','dom']},'files':['root.d.ts']}))
virtual=repo/'cohere/wave09_constructor_oracle.go';overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/literal_oracle.go')}}))
go=work/'go';run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere')
valid,_=run('valid',[go,config,manifest,'--valid-sources']);assert valid==manifest.read_bytes()
source='import {stringPatternFindings,indirectPatternFindings,regexArgumentFindings} from '+json.dumps(str(root/'constructor_pattern.a'))+';\n'
source+='import {literalTextFindings} from '+json.dumps(str(root/'literal_rule.a'))+';\n'
source+='import {Diagnostic} from '+json.dumps(str(repo/'stage1/cohere/typeaware/diagnostic.ts'))+';\n'
source+='import {written} from '+json.dumps(str(repo/'stage1/typescript/parser/nodes.ts'))+';\nimport {programArguments} from \'adamic\';\n'
source+='class Input {readonly kind:string;readonly raw:string;readonly pattern:string;readonly flags:string;readonly known:boolean;readonly hasFlags:boolean;readonly start:number;readonly end:number;readonly path:string;constructor(kind:string,raw:string,pattern:string,flags:string,known:boolean,hasFlags:boolean,start:number,end:number,path:string){this.kind=kind;this.raw=raw;this.pattern=pattern;this.flags=flags;this.known=known;this.hasFlags=hasFlags;this.start=start;this.end=end;this.path=path;}}\n'
source+='const rows:Input[]=['+','.join('new Input('+','.join(json.dumps(row[key]) for key in ['kind','raw','pattern','flags','known','hasFlags','start','end'])+','+json.dumps(path)+')' for row,path in zip(rows,paths))+'];\n'
source+='const allowed=programArguments().includes(\'--allow-escape\');let count=0;for(const row of rows){const findings:Diagnostic[]=row.kind===\'string\'?stringPatternFindings(row.raw,row.pattern,row.flags,row.start,allowed):row.kind===\'regex\'?(row.hasFlags?regexArgumentFindings(row.raw,row.flags,row.known,true,row.start,allowed):literalTextFindings(row.raw,row.start,row.end,allowed)):indirectPatternFindings(row.pattern,row.flags,row.start,row.end);for(const finding of findings){finding.sortKey=finding.written();}findings.sort((a,b)=>a.sortKey<b.sortKey?-1:a.sortKey>b.sortKey?1:0);console.log(`file\\t${written(row.path)}`);for(const finding of findings){console.log(finding.written());}count+=findings.length;}console.log(`findings ${count}`);\n'
path=write('main.a',source);commands={}
for mode in ['normal','sanitized']:
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',path,'-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);commands[mode]=[binary]
emitted,_=run('js-build',['/workspace/wave-09-core/adamic','js',path]);js=write('main.mjs',emitted.decode())
commands['source']=['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',path];commands['emitted']=['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',js]
measurements={'inputs':len(rows)}
for allowed in [False,True]:
 extra=['--allow-escape'] if allowed else [];name='allow' if allowed else 'default'
 truth,elapsed=run(name+'-go',[go,config,manifest]+extra);measurements[name]={'go':elapsed,'bytes':len(truth),'findings':truth.splitlines()[-1].decode(),'sha256':hashlib.sha256(truth).hexdigest()}
 for mode,command in commands.items():
  actual,elapsed=run(name+'-'+mode,command+extra);assert not (work/(name+'-'+mode+'.stderr')).read_bytes()
  if actual!=truth:
   diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)));raise AssertionError((name,mode,diff,actual[max(0,diff-100):diff+200],truth[max(0,diff-100):diff+200]))
  measurements[name][mode]=elapsed
  print(name,mode,'identical bytes',len(truth),flush=True)
text=(root/'constructor_pattern.a').read_text()
for file in ['character_class.a','cooked_to_raw.a','pattern_findings.a']:text=text.replace("'./"+file+"'",json.dumps(str(root/file)))
text=text.replace("'../../diagnostic.ts'",json.dumps(str(repo/'stage1/cohere/typeaware/diagnostic.ts')))
assert text.count('tokenStart+1+end')==1
mutated=write('mutated.a',text.replace('tokenStart+1+end','tokenStart+2+end'))
mutant=write('mutant-main.a',source.replace(str(root/'constructor_pattern.a'),str(mutated)));binary=work/'mutant'
run('mutant-build',['/workspace/wave-09-core/adamic','build',mutant,'-o',binary]);actual,_=run('mutant',[binary]);truth,_=run('mutant-go',[go,config,manifest]);assert actual!=truth and not (work/'mutant.stderr').read_bytes()
measurements['mutant_difference']=next(i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b)
write('measurements.json',json.dumps(measurements,indent=2)+'\n')
print('PASS resolved constructor argument judgments, full Go bytes, source/emitted Node, sanitizers and span mutant; native reference/constant resolution is not exercised',flush=True)

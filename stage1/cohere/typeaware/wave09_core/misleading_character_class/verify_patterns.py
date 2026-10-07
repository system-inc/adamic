"""Native parser and whole-pattern comparisons against unmodified Go implementations."""
import pathlib,json,subprocess,time,itertools,hashlib,shutil
root=pathlib.Path(__file__).resolve().parent
repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-patterns');work.mkdir(exist_ok=True)
def write(name,text):path=work/name;path.write_text(text);return path
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err)
 elapsed=time.monotonic()-start
 print(name,'exit',result.returncode,'seconds',round(elapsed,6),flush=True)
 assert result.returncode==0,(name,(work/(name+'.stderr')).read_text()[:3000])
 return (work/(name+'.stdout')).read_bytes(),elapsed
members=['A','́','👍','🏻','🇯','🇵','‍',r'\uD83D',r'\uDC4D',r'\u{1f44d}',r'\u{d83d}',r'\x41',r'\d',r'\B','-','[A]',r'\q{xy}',r'\p{L}',r'\cI','\\👍']
patterns=['','[]','[^]','[','[abc','[Á] [','\\',r'[\]',r'[\u{zz}]',r'[\u{ffffffffffffffff}]',r'[\uD83D\uDC4D]',r'[\x4]',r'[\u004]',r'[\u{]',r'[\p{]',r'[\q{]',r'[\c]',r'[\c🌍A]',r'\c🌍[Á]',r'[[\]]]',r'[[[[[x]]]]]',r'[\q{a]b}]','[0-[]]','[a-b-c]','[👍-￿]','[👍-\uFFFF]','[👨‍👩‍👦]','{ [Á]','a{2}[Á]','a{2,}[Á]','a{2,3}[Á]','a{,3}[Á]',r'\u{41}[Á]','}[Á]','🌍[Á]世界','[Á][👍]']
patterns += ['['+left+right+']' for left,right in itertools.product(members,repeat=2)]
rows=[{'Pattern':pattern,'Flags':flags,'AllowEscape':allowed} for pattern,flags,allowed in itertools.product(patterns,['','u','v'],[False,True])]
fixture=write('patterns.json',json.dumps(rows,ensure_ascii=False))
print('fixture rows',len(rows),flush=True)
prelude='class Input {readonly pattern:string;readonly flags:string;readonly allowed:boolean;constructor(pattern:string,flags:string,allowed:boolean){this.pattern=pattern;this.flags=flags;this.allowed=allowed;}}\n'
prelude+='const rows:Input[]=['+','.join('new Input('+json.dumps(row['Pattern'],ensure_ascii=False)+','+json.dumps(row['Flags'])+','+str(row['AllowEscape']).lower()+')' for row in rows)+'];\n'
parser_source='import {PatternBytes,RegexFlags,scanClasses,parseClassWithEnd} from '+json.dumps(str(root/'character_class.a'))+';\n'+prelude
parser_source+='for(let i=0;i<rows.length;i++){const row=rows[i];if(row!==undefined){console.log(`case ${i}`);const p=new PatternBytes(row.pattern);const f=new RegexFlags(row.flags);const scan=scanClasses(p,f);console.log(`scan ${scan.ok?1:0}`);for(const span of scan.spans){console.log(`span ${span.start} ${span.end}`);const parsed=parseClassWithEnd(p,span.start,span.end,f);console.log(`parse ${parsed.ok?1:0} ${parsed.end}`);for(const element of parsed.elements){console.log(element.written());}}}}\n'
pattern_source='import {patternFindings} from '+json.dumps(str(root/'pattern_findings.a'))+';\n'+prelude
pattern_source+='for(let i=0;i<rows.length;i++){const row=rows[i];if(row!==undefined){console.log(`case ${i}`);for(const finding of patternFindings(row.pattern,row.flags,row.allowed)){console.log(finding.written());}}}\n'
measurements={}
for kind,source,main in [('parser',parser_source,'parser_oracle.go'),('pattern',pattern_source,'pattern_oracle.go')]:
 virtual=repo/('cohere/wave09_'+kind+'_oracle.go')
 replacements={str(virtual):str(root/'testdata'/main)}
 if kind=='pattern':
  for name in ['sequence_shim.go','pattern_shim.go']:
   replacements[str(repo/'cohere/internal/lint/rules/core'/('wave09_'+name))]=str(root/'testdata'/name)
 overlay=write(kind+'-overlay.json',json.dumps({'Replace':replacements}))
 oracle=work/(kind+'-go')
 run(kind+'-go-build',['go','build','-overlay',overlay,'-o',oracle,virtual],repo/'cohere')
 truth,go_time=run(kind+'-go',[oracle,fixture])
 path=write(kind+'.a',source)
 measurements[kind]={'go':go_time,'bytes':len(truth),'sha256':hashlib.sha256(truth).hexdigest()}
 for mode in ['normal','sanitized','mutant']:
  current=path
  if mode=='mutant':
   target=root/('character_class.a' if kind=='parser' else 'pattern_findings.a')
   text=target.read_text()
   for imported in ['character_class.a','no_misleading_character_class.a']:text=text.replace("'./"+imported+"'",json.dumps(str(root/imported)))
   anchor='if(f.uv()&&high>=55296' if kind=='parser' else 'if(!scan.ok){return [];}'
   replacement='if(false&&high>=55296' if kind=='parser' else 'if(false){return [];}'
   assert text.count(anchor)==1
   mutant=write(kind+'-mutated.a',text.replace(anchor,replacement))
   current=write(kind+'-mutant-main.a',source.replace(str(target),str(mutant)))
  binary=work/(kind+'-'+mode)
  args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
  if mode=='sanitized':args.append('--sanitize')
  run(kind+'-'+mode+'-build',args)
  actual,elapsed=run(kind+'-'+mode,[binary]);measurements[kind][mode]=elapsed
  assert not (work/(kind+'-'+mode+'.stderr')).read_bytes()
  if mode=='mutant':
   assert actual!=truth
   diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
   print(kind,'mutant caught at byte',diff,flush=True);measurements[kind]['mutant_difference']=diff
  else:
   if actual!=truth:
    diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
    raise AssertionError((kind,mode,diff,actual[max(0,diff-50):diff+200],truth[max(0,diff-50):diff+200]))
   print(kind,mode,'identical bytes',len(truth),flush=True)
 # Emitted JavaScript is a second execution oracle for pure native helper code.
 emitted,_=run(kind+'-js-build',['/workspace/wave-09-core/adamic','js',path])
 js=write(kind+'.mjs',emitted.decode())
 actual,elapsed=run(kind+'-js',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',js]);assert actual==truth
 actual,_=run(kind+'-source-node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',path]);assert actual==truth
 measurements[kind]['javascript']=elapsed
write('measurements.json',json.dumps(measurements,indent=2)+'\n')
print('PASS native parser and whole-pattern helpers, Go bytes, Node, mutants, sanitizers',flush=True)

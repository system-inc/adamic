"""Compare partial native helpers with unmodified production Go helper calls."""
import itertools,json,pathlib,subprocess,time,os
root=pathlib.Path(__file__).resolve().parent
repo=root.parents[3]
work=pathlib.Path('/workspace/wave-09-core/helpers');work.mkdir(parents=True,exist_ok=True)
def write(name,text):
 p=work/name;p.write_text(text);return p
def run(name,args,env=None,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=cwd,env=env,stdout=out,stderr=err)
 elapsed=time.monotonic()-start
 print(name,result.returncode,round(elapsed,6),flush=True)
 assert result.returncode==0,(name,result.returncode)
 return (work/(name+'.stdout')).read_bytes(),elapsed
def oracle(directory,shim,main):
 virtual=repo/'cohere/wave09_helper_oracle.go'
 package=repo/'cohere/internal/lint/rules/core/wave09_helper_shim.go'
 overlay=write(directory+'-overlay.json',json.dumps({'Replace':{str(virtual):str(root/directory/'testdata'/main),str(package):str(root/directory/'testdata'/shim)}}))
 binary=work/(directory+'-go')
 run(directory+'-oracle-build',['go','build','-overlay',overlay,'-o',binary,virtual],cwd=repo/'cohere')
 return binary

def check(name,source,fixture,go,mutant_anchor,mutant_replacement,module):
 path=write(name+'.a',source)
 truth,go_time=run(name+'-go',[go,fixture])
 timings={}
 for mode in ['normal','sanitized','mutant']:
  current=path
  if mode=='mutant':
   text=module.read_text();assert text.count(mutant_anchor)==1
   mutated=write(name+'-mutated.a',text.replace(mutant_anchor,mutant_replacement).replace("from '../../", "from '"+str(repo/'stage1/cohere/typeaware')+'/').replace(str(repo/'stage1/cohere/typeaware')+'/../../typescript/',str(repo/'stage1/typescript')+'/'))
   current=write(name+'-mutant-main.a',source.replace(str(module),str(mutated)))
  binary=work/(name+'-'+mode)
  args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
  if mode=='sanitized':args.append('--sanitize')
  run(name+'-'+mode+'-build',args)
  actual,elapsed=run(name+'-'+mode,[binary]);timings[mode]=elapsed
  assert not (work/(name+'-'+mode+'.stderr')).read_bytes()
  if mode=='mutant':assert actual!=truth
  else:assert actual==truth,(name,mode,len(actual),len(truth))
 print(name,'production helper bytes',len(truth),'mutant caught',flush=True)
 return {'go':go_time,'native':timings,'bytes':len(truth)}
# Flags are scalar runes; explicit quote values exercise formatter input separately.
rows=[]
for extra in [[],['z'],['zq'],['🌍']]:
 for flags in ['', 'gim', 'uv', 'vu', 'guuv', 'gg', 'zg', 'zz', 'zgg', 'qzz', '🌍','🌍🌍','dgy','smm','vv']:
  known=set('dgimsuvy'+''.join(extra));seen=set();remaining=[]
  for c in flags:
   if c in known and c not in seen:seen.add(c)
   else:remaining.append(c)
  fault=next((c for c in remaining if c in known),remaining[0] if remaining else '')
  rows.append({'flags':flags,'extra':extra,'quoted':"'"+fault+"'"})
fixture=write('flags.json',json.dumps(rows,ensure_ascii=False))
module=root/'invalid_regexp/no_invalid_regexp.a'
source="import {invalidFlags,flagsMessage} from "+json.dumps(str(module))+";\n"
for row in rows:
 source+='console.log(flagsMessage('+json.dumps(row['flags'],ensure_ascii=False)+',invalidFlags('+json.dumps(row['flags'],ensure_ascii=False)+','+json.dumps(row['extra'],ensure_ascii=False)+'),'+json.dumps(row['quoted'],ensure_ascii=False)+'));\n'
go=oracle('invalid_regexp','flags_shim.go','flags_oracle.go')
results={'flags':check('flags',source,fixture,go,"if(flags.includes('u') && flags.includes('v'))",'if(false)',module)}
# Boundary, escape and surrogate variants plus separated and chained joiners.
values=[65,0x2ff,0x300,0x36f,0x370,0x1aaf,0x1ab0,0x1aff,0x1b00,0x1dc0,0x1dff,0x20d0,0x20ff,0xfe00,0xfe0f,0xfe20,0xfe2f,0xe0100,0xe01ef,0x1f1e5,0x1f1e6,0x1f1ff,0x1f200,0x1f3fa,0x1f3fb,0x1f3ff,0x1f400,0xd7ff,0xd800,0xdbff,0xdc00,0xdfff,0xe000]
sequences=[]
for left,right in itertools.product(values,repeat=2):
 sequences.append([left,right])
sequences += [[65,0x200d,66],[65,0x200d,66,0x200d,67],[65,0x200d,66,67,0x200d,68],[0x200d,0x200d,65],[65,0x200d,0x200d]]
rows=[]
for values_ in sequences:
 for escaped,codepoint in [(False,False),(True,False),(False,True)]:
  rows.append([{'Value':v,'Start':i*3+2,'End':i*3+5,'CodePointEscape':codepoint,'Escaped':escaped and i==1} for i,v in enumerate(values_)])
patterns=['','a','{','}','a{2}',r'\u{1f44d}',r'\u{',r'\q',r'\a',r'\d',r'\\',r'\[',r'\p{L}',r'\u{41}x{',r'\u{41}x',r'\🌍']
fixture=write('sequences.json',json.dumps({'Rows':rows,'Patterns':patterns},ensure_ascii=False))
module=root/'misleading_character_class/no_misleading_character_class.a'
source="import {ClassCharacter,sequenceFindings,patternMeaningChanges} from "+json.dumps(str(module))+";\n"
# One array table keeps generated C manageable.
source+='const rows:ClassCharacter[][]=['
source+=','.join('['+','.join('new ClassCharacter('+','.join([str(v['Value']),str(v['Start']),str(v['End']),str(v['CodePointEscape']).lower(),str(v['Escaped']).lower()])+')' for v in row)+']' for row in rows)+'];\n'
source+="for(let i=0;i<rows.length;i++){console.log(`case ${i}`);for(const found of sequenceFindings(rows[i]??[])){console.log(found.written());}}\n"
for pattern in patterns: source+='console.log(patternMeaningChanges('+json.dumps(pattern,ensure_ascii=False)+') ? \"true\" : \"false\");\n'
go=oracle('misleading_character_class','sequence_shim.go','sequence_oracle.go')
results['sequences']=check('sequences',source,fixture,go,'value >= 0x1f3fb','value >= 0x1f3fc',module)
write('measurements.json',json.dumps(results,indent=2)+'\n')
print('PASS partial helpers only; full rule adapters remain blocked',flush=True)

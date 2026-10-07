"""Hold the real constructor flag options against the upstream decoded options."""
import pathlib,subprocess,json,time
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-flag-options');work.mkdir(exist_ok=True)
base=pathlib.Path('/workspace/wave-09-core');fixtures=base/'flags-frontend'
def run(name,args,refused=False):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=repo,stdout=out,stderr=err,timeout=180)
 elapsed=time.monotonic()-start
 assert (result.returncode!=0 if refused else result.returncode==0),(name,(work/(name+'.stderr')).read_text())
 print(name,result.returncode,round(elapsed,6),flush=True)
 return (work/(name+'.stdout')).read_bytes(),elapsed
flags=['az','aa','zz','A','u','uu','uv','vu','zgg','🌍','🌍🌍','🌍A','aZ','', 'y','𝄞𝄞']
paths=[]
for index,flag in enumerate(flags):
 path=work/f'control-{index}.a'
 path.write_text('/* 世界 */\r\ndeclare const pattern:string;\r\nnew RegExp(pattern,'+json.dumps(flag)+');\r\nexport {};\r\n')
 paths.append(str(path))
manifest=work/'manifest';manifest.write_text('\n'.join(paths)+'\n')
args=[fixtures/'tsconfig.json',manifest]
commands={'normal':[fixtures/'native'],'sanitized':[fixtures/'native-asan']}
measurements={}
for name,allowed in [('default',[]),('joined',['az','u','v','🌍','𝄞']),('case',['a',''])]:
 extra=['--allow-flag='+flag for flag in allowed]
 truth,go_time=run(name+'-go',[fixtures/'oracle',*args,*extra])
 assert b'findings ' in truth
 measurements[name]={'bytes':len(truth),'findings':truth.splitlines()[-1].decode(),'go':go_time}
 for mode,command in commands.items():
  actual,elapsed=run(name+'-'+mode,[*command,*args,*extra])
  assert actual==truth,(name,mode)
  assert not (work/(name+'-'+mode+'.stderr')).read_bytes()
  measurements[name][mode]=elapsed
# Removing options must run cleanly and disagree, rather than failing to build.
text=(root/'main.a').read_text()
for filename in ['no_invalid_regexp.a','rune_quote.a','flag_options.a']:
 text=text.replace("from './"+filename+"'","from '"+str(root/filename)+"'")
text=text.replace("from '../../../../typescript/","from '"+str(repo/'stage1/typescript')+'/').replace("from '../../","from '"+str(repo/'stage1/cohere/typeaware')+'/')
assert text.count('const allowed = constructorFlags(args);')==1
probe=work/'mutant.a';probe.write_text(text.replace('const allowed = constructorFlags(args);','const allowed: string[] = [];'))
binary=work/'mutant'
run('mutant-build',[base/'adamic','build',probe,'-o',binary,'--tsgo',base/'checker.a'])
extra=['--allow-flag='+flag for flag in ['az','u','v','🌍','𝄞']]
truth,_=run('mutant-go',[fixtures/'oracle',*args,*extra]);actual,_=run('mutant',[binary,*args,*extra])
assert actual!=truth and not (work/'mutant.stderr').read_bytes()
measurements['mutant_difference']=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
# Both boundaries reject duplicate array items; error formatting is not equated.
for name,command in [('go',[fixtures/'oracle']),*commands.items()]:
 run(name+'-duplicate',[*command,*args,'--allow-flag=a','--allow-flag=a'],refused=True)
 assert b'items must be unique' in (work/(name+'-duplicate.stderr')).read_bytes(),name
(work/'results.json').write_text(json.dumps(measurements,indent=2)+'\n')
print('PASS decoded constructor options, full finding/fix/suggestion bytes, sanitizers and ignored-options mutant; pattern compilation remains blocked',flush=True)

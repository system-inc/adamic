"""Full finding/fix/suggestion bytes for the literal listener only."""
import pathlib,json,subprocess,time,hashlib,statistics,sys
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-literals');work.mkdir(exist_ok=True)
source_only='--source-only' in sys.argv
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 elapsed=time.monotonic()-start
 print(name,'exit',r.returncode,'seconds',round(elapsed,6),flush=True)
 assert r.returncode==0,(name,(work/(name+'.stderr')).read_text()[:3000])
 return (work/(name+'.stdout')).read_bytes(),elapsed
virtual=repo/'cohere/wave09_literal_oracle.go'
overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/literal_oracle.go')}}))
go=work/'go';run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere')
patterns=[('[Á]',''),('[Á]','u'),('[👍]',''),('[👍]','u'),('[👍]','v'),('[👶🏻]','u'),('[🇯🇵]','u'),('[👨‍👩‍👦]',''),('[👨‍👩‍👦]','u'),('[a-z👍]',''),('[👍-￿]',''),(r'[\uD83D\uDC4D]',''),(r'[\u{d83d}\u{dc4d}]','u'),(r'[A\u0301]','u'),(r'[\n̅]',''),(r'[\👍]',''),(r'[A\x41]',''),(r'[\uD83D\uDC4D]-\a',''),(r'[\uD83D\uDC4D]{2}',''),('[🇯[A]🇵]','v'),('[🇯--🇵]','v'),(r'[🇯\q{xy}🇵]','v'),('[Á][👍]',''),('[]',''),('[a-z]',''),('[́]',''),(r'[\uD83D]','')]
paths=[str(write(f'control-{i}.a',f'/* 世界 🌍 */\r\nconst pattern=/{pattern}/{flags};\r\nexport {{}};\r\n')) for i,(pattern,flags) in enumerate(patterns)]
manifest=write('manifest','\n'.join(paths)+'\n');config=write('tsconfig.json',json.dumps({'compilerOptions':{'target':'ES2022','module':'NodeNext','moduleResolution':'NodeNext','lib':['es2022','dom']},'files':['root.d.ts']}));write('root.d.ts','')
valid,_=run('valid-filter',[go,config,manifest,'--valid-sources']);assert len(valid.splitlines())==len(paths);manifest.write_bytes(valid)
binaries={}
for mode in ([] if source_only else ['normal','sanitized']):
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',root/'literal_main.a','-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);binaries[mode]=binary
commands={'source-node':['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',root/'literal_main.a']}
if not source_only:
 emitted,_=run('js-build',['/workspace/wave-09-core/adamic','js',root/'literal_main.a']);js=write('literal.mjs',emitted.decode())
 commands.update({'normal':[binaries['normal']],'sanitized':[binaries['sanitized']],'emitted-js':['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',js]})
measurements={}
for allowed in [False,True]:
 name='allow' if allowed else 'default';extra=['--allow-escape'] if allowed else []
 truth,go_time=run(name+'-go',[go,config,manifest]+extra)
 assert b'\tno-misleading-character-class\t' in truth
 measurements[name]={'go':go_time,'bytes':len(truth),'findings':truth.splitlines()[-1].decode()}
 for mode,command in commands.items():
  actual,elapsed=run(name+'-'+mode,command+[config,manifest]+extra)
  assert not (work/(name+'-'+mode+'.stderr')).read_bytes()
  if actual!=truth:
   diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
   raise AssertionError((name,mode,diff,actual[max(0,diff-50):diff+200],truth[max(0,diff-50):diff+200]))
  measurements[name][mode]=elapsed
  print(name,mode,'identical bytes',len(truth),flush=True)
# A clean-running suggestion-edit mutant is held by full serialization.
source=(root/'literal_rule.a').read_text()
for file in ['pattern_findings.a','no_misleading_character_class.a']:source=source.replace("'./"+file+"'",json.dumps(str(root/file)))
source=source.replace("'../../", "'"+str(repo/'stage1/cohere/typeaware')+'/').replace(str(repo/'stage1/cohere/typeaware')+'/../../typescript/',str(repo/'stage1/typescript')+'/')
assert source.count("new Repair(end,end,'u')")==1
mutated=write('mutated.a',source.replace("new Repair(end,end,'u')","new Repair(end,end+1,'u')"))
source=(root/'literal_main.a').read_text()
for file in ['character_class.a']:source=source.replace("'./"+file+"'",json.dumps(str(root/file)))
source=source.replace("'./literal_rule.a'",json.dumps(str(mutated)))
source=source.replace("'../../", "'"+str(repo/'stage1/cohere/typeaware')+'/').replace(str(repo/'stage1/cohere/typeaware')+'/../../typescript/',str(repo/'stage1/typescript')+'/')
mutant=write('mutant_main.a',source);binary=work/'mutant'
if not source_only:run('mutant-build',['/workspace/wave-09-core/adamic','build',mutant,'-o',binary])
truth,_=run('mutant-go',[go,config,manifest]);actual,_=run('mutant',(['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',mutant] if source_only else [binary])+[config,manifest]);assert actual!=truth and not (work/'mutant.stderr').read_bytes()
diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
measurements['suggestion_mutant_difference']=diff;print('suggestion edit mutant caught at byte',diff,flush=True)
# The corpus bytes exercise the literal slice, not unfinished constructor tracking.
for name,config,manifest in [('compiler',pathlib.Path('/workspace/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/tsconfig.json'),pathlib.Path('/workspace/wave-09-validation/compiler.manifest')),('repository',repo/'tsconfig.json',pathlib.Path('/workspace/wave-09-validation/repository.manifest'))]:
 truth,go_time=run(name+'-go',[go,config,manifest]);measurements[name]={'go':go_time,'bytes':len(truth),'findings':truth.splitlines()[-1].decode()}
 for mode,command in commands.items():
  actual,elapsed=run(name+'-'+mode,command+[config,manifest]);assert actual==truth
  measurements[name][mode]=elapsed;print(name,mode,'identical bytes',len(truth),flush=True)
measurements['source_only']=source_only
write('measurements.json',json.dumps(measurements,indent=2)+'\n')
print('PASS available literal diagnostics and suggestions; native whole listener and constructor tracking are not certified' if source_only else 'PASS literal slice full diagnostics and suggestions; constructor tracking remains incomplete',flush=True)

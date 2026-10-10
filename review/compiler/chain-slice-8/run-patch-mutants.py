import pathlib,json,re,subprocess,time
root=pathlib.Path('/workspace/adamic'); out=root/'review/compiler/chain-slice-8/patch-mutants';out.mkdir(exist_ok=True)
manifest=root/'stage3/checked-any/mutants/next-contracts/manifest.json'
cases=[(x['name'],manifest.parent/x['patch'],'./'+x['package'],x['selector']) for x in json.loads(manifest.read_text())]
for name,selector in [('array-layout','TestCheckedAnyUnsupportedContracts/array_view'),('evolving-any','TestTasteRepresentationLimitsStayExplicit'),('explicit-any-policy','TestCheckedAny'),('literal-result','TestCheckedAnyUnsupportedContracts/literal_result'),('primitive-prototype','TestCheckedAnyUnsupportedContracts/prototype'),('staged-field','TestCheckedAnyUnsupportedContracts/staged_field'),('string-length','TestCheckedAny/string_length'),('typed-field','TestCheckedAnyUnsupportedContracts/field_view')]:
 path=root/'stage3/checked-any/mutants'/(name+'.patch')
 if path.exists(): cases.append((name,path,'./internal/lower' if name=='evolving-any' else './internal/oracle',selector))
for name in ['source','span']:
 cases.append(('classification-'+name,root/'stage3/checked-any/mutants'/('next-classification-'+name+'.patch'),'./stage3/checked-any/probe','^TestRefusalMustBeInsideDeclaration$'))
results=[]
for name,patch,package,selector in cases:
 d=out/name;d.mkdir(exist_ok=True); repl={}; lines=patch.read_text().splitlines(keepends=True); files={}; current=None; i=0
 try:
  while i<len(lines):
   line=lines[i]
   if line.startswith('+++ b/'):
    current=line[6:].strip();files[current]=(root/current).read_text();i+=1;continue
   if line.startswith('@@'):
    i+=1; old=[];new=[]
    while i<len(lines) and not lines[i].startswith(('@@','diff --git','--- a/','+++ b/')):
     s=lines[i]
     if s.startswith((' ','-')):old.append(s[1:])
     if s.startswith((' ','+')):new.append(s[1:])
     i+=1
    a=''.join(old);b=''.join(new)
    if a not in files[current]: raise ValueError('stale hunk '+current)
    files[current]=files[current].replace(a,b,1);continue
   i+=1
  for i,(file,source) in enumerate(files.items()):
   target=d/(str(i)+pathlib.Path(file).suffix+'.txt');target.write_text(source);repl[str(root/file)]=str(target)
  overlay=d/'overlay.json';overlay.write_text(json.dumps({'Replace':repl}))
  cmd=['go','test','-p','4','-overlay='+str(overlay),package,'-run',selector,'-count=1','-timeout','90s','-v']
  start=time.monotonic()
  with (d/'test.log').open('w') as f: code=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=120).returncode
  log=(d/'test.log').read_text(); caught=code!=0 and '--- FAIL:' in log and not any(x in log for x in ['[build failed]','clang failed:','no tests to run','panic: test timed out'])
  results.append(dict(name=name,exit=code,caught=caught,seconds=time.monotonic()-start,selector=selector,log=str((d/'test.log').relative_to(root))))
 except Exception as e: results.append(dict(name=name,caught=False,error=str(e)))
 (out/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(name,results[-1],flush=True)
raise SystemExit(any(not x['caught'] for x in results))

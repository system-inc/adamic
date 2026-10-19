import pathlib,subprocess,json,difflib,time,sys,os
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-checked-writes-ac-m2';specs=json.loads((root/'review/compiler/loom-r-checked-writes/mutants.json').read_text());start_index=int(os.environ.get('MUTANT_START','0'));results=json.loads((out/'mutant-results.json').read_text())[:start_index] if start_index else []
for item in specs[start_index:]:
 replace={}
 for target,oldmut in json.loads((root/item['overlay']).read_text())['Replace'].items():
  rel=str(pathlib.Path(target).relative_to(root));base=subprocess.check_output(['git','show','5aa205ec00329ce78789ddb62277f0ccdcc1f337:'+rel],text=True);mutant=pathlib.Path(oldmut).read_text();current=(root/rel).read_text();a=base.splitlines(True);b=mutant.splitlines(True);groups=list(difflib.SequenceMatcher(None,a,b,autojunk=False).get_grouped_opcodes(3))
  for group in reversed(groups):
   old=''.join(a[group[0][1]:group[-1][2]]);new=''.join(b[group[0][3]:group[-1][4]])
   assert current.count(old)==1,(item['name'],rel,'ambiguous or stale mutation anchor',old[:120]);current=current.replace(old,new,1)
  suffix='.c.txt'if rel.endswith('.c')else'.go.txt';p=out/(item['name']+suffix);p.write_text(current);replace[str(root/rel)]=str(p)
 overlay=out/(item['name']+'.json');overlay.write_text(json.dumps({'Replace':replace},indent=2)+'\n');command=['timeout','120','go','test','-overlay',str(overlay),item['package'],'-run',item['selector'],'-count=1','-parallel=4','-v','-timeout=90s'];start=time.monotonic()
 with(out/(item['name']+'.log')).open('w')as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
 body=(out/(item['name']+'.log')).read_text();caught=r.returncode==1 and '--- FAIL:'in body and 'build failed'not in body and 'test timed out'not in body and '# github.com/'not in body and 'error: 'not in body
 results.append(dict(item,command=command,exit=r.returncode,caught=caught,seconds=round(time.monotonic()-start,3)));(out/'mutant-results.json').write_text(json.dumps(results,indent=2)+'\n');print(item['name'],'caught'if caught else'NOT CAUGHT',flush=True)
 if not caught:sys.exit(1)
